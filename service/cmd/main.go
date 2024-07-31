package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/tratteria/tconfigd/config"
	"github.com/tratteria/tconfigd/logging"
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

	if len(os.Args) < 2 {
		log.Fatalf("No configuration file provided. Please specify the configuration path as an argument when running the service.\nUsage: %s <config-path>", os.Args[0])
	}

	configPath := os.Args[1]

	config, err := config.GetConfig(configPath)
	if err != nil {
		log.Fatal("Error reading configuration.", zap.Error(err))
	}

	if err := logging.InitLogger(); err != nil {
		panic(err)
	}
	defer logging.Sync()

	logger := logging.GetLogger("main")

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

	tratteriaController := tratteriacontroller.NewTratteriaController(logging.GetLogger("controller"))

	if err := tratteriaController.Run(); err != nil {
		logger.Fatal("Failed to start TraT Controller server.", zap.Error(err))
	}

	webSocketServer := websocketserver.NewWebSocketServer(tratteriaController.Controller, x509Source, spiffeid.ID(config.TratteriaSpiffeId), logging.GetLogger("websocketserver"))

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
			Logger:                 logging.GetLogger("webhook"),
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
