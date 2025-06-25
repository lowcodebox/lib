package lib

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
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
}

func NewProcessMonitor(logPath string) *ProcessMonitor {
	return &ProcessMonitor{
		logPath: logPath,
		done:    make(chan struct{}),
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

func (pm *ProcessMonitor) Close() {
	close(pm.done)
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

func shellEscape(s string) string {
	escaped := strings.ReplaceAll(s, "'", "'\"'\"'")
	return "'" + escaped + "'"
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
func RunProcess(path, project, service, config, command, mode, dc, port string) (pid int, err error) {
	// Создаем контекст с таймаутом для проверки запуска
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
	auditLog := fmt.Sprintf("[%s] [AUDIT] Process start attempt: path=%s, project=%s, service=%s, args=%v\n",
		time.Now().Format("2006-01-02 15:04:05"), path, project, service, args)
	appendToFile(auditPath, auditLog)

	// Создаем команду с безопасным bash wrapper
	cmd, err := createSecureCommand(ctx, path, args, logPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create secure command: %w", err)
	}

	// Настраиваем процесс
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Запускаем процесс
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("unable to start process: config=%s, path=%s, command=%s, mode=%s, dc=%s, err=%w",
			config, path, command, mode, dc, err)
	}

	pid = cmd.Process.Pid
	monitor.SetProcess(cmd.Process)

	// Логируем успешный запуск
	successLog := fmt.Sprintf("[%s] [AUDIT] Process started successfully: PID=%d\n",
		time.Now().Format("2006-01-02 15:04:05"), pid)
	appendToFile(auditPath, successLog)

	// Запускаем мониторинг процесса
	processExited := make(chan *ProcessStartupError, 1)
	go monitorProcess(cmd, monitor, processExited)

	// Ждем либо успешного запуска, либо ошибки, либо таймаута
	select {
	case <-ctx.Done():
		// Таймаут - процесс запустился успешно
		return pid, nil

	case startupErr := <-processExited:
		// Процесс завершился с ошибкой в течение 5 секунд
		endLog := fmt.Sprintf("[%s] [AUDIT] Process failed during startup: PID=%d, exit_code=%d\n",
			time.Now().Format("2006-01-02 15:04:05"), pid, startupErr.ExitCode)
		appendToFile(auditPath, endLog)

		return pid, startupErr
	}
}

// Мониторинг процесса
func monitorProcess(cmd *exec.Cmd, monitor *ProcessMonitor, errorChan chan<- *ProcessStartupError) {
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

	// Читаем последнюю строку из лога
	lastLine := readLastLogLine(monitor.logPath)
	monitor.UpdateLastLogLine(lastLine)

	// Если процесс завершился с ошибкой, отправляем информацию об ошибке
	if exitCode != 0 {
		startupErr := &ProcessStartupError{
			ExitCode:    exitCode,
			LastLogLine: lastLine,
			FullError:   fmt.Sprintf("process exited with code %d", exitCode),
			ProcessPID:  cmd.Process.Pid,
		}

		select {
		case errorChan <- startupErr:
		default:
		}
	}
}

// Чтение последней строки из лога
func readLastLogLine(logPath string) string {
	file, err := os.Open(logPath)
	if err != nil {
		return fmt.Sprintf("failed to read log: %v", err)
	}
	defer file.Close()

	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lastLine = line
		}
	}

	if lastLine == "" {
		return "no log output found"
	}

	return lastLine
}

// Вспомогательная функция для добавления в файл
func appendToFile(path, content string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	file.WriteString(content)
}

// Создание безопасной команды с bash wrapper
func createSecureCommand(ctx context.Context, path string, args []string, logPath string) (*exec.Cmd, error) {
	// Экранируем все параметры
	escapedPath := shellEscape(path)
	escapedLogPath := shellEscape(logPath)
	escapedArgs := make([]string, len(args))
	for i, arg := range args {
		escapedArgs[i] = shellEscape(arg)
	}

	// Создаем безопасный bash скрипт
	bashScript := fmt.Sprintf(`
set -euo pipefail
LOG_PATH=%s
echo "[$(date '+%%Y-%%m-%%d %%H:%%M:%%S')] [INIT] Starting process: %s %s" >> "$LOG_PATH"
echo "[$(date '+%%Y-%%m-%%d %%H:%%M:%%S')] [START] Process starting" >> "$LOG_PATH"

# Запускаем процесс в фоне и получаем его PID
%s %s >> "$LOG_PATH" 2>&1 &
CHILD_PID=$!
echo "CHILD_PID:$CHILD_PID" >> "$LOG_PATH"

# Ждем завершения дочернего процесса
wait $CHILD_PID
EXIT_CODE=$?

echo "[$(date '+%%Y-%%m-%%d %%H:%%M:%%S')] [END] Process finished with exit code: $EXIT_CODE" >> "$LOG_PATH"
exit $EXIT_CODE
`, escapedLogPath, escapedPath, strings.Join(escapedArgs, " "), escapedPath, strings.Join(escapedArgs, " "))

	// Создаем команду
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", bashScript)

	// Устанавливаем безопасные переменные окружения
	cmd.Env = append(os.Environ(),
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"IFS=' \t\n'",
		"BASH_ENV=",
		"ENV=",
	)

	return cmd, nil
}
