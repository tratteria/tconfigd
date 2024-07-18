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

func (c *Controller) handleTraT(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))

		return nil
	}

	trat, err := c.traTsLister.TraTs(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("trat '%s' in work queue no longer exists", key))

			return nil
		}

		return err
	}

	verificationEndpointRules, err := trat.GetVerificationTraTRules()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving verification rules from %s trat: %w", name, err)

		c.recorder.Event(trat, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTraTStatus(ctx, trat, VerificationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s trat: %w", name, updateErr)
		}

		return messagedErr
	}

	// TODO: Implement parallel dispatching of rules using goroutines
	for service, serviceVerificationRule := range verificationEndpointRules {
		err := c.configDispatcher.DispatchVerificationTraTRule(ctx, service, namespace, serviceVerificationRule)
		if err != nil {
			messagedErr := fmt.Errorf("error dispatching %s trat verification rule to %s service: %w", name, service, err)

			c.recorder.Event(trat, corev1.EventTypeWarning, "error", messagedErr.Error())

			if updateErr := c.updateErrorTraTStatus(ctx, trat, VerificationApplicationStage, err); updateErr != nil {
				return fmt.Errorf("failed to update error status for %s trat: %w", name, updateErr)
			}

			return messagedErr
		}
	}

	c.recorder.Event(trat, corev1.EventTypeNormal, string(VerificationApplicationStage)+" successful", string(VerificationApplicationStage)+" completed successfully")

	generationEndpointRule, err := trat.GetGenerationTraTRule()
	if err != nil {
		messagedErr := fmt.Errorf("error retrieving generation rules from %s trat: %w", name, err)

		c.recorder.Event(trat, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTraTStatus(ctx, trat, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s trat: %w", name, updateErr)
		}

		return messagedErr
	}

	err = c.configDispatcher.DispatchGenerationTraTRule(ctx, namespace, generationEndpointRule)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s trat generation rule: %w", name, err)

		c.recorder.Event(trat, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateErrorTraTStatus(ctx, trat, GenerationApplicationStage, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s trat: %w", name, updateErr)
		}

		return messagedErr
	}

	if updateErr := c.updateSuccessTratStatus(ctx, trat); updateErr != nil {
		return fmt.Errorf("failed to update success status for %s trat: %w", name, updateErr)
	}

	c.recorder.Event(trat, corev1.EventTypeNormal, string(GenerationApplicationStage)+" successful", string(GenerationApplicationStage)+" completed successfully")

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

func (c *Controller) getActiveVerificationTraTRules(serviceName string, namespace string) ([]*tratteria1alpha1.VerificationTraTRule, error) {
	traTs, err := c.traTsLister.TraTs(namespace).List(labels.Everything())
	if err != nil {
		klog.Error("Failed to list TraTs in namespace:", namespace, err)

		return nil, err
	}

	var endpointVerificationRules []*tratteria1alpha1.VerificationTraTRule

	for _, trat := range traTs {
		if trat.Status.Status == "DONE" {
			endpointVerificationRule, err := trat.GetVerificationTraTRules()
			if err != nil {
				return nil, err
			}
			if serviceEndpointVerificationRule := endpointVerificationRule[serviceName]; serviceEndpointVerificationRule != nil {
				endpointVerificationRules = append(endpointVerificationRules, serviceEndpointVerificationRule)
			}
		}
	}

	return endpointVerificationRules, nil
}

func (c *Controller) getActiveGenerationEndpointRules(namespace string) ([]*tratteria1alpha1.GenerationTraTRule, error) {
	traTs, err := c.traTsLister.TraTs(namespace).List(labels.Everything())
	if err != nil {
		klog.Error("Failed to list TraTs in namespace:", namespace, err)

		return nil, err
	}

	var endpointGenerationRules []*tratteria1alpha1.GenerationTraTRule

	for _, trat := range traTs {
		if trat.Status.Status == "DONE" {
			endpointGenerationRule, err := trat.GetGenerationTraTRule()
			if err != nil {
				return nil, err
			}
			endpointGenerationRules = append(endpointGenerationRules, endpointGenerationRule)
		}
	}

	return endpointGenerationRules, nil
}
