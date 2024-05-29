package rules

import (
	"github.com/tratteria/tratd/pkg/apperrors"
)

type Rules struct {
	dir               string
	traTs             map[string]TraTDefinition
	generationRules   map[string]GenerationRule
	verificationRules map[string]map[string]VerificationRule
}

func NewRules(dir string) *Rules {
	return &Rules{
		dir:               dir,
		traTs:             make(map[string]TraTDefinition),
		generationRules:   make(map[string]GenerationRule),
		verificationRules: make(map[string]map[string]VerificationRule),
	}
}

func (r *Rules) Load() error {
	traTs, generationRules, verificationRules, err := parse(r.dir)
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
		return nil, apperrors.ErrVerificationRuleNotFound
	}

	return verificationRule, nil
}

func (r *Rules) GetGenerationRules() map[string]GenerationRule {

	return r.generationRules
}
