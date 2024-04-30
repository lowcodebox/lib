package models

import "time"

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
	Follower    string      `json:"follower"`
	Metrics     interface{} `json:"metrics"`
	Environment string      `json:"environment"`
	Cluster     string      `json:"cluster"`
	DeadTime    int64       `json:"deadtime"`
	Runtime     time.Time   `json:"runtime"`
	Uptime      string      `json:"uptime"`
	DC          string      `json:"dc"`
	Mask        string      `json:"mask"`

	ServiceVersion string `json:"service_version"`
	HashCommit     string `json:"hash_commit"`

	//deprecated
	Port         int    `json:"port"`
	State        string `json:"state"`
	Https        bool   `json:"https"`
	AccessPublic bool   `json:"access_public"`
}

type Hosts struct {
	Host     string `json:"host"`
	PortFrom int    `json:"portfrom"`
	PortTo   int    `json:"portto"`
	Protocol string `json:"protocol"`
}
