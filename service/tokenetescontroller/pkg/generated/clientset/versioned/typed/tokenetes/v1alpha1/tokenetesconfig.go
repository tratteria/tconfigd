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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"

	v1alpha1 "github.com/tokenetes/tconfigd/tokenetescontroller/pkg/apis/tokenetes/v1alpha1"
	scheme "github.com/tokenetes/tconfigd/tokenetescontroller/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// TokenetesConfigsGetter has a method to return a TokenetesConfigInterface.
// A group's client should implement this interface.
type TokenetesConfigsGetter interface {
	TokenetesConfigs(namespace string) TokenetesConfigInterface
}

// TokenetesConfigInterface has methods to work with TokenetesConfig resources.
type TokenetesConfigInterface interface {
	Create(ctx context.Context, tokenetesConfig *v1alpha1.TokenetesConfig, opts v1.CreateOptions) (*v1alpha1.TokenetesConfig, error)
	Update(ctx context.Context, tokenetesConfig *v1alpha1.TokenetesConfig, opts v1.UpdateOptions) (*v1alpha1.TokenetesConfig, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, tokenetesConfig *v1alpha1.TokenetesConfig, opts v1.UpdateOptions) (*v1alpha1.TokenetesConfig, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.TokenetesConfig, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.TokenetesConfigList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TokenetesConfig, err error)
	TokenetesConfigExpansion
}

// tokenetesConfigs implements TokenetesConfigInterface
type tokenetesConfigs struct {
	*gentype.ClientWithList[*v1alpha1.TokenetesConfig, *v1alpha1.TokenetesConfigList]
}

// newTokenetesConfigs returns a TokenetesConfigs
func newTokenetesConfigs(c *TokenetesV1alpha1Client, namespace string) *tokenetesConfigs {
	return &tokenetesConfigs{
		gentype.NewClientWithList[*v1alpha1.TokenetesConfig, *v1alpha1.TokenetesConfigList](
			"tokenetesconfigs",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha1.TokenetesConfig { return &v1alpha1.TokenetesConfig{} },
			func() *v1alpha1.TokenetesConfigList { return &v1alpha1.TokenetesConfigList{} }),
	}
}
