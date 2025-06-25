package lib

import (
	"fmt"
	"os"
	"strings"
)

// ProcessConfig содержит конфигурацию для запуска процесса
type ProcessConfig struct {
	Path    string // Путь к исполняемому файлу
	Project string // Название проекта
	Service string // Название сервиса
	Config  string // Путь к конфигурационному файлу
	Command string // Команда для выполнения
	Mode    string // Режим работы (debug, prod, etc.)
	DC      string // Дата-центр
	Port    string // Порт
}

// Validate проверяет корректность конфигурации процесса
func (pc *ProcessConfig) Validate() error {
	if pc.Path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if pc.Project == "" {
		return fmt.Errorf("project cannot be empty")
	}

	if pc.Service == "" {
		return fmt.Errorf("service cannot be empty")
	}

	if pc.Config == "" {
		return fmt.Errorf("config cannot be empty")
	}

	if pc.Command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	if pc.Mode == "" {
		return fmt.Errorf("mode cannot be empty")
	}

	if pc.DC == "" {
		return fmt.Errorf("dc cannot be empty")
	}

	if pc.Port == "" {
		return fmt.Errorf("port cannot be empty")
	}

	return nil
}

// BuildArgs создает массив аргументов для запуска процесса
func (pc *ProcessConfig) BuildArgs() []string {
	args := []string{
		pc.Command,
		"-c", pc.Config,
		"-m", pc.Mode,
		"-d", pc.DC,
		"-p", pc.Port,
	}

	return args
}

// String возвращает строковое представление конфигурации для логирования
func (pc *ProcessConfig) String() string {
	return fmt.Sprintf("ProcessConfig{Path: %s, Project: %s, Service: %s, Config: %s, Command: %s, Mode: %s, DC: %s, Port: %s}",
		pc.Path, pc.Project, pc.Service, pc.Config, pc.Command, pc.Mode, pc.DC, pc.Port)
}

// GetLogFileName возвращает имя файла лога для данной конфигурации
func (pc *ProcessConfig) GetLogFileName() string {
	return fmt.Sprintf("%s-%s.log", pc.Project, pc.Service)
}

// GetAuditInfo возвращает информацию для аудит лога
func (pc *ProcessConfig) GetAuditInfo() map[string]string {
	return map[string]string{
		"path":    pc.Path,
		"project": pc.Project,
		"service": pc.Service,
		"config":  pc.Config,
		"command": pc.Command,
		"mode":    pc.Mode,
		"dc":      pc.DC,
		"port":    pc.Port,
	}
}

// Clone создает копию конфигурации
func (pc *ProcessConfig) Clone() *ProcessConfig {
	return &ProcessConfig{
		Path:    pc.Path,
		Project: pc.Project,
		Service: pc.Service,
		Config:  pc.Config,
		Command: pc.Command,
		Mode:    pc.Mode,
		DC:      pc.DC,
		Port:    pc.Port,
	}
}

// ValidateAndSanitize проверяет и очищает конфигурацию
func (pc *ProcessConfig) ValidateAndSanitize() error {
	// Базовая валидация
	if err := pc.Validate(); err != nil {
		return err
	}

	// Очистка путей от лишних пробелов
	pc.Path = strings.TrimSpace(pc.Path)
	pc.Config = strings.TrimSpace(pc.Config)

	// Очистка строковых полей
	pc.Project = strings.TrimSpace(pc.Project)
	pc.Service = strings.TrimSpace(pc.Service)
	pc.Command = strings.TrimSpace(pc.Command)
	pc.Mode = strings.TrimSpace(pc.Mode)
	pc.DC = strings.TrimSpace(pc.DC)
	pc.Port = strings.TrimSpace(pc.Port)

	// Проверка что пути существуют
	if _, err := os.Stat(pc.Path); err != nil {
		return fmt.Errorf("executable path does not exist: %s, error: %w", pc.Path, err)
	}

	if _, err := os.Stat(pc.Config); err != nil {
		return fmt.Errorf("config file does not exist: %s, error: %w", pc.Config, err)
	}

	return nil
}

// GetCommandLine возвращает полную командную строку для выполнения
func (pc *ProcessConfig) GetCommandLine() string {
	args := pc.BuildArgs()
	return fmt.Sprintf("%s %s", pc.Path, strings.Join(args, " "))
}
