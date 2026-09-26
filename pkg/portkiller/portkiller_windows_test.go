//go:build windows

package portkiller

import "os"

func osGetpid() int { return os.Getpid() }

func TestExtractPortWin(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in       string
		wantPort uint16
		wantOk   bool
	}{
		{"0.0.0.0:8080", 8080, true},
		{"[::]:443", 443, true},
		{"0.0.0.0:0", 0, false},
		{"0.0.0.0:70000", 0, false},
		{"0.0.0.0:", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		_, port, ok := extractPortWin(c.in)
		if ok != c.wantOk || (ok && port != c.wantPort) {
			t.Errorf("extractPortWin(%q) = (%d, %v), want (%d, %v)",
				c.in, port, ok, c.wantPort, c.wantOk)
		}
	}
}
