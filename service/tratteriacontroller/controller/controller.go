package controller

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/tratteria/tconfigd/common"
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

type ServiceHash struct {
	ruleVersionNumber int64
	mu                sync.RWMutex
	hash              string
}

type NamespaceRulesHashes struct {
	mu                  sync.RWMutex
	servicesRulesHashes map[string]*ServiceHash
}

func NewNamespaceRulesHashes() *NamespaceRulesHashes {
	return &NamespaceRulesHashes{
		servicesRulesHashes: make(map[string]*ServiceHash),
	}
}

type AllRulesHashes struct {
	mu                    sync.RWMutex
	namespacesRulesHashes map[string]*NamespaceRulesHashes
}

func NewAllRulesHashes() *AllRulesHashes {
	return &AllRulesHashes{
		namespacesRulesHashes: make(map[string]*NamespaceRulesHashes),
	}
}

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
	ruleVersionNumber      int64
	allRulesHashes         *AllRulesHashes
	logger                 *zap.Logger
}

func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	tratteriaclientset clientset.Interface,
	traTInformer informers.TraTInformer,
	tratteriaConfigInformer informers.TratteriaConfigInformer,
	ruleDispatcher *ruledispatcher.RuleDispatcher,
	logger *zap.Logger) *Controller {

	utilruntime.Must(tratteriascheme.AddToScheme(scheme.Scheme))

	logger.Info("Creating event broadcaster")

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
		ruleVersionNumber:      0,
		allRulesHashes:         NewAllRulesHashes(),
		logger:                 logger,
	}

	logger.Info("Setting up event handlers")

	traTInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.AddFunc,
		UpdateFunc: controller.UpdateTraT,
	})

	tratteriaConfigInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.AddFunc,
		UpdateFunc: controller.UpdateTratteriaConfig,
	})

	return controller
}

func (c *Controller) AddFunc(obj interface{}) {
	// This version number and higher represent the rule states that are guaranteed to incorporate this particular change
	versionNumber := atomic.AddInt64(&c.ruleVersionNumber, 1)

	// Enqueuing the version number to track which clients have already incorporated this change and which clients still
	// need this change propagated
	c.enqueueObject(obj, versionNumber)
}

func (c *Controller) UpdateTraT(oldObj, newObj interface{}) {
	oldTraT, ok := oldObj.(*tratteria1alpha1.TraT)
	if !ok {
		c.logger.Error("Received unexpected object type", zap.String("expected", "TraT"), zap.Any("got", oldObj))

		return
	}

	newTraT, ok := newObj.(*tratteria1alpha1.TraT)
	if !ok {
		c.logger.Error("Received unexpected object type", zap.String("expected", "TraT"), zap.Any("got", oldObj))

		return
	}

	if !reflect.DeepEqual(oldTraT.Spec, newTraT.Spec) {
		versionNumber := atomic.AddInt64(&c.ruleVersionNumber, 1)
		c.enqueueObject(newObj, versionNumber)
	} else {
		c.logger.Debug("TraT update ignored, no spec change.", zap.String("namespace", newTraT.Namespace), zap.String("trat", newTraT.Name))
	}
}

func (c *Controller) UpdateTratteriaConfig(oldObj, newObj interface{}) {
	oldTratteriaConfig, ok := oldObj.(*tratteria1alpha1.TratteriaConfig)
	if !ok {
		c.logger.Error("Received unexpected object type", zap.String("expected", "TratteriaConfig"), zap.Any("got", oldObj))

		return
	}

	newTratteriaConfig, ok := newObj.(*tratteria1alpha1.TratteriaConfig)
	if !ok {
		c.logger.Error("Received unexpected object type", zap.String("expected", "TratteriaConfig"), zap.Any("got", oldObj))

		return
	}

	if !reflect.DeepEqual(oldTratteriaConfig.Spec, newTratteriaConfig.Spec) {
		versionNumber := atomic.AddInt64(&c.ruleVersionNumber, 1)
		c.enqueueObject(newObj, versionNumber)
	} else {
		c.logger.Debug("TratteriaConfig update ignored, no spec change.", zap.String("namespace", newTratteriaConfig.Namespace), zap.String("trat", newTratteriaConfig.Name))
	}
}

func (c *Controller) enqueueObject(obj interface{}, versionNumber int64) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)

		return
	}

	var resourceType string

	switch obj.(type) {
	case *tratteria1alpha1.TraT:
		resourceType = TraTKind
	case *tratteria1alpha1.TratteriaConfig:
		resourceType = TratteriaConfigKind
	default:
		utilruntime.HandleError(fmt.Errorf("unknown type cannot be enqueued: %T", obj))

		return
	}

	c.workqueue.Add(fmt.Sprintf("%s/%s/%d", resourceType, key, versionNumber))

	c.logger.Info("Enqueued resource with new rule version.",
		zap.String("resource-type", resourceType),
		zap.String("key", key),
		zap.Int64("version-number", versionNumber))
}

func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	c.logger.Info("Starting TraT controller")

	c.logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.traTsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	c.logger.Info("Starting workers", zap.Int("count", workers))

	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	c.logger.Info("Started workers")
	<-ctx.Done()
	c.logger.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := c.workqueue.Get()

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

		c.logger.Info("Successfully applied.", zap.String("resourceName", obj))

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
	if len(parts) < 4 {
		utilruntime.HandleError(fmt.Errorf("unexpected key format: %s", key))

		return nil
	}

	resourceType := parts[0]
	objectKey := parts[1] + "/" + parts[2]
	versionNumber, err := strconv.Atoi(parts[3])

	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid version number in key: %s", key))

		return nil
	}

	switch resourceType {
	case TraTKind:
		return c.handleTraT(ctx, objectKey, int64(versionNumber))
	case TratteriaConfigKind:
		return c.handleTratteriaConfig(ctx, objectKey, int64(versionNumber))
	default:
		utilruntime.HandleError(fmt.Errorf("unhandled resource type: %s", resourceType))

		return nil
	}
}

func (c *Controller) GetActiveVerificationRules(serviceName string, namespace string) (*tratteria1alpha1.VerificationRules, int64, error) {
	// The returned rule is guaranteed to incorporate changes at least up to and including this rule version number
	activeRuleVersionNumber := atomic.LoadInt64(&c.ruleVersionNumber)

	tratteriaConfigVerificationRule, err := c.GetActiveTratteriaConfigVerificationRule(namespace)
	if err != nil {
		return nil, 0, err
	}

	traTVerificationRules, err := c.GetActiveTraTsVerificationRules(serviceName, namespace)
	if err != nil {
		return nil, 0, err
	}

	return &tratteria1alpha1.VerificationRules{
			TratteriaConfigVerificationRule: tratteriaConfigVerificationRule,
			TraTsVerificationRules:          traTVerificationRules,
		},
		activeRuleVersionNumber,
		nil
}

func (c *Controller) GetActiveVerificationRulesHash(serviceName string, namespace string) (string, int64, error) {
	err := c.RecomputeRulesHashesIfNotLatest(serviceName, namespace)
	if err != nil {
		return "", 0, err
	}

	c.allRulesHashes.namespacesRulesHashes[namespace].servicesRulesHashes[serviceName].mu.RLock()
	defer c.allRulesHashes.namespacesRulesHashes[namespace].servicesRulesHashes[serviceName].mu.RUnlock()

	return c.allRulesHashes.namespacesRulesHashes[namespace].servicesRulesHashes[serviceName].hash,
		c.allRulesHashes.namespacesRulesHashes[namespace].servicesRulesHashes[serviceName].ruleVersionNumber,
		nil
}

func (c *Controller) GetActiveGenerationRules(namespace string) (*tratteria1alpha1.GenerationRules, int64, error) {
	// The returned rule is guaranteed to incorporate changes at least up to and including this rule version number
	activeRuleVersionNumber := atomic.LoadInt64(&c.ruleVersionNumber)

	generationTratteriaConfigRule, err := c.GetActiveTratteriaConfigGenerationRule(namespace)
	if err != nil {
		return nil, 0, err
	}

	generationTraTRules, err := c.GetActiveTraTsGenerationRules(namespace)
	if err != nil {
		return nil, 0, err
	}

	return &tratteria1alpha1.GenerationRules{
			TratteriaConfigGenerationRule: generationTratteriaConfigRule,
			TraTsGenerationRules:          generationTraTRules,
		},
		activeRuleVersionNumber,
		nil
}

func (c *Controller) GetActiveGenerationRulesHash(namespace string) (string, int64, error) {
	err := c.RecomputeRulesHashesIfNotLatest(common.TRATTERIA_SERVICE_NAME, namespace)
	if err != nil {
		return "", 0, err
	}

	c.allRulesHashes.namespacesRulesHashes[namespace].servicesRulesHashes[common.TRATTERIA_SERVICE_NAME].mu.RLock()
	defer c.allRulesHashes.namespacesRulesHashes[namespace].servicesRulesHashes[common.TRATTERIA_SERVICE_NAME].mu.RUnlock()

	return c.allRulesHashes.namespacesRulesHashes[namespace].servicesRulesHashes[common.TRATTERIA_SERVICE_NAME].hash,
		c.allRulesHashes.namespacesRulesHashes[namespace].servicesRulesHashes[common.TRATTERIA_SERVICE_NAME].ruleVersionNumber,
		nil
}

func (c *Controller) RecomputeRulesHashesIfNotLatest(serviceName string, namespace string) error {
	if c.allRulesHashes.namespacesRulesHashes[namespace] == nil {
		c.allRulesHashes.mu.Lock()
		if c.allRulesHashes.namespacesRulesHashes[namespace] == nil {
			c.allRulesHashes.namespacesRulesHashes[namespace] = NewNamespaceRulesHashes()
		}
		c.allRulesHashes.mu.Unlock()
	}

	namespaceHashes := c.allRulesHashes.namespacesRulesHashes[namespace]

	if namespaceHashes.servicesRulesHashes[serviceName] == nil {
		namespaceHashes.mu.Lock()
		if namespaceHashes.servicesRulesHashes[serviceName] == nil {
			namespaceHashes.servicesRulesHashes[serviceName] = &ServiceHash{}
		}
		namespaceHashes.mu.Unlock()
	}

	serviceHash := namespaceHashes.servicesRulesHashes[serviceName]

	serviceHash.mu.RLock()
	serviceHashVersionNumber := serviceHash.ruleVersionNumber
	serviceHash.mu.RUnlock()

	if atomic.LoadInt64(&c.ruleVersionNumber) == serviceHashVersionNumber {
		return nil
	}

	serviceHash.mu.Lock()
	defer serviceHash.mu.Unlock()

	if atomic.LoadInt64(&c.ruleVersionNumber) == serviceHash.ruleVersionNumber {
		return nil
	}

	// Rule version number that is tagged to the computed hash
	// The computed hash is guaranteed to incorporate changes at least up to and including this rule version number
	ruleVersionNumber := atomic.LoadInt64(&c.ruleVersionNumber)

	if serviceName == common.TRATTERIA_SERVICE_NAME {
		activeGenerationRules, _, err := c.GetActiveGenerationRules(namespace)
		if err != nil {
			return err
		}

		generationRuleHash, err := activeGenerationRules.ComputeStableHash()
		if err != nil {
			return err
		}

		serviceHash.hash = generationRuleHash
	} else {
		activeVerificationRules, _, err := c.GetActiveVerificationRules(serviceName, namespace)
		if err != nil {
			return err
		}

		verificationRuleHash, err := activeVerificationRules.ComputeStableHash()
		if err != nil {
			return err
		}

		serviceHash.hash = verificationRuleHash
	}

	serviceHash.ruleVersionNumber = ruleVersionNumber

	return nil
}
