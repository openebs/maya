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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CStorBackupDataLister helps list CStorBackupDatas.
type CStorBackupDataLister interface {
	// List lists all CStorBackupDatas in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.CStorBackupData, err error)
	// CStorBackupDatas returns an object that can list and get CStorBackupDatas.
	CStorBackupDatas(namespace string) CStorBackupDataNamespaceLister
	CStorBackupDataListerExpansion
}

// cStorBackupDataLister implements the CStorBackupDataLister interface.
type cStorBackupDataLister struct {
	indexer cache.Indexer
}

// NewCStorBackupDataLister returns a new CStorBackupDataLister.
func NewCStorBackupDataLister(indexer cache.Indexer) CStorBackupDataLister {
	return &cStorBackupDataLister{indexer: indexer}
}

// List lists all CStorBackupDatas in the indexer.
func (s *cStorBackupDataLister) List(selector labels.Selector) (ret []*v1alpha1.CStorBackupData, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CStorBackupData))
	})
	return ret, err
}

// CStorBackupDatas returns an object that can list and get CStorBackupDatas.
func (s *cStorBackupDataLister) CStorBackupDatas(namespace string) CStorBackupDataNamespaceLister {
	return cStorBackupDataNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// CStorBackupDataNamespaceLister helps list and get CStorBackupDatas.
type CStorBackupDataNamespaceLister interface {
	// List lists all CStorBackupDatas in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.CStorBackupData, err error)
	// Get retrieves the CStorBackupData from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.CStorBackupData, error)
	CStorBackupDataNamespaceListerExpansion
}

// cStorBackupDataNamespaceLister implements the CStorBackupDataNamespaceLister
// interface.
type cStorBackupDataNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all CStorBackupDatas in the indexer for a given namespace.
func (s cStorBackupDataNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.CStorBackupData, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CStorBackupData))
	})
	return ret, err
}

// Get retrieves the CStorBackupData from the indexer for a given namespace and name.
func (s cStorBackupDataNamespaceLister) Get(name string) (*v1alpha1.CStorBackupData, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("cstorbackupdata"), name)
	}
	return obj.(*v1alpha1.CStorBackupData), nil
}
