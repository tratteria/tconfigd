package controller

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/tratteria/tconfigd/common"
	"github.com/tratteria/tconfigd/servicemessagehandler"

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

const (
	ControllerAgentName = "trat-controller"
	TraTKind            = "TraT"
	TratteriaConfigKind = "TratteriaConfig"
)

type OperationType string

const (
	ADD    OperationType = "ADD"
	UPDATE OperationType = "UPDATE"
	DELETE OperationType = "DELETE"
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

type TraTOperation struct {
	Type          OperationType
	NewTraT       *tratteria1alpha1.TraT
	OldTraT       *tratteria1alpha1.TraT
	VersionNumber int64
}

type TratteriaConfigOperation struct {
	Type               OperationType
	NewTratteriaConfig *tratteria1alpha1.TratteriaConfig
	OldTratteriaConfig *tratteria1alpha1.TratteriaConfig
	VersionNumber      int64
}

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
	workqueue              workqueue.TypedRateLimitingInterface[any]
	recorder               record.EventRecorder
	serviceMessageHandler  *servicemessagehandler.ServiceMessageHandler
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
	serviceMessageHandler *servicemessagehandler.ServiceMessageHandler,
	logger *zap.Logger) *Controller {

	utilruntime.Must(tratteriascheme.AddToScheme(scheme.Scheme))

	logger.Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster(record.WithContext(ctx))

	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})

	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: ControllerAgentName})

	controller := &Controller{
		kubeclientset:          kubeclientset,
		tratteriaclientset:     tratteriaclientset,
		traTsLister:            traTInformer.Lister(),
		tratteriaConfigsLister: tratteriaConfigInformer.Lister(),
		traTsSynced:            traTInformer.Informer().HasSynced,
		tratteriaConfigsSynced: tratteriaConfigInformer.Informer().HasSynced,
		workqueue:              workqueue.NewTypedRateLimitingQueue[any](workqueue.DefaultTypedControllerRateLimiter[any]()),
		recorder:               recorder,
		serviceMessageHandler:  serviceMessageHandler,
		ruleVersionNumber:      0,
		allRulesHashes:         NewAllRulesHashes(),
		logger:                 logger,
	}

	logger.Info("Setting up event handlers")

	traTInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.AddFunc,
		UpdateFunc: controller.UpdateFunc,
		DeleteFunc: controller.DeleteFunc,
	})

	tratteriaConfigInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.AddFunc,
		UpdateFunc: controller.UpdateFunc,
	})

	return controller
}

func (c *Controller) AddFunc(obj interface{}) {
	switch v := obj.(type) {
	case *tratteria1alpha1.TraT:
		versionNumber := atomic.AddInt64(&c.ruleVersionNumber, 1)

		c.workqueue.Add(TraTOperation{Type: ADD, NewTraT: v, VersionNumber: versionNumber})

		c.logger.Info("Processing TraT addition operation.",
			zap.String("name", v.Name),
			zap.String("namespace", v.Namespace),
			zap.Int64("version-number", versionNumber))
	case *tratteria1alpha1.TratteriaConfig:
		versionNumber := atomic.AddInt64(&c.ruleVersionNumber, 1)

		c.workqueue.Add(TratteriaConfigOperation{Type: ADD, NewTratteriaConfig: v, VersionNumber: versionNumber})

		c.logger.Info("Processing TratteriaConfig addition operation.",
			zap.String("name", v.Name),
			zap.String("namespace", v.Namespace),
			zap.Int64("version-number", versionNumber))
	default:
		c.logger.Error("Unknown type incountered for addition operation", zap.Any("obj", obj))
	}
}

func (c *Controller) UpdateFunc(oldObj, newObj interface{}) {
	switch oldV := oldObj.(type) {
	case *tratteria1alpha1.TraT:
		newV, ok := newObj.(*tratteria1alpha1.TraT)
		if !ok {
			c.logger.Error("Received unexpected object type", zap.String("expected", "TraT"), zap.Any("got", oldObj))

			return
		}

		if reflect.DeepEqual(newV.Spec, oldV.Spec) {
			c.logger.Debug("TraT update ignored, no spec change.",
				zap.String("name", newV.Name),
				zap.String("namespace", newV.Namespace))

			return
		}

		versionNumber := atomic.AddInt64(&c.ruleVersionNumber, 1)

		c.workqueue.Add(TraTOperation{Type: UPDATE, NewTraT: newV, OldTraT: oldV, VersionNumber: versionNumber})

		c.logger.Info("Processing TraT update operation.",
			zap.String("name", oldV.Name),
			zap.String("namespace", oldV.Namespace),
			zap.Int64("version-number", versionNumber))
	case *tratteria1alpha1.TratteriaConfig:
		newV, ok := newObj.(*tratteria1alpha1.TratteriaConfig)
		if !ok {
			c.logger.Error("Received unexpected object type", zap.String("expected", "TratteriaConfig"), zap.Any("got", oldObj))

			return
		}

		if reflect.DeepEqual(newV.Spec, oldV.Spec) {
			c.logger.Debug("TratteriaConfig update ignored, no spec change.",
				zap.String("name", newV.Name),
				zap.String("namespace", newV.Namespace),
			)

			return
		}

		versionNumber := atomic.AddInt64(&c.ruleVersionNumber, 1)

		c.workqueue.Add(TratteriaConfigOperation{Type: UPDATE, NewTratteriaConfig: newV, OldTratteriaConfig: oldV, VersionNumber: versionNumber})

		c.logger.Info("Processing TratteriaConfig update operation.",
			zap.String("name", oldV.Name),
			zap.String("namespace", oldV.Namespace),
			zap.Int64("version-number", versionNumber))
	default:
		c.logger.Error("Unknown type incountered for updated operation", zap.Any("oldObj", oldObj), zap.Any("newObj", newObj))
	}
}

func (c *Controller) DeleteFunc(obj interface{}) {
	switch oldV := obj.(type) {
	case *tratteria1alpha1.TraT:
		versionNumber := atomic.AddInt64(&c.ruleVersionNumber, 1)

		c.workqueue.Add(TraTOperation{Type: DELETE, OldTraT: oldV, VersionNumber: versionNumber})
		c.logger.Info("Processing TraT deletion operation.",
			zap.String("name", oldV.Name),
			zap.String("namespace", oldV.Namespace),
			zap.Int64("version-number", versionNumber))
	case cache.DeletedFinalStateUnknown:
		switch t := oldV.Obj.(type) {
		case *tratteria1alpha1.TraT:
			versionNumber := atomic.AddInt64(&c.ruleVersionNumber, 1)

			c.workqueue.Add(TraTOperation{Type: DELETE, OldTraT: t, VersionNumber: versionNumber})
			c.logger.Info("Processing TraT deletion operation.",
				zap.String("name", t.Name),
				zap.String("namespace", t.Namespace),
				zap.Int64("version-number", versionNumber))
		default:
			c.logger.Error("Tombstone contained unknown or unexpected type", zap.Any("obj", t))
		}
	default:
		c.logger.Error("Unknown or unexpected type incountered for deletion operation", zap.Any("obj", obj))
	}
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

		return nil
	}()

	if err != nil {
		utilruntime.HandleError(err)

		return true
	}

	return true
}

func (c *Controller) syncHandler(ctx context.Context, obj any) error {
	switch op := obj.(type) {
	case TraTOperation:
		switch op.Type {
		case ADD:
			return c.handleTraTUpsert(ctx, op.NewTraT, op.VersionNumber)
		case UPDATE:
			return c.handleTraTUpdation(ctx, op.NewTraT, op.OldTraT, op.VersionNumber)
		case DELETE:
			return c.handleTraTDeletion(ctx, op.OldTraT, op.VersionNumber)
		default:
			return fmt.Errorf("unknown TraT operation type: %s", op.Type)
		}
	case TratteriaConfigOperation:
		switch op.Type {
		case ADD:
			return c.handleTratteriaConfigUpsert(ctx, op.NewTratteriaConfig, op.VersionNumber)
		case UPDATE:
			return c.handleTratteriaConfigUpsert(ctx, op.NewTratteriaConfig, op.VersionNumber)
		default:
			return fmt.Errorf("unknown TratteriaConfig operation type: %s", op.Type)
		}
	}

	return fmt.Errorf("unknown operation type: %T", obj)
}

func (c *Controller) GetActiveVerificationRules(serviceName string, namespace string) (*tratteria1alpha1.VerificationRules, int64, error) {
	// The returned rule is guaranteed to incorporate changes at least up to and including this rule version number
	activeRuleVersionNumber := atomic.LoadInt64(&c.ruleVersionNumber)

	tratteriaConfigVerificationRule, err := c.GetActiveTratteriaConfigVerificationRule(namespace)
	if err != nil {
		return nil, 0, err
	}

	traTsVerificationRules, err := c.GetActiveTraTsVerificationRules(serviceName, namespace)
	if err != nil {
		return nil, 0, err
	}

	return &tratteria1alpha1.VerificationRules{
			TratteriaConfigVerificationRule: tratteriaConfigVerificationRule,
			TraTsVerificationRules:          traTsVerificationRules,
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
