//go:build linux || darwin

package portkiller

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func osGetpid() int { return os.Getpid() }

// --- Интеграционный тест: graceful shutdown по SIGTERM ---

// TestKill_GracefulShutdown запускает дочерний процесс, который слушает TCP-порт
// и корректно обрабатывает SIGTERM (завершается за ~100 мс).
// Ожидаем: отчёт с Terminated=1, Killed=0.
func TestKill_GracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getpid() == 1 {
		t.Skip("running as pid 1 in container, cannot fork")
	}

	port := pickFreePort(t)

	// Дочерний процесс: слушает порт, ловит SIGTERM, завершается.
	script := `
package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:` + itoa(int(port)) + `")
	if err != nil {
		os.Exit(1)
	}
	defer l.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)
	<-ch
	time.Sleep(100 * time.Millisecond) // имитация graceful cleanup
	os.Exit(0)
}
`
	cmd := runGoChild(t, script)

	// Ждём, пока дочерний процесс займёт порт.
	waitForListener(t, port, 3*time.Second)

	rep, err := KillProcessesInPortRange(port, port, 3*time.Second)
	if err != nil {
		t.Fatalf("KillProcessesInPortRange: %v", err)
	}
	if rep.Terminated != 1 {
		t.Fatalf("expected Terminated=1, got %+v", rep)
	}
	if rep.Killed != 0 {
		t.Fatalf("expected Killed=0, got %+v", rep)
	}
	if rep.Failed != 0 {
		t.Fatalf("expected Failed=0, got %+v", rep)
	}

	_ = cmd.Wait()
}

// TestKill_ForceKillByTimeout запускает процесс, который игнорирует SIGTERM.
// Ожидаем: Terminated=0, Killed=1 (сработал SIGKILL после таймаута).
func TestKill_ForceKillByTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getpid() == 1 {
		t.Skip("running as pid 1 in container, cannot fork")
	}

	port := pickFreePort(t)

	script := `
package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:` + itoa(int(port)) + `")
	if err != nil {
		os.Exit(1)
	}
	defer l.Close()

	// Игнорируем SIGTERM.
	signal.Ignore(syscall.SIGTERM)
	select {}
}
`
	cmd := runGoChild(t, script)

	waitForListener(t, port, 3*time.Second)

	start := time.Now()
	rep, err := KillProcessesInPortRange(port, port, 1*time.Second)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("KillProcessesInPortRange: %v", err)
	}
	if rep.Killed != 1 {
		t.Fatalf("expected Killed=1, got %+v", rep)
	}
	if rep.Terminated != 0 {
		t.Fatalf("expected Terminated=0, got %+v", rep)
	}
	// Таймаут 1s, должно занять не меньше ~1s и не сильно больше.
	if elapsed < 900*time.Millisecond {
		t.Fatalf("killed too fast: %v, expected ~1s timeout", elapsed)
	}
	if elapsed > 3*time.Second {
		t.Fatalf("took too long: %v", elapsed)
	}

	_ = cmd.Wait()
}

// TestKill_TimeoutZero отправляет только SIGTERM без ожидания.
// Процесс, игнорирующий SIGTERM, должен остаться жив.
func TestKill_TimeoutZero(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getpid() == 1 {
		t.Skip("running as pid 1 in container, cannot fork")
	}

	port := pickFreePort(t)
	script := `
package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:` + itoa(int(port)) + `")
	if err != nil {
		os.Exit(1)
	}
	defer l.Close()
	signal.Ignore(syscall.SIGTERM)
	select {}
}
`
	cmd := runGoChild(t, script)
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	waitForListener(t, port, 3*time.Second)

	rep, err := KillProcessesInPortRange(port, port, 0)
	if err != nil {
		t.Fatalf("KillProcessesInPortRange: %v", err)
	}
	// Процесс получил SIGTERM, но не завершился; отчёт: Terminated=1 (мы не ждали).
	if rep.Terminated != 1 {
		t.Fatalf("expected Terminated=1, got %+v", rep)
	}
	// Убеждаемся, что процесс жив.
	time.Sleep(200 * time.Millisecond)
	if !processAlive(cmd.Process.Pid) {
		t.Fatal("process should still be alive after SIGTERM-only")
	}
}

// --- Вспомогательные функции ---

func pickFreePort(t *testing.T) uint16 {
	t.Helper()
	port, cleanup := freePort(t)
	cleanup() // освобождаем порт, чтобы дочерний процесс мог его занять
	return port
}

// runGoChild компилирует и запускает переданный исходник как отдельный процесс.
func runGoChild(t *testing.T, src string) *exec.Cmd {
	t.Helper()

	dir := t.TempDir()
	srcPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcPath, []byte(src), 0o600); err != nil {
		t.Fatalf("write child source: %v", err)
	}

	binPath := filepath.Join(dir, "child")
	build := exec.Command("go", "build", "-o", binPath, srcPath)
	build.Env = filterEnv(os.Environ(), "GOROOT")
	var buildErr bytes.Buffer
	build.Stderr = &buildErr
	if err := build.Run(); err != nil {
		t.Fatalf("build child: %v\nstderr:\n%s", err, buildErr.String())
	}
	t.Logf("built child: %s", binPath)

	cmd := exec.Command(binPath)
	var childOut bytes.Buffer
	cmd.Stdout = &childOut
	cmd.Stderr = &childOut
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	t.Logf("child started: pid=%d", cmd.Process.Pid)

	// Логируем смерть ребёнка, если она случится до конца теста.
	exitCh := make(chan error, 1)
	go func() {
		exitCh <- cmd.Wait()
	}()
	go func() {
		err := <-exitCh
		t.Logf("child %d exited: %v\nchild output:\n%s",
			cmd.Process.Pid, err, childOut.String())
	}()

	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		select {
		case <-exitCh:
		case <-time.After(2 * time.Second):
		}
	})
	return cmd
}

func filterEnv(env []string, key string) []string {
	out := env[:0]
	for _, e := range env {
		if !strings.HasPrefix(e, key+"=") {
			out = append(out, e)
		}
	}
	return out
}

// waitForListener ждёт, пока порт начнёт принимать соединения.
func waitForListener(t *testing.T, port uint16, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		procs, err := findListeningProcesses(port, port)
		if err == nil && len(procs) > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("listener on port %d did not appear within %v", port, timeout)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
