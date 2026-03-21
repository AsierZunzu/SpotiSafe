package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

var scopes = []string{
	"user-read-private",
	"user-read-email",
	"user-library-read",
	"user-follow-read",
	"user-top-read",
	"user-read-recently-played",
	"playlist-read-private",
	"playlist-read-collaborative",
}

// GetClient returns an authenticated HTTP client. It loads a saved token if
// available (auto-refreshing if expired), otherwise runs the PKCE browser flow.
func GetClient(ctx context.Context, clientID, callbackPort, publicURL, tokenFile string) (*http.Client, error) {
	redirectURI := publicURL + "/callback"
	oauthCfg := &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: redirectURI,
		Scopes:      scopes,
		Endpoint:    spotify.Endpoint,
	}

	// Try to load existing token
	if tok, err := loadToken(tokenFile); err == nil {
		slog.Info("loaded existing token from file", "file", tokenFile)
		src := &persistingTokenSource{
			base: oauthCfg.TokenSource(ctx, tok),
			path: tokenFile,
			last: tok,
		}
		return oauth2.NewClient(ctx, src), nil
	}

	// Run PKCE flow
	tok, err := runPKCEFlow(ctx, oauthCfg, callbackPort, publicURL, redirectURI)
	if err != nil {
		return nil, err
	}

	if err := saveToken(tokenFile, tok); err != nil {
		slog.Warn("could not save token", "err", err)
	}

	src := &persistingTokenSource{
		base: oauthCfg.TokenSource(ctx, tok),
		path: tokenFile,
		last: tok,
	}
	return oauth2.NewClient(ctx, src), nil
}

// persistingTokenSource wraps an oauth2.TokenSource and writes the token to
// disk whenever it changes (i.e. after a refresh), so subsequent runs can
// reuse the new access token without re-authorizing.
type persistingTokenSource struct {
	base oauth2.TokenSource
	path string
	last *oauth2.Token
}

func (s *persistingTokenSource) Token() (*oauth2.Token, error) {
	tok, err := s.base.Token()
	if err != nil {
		var re *oauth2.RetrieveError
		if errors.As(err, &re) && re.ErrorCode == "invalid_grant" {
			slog.Warn("refresh token revoked, deleting saved token", "file", s.path)
			if rmErr := os.Remove(s.path); rmErr != nil && !os.IsNotExist(rmErr) {
				slog.Warn("could not delete token file", "err", rmErr)
			}
			return nil, fmt.Errorf("refresh token revoked — delete the token file and re-authorize: %w", err)
		}
		return nil, err
	}
	if tok.AccessToken != s.last.AccessToken {
		slog.Info("token refreshed, persisting to disk", "file", s.path)
		if saveErr := saveToken(s.path, tok); saveErr != nil {
			slog.Warn("could not persist refreshed token", "err", saveErr)
		}
		s.last = tok
	}
	return tok, nil
}

func runPKCEFlow(ctx context.Context, cfg *oauth2.Config, port, publicURL, redirectURI string) (*oauth2.Token, error) {
	verifier, err := generateVerifier()
	if err != nil {
		return nil, fmt.Errorf("generate verifier: %w", err)
	}
	challenge := generateChallenge(verifier)

	state, err := randomState()
	if err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}

	authURL := cfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", challenge),
	)

	fmt.Printf("Redirect URI (register this in your Spotify app): %s\n", redirectURI)
	fmt.Printf("Open %s in your browser to authorize SpotiSafe.\n", publicURL)

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	code, returnedState, err := startCallbackServer(timeoutCtx, port, authURL, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("callback: %w", err)
	}

	if returnedState != state {
		return nil, fmt.Errorf("state mismatch: possible CSRF attack")
	}

	tok, err := cfg.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}

	slog.Info("authorization successful")
	return tok, nil
}

func saveToken(path string, tok *oauth2.Token) (retErr error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && retErr == nil {
			retErr = cerr
		}
	}()
	return json.NewEncoder(f).Encode(tok)
}

func loadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var tok oauth2.Token
	if err := json.NewDecoder(f).Decode(&tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
