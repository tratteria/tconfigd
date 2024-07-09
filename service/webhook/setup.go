package webhook

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"go.uber.org/zap"

	"github.com/tratteria/tconfigd/webhook/handler"
	"github.com/tratteria/tconfigd/webhook/pkg/tlscreds"
)

type Webhook struct {
	EnableTratInterception bool
	AgentHttpsApiPort      int
	AgentHttpApiPort       int
	AgentInterceptorPort   int
	SpiffeEndpointSocket   string
	TconfigdSpiffeId       spiffeid.ID
	Logger                 *zap.Logger
}

func (wh *Webhook) Run() error {
	handler := handler.NewHandlers(wh.EnableTratInterception, wh.AgentHttpsApiPort, wh.AgentHttpApiPort, wh.AgentInterceptorPort, wh.SpiffeEndpointSocket, wh.TconfigdSpiffeId, wh.Logger)
	router := mux.NewRouter()

	initializeRoutes(router, handler)

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:443",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if err := tlscreds.SetupTLSCertAndKeyFromSPIRE(wh.SpiffeEndpointSocket); err != nil {
		wh.Logger.Error("Error setting up TLS creds", zap.Error(err))

		return fmt.Errorf("error setting up TLS creds: %w", err)
	}

	wh.Logger.Info("Starting webhook server...", zap.Int("port", 443))

	if err := srv.ListenAndServeTLS(tlscreds.CertPath, tlscreds.KeyPath); err != nil && err != http.ErrServerClosed {
		wh.Logger.Error("Failed to start the webhook server", zap.Error(err))

		return fmt.Errorf("failed to start the webhook server: %w", err)
	}

	return nil
}

func initializeRoutes(router *mux.Router, handler *handler.Handlers) {
	router.HandleFunc("/inject-tratteria-agents", handler.InjectTratteriaAgent).Methods("POST")
}
