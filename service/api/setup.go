package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/tratteria/tconfigd/api/handler"
	"github.com/tratteria/tconfigd/api/pkg/rules"
	"github.com/tratteria/tconfigd/api/pkg/service"
)

type API struct {
	Router  *mux.Router
	Rules   *rules.Rules
	Handler *handler.Handlers
	Logger  *zap.Logger
}

func Run() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("cannot initialize Zap logger: %w", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Error syncing logger: %v", err)
		}
	}()

	rules := rules.NewRules()

	err = rules.Load()
	if err != nil {
		return fmt.Errorf("error loading rules: %w", err)
	}

	service := service.NewService(rules, logger)
	handler := handler.NewHandlers(service, logger)

	api := &API{
		Router:  mux.NewRouter(),
		Rules:   rules,
		Handler: handler,
		Logger:  logger,
	}

	api.initializeRulesRoutes()

	srv := &http.Server{
		Handler:      api.Router,
		Addr:         "0.0.0.0:9060",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("Starting api server on port 9060.")

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Failed to start the api server", zap.Error(err))

		return err
	}

	return nil
}

func (a *API) initializeRulesRoutes() {
	a.Router.HandleFunc("/verification-rules", a.Handler.GetVerificationRulesHandler).Methods("GET")
	a.Router.HandleFunc("/generation-rules", a.Handler.GetGenerationRulesHandler).Methods("GET")
}
