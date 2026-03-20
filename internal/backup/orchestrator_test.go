package backup

import (
	"context"
	"errors"
	"testing"
)

type mockJob struct {
	name  string
	count int
	err   error
}

func (j *mockJob) Name() string                       { return j.name }
func (j *mockJob) Run(_ context.Context) (int, error) { return j.count, j.err }

func TestPrintSummary_AllSucceeded(t *testing.T) {
	results := []Result{
		{JobName: "profile", Count: 1},
		{JobName: "tracks", Count: 100},
	}
	if code := PrintSummary(results); code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestPrintSummary_AllFailed(t *testing.T) {
	results := []Result{
		{JobName: "profile", Err: errors.New("auth error")},
		{JobName: "tracks", Err: errors.New("timeout")},
	}
	if code := PrintSummary(results); code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
}

func TestPrintSummary_Mixed(t *testing.T) {
	results := []Result{
		{JobName: "profile", Count: 1},
		{JobName: "tracks", Err: errors.New("timeout")},
	}
	if code := PrintSummary(results); code != 0 {
		t.Errorf("exit code = %d, want 0 for partial success", code)
	}
}

func TestOrchestrator_RunCollectsResults(t *testing.T) {
	someErr := errors.New("boom")
	orch := &Orchestrator{
		Jobs: []Job{
			&mockJob{name: "a", count: 5},
			&mockJob{name: "b", err: someErr},
			&mockJob{name: "c", count: 10},
		},
	}

	results := orch.Run(context.Background())

	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	if results[0].JobName != "a" || results[0].Count != 5 || results[0].Err != nil {
		t.Errorf("result[0] = %+v", results[0])
	}
	if results[1].JobName != "b" || results[1].Err != someErr {
		t.Errorf("result[1] = %+v", results[1])
	}
	if results[2].JobName != "c" || results[2].Count != 10 || results[2].Err != nil {
		t.Errorf("result[2] = %+v", results[2])
	}
}

func TestOrchestrator_ContinuesAfterFailure(t *testing.T) {
	ran := map[string]bool{}
	jobs := []Job{
		&mockJob{name: "a", err: errors.New("fail")},
		&mockJob{name: "b", count: 1},
	}
	// Override Run to track execution
	orch := &Orchestrator{Jobs: jobs}
	results := orch.Run(context.Background())

	for _, r := range results {
		ran[r.JobName] = true
	}
	if !ran["a"] || !ran["b"] {
		t.Error("orchestrator should run all jobs even after a failure")
	}
}
