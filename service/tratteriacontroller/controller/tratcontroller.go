package controller

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/tratteria/tconfigd/common"
	tratteria1alpha1 "github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"
)

func (c *Controller) handleTraTUpsert(ctx context.Context, newTraT *tratteria1alpha1.TraT, versionNumber int64) error {
	servicestraTVerificationRules, err := newTraT.GetTraTVerificationRules()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving verification rules from %s trat: %w", newTraT.Name, err)

		c.recorder.Event(newTraT, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTraTStatus(ctx, newTraT, VerificationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s trat: %w", newTraT.Name, updateErr)
		}

		return messagedErr
	}

	// TODO: Implement parallel dispatching of rules using goroutines
	for service, serviceTraTVerificationRules := range servicestraTVerificationRules {
		err := c.serviceMessageHandler.DispatchTraTVerificationRule(ctx, service, newTraT.Namespace, serviceTraTVerificationRules, versionNumber)
		if err != nil {
			messagedErr := fmt.Errorf("error dispatching %s trat verification rule to %s service: %w", newTraT.Name, service, err)

			c.recorder.Event(newTraT, corev1.EventTypeWarning, "error", messagedErr.Error())

			if updateErr := c.updateErrorTraTStatus(ctx, newTraT, VerificationApplicationStage, err); updateErr != nil {
				return fmt.Errorf("failed to update error status for %s trat: %w", newTraT.Name, updateErr)
			}

			return messagedErr
		}
	}

	c.recorder.Event(newTraT, corev1.EventTypeNormal, string(VerificationApplicationStage)+" successful", string(VerificationApplicationStage)+" completed successfully")

	generationEndpointRule, err := newTraT.GetTraTGenerationRule()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving generation rules from %s trat: %w", newTraT.Name, err)

		c.recorder.Event(newTraT, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTraTStatus(ctx, newTraT, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s trat: %w", newTraT.Name, updateErr)
		}

		return messagedErr
	}

	err = c.serviceMessageHandler.DispatchTraTGenerationRule(ctx, newTraT.Namespace, generationEndpointRule, versionNumber)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s trat generation rule: %w", newTraT.Name, err)

		c.recorder.Event(newTraT, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTraTStatus(ctx, newTraT, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s trat: %w", newTraT.Name, updateErr)
		}

		return messagedErr
	}

	if updateErr := c.updateSuccessTratStatus(ctx, newTraT); updateErr != nil {
		return fmt.Errorf("failed to update success status for %s trat: %w", newTraT.Name, updateErr)
	}

	c.recorder.Event(newTraT, corev1.EventTypeNormal, string(GenerationApplicationStage)+" successful", string(GenerationApplicationStage)+" completed successfully")

	return nil
}

func (c *Controller) handleTraTUpdation(ctx context.Context, newTraT *tratteria1alpha1.TraT, oldTraT *tratteria1alpha1.TraT, versionNumber int64) error {
	// First, handle any service removals from the TraT
	newServices := make(map[string]bool)
	for _, newServiceSpec := range newTraT.Spec.Services {
		newServices[newServiceSpec.Name] = true
	}

	oldServices := make(map[string]bool)
	for _, oldServiceSpec := range oldTraT.Spec.Services {
		oldServices[oldServiceSpec.Name] = true
	}

	for oldService := range oldServices {
		if !newServices[oldService] {
			// Service was removed, so remove the TraT from it
			err := c.serviceMessageHandler.DeleteTraT(ctx, oldService, oldTraT.Namespace, oldTraT.Name, versionNumber)
			if err != nil {
				messagedErr := fmt.Errorf("error deleting %s trat from %s service: %w", oldTraT.Name, oldService, err)

				c.recorder.Event(newTraT, corev1.EventTypeWarning, "error", messagedErr.Error())

				if updateErr := c.updateErrorTraTStatus(ctx, newTraT, VerificationApplicationStage, err); updateErr != nil {
					return fmt.Errorf("failed to update error status for %s trat: %w", newTraT.Name, updateErr)
				}

				return messagedErr
			}
		}
	}

	// Now handle additions and updates
	return c.handleTraTUpsert(ctx, newTraT, versionNumber)
}

func (c *Controller) handleTraTDeletion(ctx context.Context, oldTraT *tratteria1alpha1.TraT, versionNumber int64) error {
	services := make(map[string]bool)

	for _, serviceSpec := range oldTraT.Spec.Services {
		services[serviceSpec.Name] = true
	}

	services[common.TRATTERIA_SERVICE_NAME] = true

	// TODO: Implement parallel requests using goroutines
	for service := range services {
		err := c.serviceMessageHandler.DeleteTraT(ctx, service, oldTraT.Namespace, oldTraT.Name, versionNumber)
		if err != nil {
			messagedErr := fmt.Errorf("error deleting %s trat from %s service: %w", oldTraT.Name, service, err)

			return messagedErr
		}
	}

	return nil
}

func (c *Controller) updateErrorTraTStatus(ctx context.Context, trat *tratteria1alpha1.TraT, stage Stage, err error) error {
	tratCopy := trat.DeepCopy()

	tratCopy.Status.VerificationApplied = false
	tratCopy.Status.GenerationApplied = false
	tratCopy.Status.Status = string(PendingStatus)
	tratCopy.Status.Retries += 1

	if stage == GenerationApplicationStage {
		tratCopy.Status.VerificationApplied = true
	}

	tratCopy.Status.LastErrorMessage = err.Error()

	_, updateErr := c.tratteriaclientset.TratteriaV1alpha1().TraTs(trat.Namespace).UpdateStatus(ctx, tratCopy, metav1.UpdateOptions{})

	return updateErr
}

func (c *Controller) updateSuccessTratStatus(ctx context.Context, trat *tratteria1alpha1.TraT) error {
	tratCopy := trat.DeepCopy()

	tratCopy.Status.VerificationApplied = true
	tratCopy.Status.GenerationApplied = true
	tratCopy.Status.Status = string(DoneStatus)

	_, updateErr := c.tratteriaclientset.TratteriaV1alpha1().TraTs(trat.Namespace).UpdateStatus(ctx, tratCopy, metav1.UpdateOptions{})

	return updateErr
}

func (c *Controller) GetActiveTraTsVerificationRules(serviceName string, namespace string) (map[string]*tratteria1alpha1.ServiceTraTVerificationRules, error) {
	traTs, err := c.traTsLister.TraTs(namespace).List(labels.Everything())
	if err != nil {
		c.logger.Error("Failed to list TraTs in namespace.", zap.String("namespace", namespace), zap.Error(err))

		return nil, err
	}

	serviceTraTsVerificationRules := make(map[string]*tratteria1alpha1.ServiceTraTVerificationRules)

	for _, traT := range traTs {
		traTVerificationRules, err := traT.GetTraTVerificationRules()
		if err != nil {
			return nil, err
		}

		if serviceTraTVerificationRules := traTVerificationRules[serviceName]; serviceTraTVerificationRules != nil {
			serviceTraTsVerificationRules[traT.Name] = serviceTraTVerificationRules
		}
	}

	return serviceTraTsVerificationRules, nil
}

func (c *Controller) GetActiveTraTsGenerationRules(namespace string) (map[string]*tratteria1alpha1.TraTGenerationRule, error) {
	traTs, err := c.traTsLister.TraTs(namespace).List(labels.Everything())
	if err != nil {
		c.logger.Error("Failed to list TraTs in namespace.", zap.String("namespace", namespace), zap.Error(err))

		return nil, err
	}

	traTsGenerationRules := make(map[string]*tratteria1alpha1.TraTGenerationRule)

	for _, traT := range traTs {
		traTGenerationRule, err := traT.GetTraTGenerationRule()
		if err != nil {
			return nil, err
		}

		traTsGenerationRules[traT.Name] = traTGenerationRule
	}

	return traTsGenerationRules, nil
}
