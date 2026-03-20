package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestFetchAllOffset_SinglePage(t *testing.T) {
	c := newTestSpotifyClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(OffsetPage[string]{
			Items: []string{"a", "b", "c"},
			Total: 3,
		})
	}))

	items, err := FetchAllOffset[string](context.Background(), c, baseURL+"/test?", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("got %d items, want 3", len(items))
	}
}

func TestFetchAllOffset_MultiplePages(t *testing.T) {
	var calls atomic.Int32
	c := newTestSpotifyClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			_ = json.NewEncoder(w).Encode(OffsetPage[string]{
				Items: []string{"a", "b"},
				Total: 4,
				Next:  "non-empty",
			})
		} else {
			_ = json.NewEncoder(w).Encode(OffsetPage[string]{
				Items: []string{"c", "d"},
				Total: 4,
			})
		}
	}))

	items, err := FetchAllOffset[string](context.Background(), c, baseURL+"/test?", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 4 {
		t.Errorf("got %d items, want 4", len(items))
	}
	if calls.Load() != 2 {
		t.Errorf("expected 2 API calls, got %d", calls.Load())
	}
}

func TestFetchAllOffset_Empty(t *testing.T) {
	c := newTestSpotifyClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(OffsetPage[string]{})
	}))

	items, err := FetchAllOffset[string](context.Background(), c, baseURL+"/test?", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestFetchAllOffset_PropagatesError(t *testing.T) {
	c := newTestSpotifyClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	_, err := FetchAllOffset[string](context.Background(), c, baseURL+"/test?", 50)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
