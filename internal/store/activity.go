package store

import (
	"context"
	"database/sql"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// UpsertPeriods stores the parliamentary periods of a congressman.
func (s *Store) UpsertPeriods(ctx context.Context, congressmanID int64, periods []congress.Period) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, p := range periods {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO parliamentary_periods (congressman_id, period_from, period_to)
				VALUES (?, ?, ?)
				ON CONFLICT DO NOTHING`,
				congressmanID, p.From, p.To); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListPeriods returns the periods of a congressman, most recent first.
func (s *Store) ListPeriods(ctx context.Context, congressmanID int64) ([]congress.Period, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT period_from, period_to FROM parliamentary_periods
		WHERE congressman_id = ? ORDER BY period_from DESC`, congressmanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []congress.Period
	for rows.Next() {
		var p congress.Period
		if err := rows.Scan(&p.From, &p.To); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpsertAttendance stores the attendance summary of a congressman.
func (s *Store) UpsertAttendance(ctx context.Context, congressmanID int64, a congress.Attendance) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO attendance (congressman_id, total, present, justified_absent, unjustified_absent)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (congressman_id) DO UPDATE SET
			total = excluded.total,
			present = excluded.present,
			justified_absent = excluded.justified_absent,
			unjustified_absent = excluded.unjustified_absent`,
		congressmanID, a.Total, a.Present, a.JustifiedAbsent, a.UnjustifiedAbsent)
	return err
}

// GetAttendance returns the attendance summary, or ErrNotFound if none is stored.
func (s *Store) GetAttendance(ctx context.Context, congressmanID int64) (congress.Attendance, error) {
	var a congress.Attendance
	err := s.db.QueryRowContext(ctx, `
		SELECT total, present, justified_absent, unjustified_absent
		FROM attendance WHERE congressman_id = ?`, congressmanID,
	).Scan(&a.Total, &a.Present, &a.JustifiedAbsent, &a.UnjustifiedAbsent)
	return a, notFound(err)
}

// ReplaceVotings replaces all stored votings of a congressman with the given
// ones. The legacy flow keeps only the latest votings, so old ones are dropped.
func (s *Store) ReplaceVotings(ctx context.Context, congressmanID int64, votings []congress.Voting) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM votings WHERE congressman_id = ?`, congressmanID); err != nil {
			return err
		}
		for _, v := range votings {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO votings (congressman_id, voting_id, voting_date, bulletin_number,
					document_title, article_text, voted_option, total_approved, total_rejected,
					total_abstention, quorum, result)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT (congressman_id, voting_id) DO UPDATE SET
					voting_date = excluded.voting_date,
					bulletin_number = excluded.bulletin_number,
					document_title = excluded.document_title,
					article_text = excluded.article_text,
					voted_option = excluded.voted_option,
					total_approved = excluded.total_approved,
					total_rejected = excluded.total_rejected,
					total_abstention = excluded.total_abstention,
					quorum = excluded.quorum,
					result = excluded.result`,
				congressmanID, v.VotingID, v.Date, v.BulletinNumber, v.DocumentTitle, v.ArticleText,
				v.VotedOption, v.TotalApproved, v.TotalRejected, v.TotalAbstention, v.Quorum, v.Result); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListVotings returns the votings of a congressman, most recent first.
func (s *Store) ListVotings(ctx context.Context, congressmanID int64) ([]congress.Voting, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT voting_id, voting_date, bulletin_number, document_title, article_text,
			voted_option, total_approved, total_rejected, total_abstention, quorum, result
		FROM votings WHERE congressman_id = ?
		ORDER BY voting_date DESC, voting_id DESC`, congressmanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []congress.Voting
	for rows.Next() {
		var v congress.Voting
		if err := rows.Scan(&v.VotingID, &v.Date, &v.BulletinNumber, &v.DocumentTitle, &v.ArticleText,
			&v.VotedOption, &v.TotalApproved, &v.TotalRejected, &v.TotalAbstention, &v.Quorum, &v.Result); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// UpsertLawProjects stores the bills and motions of a congressman.
func (s *Store) UpsertLawProjects(ctx context.Context, congressmanID int64, projects []congress.LawProject) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		for _, p := range projects {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO law_projects (congressman_id, bulletin, project_name, project_type,
					project_status, project_date)
				VALUES (?, ?, ?, ?, ?, ?)
				ON CONFLICT (congressman_id, bulletin) DO UPDATE SET
					project_name = excluded.project_name,
					project_type = excluded.project_type,
					project_status = excluded.project_status,
					project_date = excluded.project_date`,
				congressmanID, p.Bulletin, p.Name, p.Type, p.Status, p.Date); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListLawProjects returns the projects of a congressman, most recent first.
func (s *Store) ListLawProjects(ctx context.Context, congressmanID int64) ([]congress.LawProject, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT bulletin, project_name, project_type, project_status, project_date
		FROM law_projects WHERE congressman_id = ?
		ORDER BY project_date DESC, bulletin DESC`, congressmanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []congress.LawProject
	for rows.Next() {
		var p congress.LawProject
		if err := rows.Scan(&p.Bulletin, &p.Name, &p.Type, &p.Status, &p.Date); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
