package service

import (
	"github.com/tratteria/tratD/pkg/rules"
	"go.uber.org/zap"
)

type Service struct {
	Rules  *rules.Rules
	Logger *zap.Logger
}

func NewService(rules *rules.Rules, logger *zap.Logger) *Service {
	return &Service{
		Rules:  rules,
		Logger: logger,
	}
}

func (s *Service) GetVerificationRule(service string) (map[string]rules.VerificationRule, error) {
	return s.Rules.GetVerificationRules(service)
}

func (s *Service) GetGenerationRule() map[string]rules.GenerationRule {
	return s.Rules.GetGenerationRules()
}
