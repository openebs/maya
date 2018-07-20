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

// This file was automatically generated by informer-gen

package v1alpha1

import (
	openebs_io_v1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	versioned "github.com/openebs/maya/pkg/client/clientset/versioned"
	internalinterfaces "github.com/openebs/maya/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/openebs/maya/pkg/client/listers/openebs/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// CStorVolumeInformer provides access to a shared informer and lister for
// CStorVolumes.
type CStorVolumeInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.CStorVolumeLister
}

type cStorVolumeInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

// NewCStorVolumeInformer constructs a new informer for CStorVolume type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCStorVolumeInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return client.OpenebsV1alpha1().CStorVolumes(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.OpenebsV1alpha1().CStorVolumes(namespace).Watch(options)
			},
		},
		&openebs_io_v1alpha1.CStorVolume{},
		resyncPeriod,
		indexers,
	)
}

func defaultCStorVolumeInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewCStorVolumeInformer(client, v1.NamespaceAll, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

func (f *cStorVolumeInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&openebs_io_v1alpha1.CStorVolume{}, defaultCStorVolumeInformer)
}

func (f *cStorVolumeInformer) Lister() v1alpha1.CStorVolumeLister {
	return v1alpha1.NewCStorVolumeLister(f.Informer().GetIndexer())
}
