package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tratteria/tconfigd/api/pkg/apierrors"
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

func (h *Handlers) GetVerificationRulesHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	serviceName := queryParams.Get("service")

	if serviceName == "" {
		http.Error(w, "Service parameter is required", http.StatusBadRequest)

		return
	}

	verificationRules, err := h.Service.GetVerificationRule(serviceName)
	if err != nil {
		if errors.Is(err, apierrors.ErrVerificationRuleNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server error", http.StatusInternalServerError)
		}

		return
	}

	response, err := json.Marshal(verificationRules)
	if err != nil {
		http.Error(w, "Failed to encode verification rules", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *Handlers) GetGenerationRulesHandler(w http.ResponseWriter, r *http.Request) {
	generationRules := h.Service.GetGenerationRule()

	response, err := json.Marshal(generationRules)
	if err != nil {
		http.Error(w, "Failed to encode generation rules", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
