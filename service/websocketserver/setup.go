package websocketserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/tokenetes/tconfigd/common"
	tokenetes1alpha1 "github.com/tokenetes/tconfigd/tokenetescontroller/pkg/apis/tokenetes/v1alpha1"
	ruleretriever "github.com/tokenetes/tconfigd/tokenetescontroller/ruleretriever"
	"go.uber.org/zap"
)

const (
	API_PORT        = 8443
	WRITE_WAIT      = 10 * time.Second
	PONG_WAIT       = 60 * time.Second
	PING_PERIOD     = (PONG_WAIT * 9) / 10
	REQUEST_TIMEOUT = 15 * time.Second
)

type ClientsRetriever interface {
	GetClientManagers(service, namespace string) []*ClientManager
	GetTokenetesAgentServices(namespace string) []string
}

type WebSocketServer struct {
	ruleRetriever     ruleretriever.RuleRetriever
	X509Source        *workloadapi.X509Source
	TokenetesSpiffeId spiffeid.ID
	Logger            *zap.Logger
	ClientManagers    map[string]map[string][]*ClientManager
	clientsMutex      sync.RWMutex
}

func NewWebSocketServer(
	ruleRetriever ruleretriever.RuleRetriever,
	x509Source *workloadapi.X509Source,
	tokenetesSpiffeId spiffeid.ID,
	logger *zap.Logger,
) *WebSocketServer {
	return &WebSocketServer{
		ruleRetriever:     ruleRetriever,
		X509Source:        x509Source,
		TokenetesSpiffeId: tokenetesSpiffeId,
		Logger:            logger,
		ClientManagers:    make(map[string]map[string][]*ClientManager),
		clientsMutex:      sync.RWMutex{},
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.TLS != nil && len(r.TLS.PeerCertificates) > 0
	},
}

func (wss *WebSocketServer) Run() error {
	router := mux.NewRouter()

	wss.initializeRoutes(router)

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%d", API_PORT),
		TLSConfig:    tlsconfig.MTLSServerConfig(wss.X509Source, wss.X509Source, tlsconfig.AuthorizeAny()),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	wss.Logger.Info("Starting websocket server...", zap.Int("port", API_PORT))

	if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		wss.Logger.Error("Failed to start the websocket server", zap.Error(err))

		return fmt.Errorf("failed to start the websocket server: %w", err)
	}

	return nil
}

func (wss *WebSocketServer) initializeRoutes(router *mux.Router) {
	router.HandleFunc("/ws", wss.handleWebSocket)
}

func (wss *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		http.Error(w, "No client certificate provided", http.StatusUnauthorized)

		return
	}

	spiffeID, err := spiffeid.FromURI(r.TLS.PeerCertificates[0].URIs[0])
	if err != nil {
		wss.Logger.Error("Failed to parse SPIFFE ID", zap.Error(err))
		http.Error(w, "Invalid SPIFFE ID", http.StatusBadRequest)

		return
	}

	serviceName := strings.TrimPrefix(spiffeID.Path(), "/")

	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		wss.Logger.Error("Namespace not provided in WebSocket connection")
		http.Error(w, "Namespace is required", http.StatusBadRequest)

		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wss.Logger.Error("Failed to upgrade connection to WebSocket",
			zap.String("namespace", namespace),
			zap.String("service", serviceName),
			zap.Error(err))

		return
	}

	client := NewClient(serviceName, namespace, conn, wss)

	wss.clientsMutex.Lock()

	if wss.ClientManagers[serviceName] == nil {
		wss.ClientManagers[serviceName] = make(map[string][]*ClientManager)
	}

	wss.ClientManagers[serviceName][namespace] = append(wss.ClientManagers[serviceName][namespace], client)

	wss.clientsMutex.Unlock()

	//nolint:unparam
	sendErrorAndClose := func(errorMsg string, statusCode int) {
		errorResponse, err := json.Marshal(Response{
			ID:      uuid.New().String(),
			Type:    MessageTypeInitialRulesResponse,
			Status:  statusCode,
			Payload: json.RawMessage(fmt.Sprintf(`{"error": "%s"}`, errorMsg)),
		})

		if err != nil {
			wss.Logger.Error("Error marshaling error response", zap.Error(err))
		} else {
			client.writeMessage(websocket.TextMessage, errorResponse)
		}

		client.Close()
	}

	// The retrieved rules are guaranteed to incorporate changes up to and including this version number
	var activeRuleVersionNumber int64

	var activeGenerationRules *tokenetes1alpha1.GenerationRules

	var activeVerificationRules *tokenetes1alpha1.VerificationRules

	initialRulesPayload := &AllActiveRulesPayload{}

	if serviceName == common.TOKENETES_SERVICE_NAME {
		activeGenerationRules, activeRuleVersionNumber, err = wss.ruleRetriever.GetActiveGenerationRules(namespace)
		if err != nil {
			wss.Logger.Error("Error getting initial generation rules from controller", zap.Error(err))
			sendErrorAndClose("Failed to retrieve initial generation rules", http.StatusInternalServerError)

			return
		}

		initialRulesPayload.GenerationRules = activeGenerationRules
	} else {
		activeVerificationRules, activeRuleVersionNumber, err = wss.ruleRetriever.GetActiveVerificationRules(serviceName, namespace)
		if err != nil {
			wss.Logger.Error("Error getting initial verification rules from controller", zap.Error(err))
			sendErrorAndClose("Failed to retrieve initial verification rules", http.StatusInternalServerError)

			return
		}

		initialRulesPayload.VerificationRules = activeVerificationRules
	}

	initialRulesPayloadJSON, err := json.Marshal(initialRulesPayload)
	if err != nil {
		wss.Logger.Error("Error marshaling initial rules payload", zap.Error(err))
		sendErrorAndClose("Failed to prepare initial rules payload", http.StatusInternalServerError)

		return
	}

	initialConfigMsg, err := json.Marshal(Response{
		ID:      uuid.New().String(),
		Type:    MessageTypeInitialRulesResponse,
		Status:  http.StatusCreated,
		Payload: initialRulesPayloadJSON,
	})
	if err != nil {
		wss.Logger.Error("Error marshaling initial rules response", zap.Error(err))
		sendErrorAndClose("Failed to prepare initial rules response", http.StatusInternalServerError)

		return
	}

	err = client.writeMessage(websocket.TextMessage, initialConfigMsg)
	if err != nil {
		wss.Logger.Error("Failed to send initial rules response", zap.Error(err))
		sendErrorAndClose("Failed to send initial rules response", http.StatusInternalServerError)

		return
	}

	// Assigning the version number to the client. This indicates that the client has rules incorporated up to and including this version number.
	atomic.StoreInt64(&client.RuleVersionNumber, activeRuleVersionNumber)

	wss.Logger.Info("Client connected and initial configuration sent successfully", zap.String("namespace", client.Namespace), zap.String("service", client.Service))

	go client.writePump()
	go client.readPump()
}

func (wss *WebSocketServer) removeClient(client *ClientManager) {
	wss.clientsMutex.Lock()
	defer wss.clientsMutex.Unlock()

	if clients, ok := wss.ClientManagers[client.Service][client.Namespace]; ok {
		for i, c := range clients {
			if c == client {
				wss.ClientManagers[client.Service][client.Namespace] = append(clients[:i], clients[i+1:]...)

				break
			}
		}

		if len(wss.ClientManagers[client.Service][client.Namespace]) == 0 {
			delete(wss.ClientManagers[client.Service], client.Namespace)
		}

		if len(wss.ClientManagers[client.Service]) == 0 {
			delete(wss.ClientManagers, client.Service)
		}
	}

	client.Close()
}

func (wss *WebSocketServer) GetClientManagers(service, namespace string) []*ClientManager {
	wss.clientsMutex.RLock()
	defer wss.clientsMutex.RUnlock()

	if serviceClients, ok := wss.ClientManagers[service]; ok {
		if namespaceClients, ok := serviceClients[namespace]; ok {
			return namespaceClients
		}
	}

	return nil
}

func (wss *WebSocketServer) GetTokenetesAgentServices(namespace string) []string {
	wss.clientsMutex.RLock()
	defer wss.clientsMutex.RUnlock()

	var tokenetesAgentsServices []string

	for service := range wss.ClientManagers {
		if service != common.TOKENETES_SERVICE_NAME {
			tokenetesAgentsServices = append(tokenetesAgentsServices, service)
		}
	}

	return tokenetesAgentsServices
}
