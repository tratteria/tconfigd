package configdispatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tratteria/tconfigd/agentsmanager"
	"github.com/tratteria/tconfigd/tratcontroller/pkg/apis/tratteria/v1alpha1"
)

const (
	verificationConfigWebhookEndpoint = "/config-webhook"
)

type ConfigDispatcher struct {
	activeAgentRetriever agentsmanager.ActiveAgentRetriever
	httpClient           *http.Client
}

func NewConfigDispatcher(activeAgentRetriever agentsmanager.ActiveAgentRetriever, httpClient *http.Client) *ConfigDispatcher {
	return &ConfigDispatcher{
		activeAgentRetriever: activeAgentRetriever,
		httpClient:           httpClient,
	}
}

// TODO: Implement parallel processing of HTTP requests using goroutines.
func (cd *ConfigDispatcher) DispatchVerificationRules(ctx context.Context, serviceName string, verificationConfig *v1alpha1.VerificationRule) error {
	agents, err := cd.activeAgentRetriever.GetServiceActiveAgents(serviceName)
	if err != nil {
		return fmt.Errorf("unable to retrieve active agents for service %s: %w", serviceName, err)
	}

	for _, agent := range agents {
		url := fmt.Sprintf("http://%s:%d%s", agent.IpAddress, agent.Port, verificationConfigWebhookEndpoint)

		jsonData, err := json.Marshal(verificationConfig)
		if err != nil {
			return fmt.Errorf("error marshaling verification config: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error creating request for agent %s: %w", agent.IpAddress, err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := cd.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request to agent %s: %w", agent.IpAddress, err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received non-ok status %d from agent %s", resp.StatusCode, agent.IpAddress)
		}
	}

	return nil
}

func (cd *ConfigDispatcher) DispatchGenerationRule(ctx context.Context, generationRule *v1alpha1.GenerationRule) error {
	return nil
}
