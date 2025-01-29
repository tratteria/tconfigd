/**
Rule Propagation Approach(Here, rule means general rules for constructing and validating TraTs.
It is the combination of all Tokenetes resources i.e TraTs, TokenetesConfig, and TraTExclusion):

When tconfigd starts (or restarts, or when the leader changes if multiple replicas are present), it initializes an
in-memory global rule version number starting at zero. Similarly, for each WebSocket client (i.e., Tokenetes Service
and Tokenetes Agents), it assigns a client rule version number also starting at zero. These rule version numbers
represent the state of the rules, with higher numbers indicating more recent states. These version numbers are
maintained only in tconfigd's memory and are not relevant to other components.

When an Add, Update, or Delete operation of a Custom Resource (CR) occurs, tconfigd's informer is notified,
updates its cache, and invokes the respective handler. Each handler performs the following:

1. Increments the global rule version number by 1 and assigns this version number to the operation.
2. Loops through the clients that need to receive this change.
3. If the client's rule version number is less than the operation's version number, push the change to the client.
4. If all pushes succeed, marks the operation as done. If any push fails, marks the operation as pending and
   requeues it to the work queue.

In addition to individual change propagation, regular rule validation and reconciliation are performed.
Each client performs the following:

1. Every 60 seconds (configurable through tconfigd static configuration), the client sends a WebSocket ping to the
   WebSocket server with its rule hash.
2. The WebSocket server replies with a WebSocket pong.
3. The WebSocket server compares the received hash with the latest hash. If they are the same, it updates the
   client's version number to the latest hash rule version number.
4. If the received hash and latest hash are different, it performs a reconciliation request to the client and
   updates the client version number to the reconciled rules version number.

The regular rule validation and reconciliation serve the following purposes:

1. It acts as a backup to the individual change propagation. If certain individual change propagations fail,
   they will be propagated through the reconciliation process.
2. If any rule propagation fails or if rules are propagated in an incorrect order resulting in inconsistent rules
   on the client side, the reconciliation process will correct the client's rules.

The reconciliation process ensures eventual consistency of the rules in the system.
*/

package websocketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/tokenetes/tconfigd/common"
	tokenetes1alpha1 "github.com/tokenetes/tconfigd/tokenetescontroller/pkg/apis/tokenetes/v1alpha1"
	"go.uber.org/zap"
)

type MessageType string

const (
	MessageTypeInitialRulesResponse                          MessageType = "INITIAL_RULES_RESPONSE"
	MessageTypeGetJWKSRequest                                MessageType = "GET_JWKS_REQUEST"
	MessageTypeGetJWKSResponse                               MessageType = "GET_JWKS_RESPONSE"
	MessageTypeTraTVerificationRuleUpsertRequest             MessageType = "TRAT_VERIFICATION_RULE_UPSERT_REQUEST"
	MessageTypeTraTVerificationRuleUpsertResponse            MessageType = "TRAT_VERIFICATION_RULE_UPSERT_RESPONSE"
	MessageTypeTraTGenerationRuleUpsertRequest               MessageType = "TRAT_GENERATION_RULE_UPSERT_REQUEST"
	MessageTypeTraTGenerationRuleUpsertResponse              MessageType = "TRAT_GENERATION_RULE_UPSERT_RESPONSE"
	MessageTypeTokenetesConfigVerificationRuleUpsertRequest  MessageType = "TOKENETES_CONFIG_VERIFICATION_RULE_UPSERT_REQUEST"
	MessageTypeTokenetesConfigVerificationRuleUpsertResponse MessageType = "TOKENETES_CONFIG_VERIFICATION_RULE_UPSERT_RESPONSE"
	MessageTypeTokenetesConfigGenerationRuleUpsertRequest    MessageType = "TOKENETES_CONFIG_GENERATION_RULE_UPSERT_REQUEST"
	MessageTypeTokenetesConfigGenerationRuleUpsertResponse   MessageType = "TOKENETES_CONFIG_GENERATION_RULE_UPSERT_RESPONSE"
	MessageTypeRuleReconciliationRequest                     MessageType = "RULE_RECONCILIATION_REQUEST"
	MessageTypeRuleReconciliationResponse                    MessageType = "RULE_RECONCILIATION_RESPONSE"
	MessageTypeTraTDeletionRequest                           MessageType = "TRAT_DELETION_REQUEST"
	MessageTypeTraTDeletionResponse                          MessageType = "TRAT_DELETION_RESPONSE"
	MessageTypeTraTExclRuleUpsertRequest                     MessageType = "TRAT_EXCL_RULE_UPSERT_REQUEST"
	MessageTypeTraTExclRuleUpsertResponse                    MessageType = "TRAT_EXCL_RULE_UPSERT_RESPONSE"
	MessageTypeTraTExclRuleDeleteRequest                     MessageType = "TRAT_EXCL_RULE_DELETE_REQUEST"
	MessageTypeTraTExclRuleDeleteResponse                    MessageType = "TRAT_EXCL_RULE_DELETE_RESPONSE"
	MessageTypeUnknown                                       MessageType = "UNKNOWN"
)

type PingData struct {
	RuleHash string `json:"ruleHash"`
}

type Request struct {
	ID      string          `json:"id"`
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Response struct {
	ID      string          `json:"id"`
	Type    MessageType     `json:"type"`
	Status  int             `json:"status"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type AllActiveRulesPayload struct {
	VerificationRules *tokenetes1alpha1.VerificationRules `json:"verificationRules,omitempty"`
	GenerationRules   *tokenetes1alpha1.GenerationRules   `json:"generationRules,omitempty"`
}

type ClientManager struct {
	Namespace         string
	Service           string
	Conn              *websocket.Conn
	Send              chan []byte
	Server            *WebSocketServer
	pendingRequests   sync.Map
	closeOnce         sync.Once
	done              chan struct{}
	RuleVersionNumber int64
}

func NewClient(service string, namespace string, conn *websocket.Conn, server *WebSocketServer) *ClientManager {
	return &ClientManager{
		Namespace:         namespace,
		Service:           service,
		Conn:              conn,
		Send:              make(chan []byte, 256),
		Server:            server,
		done:              make(chan struct{}),
		RuleVersionNumber: 0,
	}
}

func (cm *ClientManager) Close() {
	cm.closeOnce.Do(func() {
		cm.Conn.Close()
		close(cm.Send)
		close(cm.done)
		cm.Server.Logger.Info("Client connection closed and resources released", zap.String("namespace", cm.Namespace), zap.String("service", cm.Service))
	})
}

func (cm *ClientManager) readPump() {
	defer func() {
		cm.Server.removeClient(cm)
	}()

	cm.Conn.SetReadDeadline(time.Now().Add(PONG_WAIT))
	cm.Conn.SetPingHandler(func(appData string) error {
		cm.Conn.SetReadDeadline(time.Now().Add(PONG_WAIT))

		go cm.compareAndReconcileRule(appData)

		return cm.writeMessage(websocket.PongMessage, nil)
	})

	for {
		select {
		case <-cm.done:
			return
		default:
			_, message, err := cm.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					cm.Server.Logger.Error("WebSocket connection closed unexpectedly", zap.Error(err))
				}

				return
			}

			cm.handleMessage(message)
		}
	}
}

func (cm *ClientManager) compareAndReconcileRule(appData string) {
	var pingData PingData

	err := json.Unmarshal([]byte(appData), &pingData)
	if err != nil {
		cm.Server.Logger.Error("Failed to unmarshal ping data", zap.Error(err))

		return
	}

	var lateshHash string

	var activeRuleVersionNumber int64

	if cm.Service == common.TOKENETES_SERVICE_NAME {
		lateshHash, activeRuleVersionNumber, _ = cm.Server.ruleRetriever.GetActiveGenerationRulesHash(cm.Namespace)
	} else {
		lateshHash, activeRuleVersionNumber, _ = cm.Server.ruleRetriever.GetActiveVerificationRulesHash(cm.Service, cm.Namespace)
	}

	if lateshHash != pingData.RuleHash {
		cm.Server.Logger.Warn("Received ping with incorrect rule hash, triggering reconciliation...",
			zap.String("service", cm.Service),
			zap.String("namespace", cm.Namespace))

		err := cm.reconcileRules()
		if err != nil {
			cm.Server.Logger.Error("Failed to reconcile rules", zap.Error(err))
		}
	} else {
		atomic.StoreInt64(&cm.RuleVersionNumber, activeRuleVersionNumber)
	}
}

func (cm *ClientManager) reconcileRules() error {
	var err error

	var activeRuleVersionNumber int64

	var completeGenerationRules *tokenetes1alpha1.GenerationRules

	var completeVerificationRules *tokenetes1alpha1.VerificationRules

	allActiveRulesPayload := &AllActiveRulesPayload{}

	if cm.Service == common.TOKENETES_SERVICE_NAME {
		completeGenerationRules, activeRuleVersionNumber, err = cm.Server.ruleRetriever.GetActiveGenerationRules(cm.Namespace)
		if err != nil {
			cm.Server.Logger.Error("Error getting all active generation rules from controller", zap.Error(err))

			return err
		}

		allActiveRulesPayload.GenerationRules = completeGenerationRules
	} else {
		completeVerificationRules, activeRuleVersionNumber, err = cm.Server.ruleRetriever.GetActiveVerificationRules(cm.Service, cm.Namespace)
		if err != nil {
			cm.Server.Logger.Error("Error getting all active verification rules from controller", zap.Error(err))

			return err
		}

		allActiveRulesPayload.VerificationRules = completeVerificationRules
	}

	response, err := cm.SendRequest(MessageTypeRuleReconciliationRequest, allActiveRulesPayload)
	if err != nil {
		cm.Server.Logger.Error("Failed to send rule reconciliation request",
			zap.String("namespace", cm.Namespace),
			zap.String("service", cm.Service),
			zap.Error(err))

		return err
	}

	if response.Status != http.StatusOK {
		cm.Server.Logger.Error("Received non-ok status from client on rule reconciliation request rule",
			zap.Int("status", response.Status),
			zap.String("namespace", cm.Namespace),
			zap.String("service", cm.Service),
			zap.Error(err))

		return err
	}

	cm.Server.Logger.Info("Client's rule successfully reconcilized", zap.String("namespace", cm.Namespace), zap.String("service", cm.Service))

	// After successful reconciliation, set the client version number to the reconcilized rule version, which represent the version number at least
	// up which the reconciled rule incorportates changes
	atomic.StoreInt64(&cm.RuleVersionNumber, activeRuleVersionNumber)

	return nil
}

func (cm *ClientManager) writePump() {
	defer func() {
		cm.Server.removeClient(cm)
	}()

	for {
		select {
		case <-cm.done:
			return
		case message, ok := <-cm.Send:
			if !ok {
				return
			}

			if err := cm.writeMessage(websocket.TextMessage, message); err != nil {
				cm.Server.Logger.Error("Failed to write message.", zap.Error(err))

				return
			}
		}
	}
}

func (cm *ClientManager) writeMessage(messageType int, data []byte) error {
	cm.Conn.SetWriteDeadline(time.Now().Add(WRITE_WAIT))

	return cm.Conn.WriteMessage(messageType, data)
}

func (cm *ClientManager) handleMessage(message []byte) {
	var temp struct {
		Type MessageType `json:"type"`
	}

	if err := json.Unmarshal(message, &temp); err != nil {
		cm.Server.Logger.Error("Failed to unmarshal message type.", zap.Error(err))

		return
	}

	switch temp.Type {
	case MessageTypeGetJWKSRequest:
		cm.handleRequest(message)
	case MessageTypeTraTVerificationRuleUpsertResponse,
		MessageTypeTraTGenerationRuleUpsertResponse,
		MessageTypeTokenetesConfigVerificationRuleUpsertResponse,
		MessageTypeTokenetesConfigGenerationRuleUpsertResponse,
		MessageTypeRuleReconciliationResponse,
		MessageTypeGetJWKSResponse,
		MessageTypeTraTDeletionResponse,
		MessageTypeTraTExclRuleUpsertResponse,
		MessageTypeTraTExclRuleDeleteResponse:
		cm.handleResponse(message)
	default:
		cm.Server.Logger.Info("Received unknown or unexpected message type", zap.String("type", string(temp.Type)))
	}
}

func (cm *ClientManager) handleRequest(message []byte) {
	var request Request

	if err := json.Unmarshal(message, &request); err != nil {
		cm.Server.Logger.Error("Failed to unmarshal request", zap.Error(err))

		return
	}

	cm.Server.Logger.Info("Received request", zap.String("id", request.ID), zap.String("type", string(request.Type)))

	switch request.Type {
	case MessageTypeGetJWKSRequest:
		cm.handleGetJWKSRequest(request)
	default:
		cm.Server.Logger.Error("Received unknown or unexpected request type", zap.String("type", string(request.Type)))
	}
}

func (cm *ClientManager) handleGetJWKSRequest(request Request) {
	tokenetesInstances := cm.Server.GetClientManagers(common.TOKENETES_SERVICE_NAME, cm.Namespace)
	if len(tokenetesInstances) == 0 {
		cm.Server.Logger.Warn("No active tokenetes instances found", zap.String("namespace", cm.Namespace))
		cm.sendErrorResponse(request.ID, MessageTypeGetJWKSResponse, http.StatusNotFound, "No active tokenetes instances found")

		return
	}

	allKeys := jwk.NewSet()

	for _, instance := range tokenetesInstances {
		response, err := instance.SendRequest(MessageTypeGetJWKSRequest, nil)
		if err != nil {
			cm.Server.Logger.Error("Error getting JWKS from a tokenetes instance", zap.Error(err))

			continue
		}

		if response.Status != http.StatusOK {
			cm.Server.Logger.Error("Received non-OK status code from a tokenetes instance", zap.Int("status", response.Status))

			continue
		}

		set, err := jwk.Parse(response.Payload)
		if err != nil {
			cm.Server.Logger.Error("Failed to parse JWKS from a tokenetes instance", zap.Error(err))

			continue
		}

		ctx := context.Background()
		for iter := set.Iterate(ctx); iter.Next(ctx); {
			pair := iter.Pair()
			if key, ok := pair.Value.(jwk.Key); ok {
				allKeys.Add(key)
			}
		}
	}

	if allKeys.Len() == 0 {
		cm.Server.Logger.Warn("No valid keys found from any tokenetes instance", zap.String("namespace", cm.Namespace))
		cm.sendErrorResponse(request.ID, MessageTypeGetJWKSResponse, http.StatusInternalServerError, "No valid keys found")

		return
	}

	err := cm.sendResponse(request.ID, MessageTypeGetJWKSResponse, http.StatusOK, allKeys)
	if err != nil {
		cm.Server.Logger.Error("Error sending JWKS response to tokenetes-agent",
			zap.String("request-id", request.ID),
			zap.String("namespace", cm.Namespace),
			zap.String("service", cm.Service),
			zap.Error(err))
	} else {
		cm.Server.Logger.Info("Successfully sent JWKS response",
			zap.String("request-id", request.ID),
			zap.String("namespace", cm.Namespace),
			zap.String("service", cm.Service))
	}
}

func (cm *ClientManager) handleResponse(message []byte) {
	var response Response

	if err := json.Unmarshal(message, &response); err != nil {
		cm.Server.Logger.Error("Failed to unmarshal response", zap.Error(err))

		return
	}

	if pending, ok := cm.pendingRequests.Load(response.ID); ok {
		if responseChan, ok := pending.(chan Response); ok {
			select {
			case responseChan <- response:
			default:
				cm.Server.Logger.Warn("Failed to send response, request might have timed out", zap.String("id", response.ID))
			}
		} else {
			cm.Server.Logger.Error("Pending request value is not of expected type", zap.String("id", response.ID))
		}
	} else {
		cm.Server.Logger.Error("Received response for unknown request, request might have timed out", zap.String("id", response.ID))
	}
}

func (cm *ClientManager) SendRequest(msgType MessageType, payload interface{}) (Response, error) {
	id := uuid.New().String()
	request := Request{
		ID:   id,
		Type: msgType,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Response{}, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	request.Payload = payloadBytes

	msgBytes, err := json.Marshal(request)
	if err != nil {
		return Response{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	responseChan := make(chan Response, 1)

	cm.pendingRequests.Store(id, responseChan)

	defer func() {
		cm.pendingRequests.Delete(id)
		close(responseChan)
	}()

	select {
	case cm.Send <- msgBytes:
		select {
		case response := <-responseChan:
			return response, nil
		case <-time.After(REQUEST_TIMEOUT):
			return Response{}, fmt.Errorf("request timeout")
		}
	default:
		return Response{}, fmt.Errorf("send channel is full")
	}
}

func (cm *ClientManager) sendResponse(id string, respType MessageType, status int, payload interface{}) error {
	var payloadJSON json.RawMessage

	if payload != nil {
		var err error

		payloadJSON, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal response payload: %w", err)
		}
	}

	response := Response{
		ID:      id,
		Type:    respType,
		Status:  status,
		Payload: payloadJSON,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	select {
	case cm.Send <- responseJSON:
		return nil
	default:
		return fmt.Errorf("send channel is full")
	}
}

func (cm *ClientManager) sendErrorResponse(requestID string, messageType MessageType, statusCode int, errorMessage string) {
	err := cm.sendResponse(requestID, messageType, statusCode, map[string]string{"error": errorMessage})
	if err != nil {
		cm.Server.Logger.Error("Failed to send error response",
			zap.String("request-id", requestID),
			zap.String("namespace", cm.Namespace),
			zap.String("service", cm.Service),
			zap.Error(err))
	}
}
