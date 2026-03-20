package backup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/spotisafe/internal/spotify"
	"github.com/yourusername/spotisafe/internal/writer"
)

// redirectTransport rewrites all requests to a test server.
type redirectTransport struct{ target string }

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u, _ := url.Parse(t.target)
	req = req.Clone(req.Context())
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
	return http.DefaultTransport.RoundTrip(req)
}

func newJobFixture(t *testing.T, handler http.Handler) (*spotify.Client, *writer.JSONWriter) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := spotify.New(&http.Client{Transport: &redirectTransport{target: srv.URL}})
	w, err := writer.New(t.TempDir(), "user1")
	if err != nil {
		t.Fatalf("writer.New: %v", err)
	}
	return client, w
}

func jsonResponse(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func errorResponse(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func outputFileExists(t *testing.T, w *writer.JSONWriter, filename string) {
	t.Helper()
	path := filepath.Join(w.RunDir(), filename+".json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected output file %s: %v", path, err)
	}
}

// --- ProfileJob ---

func TestProfileJob_Success(t *testing.T) {
	c, w := newJobFixture(t, jsonResponse(spotify.UserProfile{ID: "u1", DisplayName: "Alice"}))
	count, err := (&ProfileJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	outputFileExists(t, w, "profile")
}

func TestProfileJob_Error(t *testing.T) {
	c, w := newJobFixture(t, http.HandlerFunc(errorResponse))
	if _, err := (&ProfileJob{Client: c, Writer: w}).Run(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

// --- SavedTracksJob ---

func TestSavedTracksJob_Success(t *testing.T) {
	page := spotify.OffsetPage[spotify.SavedTrack]{
		Items: []spotify.SavedTrack{{Track: spotify.Track{ID: "t1"}}, {Track: spotify.Track{ID: "t2"}}},
	}
	c, w := newJobFixture(t, jsonResponse(page))
	count, err := (&SavedTracksJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil || count != 2 {
		t.Fatalf("err=%v, count=%d", err, count)
	}
	outputFileExists(t, w, "saved_tracks")
}

func TestSavedTracksJob_Error(t *testing.T) {
	c, w := newJobFixture(t, http.HandlerFunc(errorResponse))
	if _, err := (&SavedTracksJob{Client: c, Writer: w}).Run(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

// --- SavedAlbumsJob ---

func TestSavedAlbumsJob_Success(t *testing.T) {
	page := spotify.OffsetPage[spotify.SavedAlbum]{Items: []spotify.SavedAlbum{{Album: spotify.Album{ID: "a1"}}}}
	c, w := newJobFixture(t, jsonResponse(page))
	count, err := (&SavedAlbumsJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("err=%v, count=%d", err, count)
	}
	outputFileExists(t, w, "saved_albums")
}

// --- SavedEpisodesJob ---

func TestSavedEpisodesJob_Success(t *testing.T) {
	page := spotify.OffsetPage[spotify.SavedEpisode]{Items: []spotify.SavedEpisode{{Episode: spotify.Episode{ID: "e1"}}}}
	c, w := newJobFixture(t, jsonResponse(page))
	count, err := (&SavedEpisodesJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("err=%v, count=%d", err, count)
	}
	outputFileExists(t, w, "saved_episodes")
}

// --- SavedShowsJob ---

func TestSavedShowsJob_Success(t *testing.T) {
	page := spotify.OffsetPage[spotify.SavedShow]{Items: []spotify.SavedShow{{Show: spotify.Show{ID: "s1"}}}}
	c, w := newJobFixture(t, jsonResponse(page))
	count, err := (&SavedShowsJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("err=%v, count=%d", err, count)
	}
	outputFileExists(t, w, "saved_shows")
}

// --- FollowedArtistsJob ---

func TestFollowedArtistsJob_Success(t *testing.T) {
	resp := spotify.FollowedArtistsResponse{
		Artists: spotify.CursorPage[spotify.Artist]{Items: []spotify.Artist{{ID: "a1"}}},
	}
	c, w := newJobFixture(t, jsonResponse(resp))
	count, err := (&FollowedArtistsJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("err=%v, count=%d", err, count)
	}
	outputFileExists(t, w, "followed_artists")
}

// --- PlaylistsJob ---

func TestPlaylistsJob_Success(t *testing.T) {
	var call int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		switch {
		case r.URL.Path == "/v1/me/playlists":
			_ = json.NewEncoder(w).Encode(spotify.OffsetPage[spotify.PlaylistSummary]{
				Items: []spotify.PlaylistSummary{{ID: "pl1", Name: "My Playlist"}},
			})
		default:
			// playlist tracks
			_ = json.NewEncoder(w).Encode(spotify.OffsetPage[spotify.PlaylistTrack]{
				Items: []spotify.PlaylistTrack{{Track: spotify.Track{ID: "t1"}}},
			})
		}
	})

	c, w := newJobFixture(t, handler)
	count, err := (&PlaylistsJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("err=%v, count=%d", err, count)
	}
	outputFileExists(t, w, "playlists")
}

func TestPlaylistsJob_TrackFetchSoftFail(t *testing.T) {
	// Playlists succeed but track fetch fails — job should still succeed.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/me/playlists" {
			_ = json.NewEncoder(w).Encode(spotify.OffsetPage[spotify.PlaylistSummary]{
				Items: []spotify.PlaylistSummary{{ID: "pl1", Name: "My Playlist"}},
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	})

	c, w := newJobFixture(t, handler)
	count, err := (&PlaylistsJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("expected success despite track error: err=%v, count=%d", err, count)
	}
}

// --- TopArtistsJob ---

func TestTopArtistsJob_Success(t *testing.T) {
	page := spotify.OffsetPage[spotify.TopArtist]{Items: []spotify.TopArtist{{ID: "a1"}}}
	c, w := newJobFixture(t, jsonResponse(page))
	count, err := (&TopArtistsJob{Client: c, Writer: w}).Run(context.Background())
	// 3 time ranges × 1 item each
	if err != nil || count != 3 {
		t.Fatalf("err=%v, count=%d", err, count)
	}
	for _, tr := range []string{"short_term", "medium_term", "long_term"} {
		outputFileExists(t, w, "top_artists_"+tr)
	}
}

func TestTopArtistsJob_Error(t *testing.T) {
	c, w := newJobFixture(t, http.HandlerFunc(errorResponse))
	if _, err := (&TopArtistsJob{Client: c, Writer: w}).Run(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

// --- TopTracksJob ---

func TestTopTracksJob_Success(t *testing.T) {
	page := spotify.OffsetPage[spotify.TopTrack]{Items: []spotify.TopTrack{{ID: "t1"}}}
	c, w := newJobFixture(t, jsonResponse(page))
	count, err := (&TopTracksJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil || count != 3 {
		t.Fatalf("err=%v, count=%d", err, count)
	}
	for _, tr := range []string{"short_term", "medium_term", "long_term"} {
		outputFileExists(t, w, "top_tracks_"+tr)
	}
}

// --- RecentlyPlayedJob ---

func TestRecentlyPlayedJob_Success(t *testing.T) {
	resp := spotify.CursorPage[spotify.RecentlyPlayedItem]{
		Items: []spotify.RecentlyPlayedItem{{Track: spotify.Track{ID: "t1"}}},
	}
	c, w := newJobFixture(t, jsonResponse(resp))
	count, err := (&RecentlyPlayedJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("err=%v, count=%d", err, count)
	}
	outputFileExists(t, w, "recently_played")
}

// --- AudiobooksJob ---

func TestAudiobooksJob_Success(t *testing.T) {
	page := spotify.OffsetPage[spotify.SavedAudiobook]{Items: []spotify.SavedAudiobook{{ID: "b1"}}}
	c, w := newJobFixture(t, jsonResponse(page))
	count, err := (&AudiobooksJob{Client: c, Writer: w}).Run(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("err=%v, count=%d", err, count)
	}
	outputFileExists(t, w, "audiobooks")
}

// --- Name() ---

func TestJobNames(t *testing.T) {
	c, w := newJobFixture(t, http.HandlerFunc(errorResponse))
	cases := []struct {
		job  Job
		want string
	}{
		{&ProfileJob{Client: c, Writer: w}, "profile"},
		{&SavedTracksJob{Client: c, Writer: w}, "saved_tracks"},
		{&SavedAlbumsJob{Client: c, Writer: w}, "saved_albums"},
		{&SavedEpisodesJob{Client: c, Writer: w}, "saved_episodes"},
		{&SavedShowsJob{Client: c, Writer: w}, "saved_shows"},
		{&FollowedArtistsJob{Client: c, Writer: w}, "followed_artists"},
		{&PlaylistsJob{Client: c, Writer: w}, "playlists"},
		{&TopArtistsJob{Client: c, Writer: w}, "top_artists"},
		{&TopTracksJob{Client: c, Writer: w}, "top_tracks"},
		{&RecentlyPlayedJob{Client: c, Writer: w}, "recently_played"},
		{&AudiobooksJob{Client: c, Writer: w}, "audiobooks"},
	}
	for _, tc := range cases {
		if got := tc.job.Name(); got != tc.want {
			t.Errorf("%T.Name() = %q, want %q", tc.job, got, tc.want)
		}
	}
}
