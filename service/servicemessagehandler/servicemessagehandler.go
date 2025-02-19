package servicemessagehandler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/tokenetes/tconfigd/common"
	"github.com/tokenetes/tconfigd/tokenetescontroller/pkg/apis/tokenetes/v1alpha1"
	"github.com/tokenetes/tconfigd/websocketserver"
)

type ServiceMessageHandler struct {
	clientsRetriever websocketserver.ClientsRetriever
}

func NewServiceMessageHandler() *ServiceMessageHandler {
	return &ServiceMessageHandler{}
}

type TraTDeletionRequestMessage struct {
	TraTName string
}

func (smh *ServiceMessageHandler) SetClientsRetriever(clientsRetriever websocketserver.ClientsRetriever) {
	smh.clientsRetriever = clientsRetriever
}

//nolint:unparam
func (smh *ServiceMessageHandler) sendMessage(ctx context.Context, serviceName string, namespace string, messageType websocketserver.MessageType, rule json.RawMessage, versionNumber int64) error {
	clients := smh.clientsRetriever.GetClientManagers(serviceName, namespace)

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

func (smh *ServiceMessageHandler) DispatchTraTVerificationRule(ctx context.Context, serviceName string, namespace string, serviceTraTVerificationRules *v1alpha1.ServiceTraTVerificationRules, versionNumber int64) error {
	jsonData, err := json.Marshal(serviceTraTVerificationRules)
	if err != nil {
		return fmt.Errorf("error marshaling verification trat rule: %w", err)
	}

	err = smh.sendMessage(ctx, serviceName, namespace, websocketserver.MessageTypeTraTVerificationRuleUpsertRequest, jsonData, versionNumber)
	if err != nil {
		return fmt.Errorf("error dispatching verification trat rule to %s service: %w", serviceName, err)
	}

	return nil
}

func (smh *ServiceMessageHandler) DeleteTraT(ctx context.Context, serviceName string, namespace string, traTName string, versionNumber int64) error {
	traTDeletionRequestMessage := TraTDeletionRequestMessage{TraTName: traTName}

	jsonData, err := json.Marshal(traTDeletionRequestMessage)
	if err != nil {
		return fmt.Errorf("error marshaling verification trat rule: %w", err)
	}

	err = smh.sendMessage(ctx, serviceName, namespace, websocketserver.MessageTypeTraTDeletionRequest, jsonData, versionNumber)
	if err != nil {
		return fmt.Errorf("error deleting trat from %s service: %w", serviceName, err)
	}

	return nil
}

func (smh *ServiceMessageHandler) DispatchTokenetesConfigVerificationRule(ctx context.Context, namespace string, verificationTokenRule *v1alpha1.TokenetesConfigVerificationRule, versionNumber int64) error {
	jsonData, err := json.Marshal(verificationTokenRule)
	if err != nil {
		return fmt.Errorf("error marshaling verification tokenetes config rule: %w", err)
	}

	var dispatchErrors []string

	for _, serviceName := range smh.clientsRetriever.GetTokenetesAgentServices(namespace) {
		err = smh.sendMessage(ctx, serviceName, namespace, websocketserver.MessageTypeTokenetesConfigVerificationRuleUpsertRequest, jsonData, versionNumber)
		if err != nil {
			dispatchErrors = append(dispatchErrors, fmt.Sprintf("error dispatching verification token rule to %s service: %v", serviceName, err))
		}
	}

	if len(dispatchErrors) > 0 {
		return fmt.Errorf("error dispatching verification tokenetes config rule: %s", strings.Join(dispatchErrors, ", "))
	}

	return nil
}

func (smh *ServiceMessageHandler) DispatchTraTGenerationRule(ctx context.Context, namespace string, traTGenerationRule *v1alpha1.TraTGenerationRule, verisionNumber int64) error {
	jsonData, err := json.Marshal(traTGenerationRule)
	if err != nil {
		return fmt.Errorf("error marshaling generation trat rule: %w", err)
	}

	err = smh.sendMessage(ctx, common.TOKENETES_SERVICE_NAME, namespace, websocketserver.MessageTypeTraTGenerationRuleUpsertRequest, jsonData, verisionNumber)
	if err != nil {
		return fmt.Errorf("error dispatching generation trat rule to tokenetes: %w", err)
	}

	return nil
}

func (smh *ServiceMessageHandler) DispatchTokenetesConfigGenerationRule(ctx context.Context, namespace string, generationTokenRule *v1alpha1.TokenetesConfigGenerationRule, versionNumber int64) error {
	jsonData, err := json.Marshal(generationTokenRule)
	if err != nil {
		return fmt.Errorf("error marshaling generation tokenetes config rule: %w", err)
	}

	err = smh.sendMessage(ctx, common.TOKENETES_SERVICE_NAME, namespace, websocketserver.MessageTypeTokenetesConfigGenerationRuleUpsertRequest, jsonData, versionNumber)
	if err != nil {
		return fmt.Errorf("error dispatching generation tokenetes config rule to tokenetes: %w", err)
	}

	return nil
}

func (smh *ServiceMessageHandler) DispatchTraTExclRule(ctx context.Context, serviceName string, namespace string, traTExclRule *v1alpha1.TraTExclRule, versionNumber int64) error {
	jsonData, err := json.Marshal(traTExclRule)
	if err != nil {
		return fmt.Errorf("error marshaling tratexcl rule: %w", err)
	}

	err = smh.sendMessage(ctx, serviceName, namespace, websocketserver.MessageTypeTraTExclRuleUpsertRequest, jsonData, versionNumber)
	if err != nil {
		return fmt.Errorf("error dispatching tratexcl rule to %s service: %w", serviceName, err)
	}

	return nil
}

func (smh *ServiceMessageHandler) DeleteTraTExcl(ctx context.Context, serviceName string, namespace string, versionNumber int64) error {
	err := smh.sendMessage(ctx, serviceName, namespace, websocketserver.MessageTypeTraTExclRuleDeleteRequest, nil, versionNumber)
	if err != nil {
		return fmt.Errorf("error deleting tratexcl rule from %s service: %w", serviceName, err)
	}

	return nil
}
