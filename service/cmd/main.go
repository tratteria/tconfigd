package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tratteria/tconfigd/api"
	"github.com/tratteria/tconfigd/webhook"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	go func() {
		log.Println("Starting API server...")

		if err := api.Run(); err != nil {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	go func() {
		log.Println("Starting Webhook server...")

		if err := webhook.Run(); err != nil {
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
