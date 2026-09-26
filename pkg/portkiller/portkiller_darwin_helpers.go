//go:build darwin

package portkiller

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// processStartTime на macOS получаем через ps -o lstart=.
// Формат: "Mon Jan  2 15:04:05 2006".
func processStartTime(pid int) (time.Time, error) {
	out, err := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return time.Time{}, err
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return time.Time{}, exec.ErrNotFound
	}
	return time.Parse("Mon Jan _2 15:04:05 2006", s)
}

func runtimeGOOS() string { return "darwin" }
