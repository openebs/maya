/*
Copyright 2017 The OpenEBS Authors

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

package spc

import (
	"fmt"
	"github.com/golang/glog"
	apisv1alpha1 "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/versioned"
	openebsScheme "github.com/openebs/maya/pkg/client/generated/clientset/versioned/scheme"
	informers "github.com/openebs/maya/pkg/client/generated/informers/externalversions"
	listers "github.com/openebs/maya/pkg/client/generated/listers/openebs.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "cspc-controller"

// Controller is the controller implementation for CSPC resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	cspcLister listers.CStorPoolClusterLister

	// cspcSynced is used for caches sync to get populated
	cspcSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// ControllerBuilder is the builder object for controller.
type ControllerBuilder struct {
	Controller *Controller
}

// NewControllerBuilder returns an empty instance of controller builder.
func NewControllerBuilder() *ControllerBuilder {
	return &ControllerBuilder{
		Controller: &Controller{},
	}
}

// withKubeClient fills kube client to controller object.
func (cb *ControllerBuilder) withKubeClient(ks kubernetes.Interface) *ControllerBuilder {
	cb.Controller.kubeclientset = ks
	return cb
}

// withOpenEBSClient fills openebs client to controller object.
func (cb *ControllerBuilder) withOpenEBSClient(cs clientset.Interface) *ControllerBuilder {
	cb.Controller.clientset = cs
	return cb
}

// withSpcLister fills cspc lister to controller object.
func (cb *ControllerBuilder) withSpcLister(sl informers.SharedInformerFactory) *ControllerBuilder {
	cspcInformer := sl.Openebs().V1alpha1().CStorPoolClusters()
	cb.Controller.cspcLister = cspcInformer.Lister()
	return cb
}

// withcspcSynced adds object sync information in cache to controller object.
func (cb *ControllerBuilder) withcspcSynced(sl informers.SharedInformerFactory) *ControllerBuilder {
	cspcInformer := sl.Openebs().V1alpha1().CStorPoolClusters()
	cb.Controller.cspcSynced = cspcInformer.Informer().HasSynced
	return cb
}

// withWorkqueue adds workqueue to controller object.
func (cb *ControllerBuilder) withWorkqueueRateLimiting() *ControllerBuilder {
	cb.Controller.workqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CSPC")
	return cb
}

// withRecorder adds recorder to controller object.
func (cb *ControllerBuilder) withRecorder(ks kubernetes.Interface) *ControllerBuilder {
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: ks.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	cb.Controller.recorder = recorder
	return cb
}

// withEventHandler adds event handlers controller object.
func (cb *ControllerBuilder) withEventHandler(cspcInformerFactory informers.SharedInformerFactory) *ControllerBuilder {
	cspcInformer := cspcInformerFactory.Openebs().V1alpha1().CStorPoolClusters()
	// Set up an event handler for when CSPC resources change
	cspcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cb.Controller.addSpc,
		UpdateFunc: cb.Controller.updateSpc,
		// This will enter the sync loop and no-op, because the cspc has been deleted from the store.
		DeleteFunc: cb.Controller.deleteSpc,
	})
	return cb
}

// Build returns a controller instance.
func (cb *ControllerBuilder) Build() (*Controller, error) {
	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	return cb.Controller, nil
}

// addSpc is the add event handler for cspc.
func (c *Controller) addSpc(obj interface{}) {
	cspc, ok := obj.(*apisv1alpha1.CStorPoolCluster)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get cspc object %#v", obj))
		return
	}
	glog.V(4).Infof("Queuing CSPC %s for add event", cspc.Name)
	c.enqueueSpc(cspc)
}

// updateSpc is the update event handler for cspc.
func (c *Controller) updateSpc(oldSpc, newSpc interface{}) {
	cspc, ok := newSpc.(*apisv1alpha1.CStorPoolCluster)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get cspc object %#v", newSpc))
		return
	}
	// Enqueue cspc only when there is a pending pool to be created.
	po := c.NewPoolOperation(cspc)
	if po.IsPoolCreationPending() {
		c.enqueueSpc(newSpc)
	}
}

// deleteSpc is the delete event handler for cspc.
func (c *Controller) deleteSpc(obj interface{}) {
	cspc, ok := obj.(*apisv1alpha1.CStorPoolCluster)
	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		cspc, ok = tombstone.Obj.(*apisv1alpha1.CStorPoolCluster)
		if !ok {
			runtime.HandleError(fmt.Errorf("Tombstone contained object that is not a storagepoolclaim %#v", obj))
			return
		}
	}
	glog.V(4).Infof("Deleting storagepoolclaim %s", cspc.Name)
	c.enqueueSpc(cspc)
}
