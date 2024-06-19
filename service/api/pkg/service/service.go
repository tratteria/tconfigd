package service

import (
	"github.com/tratteria/tconfigd/agentsmanager"
	"github.com/tratteria/tconfigd/api/pkg/rules"
	"go.uber.org/zap"
)

type Service struct {
	rules         *rules.Rules
	agentsManager *agentsmanager.AgentsManager
	logger        *zap.Logger
}

func NewService(rules *rules.Rules, agentsManager *agentsmanager.AgentsManager, logger *zap.Logger) *Service {
	return &Service{
		rules:         rules,
		agentsManager: agentsManager,
		logger:        logger,
	}
}

func (s *Service) GetVerificationRule(service string) (map[string]rules.VerificationRule, error) {
	return s.rules.GetVerificationRules(service)
}

func (s *Service) GetGenerationRule() map[string]rules.GenerationRule {
	return s.rules.GetGenerationRules()
}

func (s *Service) RegisterAgent(ipaddress string, port int, service string) {
	// TODO: return rules belonging to the service
	s.agentsManager.RegisterAgent(ipaddress, port, service)
}

func (s *Service) RegisterHeartBeat(ipaddress string, port int) {
	// TODO: if updateHeartBeat fails, notify agent to register
	// TODO: if an agent is heartbeating with an old rule version id, notify it to fetch the latest rules
	s.agentsManager.UpdateHeartbeat(ipaddress, port)
}
