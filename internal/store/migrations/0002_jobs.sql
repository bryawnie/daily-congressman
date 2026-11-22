-- +goose Up
-- job_state records which unit of work was completed for which period, so a
-- job that runs on every tick does its work once per period and a run that
-- failed halfway resumes where it stopped. key is "<chamber>:<external id>" for
-- one congressman, or "*" for a whole run.
CREATE TABLE job_state (
    job     TEXT NOT NULL,
    key     TEXT NOT NULL,
    period  TEXT NOT NULL,
    done_at TEXT NOT NULL,
    PRIMARY KEY (job, key, period)
);

-- job_errors is an append-only log of failures, for diagnosis.
CREATE TABLE job_errors (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    at     TEXT NOT NULL,
    job    TEXT NOT NULL,
    key    TEXT NOT NULL,
    period TEXT NOT NULL,
    error  TEXT NOT NULL
);

-- +goose Down
DROP TABLE job_errors;
DROP TABLE job_state;
