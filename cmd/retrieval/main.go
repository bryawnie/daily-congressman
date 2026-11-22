// Command retrieval populates the database with congressmen data.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/bryawnie/daily-congressman/internal/config"
	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/retrieval"
	"github.com/bryawnie/daily-congressman/internal/source"
	"github.com/bryawnie/daily-congressman/internal/source/deputies"
	"github.com/bryawnie/daily-congressman/internal/source/senate"
	"github.com/bryawnie/daily-congressman/internal/store"
)

const usage = `usage:
  retrieval migrate
  retrieval fetch  <deputies|senate> <external id>   print everything the source gives for a congressman
  retrieval update <deputies|senate> <external id>   save it in the database (DB_PATH)`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "migrate":
		err = migrate()
	case "fetch", "update":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(2)
		}
		if os.Args[1] == "fetch" {
			err = fetch(os.Args[2], os.Args[3])
		} else {
			err = update(os.Args[2], os.Args[3])
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n%s\n", os.Args[1], usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// migrate creates the database file if needed and applies pending migrations.
func migrate() error {
	cfg := config.Load()
	s, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer s.Close()
	if err := s.Migrate(context.Background()); err != nil {
		return err
	}
	fmt.Println("database ready:", cfg.DBPath)
	return nil
}

// section is the outcome of asking a source for one kind of data.
type section struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func run[T any](f func() (T, error)) section {
	v, err := f()
	switch {
	case errors.Is(err, congress.ErrNotSupported):
		return section{Error: "not supported by this source yet"}
	case err != nil:
		return section{Error: err.Error()}
	}
	return section{Data: v}
}

// fetch prints, as JSON, every kind of data the chamber's source provides for
// one congressman. It does not touch the database.
func fetch(chamber, id string) error {
	extID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("bad id %q: %w", id, err)
	}
	reg := source.NewRegistry(deputies.New(), senate.New())
	src, err := reg.Get(congress.Chamber(chamber))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	m := congress.Member{Chamber: src.Chamber(), ExternalID: extID}

	out := map[string]section{
		"profile":    run(func() (congress.Congressman, error) { return src.Profile(ctx, m) }),
		"attendance": run(func() (congress.Attendance, error) { return src.Attendance(ctx, m) }),
		"votings":    run(func() ([]congress.Voting, error) { return src.Votings(ctx, m, 10) }),
		"activity":   run(func() ([]congress.LawProject, error) { return src.Activity(ctx, m) }),
		"expenses":   run(func() (congress.Expenses, error) { return src.Expenses(ctx, m) }),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// update reads everything the chamber's source gives for one congressman and
// saves it. The profile is required, because every other table hangs from it;
// if another kind of data fails, the rest is still saved and the failures are
// reported at the end.
func update(chamber, id string) error {
	extID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("bad id %q: %w", id, err)
	}
	reg := source.NewRegistry(deputies.New(), senate.New())
	src, err := reg.Get(congress.Chamber(chamber))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cfg := config.Load()
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		return err
	}

	sy := &retrieval.Syncer{Store: st, Sources: reg}
	m := congress.Member{Chamber: src.Chamber(), ExternalID: extID}
	cid, profile, err := sy.Profile(ctx, m)
	if err != nil {
		return err
	}
	fmt.Printf("congressman %d: %s %s %s (%s, id %d)\n", cid, profile.FirstName, profile.FirstSurname, profile.SecondSurname, chamber, extID)

	daily, err := sy.Daily(ctx, cid, m)
	if err != nil {
		return err
	}
	expenses, err := sy.Expenses(ctx, cid, m)
	if err != nil {
		return err
	}
	report := append(daily, expenses...)
	for _, s := range report {
		switch s.Status {
		case retrieval.Pending:
			fmt.Printf("  %-11s not supported by this source yet\n", s.Name)
		case retrieval.Failed:
			fmt.Printf("  %-11s FAILED: %v\n", s.Name, s.Err)
		default:
			fmt.Printf("  %-11s %s\n", s.Name, s.Detail)
		}
	}
	if failed := report.Failed(); len(failed) > 0 {
		return fmt.Errorf("%d kind(s) of data failed: %w", len(failed), report.Err())
	}
	return nil
}
