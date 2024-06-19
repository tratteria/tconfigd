package agentsmanager

import (
	"strconv"
	"sync"
	"time"

	"github.com/tratteria/tconfigd/common"
)

const (
	EXPIRATION_BUFFER   = 2 * time.Minute
	EXPIRATION_DURATION = time.Duration(common.AGENT_HEARTBEAT_INTERVAL_MINUTES)*time.Minute + EXPIRATION_BUFFER
	CLEANUP_INTERVAL    = 5 * EXPIRATION_DURATION
)

type AgentRegistry struct {
	IpAddress     string
	Port          int
	ServiceName   string
	LastHeartbeat time.Time
}

type AgentsManager struct {
	agents map[string]*AgentRegistry
	mutex  sync.RWMutex
}

func NewAgentManager() *AgentsManager {
	am := &AgentsManager{
		agents: make(map[string]*AgentRegistry),
	}

	go am.cleanupExpiredAgents()

	return am
}

func (am *AgentsManager) RegisterAgent(ip string, port int, serviceName string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := ip + strconv.Itoa(port)

	am.agents[key] = &AgentRegistry{
		IpAddress:     ip,
		Port:          port,
		ServiceName:   serviceName,
		LastHeartbeat: time.Now(),
	}
}

func (am *AgentsManager) UpdateHeartbeat(ip string, port int) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	key := ip + strconv.Itoa(port)

	if agent, ok := am.agents[key]; ok {
		agent.LastHeartbeat = time.Now()
	} // TODO: return error if agent entry not found
}

func (am *AgentsManager) GetActiveAgents() map[string]*AgentRegistry {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	activeAgents := make(map[string]*AgentRegistry)
	now := time.Now()

	for key, agent := range am.agents {
		if now.Sub(agent.LastHeartbeat) <= EXPIRATION_DURATION {
			activeAgents[key] = agent
		}
	}

	return activeAgents
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

	for key, agent := range am.agents {
		if now.Sub(agent.LastHeartbeat) > EXPIRATION_DURATION {
			delete(am.agents, key)
		}
	}
}
