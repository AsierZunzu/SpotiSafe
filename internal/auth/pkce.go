package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// generateVerifier creates a cryptographically random PKCE code verifier (43–128 chars).
func generateVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateChallenge derives the S256 code challenge from the verifier.
func generateChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
