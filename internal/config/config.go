package config

import (
	"errors"
	"os"

	"github.com/robfig/cron/v3"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	ClientID       string
	ClientSecret   string
	RedirectURI    string
	CallbackPort   string
	OutputDir      string
	TokenFile      string
	LogLevel       string
	Schedule       string // cron expression; empty means run once and exit
	SkipInitialRun bool   // when true, skip the immediate run at startup
}

// Load reads configuration from environment variables, applies defaults, and validates.
func Load() (*Config, error) {
	cfg := &Config{
		ClientID:       os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret:   os.Getenv("SPOTIFY_CLIENT_SECRET"),
		RedirectURI:    getEnvDefault("SPOTIFY_REDIRECT_URI", "http://localhost:8888/callback"),
		CallbackPort:   getEnvDefault("SPOTIFY_CALLBACK_PORT", "8888"),
		OutputDir:      getEnvDefault("SPOTIFY_OUTPUT_DIR", "./output"),
		TokenFile:      getEnvDefault("SPOTIFY_TOKEN_FILE", "./output/.spotify_token.json"),
		LogLevel:       getEnvDefault("LOG_LEVEL", "info"),
		Schedule:       os.Getenv("SPOTIFY_SCHEDULE"),
		SkipInitialRun: os.Getenv("SPOTIFY_SKIP_INITIAL_RUN") == "true",
	}

	if cfg.ClientID == "" {
		return nil, errors.New("SPOTIFY_CLIENT_ID is required")
	}

	if cfg.Schedule != "" {
		if _, err := cron.ParseStandard(cfg.Schedule); err != nil {
			return nil, errors.New("invalid SPOTIFY_SCHEDULE: " + err.Error())
		}
	}

	return cfg, nil
}

func getEnvDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
