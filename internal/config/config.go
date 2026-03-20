package config

import (
	"errors"
	"os"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	CallbackPort string
	OutputDir    string
	TokenFile    string
	LogLevel     string
}

// Load reads configuration from environment variables, applies defaults, and validates.
func Load() (*Config, error) {
	cfg := &Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		RedirectURI:  getEnvDefault("SPOTIFY_REDIRECT_URI", "http://localhost:8888/callback"),
		CallbackPort: getEnvDefault("SPOTIFY_CALLBACK_PORT", "8888"),
		OutputDir:    getEnvDefault("SPOTIFY_OUTPUT_DIR", "./output"),
		TokenFile:    getEnvDefault("SPOTIFY_TOKEN_FILE", "./output/.spotify_token.json"),
		LogLevel:     getEnvDefault("LOG_LEVEL", "info"),
	}

	if cfg.ClientID == "" {
		return nil, errors.New("SPOTIFY_CLIENT_ID is required")
	}

	return cfg, nil
}

func getEnvDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
