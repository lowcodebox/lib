package models

// Pong тип ответа, который сервис отдает прокси при периодическом опросе (ping-е)
type Pong struct {
	Uid         string      `json:"uid"`
	Config      string      `json:"config"`
	Name        string      `json:"name"`
	Version     string      `json:"version"`
	Status      string      `json:"status"`
	Host        string      `json:"host"`
	Pid         string      `json:"pid"`
	Replicas    int         `json:"replicas"`
	PortHTTP    int         `json:"portHTTP"`
	PortGrpc    int         `json:"portGrpc"`
	PortMetric  int         `json:"portMetric"`
	PortHTTPS   int         `json:"portHTTPS"`
	EnableHttps bool        `json:"enableHttps"`
	DeadTime    int64       `json:"dead_time"`
	Follower    string      `json:"follower"`
	Metrics     interface{} `json:"metrics"`

	ServiceVersion string `json:"service_version"`
	HashCommit     string `json:"hash_commit"`

	//deprecated
	Port  int    `json:"port"`
	State string `json:"state"`
	Https bool   `json:"https"`
}

type Hosts struct {
	Host     string `json:"host"`
	PortFrom int    `json:"portfrom"`
	PortTo   int    `json:"portto"`
	Protocol string `json:"protocol"`
}
