-- +goose Up
CREATE TABLE congressmen (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    chamber             TEXT NOT NULL CHECK (chamber IN ('deputies', 'senate')),
    external_id         INTEGER NOT NULL,
    first_name          TEXT NOT NULL DEFAULT '',
    first_surname       TEXT NOT NULL DEFAULT '',
    second_surname      TEXT NOT NULL DEFAULT '',
    profile_picture     TEXT NOT NULL DEFAULT '',
    gender              TEXT NOT NULL DEFAULT '',
    birthdate           TEXT NOT NULL DEFAULT '',
    profession          TEXT NOT NULL DEFAULT '',
    party               TEXT NOT NULL DEFAULT '',
    party_alias         TEXT NOT NULL DEFAULT '',
    twitter_username    TEXT NOT NULL DEFAULT '',
    instagram_username  TEXT NOT NULL DEFAULT '',
    constituency_kind   TEXT NOT NULL DEFAULT '' CHECK (constituency_kind IN ('', 'district', 'region')),
    constituency_name   TEXT NOT NULL DEFAULT '',
    constituency_detail TEXT NOT NULL DEFAULT '',
    last_update         TEXT NOT NULL DEFAULT '',
    UNIQUE (chamber, external_id)
);

CREATE TABLE parliamentary_periods (
    congressman_id INTEGER NOT NULL REFERENCES congressmen(id) ON DELETE CASCADE,
    period_from    INTEGER NOT NULL,
    period_to      INTEGER NOT NULL,
    PRIMARY KEY (congressman_id, period_from, period_to)
);

CREATE TABLE attendance (
    congressman_id     INTEGER PRIMARY KEY REFERENCES congressmen(id) ON DELETE CASCADE,
    total              INTEGER NOT NULL,
    present            INTEGER NOT NULL,
    justified_absent   INTEGER NOT NULL,
    unjustified_absent INTEGER NOT NULL
);

CREATE TABLE votings (
    congressman_id   INTEGER NOT NULL REFERENCES congressmen(id) ON DELETE CASCADE,
    voting_id        INTEGER NOT NULL,
    voting_date      TEXT NOT NULL DEFAULT '',
    bulletin_number  TEXT NOT NULL DEFAULT '',
    document_title   TEXT NOT NULL DEFAULT '',
    article_text     TEXT NOT NULL DEFAULT '',
    voted_option     TEXT NOT NULL DEFAULT '',
    total_approved   INTEGER NOT NULL DEFAULT 0,
    total_rejected   INTEGER NOT NULL DEFAULT 0,
    total_abstention INTEGER NOT NULL DEFAULT 0,
    quorum           TEXT NOT NULL DEFAULT '',
    result           TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (congressman_id, voting_id)
);

CREATE TABLE expenses_operational (
    congressman_id INTEGER NOT NULL REFERENCES congressmen(id) ON DELETE CASCADE,
    year           INTEGER NOT NULL,
    month          INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    expense_type   TEXT NOT NULL,
    amount         INTEGER NOT NULL,
    PRIMARY KEY (congressman_id, year, month, expense_type)
);

CREATE TABLE expenses_offices (
    congressman_id INTEGER NOT NULL REFERENCES congressmen(id) ON DELETE CASCADE,
    year           INTEGER NOT NULL,
    month          INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    offices_number INTEGER NOT NULL,
    total_amount   INTEGER NOT NULL,
    PRIMARY KEY (congressman_id, year, month)
);

CREATE TABLE expenses_support_staff (
    congressman_id INTEGER NOT NULL REFERENCES congressmen(id) ON DELETE CASCADE,
    year           INTEGER NOT NULL,
    month          INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    hired_staff    INTEGER NOT NULL,
    total_amount   INTEGER NOT NULL,
    PRIMARY KEY (congressman_id, year, month)
);

CREATE TABLE law_projects (
    congressman_id INTEGER NOT NULL REFERENCES congressmen(id) ON DELETE CASCADE,
    bulletin       TEXT NOT NULL,
    project_name   TEXT NOT NULL DEFAULT '',
    project_type   TEXT NOT NULL DEFAULT '',
    project_status TEXT NOT NULL DEFAULT '',
    project_date   TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (congressman_id, bulletin)
);

-- One congressman per day. The choice is stored when made and never recomputed.
CREATE TABLE congressman_of_the_day (
    date           TEXT PRIMARY KEY,
    congressman_id INTEGER NOT NULL REFERENCES congressmen(id),
    chain_id       TEXT NOT NULL DEFAULT '',
    pulse_id       TEXT NOT NULL DEFAULT '',
    rand_out       TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_congressman_of_the_day_congressman ON congressman_of_the_day(congressman_id);

-- +goose Down
DROP TABLE congressman_of_the_day;
DROP TABLE law_projects;
DROP TABLE expenses_support_staff;
DROP TABLE expenses_offices;
DROP TABLE expenses_operational;
DROP TABLE votings;
DROP TABLE attendance;
DROP TABLE parliamentary_periods;
DROP TABLE congressmen;
