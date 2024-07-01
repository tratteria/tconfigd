package tratcontroller

import (
	"flag"
	"fmt"
	"time"

	"github.com/tratteria/tconfigd/configdispatcher"

	"github.com/tratteria/tconfigd/tratcontroller/pkg/signals"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	clientset "github.com/tratteria/tconfigd/tratcontroller/pkg/generated/clientset/versioned"
	informers "github.com/tratteria/tconfigd/tratcontroller/pkg/generated/informers/externalversions"
	"github.com/tratteria/tconfigd/tratcontroller/controller"
)

type TraTController struct {
	ConfigDispatcher *configdispatcher.ConfigDispatcher
}

func (tc *TraTController) Run() error {
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
	tratConfigInformer := tratteriaInformerFactory.Tratteria().V1alpha1().TraTConfigs()


	controller := controller.NewController(ctx, kubeClient, tratteriaClient, tratInformer, tratConfigInformer, tc.ConfigDispatcher)

	tratteriaInformerFactory.Start(ctx.Done())

	if err = controller.Run(ctx, 2); err != nil {
		return fmt.Errorf("error running controller: %w", err)
	}

	return nil
}
