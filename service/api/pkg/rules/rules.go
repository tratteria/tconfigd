package rules

import (
	"github.com/tratteria/tconfigd/api/pkg/apierrors"
)

type Rules struct {
	traTs             map[string]TraTDefinition
	generationRules   map[string]GenerationRule
	verificationRules map[string]map[string]VerificationRule
}

func NewRules() *Rules {
	return &Rules{
		traTs:             make(map[string]TraTDefinition),
		generationRules:   make(map[string]GenerationRule),
		verificationRules: make(map[string]map[string]VerificationRule),
	}
}

func (r *Rules) Load() error {
	traTs, generationRules, verificationRules, err := parse("/etc/rules/trats-rules.ndjson")
	if err != nil {
		return err
	}

	if err := validateGenerationRules(generationRules, traTs); err != nil {
		return err
	}

	if err := validateVerificationRules(verificationRules, traTs); err != nil {
		return err
	}

	r.traTs = traTs
	r.generationRules = generationRules
	r.verificationRules = verificationRules

	return nil
}

func (r *Rules) GetVerificationRules(service string) (map[string]VerificationRule, error) {
	verificationRule, exist := r.verificationRules[service]
	if !exist {
		return nil, apierrors.ErrVerificationRuleNotFound
	}

	return verificationRule, nil
}

func (r *Rules) GetGenerationRules() map[string]GenerationRule {
	return r.generationRules
}
