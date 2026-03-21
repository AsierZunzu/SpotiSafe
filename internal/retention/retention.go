package retention

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Policy defines how many backup directories to retain in each time window.
// A value of 0 disables that rule.
type Policy struct {
	KeepLast      int // N most recent backups, unconditional
	KeepLastDay   int // N oldest from the 0–24 h window
	KeepLastWeek  int // N oldest from the 1–7 day window
	KeepLastMonth int // N oldest from the 7–30 day window
	KeepLastYear  int // N oldest from the 30–365 day window
}

const dirTimestampLayout = "2006-01-02T15-04-05"

type backup struct {
	name string
	t    time.Time
}

// Apply scans outputDir for timestamped backup directories and removes any
// that are not selected by the policy. now is the reference time (use time.Now().UTC()).
func Apply(outputDir string, policy Policy, now time.Time) error {
	backups, err := loadBackups(outputDir)
	if err != nil {
		return err
	}
	if len(backups) == 0 {
		return nil
	}

	keep := selectKeep(backups, policy, now)
	deleteUnselected(outputDir, backups, keep)
	return nil
}

func loadBackups(outputDir string) ([]backup, error) {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("read output dir: %w", err)
	}

	var backups []backup
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		t, err := time.ParseInLocation(dirTimestampLayout, e.Name(), time.UTC)
		if err != nil {
			continue // not a timestamped backup directory
		}
		backups = append(backups, backup{name: e.Name(), t: t})
	}
	return backups, nil
}

func selectKeep(backups []backup, policy Policy, now time.Time) map[string]bool {
	// Sort newest first for keep_last selection.
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].t.After(backups[j].t)
	})

	keep := make(map[string]bool, len(backups))

	for i := 0; i < policy.KeepLast && i < len(backups); i++ {
		keep[backups[i].name] = true
	}

	day, week, month, year := partitionByAge(backups, now)
	markOldest(keep, day, policy.KeepLastDay)
	markOldest(keep, week, policy.KeepLastWeek)
	markOldest(keep, month, policy.KeepLastMonth)
	markOldest(keep, year, policy.KeepLastYear)

	return keep
}

func partitionByAge(backups []backup, now time.Time) (day, week, month, year []backup) {
	for _, b := range backups {
		age := now.Sub(b.t)
		switch {
		case age < 24*time.Hour:
			day = append(day, b)
		case age < 7*24*time.Hour:
			week = append(week, b)
		case age < 30*24*time.Hour:
			month = append(month, b)
		case age < 365*24*time.Hour:
			year = append(year, b)
		}
	}
	return
}

func markOldest(keep map[string]bool, window []backup, n int) {
	if n <= 0 || len(window) == 0 {
		return
	}
	sort.Slice(window, func(i, j int) bool {
		return window[i].t.Before(window[j].t) // oldest first
	})
	for i := 0; i < n && i < len(window); i++ {
		keep[window[i].name] = true
	}
}

func deleteUnselected(outputDir string, backups []backup, keep map[string]bool) {
	for _, b := range backups {
		if keep[b.name] {
			continue
		}
		path := filepath.Join(outputDir, b.name)
		slog.Info("retention: removing old backup", "dir", b.name)
		if err := os.RemoveAll(path); err != nil {
			slog.Warn("retention: failed to remove backup", "dir", b.name, "err", err)
		}
	}
}
