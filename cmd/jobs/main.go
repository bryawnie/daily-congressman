// Command jobs runs the scheduled work: the monthly synchronization of every
// congressman's profile and expenses, and the daily draw and refresh.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bryawnie/daily-congressman/internal/beacon"
	"github.com/bryawnie/daily-congressman/internal/chooser"
	"github.com/bryawnie/daily-congressman/internal/config"
	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/jobs"
	"github.com/bryawnie/daily-congressman/internal/retrieval"
	"github.com/bryawnie/daily-congressman/internal/source"
	"github.com/bryawnie/daily-congressman/internal/source/deputies"
	"github.com/bryawnie/daily-congressman/internal/source/senate"
	"github.com/bryawnie/daily-congressman/internal/store"
)

const usage = `usage:
  jobs serve                      keep running; every JOBS_INTERVAL do whatever is due
  jobs run                        do whatever is due once, then exit (for an external cron)
  jobs daily    [-date D] [-force]  draw the congressman of the day and refresh their data
                                  (-force refreshes again; a drawn day is never redrawn)
  jobs profiles [-force]          sync the profile of every congressman (monthly)
  jobs expenses [-force]          sync the expenses of every congressman (monthly)

Everything is idempotent: work already done in its period (day or month) is
skipped. -force does it again.`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]

	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	force := fs.Bool("force", false, "do the work even if already done in this period")
	day := fs.String("date", "", "day to draw for (YYYY-MM-DD, Chile time); default today")
	switch cmd {
	case "serve", "run", "daily", "profiles", "expenses":
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n%s\n", cmd, usage)
		os.Exit(2)
	}
	fs.Parse(args)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cmd, *day, *force); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd, day string, force bool) error {
	cfg := config.Load()
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		return err
	}

	var chambers []congress.Chamber
	for _, name := range cfg.Chambers {
		chambers = append(chambers, congress.Chamber(name))
	}
	active, err := source.NewRegistry(deputies.New(), senate.New()).Only(chambers...)
	if err != nil {
		return err
	}
	j := &jobs.Jobs{
		Store:   st,
		Sources: active,
		Chooser: &chooser.Chooser{Pulses: beacon.NewClient(), Sources: active},
		Syncer:  &retrieval.Syncer{Store: st, Sources: active},
		Pause:   cfg.JobsPause,
		Log:     slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}

	switch cmd {
	case "serve":
		j.Log.Info("jobs service started", "chambers", cfg.Chambers, "interval", cfg.JobsInterval)
		return j.Serve(ctx, cfg.JobsInterval)
	case "run":
		return j.RunDue(ctx)
	case "daily":
		if day == "" {
			day = j.Today()
		}
		return j.Daily(ctx, day, force)
	case "profiles":
		return j.SyncProfiles(ctx, force)
	default: // expenses
		return j.SyncExpenses(ctx, force)
	}
}
