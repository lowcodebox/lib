package portkiller

import (
	"net"
	"runtime"
	"testing"
	"time"
)

// --- Юнит-тесты парсеров ---

func TestParseHexAddr(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in       string
		wantPort uint16
		wantOk   bool
	}{
		{"0100007F:1F90", 8080, true},                         // 127.0.0.1:8080
		{"00000000:0050", 80, true},                           // 0.0.0.0:80
		{"00000000000000000000000000000000:1F90", 8080, true}, // tcp6
		{"0100007F:", 0, false},                               // нет порта
		{"0100007F", 0, false},                                // нет двоеточия
		{"0100007F:ZZZZ", 0, false},                           // не число
		{"", 0, false},
	}
	for _, c := range cases {
		_, port, ok := parseHexAddr(c.in)
		if ok != c.wantOk || (ok && port != c.wantPort) {
			t.Errorf("parseHexAddr(%q) = (%d, %v), want (%d, %v)",
				c.in, port, ok, c.wantPort, c.wantOk)
		}
	}
}

func TestExtractPortFromLocal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in       string
		wantPort uint16
		wantOk   bool
	}{
		{"*:8080", 8080, true},
		{"127.0.0.1:8080", 8080, true},
		{"[::]:443", 443, true},
		{"*:0", 0, false},     // порт 0 невалиден
		{"*:99999", 0, false}, // > 65535
		{"*:", 0, false},      // нет порта
		{"*", 0, false},       // нет разделителя
		{"", 0, false},
	}
	for _, c := range cases {
		_, port, ok := extractPortFromLocal(c.in)
		if ok != c.wantOk || (ok && port != c.wantPort) {
			t.Errorf("extractPortFromLocal(%q) = (%d, %v), want (%d, %v)",
				c.in, port, ok, c.wantPort, c.wantOk)
		}
	}
}

func TestDedupe(t *testing.T) {
	t.Parallel()
	in := []listeningProcess{
		{PID: 1, Port: 80},
		{PID: 1, Port: 80}, // дубль
		{PID: 2, Port: 80},
		{PID: 1, Port: 81}, // другой порт — не дубль
	}
	out := dedupe(in)
	if len(out) != 3 {
		t.Fatalf("dedupe: got %d, want 3", len(out))
	}
}

// --- Валидация входных параметров ---

func TestKillProcessesInPortRange_InvalidRange(t *testing.T) {
	t.Parallel()
	if _, err := KillProcessesInPortRange(9000, 8000, time.Second); err == nil {
		t.Fatal("expected error for from > to")
	}
}

// --- Поиск слушающего процесса ---

// freePort поднимает TCP-listener на свободном порту, возвращает порт и функцию очистки.
func freePort(t *testing.T) (uint16, func()) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := uint16(l.Addr().(*net.TCPAddr).Port)
	return port, func() { _ = l.Close() }
}

func TestFindListeningProcesses_FindsCurrentProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	port, cleanup := freePort(t)
	defer cleanup()

	procs, err := findListeningProcesses(port, port)
	if err != nil {
		t.Fatalf("findListeningProcesses: %v", err)
	}
	myPID := osGetpid()
	found := false
	for _, p := range procs {
		if p.PID == myPID && p.Port == port {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("current process (pid=%d) listening on %d not found; got %+v",
			myPID, port, procs)
	}
}

func TestFindListeningProcesses_EmptyRange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	// Диапазон, в котором почти наверняка ничего не слушает.
	// 1..2 — привилегированные порты, обычно заняты, но мы проверяем не это,
	// а корректность возврата пустого результата без ошибки.
	procs, err := findListeningProcesses(1, 2)
	if err != nil {
		t.Fatalf("findListeningProcesses: %v", err)
	}
	for _, p := range procs {
		if p.Port >= 1 && p.Port <= 2 {
			// ок — значит кто-то слушает, это не ошибка
			t.Logf("unexpected listener on %d: pid=%d", p.Port, p.PID)
		}
	}
}

// --- TOCTOU-проверка ---

func TestStillSameListener_True(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	port, cleanup := freePort(t)
	defer cleanup()

	pid := osGetpid()
	st, err := processStartTime(pid)
	if err != nil {
		// На Windows processStartTime не реализован — пропускаем.
		t.Skipf("processStartTime not supported: %v", err)
	}

	p := listeningProcess{PID: pid, Port: port, StartTime: st}
	if !stillSameListener(p, port, port) {
		t.Fatal("expected stillSameListener=true for live listener")
	}
}

func TestStillSameListener_WrongPID(t *testing.T) {
	t.Parallel()
	// Несуществующий PID.
	p := listeningProcess{PID: 1 << 30, Port: 80}
	if stillSameListener(p, 80, 80) {
		t.Fatal("expected stillSameListener=false for dead pid")
	}
}

func TestStillSameListener_WrongPort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	port, cleanup := freePort(t)
	defer cleanup()

	pid := osGetpid()
	st, _ := processStartTime(pid)
	p := listeningProcess{PID: pid, Port: port, StartTime: st}

	// Процесс жив, но порт в запросе — другой.
	if stillSameListener(p, port+1, port+1) {
		t.Fatal("expected stillSameListener=false when port changed")
	}
}

// --- processAlive ---

func TestProcessAlive_Self(t *testing.T) {
	t.Parallel()
	if !processAlive(osGetpid()) {
		t.Fatal("current process should be alive")
	}
}

func TestProcessAlive_Dead(t *testing.T) {
	t.Parallel()
	if processAlive(1 << 30) {
		t.Fatal("pid 1<<30 should not exist")
	}
}

// --- Проверка, что рантайм-хелперы согласованы ---

func TestRuntimeGOOS(t *testing.T) {
	t.Parallel()
	if runtimeGOOS() != runtime.GOOS {
		t.Fatalf("runtimeGOOS()=%q, runtime.GOOS=%q", runtimeGOOS(), runtime.GOOS)
	}
}
