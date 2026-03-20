package auth

import (
	"context"
	"fmt"
	"net/http"
)

type callbackResult struct {
	code  string
	state string
	err   error
}

// startCallbackServer starts a temporary HTTP server that listens for the OAuth callback.
// It returns the authorization code and state when the callback is received.
func startCallbackServer(ctx context.Context, port string) (string, string, error) {
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", port),
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		errParam := q.Get("error")
		if errParam != "" {
			_, _ = fmt.Fprintf(w, "<html><body><h2>Authorization failed: %s</h2><p>You may close this window.</p></body></html>", errParam)
			resultCh <- callbackResult{err: fmt.Errorf("spotify authorization error: %s", errParam)}
			go srv.Shutdown(context.Background()) //nolint:errcheck
			return
		}

		code := q.Get("code")
		state := q.Get("state")

		_, _ = fmt.Fprint(w, "<html><body><h2>Authorization successful!</h2><p>You may close this window and return to the terminal.</p></body></html>")
		resultCh <- callbackResult{code: code, state: state}
		go srv.Shutdown(context.Background()) //nolint:errcheck
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			resultCh <- callbackResult{err: fmt.Errorf("callback server error: %w", err)}
		}
	}()

	select {
	case result := <-resultCh:
		return result.code, result.state, result.err
	case <-ctx.Done():
		srv.Shutdown(context.Background()) //nolint:errcheck
		return "", "", ctx.Err()
	}
}
