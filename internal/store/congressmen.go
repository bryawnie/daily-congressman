package store

import (
	"context"
	"fmt"
	"time"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

const congressmanColumns = `id, chamber, external_id, first_name, first_surname, second_surname,
	profile_picture, gender, birthdate, profession, party, party_alias,
	twitter_username, instagram_username, constituency_kind, constituency_name,
	constituency_detail, last_update`

// UpsertCongressman inserts the congressman or updates it if (chamber,
// external_id) already exists, and returns its internal ID. It uses an upsert
// rather than INSERT OR REPLACE so the ID, and the rows that reference it, are
// preserved.
func (s *Store) UpsertCongressman(ctx context.Context, c *congress.Congressman) (int64, error) {
	if !c.Chamber.Valid() {
		return 0, fmt.Errorf("store: invalid chamber %q", c.Chamber)
	}
	var id int64
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO congressmen (chamber, external_id, first_name, first_surname, second_surname,
			profile_picture, gender, birthdate, profession, party, party_alias,
			twitter_username, instagram_username, constituency_kind, constituency_name,
			constituency_detail, last_update)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (chamber, external_id) DO UPDATE SET
			first_name = excluded.first_name,
			first_surname = excluded.first_surname,
			second_surname = excluded.second_surname,
			profile_picture = excluded.profile_picture,
			gender = excluded.gender,
			birthdate = excluded.birthdate,
			profession = excluded.profession,
			party = excluded.party,
			party_alias = excluded.party_alias,
			twitter_username = excluded.twitter_username,
			instagram_username = excluded.instagram_username,
			constituency_kind = excluded.constituency_kind,
			constituency_name = excluded.constituency_name,
			constituency_detail = excluded.constituency_detail,
			last_update = excluded.last_update
		RETURNING id`,
		c.Chamber, c.ExternalID, c.FirstName, c.FirstSurname, c.SecondSurname,
		c.ProfilePicture, c.Gender, c.Birthdate, c.Profession, c.Party, c.PartyAlias,
		c.TwitterUsername, c.InstagramUsername, c.ConstituencyKind, c.ConstituencyName,
		c.ConstituencyDetail, formatTime(c.LastUpdate),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("store: upsert congressman: %w", err)
	}
	c.ID = id
	return id, nil
}

// GetCongressman returns the congressman with the given internal ID.
func (s *Store) GetCongressman(ctx context.Context, id int64) (*congress.Congressman, error) {
	return s.scanCongressman(ctx, `SELECT `+congressmanColumns+` FROM congressmen WHERE id = ?`, id)
}

// FindCongressman returns the congressman identified by chamber and external ID.
func (s *Store) FindCongressman(ctx context.Context, chamber congress.Chamber, externalID int) (*congress.Congressman, error) {
	return s.scanCongressman(ctx,
		`SELECT `+congressmanColumns+` FROM congressmen WHERE chamber = ? AND external_id = ?`,
		chamber, externalID)
}

func (s *Store) scanCongressman(ctx context.Context, query string, args ...any) (*congress.Congressman, error) {
	var (
		c          congress.Congressman
		lastUpdate string
	)
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&c.ID, &c.Chamber, &c.ExternalID, &c.FirstName, &c.FirstSurname, &c.SecondSurname,
		&c.ProfilePicture, &c.Gender, &c.Birthdate, &c.Profession, &c.Party, &c.PartyAlias,
		&c.TwitterUsername, &c.InstagramUsername, &c.ConstituencyKind, &c.ConstituencyName,
		&c.ConstituencyDetail, &lastUpdate,
	)
	if err != nil {
		return nil, notFound(err)
	}
	c.LastUpdate = parseTime(lastUpdate)
	return &c, nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
