package store

import (
	"context"
	"database/sql"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// UpsertOperationalExpenses stores monthly operational expenses by category.
func (s *Store) UpsertOperationalExpenses(ctx context.Context, congressmanID int64, expenses []congress.OperationalExpense) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, e := range expenses {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO expenses_operational (congressman_id, year, month, expense_type, amount)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT (congressman_id, year, month, expense_type) DO UPDATE SET
					amount = excluded.amount`,
				congressmanID, e.Year, e.Month, e.ExpenseType, e.Amount); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListOperationalExpenses returns operational expenses, most recent month first.
func (s *Store) ListOperationalExpenses(ctx context.Context, congressmanID int64) ([]congress.OperationalExpense, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT year, month, expense_type, amount FROM expenses_operational
		WHERE congressman_id = ? ORDER BY year DESC, month DESC, expense_type`, congressmanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []congress.OperationalExpense
	for rows.Next() {
		var e congress.OperationalExpense
		if err := rows.Scan(&e.Year, &e.Month, &e.ExpenseType, &e.Amount); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// UpsertOfficeExpenses stores monthly office expenses.
func (s *Store) UpsertOfficeExpenses(ctx context.Context, congressmanID int64, expenses []congress.OfficeExpense) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, e := range expenses {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO expenses_offices (congressman_id, year, month, offices_number, total_amount)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT (congressman_id, year, month) DO UPDATE SET
					offices_number = excluded.offices_number,
					total_amount = excluded.total_amount`,
				congressmanID, e.Year, e.Month, e.OfficesNumber, e.TotalAmount); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListOfficeExpenses returns office expenses, most recent month first.
func (s *Store) ListOfficeExpenses(ctx context.Context, congressmanID int64) ([]congress.OfficeExpense, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT year, month, offices_number, total_amount FROM expenses_offices
		WHERE congressman_id = ? ORDER BY year DESC, month DESC`, congressmanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []congress.OfficeExpense
	for rows.Next() {
		var e congress.OfficeExpense
		if err := rows.Scan(&e.Year, &e.Month, &e.OfficesNumber, &e.TotalAmount); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// UpsertStaffExpenses stores monthly support staff expenses.
func (s *Store) UpsertStaffExpenses(ctx context.Context, congressmanID int64, expenses []congress.StaffExpense) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, e := range expenses {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO expenses_support_staff (congressman_id, year, month, hired_staff, total_amount)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT (congressman_id, year, month) DO UPDATE SET
					hired_staff = excluded.hired_staff,
					total_amount = excluded.total_amount`,
				congressmanID, e.Year, e.Month, e.HiredStaff, e.TotalAmount); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListStaffExpenses returns support staff expenses, most recent month first.
func (s *Store) ListStaffExpenses(ctx context.Context, congressmanID int64) ([]congress.StaffExpense, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT year, month, hired_staff, total_amount FROM expenses_support_staff
		WHERE congressman_id = ? ORDER BY year DESC, month DESC`, congressmanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []congress.StaffExpense
	for rows.Next() {
		var e congress.StaffExpense
		if err := rows.Scan(&e.Year, &e.Month, &e.HiredStaff, &e.TotalAmount); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
