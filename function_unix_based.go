package lib

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type ProcessConfig struct {
	Path    string
	Project string
	Service string
	Config  string
	Command string
	Mode    string
	DC      string
	Port    string
}

func (pc *ProcessConfig) Validate() error {
	if pc.Config == "" {
		return errors.New("config file not specified")
	}
	if pc.Command == "" {
		pc.Command = "start"
	}
	pc.Path = strings.ReplaceAll(pc.Path, "//", "/")
	return nil
}

func (pc *ProcessConfig) BuildArgs() []string {
	args := []string{pc.Command, "--config", pc.Config}

	if pc.Mode != "" {
		args = append(args, "--mode", pc.Mode)
	}
	if pc.DC != "" {
		args = append(args, "--dc", pc.DC)
	}
	if pc.Port != "" {
		args = append(args, "--port", pc.Port)
	}

	return args
}

func (pc *ProcessConfig) setupDebugLogging(cmd *exec.Cmd) error {
	if pc.Mode != "debug" {
		return nil
	}

	dirPath := filepath.Join("debug")
	if err := CreateDir(dirPath, 0755); err != nil {
		return fmt.Errorf("unable create directory for debug file, path: %s, err: %w", dirPath, err)
	}

	logWriter := &lumberjack.Logger{
		Filename: filepath.Join("debug", fmt.Sprintf("%s-%s.log", pc.Project, pc.Service)),
		MaxSize:  10, // мегабайты
		Compress: false,
	}

	// Файл будет закрыт когда процесс завершится
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	return nil
}

func waitForProcessStart(ctx context.Context, cmd *exec.Cmd) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if cmd.ProcessState != nil {
				exitCode := cmd.ProcessState.ExitCode()
				if exitCode >= 0 {
					return fmt.Errorf("process exited unexpectedly with code: %d", exitCode)
				}
			}
			// Если ProcessState == nil, процесс все еще работает
			return nil
		}
	}
}

func RunProcess(path, project, service, config, command, mode, dc, port string) (pid int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pc := &ProcessConfig{
		Path:    path,
		Project: project,
		Service: service,
		Config:  config,
		Command: command,
		Mode:    mode,
		DC:      dc,
		Port:    port,
	}

	if err := pc.Validate(); err != nil {
		return 0, err
	}

	cmd := exec.CommandContext(ctx, pc.Path, pc.BuildArgs()...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := pc.setupDebugLogging(cmd); err != nil {
		return 0, err
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("unable start process, config: %s, path: %s, command: %s, mode: %s, dc: %s, err: %w",
			config, path, command, mode, dc, err)
	}

	pid = cmd.Process.Pid

	// Запускаем горутину для ожидания завершения процесса
	go func() {
		cmd.Wait()
	}()

	// Ждем немного чтобы убедиться что процесс запустился успешно
	if err := waitForProcessStart(ctx, cmd); err != nil {
		return pid, err
	}

	return pid, nil
}
