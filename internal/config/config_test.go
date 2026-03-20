package config

import "testing"

func TestLoad_MissingClientID(t *testing.T) {
	t.Setenv("SPOTIFY_CLIENT_ID", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when SPOTIFY_CLIENT_ID is not set")
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("SPOTIFY_CLIENT_ID", "test-id")
	t.Setenv("SPOTIFY_REDIRECT_URI", "")
	t.Setenv("SPOTIFY_CALLBACK_PORT", "")
	t.Setenv("SPOTIFY_OUTPUT_DIR", "")
	t.Setenv("SPOTIFY_TOKEN_FILE", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := []struct{ got, want, name string }{
		{cfg.RedirectURI, "http://localhost:8888/callback", "RedirectURI"},
		{cfg.CallbackPort, "8888", "CallbackPort"},
		{cfg.OutputDir, "./output", "OutputDir"},
		{cfg.TokenFile, "./output/.spotify_token.json", "TokenFile"},
		{cfg.LogLevel, "info", "LogLevel"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestLoad_ValidSchedule(t *testing.T) {
	t.Setenv("SPOTIFY_CLIENT_ID", "test-id")
	t.Setenv("SPOTIFY_SCHEDULE", "0 2 * * *")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Schedule != "0 2 * * *" {
		t.Errorf("Schedule = %q, want '0 2 * * *'", cfg.Schedule)
	}
}

func TestLoad_InvalidSchedule(t *testing.T) {
	t.Setenv("SPOTIFY_CLIENT_ID", "test-id")
	t.Setenv("SPOTIFY_SCHEDULE", "not-a-cron")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

func TestLoad_SkipInitialRun(t *testing.T) {
	t.Setenv("SPOTIFY_CLIENT_ID", "test-id")
	t.Setenv("SPOTIFY_SKIP_INITIAL_RUN", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.SkipInitialRun {
		t.Error("SkipInitialRun should be true")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("SPOTIFY_CLIENT_ID", "my-client")
	t.Setenv("SPOTIFY_CLIENT_SECRET", "my-secret")
	t.Setenv("SPOTIFY_REDIRECT_URI", "http://localhost:9999/callback")
	t.Setenv("SPOTIFY_CALLBACK_PORT", "9999")
	t.Setenv("SPOTIFY_OUTPUT_DIR", "/tmp/backup")
	t.Setenv("SPOTIFY_TOKEN_FILE", "/tmp/token.json")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := []struct{ got, want, name string }{
		{cfg.ClientID, "my-client", "ClientID"},
		{cfg.ClientSecret, "my-secret", "ClientSecret"},
		{cfg.RedirectURI, "http://localhost:9999/callback", "RedirectURI"},
		{cfg.CallbackPort, "9999", "CallbackPort"},
		{cfg.OutputDir, "/tmp/backup", "OutputDir"},
		{cfg.TokenFile, "/tmp/token.json", "TokenFile"},
		{cfg.LogLevel, "debug", "LogLevel"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}
