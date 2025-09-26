package models

import "fmt"

// //////////////////////////////////////
// Deployment Config
// //////////////////////////////////////
type DeploymentConfig struct {
	ID string `json:"id" yaml:"id"` // уникальный идентификатор развертывания (возможно использовать для хранения)

	DCConfigs map[string]DCDeployment `json:"dc_configs" yaml:"dc_configs"` // конфигурации дата-центров

	Mode string `json:"mode" yaml:"mode"`

	Strategy    string `json:"strategy" yaml:"strategy"`         // rolling, canary, blue-green
	MaxParallel int    `json:"max_parallel" yaml:"max_parallel"` // max parallel deployments
	Timeout     string `json:"timeout" yaml:"timeout"`           // deployment timeout

	// S3 configs
	S3Config S3Config `json:"s3_config" yaml:"s3_config"`
}

type S3Config struct {
	Bucket string `json:"bucket" yaml:"bucket"`
	Path   string `json:"path" yaml:"path"`
}

type DCDeployment struct {
	Services map[string]ServiceDeployment `json:"services" yaml:"services"` // uid -> сервис для развертывания
	Agents   map[string]AgentConfig       `json:"agents" yaml:"agents"`
	Priority int                          `json:"priority" yaml:"priority"` // deployment order between DCs
}

type AgentConfig struct {
	OS   string `json:"os" yaml:"os"`
	Arch string `json:"arch" yaml:"arch"`
	Host string `json:"host" yaml:"host"`
}

type ServiceDeployment struct {
	Project     string        `json:"project" yaml:"project"`         // название проекта
	Name        string        `json:"name" yaml:"name"`               // имя сервиса (часть Path=project/name)
	Service     string        `json:"service" yaml:"service"`         // тип сервиса
	Version     string        `json:"version" yaml:"version"`         // версия сервиса
	Environment string        `json:"environment" yaml:"environment"` // окружение
	Replicas    ReplicaConfig `json:"replicas" yaml:"replicas"`       // конфигурация реплик
	OS          string        `json:"os" yaml:"os"`
	Arch        string        `json:"arch" yaml:"arch"`
	Config      string        `json:"config" yaml:"config"`                                 // конфигурация сервиса в base64 string или путь в с3 к конфигурации ("s3://{bucket}/{path}/item.cfg")
	Canary      *CanaryConfig `json:"canary,omitempty" yaml:"canary,omitempty"`             // конфигурация канареечного развертывания
	TargetAgent string        `json:"target_agent,omitempty" yaml:"target_agent,omitempty"` // host агента
}

func (sd ServiceDeployment) Domain() string {
	return fmt.Sprintf("%s/%s", sd.Project, sd.Name)
}

type ReplicaConfig struct {
	Desired  int `json:"desired" yaml:"desired"`     // желаемое количество реплик
	MinReady int `json:"min_ready" yaml:"min_ready"` // минимальное количество готовых реплик
	MaxSurge int `json:"max_surge" yaml:"max_surge"` // максимальное превышение реплик
}

type CanaryConfig struct {
	Percentage     int            `json:"percentage" yaml:"percentage"`           // процент канареечных реплик
	StepPercentage int            `json:"step_percentage" yaml:"step_percentage"` // процент шага
	StepInterval   string         `json:"step_interval" yaml:"step_interval"`     // интервал между шагами
	Metrics        []CanaryMetric `json:"metrics" yaml:"metrics"`                 // метрики для мониторинга
}

type CanaryMetric struct {
	Name      string  `json:"name" yaml:"name"`           // название метрики
	Threshold float64 `json:"threshold" yaml:"threshold"` // пороговое значение
}

// //////////////////////////////////////
// Deployment State
// //////////////////////////////////////

type DeploymentState struct {
	ID           string          // уникальный идентификатор развертывания
	LastError    string          `json:"last_error"`    // последняя ошибка
	Status       string          `json:"status"`        // статус развертывания
	Progress     float32         `json:"progress"`      // прогресс развертывания 0.0-1.0
	AgentsStates map[string]bool `json:"agents_states"` // готовность агентов
}

// ROLL

type ServiceRoll struct {
	Project       string
	Name          string
	TargetVersion string
	OS            string
	Arch          string
}
