package controller

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	tratteria1alpha1 "github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"
)

func (c *Controller) handleTratteriaConfigUpsert(ctx context.Context, newTratteriaConfig *tratteria1alpha1.TratteriaConfig, versionNumber int64) error {
	verificationTokenRule, err := newTratteriaConfig.GetTratteriaConfigVerificationRule()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving verification token rule from %s tratteria config: %w", newTratteriaConfig.Name, err)

		c.recorder.Event(newTratteriaConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTratteriaConfigStatus(ctx, newTratteriaConfig, VerificationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratteria config: %w", newTratteriaConfig.Name, updateErr)
		}

		return messagedErr
	}

	err = c.serviceMessageHandler.DispatchTratteriaConfigVerificationRule(ctx, newTratteriaConfig.Namespace, verificationTokenRule, versionNumber)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s tratteria config verification token rule: %w", newTratteriaConfig.Name, err)

		c.recorder.Event(newTratteriaConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTratteriaConfigStatus(ctx, newTratteriaConfig, VerificationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratteria config: %w", newTratteriaConfig.Name, updateErr)
		}

		return messagedErr
	}

	c.recorder.Event(newTratteriaConfig, corev1.EventTypeNormal, string(VerificationApplicationStage)+" successful", string(VerificationApplicationStage)+" completed successfully")

	generationTokenRule, err := newTratteriaConfig.GetTratteriaConfigGenerationRule()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving generation token rules from %s tratteria config: %w", newTratteriaConfig.Name, err)

		c.recorder.Event(newTratteriaConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTratteriaConfigStatus(ctx, newTratteriaConfig, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratteria config: %w", newTratteriaConfig.Name, updateErr)
		}

		return messagedErr
	}

	err = c.serviceMessageHandler.DispatchTratteriaConfigGenerationRule(ctx, newTratteriaConfig.Namespace, generationTokenRule, versionNumber)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s tratteria config generation token rule: %w", newTratteriaConfig.Name, err)

		c.recorder.Event(newTratteriaConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTratteriaConfigStatus(ctx, newTratteriaConfig, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratteria config: %w", newTratteriaConfig.Name, updateErr)
		}

		return messagedErr
	}

	if updateErr := c.updateSuccessTratteriaConfigStatus(ctx, newTratteriaConfig); updateErr != nil {
		return fmt.Errorf("failed to update success status for %s tratteria config: %w", newTratteriaConfig.Name, updateErr)
	}

	c.recorder.Event(newTratteriaConfig, corev1.EventTypeNormal, string(GenerationApplicationStage)+" successful", string(GenerationApplicationStage)+" completed successfully")

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

func (c *Controller) GetActiveTratteriaConfigVerificationRule(namespace string) (*tratteria1alpha1.TratteriaConfigVerificationRule, error) {
	tratteriaConfigs, err := c.tratteriaConfigsLister.TratteriaConfigs(namespace).List(labels.Everything())
	if err != nil {
		c.logger.Error("Failed to list TratteriaConfigs in namespace.", zap.String("namespace", namespace), zap.Error(err))

		return nil, err
	}

	for _, config := range tratteriaConfigs {
		verificationTokenRule, err := config.GetTratteriaConfigVerificationRule()
		if err != nil {
			return nil, err
		} else {
			return verificationTokenRule, nil
		}
	}

	return nil, nil
}

func (c *Controller) GetActiveTratteriaConfigGenerationRule(namespace string) (*tratteria1alpha1.TratteriaConfigGenerationRule, error) {
	tratteriaConfigs, err := c.tratteriaConfigsLister.TratteriaConfigs(namespace).List(labels.Everything())
	if err != nil {
		c.logger.Error("Failed to list TratteriaConfigs in namespace.", zap.String("namespace", namespace), zap.Error(err))

		return nil, err
	}

	for _, config := range tratteriaConfigs {
		generationTokenRule, err := config.GetTratteriaConfigGenerationRule()
		if err != nil {
			return nil, err
		} else {
			return generationTokenRule, nil

		}
	}

	return nil, nil
}
