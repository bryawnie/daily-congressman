// Package jobs runs the scheduled work: the monthly synchronization of every
// congressman's profile and expenses, and the daily draw of the congressman of
// the day followed by a refresh of that congressman's data.
//
// Every job is idempotent and records what it completed in the database
// (store.MarkDone), so it is safe to call RunDue as often as wanted, to run it
// from an external cron or from the built-in loop (Serve), and to resume after
// a crash or a partial failure. When everything is done a call makes no
// network request.
package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bryawnie/daily-congressman/internal/beacon"
	"github.com/bryawnie/daily-congressman/internal/chooser"
	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/retrieval"
	"github.com/bryawnie/daily-congressman/internal/source"
	"github.com/bryawnie/daily-congressman/internal/store"
)

// Job names, as stored in job_state and job_errors.
const (
	JobProfiles = "profiles" // monthly, every congressman
	JobExpenses = "expenses" // monthly, every congressman
	JobChoose   = "choose"   // daily, the draw
	JobRefresh  = "refresh"  // daily, the chosen congressman's changing data
)

// allKey is the key of the marker that says a whole run completed.
const allKey = "*"

// ErrPulseNotReady means the beacon has not published the day's pulse yet, so
// the draw cannot be made. It is not a failure: try again later.
var ErrPulseNotReady = errors.New("jobs: the day's beacon pulse is not available yet")

// Jobs holds what the scheduled jobs need.
type Jobs struct {
	Store *store.Store
	// Sources are the active chambers: the ones that take part in the draw and
	// whose congressmen are synchronized.
	Sources *source.Registry
	Chooser *chooser.Chooser
	Syncer  *retrieval.Syncer
	// Pause is the wait between two congressmen in a monthly run, to be gentle
	// with the sources.
	Pause time.Duration
	Now   func() time.Time
	Log   *slog.Logger
}

func (j *Jobs) now() time.Time {
	if j.Now != nil {
		return j.Now()
	}
	return time.Now()
}

func (j *Jobs) log() *slog.Logger {
	if j.Log != nil {
		return j.Log
	}
	return slog.Default()
}

// Today is the current day (YYYY-MM-DD) in Chile.
func (j *Jobs) Today() string { return chooser.Today(j.now()) }

// Month is the current month (YYYY-MM) in Chile.
func (j *Jobs) Month() string { return j.Today()[:7] }

func memberKey(m congress.Member) string {
	return fmt.Sprintf("%s:%d", m.Chamber, m.ExternalID)
}

func (j *Jobs) fail(ctx context.Context, job, key, period string, err error) {
	j.log().Error("job failed", "job", job, "key", key, "period", period, "error", err)
	if lerr := j.Store.LogJobError(ctx, store.JobError{At: j.now(), Job: job, Key: key, Period: period, Error: err.Error()}); lerr != nil {
		j.log().Error("could not log the failure", "error", lerr)
	}
}

func (j *Jobs) pause(ctx context.Context) {
	if j.Pause <= 0 {
		return
	}
	select {
	case <-ctx.Done():
	case <-time.After(j.Pause):
	}
}

// syncMonthly runs work for every current congressman of the active chambers,
// once per month. Congressmen already done this month are skipped unless force
// is set, so a run that failed halfway resumes with only what is missing. The
// month is marked complete only when no congressman failed.
func (j *Jobs) syncMonthly(ctx context.Context, job string, force bool, work func(context.Context, congress.Member) error) error {
	period := j.Month()
	if !force {
		if done, err := j.Store.IsDone(ctx, job, allKey, period); err != nil {
			return err
		} else if done {
			return nil
		}
	}
	j.log().Info("monthly sync starting", "job", job, "period", period, "force", force)

	var failures []error
	var synced, skipped int
	for _, ch := range j.Sources.Chambers() {
		src, err := j.Sources.Get(ch)
		if err != nil {
			return err
		}
		members, err := src.CurrentMembers(ctx)
		if err != nil {
			err = fmt.Errorf("members of %s: %w", ch, err)
			j.fail(ctx, job, string(ch), period, err)
			failures = append(failures, err)
			continue
		}
		for _, m := range members {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			key := memberKey(m)
			if !force {
				if done, err := j.Store.IsDone(ctx, job, key, period); err != nil {
					return err
				} else if done {
					skipped++
					continue
				}
			}
			if err := work(ctx, m); err != nil {
				err = fmt.Errorf("%s (%s): %w", key, m.Name, err)
				j.fail(ctx, job, key, period, err)
				failures = append(failures, err)
			} else if err := j.Store.MarkDone(ctx, job, key, period, j.now()); err != nil {
				return err
			} else {
				synced++
			}
			j.pause(ctx)
		}
	}
	j.log().Info("monthly sync finished", "job", job, "period", period, "synced", synced, "skipped", skipped, "failed", len(failures))
	if len(failures) > 0 {
		return fmt.Errorf("%s: %d failure(s), the missing ones will be retried: %w", job, len(failures), errors.Join(failures...))
	}
	return j.Store.MarkDone(ctx, job, allKey, period, j.now())
}

// SyncProfiles stores the profile of every current congressman. It runs once a
// month, so the rest of the system can read profiles from the database.
func (j *Jobs) SyncProfiles(ctx context.Context, force bool) error {
	return j.syncMonthly(ctx, JobProfiles, force, func(ctx context.Context, m congress.Member) error {
		_, _, err := j.Syncer.Profile(ctx, m)
		if errors.Is(err, congress.ErrNotSupported) {
			return nil // data pending, not a failure
		}
		return err
	})
}

// SyncExpenses stores the monthly expenses of every current congressman, so
// they can be compared. It runs once a month.
func (j *Jobs) SyncExpenses(ctx context.Context, force bool) error {
	return j.syncMonthly(ctx, JobExpenses, force, func(ctx context.Context, m congress.Member) error {
		id, err := j.Syncer.Ensure(ctx, m)
		if err != nil {
			return err
		}
		report, err := j.Syncer.Expenses(ctx, id, m)
		if err != nil {
			return err
		}
		return report.Err()
	})
}

// Choose draws the congressman of the day, once the beacon's pulse for 00:00
// of that day is available, and stores the choice. If the day already has a
// choice it is returned untouched (created is false): a day is never drawn
// twice. If the pulse is not out yet it returns ErrPulseNotReady. Any other
// failure also stores nothing, so the day is drawn later from complete data.
func (j *Jobs) Choose(ctx context.Context, day string) (choice congress.DayChoice, created bool, err error) {
	if existing, err := j.Store.GetDayChoice(ctx, day); err == nil {
		return existing, false, nil
	} else if !errors.Is(err, store.ErrNotFound) {
		return congress.DayChoice{}, false, err
	}
	if day > j.Today() {
		return congress.DayChoice{}, false, fmt.Errorf("jobs: %s is in the future", day)
	}

	res, err := j.Chooser.Choose(ctx, day)
	if errors.Is(err, beacon.ErrNotAvailable) {
		return congress.DayChoice{}, false, ErrPulseNotReady
	}
	if err != nil {
		return congress.DayChoice{}, false, err
	}
	id, err := j.Syncer.Ensure(ctx, res.Member)
	if err != nil {
		return congress.DayChoice{}, false, fmt.Errorf("jobs: profile of the chosen congressman: %w", err)
	}
	choice = congress.DayChoice{
		Date:          day,
		CongressmanID: id,
		ChainID:       res.Pulse.ChainID,
		PulseID:       res.Pulse.PulseID,
		RandOut:       res.Pulse.RandOut,
	}
	created, err = j.Store.SaveDayChoice(ctx, choice)
	if err != nil {
		return congress.DayChoice{}, false, err
	}
	if !created { // another process chose first; its choice stands
		choice, err = j.Store.GetDayChoice(ctx, day)
		return choice, false, err
	}
	j.log().Info("congressman of the day chosen", "day", day, "congressman_id", id, "member", res.Member.Name)
	return choice, true, nil
}

// Refresh updates the data of the day's congressman that changes from day to
// day (attendance, latest votings, law projects). It does nothing if it was
// already done for that day, unless force is set. Data the source does not
// provide yet is skipped as pending.
func (j *Jobs) Refresh(ctx context.Context, day string, force bool) error {
	choice, err := j.Store.GetDayChoice(ctx, day)
	if err != nil {
		return fmt.Errorf("jobs: refresh %s: %w", day, err)
	}
	c, err := j.Store.GetCongressman(ctx, choice.CongressmanID)
	if err != nil {
		return err
	}
	m := congress.Member{Chamber: c.Chamber, ExternalID: c.ExternalID}
	key := memberKey(m)
	if !force {
		if done, err := j.Store.IsDone(ctx, JobRefresh, key, day); err != nil {
			return err
		} else if done {
			return nil
		}
	}
	report, err := j.Syncer.Daily(ctx, choice.CongressmanID, m)
	if err != nil {
		return err
	}
	for _, s := range report {
		if s.Status == retrieval.Pending {
			j.log().Info("data pending, not supported by the source yet", "key", key, "data", s.Name)
		}
	}
	if err := report.Err(); err != nil {
		j.fail(ctx, JobRefresh, key, day, err)
		return err
	}
	j.log().Info("congressman of the day refreshed", "day", day, "key", key)
	return j.Store.MarkDone(ctx, JobRefresh, key, day, j.now())
}

// Daily draws the congressman of the day (if not drawn yet) and refreshes their
// data (if not refreshed yet).
func (j *Jobs) Daily(ctx context.Context, day string, force bool) error {
	if _, _, err := j.Choose(ctx, day); err != nil {
		return err
	}
	return j.Refresh(ctx, day, force)
}

// RunDue does whatever is due now: the daily draw and refresh first, since they
// are time-sensitive, then the monthly synchronizations. It continues past a
// failing job and returns all the errors together. A pulse that is not out yet
// is logged, not returned.
func (j *Jobs) RunDue(ctx context.Context) error {
	var errs []error
	day := j.Today()
	if _, _, err := j.Choose(ctx, day); errors.Is(err, ErrPulseNotReady) {
		j.log().Info("waiting for the day's beacon pulse", "day", day)
	} else if err != nil {
		j.fail(ctx, JobChoose, "", day, err)
		errs = append(errs, err)
	} else if err := j.Refresh(ctx, day, false); err != nil { // already logged by Refresh
		errs = append(errs, err)
	}
	if err := j.SyncProfiles(ctx, false); err != nil {
		errs = append(errs, err)
	}
	if err := j.SyncExpenses(ctx, false); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// Serve calls RunDue now and then every interval until ctx is cancelled. A
// failed run is logged and tried again on the next tick; only the congressmen
// or steps that are still missing are redone.
func (j *Jobs) Serve(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return errors.New("jobs: interval must be positive")
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		if err := j.RunDue(ctx); err != nil && ctx.Err() == nil {
			j.log().Error("run finished with errors", "error", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}
