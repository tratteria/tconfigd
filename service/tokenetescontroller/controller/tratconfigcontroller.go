package controller

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	tokenetes1alpha1 "github.com/tokenetes/tconfigd/tokenetescontroller/pkg/apis/tokenetes/v1alpha1"
)

func (c *Controller) handleTokenetesConfigUpsert(ctx context.Context, newTokenetesConfig *tokenetes1alpha1.TokenetesConfig, versionNumber int64) error {
	verificationTokenRule, err := newTokenetesConfig.GetTokenetesConfigVerificationRule()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving verification token rule from %s tokenetes config: %w", newTokenetesConfig.Name, err)

		c.recorder.Event(newTokenetesConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTokenetesConfigStatus(ctx, newTokenetesConfig, VerificationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tokenetes config: %w", newTokenetesConfig.Name, updateErr)
		}

		return messagedErr
	}

	err = c.serviceMessageHandler.DispatchTokenetesConfigVerificationRule(ctx, newTokenetesConfig.Namespace, verificationTokenRule, versionNumber)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s tokenetes config verification token rule: %w", newTokenetesConfig.Name, err)

		c.recorder.Event(newTokenetesConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTokenetesConfigStatus(ctx, newTokenetesConfig, VerificationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tokenetes config: %w", newTokenetesConfig.Name, updateErr)
		}

		return messagedErr
	}

	c.recorder.Event(newTokenetesConfig, corev1.EventTypeNormal, string(VerificationApplicationStage)+" successful", string(VerificationApplicationStage)+" completed successfully")

	generationTokenRule, err := newTokenetesConfig.GetTokenetesConfigGenerationRule()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving generation token rules from %s tokenetes config: %w", newTokenetesConfig.Name, err)

		c.recorder.Event(newTokenetesConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTokenetesConfigStatus(ctx, newTokenetesConfig, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tokenetes config: %w", newTokenetesConfig.Name, updateErr)
		}

		return messagedErr
	}

	err = c.serviceMessageHandler.DispatchTokenetesConfigGenerationRule(ctx, newTokenetesConfig.Namespace, generationTokenRule, versionNumber)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s tokenetes config generation token rule: %w", newTokenetesConfig.Name, err)

		c.recorder.Event(newTokenetesConfig, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTokenetesConfigStatus(ctx, newTokenetesConfig, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tokenetes config: %w", newTokenetesConfig.Name, updateErr)
		}

		return messagedErr
	}

	if updateErr := c.updateSuccessTokenetesConfigStatus(ctx, newTokenetesConfig); updateErr != nil {
		return fmt.Errorf("failed to update success status for %s tokenetes config: %w", newTokenetesConfig.Name, updateErr)
	}

	c.recorder.Event(newTokenetesConfig, corev1.EventTypeNormal, string(GenerationApplicationStage)+" successful", string(GenerationApplicationStage)+" completed successfully")

	return nil
}

func (c *Controller) updateErrorTokenetesConfigStatus(ctx context.Context, tokenetesConfig *tokenetes1alpha1.TokenetesConfig, stage Stage, err error) error {
	tokenetesConfigCopy := tokenetesConfig.DeepCopy()

	tokenetesConfigCopy.Status.VerificationApplied = false
	tokenetesConfigCopy.Status.GenerationApplied = false
	tokenetesConfigCopy.Status.Status = string(PendingStatus)
	tokenetesConfigCopy.Status.Retries += 1

	if stage == GenerationApplicationStage {
		tokenetesConfigCopy.Status.VerificationApplied = true
	}

	tokenetesConfigCopy.Status.LastErrorMessage = err.Error()

	_, updateErr := c.tokenetesclientset.TokenetesV1alpha1().TokenetesConfigs(tokenetesConfig.Namespace).UpdateStatus(ctx, tokenetesConfigCopy, metav1.UpdateOptions{})

	return updateErr
}

func (c *Controller) updateSuccessTokenetesConfigStatus(ctx context.Context, tokenetesConfig *tokenetes1alpha1.TokenetesConfig) error {
	tokenetesConfigCopy := tokenetesConfig.DeepCopy()

	tokenetesConfigCopy.Status.VerificationApplied = true
	tokenetesConfigCopy.Status.GenerationApplied = true
	tokenetesConfigCopy.Status.Status = string(DoneStatus)

	_, updateErr := c.tokenetesclientset.TokenetesV1alpha1().TokenetesConfigs(tokenetesConfig.Namespace).UpdateStatus(ctx, tokenetesConfigCopy, metav1.UpdateOptions{})

	return updateErr
}

func (c *Controller) GetActiveTokenetesConfigVerificationRule(namespace string) (*tokenetes1alpha1.TokenetesConfigVerificationRule, error) {
	tokenetesConfigs, err := c.tokenetesConfigsLister.TokenetesConfigs(namespace).List(labels.Everything())
	if err != nil {
		c.logger.Error("Failed to list TokenetesConfigs in namespace.", zap.String("namespace", namespace), zap.Error(err))

		return nil, err
	}

	for _, config := range tokenetesConfigs {
		verificationTokenRule, err := config.GetTokenetesConfigVerificationRule()
		if err != nil {
			return nil, err
		} else {
			return verificationTokenRule, nil
		}
	}

	return nil, nil
}

func (c *Controller) GetActiveTokenetesConfigGenerationRule(namespace string) (*tokenetes1alpha1.TokenetesConfigGenerationRule, error) {
	tokenetesConfigs, err := c.tokenetesConfigsLister.TokenetesConfigs(namespace).List(labels.Everything())
	if err != nil {
		c.logger.Error("Failed to list TokenetesConfigs in namespace.", zap.String("namespace", namespace), zap.Error(err))

		return nil, err
	}

	for _, config := range tokenetesConfigs {
		generationTokenRule, err := config.GetTokenetesConfigGenerationRule()
		if err != nil {
			return nil, err
		} else {
			return generationTokenRule, nil

		}
	}

	return nil, nil
}
