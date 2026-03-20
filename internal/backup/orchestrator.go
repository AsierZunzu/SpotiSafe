package backup

import (
	"fmt"
	"log/slog"
)

// Result holds the outcome of a single backup job.
type Result struct {
	JobName string
	Count   int
	Err     error
}

// Orchestrator runs all backup jobs and collects results.
type Orchestrator struct {
	Jobs []Job
}

// Run executes all jobs sequentially (soft-fail per job) and returns the results.
func (o *Orchestrator) Run() []Result {
	results := make([]Result, 0, len(o.Jobs))

	for _, job := range o.Jobs {
		slog.Info("running job", "job", job.Name())
		count, err := job.Run()
		results = append(results, Result{
			JobName: job.Name(),
			Count:   count,
			Err:     err,
		})
		if err != nil {
			slog.Warn("job failed", "job", job.Name(), "err", err)
		} else {
			slog.Info("job complete", "job", job.Name(), "count", count)
		}
	}

	return results
}

// PrintSummary prints a human-readable backup summary and returns an exit code.
// Exit codes: 0 = all OK, 2 = all jobs failed.
func PrintSummary(results []Result) int {
	succeeded := 0
	failed := 0

	for _, r := range results {
		if r.Err != nil {
			failed++
		} else {
			succeeded++
		}
	}

	fmt.Printf("\nBackup complete (%d succeeded, %d failed):\n", succeeded, failed)
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("  [FAIL] %s: %v\n", r.JobName, r.Err)
		} else {
			fmt.Printf("  [OK]   %s (%d items)\n", r.JobName, r.Count)
		}
	}
	fmt.Println()

	if failed > 0 && succeeded == 0 {
		return 2
	}
	return 0
}
