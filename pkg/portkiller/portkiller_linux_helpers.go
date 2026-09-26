//go:build linux

package portkiller

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// processStartTime читает время старта процесса из /proc/<pid>/stat
// (поле 22 — starttime в тиках после загрузки) и переводит в time.Time.
// Точность до секунды — этого достаточно для защиты от переиспользования PID.
func processStartTime(pid int) (time.Time, error) {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return time.Time{}, err
	}
	// Поле comm может содержать пробелы и скобки — берём всё после последней ')'.
	s := string(data)
	i := strings.LastIndex(s, ")")
	if i < 0 || i+2 >= len(s) {
		return time.Time{}, os.ErrInvalid
	}
	fields := strings.Fields(s[i+2:])
	// После comm: state(3) ppid(4) ... starttime(22).
	// В fields индексация с 0 для state. starttime = fields[22-3] = fields[19].
	if len(fields) < 20 {
		return time.Time{}, os.ErrInvalid
	}
	ticks, err := strconv.ParseUint(fields[19], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	// Переводим тики в секунды. Стандартный HZ = 100, но может отличаться.
	// Для защиты от TOCTOU достаточно секундной точности.
	startSec := ticks / 100
	return time.Unix(int64(startSec), 0), nil
}

func runtimeGOOS() string { return "linux" }
