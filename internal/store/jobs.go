package store

import (
	"context"
	"time"
)

// JobError is one logged failure of a job.
type JobError struct {
	At     time.Time
	Job    string
	Key    string
	Period string
	Error  string
}

// MarkDone records that the unit of work (job, key) was completed for period.
// Marking it again is harmless.
func (s *Store) MarkDone(ctx context.Context, job, key, period string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO job_state (job, key, period, done_at) VALUES (?, ?, ?, ?)
		ON CONFLICT (job, key, period) DO NOTHING`, job, key, period, formatTime(at))
	return err
}

// IsDone reports whether the unit of work was completed for period.
func (s *Store) IsDone(ctx context.Context, job, key, period string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM job_state WHERE job = ? AND key = ? AND period = ?`,
		job, key, period).Scan(&n)
	return n > 0, err
}

// ClearDone forgets the completion of a job in a period (every key), which
// forces the next run to do the work again.
func (s *Store) ClearDone(ctx context.Context, job, period string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM job_state WHERE job = ? AND period = ?`, job, period)
	return err
}

// LogJobError appends a failure to the log.
func (s *Store) LogJobError(ctx context.Context, e JobError) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO job_errors (at, job, key, period, error) VALUES (?, ?, ?, ?, ?)`,
		formatTime(e.At), e.Job, e.Key, e.Period, e.Error)
	return err
}

// ListJobErrors returns the latest logged failures, most recent first.
func (s *Store) ListJobErrors(ctx context.Context, limit int) ([]JobError, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT at, job, key, period, error FROM job_errors ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []JobError
	for rows.Next() {
		var e JobError
		var at string
		if err := rows.Scan(&at, &e.Job, &e.Key, &e.Period, &e.Error); err != nil {
			return nil, err
		}
		e.At = parseTime(at)
		out = append(out, e)
	}
	return out, rows.Err()
}
