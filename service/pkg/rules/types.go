package rules

import (
	"encoding/json"
	"fmt"
)

type TraTDefinition struct {
	Type      string                 `json:"type"`
	TraTName  string                 `json:"traT-name"`
	Scope     string                 `json:"scope"`
	AdzSchema map[string]interface{} `json:"adz-schema"`
}

type GenerationRule struct {
	Type       string                 `json:"type"`
	Route      string                 `json:"route"`
	Method     string                 `json:"method"`
	TraTName   string                 `json:"traT-name"`
	AdzMapping map[string]interface{} `json:"adz-mapping"`
}

type VerificationRule struct {
	Type    string `json:"type"`
	Service string `json:"service"`
	Route   string `json:"route"`
	Method  string `json:"method"`
	Rules   []Rule `json:"rules"`
}

type Rule struct {
	TraTName   string                          `json:"traT-name"`
	AdzMapping map[string]VerificationAdzField `json:"adz-mapping"`
}

type VerificationAdzField struct {
	Path     string `json:"path"`
	Required bool   `json:"required"`
}

func (v *VerificationAdzField) UnmarshalJSON(data []byte) error {
	var obj struct {
		Path     string `json:"path"`
		Required bool   `json:"required"`
	}

	if err := json.Unmarshal(data, &obj); err == nil {
		*v = VerificationAdzField{Path: obj.Path, Required: obj.Required}

		return nil
	}

	var path string

	if err := json.Unmarshal(data, &path); err == nil {
		*v = VerificationAdzField{Path: path, Required: false}

		return nil
	}

	return fmt.Errorf("cannot unmarshal verification rule adz-mapping VerificationAdzField: %s", data)
}
