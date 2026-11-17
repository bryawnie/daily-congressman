// Command chooser prints the congressman chosen for a day.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/bryawnie/daily-congressman/internal/beacon"
	"github.com/bryawnie/daily-congressman/internal/chooser"
	"github.com/bryawnie/daily-congressman/internal/config"
	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/source"
	"github.com/bryawnie/daily-congressman/internal/source/deputies"
	"github.com/bryawnie/daily-congressman/internal/source/senate"
)

func main() {
	day := flag.String("date", chooser.Today(time.Now()), "day to choose for (YYYY-MM-DD, Chile time)")
	flag.Parse()

	if err := run(*day); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(day string) error {
	cfg := config.Load()

	all := source.NewRegistry(deputies.New(), senate.New())
	var chambers []congress.Chamber
	for _, name := range cfg.Chambers {
		chambers = append(chambers, congress.Chamber(name))
	}
	active, err := all.Only(chambers...)
	if err != nil {
		return err
	}

	bc := beacon.NewClient()
	res, err := (&chooser.Chooser{Pulses: bc, Sources: active}).Choose(context.Background(), day)
	if err != nil {
		return err
	}

	fmt.Printf("Day:      %s\n", res.Day)
	fmt.Printf("Chamber:  %s\n", res.Member.Chamber)
	fmt.Printf("Member:   %s (id %d)\n", res.Member.Name, res.Member.ExternalID)
	fmt.Printf("Pulse:    %s\n", res.Pulse.URI(bc.BaseURL()))
	return nil
}
