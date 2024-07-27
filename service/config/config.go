package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
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
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("error parsing boolean from string: %v", err)
		}

		*b = BoolFromString(boolVal)
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

type SPIFFEIDFromString spiffeid.ID

func (s *SPIFFEIDFromString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var strValue string
	if err := unmarshal(&strValue); err != nil {
		return err
	}

	spiffeID, err := spiffeid.FromString(strValue)
	if err != nil {
		return fmt.Errorf("error parsing SPIFFE ID from string: %v", err)
	}

	*s = SPIFFEIDFromString(spiffeID)

	return nil
}

type Config struct {
	SpireAgentHostDir      string             `yaml:"spireAgentHostDir"`
	TratteriaSpiffeId      SPIFFEIDFromString `yaml:"tratteriaSpiffeId"`
	EnableTratInterception BoolFromString     `yaml:"enableTratInterception"`
	AgentHttpApiPort       IntFromString      `yaml:"agentHttpApiPort"`
	AgentInterceptorPort   IntFromString      `yaml:"agentInterceptorPort"`
}

func GetConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML configuration: %w", err)
	}

	return &cfg, nil
}
