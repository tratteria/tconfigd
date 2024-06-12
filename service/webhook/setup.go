package webhook

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/tratteria/tconfigd/webhook/handler"
	"github.com/tratteria/tconfigd/webhook/pkg/tlscreds"
)

type Webhook struct {
	enableTratInterception bool
	logger                 *zap.Logger
}

func NewWebhook(enableTratInterception bool, logger *zap.Logger) *Webhook {
	return &Webhook{
		enableTratInterception: enableTratInterception,
		logger:                 logger,
	}
}

func (webhook *Webhook) Run() error {
	handler := handler.NewHandlers(webhook.enableTratInterception, webhook.logger)
	router := mux.NewRouter()

	initializeRoutes(router, handler)

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:443",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if err := tlscreds.SetupTLSCertAndKeyFromSPIRE(); err != nil {
		webhook.logger.Error("Error setting up TLS creds", zap.Error(err))

		return fmt.Errorf("error setting up TLS creds: %w", err)
	}

	webhook.logger.Info("Starting webhook server with TLS on port 443")

	if err := srv.ListenAndServeTLS(tlscreds.CertPath, tlscreds.KeyPath); err != nil && err != http.ErrServerClosed {
		webhook.logger.Error("Failed to start the webhook server", zap.Error(err))

		return fmt.Errorf("failed to start the webhook server: %w", err)
	}

	return nil
}

func initializeRoutes(router *mux.Router, handler *handler.Handlers) {
	router.HandleFunc("/inject-tratteria-agents", handler.InjectTratteriaAgent).Methods("POST")
}
