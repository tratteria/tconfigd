package util

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mattbaird/jsonpatch"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	corev1 "k8s.io/api/core/v1"
)

const (
	AGENT_INTERCEPTION_MODE = "interception"
	AGENT_DELEGATION_MODE   = "delegation"
)

func CreatePodPatch(pod *corev1.Pod, injectInitContainer bool, agentHttpsApiPort int, agentHttpApiPort int, agentInterceptorPort int, spireAgentHostDir string, tconfigdSpiffeId spiffeid.ID) ([]jsonpatch.JsonPatchOperation, error) {
	var patch []jsonpatch.JsonPatchOperation

	shouldInject, ok := pod.Annotations["tokenetes/inject-sidecar"]
	if !ok || shouldInject != "true" {
		return patch, nil
	}

	if mode, ok := pod.Annotations["tokenetes/agent-mode"]; ok {
		if mode == AGENT_INTERCEPTION_MODE {
			injectInitContainer = true
		} else if mode == AGENT_DELEGATION_MODE {
			injectInitContainer = false
		} else {
			return nil, fmt.Errorf("invalid agent-mode %v specified", mode)
		}
	}

	var servicePort string

	if injectInitContainer {
		var portOk bool

		servicePort, portOk = pod.Annotations["tokenetes/service-port"]

		if !portOk {
			return nil, fmt.Errorf("service-port must be specified when running in the interception mode")
		}

		if _, err := strconv.Atoi(servicePort); err != nil {
			return nil, fmt.Errorf("service-port must be a valid number")
		}
	}

	volumeName := "spire-agent-socket"
	foundVolume := false

	for _, vol := range pod.Spec.Volumes {
		if vol.HostPath != nil && vol.HostPath.Path == spireAgentHostDir {
			volumeName = vol.Name
			foundVolume = true

			break
		}
	}

	if !foundVolume {
		patch = append(patch, jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      "/spec/volumes/-",
			Value: corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: spireAgentHostDir,
					},
				},
			},
		})
	}

	if injectInitContainer {
		initContainer := corev1.Container{
			Name:  "tokenetes-agent-init",
			Image: "ghcr.io/tokenetes/tokenetes-agent-init:latest",
			Args:  []string{"-i", servicePort, "-p", strconv.Itoa(agentInterceptorPort)},
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
		Name:  "tokenetes-agent",
		Image: "ghcr.io/tokenetes/tokenetes-agent:latest",
		Env: []corev1.EnvVar{
			{
				Name:  "SERVICE_PORT",
				Value: servicePort,
			},
			{
				Name:  "INTERCEPTION_MODE",
				Value: strconv.FormatBool(injectInitContainer),
			},
			{
				Name:  "TCONFIGD_HOST",
				Value: "tconfigd.tokenetes-system.svc.cluster.local:8443",
			},
			{
				Name:  "TCONFIGD_SPIFFE_ID",
				Value: tconfigdSpiffeId.String(),
			},
			{
				Name:  "SPIFFE_ENDPOINT_SOCKET",
				Value: "unix:///run/spire/sockets/agent.sock",
			},
			{
				Name:  "AGENT_API_PORT",
				Value: strconv.Itoa(agentHttpApiPort),
			},
			{
				Name:  "AGENT_INTERCEPTOR_PORT",
				Value: strconv.Itoa(agentInterceptorPort),
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
		Ports: []corev1.ContainerPort{{ContainerPort: 9070}},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      volumeName,
				MountPath: "/run/spire/sockets",
				ReadOnly:  true,
			},
		},
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
