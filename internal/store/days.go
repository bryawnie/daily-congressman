package store

import (
	"context"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// SaveDayChoice records the congressman chosen for a day. A day is chosen only
// once: if the date already has a choice, nothing changes and created is false.
func (s *Store) SaveDayChoice(ctx context.Context, c congress.DayChoice) (created bool, err error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO congressman_of_the_day (date, congressman_id, chain_id, pulse_id, rand_out)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (date) DO NOTHING`,
		c.Date, c.CongressmanID, c.ChainID, c.PulseID, c.RandOut)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// GetDayChoice returns the choice for a date, or ErrNotFound.
func (s *Store) GetDayChoice(ctx context.Context, date string) (congress.DayChoice, error) {
	var c congress.DayChoice
	err := s.db.QueryRowContext(ctx, `
		SELECT date, congressman_id, chain_id, pulse_id, rand_out
		FROM congressman_of_the_day WHERE date = ?`, date,
	).Scan(&c.Date, &c.CongressmanID, &c.ChainID, &c.PulseID, &c.RandOut)
	return c, notFound(err)
}

// ListDayChoices returns the latest choices, most recent date first.
func (s *Store) ListDayChoices(ctx context.Context, limit int) ([]congress.DayChoice, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT date, congressman_id, chain_id, pulse_id, rand_out
		FROM congressman_of_the_day ORDER BY date DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []congress.DayChoice
	for rows.Next() {
		var c congress.DayChoice
		if err := rows.Scan(&c.Date, &c.CongressmanID, &c.ChainID, &c.PulseID, &c.RandOut); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
