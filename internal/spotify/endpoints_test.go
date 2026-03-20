package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func handler200(v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func handler500(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func TestGetProfile(t *testing.T) {
	c := newTestSpotifyClient(t, handler200(UserProfile{ID: "u1", DisplayName: "Alice"}))
	p, err := c.GetProfile(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != "u1" {
		t.Errorf("ID = %q, want u1", p.ID)
	}
}

func TestGetProfile_Error(t *testing.T) {
	c := newTestSpotifyClient(t, http.HandlerFunc(handler500))
	if _, err := c.GetProfile(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetSavedTracks(t *testing.T) {
	page := OffsetPage[SavedTrack]{Items: []SavedTrack{{Track: Track{ID: "t1"}}}}
	c := newTestSpotifyClient(t, handler200(page))
	items, err := c.GetSavedTracks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Track.ID != "t1" {
		t.Errorf("unexpected items: %v", items)
	}
}

func TestGetSavedAlbums(t *testing.T) {
	page := OffsetPage[SavedAlbum]{Items: []SavedAlbum{{Album: Album{ID: "a1"}}}}
	c := newTestSpotifyClient(t, handler200(page))
	items, err := c.GetSavedAlbums(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}

func TestGetSavedEpisodes(t *testing.T) {
	page := OffsetPage[SavedEpisode]{Items: []SavedEpisode{{Episode: Episode{ID: "e1"}}}}
	c := newTestSpotifyClient(t, handler200(page))
	items, err := c.GetSavedEpisodes(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}

func TestGetSavedShows(t *testing.T) {
	page := OffsetPage[SavedShow]{Items: []SavedShow{{Show: Show{ID: "s1"}}}}
	c := newTestSpotifyClient(t, handler200(page))
	items, err := c.GetSavedShows(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}

func TestGetPlaylists(t *testing.T) {
	page := OffsetPage[PlaylistSummary]{Items: []PlaylistSummary{{ID: "pl1"}}}
	c := newTestSpotifyClient(t, handler200(page))
	items, err := c.GetPlaylists(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}

func TestGetPlaylistTracks(t *testing.T) {
	page := OffsetPage[PlaylistTrack]{Items: []PlaylistTrack{{Track: Track{ID: "t1"}}}}
	c := newTestSpotifyClient(t, handler200(page))
	items, err := c.GetPlaylistTracks(context.Background(), "pl1")
	if err != nil || len(items) != 1 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}

func TestGetTopArtists(t *testing.T) {
	page := OffsetPage[TopArtist]{Items: []TopArtist{{ID: "a1"}}}
	c := newTestSpotifyClient(t, handler200(page))
	items, err := c.GetTopArtists(context.Background(), "short_term")
	if err != nil || len(items) != 1 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}

func TestGetTopTracks(t *testing.T) {
	page := OffsetPage[TopTrack]{Items: []TopTrack{{ID: "t1"}}}
	c := newTestSpotifyClient(t, handler200(page))
	items, err := c.GetTopTracks(context.Background(), "long_term")
	if err != nil || len(items) != 1 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}

func TestGetFollowedArtists(t *testing.T) {
	resp := FollowedArtistsResponse{
		Artists: CursorPage[Artist]{Items: []Artist{{ID: "a1"}}},
	}
	c := newTestSpotifyClient(t, handler200(resp))
	items, err := c.GetFollowedArtists(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}

func TestGetFollowedArtists_Pagination(t *testing.T) {
	// First page has a cursor; second page is the last.
	var call int
	c := newTestSpotifyClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call == 1 {
			_ = json.NewEncoder(w).Encode(FollowedArtistsResponse{
				Artists: CursorPage[Artist]{
					Items:   []Artist{{ID: "a1"}},
					Next:    "non-empty",
					Cursors: struct {
						After  string `json:"after"`
						Before string `json:"before"`
					}{"cursor1", ""},
				},
			})
		} else {
			_ = json.NewEncoder(w).Encode(FollowedArtistsResponse{
				Artists: CursorPage[Artist]{Items: []Artist{{ID: "a2"}}},
			})
		}
	}))
	items, err := c.GetFollowedArtists(context.Background())
	if err != nil || len(items) != 2 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}

func TestGetRecentlyPlayed(t *testing.T) {
	resp := CursorPage[RecentlyPlayedItem]{
		Items: []RecentlyPlayedItem{{Track: Track{ID: "t1"}}},
	}
	c := newTestSpotifyClient(t, handler200(resp))
	items, err := c.GetRecentlyPlayed(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}

func TestGetAudiobooks(t *testing.T) {
	page := OffsetPage[SavedAudiobook]{Items: []SavedAudiobook{{ID: "b1"}}}
	c := newTestSpotifyClient(t, handler200(page))
	items, err := c.GetAudiobooks(context.Background())
	if err != nil || len(items) != 1 {
		t.Fatalf("err=%v, len=%d", err, len(items))
	}
}
