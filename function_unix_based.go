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

// Запуск процесса с двойным fork (полное отделение от родителя)
func startDetachedProcess(path string, args []string, logPath string, monitor *ProcessMonitor) (int, error) {
	// Проверяем, запущены ли мы как detach helper
	if len(os.Args) > 1 && os.Args[1] == "--detach-helper" {
		// Это промежуточный процесс - создаем финальный и завершаемся
		return runFinalProcess()
	}

	// Открываем лог файл для финального процесса
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to open log file: %w", err)
	}
	logFile.Close()

	// Создаем промежуточный процесс (первый fork)
	// Передаем через аргументы: --detach-helper <logPath> <path> <args...>
	intermediateArgs := []string{"--detach-helper", logPath, path}
	intermediateArgs = append(intermediateArgs, args...)

	cmd := exec.Command(os.Args[0], intermediateArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	// Промежуточный процесс не должен иметь связи с нами
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	monitor.WriteLog("[DETACH] Starting intermediate process (first fork)")

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start intermediate process: %w", err)
	}

	// Ждем завершения промежуточного процесса
	if err := cmd.Wait(); err != nil {
		return 0, fmt.Errorf("intermediate process failed: %w", err)
	}

	monitor.WriteLog("[DETACH] Intermediate process completed, searching for final process")

	// Даем время финальному процессу запуститься
	time.Sleep(500 * time.Millisecond)

	// Ищем PID финального процесса по пути к исполняемому файлу
	pid, err := findProcessByPath(path)
	if err != nil {
		return 0, fmt.Errorf("failed to find detached process: %w", err)
	}

	monitor.WriteLog(fmt.Sprintf("[DETACH] Found detached process with PID: %d", pid))
	return pid, nil
}

// Выполняется в промежуточном процессе - создает финальный и завершается
func runFinalProcess() (int, error) {
	if len(os.Args) < 4 {
		return 0, fmt.Errorf("insufficient arguments for detach helper")
	}

	logPath := os.Args[2]
	finalPath := os.Args[3]
	finalArgs := os.Args[4:]

	// Открываем лог файл
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return 0, err
	}
	defer logFile.Close()

	// Создаем финальный процесс (второй fork)
	cmd := exec.Command(finalPath, finalArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	// Перенаправляем все в лог файл
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil

	// Запускаем финальный процесс
	if err := cmd.Start(); err != nil {
		return 0, err
	}

	// Промежуточный процесс завершается
	// Финальный процесс будет "усыновлен" init/systemd (PPID = 1)
	os.Exit(0)
	return 0, nil // Не достигается
}

// Поиск процесса по пути к исполняемому файлу
func findProcessByPath(execPath string) (int, error) {
	// Получаем базовое имя файла
	execName := filepath.Base(execPath)

	// Используем pgrep для поиска
	cmd := exec.Command("pgrep", "-f", execPath)
	output, err := cmd.Output()
	if err != nil {
		// Пробуем по имени
		cmd = exec.Command("pgrep", execName)
		output, err = cmd.Output()
		if err != nil {
			return 0, fmt.Errorf("process not found: %s", execName)
		}
	}

	// Парсим первый PID
	pids := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(pids) == 0 {
		return 0, fmt.Errorf("no PIDs found")
	}

	var pid int
	_, err = fmt.Sscanf(pids[0], "%d", &pid)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID: %w", err)
	}

	return pid, nil
}
