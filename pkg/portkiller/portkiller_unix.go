//go:build linux || darwin

package portkiller

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func killProcessesInPortRange(from, to uint16, timeout time.Duration) (KillReport, error) {
	procs, err := findListeningProcesses(from, to)
	if err != nil {
		return KillReport{}, err
	}

	rep := KillReport{Found: len(procs)}
	for _, p := range procs {
		outcome, err := terminateProcess(p, from, to, timeout)
		if err != nil {
			rep.Failed++
			rep.Errors = append(rep.Errors, fmt.Errorf("pid %d: %w", p.PID, err))
			continue
		}
		switch outcome {
		case outcomeTerminated:
			rep.Terminated++
		case outcomeKilled:
			rep.Killed++
		case outcomeSkipped:
			rep.Skipped++
		}
	}
	return rep, nil
}

type outcome int

const (
	outcomeTerminated outcome = iota
	outcomeKilled
	outcomeSkipped
)

// terminateProcess выполняет: TOCTOU-проверку → SIGTERM → ожидание → SIGKILL.
func terminateProcess(p listeningProcess, from, to uint16, timeout time.Duration) (outcome, error) {
	// --- TOCTOU: проверяем, что процесс всё ещё тот же и всё ещё слушает порт ---
	if !stillSameListener(p, from, to) {
		return outcomeSkipped, nil
	}

	proc, err := os.FindProcess(p.PID)
	if err != nil {
		return outcomeSkipped, err
	}

	// --- Graceful: SIGTERM ---
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		// ESRCH = процесс уже умер — считаем как skipped.
		if errors.Is(err, syscall.ESRCH) {
			return outcomeSkipped, nil
		}
		return outcomeSkipped, fmt.Errorf("SIGTERM: %w", err)
	}

	if timeout <= 0 {
		return outcomeTerminated, nil
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processAlive(p.PID) {
			return outcomeTerminated, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	// --- Force: SIGKILL ---
	if err := proc.Signal(syscall.SIGKILL); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return outcomeTerminated, nil
		}
		return outcomeKilled, fmt.Errorf("SIGKILL: %w", err)
	}
	return outcomeKilled, nil
}

// processAlive возвращает true, если процесс существует.
func processAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = p.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

// stillSameListener проверяет, что процесс с данным PID:
//  1. всё ещё жив;
//  2. всё ещё слушает порт в диапазоне [from, to];
//  3. имеет то же время старта (защита от переиспользования PID).
func stillSameListener(p listeningProcess, from, to uint16) bool {
	if !processAlive(p.PID) {
		return false
	}
	// Сверяем время старта, если оно было известно.
	if !p.StartTime.IsZero() {
		st, err := processStartTime(p.PID)
		if err != nil || !st.Equal(p.StartTime) {
			return false
		}
	}
	// Проверяем, что процесс всё ещё слушает один из портов диапазона.
	cur, err := findListeningProcesses(from, to)
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

// ---------------------------------------------------------------------------
// Поиск процессов
// ---------------------------------------------------------------------------

// findListeningProcesses — платформенно-зависимая часть вынесена по файлам.
// В этом файле только диспетчер.
func findListeningProcesses(from, to uint16) ([]listeningProcess, error) {
	var procs []listeningProcess

	if isLinux() {
		p, err := linuxListeningProcesses(from, to)
		if err == nil {
			procs = p
		} else {
			// Фолбэк на lsof, если /proc недоступен.
			p, err2 := lsofListeningProcesses(from, to)
			if err2 != nil {
				return nil, fmt.Errorf("proc: %v; lsof: %w", err, err2)
			}
			procs = p
		}
	} else { // darwin
		p, err := darwinListeningProcesses(from, to)
		if err == nil {
			procs = p
		} else {
			p, err2 := lsofListeningProcesses(from, to)
			if err2 != nil {
				return nil, fmt.Errorf("netstat: %v; lsof: %w", err, err2)
			}
			procs = p
		}
	}
	return dedupe(procs), nil
}

func dedupe(in []listeningProcess) []listeningProcess {
	seen := map[[2]int]struct{}{}
	out := in[:0]
	for _, p := range in {
		k := [2]int{p.PID, int(p.Port)}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, p)
	}
	return out
}

func isLinux() bool {
	// runtime.GOOS == "linux" — но файл собирается и под darwin,
	// поэтому используем runtime.
	return runtimeGOOS() == "linux"
}

// ---------------------------------------------------------------------------
// Linux: /proc/net/tcp + /proc/<pid>/fd
// ---------------------------------------------------------------------------

func linuxListeningProcesses(from, to uint16) ([]listeningProcess, error) {
	// 1. Собираем inode сокетов, слушающих нужные порты.
	//    Формат /proc/net/tcp: local_address — hex "IP:PORT", st = 0A (LISTEN).
	inodes, err := linuxListeningInodes(from, to)
	if err != nil {
		return nil, err
	}
	if len(inodes) == 0 {
		return nil, nil
	}

	// 2. Ищем PID, у которых есть fd, ссылающийся на нужный inode.
	inodeToPIDs, err := linuxInodeToPIDs(inodes)
	if err != nil {
		return nil, err
	}

	// 3. Собираем результат.
	var out []listeningProcess
	for inode, pids := range inodeToPIDs {
		port := inodes[inode]
		for _, pid := range pids {
			st, _ := processStartTime(pid)
			out = append(out, listeningProcess{PID: pid, Port: port, StartTime: st})
		}
	}
	return out, nil
}

// linuxListeningInodes возвращает map[inode]port для LISTEN-сокетов в диапазоне.
func linuxListeningInodes(from, to uint16) (map[uint64]uint16, error) {
	res := map[uint64]uint16{}
	for _, path := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for i, line := range strings.Split(string(data), "\n") {
			if i == 0 || strings.TrimSpace(line) == "" {
				continue // заголовок или пустая строка
			}
			fields := strings.Fields(line)
			if len(fields) < 10 {
				continue
			}
			// st == 0A → LISTEN
			if fields[3] != "0A" {
				continue
			}
			_, port, ok := parseHexAddr(fields[1])
			if !ok || port < from || port > to {
				continue
			}
			inode, err := strconv.ParseUint(fields[9], 10, 64)
			if err != nil || inode == 0 {
				continue
			}
			res[inode] = port
		}
	}
	return res, nil
}

// parseHexAddr разбирает "0100007F:1F90" → (ip, 8080).
func parseHexAddr(s string) (string, uint16, bool) {
	i := strings.LastIndex(s, ":")
	if i < 0 {
		return "", 0, false
	}
	p, err := strconv.ParseUint(s[i+1:], 16, 32)
	if err != nil {
		return "", 0, false
	}
	return s[:i], uint16(p), true
}

// linuxInodeToPIDs проходит по /proc/<pid>/fd и сопоставляет inode → PID.
func linuxInodeToPIDs(want map[uint64]uint16) (map[uint64][]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	res := map[uint64][]int{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		fdDir := filepath.Join("/proc", e.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue // процесс умер или нет прав
		}
		for _, fd := range fds {
			target, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			// target вида "socket:[12345]"
			if !strings.HasPrefix(target, "socket:[") {
				continue
			}
			inodeStr := strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]")
			inode, err := strconv.ParseUint(inodeStr, 10, 64)
			if err != nil {
				continue
			}
			if _, ok := want[inode]; ok {
				res[inode] = append(res[inode], pid)
			}
		}
	}
	return res, nil
}

// ---------------------------------------------------------------------------
// macOS: netstat -anv -p tcp
// ---------------------------------------------------------------------------

func darwinListeningProcesses(from, to uint16) ([]listeningProcess, error) {
	// netstat на macOS с -v печатает PID в последней колонке для LISTEN.
	// Пример строки:
	// tcp4  0  0  *.8080   *.*  LISTEN  12345
	out, err := exec.Command("netstat", "-anv", "-p", "tcp").Output()
	if err != nil {
		return nil, fmt.Errorf("netstat: %w", err)
	}

	var procs []listeningProcess
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}
		if !strings.HasPrefix(fields[0], "tcp") {
			continue
		}
		if fields[len(fields)-2] != "LISTEN" {
			continue
		}
		_, port, ok := extractPortFromLocal(fields[3])
		if !ok || port < from || port > to {
			continue
		}
		pid, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil || pid <= 0 {
			continue
		}
		st, _ := processStartTime(pid)
		procs = append(procs, listeningProcess{PID: pid, Port: port, StartTime: st})
	}
	return procs, nil
}

func extractPortFromLocal(s string) (string, uint16, bool) {
	i := strings.LastIndex(s, ".")
	if i < 0 || i == len(s)-1 {
		return "", 0, false
	}
	p, err := strconv.Atoi(s[i+1:])
	if err != nil || p <= 0 || p > 65535 {
		return "", 0, false
	}
	return s[:i], uint16(p), true
}

// ---------------------------------------------------------------------------
// Фолбэк: lsof (есть и на Linux, и на macOS)
// ---------------------------------------------------------------------------

func lsofListeningProcesses(from, to uint16) ([]listeningProcess, error) {
	out, err := exec.Command("lsof", "-nP", "-iTCP", "-sTCP:LISTEN").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("lsof: %w", err)
	}
	var procs []listeningProcess
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		pid, err := strconv.Atoi(fields[1])
		if err != nil || pid <= 0 {
			continue
		}
		_, port, ok := extractPortFromLocal(fields[8])
		if !ok || port < from || port > to {
			continue
		}
		st, _ := processStartTime(pid)
		procs = append(procs, listeningProcess{PID: pid, Port: port, StartTime: st})
	}
	return procs, nil
}
