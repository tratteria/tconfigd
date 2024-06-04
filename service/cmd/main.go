package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/tratteria/tconfigd/handler"
	"github.com/tratteria/tconfigd/pkg/rules"
	"github.com/tratteria/tconfigd/pkg/service"
)

type App struct {
	Router *mux.Router
	Rules  *rules.Rules
	Logger *zap.Logger
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Cannot initialize Zap logger: %v.", err)
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Error syncing logger: %v", err)
		}
	}()

	if len(os.Args) < 2 {
		logger.Error("Rules directory not provided. Please specify the rules directory as an argument when running the service.",
			zap.String("usage", "tconfigd <rules-directory>"))
		os.Exit(1)
	}

	rulesDir := os.Args[1]
	rules := rules.NewRules(rulesDir)

	err = rules.Load()
	if err != nil {
		logger.Fatal("Error loading rules:", zap.Error(err))
	}

	app := &App{
		Router: mux.NewRouter(),
		Rules:  rules,
		Logger: logger,
	}

	appService := service.NewService(app.Rules, app.Logger)
	appHandler := handler.NewHandlers(appService, app.Logger)

	app.initializeRoutes(appHandler)

	srv := &http.Server{
		Handler:      app.Router,
		Addr:         "0.0.0.0:9060",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Info("Starting server on 9060.")
	log.Fatal(srv.ListenAndServe())
}

func (a *App) initializeRoutes(handlers *handler.Handlers) {
	a.Router.HandleFunc("/verification-rules", handlers.GetVerificationRulesHandler).Methods("GET")
	a.Router.HandleFunc("/generation-rules", handlers.GetGenerationRulesHandler).Methods("GET")
}
