package tratteriacontroller

import (
	"fmt"
	"time"

	"github.com/tratteria/tconfigd/servicemessagehandler"
	"go.uber.org/zap"

	"github.com/tratteria/tconfigd/tratteriacontroller/pkg/signals"
	"github.com/tratteria/tconfigd/websocketserver"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/tratteria/tconfigd/tratteriacontroller/controller"
	clientset "github.com/tratteria/tconfigd/tratteriacontroller/pkg/generated/clientset/versioned"
	informers "github.com/tratteria/tconfigd/tratteriacontroller/pkg/generated/informers/externalversions"
)

type TratteriaController struct {
	ServiceMessageHandler *servicemessagehandler.ServiceMessageHandler
	Controller            *controller.Controller
	Logger                *zap.Logger
}

func NewTratteriaController(logger *zap.Logger) *TratteriaController {
	return &TratteriaController{
		ServiceMessageHandler: servicemessagehandler.NewServiceMessageHandler(),
		Logger:                logger,
	}
}

func (tc *TratteriaController) SetClientsRetriever(clientsRetriever websocketserver.ClientsRetriever) {
	tc.ServiceMessageHandler.SetClientsRetriever(clientsRetriever)
}

func (tc *TratteriaController) Run() error {
	ctx := signals.SetupSignalHandler()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("error building in-cluster config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset: %w", err)
	}

	tratteriaClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset: %w", err)
	}

	tratteriaInformerFactory := informers.NewSharedInformerFactory(tratteriaClient, time.Second*30)
	tratInformer := tratteriaInformerFactory.Tratteria().V1alpha1().TraTs()
	tratteriaConfigInformer := tratteriaInformerFactory.Tratteria().V1alpha1().TratteriaConfigs()

	tc.Controller = controller.NewController(ctx, kubeClient, tratteriaClient, tratInformer, tratteriaConfigInformer, tc.ServiceMessageHandler, tc.Logger)

	go func() {
		tc.Logger.Info("Starting TraT Controller...")

		tratteriaInformerFactory.Start(ctx.Done())

		if err := tc.Controller.Run(ctx, 2); err != nil {
			tc.Logger.Error("Error running controller.", zap.Error(err))
		}
	}()

	return nil
}
