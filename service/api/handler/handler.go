package handler

import (
	"encoding/json"
	"net/http"

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
	IpAddress   string `json:"ipAddress"`
	Port        int    `json:"port"`
	ServiceName string `json:"serviceName"`
}

type heartBeatRequest struct {
	IpAddress      string `json:"ipAddress"`
	Port           int    `json:"port"`
	ServiceName    string `json:"serviceName"`
	RulesVersionID string `json:"rulesVersionId"`
}

func (h *Handlers) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	var registrationRequest registrationRequest

	err := json.NewDecoder(r.Body).Decode(&registrationRequest)
	if err != nil {
		h.Logger.Error("Invalid registration request.", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)

		return
	}

	h.Logger.Info("Received a registration request.", zap.String("service", registrationRequest.ServiceName))

	h.Service.RegisterAgent(registrationRequest.IpAddress, registrationRequest.Port, registrationRequest.ServiceName)

	// TODO: return rules belonging to this service

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) HeartBeatHandler(w http.ResponseWriter, r *http.Request) {
	var heartBeatRequest heartBeatRequest

	err := json.NewDecoder(r.Body).Decode(&heartBeatRequest)
	if err != nil {
		h.Logger.Error("Invalid heartbeat request.", zap.Error(err))
		http.Error(w, "Invalid request", http.StatusBadRequest)

		return
	}

	h.Logger.Info("Received a heartbeat.", zap.String("service", heartBeatRequest.ServiceName))

	h.Service.RegisterHeartBeat(heartBeatRequest.IpAddress, heartBeatRequest.Port, heartBeatRequest.ServiceName)

	// TODO: if an agent is heartbeating with an old rule version id, notify it to fetch the latest rules

	w.WriteHeader(http.StatusOK)
}
