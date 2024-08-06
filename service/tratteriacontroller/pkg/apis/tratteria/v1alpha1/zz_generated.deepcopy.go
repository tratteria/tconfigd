//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessEvaluationAPI) DeepCopyInto(out *AccessEvaluationAPI) {
	*out = *in
	out.Authentication = in.Authentication
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessEvaluationAPI.
func (in *AccessEvaluationAPI) DeepCopy() *AccessEvaluationAPI {
	if in == nil {
		return nil
	}
	out := new(AccessEvaluationAPI)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Authentication) DeepCopyInto(out *Authentication) {
	*out = *in
	out.Token = in.Token
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Authentication.
func (in *Authentication) DeepCopy() *Authentication {
	if in == nil {
		return nil
	}
	out := new(Authentication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AzdField) DeepCopyInto(out *AzdField) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzdField.
func (in *AzdField) DeepCopy() *AzdField {
	if in == nil {
		return nil
	}
	out := new(AzdField)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in AzdMapping) DeepCopyInto(out *AzdMapping) {
	{
		in := &in
		*out = make(AzdMapping, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AzdMapping.
func (in AzdMapping) DeepCopy() AzdMapping {
	if in == nil {
		return nil
	}
	out := new(AzdMapping)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DynamicMap.
func (in *DynamicMap) DeepCopy() *DynamicMap {
	if in == nil {
		return nil
	}
	out := new(DynamicMap)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Endpoint) DeepCopyInto(out *Endpoint) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Endpoint.
func (in *Endpoint) DeepCopy() *Endpoint {
	if in == nil {
		return nil
	}
	out := new(Endpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GenerationRules) DeepCopyInto(out *GenerationRules) {
	*out = *in
	if in.TratteriaConfigGenerationRule != nil {
		in, out := &in.TratteriaConfigGenerationRule, &out.TratteriaConfigGenerationRule
		*out = new(TratteriaConfigGenerationRule)
		(*in).DeepCopyInto(*out)
	}
	if in.TraTsGenerationRules != nil {
		in, out := &in.TraTsGenerationRules, &out.TraTsGenerationRules
		*out = make(map[string]*TraTGenerationRule, len(*in))
		for key, val := range *in {
			var outVal *TraTGenerationRule
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(TraTGenerationRule)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GenerationRules.
func (in *GenerationRules) DeepCopy() *GenerationRules {
	if in == nil {
		return nil
	}
	out := new(GenerationRules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCToken) DeepCopyInto(out *OIDCToken) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCToken.
func (in *OIDCToken) DeepCopy() *OIDCToken {
	if in == nil {
		return nil
	}
	out := new(OIDCToken)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SelfSignedToken) DeepCopyInto(out *SelfSignedToken) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SelfSignedToken.
func (in *SelfSignedToken) DeepCopy() *SelfSignedToken {
	if in == nil {
		return nil
	}
	out := new(SelfSignedToken)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSpec) DeepCopyInto(out *ServiceSpec) {
	*out = *in
	if in.AzdMapping != nil {
		in, out := &in.AzdMapping, &out.AzdMapping
		*out = make(AzdMapping, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSpec.
func (in *ServiceSpec) DeepCopy() *ServiceSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceTraTVerificationRules) DeepCopyInto(out *ServiceTraTVerificationRules) {
	*out = *in
	if in.TraTVerificationRules != nil {
		in, out := &in.TraTVerificationRules, &out.TraTVerificationRules
		*out = make([]*TraTVerificationRule, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(TraTVerificationRule)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceTraTVerificationRules.
func (in *ServiceTraTVerificationRules) DeepCopy() *ServiceTraTVerificationRules {
	if in == nil {
		return nil
	}
	out := new(ServiceTraTVerificationRules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubjectTokens) DeepCopyInto(out *SubjectTokens) {
	*out = *in
	if in.OIDC != nil {
		in, out := &in.OIDC, &out.OIDC
		*out = new(OIDCToken)
		**out = **in
	}
	if in.SelfSigned != nil {
		in, out := &in.SelfSigned, &out.SelfSigned
		*out = new(SelfSignedToken)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubjectTokens.
func (in *SubjectTokens) DeepCopy() *SubjectTokens {
	if in == nil {
		return nil
	}
	out := new(SubjectTokens)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Token) DeepCopyInto(out *Token) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Token.
func (in *Token) DeepCopy() *Token {
	if in == nil {
		return nil
	}
	out := new(Token)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraT) DeepCopyInto(out *TraT) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraT.
func (in *TraT) DeepCopy() *TraT {
	if in == nil {
		return nil
	}
	out := new(TraT)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TraT) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraTExclRule) DeepCopyInto(out *TraTExclRule) {
	*out = *in
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]Endpoint, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraTExclRule.
func (in *TraTExclRule) DeepCopy() *TraTExclRule {
	if in == nil {
		return nil
	}
	out := new(TraTExclRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraTExclusion) DeepCopyInto(out *TraTExclusion) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraTExclusion.
func (in *TraTExclusion) DeepCopy() *TraTExclusion {
	if in == nil {
		return nil
	}
	out := new(TraTExclusion)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TraTExclusion) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraTExclusionList) DeepCopyInto(out *TraTExclusionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TraTExclusion, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraTExclusionList.
func (in *TraTExclusionList) DeepCopy() *TraTExclusionList {
	if in == nil {
		return nil
	}
	out := new(TraTExclusionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TraTExclusionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraTExclusionSpec) DeepCopyInto(out *TraTExclusionSpec) {
	*out = *in
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]Endpoint, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraTExclusionSpec.
func (in *TraTExclusionSpec) DeepCopy() *TraTExclusionSpec {
	if in == nil {
		return nil
	}
	out := new(TraTExclusionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraTExclusionStatus) DeepCopyInto(out *TraTExclusionStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraTExclusionStatus.
func (in *TraTExclusionStatus) DeepCopy() *TraTExclusionStatus {
	if in == nil {
		return nil
	}
	out := new(TraTExclusionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraTGenerationRule) DeepCopyInto(out *TraTGenerationRule) {
	*out = *in
	if in.AzdMapping != nil {
		in, out := &in.AzdMapping, &out.AzdMapping
		*out = make(AzdMapping, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.AccessEvaluation != nil {
		in, out := &in.AccessEvaluation, &out.AccessEvaluation
		*out = (*in).DeepCopy()
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraTGenerationRule.
func (in *TraTGenerationRule) DeepCopy() *TraTGenerationRule {
	if in == nil {
		return nil
	}
	out := new(TraTGenerationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraTList) DeepCopyInto(out *TraTList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TraT, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraTList.
func (in *TraTList) DeepCopy() *TraTList {
	if in == nil {
		return nil
	}
	out := new(TraTList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TraTList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraTSpec) DeepCopyInto(out *TraTSpec) {
	*out = *in
	if in.AzdMapping != nil {
		in, out := &in.AzdMapping, &out.AzdMapping
		*out = make(map[string]AzdField, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]ServiceSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AccessEvaluation != nil {
		in, out := &in.AccessEvaluation, &out.AccessEvaluation
		*out = (*in).DeepCopy()
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraTSpec.
func (in *TraTSpec) DeepCopy() *TraTSpec {
	if in == nil {
		return nil
	}
	out := new(TraTSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraTStatus) DeepCopyInto(out *TraTStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraTStatus.
func (in *TraTStatus) DeepCopy() *TraTStatus {
	if in == nil {
		return nil
	}
	out := new(TraTStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TraTVerificationRule) DeepCopyInto(out *TraTVerificationRule) {
	*out = *in
	if in.AzdMapping != nil {
		in, out := &in.AzdMapping, &out.AzdMapping
		*out = make(AzdMapping, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TraTVerificationRule.
func (in *TraTVerificationRule) DeepCopy() *TraTVerificationRule {
	if in == nil {
		return nil
	}
	out := new(TraTVerificationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TratteriaConfig) DeepCopyInto(out *TratteriaConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TratteriaConfig.
func (in *TratteriaConfig) DeepCopy() *TratteriaConfig {
	if in == nil {
		return nil
	}
	out := new(TratteriaConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TratteriaConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TratteriaConfigGenerationRule) DeepCopyInto(out *TratteriaConfigGenerationRule) {
	*out = *in
	out.Token = in.Token
	in.SubjectTokens.DeepCopyInto(&out.SubjectTokens)
	out.AccessEvaluationAPI = in.AccessEvaluationAPI
	if in.TokenGenerationAuthorizedServiceIds != nil {
		in, out := &in.TokenGenerationAuthorizedServiceIds, &out.TokenGenerationAuthorizedServiceIds
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TratteriaConfigGenerationRule.
func (in *TratteriaConfigGenerationRule) DeepCopy() *TratteriaConfigGenerationRule {
	if in == nil {
		return nil
	}
	out := new(TratteriaConfigGenerationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TratteriaConfigList) DeepCopyInto(out *TratteriaConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TratteriaConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TratteriaConfigList.
func (in *TratteriaConfigList) DeepCopy() *TratteriaConfigList {
	if in == nil {
		return nil
	}
	out := new(TratteriaConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TratteriaConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TratteriaConfigSpec) DeepCopyInto(out *TratteriaConfigSpec) {
	*out = *in
	out.Token = in.Token
	in.SubjectTokens.DeepCopyInto(&out.SubjectTokens)
	out.AccessEvaluationAPI = in.AccessEvaluationAPI
	if in.TokenGenerationAuthorizedServiceIds != nil {
		in, out := &in.TokenGenerationAuthorizedServiceIds, &out.TokenGenerationAuthorizedServiceIds
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TratteriaConfigSpec.
func (in *TratteriaConfigSpec) DeepCopy() *TratteriaConfigSpec {
	if in == nil {
		return nil
	}
	out := new(TratteriaConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TratteriaConfigStatus) DeepCopyInto(out *TratteriaConfigStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TratteriaConfigStatus.
func (in *TratteriaConfigStatus) DeepCopy() *TratteriaConfigStatus {
	if in == nil {
		return nil
	}
	out := new(TratteriaConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TratteriaConfigVerificationRule) DeepCopyInto(out *TratteriaConfigVerificationRule) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TratteriaConfigVerificationRule.
func (in *TratteriaConfigVerificationRule) DeepCopy() *TratteriaConfigVerificationRule {
	if in == nil {
		return nil
	}
	out := new(TratteriaConfigVerificationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VerificationRules) DeepCopyInto(out *VerificationRules) {
	*out = *in
	if in.TratteriaConfigVerificationRule != nil {
		in, out := &in.TratteriaConfigVerificationRule, &out.TratteriaConfigVerificationRule
		*out = new(TratteriaConfigVerificationRule)
		**out = **in
	}
	if in.TraTsVerificationRules != nil {
		in, out := &in.TraTsVerificationRules, &out.TraTsVerificationRules
		*out = make(map[string]*ServiceTraTVerificationRules, len(*in))
		for key, val := range *in {
			var outVal *ServiceTraTVerificationRules
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(ServiceTraTVerificationRules)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.TraTExclRule != nil {
		in, out := &in.TraTExclRule, &out.TraTExclRule
		*out = new(TraTExclRule)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VerificationRules.
func (in *VerificationRules) DeepCopy() *VerificationRules {
	if in == nil {
		return nil
	}
	out := new(VerificationRules)
	in.DeepCopyInto(out)
	return out
}
