package writer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Envelope is the top-level structure written to every backup JSON file.
type Envelope struct {
	Metadata Metadata `json:"metadata"`
	Data     any      `json:"data"`
}

// Metadata describes the backup run context for a single file.
type Metadata struct {
	Category  string `json:"category"`
	FetchedAt string `json:"fetched_at"`
	Count     int    `json:"count"`
	UserID    string `json:"user_id"`
}

// JSONWriter writes backup files to a timestamped output directory.
type JSONWriter struct {
	runDir string
	userID string
}

// New creates a JSONWriter for the current backup run.
// outputDir is the base output directory; a timestamped subdirectory is created inside it.
func New(outputDir, userID string) (*JSONWriter, error) {
	ts := time.Now().UTC().Format("2006-01-02T15-04-05")
	runDir := filepath.Join(outputDir, ts)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return nil, fmt.Errorf("create run directory %s: %w", runDir, err)
	}
	return &JSONWriter{runDir: runDir, userID: userID}, nil
}

// RunDir returns the path to the current run's output directory.
func (w *JSONWriter) RunDir() string {
	return w.runDir
}

// Write atomically writes data to <runDir>/<filename>.json.
// count is the number of items in data (used for metadata).
func (w *JSONWriter) Write(filename, category string, count int, data any) error {
	env := Envelope{
		Metadata: Metadata{
			Category:  category,
			FetchedAt: time.Now().UTC().Format(time.RFC3339),
			Count:     count,
			UserID:    w.userID,
		},
		Data: data,
	}

	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", category, err)
	}

	dest := filepath.Join(w.runDir, filename+".json")
	return atomicWrite(dest, b)
}

// WriteToSubdir writes data into a subdirectory of the run directory.
func (w *JSONWriter) WriteToSubdir(subdir, filename, category string, count int, data any) error {
	dir := filepath.Join(w.runDir, subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create subdir %s: %w", dir, err)
	}

	env := Envelope{
		Metadata: Metadata{
			Category:  category,
			FetchedAt: time.Now().UTC().Format(time.RFC3339),
			Count:     count,
			UserID:    w.userID,
		},
		Data: data,
	}

	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", category, err)
	}

	dest := filepath.Join(dir, filename+".json")
	return atomicWrite(dest, b)
}

// atomicWrite writes b to path via a temp file + rename to avoid partial writes.
func atomicWrite(path string, b []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".spotisafe-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename to %s: %w", path, err)
	}
	return nil
}

// SanitizeName replaces characters unsafe for filenames with underscores.
func SanitizeName(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	result := replacer.Replace(name)
	if len(result) > 80 {
		result = result[:80]
	}
	return strings.TrimSpace(result)
}
