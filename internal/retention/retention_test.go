package retention

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// makeBackups creates timestamped directories inside a temp dir.
// ages is a list of durations subtracted from now.
func makeBackups(t *testing.T, now time.Time, ages []time.Duration) (dir string, names []string) {
	t.Helper()
	dir = t.TempDir()
	for _, age := range ages {
		ts := now.Add(-age).UTC().Format(dirTimestampLayout)
		if err := os.Mkdir(filepath.Join(dir, ts), 0o755); err != nil {
			t.Fatal(err)
		}
		names = append(names, ts)
	}
	return dir, names
}

func listDirs(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs)
	return dirs
}

func TestApply_KeepLast(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	ages := []time.Duration{
		1 * time.Hour,
		2 * time.Hour,
		3 * time.Hour,
		4 * time.Hour,
		5 * time.Hour,
		6 * time.Hour,
	}
	dir, _ := makeBackups(t, now, ages)

	policy := Policy{KeepLast: 3}
	if err := Apply(dir, policy, now); err != nil {
		t.Fatal(err)
	}

	remaining := listDirs(t, dir)
	if len(remaining) != 3 {
		t.Errorf("expected 3 dirs, got %d: %v", len(remaining), remaining)
	}
	// The 3 newest should survive: 1h, 2h, 3h ago.
	for _, d := range remaining {
		ts, _ := time.ParseInLocation(dirTimestampLayout, d, time.UTC)
		age := now.Sub(ts)
		if age > 3*time.Hour+time.Minute {
			t.Errorf("dir %s is too old (%v), expected only newest 3", d, age)
		}
	}
}

func TestApply_TimeWindows(t *testing.T) {
	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	// One backup per day for 40 days.
	ages := make([]time.Duration, 40)
	for i := range ages {
		ages[i] = time.Duration(i+1) * 24 * time.Hour
	}
	dir, _ := makeBackups(t, now, ages)

	// Add a recent backup (2h ago).
	ts := now.Add(-2 * time.Hour).UTC().Format(dirTimestampLayout)
	if err := os.Mkdir(filepath.Join(dir, ts), 0o755); err != nil {
		t.Fatal(err)
	}

	policy := Policy{
		KeepLast:      2,
		KeepLastDay:   1,
		KeepLastWeek:  2,
		KeepLastMonth: 3,
		KeepLastYear:  0, // disabled
	}
	if err := Apply(dir, policy, now); err != nil {
		t.Fatal(err)
	}

	remaining := listDirs(t, dir)
	// keep_last=2: 2h ago, 1 day ago
	// keep_last_day=1 oldest in 0-24h: 2h ago (only one in window, already kept)
	// keep_last_week=2 oldest in 1-7 days: days 6 and 7 (ages 6d and 7d)
	// keep_last_month=3 oldest in 7-30 days: days 28, 29, 30 (ages 28d, 29d, 30d)
	// keep_last_year disabled
	// Days 31-40 are in 30-365d window but year rule is disabled → deleted
	// Days 2-5 in week window but not oldest 2 → deleted
	// Days 8-27 in month window but not oldest 3 → deleted

	// Expected kept: 2h, 1d, 6d, 7d, 28d, 29d, 30d = 7 dirs
	if len(remaining) != 7 {
		t.Errorf("expected 7 dirs, got %d: %v", len(remaining), remaining)
	}
}

func TestApply_NoBackups(t *testing.T) {
	dir := t.TempDir()
	// Create a non-backup directory to ensure it's not touched.
	if err := os.Mkdir(filepath.Join(dir, "playlist_tracks"), 0o755); err != nil {
		t.Fatal(err)
	}

	policy := Policy{KeepLast: 5}
	if err := Apply(dir, policy, time.Now()); err != nil {
		t.Fatal(err)
	}

	remaining := listDirs(t, dir)
	if len(remaining) != 1 || remaining[0] != "playlist_tracks" {
		t.Errorf("non-backup dir should be untouched, got: %v", remaining)
	}
}

func TestApply_AllDisabled(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	ages := []time.Duration{1 * time.Hour, 2 * time.Hour, 3 * time.Hour}
	dir, _ := makeBackups(t, now, ages)

	policy := Policy{} // all zeros → delete everything
	if err := Apply(dir, policy, now); err != nil {
		t.Fatal(err)
	}

	remaining := listDirs(t, dir)
	if len(remaining) != 0 {
		t.Errorf("expected all deleted, got: %v", remaining)
	}
}

func TestApply_OldestKeptInWindow(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	// 5 backups within the 1-7 day window.
	ages := []time.Duration{
		2 * 24 * time.Hour,
		3 * 24 * time.Hour,
		4 * 24 * time.Hour,
		5 * 24 * time.Hour,
		6 * 24 * time.Hour,
	}
	dir, _ := makeBackups(t, now, ages)

	policy := Policy{KeepLastWeek: 2} // keep 2 oldest in week window = 5d and 6d ago
	if err := Apply(dir, policy, now); err != nil {
		t.Fatal(err)
	}

	remaining := listDirs(t, dir)
	if len(remaining) != 2 {
		t.Errorf("expected 2 dirs, got %d: %v", len(remaining), remaining)
	}
	for _, d := range remaining {
		ts, _ := time.ParseInLocation(dirTimestampLayout, d, time.UTC)
		age := now.Sub(ts)
		if age < 5*24*time.Hour-time.Minute {
			t.Errorf("dir %s is too new (%v), expected oldest 2 from week window", d, age)
		}
	}
}
