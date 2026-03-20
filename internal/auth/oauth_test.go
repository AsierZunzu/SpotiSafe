package auth

import (
	"path/filepath"
	"testing"

	"golang.org/x/oauth2"
)

func TestSaveAndLoadToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "token.json")

	tok := &oauth2.Token{
		AccessToken:  "access-abc",
		RefreshToken: "refresh-xyz",
		TokenType:    "Bearer",
	}

	if err := saveToken(path, tok); err != nil {
		t.Fatalf("saveToken: %v", err)
	}

	got, err := loadToken(path)
	if err != nil {
		t.Fatalf("loadToken: %v", err)
	}

	if got.AccessToken != tok.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, tok.AccessToken)
	}
	if got.RefreshToken != tok.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, tok.RefreshToken)
	}
	if got.TokenType != tok.TokenType {
		t.Errorf("TokenType = %q, want %q", got.TokenType, tok.TokenType)
	}
}

func TestSaveToken_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "token.json")

	if err := saveToken(path, &oauth2.Token{AccessToken: "x"}); err != nil {
		t.Fatalf("saveToken: %v", err)
	}
}

func TestLoadToken_MissingFile(t *testing.T) {
	_, err := loadToken("/nonexistent/path/token.json")
	if err == nil {
		t.Fatal("expected error loading missing file")
	}
}
