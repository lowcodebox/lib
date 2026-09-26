package portkiller

import (
	"fmt"
	"time"
)

// KillReport описывает результат работы.
type KillReport struct {
	Found      int // процессов найдено на портах в диапазоне
	Terminated int // завершились по SIGTERM (graceful)
	Killed     int // пришлось добить SIGKILL
	Failed     int // не удалось завершить
	Skipped    int // пропущено из-за TOCTOU-проверки
	Errors     []error
}

// KillProcessesInPortRange завершает все процессы, слушающие TCP-порты
// в диапазоне [from, to] включительно.
//
// Сначала каждому процессу отправляется SIGTERM (SIGBREAK/CTRL_BREAK на Windows).
// Если процесс не завершился за timeout — отправляется SIGKILL (TerminateProcess на Windows).
// Если timeout <= 0 — используется только SIGTERM.
func KillProcessesInPortRange(from, to uint16, timeout time.Duration) (KillReport, error) {
	if from > to {
		return KillReport{}, fmt.Errorf("invalid port range: %d > %d", from, to)
	}
	return killProcessesInPortRange(from, to, timeout)
}

// listeningProcess — описание процесса, слушающего порт.
// startTime используется для защиты от TOCTOU: PID может быть переиспользован ОС.
type listeningProcess struct {
	PID       int
	Port      uint16
	StartTime time.Time // нулевое значение = неизвестно
}
