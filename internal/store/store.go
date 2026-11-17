// Package store handles SQLite persistence, migrations and repositories.
package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite" // pure Go SQLite driver
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// ErrNotFound is returned when a requested row does not exist.
var ErrNotFound = errors.New("store: not found")

// Store gives access to the SQLite database.
type Store struct {
	db *sql.DB
}

// Open opens (creating it if needed) the SQLite database at path. It enables
// WAL, a busy timeout and foreign keys so the API and the jobs can share the
// file safely. Call Migrate to bring the schema up to date.
func Open(path string) (*Store, error) {
	q := url.Values{}
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "journal_mode(WAL)")
	q.Add("_pragma", "foreign_keys(1)")
	q.Set("_txlock", "immediate")
	dsn := (&url.URL{Scheme: "file", Opaque: path, RawQuery: q.Encode()}).String()

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}
	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// Migrate applies all pending migrations.
func (s *Store) Migrate(ctx context.Context) error {
	fsys, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectSQLite3, s.db, fsys)
	if err != nil {
		return fmt.Errorf("store: migrations: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("store: migrate: %w", err)
	}
	return nil
}

// inTx runs fn in a transaction, committing if it returns nil.
func (s *Store) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func notFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
