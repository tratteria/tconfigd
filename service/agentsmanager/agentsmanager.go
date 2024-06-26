package agentsmanager

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/tratteria/tconfigd/common"
	"github.com/tratteria/tconfigd/tconfigderrors"
)

const (
	EXPIRATION_BUFFER   = 2 * time.Minute
	EXPIRATION_DURATION = time.Duration(common.AGENT_HEARTBEAT_INTERVAL_MINUTES)*time.Minute + EXPIRATION_BUFFER
	CLEANUP_INTERVAL    = 5 * EXPIRATION_DURATION
)

type Agent struct {
	IpAddress     string
	Port          int
	ServiceName   string
	LastHeartbeat time.Time
}

type AgentLifecycleManager interface {
	RegisterAgent(string, int, string)
	UpdateAgentHeartbeat(string, int, string)
}

type ActiveAgentRetriever interface {
	GetServiceActiveAgents(string) ([]*Agent, error)
}

type AgentsManager struct {
	agents map[string]map[string]*Agent
	mutex  sync.RWMutex
}

func NewAgentManager() *AgentsManager {
	am := &AgentsManager{
		agents: make(map[string]map[string]*Agent),
	}

	go am.cleanupExpiredAgents()

	return am
}

func (am *AgentsManager) RegisterAgent(ip string, port int, serviceName string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if am.agents[serviceName] == nil {
		am.agents[serviceName] = make(map[string]*Agent)
	}

	key := ip + strconv.Itoa(port)

	am.agents[serviceName][key] = &Agent{
		IpAddress:     ip,
		Port:          port,
		ServiceName:   serviceName,
		LastHeartbeat: time.Now(),
	}
}

func (am *AgentsManager) UpdateAgentHeartbeat(ip string, port int, serviceName string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := ip + strconv.Itoa(port)

	if agent, ok := am.agents[serviceName][key]; ok {
		agent.LastHeartbeat = time.Now()
	} // TODO: return error if agent entry not found
}

func (am *AgentsManager) GetServiceActiveAgents(serviceName string) ([]*Agent, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var activeAgents []*Agent

	now := time.Now()

	serviceAgents, ok := am.agents[serviceName]
	if !ok {
		return nil, fmt.Errorf("%w: service %s", tconfigderrors.ErrNotFound, serviceName)
	}

	for _, agent := range serviceAgents {
		if now.Sub(agent.LastHeartbeat) <= EXPIRATION_DURATION {
			activeAgents = append(activeAgents, agent)
		}
	}

	return activeAgents, nil
}

func (am *AgentsManager) cleanupExpiredAgents() {
	ticker := time.NewTicker(CLEANUP_INTERVAL)
	defer ticker.Stop()

	for {
		<-ticker.C
		am.removeExpiredAgents()
	}
}

func (am *AgentsManager) removeExpiredAgents() {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	now := time.Now()

	for _, serviceAgents := range am.agents {
		for key, agent := range serviceAgents {
			if now.Sub(agent.LastHeartbeat) > EXPIRATION_DURATION {
				delete(serviceAgents, key)
			}
		}
	}
}
