package v1alpha1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/tokenetes/tconfigd/tconfigderrors"
	"github.com/tokenetes/tconfigd/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DynamicMap is a wrapper around map[string]interface{} that implements DeepCopyInterface
type DynamicMap struct {
	Map map[string]interface{} `json:"-"`
}

func (in *DynamicMap) DeepCopyInterface() interface{} {
	if in == nil {
		return nil
	}

	out := new(DynamicMap)

	in.DeepCopyInto(out)

	return out
}

func (in *DynamicMap) DeepCopyInto(out *DynamicMap) {
	clone := make(map[string]interface{})

	for k, v := range in.Map {
		clone[k] = deepCopyJSONValue(v)
	}

	out.Map = clone
}

func deepCopyJSONValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case []interface{}:
		arr := make([]interface{}, len(v))

		for i, elem := range v {
			arr[i] = deepCopyJSONValue(elem)
		}

		return arr
	case map[string]interface{}:
		m := make(map[string]interface{})

		for k, val := range v {
			m[k] = deepCopyJSONValue(val)
		}

		return m
	default:
		return v
	}
}

func (in *DynamicMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(in.Map)
}

func (in *DynamicMap) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &in.Map)
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TraT struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TraTSpec   `json:"spec"`
	Status TraTStatus `json:"status"`
}

type TraTSpec struct {
	Path             string              `json:"path"`
	Method           string              `json:"method"`
	Purp             string              `json:"purp"`
	AzdMapping       map[string]AzdField `json:"azdMapping,omitempty"`
	Services         []ServiceSpec       `json:"services"`
	AccessEvaluation *DynamicMap         `json:"accessEvaluation,omitempty"`
}

type ServiceSpec struct {
	Name       string     `json:"name"`
	Path       string     `json:"path,omitempty"`
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
	Path       string     `json:"path"`
	Method     string     `json:"method"`
	Purp       string     `json:"purp"`
	AzdMapping AzdMapping `json:"azdmapping,omitempty"`
}

type ServiceTraTVerificationRules struct {
	TraTName              string
	TraTVerificationRules []*TraTVerificationRule
}

type TraTGenerationRule struct {
	TraTName         string      `json:"traTName"`
	Path             string      `json:"path"`
	Method           string      `json:"method"`
	Purp             string      `json:"purp"`
	AzdMapping       AzdMapping  `json:"azdmapping,omitempty"`
	AccessEvaluation *DynamicMap `json:"accessEvaluation,omitempty"`
}

// constructs TraT verification for each service present in the call chain
// a single service can have multiple different APIs present in the call chain, so it return the map of list of TraTVerificationRule
func (traT *TraT) GetTraTVerificationRules() (map[string]*ServiceTraTVerificationRules, error) {
	servicesTraTVerificationRules := make(map[string]*ServiceTraTVerificationRules)

	for _, serviceSpec := range traT.Spec.Services {
		path := traT.Spec.Path
		method := traT.Spec.Method
		azdMapping := traT.Spec.AzdMapping

		if serviceSpec.Path != "" {
			path = serviceSpec.Path
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
				Path:       path,
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
		TraTName:         traT.Name,
		Path:             traT.Spec.Path,
		Method:           traT.Spec.Method,
		Purp:             traT.Spec.Purp,
		AzdMapping:       traT.Spec.AzdMapping,
		AccessEvaluation: traT.Spec.AccessEvaluation,
	}, nil
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TokenetesConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TokenetesConfigSpec   `json:"spec"`
	Status TokenetesConfigStatus `json:"status"`
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

type TokenetesConfigSpec struct {
	Token struct {
		Issuer   string `json:"issuer"`
		Audience string `json:"audience"`
		LifeTime string `json:"lifeTime"`
	} `json:"token"`
	SubjectTokens                       SubjectTokens       `json:"subjectTokens"`
	AccessEvaluationAPI                 AccessEvaluationAPI `json:"accessEvaluationAPI"`
	TokenGenerationAuthorizedServiceIds []string            `json:"tokenGenerationAuthorizedServiceIds"`
}

type TokenetesConfigStatus struct {
	VerificationApplied bool   `json:"verificationApplied"`
	GenerationApplied   bool   `json:"generationApplied"`
	Status              string `json:"status"`
	LastErrorMessage    string `json:"lastErrorMessage,omitempty"`
	Retries             int32  `json:"retries"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TokenetesConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TokenetesConfig `json:"items"`
}

type TokenetesConfigVerificationRule struct {
	Issuer   string `json:"issuer"`
	Audience string `json:"audience"`
}

type TokenetesConfigGenerationRule TokenetesConfigSpec

func (tokenetesConfig *TokenetesConfig) GetTokenetesConfigVerificationRule() (*TokenetesConfigVerificationRule, error) {
	return &TokenetesConfigVerificationRule{
		Issuer:   tokenetesConfig.Spec.Token.Issuer,
		Audience: tokenetesConfig.Spec.Token.Audience,
	}, nil
}

func (tokenetesConfig *TokenetesConfig) GetTokenetesConfigGenerationRule() (*TokenetesConfigGenerationRule, error) {
	generationTokenetesConfigRule := TokenetesConfigGenerationRule(tokenetesConfig.Spec)

	return &generationTokenetesConfigRule, nil
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
	TokenetesConfigVerificationRule *TokenetesConfigVerificationRule         `json:"tokenetesConfigVerificationRule"`
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
	TokenetesConfigGenerationRule *TokenetesConfigGenerationRule `json:"tokenetesConfigGenerationRule"`
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
