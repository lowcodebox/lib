package models

import (
	"sync"
)

type Service struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Version      string `json:"version"`
	Replicas     int    `json:"replicas"`
	Follower     string `json:"follower"`
	EnableHTTPS  bool   `json:"enable_https"`
	Status       string `json:"status"`
	Enviroment   string `json:"environment"`
	Cluster      string `json:"cluster"`
	AccessPublic bool   `json:"access_public"`
	DC           string `json:"dc"`
	Mask         string `json:"mask"`

	Instances []Pong `json:"services"`
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
		// Пропуск пустышек
		if len(v.Instances) == 0 {
			continue
		}

		path := v.Path
		mapDomainServices[path] = v.Instances
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
