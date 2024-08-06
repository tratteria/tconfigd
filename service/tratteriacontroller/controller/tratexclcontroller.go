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

func (c *Controller) handleTraTExclUpsert(ctx context.Context, newTraTExcl *tratteria1alpha1.TraTExclusion, versionNumber int64) error {
	traTExclRule := newTraTExcl.GetTraTExclRules()

	err := c.serviceMessageHandler.DispatchTraTExclRule(ctx, newTraTExcl.Spec.Service, newTraTExcl.Namespace, traTExclRule, versionNumber)
	if err != nil {
		messagedErr := fmt.Errorf("error dispatching %s tratexcl rule to %s service: %w", newTraTExcl.Name, newTraTExcl.Spec.Service, err)

		c.recorder.Event(newTraTExcl, corev1.EventTypeWarning, "error", messagedErr.Error())

		if updateErr := c.updateTraTExclErrorStatus(ctx, newTraTExcl, err); updateErr != nil {
			return fmt.Errorf("failed to update error status for %s tratexcl: %w", newTraTExcl.Name, updateErr)
		}

		return messagedErr
	}

	c.recorder.Event(newTraTExcl, corev1.EventTypeNormal, "success", "TraTExcl applied successfully")

	if updateErr := c.updateTraTExclSuccessStatus(ctx, newTraTExcl); updateErr != nil {
		return fmt.Errorf("failed to update success status for %s tratexcl: %w", newTraTExcl.Name, updateErr)
	}

	return nil
}

func (c *Controller) handleTraTExclDeletion(ctx context.Context, oldTraTExcl *tratteria1alpha1.TraTExclusion, versionNumber int64) error {
	err := c.serviceMessageHandler.DeleteTraTExcl(ctx, oldTraTExcl.Spec.Service, oldTraTExcl.Namespace, versionNumber)
	if err != nil {
		messagedErr := fmt.Errorf("error deleting %s tratexcl from %s service: %w", oldTraTExcl.Name, oldTraTExcl.Spec.Service, err)

		return messagedErr
	}

	return nil
}

func (c *Controller) updateTraTExclErrorStatus(ctx context.Context, tratExcl *tratteria1alpha1.TraTExclusion, err error) error {
	tratExclCopy := tratExcl.DeepCopy()

	tratExclCopy.Status.Status = string(PendingStatus)
	tratExclCopy.Status.Retries += 1

	tratExclCopy.Status.LastErrorMessage = err.Error()

	_, updateErr := c.tratteriaclientset.TratteriaV1alpha1().TraTExclusions(tratExcl.Namespace).UpdateStatus(ctx, tratExclCopy, metav1.UpdateOptions{})

	return updateErr
}

func (c *Controller) updateTraTExclSuccessStatus(ctx context.Context, tratExcl *tratteria1alpha1.TraTExclusion) error {
	tratExclCopy := tratExcl.DeepCopy()

	tratExclCopy.Status.Status = string(DoneStatus)

	_, updateErr := c.tratteriaclientset.TratteriaV1alpha1().TraTExclusions(tratExcl.Namespace).UpdateStatus(ctx, tratExclCopy, metav1.UpdateOptions{})

	return updateErr
}

func (c *Controller) GetActiveTraTExclRules(serviceName string, namespace string) (*tratteria1alpha1.TraTExclRule, error) {
	traTExcls, err := c.tratExclusionsLister.TraTExclusions(namespace).List(labels.Everything())
	if err != nil {
		c.logger.Error("Failed to list TraTExcls in namespace.", zap.String("namespace", namespace), zap.Error(err))

		return nil, err
	}

	for _, traTExcl := range traTExcls {
		if traTExcl.Spec.Service == serviceName {
			return traTExcl.GetTraTExclRules(), nil
		}
	}

	return nil, nil
}
