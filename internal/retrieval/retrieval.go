// Package retrieval copies data from a chamber's source into the database.
//
// Data is grouped by how often it changes, because the jobs refresh each group
// on its own schedule: the profile and the expenses change monthly, while
// attendance, votings and law projects change from day to day.
package retrieval

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/source"
	"github.com/bryawnie/daily-congressman/internal/store"
)

// Status is the outcome of one kind of data.
type Status int

const (
	// Saved means the data was read and stored.
	Saved Status = iota
	// Pending means the source does not provide this data yet
	// (congress.ErrNotSupported). It is not a failure.
	Pending
	// Failed means the source or the database returned an error.
	Failed
)

// Step is the outcome of retrieving one kind of data.
type Step struct {
	Name   string
	Status Status
	Detail string // what was saved, for Saved
	Err    error  // for Failed
}

// Report lists the steps of one retrieval, in order.
type Report []Step

// Failed returns the steps that failed.
func (r Report) Failed() []Step {
	var out []Step
	for _, s := range r {
		if s.Status == Failed {
			out = append(out, s)
		}
	}
	return out
}

// Err returns an error naming the failed steps, or nil.
func (r Report) Err() error {
	var errs []error
	for _, s := range r.Failed() {
		errs = append(errs, fmt.Errorf("%s: %w", s.Name, s.Err))
	}
	return errors.Join(errs...)
}

// Syncer reads from the sources and writes to the store.
type Syncer struct {
	Store   *store.Store
	Sources *source.Registry
	// VotingsLimit is how many of the latest votings are kept.
	VotingsLimit int
	// Attempts is how many times a read is tried before it is reported as
	// failed (default 3). Backoff is the wait before the second try, doubled
	// after each one (default 2s).
	Attempts int
	Backoff  time.Duration
	Now      func() time.Time
}

func (s *Syncer) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Syncer) votingsLimit() int {
	if s.VotingsLimit > 0 {
		return s.VotingsLimit
	}
	return 10
}

// retry runs fn up to Attempts times, waiting between tries. A source that does
// not support the data, or a cancelled context, is never retried.
func retry[T any](ctx context.Context, s *Syncer, fn func() (T, error)) (T, error) {
	attempts, wait := s.Attempts, s.Backoff
	if attempts <= 0 {
		attempts = 3
	}
	if wait <= 0 {
		wait = 2 * time.Second
	}
	var v T
	var err error
	for i := 0; i < attempts; i++ {
		if v, err = fn(); err == nil || errors.Is(err, congress.ErrNotSupported) || ctx.Err() != nil {
			return v, err
		}
		if i < attempts-1 {
			select {
			case <-ctx.Done():
				return v, ctx.Err()
			case <-time.After(wait):
			}
			wait *= 2
		}
	}
	return v, err
}

func step(name string, err error, detail string) Step {
	switch {
	case errors.Is(err, congress.ErrNotSupported):
		return Step{Name: name, Status: Pending}
	case err != nil:
		return Step{Name: name, Status: Failed, Err: err}
	}
	return Step{Name: name, Status: Saved, Detail: detail}
}

// Profile reads and stores a congressman's profile and returns its internal ID.
// Every other table hangs from it, so a failure here is returned as an error.
func (s *Syncer) Profile(ctx context.Context, m congress.Member) (int64, congress.Congressman, error) {
	src, err := s.Sources.Get(m.Chamber)
	if err != nil {
		return 0, congress.Congressman{}, err
	}
	p, err := retry(ctx, s, func() (congress.Congressman, error) { return src.Profile(ctx, m) })
	if err != nil {
		return 0, p, fmt.Errorf("profile of %s/%d: %w", m.Chamber, m.ExternalID, err)
	}
	p.LastUpdate = s.now()
	id, err := s.Store.UpsertCongressman(ctx, &p)
	return id, p, err
}

// Ensure returns the internal ID of a congressman, reading the profile first if
// the database does not have it yet.
func (s *Syncer) Ensure(ctx context.Context, m congress.Member) (int64, error) {
	c, err := s.Store.FindCongressman(ctx, m.Chamber, m.ExternalID)
	if err == nil {
		return c.ID, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return 0, err
	}
	id, _, err := s.Profile(ctx, m)
	return id, err
}

// Daily refreshes what changes from day to day: attendance, latest votings
// and law projects.
func (s *Syncer) Daily(ctx context.Context, id int64, m congress.Member) (Report, error) {
	src, err := s.Sources.Get(m.Chamber)
	if err != nil {
		return nil, err
	}
	var r Report

	att, err := retry(ctx, s, func() (congress.Attendance, error) { return src.Attendance(ctx, m) })
	if err == nil {
		err = s.Store.UpsertAttendance(ctx, id, att)
	}
	r = append(r, step("attendance", err, fmt.Sprintf("%d of %d sessions", att.Present, att.Total)))

	votings, err := retry(ctx, s, func() ([]congress.Voting, error) { return src.Votings(ctx, m, s.votingsLimit()) })
	if err == nil {
		err = s.Store.ReplaceVotings(ctx, id, votings)
	}
	r = append(r, step("votings", err, fmt.Sprintf("%d saved", len(votings))))

	projects, err := retry(ctx, s, func() ([]congress.LawProject, error) { return src.Activity(ctx, m) })
	if err == nil {
		err = s.Store.UpsertLawProjects(ctx, id, projects)
	}
	r = append(r, step("activity", err, fmt.Sprintf("%d projects", len(projects))))
	return r, ctx.Err()
}

// Expenses refreshes the monthly expenses (operational, offices, staff).
func (s *Syncer) Expenses(ctx context.Context, id int64, m congress.Member) (Report, error) {
	src, err := s.Sources.Get(m.Chamber)
	if err != nil {
		return nil, err
	}
	exp, err := retry(ctx, s, func() (congress.Expenses, error) { return src.Expenses(ctx, m) })
	if err == nil {
		if err = s.Store.UpsertOperationalExpenses(ctx, id, exp.Operational); err == nil {
			if err = s.Store.UpsertOfficeExpenses(ctx, id, exp.Offices); err == nil {
				err = s.Store.UpsertStaffExpenses(ctx, id, exp.Staff)
			}
		}
	}
	detail := fmt.Sprintf("%d operational, %d offices, %d staff rows", len(exp.Operational), len(exp.Offices), len(exp.Staff))
	return Report{step("expenses", err, detail)}, ctx.Err()
}
