package tratcontroller

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	samplev1alpha1 "github.com/tratteria/tconfigd/tratcontroller/pkg/apis/tratteria/v1alpha1"
	clientset "github.com/tratteria/tconfigd/tratcontroller/pkg/generated/clientset/versioned"
	samplescheme "github.com/tratteria/tconfigd/tratcontroller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/tratteria/tconfigd/tratcontroller/pkg/generated/informers/externalversions/tratteria/v1alpha1"
	listers "github.com/tratteria/tconfigd/tratcontroller/pkg/generated/listers/tratteria/v1alpha1"
)

const controllerAgentName = "trat-controller"

const (
	SuccessApplied         = "Applied"
	MessageResourceApplied = "TraT applied successfully"
)

type Controller struct {
	kubeclientset   kubernetes.Interface
	sampleclientset clientset.Interface
	tratsLister     listers.TraTLister
	tratsSynced     cache.InformerSynced
	workqueue       workqueue.TypedRateLimitingInterface[string]
	recorder        record.EventRecorder
}

func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	tratInformer informers.TraTInformer) *Controller {
	logger := klog.FromContext(ctx)

	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))

	logger.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster(record.WithContext(ctx))

	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})

	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	ratelimiter := workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[string](5*time.Millisecond, 1000*time.Second),
		&workqueue.TypedBucketRateLimiter[string]{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	controller := &Controller{
		kubeclientset:   kubeclientset,
		sampleclientset: sampleclientset,
		tratsLister:     tratInformer.Lister(),
		tratsSynced:     tratInformer.Informer().HasSynced,
		workqueue:       workqueue.NewTypedRateLimitingQueue(ratelimiter),
		recorder:        recorder,
	}

	logger.Info("Setting up event handlers")

	tratInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueTraT,
	})

	return controller
}

func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	logger := klog.FromContext(ctx)

	logger.Info("Starting TraT controller")

	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.tratsSynced); !ok {
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
	namespace, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))

		return nil
	}

	trat, err := c.tratsLister.TraTs(namespace).Get(name)

	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("trat '%s' in work queue no longer exists", key))

			return nil
		}

		return err
	}

	err = c.updateTraTStatus(trat)

	if err != nil {
		return err
	}

	c.recorder.Event(trat, corev1.EventTypeNormal, SuccessApplied, MessageResourceApplied)

	return nil
}

func (c *Controller) updateTraTStatus(trat *samplev1alpha1.TraT) error {
	tratCopy := trat.DeepCopy()
	tratCopy.Status.Applied = true
	_, err := c.sampleclientset.TratteriaV1alpha1().TraTs(trat.Namespace).UpdateStatus(context.TODO(), tratCopy, metav1.UpdateOptions{})

	return err
}

func (c *Controller) enqueueTraT(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)

		return
	}

	c.workqueue.Add(key)
}
