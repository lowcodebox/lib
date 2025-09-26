package models

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type DCMap map[string]*DCServices

type DCServices struct {
	Agents          []*Agent `json:"agents"`           // список агентов
	LocalGeneration uint64   `json:"local_generation"` // локальное поколение
}

type Agent struct {
	Host         string                     `json:"host"`         // хост агента
	Replicas     []ServiceReplica           `json:"replicas"`     // реплики агента (на будущее)
	Dependencies map[string]*ServiceReplica `json:"dependencies"` // зависимости сервиса, ключ ReplicaID
	Healthy      bool                       `json:"healthy"`      // состояние здоровья
}

type ServiceReplica struct {
	Uid          string         `json:"service_uid"`
	ReplicaID    string         `json:"replica_id"`
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
	LastPinged   time.Time      `json:"last_pinged"`
	PortHTTPS    int            `json:"portHTTPS"`
	EnableHTTPS  bool           `json:"enable_https"`
	Enviroment   string         `json:"environment"`
	AccessPublic bool           `json:"access_public"`
	DC           string         `json:"dc"`
	Healthy      bool           `json:"healthy"`
	Uptime       string         `json:"uptime"`
	Mask         string         `json:"mask"`
	StartedAt    int64          `json:"started_at"`
	Error        string         `json:"error"`
	Metrics      ServiceMetrics `json:"metrics"`

	OS   string `json:"os"`
	Arch string `json:"arch"`

	Strategy RollingStrategy `json:"straregy"`
}

func (sr ServiceReplica) Domain() string {
	return fmt.Sprintf("%s/%s", sr.Project, sr.Name)
}

func (r *ServiceReplica) GetUrl() string {
	return fmt.Sprintf("%s:%d", r.AgentHost, r.PortHTTP)
}

var serviceMapInstance *serviceMap
var once sync.Once

////////////////////////////////////////
// Service Map
////////////////////////////////////////

type ServiceMap interface {
	// добавить с заменой
	Upsert(replica *ServiceReplica, agentHost string) error
	// убрать сервис
	Remove(service *ServiceReplica, agentHost string) error
	// строит общую карту на основе списка реплик
	Rebuild(repls []ServiceReplica)
	// возвращает хосты учитывая фильтрацию
	Endpoints(path string, filter EndpointOpts) []string
	// чистит текущую карту от сервисов более не присутствующих в свежей карте
	// agentMap - карта всех агентов с их сервис-репликами
	Clean(host string, agentMap map[string][]ServiceReplica) error
	// сброс карты
	Reset()
	// прогресс деплоя 0.0..1.0
	Progress(target *DeploymentConfig) float32
	// получить полную карту по всем ДЦ
	Set(dcMap DCMap, gen uint64, lastUpdate time.Time)
	// получить полную карту по всем ДЦ
	Get() DCMap
	// проверяет доступность сервиса по всем ДЦ
	IsServiceAvailable(path string) bool
}

// Update the serviceMap struct to include the indexes
type serviceMap struct {
	mut          sync.RWMutex
	DCMap        DCMap                       `json:"dc_map"`      // карта сервисов по дата-центрам
	Generation   uint64                      `json:"generation"`  // поколение карты сервисов
	LastUpdate   time.Time                   `json:"last_update"` // время последнего обновления
	pathIndex    map[string][]ServiceReplica // Index of replicas by path
	replicaIndex map[string]*ServiceReplica  // Index of replicas by ReplicaID
}

// Update the NewServiceMap function to initialize the indexes
func NewServiceMap() ServiceMap {
	once.Do(func() {
		serviceMapInstance = &serviceMap{
			DCMap:        make(DCMap),
			pathIndex:    make(map[string][]ServiceReplica),
			replicaIndex: make(map[string]*ServiceReplica),
		}
	})
	return serviceMapInstance
}

// строит общую карту на основе списка реплик
func (s *serviceMap) Rebuild(repls []ServiceReplica) {
	s.mut.Lock()
	defer s.mut.Unlock()

	// Reset all maps
	s.DCMap = make(DCMap)
	s.pathIndex = make(map[string][]ServiceReplica)
	s.replicaIndex = make(map[string]*ServiceReplica)
	s.LastUpdate = time.Now()

	// Rebuild maps from replicas list
	for _, replica := range repls {
		// Add to replica index
		replicaCopy := replica
		s.replicaIndex[replica.ReplicaID] = &replicaCopy

		// Add to path index
		s.pathIndex[replica.Path] = append(s.pathIndex[replica.Path], replica)

		// Add to DC map
		dcServices, ok := s.DCMap[replica.DC]
		if !ok {
			dcServices = &DCServices{
				Agents:          make([]*Agent, 0),
				LocalGeneration: 0,
			}
			s.DCMap[replica.DC] = dcServices
		}

		// Find or create agent
		var agent *Agent
		for _, a := range dcServices.Agents {
			if a.Host == replica.AgentHost {
				agent = a
				break
			}
		}

		if agent == nil {
			agent = &Agent{
				Host:         replica.AgentHost,
				Dependencies: make(map[string]*ServiceReplica),
				Replicas:     make([]ServiceReplica, 0),
				Healthy:      true,
			}
			dcServices.Agents = append(dcServices.Agents, agent)
		}

		// Add replica to agent - ИСПРАВЛЕНИЕ
		agent.Dependencies[replica.ReplicaID] = &replicaCopy
		agent.Replicas = append(agent.Replicas, replica) // Добавляем в Replicas
	}
}

func (s *serviceMap) Reset() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.LastUpdate = time.Now()
	s.DCMap = make(DCMap)
	s.pathIndex = make(map[string][]ServiceReplica)
	s.replicaIndex = make(map[string]*ServiceReplica)
}

// Fix the Upsert method to properly update the indexes
func (s *serviceMap) Upsert(replica *ServiceReplica, agentHost string) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	// Initialize DCMap if needed
	if s.DCMap == nil {
		s.DCMap = make(DCMap)
	}

	// Continue with the existing DC and agent management logic
	dcServices, ok := s.DCMap[replica.DC]
	if !ok {
		slog.Info("создаем новую карту для датацентра", "dc", replica.DC)
		dcServices = new(DCServices)
		dcServices.Agents = make([]*Agent, 0)
		dcServices.Agents = append(dcServices.Agents, &Agent{
			Host:         agentHost,
			Dependencies: map[string]*ServiceReplica{},
			Replicas:     []ServiceReplica{},
		})
	}

	agentFound := false
	for i, agent := range dcServices.Agents {
		if agent.Host != agentHost {
			continue
		}
		agentFound = true
		agent.Dependencies[replica.ReplicaID] = replica
		dcServices.Agents[i] = agent
	}

	if !agentFound {
		newAgent := &Agent{
			Host:         agentHost,
			Dependencies: map[string]*ServiceReplica{},
			Replicas:     []ServiceReplica{*replica},
		}
		newAgent.Dependencies[replica.ReplicaID] = replica
		dcServices.Agents = append(dcServices.Agents, newAgent)
	}

	s.DCMap[replica.DC] = dcServices
	s.LastUpdate = time.Now()

	s.rebuildPathIndex()

	return nil
}

// Set заменяет текущую DCMap и перестраивает индексы
func (s *serviceMap) Set(dcMap DCMap, gen uint64, lastUpdate time.Time) {
	s.mut.Lock()
	defer s.mut.Unlock()

	// Заменяем DCMap
	s.DCMap = dcMap
	s.LastUpdate = lastUpdate
	s.Generation = gen

	// Очищаем существующие индексы
	s.pathIndex = make(map[string][]ServiceReplica)
	s.replicaIndex = make(map[string]*ServiceReplica)

	// Перестраиваем индексы из новой DCMap
	for _, dcServices := range s.DCMap {
		for _, agent := range dcServices.Agents {
			// Добавляем все реплики из Dependencies в индексы
			for replicaID, replica := range agent.Dependencies {
				// Добавляем в индекс реплик
				s.replicaIndex[replicaID] = replica

				// Добавляем в индекс по путям
				s.pathIndex[replica.Path] = append(s.pathIndex[replica.Path], *replica)
			}

			// Также добавляем реплики из Replicas slice (если они не дублируются)
			for _, replica := range agent.Replicas {
				// Проверяем, что реплика еще не добавлена через Dependencies
				if _, exists := s.replicaIndex[replica.ReplicaID]; !exists {
					// Добавляем в индекс реплик
					replicaCopy := replica
					s.replicaIndex[replica.ReplicaID] = &replicaCopy

					// Добавляем в индекс по путям
					s.pathIndex[replica.Path] = append(s.pathIndex[replica.Path], replica)
				}
			}
		}
	}
}

func (s *serviceMap) rebuildPathIndex() {
	// Clear existing indexes
	s.pathIndex = make(map[string][]ServiceReplica)
	s.replicaIndex = make(map[string]*ServiceReplica)

	// Iterate through all DCs and their agents
	for _, dcServices := range s.DCMap {
		for _, agent := range dcServices.Agents {
			// Add all replicas to both indexes
			for replicaID, replica := range agent.Dependencies {
				// Add to replica index
				s.replicaIndex[replicaID] = replica

				// Add to path index
				s.pathIndex[replica.Path] = append(s.pathIndex[replica.Path], *replica)
			}
		}
	}
}

// Fix the Remove method to update the indexes
func (s *serviceMap) Remove(replica *ServiceReplica, agentHost string) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	// Remove from replica index
	delete(s.replicaIndex, replica.ReplicaID)

	// Remove from path index
	if replicas, exists := s.pathIndex[replica.Path]; exists {
		updatedReplicas := make([]ServiceReplica, 0, len(replicas))
		for _, r := range replicas {
			if r.ReplicaID != replica.ReplicaID {
				updatedReplicas = append(updatedReplicas, r)
			}
		}
		if len(updatedReplicas) > 0 {
			s.pathIndex[replica.Path] = updatedReplicas
		} else {
			delete(s.pathIndex, replica.Path)
		}
	}

	// Continue with existing removal logic
	dcServices, ok := s.DCMap[replica.DC]
	if !ok {
		return fmt.Errorf("datacenter %s not found", replica.DC)
	}

	key := replica.ReplicaID
	agentFound := false

	for i, agent := range dcServices.Agents {
		if agent.Host != agentHost {
			continue
		}
		agentFound = true

		// Remove from Dependencies
		delete(agent.Dependencies, key)

		// Remove from Replicas
		updatedReplicas := make([]ServiceReplica, 0, len(agent.Replicas))
		for _, r := range agent.Replicas {
			if r.ReplicaID != replica.ReplicaID {
				updatedReplicas = append(updatedReplicas, r)
			}
		}
		agent.Replicas = updatedReplicas

		// If agent has no more dependencies, remove it
		if len(agent.Dependencies) == 0 {
			dcServices.Agents = append(dcServices.Agents[:i], dcServices.Agents[i+1:]...)
		}

		// If no more agents in DC, remove DC
		if len(dcServices.Agents) == 0 {
			delete(s.DCMap, replica.DC)
		}

		return nil
	}

	if !agentFound {
		return fmt.Errorf("agent %s not found in datacenter %s", agentHost, replica.DC)
	}

	return nil
}

func (s *serviceMap) Clean(host string, agentMap map[string][]ServiceReplica) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	// Create a map of current services from the agent for O(1) lookups
	currentServices := make(map[string]bool)
	for _, services := range agentMap {
		for _, service := range services {
			currentServices[service.ReplicaID] = true
		}
	}

	if len(currentServices) == 0 {
		return nil
	}

	// Track if any changes were made
	changesMade := false

	// Clean up services that no longer exist on this agent
	for dcName, dcServices := range s.DCMap {
		agentIndex := -1
		var agent *Agent

		// Find the agent
		for i, a := range dcServices.Agents {
			if a.Host == host {
				agentIndex = i
				agent = a
				break
			}
		}

		// Skip if agent not found in this DC
		if agentIndex == -1 || agent == nil {
			continue
		}

		// Track replicas to remove
		replicasToRemove := make([]string, 0)

		// Find dependencies that are not in the current services
		for replicaID, replica := range agent.Dependencies {
			if !currentServices[replicaID] {
				replicasToRemove = append(replicasToRemove, replicaID)
				changesMade = true

				// Remove from indexes
				delete(s.replicaIndex, replicaID)

				// Update path index
				if replicas, exists := s.pathIndex[replica.Path]; exists {
					if len(replicas) == 1 {
						// If this is the only replica for this path, delete the path entry
						delete(s.pathIndex, replica.Path)
					} else {
						// Otherwise, filter out this replica
						updatedReplicas := make([]ServiceReplica, 0, len(replicas)-1)
						for _, r := range replicas {
							if r.ReplicaID != replicaID {
								updatedReplicas = append(updatedReplicas, r)
							}
						}
						s.pathIndex[replica.Path] = updatedReplicas
					}
				}
			}
		}

		// If no replicas to remove, continue to next DC
		if len(replicasToRemove) == 0 {
			continue
		}

		// Remove the dependencies
		for _, replicaID := range replicasToRemove {
			delete(agent.Dependencies, replicaID)
		}

		// Update Replicas slice
		if len(agent.Replicas) > 0 {
			updatedReplicas := make([]ServiceReplica, 0, len(agent.Replicas))
			for _, replica := range agent.Replicas {
				if currentServices[replica.ReplicaID] {
					updatedReplicas = append(updatedReplicas, replica)
				}
			}
			agent.Replicas = updatedReplicas
		}

		// If agent has no more dependencies, remove it
		if len(agent.Dependencies) == 0 {
			dcServices.Agents = append(dcServices.Agents[:agentIndex], dcServices.Agents[agentIndex+1:]...)
			changesMade = true
		}

		// If no more agents in DC, remove DC
		if len(dcServices.Agents) == 0 {
			delete(s.DCMap, dcName)
			changesMade = true
		}
	}

	// Update generation and timestamp only if changes were made
	if changesMade {
		s.Generation++
		s.LastUpdate = time.Now()
	}

	return nil
}

func (s *serviceMap) Get() DCMap {
	s.mut.Lock()
	defer s.mut.Unlock()
	return s.DCMap
}

func (s *serviceMap) Progress(target *DeploymentConfig) float32 {
	totalDesiredReplicas := 0
	runningReplicas := 0

	for _, dcConfig := range target.DCConfigs {
		for _, svc := range dcConfig.Services {
			totalDesiredReplicas += svc.Replicas.Desired
		}
	}

	s.mut.Lock()
	for _, dc := range s.DCMap {
		for _, agent := range dc.Agents {
			for range agent.Dependencies {
				runningReplicas++
			}
		}
	}
	s.mut.Unlock()

	if totalDesiredReplicas == 0 {
		return 1.0
	}

	return float32(runningReplicas) / float32(totalDesiredReplicas)
}

// Endpoints returns a list of endpoints for a given path with optional filtering
func (s *serviceMap) Endpoints(path string, opts EndpointOpts) []string {
	s.mut.RLock()
	defer s.mut.RUnlock()

	// If no path is specified, return empty list
	if path == "" {
		return []string{}
	}

	// Get replicas from the path index
	replicas, exists := s.pathIndex[path]
	if !exists || len(replicas) == 0 {
		return []string{}
	}

	// Make a copy of the replicas to avoid modifying the original slice
	replicasCopy := make([]ServiceReplica, len(replicas))
	copy(replicasCopy, replicas)

	// Sort replicas based on the strategy
	switch opts.SortBy {
	case SortRandom:
		// Shuffle the replicas
		rand.Shuffle(len(replicasCopy), func(i, j int) {
			replicasCopy[i], replicasCopy[j] = replicasCopy[j], replicasCopy[i]
		})
	case SortMinLatency:
		// Sort by minimum response time
		sort.Slice(replicasCopy, func(i, j int) bool {
			return replicasCopy[i].Metrics.MinResponseTime < replicasCopy[j].Metrics.MinResponseTime
		})
	case SortNewest:
		// Sort by newest first (highest StartedAt timestamp)
		sort.Slice(replicasCopy, func(i, j int) bool {
			return replicasCopy[i].StartedAt > replicasCopy[j].StartedAt
		})
	}

	// Apply limit if specified
	if opts.Limit > 0 && opts.Limit < len(replicasCopy) {
		replicasCopy = replicasCopy[:opts.Limit]
	}

	// Convert replicas to endpoint URLs
	endpoints := make([]string, len(replicasCopy))
	for i, replica := range replicasCopy {
		endpoints[i] = replica.GetUrl()
	}

	return endpoints
}

func (s *serviceMap) IsServiceAvailable(path string) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()

	// If no path is specified, service is not available
	if path == "" {
		return false
	}

	// Get replicas from the path index
	replicas, exists := s.pathIndex[path]
	if !exists {
		return false
	}

	// Check if there are any active replicas
	return len(replicas) > 0
}

// ENDPOINTS OPTS
type SortStrategy string

const (
	// SortRandom returns endpoints in random order
	SortRandom SortStrategy = "random"
	// SortMinLatency returns endpoints sorted by minimum response time
	SortMinLatency SortStrategy = "min_latency"
	// SortNewest returns endpoints sorted by newest first (based on StartedAt)
	SortNewest SortStrategy = "newest"
)

// EndpointOpts defines options for filtering and sorting endpoints
type EndpointOpts struct {
	SortBy SortStrategy
	Limit  int
}

// EndpointOptsBuilder is a builder for EndpointOpts
type EndpointOptsBuilder struct {
	opts EndpointOpts
}

// NewEndpointOptsBuilder creates a new builder for EndpointOpts
func NewEndpointOptsBuilder() *EndpointOptsBuilder {
	return &EndpointOptsBuilder{
		opts: EndpointOpts{
			SortBy: SortRandom, // Default to random sorting
			Limit:  0,          // 0 means no limit
		},
	}
}

// WithSortRandom sets the sort strategy to random
func (b *EndpointOptsBuilder) WithSortRandom() *EndpointOptsBuilder {
	b.opts.SortBy = SortRandom
	return b
}

// WithSortMinLatency sets the sort strategy to minimum latency
func (b *EndpointOptsBuilder) WithSortMinLatency() *EndpointOptsBuilder {
	b.opts.SortBy = SortMinLatency
	return b
}

// WithSortNewest sets the sort strategy to newest first
func (b *EndpointOptsBuilder) WithSortNewest() *EndpointOptsBuilder {
	b.opts.SortBy = SortNewest
	return b
}

// WithLimit sets the maximum number of endpoints to return
func (b *EndpointOptsBuilder) WithLimit(limit int) *EndpointOptsBuilder {
	b.opts.Limit = limit
	return b
}

// Build returns the built EndpointOpts
func (b *EndpointOptsBuilder) Build() EndpointOpts {
	return b.opts
}
