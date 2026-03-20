package auth

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateVerifier_Length(t *testing.T) {
	v, err := generateVerifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// base64 RawURL of 32 bytes = 43 chars
	if len(v) != 43 {
		t.Errorf("len = %d, want 43", len(v))
	}
}

func TestGenerateVerifier_ValidBase64URL(t *testing.T) {
	v, err := generateVerifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := base64.RawURLEncoding.DecodeString(v); err != nil {
		t.Errorf("not valid base64url: %v", err)
	}
	if strings.ContainsAny(v, "+/=") {
		t.Error("verifier contains chars invalid for base64url")
	}
}

func TestGenerateVerifier_Unique(t *testing.T) {
	v1, _ := generateVerifier()
	v2, _ := generateVerifier()
	if v1 == v2 {
		t.Error("expected two distinct verifiers")
	}
}

func TestGenerateChallenge_KnownVector(t *testing.T) {
	// Test vector from RFC 7636 Appendix B.
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got := generateChallenge(verifier); got != want {
		t.Errorf("generateChallenge = %q, want %q", got, want)
	}
}

func TestGenerateChallenge_Deterministic(t *testing.T) {
	c1 := generateChallenge("same-verifier")
	c2 := generateChallenge("same-verifier")
	if c1 != c2 {
		t.Error("generateChallenge should be deterministic")
	}
}

func TestGenerateChallenge_DifferentInputs(t *testing.T) {
	c1 := generateChallenge("verifier-one")
	c2 := generateChallenge("verifier-two")
	if c1 == c2 {
		t.Error("different verifiers should produce different challenges")
	}
}

func TestRandomState_ValidBase64URL(t *testing.T) {
	s, err := randomState()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s) == 0 {
		t.Error("state must not be empty")
	}
	if _, err := base64.RawURLEncoding.DecodeString(s); err != nil {
		t.Errorf("not valid base64url: %v", err)
	}
}

func TestRandomState_Unique(t *testing.T) {
	s1, _ := randomState()
	s2, _ := randomState()
	if s1 == s2 {
		t.Error("expected distinct states")
	}
}
