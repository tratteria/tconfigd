package v1alpha1

import (
	"fmt"

	"github.com/tratteria/tconfigd/tconfigderrors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TraT struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TraTSpec   `json:"spec"`
	Status TraTStatus `json:"status"`
}

type TraTSpec struct {
	Endpoint   string              `json:"endpoint"`
	Method     string              `json:"method"`
	Purp       string              `json:"purp"`
	AzdMapping map[string]AzdField `json:"azdMapping,omitempty"`
	Services   []ServiceSpec       `json:"services"`
}

type ServiceSpec struct {
	Name       string     `json:"name"`
	Endpoint   string     `json:"endpoint,omitempty"`
	AzdMapping AzdMapping `json:"azdMapping,omitempty"`
}

type AzdMapping map[string]AzdField
type AzdField struct {
	Required bool   `json:"required"`
	Value    string `json:"value"`
}

type TraTStatus struct {
	VerificationApplied bool   `json:"verificationApplied"`
	GenerationApplied   bool   `json:"generationApplied"`
	Status              string `json:"status"`
	LastErrorMessage    string `json:"lastErrorMessage,omitempty"`
	Retries             int32  `json:"retries"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TraTList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TraT `json:"items"`
}

type VerificationRule struct {
	Endpoint   string     `json:"endpoint"`
	Method     string     `json:"method"`
	Purp       string     `json:"purp"`
	AzdMapping AzdMapping `json:"azdmapping,omitempty"`
}

type GenerationRule struct {
	Endpoint   string     `json:"endpoint"`
	Method     string     `json:"method"`
	Purp       string     `json:"purp"`
	AzdMapping AzdMapping `json:"azdmapping,omitempty"`
}

func (trat *TraT) GetVerificationRules() (map[string]*VerificationRule, error) {
	verificationRules := make(map[string]*VerificationRule)

	// TODO: do basic check and return err if failed

	for _, serviceSpec := range trat.Spec.Services {
		endpoint := trat.Spec.Endpoint
		azdMapping := trat.Spec.AzdMapping

		if serviceSpec.Endpoint != "" {
			endpoint = serviceSpec.Endpoint
		}

		if serviceSpec.AzdMapping != nil {
			azdMapping = serviceSpec.AzdMapping
		}

		verificationRules[serviceSpec.Name] = &VerificationRule{
			Endpoint:   endpoint,
			Method:     trat.Spec.Method,
			Purp:       trat.Spec.Purp,
			AzdMapping: azdMapping,
		}

	}

	if len(verificationRules) == 0 {
		return nil, fmt.Errorf("%w: verification rules for %s trat", tconfigderrors.ErrNotFound, trat.Name)
	}

	return verificationRules, nil
}

func (trat *TraT) GetGenerationRule() (*GenerationRule, error) {
	// TODO: do basic check and return err if failed

	return &GenerationRule{
		Endpoint:   trat.Spec.Endpoint,
		Method:     trat.Spec.Method,
		Purp:       trat.Spec.Purp,
		AzdMapping: trat.Spec.AzdMapping,
	}, nil
}
