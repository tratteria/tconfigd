package rules

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func parse(path string) (map[string]TraTDefinition, map[string]GenerationRule, map[string]map[string]VerificationRule, error) {
	traTs := make(map[string]TraTDefinition)
	generationRules := make(map[string]GenerationRule)
	verificationRules := make(map[string]map[string]VerificationRule)

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, err
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)

	for {
		var jsonData map[string]interface{}

		if err := decoder.Decode(&jsonData); err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, nil, err
		}

		switch jsonType := jsonData["type"]; jsonType {
		case "TraT":
			var definition TraTDefinition

			jsonBytes, err := json.Marshal(jsonData)

			if err != nil {
				return nil, nil, nil, err
			}

			if err := json.Unmarshal(jsonBytes, &definition); err != nil {
				return nil, nil, nil, err
			}

			traTs[definition.TraTName] = definition

		case "TraT-Generation-Rule":
			var genRule GenerationRule

			jsonBytes, err := json.Marshal(jsonData)

			if err != nil {
				return nil, nil, nil, err
			}

			if err := json.Unmarshal(jsonBytes, &genRule); err != nil {
				return nil, nil, nil, err
			}

			key := genRule.Method + genRule.Route

			if _, exists := generationRules[key]; exists {
				return nil, nil, nil, fmt.Errorf("multiple generation rules for route: %s,  method: %s provided", genRule.Route, genRule.Method)
			}

			generationRules[key] = genRule

		case "TraT-Verification-Rule":
			var verRule VerificationRule

			jsonBytes, err := json.Marshal(jsonData)

			if err != nil {
				return nil, nil, nil, err
			}

			if err := json.Unmarshal(jsonBytes, &verRule); err != nil {
				return nil, nil, nil, err
			}

			serviceVerificatinRules, exists := verificationRules[verRule.Service]
			if !exists {
				serviceVerificatinRules = make(map[string]VerificationRule)
				verificationRules[verRule.Service] = serviceVerificatinRules
			}

			key := verRule.Method + verRule.Route
			if _, exists := serviceVerificatinRules[key]; exists {
				return nil, nil, nil, fmt.Errorf("multiple verification rules for service: %s, route: %s,  method: %s provided", verRule.Service, verRule.Route, verRule.Method)
			}

			serviceVerificatinRules[key] = verRule
		}
	}

	return traTs, generationRules, verificationRules, nil
}