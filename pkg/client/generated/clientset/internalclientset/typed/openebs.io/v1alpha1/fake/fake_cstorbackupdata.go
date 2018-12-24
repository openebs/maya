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

// Code generated by client-gen. DO NOT EDIT.

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

// FakeCStorBackupDatas implements CStorBackupDataInterface
type FakeCStorBackupDatas struct {
	Fake *FakeOpenebsV1alpha1
	ns   string
}

var cstorbackupdatasResource = schema.GroupVersionResource{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorbackupdatas"}

var cstorbackupdatasKind = schema.GroupVersionKind{Group: "openebs.io", Version: "v1alpha1", Kind: "CStorBackupData"}

// Get takes name of the cStorBackupData, and returns the corresponding cStorBackupData object, and an error if there is any.
func (c *FakeCStorBackupDatas) Get(name string, options v1.GetOptions) (result *v1alpha1.CStorBackupData, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cstorbackupdatasResource, c.ns, name), &v1alpha1.CStorBackupData{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorBackupData), err
}

// List takes label and field selectors, and returns the list of CStorBackupDatas that match those selectors.
func (c *FakeCStorBackupDatas) List(opts v1.ListOptions) (result *v1alpha1.CStorBackupDataList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cstorbackupdatasResource, cstorbackupdatasKind, c.ns, opts), &v1alpha1.CStorBackupDataList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CStorBackupDataList{ListMeta: obj.(*v1alpha1.CStorBackupDataList).ListMeta}
	for _, item := range obj.(*v1alpha1.CStorBackupDataList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cStorBackupDatas.
func (c *FakeCStorBackupDatas) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cstorbackupdatasResource, c.ns, opts))

}

// Create takes the representation of a cStorBackupData and creates it.  Returns the server's representation of the cStorBackupData, and an error, if there is any.
func (c *FakeCStorBackupDatas) Create(cStorBackupData *v1alpha1.CStorBackupData) (result *v1alpha1.CStorBackupData, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cstorbackupdatasResource, c.ns, cStorBackupData), &v1alpha1.CStorBackupData{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorBackupData), err
}

// Update takes the representation of a cStorBackupData and updates it. Returns the server's representation of the cStorBackupData, and an error, if there is any.
func (c *FakeCStorBackupDatas) Update(cStorBackupData *v1alpha1.CStorBackupData) (result *v1alpha1.CStorBackupData, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cstorbackupdatasResource, c.ns, cStorBackupData), &v1alpha1.CStorBackupData{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorBackupData), err
}

// Delete takes name of the cStorBackupData and deletes it. Returns an error if one occurs.
func (c *FakeCStorBackupDatas) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cstorbackupdatasResource, c.ns, name), &v1alpha1.CStorBackupData{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCStorBackupDatas) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cstorbackupdatasResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CStorBackupDataList{})
	return err
}

// Patch applies the patch and returns the patched cStorBackupData.
func (c *FakeCStorBackupDatas) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CStorBackupData, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cstorbackupdatasResource, c.ns, name, data, subresources...), &v1alpha1.CStorBackupData{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CStorBackupData), err
}
