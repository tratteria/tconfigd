package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/tratteria/tconfigd/api/pkg/service"

	"go.uber.org/zap"
)

type Handlers struct {
	Service *service.Service
	Logger  *zap.Logger
}

func NewHandlers(service *service.Service, logger *zap.Logger) *Handlers {
	return &Handlers{
		Service: service,
		Logger:  logger,
	}
}

type registrationRequest struct {
	IpAddress string `json:"ipAddress"`
	Port      int    `json:"port"`
	NameSpace string `json:"namespace"`
}

type heartBeatRequest struct {
	IpAddress      string `json:"ipAddress"`
	Port           int    `json:"port"`
	NameSpace      string `json:"namespace"`
	RulesVersionID string `json:"rulesVersionId"`
}

func (h *Handlers) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		http.Error(w, "No client certificate provided", http.StatusUnauthorized)

		return
	}

	spiffeID, err := spiffeid.FromURI(r.TLS.PeerCertificates[0].URIs[0])
	if err != nil {
		h.Logger.Error("Failed to parse SPIFFE ID", zap.Error(err))
		http.Error(w, "Invalid SPIFFE ID", http.StatusBadRequest)

		return
	}

	serviceName := strings.TrimPrefix(spiffeID.Path(), "/")

	var registrationRequest registrationRequest

	err = json.NewDecoder(r.Body).Decode(&registrationRequest)
	if err != nil {
		h.Logger.Error("Invalid registration request.", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)

		return
	}

	h.Logger.Info("Received a registration request.", zap.String("service", serviceName))

	registrationResponse, err := h.Service.RegisterService(registrationRequest.IpAddress, registrationRequest.Port, serviceName, registrationRequest.NameSpace)
	if err != nil {
		h.Logger.Error("Failed to register service.", zap.Error(err))
		http.Error(w, "Failed to register service", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(registrationResponse); err != nil {
		h.Logger.Error("Failed to encode registration response.", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)

		return
	}
}

func (h *Handlers) HeartBeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		http.Error(w, "No client certificate provided", http.StatusUnauthorized)

		return
	}

	spiffeID, err := spiffeid.FromURI(r.TLS.PeerCertificates[0].URIs[0])
	if err != nil {
		h.Logger.Error("Failed to parse SPIFFE ID", zap.Error(err))
		http.Error(w, "Invalid SPIFFE ID", http.StatusBadRequest)

		return
	}

	serviceName := strings.TrimPrefix(spiffeID.Path(), "/")

	var heartBeatRequest heartBeatRequest

	err = json.NewDecoder(r.Body).Decode(&heartBeatRequest)
	if err != nil {
		h.Logger.Error("Invalid heartbeat request.", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)

		return
	}

	h.Logger.Info("Received a heartbeat request.", zap.String("service", serviceName))

	h.Service.RegisterHeartBeat(heartBeatRequest.IpAddress, heartBeatRequest.Port, serviceName, heartBeatRequest.NameSpace)

	// TODO: if an agent is heartbeating with an old rule version id, notify it to fetch the latest rules

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetJwksHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Get-Jwks request received.")

	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		http.Error(w, "Namespace parameter is required", http.StatusBadRequest)

		return
	}

	ctx := r.Context()

	jwks, err := h.Service.CollectJwks(ctx, namespace)
	if err != nil {
		h.Logger.Error("Error collecting JWKS from tratteria instances.", zap.Error(err))
		http.Error(w, "Error collecting JWKS", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(jwks); err != nil {
		h.Logger.Error("Failed to encode response of a get-jwks request.", zap.Error(err))

		return
	}

	h.Logger.Info("Get-Jwks request processed successfully.")
}
