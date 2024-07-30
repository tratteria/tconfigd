package ruledispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

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
func (rd *RuleDispatcher) dispatchRule(ctx context.Context, serviceName string, namespace string, messageType websocketserver.MessageType, rule json.RawMessage, versionNumber int64) error {
	clients := rd.clientsRetriever.GetClientManagers(serviceName, namespace)

	var dispatchErrors []string
	/*
		Only propagate rules to clients that don't already have the change. If a client's version number
		is equal to or greater than this particular change's version number, the client already
		incorporates this change.

		Clients can incorporate changes without being explicitly pushed through the reconciliation process.

		Reconciliation serves as a backup process for propagating changes and correcting any out-of-order
		or missed propagations.

		If a client's rule hash is found to be different, the client replaces its complete rule set with
		the latest rule set and adopts the version number associated with the rules it just received.
	*/
	for _, client := range clients {
		if atomic.LoadInt64(&client.RuleVersionNumber) < versionNumber {
			_, err := client.SendRequest(messageType, rule)

			if err != nil {
				dispatchErrors = append(dispatchErrors, err.Error())
			}
		}
	}

	if len(dispatchErrors) > 0 {
		return fmt.Errorf("error dispatching config to %s service: %s", serviceName, strings.Join(dispatchErrors, ", "))
	}

	return nil
}

func (rd *RuleDispatcher) DispatchTraTVerificationRule(ctx context.Context, serviceName string, namespace string, verificationEndpointRule *v1alpha1.TraTVerificationRule, versionNumber int64) error {
	jsonData, err := json.Marshal(verificationEndpointRule)
	if err != nil {
		return fmt.Errorf("error marshaling verification trat rule: %w", err)
	}

	err = rd.dispatchRule(ctx, serviceName, namespace, websocketserver.MessageTypeTraTVerificationRuleUpsertRequest, jsonData, versionNumber)
	if err != nil {
		return fmt.Errorf("error dispatching verification trat rule to %s service: %w", serviceName, err)
	}

	return nil
}

func (rd *RuleDispatcher) DispatchTratteriaConfigVerificationRule(ctx context.Context, namespace string, verificationTokenRule *v1alpha1.TratteriaConfigVerificationRule, versionNumber int64) error {
	jsonData, err := json.Marshal(verificationTokenRule)
	if err != nil {
		return fmt.Errorf("error marshaling verification tratteria config rule: %w", err)
	}

	var dispatchErrors []string

	for _, serviceName := range rd.clientsRetriever.GetTratteriaAgentServices(namespace) {
		err = rd.dispatchRule(ctx, serviceName, namespace, websocketserver.MessageTypeTratteriaConfigVerificationRuleUpsertRequest, jsonData, versionNumber)
		if err != nil {
			dispatchErrors = append(dispatchErrors, fmt.Sprintf("error dispatching verification token rule to %s service: %v", serviceName, err))
		}
	}

	if len(dispatchErrors) > 0 {
		return fmt.Errorf("error dispatching verification tratteria config rule: %s", strings.Join(dispatchErrors, ", "))
	}

	return nil
}

func (rd *RuleDispatcher) DispatchTraTGenerationRule(ctx context.Context, namespace string, generationEndpointRule *v1alpha1.TraTGenerationRule, verisionNumber int64) error {
	jsonData, err := json.Marshal(generationEndpointRule)
	if err != nil {
		return fmt.Errorf("error marshaling generation trat rule: %w", err)
	}

	err = rd.dispatchRule(ctx, common.TRATTERIA_SERVICE_NAME, namespace, websocketserver.MessageTypeTraTGenerationRuleUpsertRequest, jsonData, verisionNumber)
	if err != nil {
		return fmt.Errorf("error dispatching generation trat rule to tratteria: %w", err)
	}

	return nil
}

func (rd *RuleDispatcher) DispatchTratteriaConfigGenerationRule(ctx context.Context, namespace string, generationTokenRule *v1alpha1.TratteriaConfigGenerationRule, versionNumber int64) error {
	jsonData, err := json.Marshal(generationTokenRule)
	if err != nil {
		return fmt.Errorf("error marshaling generation tratteria config rule: %w", err)
	}

	err = rd.dispatchRule(ctx, common.TRATTERIA_SERVICE_NAME, namespace, websocketserver.MessageTypeTratteriaConfigGenerationRuleUpsertRequest, jsonData, versionNumber)
	if err != nil {
		return fmt.Errorf("error dispatching generation tratteria config rule to tratteria: %w", err)
	}

	return nil
}
