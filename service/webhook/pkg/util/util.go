package util

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mattbaird/jsonpatch"
	"github.com/tratteria/tconfigd/common"
	corev1 "k8s.io/api/core/v1"
)

func CreatePodPatch(pod *corev1.Pod, injectInitContainer bool, agentApiPort int, agentInterceptorPort int) ([]jsonpatch.JsonPatchOperation, error) {
	var patch []jsonpatch.JsonPatchOperation

	shouldInject, ok := pod.Annotations["tratteria/inject-sidecar"]
	if !ok || shouldInject != "true" {
		return patch, nil
	}

	serviceName, nameOk := pod.Annotations["tratteria/service-name"]
	servicePort, portOk := pod.Annotations["tratteria/service-port"]

	if !nameOk || !portOk {
		return nil, fmt.Errorf("service-name and service-port must be specified when inject-sidecar is 'true'")
	}

	if _, err := strconv.Atoi(servicePort); err != nil {
		return nil, fmt.Errorf("service-port must be a valid number")
	}

	if injectInitContainer {
		initContainer := corev1.Container{
			Name:            "tratteria-agent-init",
			Image:           "tratteria-agent-init:latest",
			Args:            []string{"-i", servicePort, "-p", strconv.Itoa(agentInterceptorPort)},
			ImagePullPolicy: corev1.PullNever,
			SecurityContext: &corev1.SecurityContext{
				Capabilities: &corev1.Capabilities{
					Add: []corev1.Capability{"NET_ADMIN"},
				},
			},
		}

		initContainerJson, err := json.Marshal(initContainer)
		if err != nil {
			return nil, err
		}

		if pod.Spec.InitContainers == nil {
			patch = append(patch, jsonpatch.JsonPatchOperation{
				Operation: "add",
				Path:      "/spec/initContainers",
				Value:     []json.RawMessage{json.RawMessage(initContainerJson)},
			})
		} else {
			patch = append(patch, jsonpatch.JsonPatchOperation{
				Operation: "add",
				Path:      "/spec/initContainers/-",
				Value:     json.RawMessage(initContainerJson),
			})
		}
	}

	sidecar := corev1.Container{
		Name:  "tratteria-agent",
		Image: "tratteria-agent:latest",
		Env: []corev1.EnvVar{
			{
				Name:  "SERVICE_NAME",
				Value: serviceName,
			},
			{
				Name:  "SERVICE_PORT",
				Value: servicePort,
			},
			{
				Name:  "TCONFIGD_URL",
				Value: "http://tconfigd.tratteria-system.svc.cluster.local:9060",
			},
			{
				Name:  "AGENT_API_PORT",
				Value: strconv.Itoa(agentApiPort),
			},
			{
				Name:  "AGENT_INTERCEPTOR_PORT",
				Value: strconv.Itoa(agentInterceptorPort),
			},
			{
				Name:  "HEARTBEAT_INTERVAL_MINUTES",
				Value: strconv.Itoa(common.DATA_PLANE_HEARTBEAT_INTERVAL_MINUTES),
			},
			{
				Name: "MY_NAMESPACE",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
		},
		Ports:           []corev1.ContainerPort{{ContainerPort: 9070}},
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
