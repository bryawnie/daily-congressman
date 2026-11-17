// Command retrieval populates the database with congressmen data.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bryawnie/daily-congressman/internal/config"
	"github.com/bryawnie/daily-congressman/internal/store"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: retrieval migrate")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "migrate":
		if err := migrate(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(2)
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
