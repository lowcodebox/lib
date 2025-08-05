package models

import "time"

type ServiceMetrics struct {
	// Основные счетчики
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`   // 5xx
	WarningCount int64   `json:"warning_count"` // 4xx
	SuccessRate  float64 `json:"success_rate"`

	// Статистика времени отклика
	MinResponseTime time.Duration `json:"min_response_time"`
	MaxResponseTime time.Duration `json:"max_response_time"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	P95ResponseTime time.Duration `json:"p95_response_time"`

	// Ресурсы
	AvgCPUUsage    float64 `json:"avg_cpu_usage"`
	AvgMemoryUsage uint64  `json:"avg_memory_usage"`
	AvgGoroutines  int     `json:"avg_goroutines"`

	Custom map[string]interface{} `json:"custom"`
}
