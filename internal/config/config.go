// Package config holds environment-based configuration.
package config

import (
	"os"
	"strings"
)

const (
	// DefaultDBPath is used when DB_PATH is not set.
	DefaultDBPath = "data/congressmen.db"
	// DefaultChambers is used when CHOOSER_CHAMBERS is not set. The Senate joins
	// the draw once its source is implemented.
	DefaultChambers = "deputies"
)

// Config is the runtime configuration shared by the binaries.
type Config struct {
	DBPath string
	// Chambers lists the chambers that take part in the daily draw.
	Chambers []string
}

// Load reads the configuration from the environment.
func Load() Config {
	return Config{
		DBPath:   getenv("DB_PATH", DefaultDBPath),
		Chambers: splitList(getenv("CHOOSER_CHAMBERS", DefaultChambers)),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
