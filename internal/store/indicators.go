package store

import (
	"context"
	"database/sql"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// ExpenseStats summarizes what the congressmen of a chamber spent in a month.
// Count is how many of them have a record; with none, the other fields are 0.
type ExpenseStats struct {
	Avg   float64
	Min   int
	Max   int
	Count int
}

// Scope selects the congressmen a statistic is computed over: the whole
// chamber, or only those of one constituency (a district of the Chamber of
// Deputies, a region of the Senate) when ConstituencyName is set.
type Scope struct {
	Chamber          congress.Chamber
	ConstituencyKind congress.ConstituencyKind
	ConstituencyName string
}

// scopeSQL is the WHERE condition for a Scope; its four placeholders are
// filled by Scope.args.
const scopeSQL = `c.chamber = ? AND (? = '' OR (c.constituency_name = ? AND c.constituency_kind = ?))`

// args returns the placeholders of scopeSQL followed by rest.
func (sc Scope) args(rest ...any) []any {
	return append([]any{sc.Chamber, sc.ConstituencyName, sc.ConstituencyName, sc.ConstituencyKind}, rest...)
}

// Local reports whether the scope is narrower than the chamber.
func (sc Scope) Local() bool { return sc.ConstituencyName != "" }

// Month identifies a calendar month.
type Month struct {
	Year  int
	Month int
}

// EverDone reports whether the unit of work (job, key) was completed in any
// period. The API uses it to tell "not fetched yet" from "fetched, and empty".
func (s *Store) EverDone(ctx context.Context, job, key string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM job_state WHERE job = ? AND key = ?`, job, key).Scan(&n)
	return n > 0, err
}

// ExpenseMonths returns the latest months with operational expenses recorded
// for a congressman, most recent first.
func (s *Store) ExpenseMonths(ctx context.Context, congressmanID int64, limit int) ([]Month, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT year, month FROM expenses_operational
		WHERE congressman_id = ? ORDER BY year DESC, month DESC LIMIT ?`, congressmanID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Month
	for rows.Next() {
		var m Month
		if err := rows.Scan(&m.Year, &m.Month); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) stats(ctx context.Context, scope Scope, query string, args ...any) (ExpenseStats, error) {
	var (
		avg      sql.NullFloat64
		min, max sql.NullInt64
		st       ExpenseStats
	)
	if err := s.db.QueryRowContext(ctx, query, scope.args(args...)...).Scan(&avg, &min, &max, &st.Count); err != nil {
		return ExpenseStats{}, err
	}
	st.Avg, st.Min, st.Max = avg.Float64, int(min.Int64), int(max.Int64)
	return st, nil
}

// OperationalCategoryStats compares one expense category in a month across the
// congressmen of a chamber.
func (s *Store) OperationalCategoryStats(ctx context.Context, scope Scope, year, month int, category string) (ExpenseStats, error) {
	return s.stats(ctx, scope, `
		SELECT AVG(e.amount), MIN(e.amount), MAX(e.amount), COUNT(*)
		FROM expenses_operational e JOIN congressmen c ON c.id = e.congressman_id
		WHERE `+scopeSQL+` AND e.year = ? AND e.month = ? AND e.expense_type = ?`,
		year, month, category)
}

// OperationalMonthStats compares the total operational spending of a month
// across the congressmen of a chamber.
func (s *Store) OperationalMonthStats(ctx context.Context, scope Scope, year, month int) (ExpenseStats, error) {
	return s.stats(ctx, scope, `
		WITH totals AS (
			SELECT SUM(e.amount) AS amount
			FROM expenses_operational e JOIN congressmen c ON c.id = e.congressman_id
			WHERE `+scopeSQL+` AND e.year = ? AND e.month = ?
			GROUP BY e.congressman_id
		)
		SELECT AVG(amount), MIN(amount), MAX(amount), COUNT(*) FROM totals`,
		year, month)
}

// OperationalRank is the position of a congressman among those of their chamber
// with the highest operational spending in a month (1 = the most). ok is false
// when they have no record that month.
func (s *Store) OperationalRank(ctx context.Context, scope Scope, congressmanID int64, year, month int) (rank int, ok bool, err error) {
	err = s.db.QueryRowContext(ctx, `
		WITH ranked AS (
			SELECT e.congressman_id AS id,
				RANK() OVER (ORDER BY SUM(e.amount) DESC) AS rnk
			FROM expenses_operational e JOIN congressmen c ON c.id = e.congressman_id
			WHERE `+scopeSQL+` AND e.year = ? AND e.month = ?
			GROUP BY e.congressman_id
		)
		SELECT rnk FROM ranked WHERE id = ?`,
		scope.args(year, month, congressmanID)...).Scan(&rank)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	return rank, err == nil, err
}

// StaffMonthStats compares support staff spending in a month across the
// congressmen of a chamber.
func (s *Store) StaffMonthStats(ctx context.Context, scope Scope, year, month int) (ExpenseStats, error) {
	return s.stats(ctx, scope, `
		SELECT AVG(e.total_amount), MIN(e.total_amount), MAX(e.total_amount), COUNT(*)
		FROM expenses_support_staff e JOIN congressmen c ON c.id = e.congressman_id
		WHERE `+scopeSQL+` AND e.year = ? AND e.month = ?`,
		year, month)
}

// StaffRank is the position of a congressman among those of their chamber with
// the highest support staff spending in a month (1 = the most).
func (s *Store) StaffRank(ctx context.Context, scope Scope, congressmanID int64, year, month int) (rank int, ok bool, err error) {
	err = s.db.QueryRowContext(ctx, `
		WITH ranked AS (
			SELECT e.congressman_id AS id,
				RANK() OVER (ORDER BY e.total_amount DESC) AS rnk
			FROM expenses_support_staff e JOIN congressmen c ON c.id = e.congressman_id
			WHERE `+scopeSQL+` AND e.year = ? AND e.month = ?
		)
		SELECT rnk FROM ranked WHERE id = ?`,
		scope.args(year, month, congressmanID)...).Scan(&rank)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	return rank, err == nil, err
}
