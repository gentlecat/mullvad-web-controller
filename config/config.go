package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Addr   string // Address to listen on
	DryRun bool   // Skip calling the Mullvad CLI when changing relays
}

// Load reads the configuration from environment variables, falling back to defaults.
func Load() (Config, error) {
	cfg := Config{
		Addr: getEnv("MULLVAD_WEB_ADDR", "0.0.0.0:8666"),
	}

	if value := os.Getenv("MULLVAD_WEB_DRY_RUN"); value != "" {
		dryRun, err := strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid MULLVAD_WEB_DRY_RUN value %q: %w", value, err)
		}
		cfg.DryRun = dryRun
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
