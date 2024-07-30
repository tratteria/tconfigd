package ruleretriever

import (
	tratteria1alpha1 "github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"
)

type RuleRetriever interface {
	GetActiveVerificationRules(serviceName string, namespace string) (*tratteria1alpha1.VerificationRules, int64, error)
	GetActiveVerificationRulesHash(serviceName string, namespace string) (string, int64, error)
	GetActiveGenerationRules(namespace string) (*tratteria1alpha1.GenerationRules, int64, error)
	GetActiveGenerationRulesHash(namespace string) (string, int64, error)
}
