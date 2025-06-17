//go:build !windows

package lib

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// RunProcess стартуем сервис из конфига
func RunProcess(path, domain, config, command, mode, dc, port string) (pid int, err error) {
	var cmd *exec.Cmd
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if config == "" {
		return 0, errors.New("config file not specified")
	}
	if command == "" {
		command = "start"
	}

	path = strings.Replace(path, "//", "/", -1)

	args := []string{command, "--config", config}

	if mode != "" {
		args = append(args, "--mode", mode)
	}
	if dc != "" {
		args = append(args, "--dc", dc)
	}
	if port != "" {
		args = append(args, "--port", port)
	}

	cmd = exec.Command(path, args...)

	if mode == "debug" {
		dirPath := "debug" + sep + domain
		err = CreateDir(dirPath, 0777)
		if err != nil {
			return 0, fmt.Errorf("unable create directory for debug file, path: %s, err: %w", dirPath, err)
		}

		filePath := "debug" + sep + domain + ".log"
		f, err := os.Create(filePath)
		if err != nil {
			return 0, fmt.Errorf("unable create debug file, path: %s, err: %w", filePath, err)
		}
		cmd.Stdout = f
		cmd.Stderr = f
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("unable start process, status: %d, config: %s, path: %s, command: %s, mode: %s, dc: %s, err: %w",
			cmd.ProcessState.ExitCode(), config, path, command, mode, dc, err)
	}

	go cmd.Process.Wait()

	pid = cmd.Process.Pid

	// в течение заданного интервала ожидаем завершающий статус запуска
	// или выходим если -1 (в процессе или прибит сигналом)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		exitCode := cmd.ProcessState.ExitCode()

		// завершился
		if exitCode >= 0 {
			return
		}

		select {
		case <-ctx.Done():
			return

		default:
			// -1 — работает или прибит сигналом
		}
	}

	return
}
