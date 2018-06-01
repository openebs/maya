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
package v1alpha1

import (
	v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	scheme "github.com/openebs/maya/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// StoragePoolsGetter has a method to return a StoragePoolInterface.
// A group's client should implement this interface.
type StoragePoolsGetter interface {
	StoragePools() StoragePoolInterface
}

// StoragePoolInterface has methods to work with StoragePool resources.
type StoragePoolInterface interface {
	Create(*v1alpha1.StoragePool) (*v1alpha1.StoragePool, error)
	Update(*v1alpha1.StoragePool) (*v1alpha1.StoragePool, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.StoragePool, error)
	List(opts v1.ListOptions) (*v1alpha1.StoragePoolList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.StoragePool, err error)
	StoragePoolExpansion
}

// storagePools implements StoragePoolInterface
type storagePools struct {
	client rest.Interface
}

// newStoragePools returns a StoragePools
func newStoragePools(c *OpenebsV1alpha1Client) *storagePools {
	return &storagePools{
		client: c.RESTClient(),
	}
}

// Get takes name of the storagePool, and returns the corresponding storagePool object, and an error if there is any.
func (c *storagePools) Get(name string, options v1.GetOptions) (result *v1alpha1.StoragePool, err error) {
	result = &v1alpha1.StoragePool{}
	err = c.client.Get().
		Resource("storagepools").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of StoragePools that match those selectors.
func (c *storagePools) List(opts v1.ListOptions) (result *v1alpha1.StoragePoolList, err error) {
	result = &v1alpha1.StoragePoolList{}
	err = c.client.Get().
		Resource("storagepools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested storagePools.
func (c *storagePools) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("storagepools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a storagePool and creates it.  Returns the server's representation of the storagePool, and an error, if there is any.
func (c *storagePools) Create(storagePool *v1alpha1.StoragePool) (result *v1alpha1.StoragePool, err error) {
	result = &v1alpha1.StoragePool{}
	err = c.client.Post().
		Resource("storagepools").
		Body(storagePool).
		Do().
		Into(result)
	return
}

// Update takes the representation of a storagePool and updates it. Returns the server's representation of the storagePool, and an error, if there is any.
func (c *storagePools) Update(storagePool *v1alpha1.StoragePool) (result *v1alpha1.StoragePool, err error) {
	result = &v1alpha1.StoragePool{}
	err = c.client.Put().
		Resource("storagepools").
		Name(storagePool.Name).
		Body(storagePool).
		Do().
		Into(result)
	return
}

// Delete takes name of the storagePool and deletes it. Returns an error if one occurs.
func (c *storagePools) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("storagepools").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *storagePools) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("storagepools").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched storagePool.
func (c *storagePools) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.StoragePool, err error) {
	result = &v1alpha1.StoragePool{}
	err = c.client.Patch(pt).
		Resource("storagepools").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
