package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/tokenetes/tconfigd/config"
	"github.com/tokenetes/tconfigd/logging"
	"github.com/tokenetes/tconfigd/spiffe"
	"github.com/tokenetes/tconfigd/tokenetescontroller"
	"github.com/tokenetes/tconfigd/webhook"
	"github.com/tokenetes/tconfigd/websocketserver"
	"go.uber.org/zap"
)

const X509_SOURCE_TIMEOUT = 15 * time.Second

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	if err := logging.InitLogger(); err != nil {
		panic(err)
	}
	defer logging.Sync()

	mainLogger := logging.GetLogger("main")

	if len(os.Args) < 2 {
		mainLogger.Fatal(fmt.Sprintf("No configuration file provided. Please specify the configuration path as an argument when running the service.\nUsage: %s <config-path>", os.Args[0]))
	}

	configPath := os.Args[1]

	config, err := config.GetConfig(configPath)
	if err != nil {
		mainLogger.Fatal("Error reading configuration.", zap.Error(err))
	}

	x509SrcCtx, cancel := context.WithTimeout(context.Background(), X509_SOURCE_TIMEOUT)
	defer cancel()

	x509Source, err := workloadapi.NewX509Source(x509SrcCtx)
	if err != nil {
		mainLogger.Fatal("Failed to create X.509 source", zap.Error(err))
	}

	defer x509Source.Close()

	tconfigdSpiffeId, err := spiffe.FetchSpiffeIdFromX509(x509Source)
	if err != nil {
		mainLogger.Fatal("Error getting tconfigd spiffe id.", zap.Error(err))
	}

	tokenetesController := tokenetescontroller.NewTokenetesController(logging.GetLogger("controller"))

	if err := tokenetesController.Run(); err != nil {
		mainLogger.Fatal("Failed to start TraT Controller server.", zap.Error(err))
	}

	webSocketServer := websocketserver.NewWebSocketServer(tokenetesController.Controller, x509Source, spiffeid.ID(config.TokenetesSpiffeId), logging.GetLogger("websocket-server"))

	tokenetesController.SetClientsRetriever(webSocketServer)

	go func() {
		if err := webSocketServer.Run(); err != nil {
			mainLogger.Fatal("Failed to start websocket server.", zap.Error(err))
		}
	}()

	go func() {
		webhook := &webhook.Webhook{
			EnableTratInterception: bool(config.EnableTratInterception),
			AgentApiPort:           int(config.AgentApiPort),
			AgentInterceptorPort:   int(config.AgentInterceptorPort),
			SpireAgentHostDir:      config.SpireAgentHostDir,
			TconfigdSpiffeId:       tconfigdSpiffeId,
			Logger:                 logging.GetLogger("webhook"),
		}

		if err := webhook.Run(); err != nil {
			mainLogger.Fatal("Failed to start Webhook server.", zap.Error(err))
		}
	}()

	<-ctx.Done()

	mainLogger.Info("Shutting down tconfigd...")
}

func setupSignalHandler(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()
}
