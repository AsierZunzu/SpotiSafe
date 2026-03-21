package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/asierzunzu/spotisafe/internal/auth"
	"github.com/asierzunzu/spotisafe/internal/backup"
	"github.com/asierzunzu/spotisafe/internal/config"
	"github.com/asierzunzu/spotisafe/internal/retention"
	"github.com/asierzunzu/spotisafe/internal/spotify"
	"github.com/asierzunzu/spotisafe/internal/writer"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "err", err)
		return 1
	}

	setupLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpClient, err := auth.GetClient(ctx, cfg.ClientID, cfg.CallbackPort, cfg.PublicURL, cfg.TokenFile)
	if err != nil {
		slog.Error("authentication failed", "err", err)
		return 1
	}

	spotifyClient := spotify.New(httpClient)

	profile, err := spotifyClient.GetProfile(ctx)
	if err != nil {
		slog.Error("failed to fetch user profile", "err", err)
		return 1
	}
	slog.Info("authenticated", "user", profile.DisplayName, "id", profile.ID)

	// No schedule: run once and exit.
	if cfg.Schedule == "" {
		return runBackup(ctx, cfg, spotifyClient, profile.ID)
	}

	slog.Info("running on schedule", "cron", cfg.Schedule)

	c := cron.New()
	if _, err := c.AddFunc(cfg.Schedule, func() {
		runBackup(ctx, cfg, spotifyClient, profile.ID)
	}); err != nil {
		slog.Error("invalid schedule", "err", err)
		return 1
	}
	c.Start()
	defer c.Stop()

	if !cfg.SkipInitialRun {
		runBackup(ctx, cfg, spotifyClient, profile.ID)
	}

	<-ctx.Done()
	slog.Info("shutting down")
	return 0
}

func runBackup(ctx context.Context, cfg *config.Config, spotifyClient *spotify.Client, userID string) int {
	w, err := writer.New(cfg.OutputDir, userID)
	if err != nil {
		slog.Error("failed to create output directory", "err", err)
		return 1
	}
	slog.Info("output directory", "path", w.RunDir())

	orch := &backup.Orchestrator{
		Jobs: []backup.Job{
			&backup.ProfileJob{Client: spotifyClient, Writer: w},
			&backup.SavedTracksJob{Client: spotifyClient, Writer: w},
			&backup.SavedAlbumsJob{Client: spotifyClient, Writer: w},
			&backup.SavedEpisodesJob{Client: spotifyClient, Writer: w},
			&backup.SavedShowsJob{Client: spotifyClient, Writer: w},
			&backup.FollowedArtistsJob{Client: spotifyClient, Writer: w},
			&backup.PlaylistsJob{Client: spotifyClient, Writer: w},
			&backup.TopArtistsJob{Client: spotifyClient, Writer: w},
			&backup.TopTracksJob{Client: spotifyClient, Writer: w},
			&backup.RecentlyPlayedJob{Client: spotifyClient, Writer: w},
			&backup.AudiobooksJob{Client: spotifyClient, Writer: w},
		},
	}

	results := orch.Run(ctx)
	code := backup.PrintSummary(results)

	policy := retention.Policy{
		KeepLast:      cfg.RetentionKeepLast,
		KeepLastDay:   cfg.RetentionKeepLastDay,
		KeepLastWeek:  cfg.RetentionKeepLastWeek,
		KeepLastMonth: cfg.RetentionKeepLastMonth,
		KeepLastYear:  cfg.RetentionKeepLastYear,
	}
	if err := retention.Apply(cfg.OutputDir, policy, time.Now().UTC()); err != nil {
		slog.Warn("retention policy failed", "err", err)
	}

	return code
}

func setupLogger(level string) {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(handler))
}
