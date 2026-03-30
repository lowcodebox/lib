package lib

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// GetOutboundIP получение текущего IP-адреса компьютера
// publicPingHost - хост, через который делаем соединение
// если подключаемся = получим наш IP
// по-умолчанию DNS Google
func GetOutboundIP(publicPingHost string) (net.IP, error) {
	if publicPingHost == "" {
		publicPingHost = "8.8.8.8:80"
	}
	conn, err := net.Dial("udp", publicPingHost)
	if err != nil {
		return nil, fmt.Errorf("connect to %s failed, err: %w", publicPingHost, err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

// CheckPort проверяет доступность UDP порта
// true - порт занят
// false - порт свободен (нет соединения)
// network: tcp/udp
func CheckPort(network string, host string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", host, port)

	network = strings.ToLower(network)
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Для UDP отправляем тестовые данные
	if network == "udp" {
		_, err = conn.Write([]byte("ping"))
		if err != nil {
			return false
		}

		// Устанавливаем таймаут на чтение
		err = conn.SetReadDeadline(time.Now().Add(timeout))
		if err != nil {
			return false
		}

		// Пытаемся прочитать ответ
		buf := make([]byte, 1024)
		_, err = conn.Read(buf)

		// Для UDP отсутствие ответа не всегда означает недоступность
		// Возвращаем true, если соединение установлено успешно
	}

	return true
}

// GetPIDByPort возвращает PID процесса, слушающего указанный порт (кросс-платформенная версия)
func GetPIDByPort(port int) (int, error) {
	switch runtime.GOOS {
	case "linux":
		return getPIDByPortLinux(port)
	case "darwin":
		return getPIDByPortDarwin(port)
	case "windows":
		return getPIDByPortWindows(port)
	default:
		return 0, fmt.Errorf("неподдерживаемая ОС: %s", runtime.GOOS)
	}
}

// getPIDByPortLinux для Linux
func getPIDByPortLinux(port int) (int, error) {
	// Используем команду ss или netstat
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ss -lpn | grep :%d", port))
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("процесс не найден: %w", err)
	}

	return parsePIDFromOutput(string(output))
}

// getPIDByPortDarwin для macOS
func getPIDByPortDarwin(port int) (int, error) {
	// Используем команду lsof
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("процесс не найден: %w", err)
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return 0, fmt.Errorf("процесс не найден")
	}

	return strconv.Atoi(pidStr)
}

// getPIDByPortWindows для Windows
func getPIDByPortWindows(port int) (int, error) {
	// Используем команду netstat
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("не удалось выполнить netstat: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	portStr := fmt.Sprintf(":%d", port)

	for _, line := range lines {
		if strings.Contains(line, "LISTENING") && strings.Contains(line, portStr) {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				pid, err := strconv.Atoi(fields[4])
				if err == nil {
					return pid, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("процесс, слушающий порт %d, не найден", port)
}

// parsePIDFromOutput парсит PID из вывода ss
func parsePIDFromOutput(output string) (int, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "pid=") {
			// Ищем pid=12345
			start := strings.Index(line, "pid=")
			if start != -1 {
				start += 4
				end := strings.IndexAny(line[start:], ",)")
				if end != -1 {
					pidStr := line[start : start+end]
					return strconv.Atoi(pidStr)
				}
			}
		}
	}
	return 0, fmt.Errorf("PID не найден")
}
