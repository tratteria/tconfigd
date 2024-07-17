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

	v1alpha1 "github.com/tratteria/tconfigd/tratteriacontroller/pkg/apis/tratteria/v1alpha1"
	scheme "github.com/tratteria/tconfigd/tratteriacontroller/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// TratteriaConfigsGetter has a method to return a TratteriaConfigInterface.
// A group's client should implement this interface.
type TratteriaConfigsGetter interface {
	TratteriaConfigs(namespace string) TratteriaConfigInterface
}

// TratteriaConfigInterface has methods to work with TratteriaConfig resources.
type TratteriaConfigInterface interface {
	Create(ctx context.Context, tratteriaConfig *v1alpha1.TratteriaConfig, opts v1.CreateOptions) (*v1alpha1.TratteriaConfig, error)
	Update(ctx context.Context, tratteriaConfig *v1alpha1.TratteriaConfig, opts v1.UpdateOptions) (*v1alpha1.TratteriaConfig, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, tratteriaConfig *v1alpha1.TratteriaConfig, opts v1.UpdateOptions) (*v1alpha1.TratteriaConfig, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.TratteriaConfig, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.TratteriaConfigList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TratteriaConfig, err error)
	TratteriaConfigExpansion
}

// tratteriaConfigs implements TratteriaConfigInterface
type tratteriaConfigs struct {
	*gentype.ClientWithList[*v1alpha1.TratteriaConfig, *v1alpha1.TratteriaConfigList]
}

// newTratteriaConfigs returns a TratteriaConfigs
func newTratteriaConfigs(c *TratteriaV1alpha1Client, namespace string) *tratteriaConfigs {
	return &tratteriaConfigs{
		gentype.NewClientWithList[*v1alpha1.TratteriaConfig, *v1alpha1.TratteriaConfigList](
			"tratteriaconfigs",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha1.TratteriaConfig { return &v1alpha1.TratteriaConfig{} },
			func() *v1alpha1.TratteriaConfigList { return &v1alpha1.TratteriaConfigList{} }),
	}
}