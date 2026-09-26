//go:build darwin

package portkiller

import "testing"

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
