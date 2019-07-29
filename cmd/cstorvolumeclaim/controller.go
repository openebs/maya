/*
Copyright 2019 The OpenEBS Authors

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

package cstorvolumeclaim

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	merrors "github.com/openebs/maya/pkg/errors/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	ref "k8s.io/client-go/tools/reference"
	"k8s.io/kubernetes/pkg/util/slice"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a
	// cstorvolumeclaim is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a
	// cstorvolumeclaim fails to sync due to a cstorvolumeclaim of the same
	// name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a cstorvolumeclaim already existing
	MessageResourceExists = "Resource %q already exists and is not managed by CVC"
	// MessageResourceSynced is the message used for an Event fired when a
	// cstorvolumeclaim is synced successfully
	MessageResourceSynced = "cstorvolumeclaim synced successfully"
	// MessageResourceCreated ...
	MessageResourceCreated = "cstorvolumeclaim created successfully"
	// CStorVolumeClaimFinalizer name of finalizer on CStorVolumeClaim that
	// are bound by CStorVolume
	CStorVolumeClaimFinalizer = "cvc.openebs.io/finalizer"
)

var knownResizeConditions = map[apis.CStorVolumeClaimConditionType]bool{
	apis.CStorVolumeClaimResizing:      true,
	apis.CStorVolumeClaimResizePending: true,
}

// Patch struct represent the struct used to patch
// the cstorvolumeclaim object
type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the spcPoolUpdated resource
// with the current status of the resource.
func (c *CVCController) syncHandler(key string) error {
	startTime := time.Now()
	glog.V(4).Infof("Started syncing cstorvolumeclaim %q (%v)", key, startTime)
	defer func() {
		glog.V(4).Infof("Finished syncing cstorvolumeclaim %q (%v)", key, time.Since(startTime))
	}()

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the cvc resource with this namespace/name
	cvc, err := c.cvcLister.CStorVolumeClaims(namespace).Get(name)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(fmt.Errorf("cstorvolumeclaim '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}
	cvcCopy := cvc.DeepCopy()
	err = c.syncCVC(cvcCopy)
	return err
}

// enqueueCVC takes a CVC resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CStorVolumeClaims.
func (c *CVCController) enqueueCVC(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)

	/*	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
			obj = unknown.Obj
		}
		if cvc, ok := obj.(*apis.CStorVolumeClaim); ok {
			objName, err := cache.DeletionHandlingMetaNamespaceKeyFunc(cvc)
			if err != nil {
				glog.Errorf("failed to get key from object: %v, %v", err, cvc)
				return
			}
			glog.V(5).Infof("enqueued %q for sync", objName)
			c.workqueue.Add(objName)
		}
	*/
}

// synCVC is the function which tries to converge to a desired state for the
// CStorVolumeClaims
func (c *CVCController) syncCVC(cvc *apis.CStorVolumeClaim) error {
	// CStor Volume Claim should be deleted. Check if deletion timestamp is set
	// and remove finalizer.
	if c.isClaimDeletionCandidate(cvc) {
		glog.Infof("syncClaim: remove finalizer for CStorVolumeClaimVolume [%s]", cvc.Name)
		return c.removeClaimFinalizer(cvc)
	}

	volName := cvc.Name
	if volName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%+v: cvc name must be specified", cvc))
		return nil
	}

	//NodeID indicates where the volume is needed to be mounted, i.e the node
	// where the app has been scheduled.
	nodeID := cvc.Publish.NodeID
	if nodeID == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("cvc must be publish/attached to Node: %v", cvc))
		return nil
	}

	// Get the cstorvolume with the name specified, if not found create all the
	// required resources
	_, err := c.cvLister.CStorVolumes(cvc.Namespace).Get(volName)
	if k8serror.IsNotFound(err) {
		glog.Infof("create cstor based volume using cvc %+v", cvc)
		_, err = c.createVolumeOperation(cvc)
	}
	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// IsCVRPending checks for the pending cstorvolume replica and requeue the
	// create operation if count doesn't matches the desired count
	pending, err := c.IsCVRPending(cvc)
	if err != nil {
		return err
	}
	if pending {
		glog.Infof("create remaining volume replica %+v", cvc)
		_, err = c.createVolumeOperation(cvc)
		if err != nil {
			return err
		}
	}

	if c.cvcNeedResize(cvc) {
		err = c.resizeCVC(cvc)
		if err != nil {
			return err
		}
	}
	// Finally, we update the status block of the CVC resource to reflect the
	// current state of the world
	//c.recorder.Event(cvc, corev1.EventTypeNormal,
	//	SuccessSynced,
	//	MessageResourceSynced,
	//)
	return nil
}

// UpdateCVCObj updates the cstorvolumeclaim object resource to reflect the
// current state of the world
func (c *CVCController) updateCVCObj(
	cvc *apis.CStorVolumeClaim,
	cv *apis.CStorVolume,
) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	cvcCopy := cvc.DeepCopy()
	if cvc.Name != cv.Name {
		return fmt.
			Errorf("could not bind cstorvolumeclaim %s and cstorvolume %s, name does not match",
				cvc.Name,
				cv.Name)
	}

	_, err := c.clientset.OpenebsV1alpha1().CStorVolumeClaims(cvc.Namespace).Update(cvcCopy)

	c.recorder.Event(cvc, corev1.EventTypeNormal,
		SuccessSynced,
		MessageResourceCreated,
	)
	return err
}

// createVolumeOperation trigers the all required resource create operation.
// 1. Create volume service.
// 2. Create cstorvolume resource with required iscsi information.
// 3. Create target deployment.
// 4. Create cstorvolumeclaim resource.
// 5. Update the cstorvolumeclaim with claimRef info and bound with cstorvolume.
func (c *CVCController) createVolumeOperation(cvc *apis.CStorVolumeClaim) (*apis.CStorVolumeClaim, error) {

	_ = cvc.Annotations[string(apis.ConfigClassKey)]

	glog.V(2).Infof("creating cstorvolume service resource")
	svcObj, err := getOrCreateTargetService(cvc)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("creating cstorvolume resource")
	cvObj, err := getOrCreateCStorVolumeResource(svcObj, cvc)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("creating cstorvolume target deployment")
	_, err = getOrCreateCStorTargetDeployment(cvObj)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("creating cstorvolume replica resource")
	err = c.distributePendingCVRs(cvc, cvObj, svcObj)
	if err != nil {
		return nil, err
	}

	volumeRef, err := ref.GetReference(scheme.Scheme, cvObj)
	if err != nil {
		return nil, err
	}

	// update the cstorvolume reference, phase as "Bound" and requested
	// capacity
	cvc.Spec.CStorVolumeRef = volumeRef
	cvc.Status.Phase = apis.CStorVolumeClaimPhaseBound
	cvc.Status.Capacity = cvc.Spec.Capacity

	err = c.updateCVCObj(cvc, cvObj)
	if err != nil {
		return nil, err
	}
	return cvc, nil
}

// distributePendingCVRs trigers create and distribute pending cstorvolumereplica
// resource among the available cstor pools
func (c *CVCController) distributePendingCVRs(
	cvc *apis.CStorVolumeClaim,
	cv *apis.CStorVolume,
	service *corev1.Service,
) error {

	pendingReplicaCount, err := c.getPendingCVRCount(cvc)
	if err != nil {
		return err
	}
	err = distributeCVRs(pendingReplicaCount, cvc, service, cv)
	if err != nil {
		return err
	}
	return nil
}

// isClaimDeletionCandidate checks if a cstorvolumeclaim is a deletion candidate.
func (c *CVCController) isClaimDeletionCandidate(cvc *apis.CStorVolumeClaim) bool {
	return cvc.ObjectMeta.DeletionTimestamp != nil &&
		slice.ContainsString(cvc.ObjectMeta.Finalizers, CStorVolumeClaimFinalizer, nil)
}

// removeFinalizer removes finalizers present in CStorVolumeClaim resource
func (c *CVCController) removeClaimFinalizer(
	cvc *apis.CStorVolumeClaim,
) error {
	cvcPatch := []Patch{
		Patch{
			Op:   "remove",
			Path: "/metadata/finalizers",
		},
	}

	cvcPatchBytes, err := json.Marshal(cvcPatch)
	if err != nil {
		return merrors.Wrapf(
			err,
			"failed to remove finalizers from cstorvolumeclaim {%s}",
			cvc.Name,
		)
	}

	_, err = c.clientset.
		OpenebsV1alpha1().
		CStorVolumeClaims(cvc.Namespace).
		Patch(cvc.Name, types.JSONPatchType, cvcPatchBytes)
	if err != nil {
		return merrors.Wrapf(
			err,
			"failed to remove finalizers from cstorvolumeclaim {%s}",
			cvc.Name,
		)
	}
	glog.Infof("finalizers removed successfully from cstorvolumeclaim {%s}", cvc.Name)
	return nil
}

// getPendingCVRCount gets the pending replica count to be created
// in case of any failures
func (c *CVCController) getPendingCVRCount(
	cvc *apis.CStorVolumeClaim,
) (int, error) {

	currentReplicaCount, err := c.getCurrentReplicaCount(cvc)
	if err != nil {
		runtime.HandleError(err)
		return 0, err
	}
	return cvc.Spec.ReplicaCount - currentReplicaCount, nil
}

// getCurrentReplicaCount give the current cstorvolumereplicas count for the
// given volume.
func (c *CVCController) getCurrentReplicaCount(cvc *apis.CStorVolumeClaim) (int, error) {
	// TODO use lister
	//	CVRs, err := c.cvrLister.CStorVolumeReplicas(cvc.Namespace).
	//		List(klabels.Set(pvLabel).AsSelector())

	pvLabel := pvSelector + "=" + cvc.Name

	cvrList, err := c.clientset.
		OpenebsV1alpha1().
		CStorVolumeReplicas(cvc.Namespace).
		List(metav1.ListOptions{LabelSelector: pvLabel})

	if err != nil {
		return 0, merrors.Errorf("unable to get current replica count: %v", err)
	}
	return len(cvrList.Items), nil
}

// IsCVRPending look for pending cstorvolume replicas compared to desired
// replica count. returns true if count doesn't matches.
func (c *CVCController) IsCVRPending(cvc *apis.CStorVolumeClaim) (bool, error) {

	selector := klabels.SelectorFromSet(BaseLabels(cvc))
	CVRs, err := c.cvrLister.CStorVolumeReplicas(cvc.Namespace).
		List(selector)
	if err != nil {
		return false, merrors.Errorf("failed to list cvr : %v", err)
	}
	// TODO: check for greater values
	return cvc.Spec.ReplicaCount != len(CVRs), nil
}

// BaseLabels returns the base labels we apply to cstorvolumereplicas created
func BaseLabels(cvc *apis.CStorVolumeClaim) map[string]string {
	base := map[string]string{
		pvSelector: cvc.Name,
	}
	return base
}

// cvcNeedResize returns true is a cvc requests a resize operation.
func (c *CVCController) cvcNeedResize(cvc *apis.CStorVolumeClaim) bool {
	// only bound cstorvolumeclaim can be resized
	if cvc.Status.Phase != apis.CStorVolumeClaimPhaseBound {
		return false
	}

	cv, err := c.clientset.OpenebsV1alpha1().CStorVolumes(cvc.Namespace).
		Get(cvc.Name, metav1.GetOptions{})
	if err != nil {
		runtime.HandleError(fmt.Errorf("falied to get cv %s: %v", cvc.Name, err))
		return false
	}

	requestCVCSize := cvc.Spec.Capacity[v1.ResourceStorage]
	actualCVCSize := cvc.Status.Capacity[v1.ResourceStorage]
	if requestCVCSize.Cmp(actualCVCSize) > 0 {
		return true
	}

	actualCVSize := cv.Spec.Capacity
	return requestCVCSize.Cmp(actualCVSize) > 0
}

// resizeCVC will:
// 1. Mark cvc as resizing.
// 2. Resize the cstorvolume object.
// 3. Mark cvc as resizing finished
func (c *CVCController) resizeCVC(cvc *apis.CStorVolumeClaim) error {
	if updatedCVC, err := c.markCVCResizeInProgress(cvc); err != nil {
		glog.Errorf("mark cvc %q as resizing failed: %v", cvc.Name, err)
		return err
	} else if updatedCVC != nil {
		cvc = updatedCVC
	}

	// Record an event to indicate that cvc-controller is resizing this volume.
	c.recorder.Event(cvc, corev1.EventTypeNormal, string(apis.CStorVolumeClaimResizing),
		fmt.Sprintf("CVCController is resizing volume %s", cvc.Name))

	cv, err := c.clientset.OpenebsV1alpha1().CStorVolumes(cvc.Namespace).
		Get(cvc.Name, metav1.GetOptions{})
	if err != nil {
		runtime.HandleError(fmt.Errorf("falied to get cv %s: %v", cvc.Name, err))
		return err
	}

	err = func() error {
		err = c.resizeVolume(cvc, cv)
		if err != nil {
			return err
		}
		// Resize volume succeeded mark it as resizing finished.
		return c.markCVCResizeFinished(cvc)
	}()

	if err != nil {
		// Record an event to indicate that resize operation is failed.
		c.recorder.Eventf(cvc, corev1.EventTypeWarning, string(apis.CStorVolumeClaimResizeFailed), err.Error())
	}

	return err
}

func (c *CVCController) markCVCResizeInProgress(cvc *apis.CStorVolumeClaim) (*apis.CStorVolumeClaim, error) {
	// Mark CVC as Resize Started
	progressCondition := apis.CStorVolumeClaimCondition{
		Type:               apis.CStorVolumeClaimResizing,
		LastTransitionTime: metav1.Now(),
	}
	newCVC := cvc.DeepCopy()
	newCVC.Status.Conditions = c.MergeResizeConditionsOfCVC(newCVC.Status.Conditions,
		[]apis.CStorVolumeClaimCondition{progressCondition})
	return c.PatchCVCStatus(cvc, newCVC)
}

// MergeResizeConditionsOfCVC updates cvc with requested resize conditions
// leaving other conditions untouched.
func (c *CVCController) MergeResizeConditionsOfCVC(oldConditions, newConditions []apis.CStorVolumeClaimCondition) []apis.CStorVolumeClaimCondition {
	newConditionSet := make(map[apis.CStorVolumeClaimConditionType]apis.CStorVolumeClaimCondition, len(newConditions))
	for _, condition := range newConditions {
		newConditionSet[condition.Type] = condition
	}

	var resultConditions []apis.CStorVolumeClaimCondition
	for _, condition := range oldConditions {
		// If Condition is of not resize type, we keep and append it
		if _, ok := knownResizeConditions[condition.Type]; !ok {
			newConditions = append(newConditions, condition)
			continue
		}
		if newCondition, ok := newConditionSet[condition.Type]; ok {
			// Use the new condition to replace old condition with same type
			resultConditions = append(resultConditions, newCondition)
			delete(newConditionSet, condition.Type)
		}
	}

	// Append remains resize conditions
	for _, condition := range newConditionSet {
		resultConditions = append(resultConditions, condition)
	}

	return resultConditions
}

func (c *CVCController) markCVCResizeFinished(cvc *apis.CStorVolumeClaim) error {
	newCVC := cvc.DeepCopy()
	newCVC.Status.Capacity = cvc.Spec.Capacity

	newCVC.Status.Conditions = c.MergeResizeConditionsOfCVC(cvc.Status.Conditions, []apis.CStorVolumeClaimCondition{})
	_, err := c.PatchCVCStatus(cvc, newCVC)
	if err != nil {
		glog.Errorf("Mark CVC %q as resize finished failed: %v", cvc.Name, err)
		return err
	}

	glog.V(4).Infof("Resize CVC %q finished", cvc.Name)
	c.recorder.Eventf(cvc, corev1.EventTypeNormal, string(apis.CStorVolumeClaimResizeSuccess), "Resize volume succeeded")

	return nil
}

// PatchCVCStatus updates CVC status using patch api
func (c *CVCController) PatchCVCStatus(oldCVC,
	newCVC *apis.CStorVolumeClaim,
) (*apis.CStorVolumeClaim, error) {
	patchBytes, err := getPatchData(oldCVC, newCVC)
	if err != nil {
		return nil, fmt.Errorf("can't patch status of CVC %s as generate path data failed: %v", oldCVC.Name, err)
	}
	updatedClaim, updateErr := c.clientset.OpenebsV1alpha1().CStorVolumeClaims(oldCVC.Namespace).
		Patch(oldCVC.Name, types.MergePatchType, patchBytes)

	if updateErr != nil {
		return nil, fmt.Errorf("can't patch status of CVC %s with %v", oldCVC.Name, updateErr)
	}
	return updatedClaim, nil
}

func getPatchData(oldObj, newObj interface{}) ([]byte, error) {
	oldData, err := json.Marshal(oldObj)
	if err != nil {
		return nil, fmt.Errorf("marshal old object failed: %v", err)
	}
	newData, err := json.Marshal(newObj)
	if err != nil {
		return nil, fmt.Errorf("mashal new object failed: %v", err)
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, oldObj)
	if err != nil {
		return nil, fmt.Errorf("CreateTwoWayMergePatch failed: %v", err)
	}
	return patchBytes, nil
}

// resizeVolume resize the volume to request size, and update CV's capacity
func (c *CVCController) resizeVolume(
	cvc *apis.CStorVolumeClaim,
	cv *apis.CStorVolume) error {

	newSize := cvc.Spec.Capacity[corev1.ResourceStorage]
	err := c.UpdateCVCapacity(cv, newSize)
	if err != nil {
		glog.Errorf("Update capacity of CV %q to %s failed: %v", cv.Name, newSize.String(), err)
		return err
	}
	glog.V(4).Infof("Update capacity of CV %q to %s succeeded", cv.Name, newSize.String())

	return err

}

// UpdateCVCapacity updates CV capacity with requested size.
func (c *CVCController) UpdateCVCapacity(cv *apis.CStorVolume, newCapacity resource.Quantity) error {
	progressCondition := apis.CStorVolumeCondition{
		Type:               apis.CStorVolumeResizing,
		LastTransitionTime: metav1.Now(),
		Status:             apis.ConditionInProgress,
	}

	newCV := cv.DeepCopy()
	newCV.Status.Conditions = append(newCV.Status.Conditions, progressCondition)

	newCV.Spec.Capacity = newCapacity
	patchBytes, err := getPatchData(cv, newCV)
	if err != nil {
		return fmt.Errorf("can't update capacity of CV %s as generate patch data failed: %v", cv.Name, err)
	}
	_, updateErr := c.clientset.OpenebsV1alpha1().CStorVolumes(getNamespace()).
		Patch(cv.Name, types.MergePatchType, patchBytes)
	if updateErr != nil {
		return updateErr
	}
	return nil
}

// HasResizePendingCondition returns true if a cv has a ResizePending condition.
// This means the controller side resize operation is finished,
// and volume side operation is in progress.
func HasResizePendingCondition(cv *apis.CStorVolume) bool {
	for _, condition := range cv.Status.Conditions {
		if condition.Type == apis.CStorVolumeResizing && condition.Status == apis.ConditionInProgress {
			return true
		}
	}
	return false
}
