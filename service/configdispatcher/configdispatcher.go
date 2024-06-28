package configdispatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tratteria/tconfigd/common"
	"github.com/tratteria/tconfigd/dataplaneregistry"
	"github.com/tratteria/tconfigd/tratcontroller/pkg/apis/tratteria/v1alpha1"
)

const (
	CONFIG_WEBHOOK_ENDPOINT = "/config-webhook"
)

type ConfigDispatcher struct {
	activeAgentRetriever dataplaneregistry.Retriever
	httpClient           *http.Client
}

func NewConfigDispatcher(activeAgentRetriever dataplaneregistry.Retriever, httpClient *http.Client) *ConfigDispatcher {
	return &ConfigDispatcher{
		activeAgentRetriever: activeAgentRetriever,
		httpClient:           httpClient,
	}
}

// TODO: Implement parallel processing of HTTP requests using goroutines.
func (cd *ConfigDispatcher) DispatchVerificationRules(ctx context.Context, serviceName string, verificationConfig *v1alpha1.VerificationRule) error {
	agents, err := cd.activeAgentRetriever.GetActiveEntries(serviceName)
	if err != nil {
		return fmt.Errorf("unable to retrieve active agents for service %s: %w", serviceName, err)
	}

	for _, agent := range agents {
		url := fmt.Sprintf("http://%s:%d%s", agent.IpAddress, agent.Port, CONFIG_WEBHOOK_ENDPOINT)

		jsonData, err := json.Marshal(verificationConfig)
		if err != nil {
			return fmt.Errorf("error marshaling verification config: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error creating request for a %s agent: %w", serviceName, err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := cd.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request to a %s agent: %w", serviceName, err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received non-ok status %d from a %s agent", resp.StatusCode, serviceName)
		}
	}

	return nil
}

// TODO: Implement parallel processing of HTTP requests using goroutines.
func (cd *ConfigDispatcher) DispatchGenerationRule(ctx context.Context, generationConfig *v1alpha1.GenerationRule) error {
	tratteriaInstances, err := cd.activeAgentRetriever.GetActiveEntries(common.TRATTERIA_SERVICE_NAME)
	if err != nil {
		return fmt.Errorf("unable to retrieve active tratteria instances: %w", err)
	}

	for _, instance := range tratteriaInstances {
		url := fmt.Sprintf("http://%s:%d%s", instance.IpAddress, instance.Port, CONFIG_WEBHOOK_ENDPOINT)

		jsonData, err := json.Marshal(generationConfig)
		if err != nil {
			return fmt.Errorf("error marshaling generation config: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error creating request for a tratteria instance: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := cd.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request to a tratteria instance: %w", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received non-ok status %d from a tratteria instance", resp.StatusCode)
		}
	}

	return nil
}
