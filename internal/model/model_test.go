package model

import "testing"

func TestExitCode(t *testing.T) {
	if got := ExitCode(nil, nil); got != 0 {
		t.Fatalf("clean exit code = %d, want 0", got)
	}
	if got := ExitCode([]Finding{{}}, nil); got != 1 {
		t.Fatalf("finding exit code = %d, want 1", got)
	}
	if got := ExitCode(nil, errSentinel{}); got != 2 {
		t.Fatalf("error exit code = %d, want 2", got)
	}
}

type errSentinel struct{}

func (errSentinel) Error() string { return "sentinel" }
