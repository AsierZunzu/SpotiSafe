package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

// rewriteTransport redirects all requests to the given target server,
// preserving the original path and query.
type rewriteTransport struct {
	target string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	u, _ := url.Parse(t.target)
	req = req.Clone(req.Context())
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
	return http.DefaultTransport.RoundTrip(req)
}

func newTestSpotifyClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(&http.Client{Transport: &rewriteTransport{target: srv.URL}})
}

func TestGet_Success(t *testing.T) {
	type payload struct {
		ID string `json:"id"`
	}
	c := newTestSpotifyClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(payload{ID: "user123"})
	}))

	var out payload
	if err := c.get(context.Background(), baseURL+"/me", &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ID != "user123" {
		t.Errorf("ID = %q, want user123", out.ID)
	}
}

func TestGet_UnexpectedStatus(t *testing.T) {
	c := newTestSpotifyClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))

	var out any
	if err := c.get(context.Background(), baseURL+"/me", &out); err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestGet_InvalidJSON(t *testing.T) {
	c := newTestSpotifyClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))

	var out map[string]any
	if err := c.get(context.Background(), baseURL+"/me", &out); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGet_ContextCancelled(t *testing.T) {
	c := newTestSpotifyClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{})
	}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out any
	if err := c.get(ctx, baseURL+"/me", &out); err == nil {
		t.Fatal("expected error with cancelled context")
	}
}

func TestGet_RetriesOn429ThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	c := newTestSpotifyClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "ok"})
	}))

	var out map[string]string
	if err := c.get(context.Background(), baseURL+"/me", &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 2 {
		t.Errorf("expected 2 calls, got %d", calls.Load())
	}
}

func TestRetryAfter_WithHeader(t *testing.T) {
	resp := &http.Response{Header: http.Header{"Retry-After": {"10"}}}
	if d := retryAfter(resp); d != 11*time.Second {
		t.Errorf("retryAfter = %v, want 11s", d)
	}
}

func TestRetryAfter_NoHeader(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	if d := retryAfter(resp); d != 5*time.Second {
		t.Errorf("retryAfter = %v, want 5s", d)
	}
}

func TestRetryAfter_InvalidHeader(t *testing.T) {
	resp := &http.Response{Header: http.Header{"Retry-After": {"not-a-number"}}}
	if d := retryAfter(resp); d != 5*time.Second {
		t.Errorf("retryAfter = %v, want 5s", d)
	}
}

func TestExponentialBackoff_Bounds(t *testing.T) {
	for attempt := 0; attempt < 4; attempt++ {
		base := time.Duration(1<<uint(attempt)) * time.Second
		d := exponentialBackoff(attempt)
		if d < base || d >= base+time.Second {
			t.Errorf("attempt %d: backoff %v not in [%v, %v)", attempt, d, base, base+time.Second)
		}
	}
}

func TestExponentialBackoff_Increases(t *testing.T) {
	// Base values increase strictly: 1s, 2s, 4s, 8s — jitter is at most 999ms,
	// so attempt N base always exceeds attempt N-1 base + max jitter for N >= 2.
	if exponentialBackoff(2) <= exponentialBackoff(0) {
		// probabilistically near-impossible to fail (4s+jitter > 1s+jitter)
		t.Error("backoff should increase with attempt number")
	}
}
