package models

import (
	"sync"
	"time"
)

type Service struct {
	Uid          string `json:"uid"`
	Pid          int64  `json:"pid"`
	Agent        string `json:"agent"`
	Project      string `json:"project"`
	Service      string `json:"service"`
	Path         string `json:"path"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Status       string `json:"status"`
	Replicas     int    `json:"replicas"`
	PortHTTP     int    `json:"portHTTP"`
	PortGrpc     int    `json:"portGrpc"`
	PortHTTPS    int    `json:"portHTTPS"`
	EnableHTTPS  bool   `json:"enable_https,enableHttps"`
	Follower     string `json:"follower"`
	Enviroment   string `json:"environment"`
	Cluster      string `json:"cluster"`
	AccessPublic bool   `json:"access_public"`
	DC           string `json:"dc"`
	Mask         string `json:"mask"`
	StartedAt    int64  `json:"started_at"`
	Error        string `json:"error"`

	LastPing time.Time `json:"last_ping"`

	Instances []Pong `json:"services"`
}

type ServiceReplica struct {
	ServiceUid   string         `json:"service_uid"`
	Pid          int64          `json:"pid"`
	AgentHost    string         `json:"agent_host"`
	Project      string         `json:"project"`
	Service      string         `json:"service"`
	Path         string         `json:"path"`
	Name         string         `json:"name"`
	Version      string         `json:"version"`
	Status       string         `json:"status"`
	PortHTTP     int            `json:"portHTTP"`
	PortGrpc     int            `json:"portGrpc"`
	PortHTTPS    int            `json:"portHTTPS"`
	EnableHTTPS  bool           `json:"enable_https,enableHttps"`
	Enviroment   string         `json:"environment"`
	AccessPublic bool           `json:"access_public"`
	DC           string         `json:"dc"`
	Healthy      bool           `json:"healthy"`
	Uptime       string         `json:"uptime"`
	Mask         string         `json:"mask"`
	StartedAt    int64          `json:"started_at"`
	Error        string         `json:"error"`
	Metrics      ServiceMetrics `json:"metrics"`
}

type SyncServiceMap struct {
	m  map[string][]Pong
	mx *sync.RWMutex
}

func NewSyncServiceMap() *SyncServiceMap {
	return &SyncServiceMap{
		m:  map[string][]Pong{},
		mx: &sync.RWMutex{},
	}
}

// Set обновить карту сервисов
func (s *SyncServiceMap) Set(m map[string]Service) {
	mapDomainServices := make(map[string][]Pong, len(m))

	// Преобразование из uid - service в domain - service
	for _, v := range m {
		instances := deleteEmpty(v.Instances)

		// Пропуск пустышек
		if len(instances) == 0 {
			continue
		}

		path := v.Path
		mapDomainServices[path] = instances
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	s.m = mapDomainServices
}

// Get получить текущие сервисы по домену
func (s *SyncServiceMap) Get(domain string) []Pong {
	s.mx.RLock()
	defer s.mx.RUnlock()
	val, ok := s.m[domain]
	if !ok {
		return nil
	}

	newSlice := make([]Pong, len(val))
	copy(newSlice, val)
	return newSlice
}

// deleteEmpty подтирает пустые инстансы, где нет порта или хоста
func deleteEmpty(p []Pong) []Pong {
	lastidx := len(p) - 1
	for i := 0; i <= lastidx; i++ {
		if p[i].Host != "" && p[i].PortHTTP > 0 {
			continue
		}
		p[i], p[lastidx] = p[lastidx], p[i]
		lastidx--
		i--
	}
	return p[0 : lastidx+1]
}
