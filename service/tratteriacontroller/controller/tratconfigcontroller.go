package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	tratteria1alpha1 "github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"
)

func (c *Controller) handleTratteriaConfig(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))

		return nil
	}

	tratteriaConfig, err := c.tratteriaConfigsLister.TratteriaConfigs(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("tratteria config '%s' in work queue no longer exists", key))

			return nil
		}

		return err
	}

	verificationTokenRule, err := tratteriaConfig.GetVerificationTokenRule()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving verification token rule from %s tratteria config: %w", name, err)

		c.recorder.Event(tratteriaConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTratteriaConfigStatus(ctx, tratteriaConfig, VerificationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratteria config: %w", name, updateErr)
		}

		return messagedErr
	}

	err = c.configDispatcher.DispatchVerificationTokenRule(ctx, namespace, verificationTokenRule)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s tratteria config verification token rule: %w", name, err)

		c.recorder.Event(tratteriaConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTratteriaConfigStatus(ctx, tratteriaConfig, VerificationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratteria config: %w", name, updateErr)
		}

		return messagedErr
	}

	c.recorder.Event(tratteriaConfig, corev1.EventTypeNormal, string(VerificationApplicationStage)+" successful", string(VerificationApplicationStage)+" completed successfully")

	generationTokenRule, err := tratteriaConfig.GetGenerationTokenRule()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving generation token rules from %s tratteria config: %w", name, err)

		c.recorder.Event(tratteriaConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTratteriaConfigStatus(ctx, tratteriaConfig, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratteria config: %w", name, updateErr)
		}

		return messagedErr
	}

	err = c.configDispatcher.DispatchGenerationTokentRule(ctx, namespace, generationTokenRule)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s tratteria config generation token rule: %w", name, err)

		c.recorder.Event(tratteriaConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTratteriaConfigStatus(ctx, tratteriaConfig, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratteria config: %w", name, updateErr)
		}

		return messagedErr
	}

	if updateErr := c.updateSuccessTratteriaConfigStatus(ctx, tratteriaConfig); updateErr != nil {
		return fmt.Errorf("failed to update success status for %s tratteria config: %w", name, updateErr)
	}

	c.recorder.Event(tratteriaConfig, corev1.EventTypeNormal, string(GenerationApplicationStage)+" successful", string(GenerationApplicationStage)+" completed successfully")

	return nil
}

func (c *Controller) updateErrorTratteriaConfigStatus(ctx context.Context, tratteriaConfig *tratteria1alpha1.TratteriaConfig, stage Stage, err error) error {
	tratteriaConfigCopy := tratteriaConfig.DeepCopy()

	tratteriaConfigCopy.Status.VerificationApplied = false
	tratteriaConfigCopy.Status.GenerationApplied = false
	tratteriaConfigCopy.Status.Status = string(PendingStatus)
	tratteriaConfigCopy.Status.Retries += 1

	if stage == GenerationApplicationStage {
		tratteriaConfigCopy.Status.VerificationApplied = true
	}

	tratteriaConfigCopy.Status.LastErrorMessage = err.Error()

	_, updateErr := c.tratteriaclientset.TratteriaV1alpha1().TratteriaConfigs(tratteriaConfig.Namespace).UpdateStatus(ctx, tratteriaConfigCopy, metav1.UpdateOptions{})

	return updateErr
}

func (c *Controller) updateSuccessTratteriaConfigStatus(ctx context.Context, tratteriaConfig *tratteria1alpha1.TratteriaConfig) error {
	tratteriaConfigCopy := tratteriaConfig.DeepCopy()

	tratteriaConfigCopy.Status.VerificationApplied = true
	tratteriaConfigCopy.Status.GenerationApplied = true
	tratteriaConfigCopy.Status.Status = string(DoneStatus)

	_, updateErr := c.tratteriaclientset.TratteriaV1alpha1().TratteriaConfigs(tratteriaConfig.Namespace).UpdateStatus(ctx, tratteriaConfigCopy, metav1.UpdateOptions{})

	return updateErr
}

func (c *Controller) getActiveVerificationTokenRule(namespace string) (*tratteria1alpha1.VerificationTokenRule, error) {
	tratteriaConfigs, err := c.tratteriaConfigsLister.TratteriaConfigs(namespace).List(labels.Everything())
	if err != nil {
		klog.Error("Failed to list TratteriaConfigs in namespace:", namespace, err)
		
		return nil, err
	}

	for _, config := range tratteriaConfigs {
		if config.Status.Status == "DONE" {
			verificationTokenRule, err := config.GetVerificationTokenRule()
			if err != nil {
				return nil, err
			}

			return verificationTokenRule, nil
		}
	}

	return nil, nil
}

func (c *Controller) getActiveGenerationTokenRule(namespace string) (*tratteria1alpha1.GenerationTokenRule, error) {
	tratteriaConfigs, err := c.tratteriaConfigsLister.TratteriaConfigs(namespace).List(labels.Everything())
	if err != nil {
		klog.Error("Failed to list TratteriaConfigs in namespace:", namespace, err)
		
		return nil, err
	}

	for _, config := range tratteriaConfigs {
		if config.Status.Status == "DONE" {
			generationTokenRule, err := config.GetGenerationTokenRule()
			if err != nil {
				return nil, err
			}

			return generationTokenRule, nil
		}
	}

	return nil, nil
}
