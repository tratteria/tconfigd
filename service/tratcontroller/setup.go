package tratcontroller

import (
	"flag"
	"fmt"
	"time"

	"github.com/tratteria/tconfigd/tratcontroller/pkg/signals"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	clientset "github.com/tratteria/tconfigd/tratcontroller/pkg/generated/clientset/versioned"
	informers "github.com/tratteria/tconfigd/tratcontroller/pkg/generated/informers/externalversions"
)

type TraTController struct{}

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

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes clientset: %w", err)
	}

	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	controller := NewController(ctx, kubeClient, exampleClient, exampleInformerFactory.Tratteria().V1alpha1().TraTs())

	exampleInformerFactory.Start(ctx.Done())

	if err = controller.Run(ctx, 2); err != nil {
		return fmt.Errorf("error running controller: %w", err)
	}

	return nil
}
