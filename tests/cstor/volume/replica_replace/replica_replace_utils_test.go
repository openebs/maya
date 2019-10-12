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

package replicareplace

import (
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/cstor"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func deleteVolumeResources() {
	ops.DeletePersistentVolumeClaim(pvcObj.Name, pvcObj.Namespace)
	ops.VerifyVolumeResources(pvcObj.Spec.VolumeName, openebsNamespace)
	err := ops.SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	Expect(err).To(BeNil())
}

func deletePoolResources() {
	ops.DeleteStoragePoolClaim(spcObj.Name)
}

func verifyDesiredCSPCount() {
	cspCount := ops.GetHealthyCSPCount(spcObj.Name, cstor.PoolCount)
	Expect(cspCount).To(Equal(cstor.PoolCount))

	// Check are there any extra csps
	cspCount = ops.GetCSPCount(getLabelSelector(spcObj))
	Expect(cspCount).To(Equal(cstor.PoolCount), "Mismatch Of CSP Count")
}

func verifyVolumeStatus() {
	var err error
	status := ops.IsPVCBoundEventually(pvcObj.Name)
	Expect(status).To(Equal(true), "while checking status equal to bound")

	// GetLatest PVC object
	pvcObj, err = ops.PVCClient.
		WithNamespace(nsObj.Name).
		Get(pvcObj.Name, metav1.GetOptions{})
	Expect(err).To(BeNil())

	volumeLabel := pvLabel + pvcObj.Spec.VolumeName
	cvCount := ops.GetCstorVolumeCount(openebsNamespace, volumeLabel, 1)
	Expect(cvCount).To(Equal(1), "while checking cstorvolume count")

	isExpectedCVRCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, volumeLabel, ReplicaCount)
	Expect(isExpectedCVRCount).To(Equal(true), "while checking cstorvolume replica count")
}

func verifyVolumeConfigurationEventually() {
	var err error
	consistencyFactor := (ReplicaCount / 2) + 1
	for i := 0; i < MaxRetry; i++ {
		cvObj, err = ops.CVClient.WithNamespace(openebsNamespace).
			Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
		Expect(err).To(BeNil())
		if cvObj.Spec.ReplicationFactor == ReplicaCount {
			break
		}
		time.Sleep(5 * time.Second)
	}
	Expect(
		cvObj.Spec.ConsistencyFactor).To(Equal(consistencyFactor),
		"mismatch of consistencyFactor",
	)
	Expect(
		len(cvObj.Status.ReplicaDetails.KnownReplicas)).To(Equal(ReplicaCount),
		"mismatch of known replica list",
	)
	Expect(cvObj.Status.Phase).To(Equal(apis.CStorVolumePhase("Healthy")))
}

// This function is local to this package
func getLabelSelector(spc *apis.StoragePoolClaim) string {
	return string(apis.StoragePoolClaimCPK) + "=" + spc.Name
}

func buildAndCreateSC() {
	ReplicaCount = cstor.ReplicaCount
	FailureReplicaCount = (cstor.ReplicaCount / 2) + 1
	casConfig := strings.Replace(
		openebsCASConfigValue, "$spcName", spcObj.Name, 1)
	casConfig = strings.Replace(
		casConfig, "$count", strconv.Itoa(ReplicaCount), 1)
	annotations[string(apis.CASTypeKey)] = string(apis.CstorVolume)
	annotations[string(apis.CASConfigKey)] = casConfig
	scConfig := &tests.SCConfig{
		Name:        scName,
		Annotations: annotations,
		Provisioner: openebsProvisioner,
	}
	ops.Config = scConfig
	scObj = ops.CreateStorageClass()
}

func deleteZFSDataSets() {
	volumeLabel := pvLabel + "=" + pvcObj.Spec.VolumeName
	cvrObjList, err := ops.CVRClient.
		WithNamespace(nsObj.Namespace).
		List(metav1.ListOptions{LabelSelector: volumeLabel})
	Expect(err).To(BeNil())
	//TODO: Need to uncomment
	//Expect(len(cvrObjList.Items)).To(BeNumerically(">", FailureReplicaCount),
	//	"Available replica count should be greater than replicas targeted for test",
	//)

	targetCVRObjList = &apis.CStorVolumeReplicaList{}
	//pick FailureReplicaCount number of cvr's
	for i := 0; i < FailureReplicaCount; i++ {
		targetCVRObjList.Items = append(targetCVRObjList.Items, cvrObjList.Items[i])
	}

	poolPodObjList = &corev1.PodList{}
	for _, cvrObj := range targetCVRObjList.Items {
		poolLabel := cspLabel + "=" + cvrObj.GetLabels()[CstorPoolNameLabel]
		podObjList, err := ops.PodClient.
			WithNamespace(openebsNamespace).
			List(metav1.ListOptions{LabelSelector: poolLabel})
		Expect(err).To(BeNil())
		//One pool pod should present per pool
		Expect(len(podObjList.Items)).To(BeNumerically("==", 1))
		poolUID := cvrObj.GetLabels()[CstorPoolUIDLabel]
		volumeDataSet := PoolPrefix + poolUID + "/" + pvcObj.Spec.VolumeName
		cmd := "zfs destroy " + volumeDataSet
		_ = ops.ExecuteCMDEventually(&podObjList.Items[0], CstorPoolMgmtContainer, cmd, false)
		poolPodObjList.Items = append(poolPodObjList.Items, podObjList.Items[0])
	}

	// Make Sure data sets are deleted ( we can get it by checking the phase of
	// cvr)
	isExpectedCVRCount := ops.GetCstorVolumeReplicaCountEventually(openebsNamespace, volumeLabel, ReplicaCount-FailureReplicaCount)
	Expect(isExpectedCVRCount).To(Equal(true), "while checking cstorvolume replica count after deleting volume datasets")
}

func updateCVConfigurationsAndVerifyStatus() {
	volumeLabel := pvLabel + pvcObj.Spec.VolumeName
	cvObjList, err := ops.CVClient.WithNamespace(openebsNamespace).
		List(metav1.ListOptions{LabelSelector: volumeLabel})
	Expect(err).To(BeNil())
	cvObj = &cvObjList.Items[0]
	cvObj.Spec.ReplicationFactor = ReplicaCount - FailureReplicaCount
	cvObj.Spec.ConsistencyFactor = (cvObj.Spec.ReplicationFactor / 2) + 1
	knownReplicaDetails := getAvailableKnownReplicaDetails(cvObj)
	cvObj.Status.ReplicaDetails = knownReplicaDetails

	// Namespace is already set to CVClient in above step
	cvObj, err = ops.CVClient.Update(cvObj)
	Expect(err).To(BeNil())

	targetPodObjList, err := ops.PodClient.
		WithNamespace(openebsNamespace).
		List(metav1.ListOptions{LabelSelector: volumeLabel})
	Expect(err).To(BeNil())
	err = ops.RestartPodEventually(&targetPodObjList.Items[0])
	Expect(err).To(BeNil())

	cvCount := ops.GetCstorVolumeCount(openebsNamespace, volumeLabel, 1)
	Expect(cvCount).To(Equal(1), "while checking cstorvolume count after updating cstorvolume configurations")
}

func getAvailableKnownReplicaDetails(cvObj *apis.CStorVolume) apis.CStorVolumeReplicaDetails {
	cvKnownReplicaDetails := apis.CStorVolumeReplicaDetails{
		KnownReplicas: map[string]string{},
	}
	failedReplicaIDs := map[string]bool{}
	for _, cvrObj := range targetCVRObjList.Items {
		failedReplicaIDs[cvrObj.Spec.ReplicaID] = true
	}
	for replicaID, zvolGUID := range cvObj.Status.ReplicaDetails.KnownReplicas {
		if !failedReplicaIDs[replicaID] {
			cvKnownReplicaDetails.KnownReplicas[replicaID] = zvolGUID
		}
	}
	return cvKnownReplicaDetails
}

//TODO: Restart container instead of pod

func restartPoolPods() {
	for _, podObj := range poolPodObjList.Items {
		podObj := podObj
		err := ops.RestartPodEventually(&podObj)
		Expect(err).To(BeNil())
	}
}