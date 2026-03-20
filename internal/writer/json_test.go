package writer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain", "My Playlist", "My Playlist"},
		{"slash", "Rock/Pop", "Rock_Pop"},
		{"backslash", `Back\slash`, "Back_slash"},
		{"colon", "Morning:Vibes", "Morning_Vibes"},
		{"star", "Top*Tracks", "Top_Tracks"},
		{"question", "What?", "What_"},
		{"quote", `Say "Hi"`, "Say _Hi_"},
		{"angle brackets", "<best>", "_best_"},
		{"pipe", "a|b", "a_b"},
		{"truncate", strings.Repeat("x", 90), strings.Repeat("x", 80)},
		{"trim spaces", "  hello  ", "hello"},
		{"empty", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := SanitizeName(tc.input); got != tc.want {
				t.Errorf("SanitizeName(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNew_CreatesRunDir(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, "user1")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := os.Stat(w.RunDir()); err != nil {
		t.Errorf("run dir does not exist: %v", err)
	}
}

func TestWrite_CreatesFileWithEnvelope(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, "user1")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	data := []string{"track1", "track2"}
	if err := w.Write("saved_tracks", "saved_tracks", len(data), data); err != nil {
		t.Fatalf("Write: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(w.RunDir(), "saved_tracks.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if env.Metadata.Category != "saved_tracks" {
		t.Errorf("Category = %q, want saved_tracks", env.Metadata.Category)
	}
	if env.Metadata.Count != 2 {
		t.Errorf("Count = %d, want 2", env.Metadata.Count)
	}
	if env.Metadata.UserID != "user1" {
		t.Errorf("UserID = %q, want user1", env.Metadata.UserID)
	}
	if env.Metadata.FetchedAt == "" {
		t.Error("FetchedAt must not be empty")
	}
}

func TestWriteToSubdir_CreatesSubdirAndFile(t *testing.T) {
	dir := t.TempDir()
	w, err := New(dir, "user1")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := w.WriteToSubdir("playlist_tracks", "pl123_name", "playlist_tracks:pl123", 3, []string{"a", "b", "c"}); err != nil {
		t.Fatalf("WriteToSubdir: %v", err)
	}

	path := filepath.Join(w.RunDir(), "playlist_tracks", "pl123_name.json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created at expected path: %v", err)
	}
}

func TestWrite_Idempotent(t *testing.T) {
	dir := t.TempDir()
	w, _ := New(dir, "u")

	if err := w.Write("file", "cat", 1, "data"); err != nil {
		t.Fatalf("first Write: %v", err)
	}
	if err := w.Write("file", "cat", 1, "data"); err != nil {
		t.Fatalf("second Write: %v", err)
	}
}
