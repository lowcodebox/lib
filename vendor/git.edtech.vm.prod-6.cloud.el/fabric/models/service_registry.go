package models

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type AgentMapUpdate struct {
	AgentHost   string                      `json:"agent_host"`
	DC          string                      `json:"dc"`
	Environment string                      `json:"environment"`
	ServiceMap  map[string][]ServiceReplica `json:"service_map"`
	Timestamp   time.Time                   `json:"timestamp"`
}

type VersionedDCMap struct {
	DCMap            DCMap     `json:"dc_map"`             // ИСПРАВЛЕНО: используем правильный тип
	Version          uint64    `json:"version"`            // глобальная версия карты
	Timestamp        time.Time `json:"timestamp"`          // время последнего обновления
	Checksum         string    `json:"checksum"`           // контрольная сумма для проверки целостности
	LastGossipUpdate time.Time `json:"last_gossip_update"` // время последнего gossip обновления
	SourceInstance   string    `json:"source_instance"`    // ID инстанса, который последний обновлял карту
}

// Вычисляем контрольную сумму карты
func (vdm *VersionedDCMap) CalculateChecksum() string {
	// Сериализуем только DCMap без метаданных
	data, err := json.Marshal(vdm.DCMap)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// Проверяем валидность карты
func (vdm *VersionedDCMap) IsValid() bool {
	return vdm.Checksum == vdm.CalculateChecksum()
}

// Обновляем метаданные после изменения карты
func (vdm *VersionedDCMap) UpdateMetadata(instanceID string) {
	vdm.Version++
	vdm.Timestamp = time.Now()
	vdm.Checksum = vdm.CalculateChecksum()
	vdm.SourceInstance = instanceID
}

// Добавляем метод для получения статистики
func (vdm *VersionedDCMap) GetStats() DCMapStats {
	stats := DCMapStats{
		Version:      vdm.Version,
		LastUpdate:   vdm.Timestamp,
		ServicesByDC: make(map[string]int),
	}

	totalAgents := make(map[string]bool)

	for dc, dcServices := range vdm.DCMap {
		serviceCount := 0
		for _, agent := range dcServices.Agents {
			totalAgents[agent.Host] = true
			serviceCount += len(agent.Dependencies)
			stats.TotalReplicas += len(agent.Dependencies)
		}
		stats.ServicesByDC[dc] = serviceCount
		stats.TotalServices += serviceCount
	}

	stats.AgentCount = len(totalAgents)
	return stats
}

// Добавляем методы для работы с картой (делегируем к внутренней DCMap)
func (vdm *VersionedDCMap) GetServiceReplicas(path string) []ServiceReplica {
	var replicas []ServiceReplica

	for _, dcServices := range vdm.DCMap {
		for _, agent := range dcServices.Agents {
			for _, replica := range agent.Dependencies {
				if replica.Path == path {
					replicas = append(replicas, *replica)
				}
			}
		}
	}

	return replicas
}

// Проверяем доступность сервиса
func (vdm *VersionedDCMap) IsServiceAvailable(path string) bool {
	replicas := vdm.GetServiceReplicas(path)
	return len(replicas) > 0
}

// Получаем endpoints для сервиса
func (vdm *VersionedDCMap) GetEndpoints(path string, opts EndpointOpts) []string {
	replicas := vdm.GetServiceReplicas(path)
	if len(replicas) == 0 {
		return []string{}
	}

	// Применяем фильтрацию и сортировку (упрощенная версия)
	endpoints := make([]string, 0, len(replicas))
	for _, replica := range replicas {
		if replica.Healthy { // Только здоровые реплики
			endpoints = append(endpoints, replica.GetUrl())
		}
	}

	// Применяем лимит если указан
	if opts.Limit > 0 && opts.Limit < len(endpoints) {
		endpoints = endpoints[:opts.Limit]
	}

	return endpoints
}

// Создаем новую версионную карту
func NewVersionedDCMap(instanceID string) *VersionedDCMap {
	vdm := &VersionedDCMap{
		DCMap:          make(DCMap),
		Version:        1,
		Timestamp:      time.Now(),
		SourceInstance: instanceID,
	}
	vdm.Checksum = vdm.CalculateChecksum()
	return vdm
}

type DCMapStats struct {
	Version       uint64         `json:"version"`
	LastUpdate    time.Time      `json:"last_update"`
	TotalServices int            `json:"total_services"`
	TotalReplicas int            `json:"total_replicas"`
	AgentCount    int            `json:"agent_count"`
	ServicesByDC  map[string]int `json:"services_by_dc"`
}
