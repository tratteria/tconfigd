package websocketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/tratteria/tconfigd/common"
	tratteria1alpha1 "github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"
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
	MessageTypeTratteriaConfigVerificationRuleUpsertRequest  MessageType = "TRATTERIA_CONFIG_VERIFICATION_RULE_UPSERT_REQUEST"
	MessageTypeTratteriaConfigVerificationRuleUpsertResponse MessageType = "TRATTERIA_CONFIG_VERIFICATION_RULE_UPSERT_RESPONSE"
	MessageTypeTratteriaConfigGenerationRuleUpsertRequest    MessageType = "TRATTERIA_CONFIG_GENERATION_RULE_UPSERT_REQUEST"
	MessageTypeTratteriaConfigGenerationRuleUpsertResponse   MessageType = "TRATTERIA_CONFIG_GENERATION_RULE_UPSERT_RESPONSE"
)

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

type InitialRulesPayload struct {
	VerificationRules *tratteria1alpha1.VerificationRules `json:"verificationRules,omitempty"`
	GenerationRules   *tratteria1alpha1.GenerationRules   `json:"generationRules,omitempty"`
}

type ClientManager struct {
	Namespace       string
	Service         string
	Conn            *websocket.Conn
	Send            chan []byte
	Server          *WebSocketServer
	pendingRequests sync.Map
	closeOnce       sync.Once
	done            chan struct{}
}

func NewClient(service string, namespace string, conn *websocket.Conn, server *WebSocketServer) *ClientManager {
	return &ClientManager{
		Namespace: namespace,
		Service:   service,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Server:    server,
		done:      make(chan struct{}),
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
	cm.Conn.SetPongHandler(func(string) error {
		cm.Conn.SetReadDeadline(time.Now().Add(PONG_WAIT))

		return nil
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

func (cm *ClientManager) writePump() {
	ticker := time.NewTicker(PING_PERIOD)

	defer func() {
		ticker.Stop()
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
		case <-ticker.C:
			if err := cm.writeMessage(websocket.PingMessage, nil); err != nil {
				cm.Server.Logger.Error("Failed to write ping message.", zap.Error(err))

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
		MessageTypeTratteriaConfigVerificationRuleUpsertResponse,
		MessageTypeTratteriaConfigGenerationRuleUpsertResponse,
		MessageTypeGetJWKSResponse:
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
	tratteriaInstances := cm.Server.GetClientManagers(common.TRATTERIA_SERVICE_NAME, cm.Namespace)
	if len(tratteriaInstances) == 0 {
		cm.Server.Logger.Warn("No active tratteria instances found", zap.String("namespace", cm.Namespace))
		cm.sendErrorResponse(request.ID, MessageTypeGetJWKSResponse, http.StatusNotFound, "No active tratteria instances found")

		return
	}

	allKeys := jwk.NewSet()

	for _, instance := range tratteriaInstances {
		response, err := instance.SendRequest(MessageTypeGetJWKSRequest, nil)
		if err != nil {
			cm.Server.Logger.Error("Error getting JWKS from a tratteria instance", zap.Error(err))

			continue
		}

		if response.Status != http.StatusOK {
			cm.Server.Logger.Error("Received non-OK status code from a tratteria instance", zap.Int("status", response.Status))

			continue
		}

		set, err := jwk.Parse(response.Payload)
		if err != nil {
			cm.Server.Logger.Error("Failed to parse JWKS from a tratteria instance", zap.Error(err))

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
		cm.Server.Logger.Warn("No valid keys found from any tratteria instance", zap.String("namespace", cm.Namespace))
		cm.sendErrorResponse(request.ID, MessageTypeGetJWKSResponse, http.StatusInternalServerError, "No valid keys found")

		return
	}

	err := cm.sendResponse(request.ID, MessageTypeGetJWKSResponse, http.StatusOK, allKeys)
	if err != nil {
		cm.Server.Logger.Error("Error sending JWKS response to tratteria-agent",
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
