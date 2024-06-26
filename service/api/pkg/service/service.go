package service

import (
	"github.com/tratteria/tconfigd/agentsmanager"
	"go.uber.org/zap"
)

type Service struct {
	agentsLifecycleManager agentsmanager.AgentLifecycleManager
	logger                 *zap.Logger
}

func NewService(agentsLifecycleManager agentsmanager.AgentLifecycleManager, logger *zap.Logger) *Service {
	return &Service{
		agentsLifecycleManager: agentsLifecycleManager,
		logger:                 logger,
	}
}

func (s *Service) RegisterAgent(ipaddress string, port int, serviceName string) {
	// TODO: return rules belonging to the service
	s.agentsLifecycleManager.RegisterAgent(ipaddress, port, serviceName)
}

func (s *Service) RegisterHeartBeat(ipaddress string, port int, serviceName string) {
	// TODO: if updateHeartBeat fails, notify agent to register
	// TODO: if an agent is heartbeating with an old rule version id, notify it to fetch the latest rules
	s.agentsLifecycleManager.UpdateAgentHeartbeat(ipaddress, port, serviceName)
}
