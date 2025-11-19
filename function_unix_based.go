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

	// Логируем начало запуска
	monitor.WriteLog(fmt.Sprintf("[INIT] Starting detached process: %s %s", path, strings.Join(args, " ")))

	// Запускаем процесс с двойным fork (полное отделение от родителя)
	pid, err = startDetachedProcess(path, args, logPath, monitor)
	if err != nil {
		return 0, fmt.Errorf("unable to start detached process: config=%s, path=%s, command=%s, mode=%s, dc=%s, err=%w",
			config, path, command, mode, dc, err)
	}

	// Логируем успешный запуск
	monitor.WriteLog(fmt.Sprintf("[START] Detached process started with PID: %d", pid))
	successLog := fmt.Sprintf("[%s] [AUDIT] Detached process started successfully: PID=%d\n",
		time.Now().Format("2006-01-02 15:04:05"), pid)
	auditLogger.Write([]byte(successLog))

	// Запускаем мониторинг логов в фоне (читаем из файла)
	go monitorDetachedProcessLogs(logPath, monitor)

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

// Мониторинг логов отделенного процесса (читаем из файла)
func monitorDetachedProcessLogs(logPath string, monitor *ProcessMonitor) {
	// Открываем файл для чтения
	file, err := os.Open(logPath)
	if err != nil {
		monitor.WriteLog(fmt.Sprintf("[ERROR] Failed to open log file for monitoring: %v", err))
		return
	}
	defer file.Close()

	// Переходим в конец файла
	file.Seek(0, io.SeekEnd)

	// Читаем новые строки по мере их появления
	scanner := bufio.NewScanner(file)
	for {
		if scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) != "" {
				monitor.UpdateLastLogLine(line)
			}
		} else {
			// Нет новых строк - ждем немного
			time.Sleep(100 * time.Millisecond)
			// Проверяем, не закрыт ли файл
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				break
			}
		}
	}
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

// Запуск процесса с двойным fork через bash-скрипт (полное отделение от родителя)
func startDetachedProcess(path string, args []string, logPath string, monitor *ProcessMonitor) (int, error) {
	// Путь к скрипту отделения
	detachScript := "/opt/lowcodebox/scripts/detach.sh"

	// Проверяем наличие скрипта
	if _, err := os.Stat(detachScript); err != nil {
		return 0, fmt.Errorf("detach script not found at %s: %w", detachScript, err)
	}

	// Формируем аргументы для скрипта: detach.sh <log_file> <executable> <args...>
	scriptArgs := []string{logPath, path}
	scriptArgs = append(scriptArgs, args...)

	monitor.WriteLog(fmt.Sprintf("[DETACH] Using script: %s", detachScript))
	monitor.WriteLog(fmt.Sprintf("[DETACH] Starting: %s %v", path, args))

	// Запускаем скрипт
	cmd := exec.Command(detachScript, scriptArgs...)

	// Скрипту не нужны pipes - он сам управляет потоками
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Запускаем
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start detach script: %w", err)
	}

	// Ждем завершения скрипта (он завершится быстро после двойного fork)
	if err := cmd.Wait(); err != nil {
		return 0, fmt.Errorf("detach script failed: %w", err)
	}

	monitor.WriteLog("[DETACH] Script completed, searching for process")

	// Даем время процессу запуститься
	time.Sleep(1 * time.Second)

	// Ищем PID запущенного процесса
	pid, err := findProcessByPath(path)
	if err != nil {
		// Пытаемся несколько раз с интервалом
		for i := 0; i < 5; i++ {
			monitor.WriteLog(fmt.Sprintf("[DETACH] Retry %d/5 finding process...", i+1))
			time.Sleep(500 * time.Millisecond)
			pid, err = findProcessByPath(path)
			if err == nil {
				break
			}
		}
		if err != nil {
			return 0, fmt.Errorf("failed to find detached process after retries: %w", err)
		}
	}

	monitor.WriteLog(fmt.Sprintf("[DETACH] Found process with PID: %d", pid))
	return pid, nil
}

// Поиск процесса по пути к исполняемому файлу
func findProcessByPath(execPath string) (int, error) {
	// Получаем базовое имя файла
	execName := filepath.Base(execPath)

	// Используем pgrep для поиска самого свежего процесса (-n = newest)
	cmd := exec.Command("pgrep", "-n", "-f", execPath)
	output, err := cmd.Output()
	if err != nil {
		// Пробуем по базовому имени
		cmd = exec.Command("pgrep", "-n", execName)
		output, err = cmd.Output()
		if err != nil {
			return 0, fmt.Errorf("process not found: %s", execName)
		}
	}

	// Парсим PID
	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return 0, fmt.Errorf("no PID found")
	}

	var pid int
	_, err = fmt.Sscanf(pidStr, "%d", &pid)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID '%s': %w", pidStr, err)
	}

	return pid, nil
}
