package tokenetescontroller

import (
	"fmt"
	"time"

	"github.com/tokenetes/tconfigd/servicemessagehandler"
	"go.uber.org/zap"

	"github.com/tokenetes/tconfigd/tokenetescontroller/pkg/signals"
	"github.com/tokenetes/tconfigd/websocketserver"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/tokenetes/tconfigd/tokenetescontroller/controller"
	clientset "github.com/tokenetes/tconfigd/tokenetescontroller/pkg/generated/clientset/versioned"
	informers "github.com/tokenetes/tconfigd/tokenetescontroller/pkg/generated/informers/externalversions"
)

type TokenetesController struct {
	ServiceMessageHandler *servicemessagehandler.ServiceMessageHandler
	Controller            *controller.Controller
	Logger                *zap.Logger
}

func NewTokenetesController(logger *zap.Logger) *TokenetesController {
	return &TokenetesController{
		ServiceMessageHandler: servicemessagehandler.NewServiceMessageHandler(),
		Logger:                logger,
	}
}

func (tc *TokenetesController) SetClientsRetriever(clientsRetriever websocketserver.ClientsRetriever) {
	tc.ServiceMessageHandler.SetClientsRetriever(clientsRetriever)
}

func (tc *TokenetesController) Run() error {
	ctx := signals.SetupSignalHandler()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("error building in-cluster config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset: %w", err)
	}

	tokenetesClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset: %w", err)
	}

	tokenetesInformerFactory := informers.NewSharedInformerFactory(tokenetesClient, time.Second*30)
	tratInformer := tokenetesInformerFactory.Tokenetes().V1alpha1().TraTs()
	tokenetesConfigInformer := tokenetesInformerFactory.Tokenetes().V1alpha1().TokenetesConfigs()
	tratExclusionInformer := tokenetesInformerFactory.Tokenetes().V1alpha1().TraTExclusions()

	tc.Controller = controller.NewController(ctx, kubeClient, tokenetesClient, tratInformer, tokenetesConfigInformer, tratExclusionInformer, tc.ServiceMessageHandler, tc.Logger)

	go func() {
		tc.Logger.Info("Starting TraT Controller...")

		tokenetesInformerFactory.Start(ctx.Done())

		if err := tc.Controller.Run(ctx, 2); err != nil {
			tc.Logger.Error("Error running controller.", zap.Error(err))
		}
	}()

	return nil
}
