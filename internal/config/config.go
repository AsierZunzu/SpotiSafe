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
	CallbackPort   string
	PublicURL      string // base URL served to the user; /callback is appended for the redirect URI
	OutputDir      string
	TokenFile      string
	LogLevel       string
	Schedule       string // cron expression; empty means run once and exit
	SkipInitialRun bool   // when true, skip the immediate run at startup
}

// RedirectURI returns the OAuth redirect URI derived from PublicURL.
func (c *Config) RedirectURI() string {
	return c.PublicURL + "/callback"
}

// Load reads configuration from environment variables, applies defaults, and validates.
func Load() (*Config, error) {
	port := getEnvDefault("SPOTIFY_CALLBACK_PORT", "8888")
	cfg := &Config{
		ClientID:       os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret:   os.Getenv("SPOTIFY_CLIENT_SECRET"),
		CallbackPort:   port,
		PublicURL:      getEnvDefault("SPOTIFY_PUBLIC_URL", "http://localhost:"+port),
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
