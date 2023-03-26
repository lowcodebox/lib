package models

import "time"

// Pong тип ответа, который сервис отдает прокси при периодическом опросе (ping-е)
type Pong struct {
	Uid      string         `json:"uid"`
	Config   string         `json:"config"`
	Name     string         `json:"name"`
	Version  string         `json:"version"`
	Path     string         `json:"path"`
	Status   string         `json:"status"`
	Port     int            `json:"port"`
	Pid      string         `json:"pid"`
	State    string         `json:"state"`
	Replicas int            `json:"replicas"`
	Https    bool           `json:"https"`
	DeadTime int64          `json:"dead_time"`
	Follower string         `json:"follower"`
	Grpc     int            `json:"grpc"`
	Metric   int            `json:"metric"`
	Host     string         `json:"host"`
	Metrics  []MetricsField `json:"metrics"`
}

type MetricsField struct {
	Help         string        `json:"help"`
	Type         string        `json:"type"`
	Count        float64       `json:"count"`
	Value        string        `json:"value"`
	Viewer       string        `json:"viewer"`
	SaveInterval time.Duration `json:"saveInterval"`
	SavePeriod   time.Duration `json:"savePeriod"`
}

type Hosts struct {
	Host     string `json:"host"`
	PortFrom int    `json:"portfrom"`
	PortTo   int    `json:"portto"`
	Protocol string `json:"protocol"`
}
