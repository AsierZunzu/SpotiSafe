package backup

import "context"

// Job is the interface implemented by each backup category.
type Job interface {
	// Name returns a short identifier for this job (used in logs and summary).
	Name() string
	// Run executes the backup job and returns the number of items backed up.
	Run(ctx context.Context) (int, error)
}
