package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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
func GetClient(ctx context.Context, clientID, redirectURI, callbackPort, tokenFile string) (*http.Client, error) {
	oauthCfg := &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: redirectURI,
		Scopes:      scopes,
		Endpoint:    spotify.Endpoint,
	}

	// Try to load existing token
	if tok, err := loadToken(tokenFile); err == nil {
		slog.Info("loaded existing token from file", "file", tokenFile)
		return oauthCfg.Client(ctx, tok), nil
	}

	// Run PKCE flow
	tok, err := runPKCEFlow(ctx, oauthCfg, callbackPort)
	if err != nil {
		return nil, err
	}

	if err := saveToken(tokenFile, tok); err != nil {
		slog.Warn("could not save token", "err", err)
	}

	return oauthCfg.Client(ctx, tok), nil
}

func runPKCEFlow(ctx context.Context, cfg *oauth2.Config, port string) (*oauth2.Token, error) {
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

	fmt.Println()
	fmt.Println("=============================================================")
	fmt.Println("  SpotiSafe — Authorization Required")
	fmt.Println("=============================================================")
	fmt.Println("Open the following URL in your browser to authorize SpotiSafe:")
	fmt.Println()
	fmt.Println("  " + authURL)
	fmt.Println()
	fmt.Printf("Waiting for callback on port %s...\n", port)
	fmt.Println("=============================================================")
	fmt.Println()

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	code, returnedState, err := startCallbackServer(timeoutCtx, port)
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

func saveToken(path string, tok *oauth2.Token) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(tok)
}

func loadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
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
