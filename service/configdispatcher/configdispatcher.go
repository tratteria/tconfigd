package configdispatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tratteria/tconfigd/common"
	"github.com/tratteria/tconfigd/dataplaneregistry"
	"github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"

	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

const (
	VERIFICATION_ENDPOINT_RULE_WEBHOOK_ENDPOINT = "/verification-trat-rule-webhook"
	VERIFICATION_TOKEN_RULE_WEBHOOK_ENDPOINT    = "/verification-tratteria-config-rule-webhook"
	GENERATION_ENDPOINT_RULE_WEBHOOK_ENDPOINT   = "/generation-trat-rule-webhook"
	GENERATION_TOKEN_RULE_WEBHOOK_ENDPOINT      = "/generation-tratteria-config-rule-webhook"
)

type ConfigDispatcher struct {
	dataplaneRegistryRetriever dataplaneregistry.Retriever
	httpClient                 *http.Client
}

func NewConfigDispatcher(dataplaneRegistryRetriever dataplaneregistry.Retriever, x509Source *workloadapi.X509Source) *ConfigDispatcher {
	tlsConfig := tlsconfig.MTLSClientConfig(x509Source, x509Source, tlsconfig.AuthorizeAny())

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return &ConfigDispatcher{
		dataplaneRegistryRetriever: dataplaneRegistryRetriever,
		httpClient:                 &client,
	}
}

func (cd *ConfigDispatcher) dispatchConfigUtil(ctx context.Context, url string, config json.RawMessage) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(config))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := cd.httpClient.Do(req)

	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}

		return fmt.Errorf("received non-ok status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// TODO: Implement parallel processing of HTTP requests using goroutines.
func (cd *ConfigDispatcher) dispatchConfig(ctx context.Context, serviceName string, namespace string, endpoint string, config json.RawMessage) error {
	entries := cd.dataplaneRegistryRetriever.GetActiveEntries(serviceName, namespace)

	var dispatchErrors []string

	for _, entry := range entries {
		url := fmt.Sprintf("https://%s:%d%s", entry.IpAddress, entry.Port, endpoint)
		err := cd.dispatchConfigUtil(ctx, url, config)

		if err != nil {
			dispatchErrors = append(dispatchErrors, err.Error())
		}
	}

	if len(dispatchErrors) > 0 {
		return fmt.Errorf("error dispatching config to %s service: %s", serviceName, strings.Join(dispatchErrors, ", "))
	}

	return nil
}

func (cd *ConfigDispatcher) DispatchVerificationTraTRule(ctx context.Context, serviceName string, namespace string, verificationEndpointRule *v1alpha1.VerificationTraTRule) error {
	jsonData, err := json.Marshal(verificationEndpointRule)
	if err != nil {
		return fmt.Errorf("error marshaling verification trat rule: %w", err)
	}

	err = cd.dispatchConfig(ctx, serviceName, namespace, VERIFICATION_ENDPOINT_RULE_WEBHOOK_ENDPOINT, jsonData)
	if err != nil {
		return fmt.Errorf("error dispatching verification trat rule to %s service: %w", serviceName, err)
	}

	return nil
}

func (cd *ConfigDispatcher) DispatchVerificationTratteriaConfigRule(ctx context.Context, namespace string, verificationTokenRule *v1alpha1.VerificationTratteriaConfigRule) error {
	jsonData, err := json.Marshal(verificationTokenRule)
	if err != nil {
		return fmt.Errorf("error marshaling verification tratteria config rule: %w", err)
	}

	var dispatchErrors []string

	for _, serviceName := range cd.dataplaneRegistryRetriever.GetAgentServices(namespace) {
		err = cd.dispatchConfig(ctx, serviceName, namespace, VERIFICATION_TOKEN_RULE_WEBHOOK_ENDPOINT, jsonData)
		if err != nil {
			dispatchErrors = append(dispatchErrors, fmt.Sprintf("error dispatching verification token rule to %s service: %v", serviceName, err))
		}
	}

	if len(dispatchErrors) > 0 {
		return fmt.Errorf("error dispatching verification tratteria config rule: %s", strings.Join(dispatchErrors, ", "))
	}

	return nil
}

func (cd *ConfigDispatcher) DispatchGenerationTraTRule(ctx context.Context, namespace string, generationEndpointRule *v1alpha1.GenerationTraTRule) error {
	jsonData, err := json.Marshal(generationEndpointRule)
	if err != nil {
		return fmt.Errorf("error marshaling generation trat rule: %w", err)
	}

	err = cd.dispatchConfig(ctx, common.TRATTERIA_SERVICE_NAME, namespace, GENERATION_ENDPOINT_RULE_WEBHOOK_ENDPOINT, jsonData)
	if err != nil {
		return fmt.Errorf("error dispatching generation trat rule to tratteria: %w", err)
	}

	return nil
}

func (cd *ConfigDispatcher) DispatchGenerationTratteriaConfigRule(ctx context.Context, namespace string, generationTokenRule *v1alpha1.GenerationTratteriaConfigRule) error {
	jsonData, err := json.Marshal(generationTokenRule)
	if err != nil {
		return fmt.Errorf("error marshaling generation tratteria config rule: %w", err)
	}

	err = cd.dispatchConfig(ctx, common.TRATTERIA_SERVICE_NAME, namespace, GENERATION_TOKEN_RULE_WEBHOOK_ENDPOINT, jsonData)
	if err != nil {
		return fmt.Errorf("error dispatching generation tratteria config rule to tratteria: %w", err)
	}

	return nil
}
