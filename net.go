package lib

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	gopsutil "github.com/shirou/gopsutil/v3/net"
	gopsutilprocess "github.com/shirou/gopsutil/v3/process"
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

// GetPIDByPort возвращает PID процесса, слушающего указанный порт
func GetPIDByPort(port int) (int, error) {
	// Получаем все TCP соединения
	connections, err := gopsutil.Connections("tcp")
	if err != nil {
		return 0, fmt.Errorf("failed to get connection list: %w", err)
	}

	portStr := strconv.Itoa(port)

	for _, conn := range connections {
		// Проверяем, что соединение в состоянии LISTEN
		if conn.Status == "LISTEN" {
			// Получаем порт из локального адреса
			_, localPort, err := net.SplitHostPort(conn.Laddr.String())
			if err != nil {
				continue
			}

			if localPort == portStr {
				// Если есть PID, возвращаем его
				if conn.Pid > 0 {
					return int(conn.Pid), nil
				}
			}
		}
	}

	return 0, fmt.Errorf("process listening on port %d was not found", port)
}

type ProcessInfo struct {
	PID     int32
	Name    string
	Port    int
	Status  string
	Exe     string
	Cmdline string
}

// GetProcessByPort возвращает информацию о процессе по порту
func GetProcessByPort(port int) (*ProcessInfo, error) {
	// Получаем все TCP соединения
	connections, err := gopsutil.Connections("tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection list: %w", err)
	}

	portStr := strconv.Itoa(port)

	for _, conn := range connections {
		if conn.Status == "LISTEN" {
			_, localPort, err := net.SplitHostPort(conn.Laddr.String())
			if err != nil {
				continue
			}

			if localPort == portStr && conn.Pid > 0 {
				// Получаем информацию о процессе
				proc, err := gopsutilprocess.NewProcess(conn.Pid)
				if err != nil {
					return nil, fmt.Errorf("failed to get information about process: %w", err)
				}

				name, _ := proc.Name()
				exe, _ := proc.Exe()
				cmdline, _ := proc.Cmdline()

				return &ProcessInfo{
					PID:     conn.Pid,
					Name:    name,
					Port:    port,
					Status:  conn.Status,
					Exe:     exe,
					Cmdline: cmdline,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("process listening on port %d was not found", port)
}
