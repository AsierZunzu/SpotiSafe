package backup

import (
	"context"
	"fmt"

	"github.com/yourusername/spotisafe/internal/spotify"
	"github.com/yourusername/spotisafe/internal/writer"
)

// ProfileJob backs up the user's Spotify profile.
type ProfileJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *ProfileJob) Name() string { return "profile" }
func (j *ProfileJob) Run(ctx context.Context) (int, error) {
	profile, err := j.Client.GetProfile(ctx)
	if err != nil {
		return 0, err
	}
	return 1, j.Writer.Write("profile", "profile", 1, profile)
}

// SavedTracksJob backs up all saved tracks.
type SavedTracksJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *SavedTracksJob) Name() string { return "saved_tracks" }
func (j *SavedTracksJob) Run(ctx context.Context) (int, error) {
	items, err := j.Client.GetSavedTracks(ctx)
	if err != nil {
		return 0, err
	}
	return len(items), j.Writer.Write("saved_tracks", "saved_tracks", len(items), items)
}

// SavedAlbumsJob backs up all saved albums.
type SavedAlbumsJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *SavedAlbumsJob) Name() string { return "saved_albums" }
func (j *SavedAlbumsJob) Run(ctx context.Context) (int, error) {
	items, err := j.Client.GetSavedAlbums(ctx)
	if err != nil {
		return 0, err
	}
	return len(items), j.Writer.Write("saved_albums", "saved_albums", len(items), items)
}

// SavedEpisodesJob backs up all saved podcast episodes.
type SavedEpisodesJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *SavedEpisodesJob) Name() string { return "saved_episodes" }
func (j *SavedEpisodesJob) Run(ctx context.Context) (int, error) {
	items, err := j.Client.GetSavedEpisodes(ctx)
	if err != nil {
		return 0, err
	}
	return len(items), j.Writer.Write("saved_episodes", "saved_episodes", len(items), items)
}

// SavedShowsJob backs up all saved podcast shows.
type SavedShowsJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *SavedShowsJob) Name() string { return "saved_shows" }
func (j *SavedShowsJob) Run(ctx context.Context) (int, error) {
	items, err := j.Client.GetSavedShows(ctx)
	if err != nil {
		return 0, err
	}
	return len(items), j.Writer.Write("saved_shows", "saved_shows", len(items), items)
}

// FollowedArtistsJob backs up all followed artists.
type FollowedArtistsJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *FollowedArtistsJob) Name() string { return "followed_artists" }
func (j *FollowedArtistsJob) Run(ctx context.Context) (int, error) {
	items, err := j.Client.GetFollowedArtists(ctx)
	if err != nil {
		return 0, err
	}
	return len(items), j.Writer.Write("followed_artists", "followed_artists", len(items), items)
}

// PlaylistsJob backs up the playlist list and each playlist's tracks.
type PlaylistsJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *PlaylistsJob) Name() string { return "playlists" }
func (j *PlaylistsJob) Run(ctx context.Context) (int, error) {
	playlists, err := j.Client.GetPlaylists(ctx)
	if err != nil {
		return 0, err
	}

	if err := j.Writer.Write("playlists", "playlists", len(playlists), playlists); err != nil {
		return 0, err
	}

	for _, pl := range playlists {
		tracks, err := j.Client.GetPlaylistTracks(ctx, pl.ID)
		if err != nil {
			// Soft-fail individual playlists — log but continue
			fmt.Printf("  [WARN] could not fetch tracks for playlist %q (%s): %v\n", pl.Name, pl.ID, err)
			continue
		}
		filename := fmt.Sprintf("%s_%s", pl.ID, writer.SanitizeName(pl.Name))
		category := fmt.Sprintf("playlist_tracks:%s", pl.ID)
		if err := j.Writer.WriteToSubdir("playlist_tracks", filename, category, len(tracks), tracks); err != nil {
			fmt.Printf("  [WARN] could not write tracks for playlist %q: %v\n", pl.Name, err)
		}
	}

	return len(playlists), nil
}

// TopArtistsJob backs up top artists for all three time ranges.
type TopArtistsJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *TopArtistsJob) Name() string { return "top_artists" }
func (j *TopArtistsJob) Run(ctx context.Context) (int, error) {
	total := 0
	for _, tr := range []string{"short_term", "medium_term", "long_term"} {
		items, err := j.Client.GetTopArtists(ctx, tr)
		if err != nil {
			return total, fmt.Errorf("time range %s: %w", tr, err)
		}
		filename := "top_artists_" + tr
		if err := j.Writer.Write(filename, filename, len(items), items); err != nil {
			return total, err
		}
		total += len(items)
	}
	return total, nil
}

// TopTracksJob backs up top tracks for all three time ranges.
type TopTracksJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *TopTracksJob) Name() string { return "top_tracks" }
func (j *TopTracksJob) Run(ctx context.Context) (int, error) {
	total := 0
	for _, tr := range []string{"short_term", "medium_term", "long_term"} {
		items, err := j.Client.GetTopTracks(ctx, tr)
		if err != nil {
			return total, fmt.Errorf("time range %s: %w", tr, err)
		}
		filename := "top_tracks_" + tr
		if err := j.Writer.Write(filename, filename, len(items), items); err != nil {
			return total, err
		}
		total += len(items)
	}
	return total, nil
}

// RecentlyPlayedJob backs up recently played tracks.
type RecentlyPlayedJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *RecentlyPlayedJob) Name() string { return "recently_played" }
func (j *RecentlyPlayedJob) Run(ctx context.Context) (int, error) {
	items, err := j.Client.GetRecentlyPlayed(ctx)
	if err != nil {
		return 0, err
	}
	return len(items), j.Writer.Write("recently_played", "recently_played", len(items), items)
}

// AudiobooksJob backs up all saved audiobooks.
type AudiobooksJob struct {
	Client *spotify.Client
	Writer *writer.JSONWriter
}

func (j *AudiobooksJob) Name() string { return "audiobooks" }
func (j *AudiobooksJob) Run(ctx context.Context) (int, error) {
	items, err := j.Client.GetAudiobooks(ctx)
	if err != nil {
		return 0, err
	}
	return len(items), j.Writer.Write("audiobooks", "audiobooks", len(items), items)
}
