package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/tratteria/tconfigd/ruledispatcher"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	tratteria1alpha1 "github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"
	clientset "github.com/tratteria/tconfigd/tratteriacontroller/pkg/generated/clientset/versioned"
	tratteriascheme "github.com/tratteria/tconfigd/tratteriacontroller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/tratteria/tconfigd/tratteriacontroller/pkg/generated/informers/externalversions/tratteria/v1alpha1"
	listers "github.com/tratteria/tconfigd/tratteriacontroller/pkg/generated/listers/tratteria/v1alpha1"
)

type Status string

const (
	PendingStatus Status = "PENDING"
	DoneStatus    Status = "DONE"
)

type Stage string

const (
	VerificationApplicationStage Stage = "verification application stage"
	GenerationApplicationStage   Stage = "generation application stage"
)

const (
	ControllerAgentName = "trat-controller"
	TraTKind            = "TraT"
	TratteriaConfigKind = "TratteriaConfig"
)

type Controller struct {
	kubeclientset          kubernetes.Interface
	tratteriaclientset     clientset.Interface
	traTsLister            listers.TraTLister
	tratteriaConfigsLister listers.TratteriaConfigLister
	traTsSynced            cache.InformerSynced
	tratteriaConfigsSynced cache.InformerSynced
	workqueue              workqueue.TypedRateLimitingInterface[string]
	recorder               record.EventRecorder
	ruleDispatcher         *ruledispatcher.RuleDispatcher
}

func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	tratteriaclientset clientset.Interface,
	traTInformer informers.TraTInformer,
	tratteriaConfigInformer informers.TratteriaConfigInformer,
	ruleDispatcher *ruledispatcher.RuleDispatcher) *Controller {
	logger := klog.FromContext(ctx)

	utilruntime.Must(tratteriascheme.AddToScheme(scheme.Scheme))

	logger.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster(record.WithContext(ctx))

	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})

	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: ControllerAgentName})
	ratelimiter := workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[string](5*time.Millisecond, 1000*time.Second),
		&workqueue.TypedBucketRateLimiter[string]{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	controller := &Controller{
		kubeclientset:          kubeclientset,
		tratteriaclientset:     tratteriaclientset,
		traTsLister:            traTInformer.Lister(),
		tratteriaConfigsLister: tratteriaConfigInformer.Lister(),
		traTsSynced:            traTInformer.Informer().HasSynced,
		tratteriaConfigsSynced: tratteriaConfigInformer.Informer().HasSynced,
		workqueue:              workqueue.NewTypedRateLimitingQueue(ratelimiter),
		recorder:               recorder,
		ruleDispatcher:         ruleDispatcher,
	}

	logger.Info("Setting up event handlers")

	traTInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueObject,
	})

	tratteriaConfigInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueObject,
	})

	return controller
}

func (c *Controller) enqueueObject(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)

		return
	}

	switch obj.(type) {
	case *tratteria1alpha1.TraT:
		c.workqueue.Add(TraTKind + "/" + key)
	case *tratteria1alpha1.TratteriaConfig:
		c.workqueue.Add(TratteriaConfigKind + "/" + key)
	default:
		utilruntime.HandleError(fmt.Errorf("unknown type cannot be enqueued: %T", obj))
	}
}

func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	logger := klog.FromContext(ctx)

	logger.Info("Starting TraT controller")

	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.traTsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers", "count", workers)

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	logger.Info("Started workers")
	<-ctx.Done()
	logger.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := c.workqueue.Get()
	logger := klog.FromContext(ctx)

	if shutdown {
		return false
	}

	err := func() error {
		defer c.workqueue.Done(obj)

		if err := c.syncHandler(ctx, obj); err != nil {
			c.workqueue.AddRateLimited(obj)

			return fmt.Errorf("error applying '%s': %s, requeuing", obj, err.Error())
		}

		c.workqueue.Forget(obj)

		logger.Info("Successfully applied", "resourceName", obj)

		return nil
	}()

	if err != nil {
		utilruntime.HandleError(err)

		return true
	}

	return true
}

func (c *Controller) syncHandler(ctx context.Context, key string) error {
	parts := strings.Split(key, "/")
	if len(parts) < 3 {
		utilruntime.HandleError(fmt.Errorf("unexpected key format: %s", key))
		return nil
	}

	resourceType := parts[0]
	key = parts[1] + "/" + parts[2]

	switch resourceType {
	case TraTKind:
		return c.handleTraT(ctx, key)
	case TratteriaConfigKind:
		return c.handleTratteriaConfig(ctx, key)
	default:
		utilruntime.HandleError(fmt.Errorf("unhandled resource type: %s", resourceType))

		return nil
	}
}

func (c *Controller) GetActiveVerificationRules(serviceName string, namespace string) (*tratteria1alpha1.VerificationRules, error) {
	tratteriaConfigVerificationRule, err := c.GetActiveTratteriaConfigVerificationRule(namespace)
	if err != nil {
		return nil, err
	}

	traTVerificationRules, err := c.GetActiveTraTVerificationRules(serviceName, namespace)
	if err != nil {
		return nil, err
	}

	return &tratteria1alpha1.VerificationRules{
		TratteriaConfigVerificationRule: tratteriaConfigVerificationRule,
		TraTVerificationRules:           traTVerificationRules,
	}, nil
}

func (c *Controller) GetActiveGenerationRules(namespace string) (*tratteria1alpha1.GenerationRules, error) {
	generationTratteriaConfigRule, err := c.GetActiveGenerationTokenRule(namespace)
	if err != nil {
		return nil, err
	}

	generationTraTRules, err := c.GetActiveGenerationEndpointRules(namespace)
	if err != nil {
		return nil, err
	}

	return &tratteria1alpha1.GenerationRules{
		TratteriaConfigGenerationRule: generationTratteriaConfigRule,
		TraTGenerationRules:           generationTraTRules,
	}, nil
}
