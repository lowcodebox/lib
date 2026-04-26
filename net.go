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
func GetPIDByPort(port int) ([]int, error) {
	switch runtime.GOOS {
	case "linux":
		return getPIDByPortLinux(port)
	case "darwin":
		return getPIDByPortDarwin(port)
	case "windows":
		return getPIDByPortWindows(port)
	default:
		return []int{0}, fmt.Errorf("unsupport OS: %s", runtime.GOOS)
	}
}

// getPIDByPortLinux для Linux
func getPIDByPortLinux(port int) ([]int, error) {
	// Используем команду ss или netstat
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ss -lpn | grep :%d", port))
	output, err := cmd.Output()
	if err != nil {
		return []int{0}, fmt.Errorf("process is not found: %w", err)
	}

	return parsePIDFromOutputLinux(string(output))
}

// getPIDByPortDarwin для macOS
func getPIDByPortDarwin(port int) (res []int, err error) {
	// Используем команду lsof
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t")
	output, err := cmd.Output()
	if err != nil {
		return []int{0}, fmt.Errorf("process is not found: %w", err)
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return []int{0}, fmt.Errorf("process is not found")
	}

	return parsePIDFromOutputMac(pidStr)
}

// getPIDByPortWindows для Windows
func getPIDByPortWindows(port int) (res []int, err error) {
	// Используем команду netstat
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return res, fmt.Errorf("not exec netstat: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	portStr := fmt.Sprintf(":%d", port)

	for _, line := range lines {
		if strings.Contains(line, "LISTENING") && strings.Contains(line, portStr) {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				pid, err := strconv.Atoi(fields[4])
				if err != nil {
					return res, fmt.Errorf("parse pid (%s) failed, err: %w", fields[4], err)
				}
				res = append(res, pid)
			}
		}
	}

	return res, nil
}

// parsePIDFromOutputMac парсит PID из вывода ss
func parsePIDFromOutputMac(output string) (res []int, err error) {
	lines := strings.Split(output, "\n")
	for _, pidStr := range lines {
		p, err := strconv.Atoi(pidStr)
		if err != nil {
			return res, fmt.Errorf("parse pid (%s) failed, err: %w", pidStr, err)
		}
		res = append(res, p)
	}

	return res, nil
}

// parsePIDFromOutputLinux парсит PID из вывода ss
func parsePIDFromOutputLinux(output string) (res []int, err error) {
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
					p, err := strconv.Atoi(pidStr)
					if err != nil {
						return res, fmt.Errorf("parse pid (%s) failed, err: %w", pidStr, err)
					}
					res = append(res, p)
				}
			}
		}
	}
	return res, nil
}
