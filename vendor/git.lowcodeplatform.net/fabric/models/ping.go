package models

// Pong тип ответа, который сервис отдает прокси при периодическом опросе (ping-е)
type Pong struct {
	Uid      string `json:"uid"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Status   string `json:"status"`
	Port     int    `json:"port"`
	Pid      string `json:"pid"`
	State    string `json:"state"`
	Replicas int    `json:"replicas"`
	Https    bool   `json:"https"`
	DeadTime int64  `json:"dead_time"`
	Follower string `json:"follower"`
	Grpc     int    `json:"grpc"`
	Metric   int    `json:"metric"`
}

type Hosts struct {
	Host     string `json:"host"`
	PortFrom int    `json:"portfrom"`
	PortTo   int    `json:"portto"`
	Protocol string `json:"protocol"`
}
