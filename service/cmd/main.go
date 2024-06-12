package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tratteria/tconfigd/api"
	"github.com/tratteria/tconfigd/config"
	"github.com/tratteria/tconfigd/webhook"
	"github.com/tratteria/tconfigd/webhook/webhookconfig"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	if len(os.Args) < 2 {
		log.Fatalf("No configuration file provided. Please specify the configuration path as an argument when running the service.\nUsage: %s <config-path>", os.Args[0])
	}

	configPath := os.Args[1]

	appConfig, err := config.GetAppConfig(configPath)
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	go func() {
		log.Println("Starting API server...")

		if err := api.Run(); err != nil {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	go func() {
		log.Println("Starting Webhook server...")

		webhookConfig := webhookconfig.WebhookConfig{EnableTratInterception: bool(appConfig.EnableTratInterception)}

		if err := webhook.Run(&webhookConfig); err != nil {
			log.Fatalf("Webhook server failed: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down servers and controllers...")
}

func setupSignalHandler(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()
}
