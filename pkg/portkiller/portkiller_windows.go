//go:build windows

package portkiller

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func killProcessesInPortRange(from, to uint16, timeout time.Duration) (KillReport, error) {
	procs, err := windowsListeningProcesses(from, to)
	if err != nil {
		return KillReport{}, err
	}
	rep := KillReport{Found: len(procs)}
	for _, p := range procs {
		// TOCTOU: перечитываем список и сверяем PID.
		if !windowsStillListener(p, from, to) {
			rep.Skipped++
			continue
		}

		// Windows не имеет SIGTERM в POSIX-смысле.
		// Пробуем мягко: taskkill без /F. Если не вышло — /F (аналог SIGKILL).
		if err := exec.Command("taskkill", "/PID", strconv.Itoa(p.PID)).Run(); err == nil {
			if timeout > 0 {
				deadline := time.Now().Add(timeout)
				for time.Now().Before(deadline) {
					if !windowsProcessAlive(p.PID) {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
			if !windowsProcessAlive(p.PID) {
				rep.Terminated++
				continue
			}
		}

		if err := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(p.PID)).Run(); err != nil {
			rep.Failed++
			rep.Errors = append(rep.Errors, fmt.Errorf("pid %d: %w", p.PID, err))
			continue
		}
		rep.Killed++
	}
	return rep, nil
}

func windowsListeningProcesses(from, to uint16) ([]listeningProcess, error) {
	out, err := exec.Command("netstat", "-ano", "-p", "TCP").Output()
	if err != nil {
		return nil, err
	}
	var procs []listeningProcess
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if !strings.EqualFold(fields[0], "TCP") || !strings.EqualFold(fields[3], "LISTENING") {
			continue
		}
		_, port, ok := extractPortWin(fields[1])
		if !ok || port < from || port > to {
			continue
		}
		pid, err := strconv.Atoi(fields[4])
		if err != nil || pid <= 0 {
			continue
		}
		procs = append(procs, listeningProcess{PID: pid, Port: port})
	}
	return procs, nil
}

func windowsStillListener(p listeningProcess, from, to uint16) bool {
	if !windowsProcessAlive(p.PID) {
		return false
	}
	cur, err := windowsListeningProcesses(from, to)
	if err != nil {
		return false
	}
	for _, c := range cur {
		if c.PID == p.PID && c.Port == p.Port {
			return true
		}
	}
	return false
}

func windowsProcessAlive(pid int) bool {
	// OpenProcess с PROCESS_QUERY_LIMITED_INFORMATION — самый дешёвый способ.
	const PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	h, err := syscall.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	syscall.CloseHandle(h)
	return true
}

func extractPortWin(s string) (string, uint16, bool) {
	i := strings.LastIndex(s, ":")
	if i < 0 || i == len(s)-1 {
		return "", 0, false
	}
	p, err := strconv.Atoi(s[i+1:])
	if err != nil || p <= 0 || p > 65535 {
		return "", 0, false
	}
	return s[:i], uint16(p), true
}
