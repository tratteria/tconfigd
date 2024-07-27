package ruleretriever

import (
	tratteria1alpha1 "github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"
)

type RuleRetriever interface {
	GetActiveVerificationRules(serviceName string, namespace string) (*tratteria1alpha1.VerificationRules, error)
	GetActiveGenerationRules(namespace string) (*tratteria1alpha1.GenerationRules, error)
}
