package util

import (
	"encoding/json"

	"github.com/mattbaird/jsonpatch"
	corev1 "k8s.io/api/core/v1"
)

func CreatePodPatch(pod *corev1.Pod) ([]jsonpatch.JsonPatchOperation, error) {
	var patch []jsonpatch.JsonPatchOperation

	serviceName := pod.Labels["app"]

	sidecar := corev1.Container{
		Name:  "tratteria-agent",
		Image: "tratteria-agent:latest",
		Env: []corev1.EnvVar{
			{
				Name:  "SERVICE",
				Value: serviceName,
			},
			{
				Name:  "TRATTERIA_URL",
				Value: "http://tconfigd.tratteria.svc.cluster.local:9060",
			},
		},
		Ports:           []corev1.ContainerPort{{ContainerPort: 80}},
		ImagePullPolicy: corev1.PullNever,
	}

	sidecarJson, err := json.Marshal(sidecar)
	if err != nil {
		return nil, err
	}

	patch = append(patch, jsonpatch.JsonPatchOperation{
		Operation: "add",
		Path:      "/spec/containers/-",
		Value:     json.RawMessage(sidecarJson),
	})

	return patch, nil
}
