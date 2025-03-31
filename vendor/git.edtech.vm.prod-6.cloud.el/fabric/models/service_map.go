package models

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

////////////////////////////////////////
// Service Map
////////////////////////////////////////

// уникальный ключ по которому храним реплику в карте
func makeReplicaKey(uid string, pid int64) string {
	return fmt.Sprintf("%s:%d", uid, pid)
}

type ServiceMap interface {
	Progress(target *DeploymentConfig) float32
	// добавить с заменой
	Upsert(service *Service, agentHost string) error
	// убрать сервис
	Remove(service *Service, agentHost string) error
	// обновляет карту по данным от агента
	Update(host string, agentMap map[string][]Service) error
	// сброс карты
	Clean()
	Get() DCMap
}
type DCMap map[string]*DCServices

type serviceMap struct {
	mut        sync.RWMutex
	DCMap      DCMap     `json:"dc_map"`      // карта сервисов по дата-центрам
	Generation uint64    `json:"generation"`  // поколение карты сервисов
	LastUpdate time.Time `json:"last_update"` // время последнего обновления
}

type DCServices struct {
	Agents          []*Agent `json:"agents"`           // список агентов
	LocalGeneration uint64   `json:"local_generation"` // локальное поколение
}

type Agent struct {
	Host         string                     `json:"host"`         // хост агента
	Replicas     []ServiceReplica           `json:"replicas"`     // список реплик
	Dependencies map[string]*ServiceReplica `json:"dependencies"` // зависимости сервиса, ключ "uid:pid"
	LastBleep    time.Time                  `json:"last_bleep"`   // время последнего сигнала
	Healthy      bool                       `json:"healthy"`      // состояние здоровья
}

var serviceMapInstance *serviceMap
var once sync.Once

func NewServiceMap() ServiceMap {
	once.Do(func() {
		serviceMapInstance = &serviceMap{}
	})
	return serviceMapInstance
}

func (s *serviceMap) Get() DCMap {
	s.mut.Lock()
	defer s.mut.Unlock()
	return s.DCMap
}

func (s *serviceMap) Clean() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.LastUpdate = time.Now()
	s.DCMap = DCMap{}
}

func (s *serviceMap) Remove(service *Service, agentHost string) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	dcServices, ok := s.DCMap[service.DC]
	if !ok {
		return fmt.Errorf("datacenter %s not found", service.DC)
	}

	key := makeReplicaKey(service.Uid, service.Pid)

	for i, agent := range dcServices.Agents {
		if agent.Host != agentHost {
			continue
		}

		delete(agent.Dependencies, key)

		// If agent has no more dependencies, remove it
		if len(agent.Dependencies) == 0 {
			dcServices.Agents = append(dcServices.Agents[:i], dcServices.Agents[i+1:]...)
		}

		// If no more agents in DC, remove DC
		if len(dcServices.Agents) == 0 {
			delete(s.DCMap, service.DC)
		}

		return nil
	}

	return fmt.Errorf("agent %s not found in datacenter %s", agentHost, service.DC)
}

func (s *serviceMap) Update(host string, agentMap map[string][]Service) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	// Track which DCs we've updated
	updatedDCs := make(map[string]bool)

	// Process each service in the agent map
	for _, services := range agentMap {
		for _, service := range services {
			// Track this DC as updated
			updatedDCs[service.DC] = true

			// Get or create DC services
			dcServices, ok := s.DCMap[service.DC]
			if !ok {
				if s.DCMap == nil {
					s.DCMap = make(DCMap)
				}
				dcServices = &DCServices{
					Agents: make([]*Agent, 0),
				}
				s.DCMap[service.DC] = dcServices
			}

			// Find or create agent
			var agent *Agent
			agentIndex := -1
			for i, a := range dcServices.Agents {
				if a.Host == host {
					agent = a
					agentIndex = i
					break
				}
			}

			if agent == nil {
				agent = &Agent{
					Host:         host,
					Dependencies: make(map[string]*ServiceReplica),
					LastBleep:    time.Now(),
					Healthy:      true,
				}
				dcServices.Agents = append(dcServices.Agents, agent)
			} else {
				// Update agent health status
				agent.LastBleep = time.Now()
				agent.Healthy = true
			}

			// Create or update service replica
			key := makeReplicaKey(service.Uid, service.Pid)
			uptime := time.Since(time.Unix(service.StartedAt, 0)).String()

			replica := &ServiceReplica{
				ServiceUid:  service.Uid,
				Pid:         service.Pid,
				Name:        service.Name,
				AgentHost:   host,
				Project:     service.Project,
				Service:     service.Name,
				Version:     service.Version,
				Uptime:      uptime,
				Healthy:     true,
				Status:      service.Status,
				PortHTTP:    service.PortHTTP,
				PortGrpc:    service.PortGrpc,
				PortHTTPS:   service.PortHTTPS,
				EnableHTTPS: service.EnableHTTPS,
				Enviroment:  service.Enviroment,
				DC:          service.DC,
				Mask:        service.Mask,
				StartedAt:   service.StartedAt,
			}

			agent.Dependencies[key] = replica

			if agentIndex >= 0 {
				dcServices.Agents[agentIndex] = agent
			}
		}
	}

	// Clean up services that no longer exist on this agent
	for dcName, dcServices := range s.DCMap {
		for i, agent := range dcServices.Agents {
			if agent.Host != host {
				continue
			}

			// Create a map of current services from the agent
			currentServices := make(map[string]bool)
			for uid, services := range agentMap {
				for _, service := range services {
					key := makeReplicaKey(uid, service.Pid)
					currentServices[key] = true
				}
			}

			// Remove dependencies that are not in the current services
			for key := range agent.Dependencies {
				if !currentServices[key] {
					delete(agent.Dependencies, key)
				}
			}

			// If agent has no more dependencies, remove it
			if len(agent.Dependencies) == 0 {
				dcServices.Agents = append(dcServices.Agents[:i], dcServices.Agents[i+1:]...)
				i-- // Adjust index after removal
			}
		}

		// If no more agents in DC, remove DC
		if len(dcServices.Agents) == 0 {
			delete(s.DCMap, dcName)
		}
	}

	// Update generation and timestamp
	s.Generation++
	s.LastUpdate = time.Now()

	return nil
}

func (s *serviceMap) Upsert(service *Service, agentHost string) error {
	s.mut.Lock()
	defer s.mut.Unlock()

	dcServices, ok := s.DCMap[service.DC]
	if !ok {
		slog.Info("создаем новую карту для датацентра", "dc", service.DC)
		s.DCMap = make(DCMap, 0)
		dcServices = new(DCServices)
		dcServices.Agents = make([]*Agent, 0)
		dcServices.Agents = append(dcServices.Agents, &Agent{
			Host:         agentHost,
			Dependencies: map[string]*ServiceReplica{},
		})
	}

	uptime := time.Since(time.Unix(service.StartedAt, 0)).String()

	agents := dcServices.Agents
	for n, agent := range agents {
		if agent.Host != agentHost {
			continue
		}
		key := makeReplicaKey(service.Uid, service.Pid)
		svc, ok := agent.Dependencies[key]
		if !ok {
			slog.Info("обновляем реплику в зависимостях агента", "service", service)
			replica := &ServiceReplica{
				ServiceUid:  service.Uid,
				Pid:         service.Pid,
				Name:        service.Name,
				AgentHost:   agentHost,
				Project:     service.Project,
				Service:     service.Name,
				Version:     service.Version,
				Uptime:      uptime,
				Healthy:     true,
				Status:      service.Status,
				PortHTTP:    service.PortHTTP,
				PortGrpc:    service.PortGrpc,
				PortHTTPS:   service.PortHTTPS,
				EnableHTTPS: service.EnableHTTPS,
				Enviroment:  service.Enviroment,
				DC:          service.DC,
				Mask:        service.Mask,
				StartedAt:   service.StartedAt,
			}
			agent.Dependencies[key] = replica
		} else {
			svc.Healthy = true
			svc.Status = service.Status
			svc.Uptime = uptime
			agent.Dependencies[key] = svc
		}

		dcServices.Agents[n] = agent
	}

	s.DCMap[service.DC] = dcServices

	s.LastUpdate = time.Now()

	return nil
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
			for _, replica := range agent.Dependencies {
				if replica.Status == "running" {
					runningReplicas++
				}
			}
		}
	}
	s.mut.Unlock()

	if totalDesiredReplicas == 0 {
		return 1.0
	}

	return float32(runningReplicas) / float32(totalDesiredReplicas)
}
