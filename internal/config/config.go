// Package config holds environment-based configuration.
package config

import "os"

// DefaultDBPath is used when DB_PATH is not set.
const DefaultDBPath = "data/congressmen.db"

// Config is the runtime configuration shared by the binaries.
type Config struct {
	DBPath string
}

// Load reads the configuration from the environment.
func Load() Config {
	return Config{DBPath: getenv("DB_PATH", DefaultDBPath)}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
