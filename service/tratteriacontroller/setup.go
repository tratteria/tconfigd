package tratteriacontroller

import (
	"flag"
	"fmt"
	"time"

	"github.com/tratteria/tconfigd/configdispatcher"

	"github.com/tratteria/tconfigd/tratteriacontroller/pkg/signals"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/tratteria/tconfigd/tratteriacontroller/controller"
	clientset "github.com/tratteria/tconfigd/tratteriacontroller/pkg/generated/clientset/versioned"
	informers "github.com/tratteria/tconfigd/tratteriacontroller/pkg/generated/informers/externalversions"
)

type TratteriaController struct {
	ConfigDispatcher *configdispatcher.ConfigDispatcher
	Controller       *controller.Controller
}

func (tc *TratteriaController) Run() error {
	klog.InitFlags(nil)
	defer klog.Flush()

	flag.Parse()

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

	controller := controller.NewController(ctx, kubeClient, tratteriaClient, tratInformer, tratteriaConfigInformer, tc.ConfigDispatcher)
	tc.Controller = controller

	go func() {
		klog.Info("Starting TraT Controller...")

		tratteriaInformerFactory.Start(ctx.Done())

		if err := controller.Run(ctx, 2); err != nil {
			klog.Errorf("error running controller: %v", err)
		}
	}()

	return nil
}
