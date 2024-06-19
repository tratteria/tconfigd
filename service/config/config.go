package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v2"
)

type BoolFromString bool

func (b *BoolFromString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tmp interface{}
	if err := unmarshal(&tmp); err != nil {
		return err
	}

	switch value := tmp.(type) {
	case bool:
		*b = BoolFromString(value)
	case string:
		if matched, envVarName := extractEnvVarName(value); matched {
			envValue, err := getEnvVarValue(envVarName)
			if err != nil {
				return err
			}

			boolVal, err := strconv.ParseBool(envValue)

			if err != nil {
				return fmt.Errorf("error parsing boolean from environment variable: %v", err)
			}

			*b = BoolFromString(boolVal)
		} else {
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("error parsing boolean from string: %v", err)
			}

			*b = BoolFromString(boolVal)
		}
	default:
		return fmt.Errorf("invalid type for a bool variable, expected bool or string, got %T", tmp)
	}

	return nil
}

type IntFromString int

func (i *IntFromString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var strValue string

	if err := unmarshal(&strValue); err != nil {
		return err
	}

	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		return err
	}

	*i = IntFromString(intValue)

	return nil
}

type AppConfig struct {
	EnableTratInterception BoolFromString `yaml:"enableTratInterception"`
	AgentApiPort           IntFromString  `yaml:"agentApiPort"`
	AgentInterceptorPort   IntFromString  `yaml:"agentInterceptorPort"`
}

func GetAppConfig(configPath string) (*AppConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML configuration: %w", err)
	}

	resolveEnvVariables(&cfg)

	return &cfg, nil
}

func extractEnvVarName(s string) (bool, string) {
	envVarRegex := regexp.MustCompile(`^\$\{([^}]+)\}$`)
	matches := envVarRegex.FindStringSubmatch(s)

	if len(matches) > 1 {
		return true, matches[1]
	}

	return false, ""
}

func getEnvVarValue(envVarName string) (string, error) {
	if envValue, exists := os.LookupEnv(envVarName); exists {
		return envValue, nil
	}

	return "", fmt.Errorf("environment variable %s not set", envVarName)
}

func resolveEnvVariablesUtil(v reflect.Value) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String {
			fieldValue := field.String()

			if matched, envVarName := extractEnvVarName(fieldValue); matched {
				envValue, err := getEnvVarValue(envVarName)
				if err != nil {
					panic(err.Error())
				}

				field.SetString(envValue)
			}
		} else if field.Kind() == reflect.Struct {
			resolveEnvVariablesUtil(field)
		} else if field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
			resolveEnvVariablesUtil(field.Elem())
		}
	}
}

func resolveEnvVariables(cfg *AppConfig) {
	v := reflect.ValueOf(cfg)
	resolveEnvVariablesUtil(v)
}
