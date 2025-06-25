package lib

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// ProcessStartupError содержит детальную информацию об ошибке запуска
type ProcessStartupError struct {
	ExitCode    int
	LastLogLine string
	FullError   string
	ProcessPID  int
}

func (e *ProcessStartupError) Error() string {
	return fmt.Sprintf("process failed to start: exit_code=%d, pid=%d, last_log=%s, error=%s",
		e.ExitCode, e.ProcessPID, e.LastLogLine, e.FullError)
}

// ProcessMonitor отслеживает состояние процесса и логи
type ProcessMonitor struct {
	logPath     string
	process     *os.Process
	lastLogLine string
	exitCode    int
	mutex       sync.RWMutex
	done        chan struct{}
	logger      *lumberjack.Logger
}

func NewProcessMonitor(logPath string) *ProcessMonitor {
	logger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    2, // 1 MB
		MaxBackups: 1,
		MaxAge:     7, // days
		Compress:   false,
	}

	return &ProcessMonitor{
		logPath: logPath,
		done:    make(chan struct{}),
		logger:  logger,
	}
}

func (pm *ProcessMonitor) SetProcess(process *os.Process) {
	pm.mutex.Lock()
	pm.process = process
	pm.mutex.Unlock()
}

func (pm *ProcessMonitor) GetLastLogLine() string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.lastLogLine
}

func (pm *ProcessMonitor) GetExitCode() int {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	return pm.exitCode
}

func (pm *ProcessMonitor) UpdateLastLogLine(line string) {
	pm.mutex.Lock()
	pm.lastLogLine = strings.TrimSpace(line)
	pm.mutex.Unlock()
}

func (pm *ProcessMonitor) SetExitCode(code int) {
	pm.mutex.Lock()
	pm.exitCode = code
	pm.mutex.Unlock()
}

func (pm *ProcessMonitor) WriteLog(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)
	pm.logger.Write([]byte(logEntry))
}

func (pm *ProcessMonitor) Close() {
	close(pm.done)
	if pm.logger != nil {
		pm.logger.Close()
	}
}

// Безопасная валидация
func validateSecurityConstraints(path string, args []string, project, service string) error {
	// Валидация пути к исполняемому файлу
	if err := validateExecutablePath(path); err != nil {
		return fmt.Errorf("unsafe executable path: %w", err)
	}

	// Валидация аргументов
	for i, arg := range args {
		if err := validateArgument(arg); err != nil {
			return fmt.Errorf("unsafe argument at position %d (%s): %w", i, arg, err)
		}
	}

	// Валидация имен проекта и сервиса
	if err := validateName(project, "project"); err != nil {
		return err
	}
	if err := validateName(service, "service"); err != nil {
		return err
	}

	return nil
}

func validateExecutablePath(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal detected")
	}

	dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "{", "}", "<", ">", "\"", "'", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("dangerous character '%s' detected in path", char)
		}
	}

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("executable not found or not accessible: %w", err)
	}

	return nil
}

func validateArgument(arg string) error {
	dangerousPatterns := []string{
		";", "&", "|", "`", "$", "$(", "${", "&&", "||",
		"<", ">", ">>", "<<", "\n", "\r",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(arg, pattern) {
			return fmt.Errorf("potentially dangerous pattern '%s' detected", pattern)
		}
	}

	if strings.Contains(arg, "..") {
		return fmt.Errorf("path traversal detected")
	}

	return nil
}

func validateName(name, nameType string) error {
	namePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !namePattern.MatchString(name) {
		return fmt.Errorf("invalid %s name: must contain only alphanumeric, underscore, and dash characters", nameType)
	}
	return nil
}

func createSecureLogPath(project, service string) (string, error) {
	logFileName := fmt.Sprintf("%s-%s.log", project, service)
	logPath := filepath.Join("debug", logFileName)

	absLogPath, err := filepath.Abs(logPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	workDir, _ := os.Getwd()
	expectedPrefix := filepath.Join(workDir, "debug")
	if !strings.HasPrefix(absLogPath, expectedPrefix) {
		return "", fmt.Errorf("log path outside of allowed directory")
	}

	return absLogPath, nil
}

// Основная функция запуска процесса
// Основная функция запуска процесса
func RunProcess(path, project, service, config, command, mode, dc, port string) (pid int, err error) {
	// Создаем контекст только для проверки запуска (не для выполнения процесса)
	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()

	// Создаем конфигурацию процесса
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

	// Валидация конфигурации
	if err := pc.Validate(); err != nil {
		return 0, fmt.Errorf("process config validation failed: %w", err)
	}

	// Проверка безопасности
	args := pc.BuildArgs()
	if err := validateSecurityConstraints(path, args, project, service); err != nil {
		return 0, fmt.Errorf("security validation failed: %w", err)
	}

	// Создаем директорию для логов
	debugDir := filepath.Join("debug")
	if err := CreateDir(debugDir, 0755); err != nil {
		return 0, fmt.Errorf("unable to create debug directory: %w", err)
	}

	// Создаем безопасный путь к логу
	logPath, err := createSecureLogPath(project, service)
	if err != nil {
		return 0, fmt.Errorf("failed to create secure log path: %w", err)
	}

	// Создаем монитор процесса
	monitor := NewProcessMonitor(logPath)
	defer monitor.Close()

	// Записываем аудит лог
	auditPath := filepath.Join("debug", "audit.log")
	auditLogger := &lumberjack.Logger{
		Filename:   auditPath,
		MaxSize:    1, // 1 MB
		MaxBackups: 1,
		MaxAge:     30, // days
		Compress:   false,
	}
	defer auditLogger.Close()

	auditLog := fmt.Sprintf("[%s] [AUDIT] Process start attempt: path=%s, project=%s, service=%s, args=%v\n",
		time.Now().Format("2006-01-02 15:04:05"), path, project, service, args)
	auditLogger.Write([]byte(auditLog))

	// Создаем команду БЕЗ контекста для долгосрочного выполнения
	cmd := exec.Command(path, args...)

	// Настраиваем процесс
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Создаем pipe для захвата вывода
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Логируем начало запуска
	monitor.WriteLog(fmt.Sprintf("[INIT] Starting process: %s %s", path, strings.Join(args, " ")))

	// Запускаем процесс
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("unable to start process: config=%s, path=%s, command=%s, mode=%s, dc=%s, err=%w",
			config, path, command, mode, dc, err)
	}

	// Получаем PID реального процесса
	pid = cmd.Process.Pid
	monitor.SetProcess(cmd.Process)

	// Логируем успешный запуск
	monitor.WriteLog(fmt.Sprintf("[START] Process started with PID: %d", pid))
	successLog := fmt.Sprintf("[%s] [AUDIT] Process started successfully: PID=%d\n",
		time.Now().Format("2006-01-02 15:04:05"), pid)
	auditLogger.Write([]byte(successLog))

	// Запускаем горутины для чтения вывода
	go captureOutput(stdout, monitor, "STDOUT")
	go captureOutput(stderr, monitor, "STDERR")

	// Запускаем мониторинг процесса в фоне (не блокирующий)
	go monitorProcessBackground(cmd, monitor, auditLogger)

	// Ждем только успешного запуска (не завершения процесса)
	select {
	case <-time.After(3 * time.Second):
		// Проверяем, что процесс еще работает
		if isProcessRunning(pid) {
			monitor.WriteLog("[SUCCESS] Process startup completed successfully")
			return pid, nil
		} else {
			monitor.WriteLog("[ERROR] Process died shortly after startup")
			return pid, fmt.Errorf("process died shortly after startup")
		}

	case <-startupCtx.Done():
		// Таймаут контекста запуска
		if isProcessRunning(pid) {
			monitor.WriteLog("[SUCCESS] Process startup completed (timeout reached, but process is running)")
			return pid, nil
		} else {
			monitor.WriteLog("[TIMEOUT] Process startup timeout")
			return pid, fmt.Errorf("process startup timeout")
		}
	}
}

// Проверка, работает ли процесс
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// На Unix системах проверяем отправкой сигнала 0
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// Фоновый мониторинг процесса (не блокирует возврат из RunProcess)
func monitorProcessBackground(cmd *exec.Cmd, monitor *ProcessMonitor, auditLogger *lumberjack.Logger) {
	// Ждем завершения процесса
	err := cmd.Wait()

	// Получаем код выхода
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	monitor.SetExitCode(exitCode)
	monitor.WriteLog(fmt.Sprintf("[END] Process finished with exit code: %d", exitCode))

	// Логируем завершение в аудит
	endLog := fmt.Sprintf("[%s] [AUDIT] Process finished: PID=%d, exit_code=%d\n",
		time.Now().Format("2006-01-02 15:04:05"), cmd.Process.Pid, exitCode)
	auditLogger.Write([]byte(endLog))
}

// Захват вывода процесса
func captureOutput(reader io.Reader, monitor *ProcessMonitor, outputType string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			monitor.WriteLog(fmt.Sprintf("[%s] %s", outputType, line))
			monitor.UpdateLastLogLine(line)
		}
	}
}
