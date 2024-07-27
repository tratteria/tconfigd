package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/tratteria/tconfigd/config"
	"github.com/tratteria/tconfigd/spiffe"
	"github.com/tratteria/tconfigd/tratteriacontroller"
	"github.com/tratteria/tconfigd/webhook"
	"github.com/tratteria/tconfigd/websocketserver"
	"go.uber.org/zap"
)

const X509_SOURCE_TIMEOUT = 15 * time.Second

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

	config, err := config.GetConfig(configPath)
	if err != nil {
		logger.Fatal("Error reading configuration.", zap.Error(err))
	}

	x509SrcCtx, cancel := context.WithTimeout(context.Background(), X509_SOURCE_TIMEOUT)
	defer cancel()

	x509Source, err := workloadapi.NewX509Source(x509SrcCtx)
	if err != nil {
		logger.Fatal("Failed to create X.509 source", zap.Error(err))
	}

	defer x509Source.Close()

	tconfigdSpiffeId, err := spiffe.FetchSpiffeIdFromX509(x509Source)
	if err != nil {
		logger.Fatal("Error getting tconfigd spiffe id.", zap.Error(err))
	}

	tratteriaController := tratteriacontroller.NewTratteriaController()

	if err := tratteriaController.Run(); err != nil {
		logger.Fatal("Failed to start TraT Controller server.", zap.Error(err))
	}

	webSocketServer := websocketserver.NewWebSocketServer(tratteriaController.Controller, x509Source, spiffeid.ID(config.TratteriaSpiffeId), logger)

	tratteriaController.SetClientsRetriever(webSocketServer)

	go func() {
		if err := webSocketServer.Run(); err != nil {
			logger.Fatal("Failed to start websocket server.", zap.Error(err))
		}
	}()

	go func() {
		webhook := &webhook.Webhook{
			EnableTratInterception: bool(config.EnableTratInterception),
			AgentHttpApiPort:       int(config.AgentHttpApiPort),
			AgentInterceptorPort:   int(config.AgentInterceptorPort),
			SpireAgentHostDir:      config.SpireAgentHostDir,
			TconfigdSpiffeId:       tconfigdSpiffeId,
			Logger:                 logger,
		}

		if err := webhook.Run(); err != nil {
			logger.Fatal("Failed to start Webhook server.", zap.Error(err))
		}
	}()

	<-ctx.Done()

	logger.Info("Shutting down tconfigd...")
}

func setupSignalHandler(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()
}
