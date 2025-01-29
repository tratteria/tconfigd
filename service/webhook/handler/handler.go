package handler

import (
	"encoding/json"
	"net/http"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/tokenetes/tconfigd/webhook/pkg/util"

	"go.uber.org/zap"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Handlers struct {
	enableTratInterception bool
	agentHttpsApiPort      int
	agentHttpApiPort       int
	agentInterceptorPort   int
	spireAgentHostDir      string
	tconfigdSpiffeId       spiffeid.ID
	logger                 *zap.Logger
}

func NewHandlers(enableTratInterception bool, agentHttpApiPort int, agentInterceptorPort int, spireAgentHostDir string, tconfigdSpiffeId spiffeid.ID, logger *zap.Logger) *Handlers {
	return &Handlers{
		enableTratInterception: enableTratInterception,
		agentHttpApiPort:       agentHttpApiPort,
		agentInterceptorPort:   agentInterceptorPort,
		spireAgentHostDir:      spireAgentHostDir,
		tconfigdSpiffeId:       tconfigdSpiffeId,
		logger:                 logger,
	}
}

func (h *Handlers) InjectTokenetesAgent(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Received Agent Injection Request")

	var admissionReview admissionv1.AdmissionReview

	if err := json.NewDecoder(r.Body).Decode(&admissionReview); err != nil {
		h.logger.Error("Failed to decode admission review", zap.Error(err))
		http.Error(w, "could not decode admission review", http.StatusBadRequest)

		return
	}

	if admissionReview.Request == nil {
		h.logger.Error("Received an AdmissionReview with no Request")
		http.Error(w, "received an AdmissionReview with no Request", http.StatusBadRequest)

		return
	}

	admissionResponse := &admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
	}

	var pod corev1.Pod
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
		h.logger.Error("Could not unmarshal raw object into pod", zap.Error(err))
		admissionResponse.Result = &metav1.Status{
			Message: err.Error(),
		}
	} else {
		patchOps, err := util.CreatePodPatch(&pod, h.enableTratInterception, h.agentHttpsApiPort, h.agentHttpApiPort, h.agentInterceptorPort, h.spireAgentHostDir, h.tconfigdSpiffeId)

		if err != nil {
			h.logger.Error("Could not create patch for pod", zap.Error(err))
			admissionResponse.Result = &metav1.Status{
				Message: err.Error(),
			}
		} else {
			patchBytes, err := json.Marshal(patchOps)

			if err != nil {
				h.logger.Error("Failed to marshal patch operations", zap.Error(err))
				admissionResponse.Result = &metav1.Status{
					Message: err.Error(),
				}
			} else {
				admissionResponse.Patch = patchBytes
				admissionResponse.PatchType = new(admissionv1.PatchType)
				*admissionResponse.PatchType = admissionv1.PatchTypeJSONPatch
			}
		}
	}

	responseAdmissionReview := admissionv1.AdmissionReview{
		Response: admissionResponse,
	}

	responseAdmissionReview.TypeMeta = metav1.TypeMeta{
		Kind:       "AdmissionReview",
		APIVersion: "admission.k8s.io/v1",
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(responseAdmissionReview); err != nil {
		h.logger.Error("Failed to write response", zap.Error(err))
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}

	h.logger.Info("Agent Injection Request Processed Successfully")
}
