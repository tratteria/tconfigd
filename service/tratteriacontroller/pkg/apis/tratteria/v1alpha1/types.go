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

type TraTVerificationRule struct {
	Endpoint   string     `json:"endpoint"`
	Method     string     `json:"method"`
	Purp       string     `json:"purp"`
	AzdMapping AzdMapping `json:"azdmapping,omitempty"`
}

type TraTGenerationRule struct {
	Endpoint   string     `json:"endpoint"`
	Method     string     `json:"method"`
	Purp       string     `json:"purp"`
	AzdMapping AzdMapping `json:"azdmapping,omitempty"`
}

func (traT *TraT) GetTraTVerificationRules() (map[string]*TraTVerificationRule, error) {
	verificationRules := make(map[string]*TraTVerificationRule)

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

		verificationRules[serviceSpec.Name] = &TraTVerificationRule{
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

func (traT *TraT) GetTraTGenerationRule() (*TraTGenerationRule, error) {
	// TODO: do basic check and return err if failed

	return &TraTGenerationRule{
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

type TratteriaConfigVerificationRule struct {
	Issuer   string `json:"issuer"`
	Audience string `json:"audience"`
}

type TratteriaConfigGenerationRule TratteriaConfigSpec

func (tratteriaConfig *TratteriaConfig) GetTratteriaConfigVerificationRule() (*TratteriaConfigVerificationRule, error) {
	return &TratteriaConfigVerificationRule{
		Issuer:   tratteriaConfig.Spec.Token.Issuer,
		Audience: tratteriaConfig.Spec.Token.Audience,
	}, nil
}

func (tratteriaConfig *TratteriaConfig) GetTratteriaConfigGenerationRule() (*TratteriaConfigGenerationRule, error) {
	generationTratteriaConfigRule := TratteriaConfigGenerationRule(tratteriaConfig.Spec)

	return &generationTratteriaConfigRule, nil
}

type VerificationRules struct {
	TratteriaConfigVerificationRule *TratteriaConfigVerificationRule `json:"tratteriaConfigVerificationRule"`
	TraTVerificationRules           []*TraTVerificationRule          `json:"traTVerificationRules"`
}

type GenerationRules struct {
	TratteriaConfigGenerationRule *TratteriaConfigGenerationRule `json:"tratteriaConfigGenerationRule"`
	TraTGenerationRules           []*TraTGenerationRule          `json:"traTGenerationRules"`
}
