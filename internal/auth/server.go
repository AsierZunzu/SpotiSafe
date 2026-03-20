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

// startCallbackServer starts a temporary HTTP server with two routes:
//   - GET /        — landing page with an "Authorize with Spotify" button
//   - GET /callback — receives the OAuth redirect from Spotify
//
// The server shuts itself down after the callback is received or the context is cancelled.
func startCallbackServer(ctx context.Context, port, authURL, redirectURI string) (string, string, error) {
	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", port),
		Handler: mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>SpotiSafe — Authorization</title></head>
<body style="font-family:sans-serif;max-width:480px;margin:80px auto;text-align:center">
  <h2>SpotiSafe</h2>
  <p>Click the button below to authorize SpotiSafe to read your Spotify library.</p>
  <a href="%s" style="display:inline-block;padding:12px 24px;background:#1DB954;color:#fff;text-decoration:none;border-radius:24px;font-weight:bold">
    Authorize with Spotify
  </a>
  <p style="margin-top:40px;font-size:0.85em;color:#666">
    Redirect URI registered in your Spotify app:<br>
    <code>%s</code>
  </p>
</body>
</html>`, authURL, redirectURI)
	})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		errParam := q.Get("error")
		if errParam != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>SpotiSafe — Error</title></head>
<body style="font-family:sans-serif;max-width:480px;margin:80px auto;text-align:center">
  <h2>Authorization failed</h2>
  <p>%s</p>
</body>
</html>`, errParam)
			resultCh <- callbackResult{err: fmt.Errorf("spotify authorization error: %s", errParam)}
			go srv.Shutdown(context.Background()) //nolint:errcheck
			return
		}

		code := q.Get("code")
		state := q.Get("state")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>SpotiSafe — Success</title></head>
<body style="font-family:sans-serif;max-width:480px;margin:80px auto;text-align:center">
  <h2>Authorization successful!</h2>
  <p>SpotiSafe is now backing up your library. You may close this tab.</p>
</body>
</html>`)
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
