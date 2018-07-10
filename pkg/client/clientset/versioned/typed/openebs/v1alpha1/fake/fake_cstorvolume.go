/*
Copyright 2018 The OpenEBS Authors

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
package fake

import (
	v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCStorVolumes implements CStorVolumeInterface
type FakeCStorVolumes struct {
	Fake *FakeOpenebsV1alpha1
}

var cstorvolumesResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorvolumes"}

var cstorvolumesKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "CStorVolume"}

// Get takes name of the cStorVolume, and returns the corresponding cStorVolume object, and an error if there is any.
func (c *FakeCStorVolumes) Get(name string, options v1.GetOptions) (result *v1alpha1.CStorVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(cstorvolumesResource, name), &v1alpha1.CStorVolume{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorVolume), err
}

// List takes label and field selectors, and returns the list of CStorVolumes that match those selectors.
func (c *FakeCStorVolumes) List(opts v1.ListOptions) (result *v1alpha1.CStorVolumeList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(cstorvolumesResource, cstorvolumesKind, opts), &v1alpha1.CStorVolumeList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CStorVolumeList{}
	for _, item := range obj.(*v1alpha1.CStorVolumeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cStorVolumes.
func (c *FakeCStorVolumes) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(cstorvolumesResource, opts))
}

// Create takes the representation of a cStorVolume and creates it.  Returns the server's representation of the cStorVolume, and an error, if there is any.
func (c *FakeCStorVolumes) Create(cStorVolume *v1alpha1.CStorVolume) (result *v1alpha1.CStorVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(cstorvolumesResource, cStorVolume), &v1alpha1.CStorVolume{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorVolume), err
}

// Update takes the representation of a cStorVolume and updates it. Returns the server's representation of the cStorVolume, and an error, if there is any.
func (c *FakeCStorVolumes) Update(cStorVolume *v1alpha1.CStorVolume) (result *v1alpha1.CStorVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(cstorvolumesResource, cStorVolume), &v1alpha1.CStorVolume{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorVolume), err
}

// Delete takes name of the cStorVolume and deletes it. Returns an error if one occurs.
func (c *FakeCStorVolumes) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(cstorvolumesResource, name), &v1alpha1.CStorVolume{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCStorVolumes) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(cstorvolumesResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CStorVolumeList{})
	return err
}

// Patch applies the patch and returns the patched cStorVolume.
func (c *FakeCStorVolumes) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CStorVolume, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(cstorvolumesResource, name, data, subresources...), &v1alpha1.CStorVolume{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorVolume), err
}
