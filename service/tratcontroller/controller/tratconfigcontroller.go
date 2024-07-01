package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	tratteria1alpha1 "github.com/tratteria/tconfigd/tratcontroller/pkg/apis/tratteria/v1alpha1"
)

func (c *Controller) handleTratConfig(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))

		return nil
	}

	traTconfig, err := c.traTConfigsLister.TraTConfigs(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("tratconfig '%s' in work queue no longer exists", key))

			return nil
		}

		return err
	}

	verificationTokenRule, err := traTconfig.GetVerificationTokenRule()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving verification token rule from %s tratconfig: %w", name, err)

		c.recorder.Event(traTconfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTraTConfigStatus(ctx, traTconfig, VerificationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratconfig: %w", name, updateErr)
		}

		return messagedErr
	}

	err = c.configDispatcher.DispatchVerificationTokenRule(ctx, namespace, verificationTokenRule)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s tratconfig verification token rule: %w", name, err)

		c.recorder.Event(traTconfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTraTConfigStatus(ctx, traTconfig, VerificationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratconfig: %w", name, updateErr)
		}

		return messagedErr
	}

	c.recorder.Event(traTconfig, corev1.EventTypeNormal, string(VerificationApplicationStage)+" successful", string(VerificationApplicationStage)+" completed successfully")

	generationTokenRule, err := traTconfig.GetGenerationTokenRule()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving generation token rules from %s tratconfig: %w", name, err)

		c.recorder.Event(traTconfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTraTConfigStatus(ctx, traTconfig, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratconfig: %w", name, updateErr)
		}

		return messagedErr
	}

	err = c.configDispatcher.DispatchGenerationTokentRule(ctx, namespace, generationTokenRule)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s tratconfig generation token rule: %w", name, err)

		c.recorder.Event(traTconfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTraTConfigStatus(ctx, traTconfig, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratconfig: %w", name, updateErr)
		}

		return messagedErr
	}

	if updateErr := c.updateSuccessTratConfigStatus(ctx, traTconfig); updateErr != nil {
		return fmt.Errorf("failed to update success status for %s tratconfig: %w", name, updateErr)
	}

	c.recorder.Event(traTconfig, corev1.EventTypeNormal, string(GenerationApplicationStage)+" successful", string(GenerationApplicationStage)+" completed successfully")

	return nil
}

func (c *Controller) updateErrorTraTConfigStatus(ctx context.Context, traTConfig *tratteria1alpha1.TraTConfig, stage Stage, err error) error {
	traTConfigCopy := traTConfig.DeepCopy()

	traTConfigCopy.Status.VerificationApplied = false
	traTConfigCopy.Status.GenerationApplied = false
	traTConfigCopy.Status.Status = string(PendingStatus)
	traTConfigCopy.Status.Retries += 1

	if stage == GenerationApplicationStage {
		traTConfigCopy.Status.VerificationApplied = true
	}

	traTConfigCopy.Status.LastErrorMessage = err.Error()

	_, updateErr := c.sampleclientset.TratteriaV1alpha1().TraTConfigs(traTConfig.Namespace).UpdateStatus(ctx, traTConfigCopy, metav1.UpdateOptions{})

	return updateErr
}

func (c *Controller) updateSuccessTratConfigStatus(ctx context.Context, traTConfig *tratteria1alpha1.TraTConfig) error {
	traTConfigCopy := traTConfig.DeepCopy()

	traTConfigCopy.Status.VerificationApplied = true
	traTConfigCopy.Status.GenerationApplied = true
	traTConfigCopy.Status.Status = string(DoneStatus)

	_, updateErr := c.sampleclientset.TratteriaV1alpha1().TraTConfigs(traTConfig.Namespace).UpdateStatus(ctx, traTConfigCopy, metav1.UpdateOptions{})

	return updateErr
}
