//go:build linux

package portkiller

import (
	"testing"
)

func TestProcessStartTime_Self(t *testing.T) {
	t.Parallel()
	st, err := processStartTime(osGetpid())
	if err != nil {
		t.Fatalf("processStartTime: %v", err)
	}
	if st.IsZero() {
		t.Fatal("expected non-zero start time")
	}
}

func TestProcessStartTime_Stable(t *testing.T) {
	t.Parallel()
	pid := osGetpid()
	st1, err := processStartTime(pid)
	if err != nil {
		t.Fatalf("processStartTime #1: %v", err)
	}
	st2, err := processStartTime(pid)
	if err != nil {
		t.Fatalf("processStartTime #2: %v", err)
	}
	if !st1.Equal(st2) {
		t.Fatalf("start time changed between calls: %v vs %v", st1, st2)
	}
}

func TestProcessStartTime_Dead(t *testing.T) {
	t.Parallel()
	if _, err := processStartTime(1 << 30); err == nil {
		t.Fatal("expected error for non-existent pid")
	}
}
