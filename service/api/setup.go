package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"go.uber.org/zap"

	"github.com/tratteria/tconfigd/api/handler"
	"github.com/tratteria/tconfigd/api/pkg/service"
	"github.com/tratteria/tconfigd/dataplaneregistry"
)

const API_PORT = 8443

type API struct {
	DataPlaneRegistryManager dataplaneregistry.Manager
	X509Source               *workloadapi.X509Source
	TratteriaSpiffeId        spiffeid.ID
	Logger                   *zap.Logger
}

func (api *API) Run() error {
	service := service.NewService(api.DataPlaneRegistryManager, api.X509Source, api.TratteriaSpiffeId, api.Logger)
	handler := handler.NewHandlers(service, api.Logger)
	router := mux.NewRouter()

	initializeRulesRoutes(router, handler)

	serverTLSConfig := tlsconfig.MTLSServerConfig(api.X509Source, api.X509Source, tlsconfig.AuthorizeAny())

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("0.0.0.0:%d", API_PORT),
		TLSConfig:    serverTLSConfig,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	api.Logger.Info("Starting api server...", zap.Int("port", API_PORT))

	if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
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
