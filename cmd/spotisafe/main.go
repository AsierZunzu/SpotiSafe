package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/yourusername/spotisafe/internal/auth"
	"github.com/yourusername/spotisafe/internal/backup"
	"github.com/yourusername/spotisafe/internal/config"
	"github.com/yourusername/spotisafe/internal/spotify"
	"github.com/yourusername/spotisafe/internal/writer"
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

	ctx := context.Background()

	httpClient, err := auth.GetClient(ctx, cfg.ClientID, cfg.RedirectURI, cfg.CallbackPort, cfg.TokenFile)
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

	w, err := writer.New(cfg.OutputDir, profile.ID)
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
	return backup.PrintSummary(results)
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
