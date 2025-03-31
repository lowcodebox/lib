package models

type ServiceMetrics struct {
	ErrorRate   float64
	Latency     float64
	CPUUsage    float64
	MemoryUsage float64
	Custom      map[string]interface{}
}
