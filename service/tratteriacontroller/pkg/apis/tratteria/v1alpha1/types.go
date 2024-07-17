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

type VerificationEndpointRule struct {
	Endpoint   string     `json:"endpoint"`
	Method     string     `json:"method"`
	Purp       string     `json:"purp"`
	AzdMapping AzdMapping `json:"azdmapping,omitempty"`
}

type GenerationEndpointRule struct {
	Endpoint   string     `json:"endpoint"`
	Method     string     `json:"method"`
	Purp       string     `json:"purp"`
	AzdMapping AzdMapping `json:"azdmapping,omitempty"`
}

func (traT *TraT) GetVerificationEndpointRules() (map[string]*VerificationEndpointRule, error) {
	verificationRules := make(map[string]*VerificationEndpointRule)

	// TODO: do basic check and return err if failed

	for _, serviceSpec := range traT.Spec.Services {
		endpoint := traT.Spec.Endpoint
		azdMapping := traT.Spec.AzdMapping

		if serviceSpec.Endpoint != "" {
			endpoint = serviceSpec.Endpoint
		}

		if serviceSpec.AzdMapping != nil {
			azdMapping = serviceSpec.AzdMapping
		}

		verificationRules[serviceSpec.Name] = &VerificationEndpointRule{
			Endpoint:   endpoint,
			Method:     traT.Spec.Method,
			Purp:       traT.Spec.Purp,
			AzdMapping: azdMapping,
		}

	}

	if len(verificationRules) == 0 {
		return nil, fmt.Errorf("%w: verification rules for %s trat", tconfigderrors.ErrNotFound, traT.Name)
	}

	return verificationRules, nil
}

func (traT *TraT) GetGenerationEndpointRule() (*GenerationEndpointRule, error) {
	// TODO: do basic check and return err if failed

	return &GenerationEndpointRule{
		Endpoint:   traT.Spec.Endpoint,
		Method:     traT.Spec.Method,
		Purp:       traT.Spec.Purp,
		AzdMapping: traT.Spec.AzdMapping,
	}, nil
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TratteriaConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TratteriaConfigSpec   `json:"spec"`
	Status TratteriaConfigStatus `json:"status"`
}

type SubjectTokens struct {
	OIDC       *OIDCToken       `json:"OIDC,omitempty"`
	SelfSigned *SelfSignedToken `json:"selfSigned,omitempty"`
}

type OIDCToken struct {
	ClientID     string `json:"clientId"`
	ProviderURL  string `json:"providerURL"`
	SubjectField string `json:"subjectField"`
}

type SelfSignedToken struct {
	Validation    bool   `json:"validation"`
	JWKSSEndpoint string `json:"jwksEndpoint"`
}

type AccessEvaluationAPI struct {
	AnableAccessEvaluation bool           `json:"enableAccessEvaluation"`
	Endpoint               string         `json:"endpoint"`
	Authentication         Authentication `json:"authentication"`
}

type Authentication struct {
	Method string `json:"method"`
	Token  Token  `json:"token"`
}

type Token struct {
	Value string `json:"value"`
}

type TratteriaConfigSpec struct {
	Token struct {
		Issuer   string `json:"issuer"`
		Audience string `json:"audience"`
		LifeTime string `json:"lifeTime"`
	} `json:"token"`
	SubjectTokens                       SubjectTokens       `json:"subjectTokens"`
	AccessEvaluationAPI                 AccessEvaluationAPI `json:"accessEvaluationAPI"`
	TokenGenerationAuthorizedServiceIds []string            `json:"tokenGenerationAuthorizedServiceIds"`
}

type TratteriaConfigStatus struct {
	VerificationApplied bool   `json:"verificationApplied"`
	GenerationApplied   bool   `json:"generationApplied"`
	Status              string `json:"status"`
	LastErrorMessage    string `json:"lastErrorMessage,omitempty"`
	Retries             int32  `json:"retries"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TratteriaConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TratteriaConfig `json:"items"`
}

type VerificationTokenRule struct {
	Issuer   string `json:"issuer"`
	Audience string `json:"audience"`
}

type GenerationTokenRule TratteriaConfigSpec

func (tratteriaConfig *TratteriaConfig) GetVerificationTokenRule() (*VerificationTokenRule, error) {
	return &VerificationTokenRule{
		Issuer:   tratteriaConfig.Spec.Token.Issuer,
		Audience: tratteriaConfig.Spec.Token.Audience,
	}, nil
}

func (tratteriaConfig *TratteriaConfig) GetGenerationTokenRule() (*GenerationTokenRule, error) {
	generationTokenRule := GenerationTokenRule(tratteriaConfig.Spec)
	
	return &generationTokenRule, nil
}

type VerificationRules struct {
	VerificationTokenRule     *VerificationTokenRule      `json:"verificationTokenRule"`
	VerificationEndpointRules []*VerificationEndpointRule `json:"verificationEndpointRules"`
}

type GenerationRules struct {
	GenerationTokenRule     *GenerationTokenRule      `json:"generationTokenRule"`
	GenerationEndpointRules []*GenerationEndpointRule `json:"generationEndpointRules"`
}
