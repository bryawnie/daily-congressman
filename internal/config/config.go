// Package config holds environment-based configuration.
package config

import (
	"os"
	"strings"
	"time"
)

const (
	// DefaultDBPath is used when DB_PATH is not set.
	DefaultDBPath = "data/congressmen.db"
	// DefaultChambers is used when CHOOSER_CHAMBERS is not set.
	DefaultChambers = "senate"
	// DefaultJobsInterval is how often the jobs service checks what is due.
	DefaultJobsInterval = 15 * time.Minute
	// DefaultJobsPause is the wait between two congressmen in a monthly sync.
	DefaultJobsPause = 500 * time.Millisecond
	// DefaultAPIAddr is where the API listens when API_ADDR is not set.
	DefaultAPIAddr = ":8080"
	// DefaultCORSOrigins is used when API_CORS_ORIGINS is not set: the API is
	// public and read-only.
	DefaultCORSOrigins = "*"
)

// Config is the runtime configuration shared by the binaries.
type Config struct {
	DBPath string
	// Chambers lists the chambers that take part in the daily draw.
	Chambers []string
	// JobsInterval is how often `jobs serve` checks what is due (JOBS_INTERVAL,
	// a Go duration such as 15m).
	JobsInterval time.Duration
	// JobsPause is the wait between congressmen in a monthly sync (JOBS_PAUSE).
	JobsPause time.Duration
	// APIAddr is the address the API listens on (API_ADDR).
	APIAddr string
	// CORSOrigins are the origins allowed to call the API (API_CORS_ORIGINS,
	// comma separated; "*" allows any).
	CORSOrigins []string
}

// Load reads the configuration from the environment.
func Load() Config {
	return Config{
		DBPath:   getenv("DB_PATH", DefaultDBPath),
		Chambers: splitList(getenv("CHOOSER_CHAMBERS", DefaultChambers)),

		JobsInterval: getduration("JOBS_INTERVAL", DefaultJobsInterval),
		JobsPause:    getduration("JOBS_PAUSE", DefaultJobsPause),

		APIAddr:     getenv("API_ADDR", DefaultAPIAddr),
		CORSOrigins: splitList(getenv("API_CORS_ORIGINS", DefaultCORSOrigins)),
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

// getduration reads a Go duration; an unset or invalid value gives the fallback.
func getduration(key string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(os.Getenv(key)); err == nil && d >= 0 {
		return d
	}
	return fallback
}
