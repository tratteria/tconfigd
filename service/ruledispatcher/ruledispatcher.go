package ruledispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tratteria/tconfigd/common"
	"github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"
	"github.com/tratteria/tconfigd/websocketserver"
)

type RuleDispatcher struct {
	clientsRetriever websocketserver.ClientsRetriever
}

func NewRuleDispatcher() *RuleDispatcher {
	return &RuleDispatcher{}
}

func (rd *RuleDispatcher) SetClientsRetriever(clientsRetriever websocketserver.ClientsRetriever) {
	rd.clientsRetriever = clientsRetriever
}

//nolint:unparam
func (rd *RuleDispatcher) dispatchRule(ctx context.Context, serviceName string, namespace string, messageType websocketserver.MessageType, rule json.RawMessage) error {
	clients := rd.clientsRetriever.GetClientManagers(serviceName, namespace)

	var dispatchErrors []string

	for _, client := range clients {
		_, err := client.SendRequest(messageType, rule)

		if err != nil {
			dispatchErrors = append(dispatchErrors, err.Error())
		}
	}

	if len(dispatchErrors) > 0 {
		return fmt.Errorf("error dispatching config to %s service: %s", serviceName, strings.Join(dispatchErrors, ", "))
	}

	return nil
}

func (rd *RuleDispatcher) DispatchTraTVerificationRule(ctx context.Context, serviceName string, namespace string, verificationEndpointRule *v1alpha1.TraTVerificationRule) error {
	jsonData, err := json.Marshal(verificationEndpointRule)
	if err != nil {
		return fmt.Errorf("error marshaling verification trat rule: %w", err)
	}

	err = rd.dispatchRule(ctx, serviceName, namespace, websocketserver.MessageTypeTraTVerificationRuleUpsertRequest, jsonData)
	if err != nil {
		return fmt.Errorf("error dispatching verification trat rule to %s service: %w", serviceName, err)
	}

	return nil
}

func (rd *RuleDispatcher) DispatchTratteriaConfigVerificationRule(ctx context.Context, namespace string, verificationTokenRule *v1alpha1.TratteriaConfigVerificationRule) error {
	jsonData, err := json.Marshal(verificationTokenRule)
	if err != nil {
		return fmt.Errorf("error marshaling verification tratteria config rule: %w", err)
	}

	var dispatchErrors []string

	for _, serviceName := range rd.clientsRetriever.GetTratteriaAgentServices(namespace) {
		err = rd.dispatchRule(ctx, serviceName, namespace, websocketserver.MessageTypeTratteriaConfigVerificationRuleUpsertRequest, jsonData)
		if err != nil {
			dispatchErrors = append(dispatchErrors, fmt.Sprintf("error dispatching verification token rule to %s service: %v", serviceName, err))
		}
	}

	if len(dispatchErrors) > 0 {
		return fmt.Errorf("error dispatching verification tratteria config rule: %s", strings.Join(dispatchErrors, ", "))
	}

	return nil
}

func (rd *RuleDispatcher) DispatchTraTGenerationRule(ctx context.Context, namespace string, generationEndpointRule *v1alpha1.TraTGenerationRule) error {
	jsonData, err := json.Marshal(generationEndpointRule)
	if err != nil {
		return fmt.Errorf("error marshaling generation trat rule: %w", err)
	}

	err = rd.dispatchRule(ctx, common.TRATTERIA_SERVICE_NAME, namespace, websocketserver.MessageTypeTraTGenerationRuleUpsertRequest, jsonData)
	if err != nil {
		return fmt.Errorf("error dispatching generation trat rule to tratteria: %w", err)
	}

	return nil
}

func (rd *RuleDispatcher) DispatchTratteriaConfigGenerationRule(ctx context.Context, namespace string, generationTokenRule *v1alpha1.TratteriaConfigGenerationRule) error {
	jsonData, err := json.Marshal(generationTokenRule)
	if err != nil {
		return fmt.Errorf("error marshaling generation tratteria config rule: %w", err)
	}

	err = rd.dispatchRule(ctx, common.TRATTERIA_SERVICE_NAME, namespace, websocketserver.MessageTypeTratteriaConfigGenerationRuleUpsertRequest, jsonData)
	if err != nil {
		return fmt.Errorf("error dispatching generation tratteria config rule to tratteria: %w", err)
	}

	return nil
}
