package spotify

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://api.spotify.com/v1"

// Client wraps an oauth2-authenticated HTTP client with retry logic.
type Client struct {
	http *http.Client
}

// New creates a new Spotify API client from an authenticated HTTP client.
func New(httpClient *http.Client) *Client {
	return &Client{http: httpClient}
}

// get performs a GET request to the Spotify API and decodes the JSON response.
// It handles 429 (rate limit) and 5xx (server error) retries automatically.
func (c *Client) get(url string, out any) error {
	const maxRetries = 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.http.Get(url)
		if err != nil {
			return fmt.Errorf("http get %s: %w", url, err)
		}

		switch {
		case resp.StatusCode == http.StatusOK:
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
				return fmt.Errorf("decode response from %s: %w", url, err)
			}
			return nil

		case resp.StatusCode == http.StatusTooManyRequests:
			resp.Body.Close()
			if attempt == maxRetries {
				return fmt.Errorf("rate limited after %d retries: %s", maxRetries, url)
			}
			wait := retryAfter(resp)
			slog.Warn("rate limited, waiting", "seconds", wait.Seconds(), "attempt", attempt+1)
			time.Sleep(wait)

		case resp.StatusCode >= 500:
			resp.Body.Close()
			if attempt == maxRetries {
				return fmt.Errorf("server error %d after %d retries: %s", resp.StatusCode, maxRetries, url)
			}
			wait := exponentialBackoff(attempt)
			slog.Warn("server error, retrying", "status", resp.StatusCode, "wait", wait, "attempt", attempt+1)
			time.Sleep(wait)

		default:
			resp.Body.Close()
			return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, url)
		}
	}

	return fmt.Errorf("exhausted retries for %s", url)
}

func retryAfter(resp *http.Response) time.Duration {
	if v := resp.Header.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			return time.Duration(secs+1) * time.Second
		}
	}
	return 5 * time.Second
}

func exponentialBackoff(attempt int) time.Duration {
	base := time.Duration(1<<uint(attempt)) * time.Second
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	return base + jitter
}
