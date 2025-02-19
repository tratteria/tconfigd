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

package fake

import (
	"context"

	v1alpha1 "github.com/tokenetes/tconfigd/tokenetescontroller/pkg/apis/tokenetes/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTraTExclusions implements TraTExclusionInterface
type FakeTraTExclusions struct {
	Fake *FakeTokenetesV1alpha1
	ns   string
}

var tratexclusionsResource = v1alpha1.SchemeGroupVersion.WithResource("tratexclusions")

var tratexclusionsKind = v1alpha1.SchemeGroupVersion.WithKind("TraTExclusion")

// Get takes name of the traTExclusion, and returns the corresponding traTExclusion object, and an error if there is any.
func (c *FakeTraTExclusions) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.TraTExclusion, err error) {
	emptyResult := &v1alpha1.TraTExclusion{}
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(tratexclusionsResource, c.ns, name), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.TraTExclusion), err
}

// List takes label and field selectors, and returns the list of TraTExclusions that match those selectors.
func (c *FakeTraTExclusions) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.TraTExclusionList, err error) {
	emptyResult := &v1alpha1.TraTExclusionList{}
	obj, err := c.Fake.
		Invokes(testing.NewListAction(tratexclusionsResource, tratexclusionsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.TraTExclusionList{ListMeta: obj.(*v1alpha1.TraTExclusionList).ListMeta}
	for _, item := range obj.(*v1alpha1.TraTExclusionList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested traTExclusions.
func (c *FakeTraTExclusions) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(tratexclusionsResource, c.ns, opts))

}

// Create takes the representation of a traTExclusion and creates it.  Returns the server's representation of the traTExclusion, and an error, if there is any.
func (c *FakeTraTExclusions) Create(ctx context.Context, traTExclusion *v1alpha1.TraTExclusion, opts v1.CreateOptions) (result *v1alpha1.TraTExclusion, err error) {
	emptyResult := &v1alpha1.TraTExclusion{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(tratexclusionsResource, c.ns, traTExclusion), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.TraTExclusion), err
}

// Update takes the representation of a traTExclusion and updates it. Returns the server's representation of the traTExclusion, and an error, if there is any.
func (c *FakeTraTExclusions) Update(ctx context.Context, traTExclusion *v1alpha1.TraTExclusion, opts v1.UpdateOptions) (result *v1alpha1.TraTExclusion, err error) {
	emptyResult := &v1alpha1.TraTExclusion{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(tratexclusionsResource, c.ns, traTExclusion), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.TraTExclusion), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeTraTExclusions) UpdateStatus(ctx context.Context, traTExclusion *v1alpha1.TraTExclusion, opts v1.UpdateOptions) (result *v1alpha1.TraTExclusion, err error) {
	emptyResult := &v1alpha1.TraTExclusion{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(tratexclusionsResource, "status", c.ns, traTExclusion), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.TraTExclusion), err
}

// Delete takes name of the traTExclusion and deletes it. Returns an error if one occurs.
func (c *FakeTraTExclusions) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(tratexclusionsResource, c.ns, name, opts), &v1alpha1.TraTExclusion{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTraTExclusions) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(tratexclusionsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.TraTExclusionList{})
	return err
}

// Patch applies the patch and returns the patched traTExclusion.
func (c *FakeTraTExclusions) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TraTExclusion, err error) {
	emptyResult := &v1alpha1.TraTExclusion{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(tratexclusionsResource, c.ns, name, pt, data, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.TraTExclusion), err
}
