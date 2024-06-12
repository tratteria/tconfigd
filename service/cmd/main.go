package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tratteria/tconfigd/api"
	"github.com/tratteria/tconfigd/config"
	"github.com/tratteria/tconfigd/webhook"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

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
		logger.Fatal(fmt.Sprintf("No configuration file provided. Please specify the configuration path as an argument when running the service.\nUsage: %s <config-path>", os.Args[0]))
	}

	configPath := os.Args[1]

	appConfig, err := config.GetAppConfig(configPath)
	if err != nil {
		logger.Fatal("Error reading configuration.", zap.Error(err))
	}

	go func() {
		logger.Info("Starting API server...")

		apiServer := api.NewAPI(logger)

		if err := apiServer.Run(); err != nil {
			logger.Fatal("Failed to start API server.", zap.Error(err))
		}
	}()

	go func() {
		logger.Info("Starting Webhook server...")

		webhook := webhook.NewWebhook(bool(appConfig.EnableTratInterception), logger)

		if err := webhook.Run(); err != nil {
			logger.Fatal("Failed to start webhook server.", zap.Error(err))
		}
	}()

	<-ctx.Done()

	logger.Info("Shutting down servers and controllers...")
}

func setupSignalHandler(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()
}
