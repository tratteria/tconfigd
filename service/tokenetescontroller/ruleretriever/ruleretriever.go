package ruleretriever

import (
	tokenetes1alpha1 "github.com/tokenetes/tconfigd/tokenetescontroller/pkg/apis/tokenetes/v1alpha1"
)

type RuleRetriever interface {
	GetActiveVerificationRules(serviceName string, namespace string) (*tokenetes1alpha1.VerificationRules, int64, error)
	GetActiveVerificationRulesHash(serviceName string, namespace string) (string, int64, error)
	GetActiveGenerationRules(namespace string) (*tokenetes1alpha1.GenerationRules, int64, error)
	GetActiveGenerationRulesHash(namespace string) (string, int64, error)
}
