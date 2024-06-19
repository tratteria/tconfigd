package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/tratteria/tconfigd/agentsmanager"
	"github.com/tratteria/tconfigd/api/handler"
	"github.com/tratteria/tconfigd/api/pkg/rules"
	"github.com/tratteria/tconfigd/api/pkg/service"
)

type API struct {
	AgentsManager *agentsmanager.AgentsManager
	Logger        *zap.Logger
}

func (api *API) Run() error {
	rules := rules.NewRules()

	err := rules.Load()
	if err != nil {
		return fmt.Errorf("error loading rules: %w", err)
	}

	service := service.NewService(rules, api.AgentsManager, api.Logger)
	handler := handler.NewHandlers(service, api.Logger)
	router := mux.NewRouter()

	initializeRulesRoutes(router, handler)

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:9060",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	api.Logger.Info("Starting api server on port 9060.")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		api.Logger.Error("Failed to start the api server", zap.Error(err))

		return fmt.Errorf("failed to start the api server :%w", err)
	}

	return nil
}

func initializeRulesRoutes(router *mux.Router, handler *handler.Handlers) {
	router.HandleFunc("/agent-register", handler.RegistrationHandler).Methods("POST")
	router.HandleFunc("/agent-heartbeat", handler.HeartBeatHandler).Methods("POST")
	router.HandleFunc("/verification-rules", handler.GetVerificationRulesHandler).Methods("GET")
	router.HandleFunc("/generation-rules", handler.GetGenerationRulesHandler).Methods("GET")
}
