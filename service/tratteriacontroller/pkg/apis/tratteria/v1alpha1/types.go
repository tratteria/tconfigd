package v1alpha1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/tratteria/tconfigd/tconfigderrors"
	"github.com/tratteria/tconfigd/utils"
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
	Method     string     `json:"method,omitempty"`
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
	TraTName   string     `json:"traTName"`
	Endpoint   string     `json:"endpoint"`
	Method     string     `json:"method"`
	Purp       string     `json:"purp"`
	AzdMapping AzdMapping `json:"azdmapping,omitempty"`
}

type ServiceTraTVerificationRules struct {
	TraTName              string
	TraTVerificationRules []*TraTVerificationRule
}

type TraTGenerationRule struct {
	TraTName   string     `json:"traTName"`
	Endpoint   string     `json:"endpoint"`
	Method     string     `json:"method"`
	Purp       string     `json:"purp"`
	AzdMapping AzdMapping `json:"azdmapping,omitempty"`
}

// constructs TraT verification for each service present in the call chain
// a single service can have multiple different APIs present in the call chain, so it return the map of list of TraTVerificationRule
func (traT *TraT) GetTraTVerificationRules() (map[string]*ServiceTraTVerificationRules, error) {
	servicesTraTVerificationRules := make(map[string]*ServiceTraTVerificationRules)

	for _, serviceSpec := range traT.Spec.Services {
		endpoint := traT.Spec.Endpoint
		method := traT.Spec.Method
		azdMapping := traT.Spec.AzdMapping

		if serviceSpec.Endpoint != "" {
			endpoint = serviceSpec.Endpoint
		}

		if serviceSpec.Method != "" {
			method = serviceSpec.Method
		}

		if serviceSpec.AzdMapping != nil {
			azdMapping = serviceSpec.AzdMapping
		}

		if servicesTraTVerificationRules[serviceSpec.Name] == nil {
			servicesTraTVerificationRules[serviceSpec.Name] = &ServiceTraTVerificationRules{
				TraTName: traT.Name,
			}
		}

		servicesTraTVerificationRules[serviceSpec.Name].TraTVerificationRules = append(
			servicesTraTVerificationRules[serviceSpec.Name].TraTVerificationRules,
			&TraTVerificationRule{
				TraTName:   traT.Name,
				Endpoint:   endpoint,
				Method:     method,
				Purp:       traT.Spec.Purp,
				AzdMapping: azdMapping,
			})
	}

	if len(servicesTraTVerificationRules) == 0 {
		return nil, fmt.Errorf("%w: verification rules for %s trat", tconfigderrors.ErrNotFound, traT.Name)
	}

	return servicesTraTVerificationRules, nil
}

func (traT *TraT) GetTraTGenerationRule() (*TraTGenerationRule, error) {

	return &TraTGenerationRule{
		TraTName:   traT.Name,
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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TraTExclusion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TraTExclusionSpec   `json:"spec"`
	Status TraTExclusionStatus `json:"status,omitempty"`
}

type TraTExclusionSpec struct {
	Service   string     `json:"service"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type TraTExclusionStatus struct {
	Status           string `json:"status,omitempty"`
	LastErrorMessage string `json:"lastErrorMessage,omitempty"`
	Retries          int32  `json:"retries,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TraTExclusionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TraTExclusion `json:"items"`
}

type TraTExclRule struct {
	Endpoints []Endpoint `json:"endpoints"`
}

func (traTExclusion *TraTExclusion) GetTraTExclRules() *TraTExclRule {
	return &TraTExclRule{
		Endpoints: traTExclusion.Spec.Endpoints,
	}
}

type VerificationRules struct {
	TratteriaConfigVerificationRule *TratteriaConfigVerificationRule         `json:"tratteriaConfigVerificationRule"`
	TraTsVerificationRules          map[string]*ServiceTraTVerificationRules `json:"traTsVerificationRules"`
	TraTExclRule                    *TraTExclRule                            `json:"traTExclRule"`
}

func (verificationRules *VerificationRules) ComputeStableHash() (string, error) {
	data, err := json.Marshal(verificationRules)
	if err != nil {
		return "", fmt.Errorf("failed to marshal rules: %w", err)
	}

	var jsonData interface{}

	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal for canonicalization: %w", err)
	}

	canonicalizedData, err := utils.CanonicalizeJSON(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to canonicalize JSON: %w", err)
	}

	hash := sha256.Sum256([]byte(canonicalizedData))

	return hex.EncodeToString(hash[:]), nil
}

type GenerationRules struct {
	TratteriaConfigGenerationRule *TratteriaConfigGenerationRule `json:"tratteriaConfigGenerationRule"`
	TraTsGenerationRules          map[string]*TraTGenerationRule `json:"traTsGenerationRules"`
}

func (generationRules *GenerationRules) ComputeStableHash() (string, error) {
	data, err := json.Marshal(generationRules)
	if err != nil {
		return "", fmt.Errorf("failed to marshal rules: %w", err)
	}

	var jsonData interface{}

	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal for canonicalization: %w", err)
	}

	canonicalizedData, err := utils.CanonicalizeJSON(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to canonicalize JSON: %w", err)
	}

	hash := sha256.Sum256([]byte(canonicalizedData))

	return hex.EncodeToString(hash[:]), nil
}
