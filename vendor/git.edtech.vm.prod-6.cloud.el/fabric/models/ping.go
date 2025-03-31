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

	// system related
	OS   string `json:"os"`
	Arch string `json:"arch"`

	ServiceVersion string `json:"service_version"`
	HashCommit     string `json:"hash_commit"`

	Code  int    `json:"code"`
	Error string `json:"error"`

	//deprecated
	Port         int    `json:"port"`
	State        string `json:"state"`
	Https        bool   `json:"https"`
	AccessPublic bool   `json:"access_public"`
}

// PongObj — тип ответа, который сервис отдает прокси при периодическом опросе (ping’е)
type PongObj struct {
	Uid          string    `json:"uid"`                   // uid сервиса
	ReplicaID    string    `json:"replica_id"`            // id инстанса
	ProjectUid   string    `json:"project_uid,omitempty"` // uid проекта
	Project      string    `json:"project"`               // имя проекта
	Name         string    `json:"name"`                  // имя сервиса в проекте
	Service      string    `json:"service"`               // имя сервиса (файла)
	Version      string    `json:"version"`               // версия сервиса
	HashCommit   string    `json:"hash_commit"`
	Status       string    `json:"status"`
	Host         string    `json:"host"`
	Pid          int       `json:"pid"`
	Replicas     int       `json:"replicas"`
	PortHTTP     int       `json:"portHTTP,omitempty"`
	PortHTTPS    int       `json:"portHTTPS,omitempty"`
	PortGrpc     int       `json:"portGrpc,omitempty"`
	Follower     string    `json:"follower"`
	Environment  string    `json:"environment"`
	Cluster      string    `json:"cluster"`
	DeadTime     int64     `json:"dead_time"`
	StartTime    time.Time `json:"start_time"`
	Uptime       string    `json:"uptime"`
	DC           string    `json:"dc"`
	Mask         string    `json:"mask"`
	AccessPublic bool      `json:"access_public"`

	// System related
	OS   string `json:"os"`
	Arch string `json:"arch"`

	Code  int    `json:"code"`
	Error string `json:"error"`

	// Deprecated
	Port       int         `json:"port,omitempty"`
	State      string      `json:"state,omitempty"`
	Https      bool        `json:"https"`
	Metrics    interface{} `json:"metrics"`
	PortMetric int         `json:"portMetric,omitempty"`
}

type Hosts struct {
	Host     string `json:"host"`
	PortFrom int    `json:"portfrom"`
	PortTo   int    `json:"portto"`
	Protocol string `json:"protocol"`
}
