package lib

import "net"

// получает имя агента по IPC из сокета, открытого агентом при запуске
func GetAgentHostname() (string, error) {
	socketPath := "/var/run/agent.sock"
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	return string(buf[:n]), nil
}
