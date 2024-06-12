package webhook

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/tratteria/tconfigd/webhook/handler"
	"github.com/tratteria/tconfigd/webhook/pkg/tlscreds"
	"github.com/tratteria/tconfigd/webhook/webhookconfig"
)

type Webhook struct {
	Router              *mux.Router
	SpireWrokloadClient *workloadapi.Client
	Handler             *handler.Handlers
	Logger              *zap.Logger
}

func Run(config *webhookconfig.WebhookConfig) error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("cannot initialize Zap logger: %w", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Error syncing logger: %v", err)
		}
	}()

	handler := handler.NewHandlers(config, logger)

	webhook := &Webhook{
		Router:  mux.NewRouter(),
		Handler: handler,
		Logger:  logger,
	}

	webhook.initializeRoutes()

	srv := &http.Server{
		Handler:      webhook.Router,
		Addr:         "0.0.0.0:443",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if err := tlscreds.SetupTLSCertAndKeyFromSPIRE(); err != nil {
		logger.Error("Error setting up TLS creds", zap.Error(err))

		return err
	}

	logger.Info("Starting webhook server with TLS on port 443")

	if err := srv.ListenAndServeTLS(tlscreds.CertPath, tlscreds.KeyPath); err != nil && err != http.ErrServerClosed {
		logger.Error("Failed to start the webhook server", zap.Error(err))

		return err
	}

	return nil
}

func (w *Webhook) initializeRoutes() {
	w.Router.HandleFunc("/inject-tratteria-agents", w.Handler.InjectTratteriaAgent).Methods("POST")
}
