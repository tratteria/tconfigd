package rules

import (
	"fmt"
)

// TODO
// 1. Add validation if JSON paths are valid and referencing valid variables i.e body, header, or queryParameters
// 2. Add validation if referenced url-path parameters are correct
// 3. Add validation if route is correct i.e begins with leading slash

var validMethods = map[string]struct{}{
	"GET":     {},
	"POST":    {},
	"PUT":     {},
	"DELETE":  {},
	"PATCH":   {},
	"OPTIONS": {},
	"HEAD":    {},
}

func isValidHTTPMethod(method string) bool {
	_, exists := validMethods[method]

	return exists
}

func validateGenerationRules(genRules map[string]GenerationRule, traTs map[string]TraTDefinition) error {
	for _, genRule := range genRules {
		traT, exist := traTs[genRule.TraTName]
		if !exist {
			return fmt.Errorf("invalid TraT %s referenced in TraT generation rule for route: %s, method: %s", genRule.TraTName, genRule.Route, genRule.Method)
		}

		if !isValidHTTPMethod(genRule.Method) {
			return fmt.Errorf("invalid HTTP method %s in generation rule for route: %s", genRule.Method, genRule.Route)
		}

		if err := validateAdzMapping(genRule.AdzMapping, traT); err != nil {
			return fmt.Errorf("invalid adz mapping in TraT generation rule; traT: %s, route: %s, method: %s; error: %s", genRule.TraTName, genRule.Route, genRule.Method, err.Error())
		}
	}

	return nil
}

func validateVerificationRules(verRules map[string]map[string]VerificationRule, traTs map[string]TraTDefinition) error {
	for _, serviceVerRules := range verRules {
		for _, verRule := range serviceVerRules {
			if !isValidHTTPMethod(verRule.Method) {
				return fmt.Errorf("invalid HTTP method %s in verification rule for route: %s", verRule.Method, verRule.Route)
			}

			for _, rule := range verRule.Rules {
				traT, exist := traTs[rule.TraTName]
				if !exist {
					return fmt.Errorf("invalid TraT %s referenced in TraT verification rule for route: %s, method: %s", rule.TraTName, verRule.Route, verRule.Method)
				}

				if err := validateAdzMapping(rule.AdzMapping, traT); err != nil {
					return fmt.Errorf("invalid adz mapping in TraT verification rule; traT: %s, route: %s, method: %s; error: %s", rule.TraTName, verRule.Route, verRule.Method, err.Error())
				}
			}
		}
	}

	return nil
}

func validateAdzMapping(adzMapping interface{}, traT TraTDefinition) error {
	switch v := adzMapping.(type) {
	case map[string]interface{}:
		for field := range v {
			if _, ok := traT.AdzSchema[field]; !ok {
				return fmt.Errorf("invalid key '%s' in generation rule adz mapping", field)
			}
		}
	case map[string]VerificationAdzField:
		for field := range v {
			if _, ok := traT.AdzSchema[field]; !ok {
				return fmt.Errorf("invalid key '%s' in verification rule adz mapping", field)
			}
		}
	default:
		return fmt.Errorf("unknown type for adz mapping")
	}

	return nil
}
