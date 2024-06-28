package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/tratteria/tconfigd/api/handler"
	"github.com/tratteria/tconfigd/api/pkg/service"
	"github.com/tratteria/tconfigd/dataplaneregistry"
)

type API struct {
	DataPlaneRegistryManager dataplaneregistry.Manager
	HttpClient               *http.Client
	Logger                   *zap.Logger
}

func (api *API) Run() error {
	service := service.NewService(api.DataPlaneRegistryManager, api.HttpClient, api.Logger)
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
	router.HandleFunc("/register", handler.RegistrationHandler).Methods("POST")
	router.HandleFunc("/heartbeat", handler.HeartBeatHandler).Methods("POST")
	router.HandleFunc("/.well-known/jwks.json", handler.GetJwksHandler).Methods("GET")
}
