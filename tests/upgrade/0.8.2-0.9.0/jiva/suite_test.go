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

package jiva

import (
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	"github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	k8s "github.com/openebs/maya/pkg/client/k8s/v1alpha1"
	pvc "github.com/openebs/maya/pkg/kubernetes/persistentvolumeclaim/v1alpha1"
	sc "github.com/openebs/maya/pkg/kubernetes/storageclass/v1alpha1"
	"github.com/openebs/maya/tests"
	"github.com/openebs/maya/tests/artifacts"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	kubeConfigPath        string
	replicaCount          int
	nsName                = "default"
	scName                = "jiva-upgrade-sc"
	openebsProvisioner    = "openebs.io/provisioner-iscsi"
	replicaLabel          = "openebs.io/replica=jiva-replica"
	ctrlLabel             = "openebs.io/controller=jiva-controller"
	openebsCASConfigValue = "- name: ReplicaCount\n  Value: "
	accessModes           = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity              = "5G"
	pvcObj                *corev1.PersistentVolumeClaim
	pvcName               = "jiva-volume-claim"
	scObj                 *storagev1.StorageClass
	openebsArtifact,
	rbacArtifact,
	crArtifact,
	runtaskArtifact,
	jobArtifact artifacts.Artifact
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test jiva volume upgrade")
}

func init() {
	flag.StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig to invoke kubernetes API calls")
	flag.IntVar(&replicaCount, "replicas", 1, "number of replicas to be created")
}

var ops *tests.Operations

var _ = BeforeSuite(func() {

	openebsArtifact = getArtifactFromURL("https://openebs.github.io/charts/openebs-operator-0.8.2.yaml")
	rbacArtifact = getArtifactFromURL("https://raw.githubusercontent.com/openebs/openebs/master/k8s/upgrades/0.8.2-0.9.0/rbac.yaml")
	crArtifact = getArtifactFromURL("https://raw.githubusercontent.com/openebs/openebs/master/k8s/upgrades/0.8.2-0.9.0/jiva/cr.yaml")
	runtaskArtifact = getArtifactFromURL("https://raw.githubusercontent.com/openebs/openebs/master/k8s/upgrades/0.8.2-0.9.0/jiva/jiva_upgrade_runtask.yaml")
	jobArtifact = getArtifactFromURL("https://raw.githubusercontent.com/openebs/openebs/master/k8s/upgrades/0.8.2-0.9.0/jiva/volume-upgrade-job.yaml")

	ops = tests.NewOperations(tests.WithKubeConfigPath(kubeConfigPath))
	openebsCASConfigValue = openebsCASConfigValue + strconv.Itoa(replicaCount)

	// Setting the path in environemnt variable
	err := os.Setenv(string(v1alpha1.KubeConfigEnvironmentKey), kubeConfigPath)
	Expect(err).ShouldNot(HaveOccurred())

	By("applying openebs 0.8.2")
	applyArtifact(openebsArtifact, "")

	By("waiting for maya-apiserver pod to come into running state")
	podCount := ops.GetPodRunningCountEventually(
		string(artifacts.OpenebsNamespace),
		string(artifacts.MayaAPIServerLabelSelector),
		1,
	)
	Expect(podCount).To(Equal(1))

	annotations := map[string]string{
		string(apis.CASTypeKey):   string(apis.JivaVolume),
		string(apis.CASConfigKey): openebsCASConfigValue,
	}

	By("building a storageclass")
	scObj, err = sc.NewBuilder().
		WithName(scName).
		WithAnnotations(annotations).
		WithProvisioner(openebsProvisioner).Build()
	Expect(err).ShouldNot(HaveOccurred(), "while building storageclass {%s}", scName)

	By("creating above storageclass")
	_, err = ops.SCClient.Create(scObj)
	Expect(err).To(BeNil(), "while creating storageclass {%s}", scObj.Name)

	By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(nsName).
		WithStorageClass(scName).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		nsName,
	)

	By("creating above pvc")
	_, err = ops.PVCClient.WithNamespace(nsName).Create(pvcObj)
	Expect(err).To(
		BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		nsName,
	)

	By("verifying controller pod count ")
	controllerPodCount := ops.GetPodRunningCountEventually(nsName, ctrlLabel, replicaCount)
	Expect(controllerPodCount).To(Equal(replicaCount), "while checking controller pod count")

	By("verifying replica pod count ")
	replicaPodCount := ops.GetPodRunningCountEventually(nsName, replicaLabel, replicaCount)
	Expect(replicaPodCount).To(Equal(replicaCount), "while checking replica pod count")

	By("verifying status as bound")
	status := ops.IsPVCBound(pvcName)
	Expect(status).To(Equal(true), "while checking status equal to bound")

	// TODO
	// By("applying openebs 0.9.0")
	// applyYAMLFromURL("https://openebs.github.io/charts/openebs-operator-0.9.0-RC3.yaml", "")

})

var _ = AfterSuite(func() {

	By("deleting above pvc")
	err := ops.PVCClient.Delete(pvcName, &metav1.DeleteOptions{})
	Expect(err).To(
		BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcName,
		nsName,
	)

	By("verifying controller pod count")
	controllerPodCount := ops.GetPodRunningCountEventually(nsName, ctrlLabel, 0)
	Expect(controllerPodCount).To(Equal(0), "while checking controller pod count")

	By("verifying replica pod count")
	replicaPodCount := ops.GetPodRunningCountEventually(nsName, replicaLabel, 0)
	Expect(replicaPodCount).To(Equal(0), "while checking replica pod count")

	By("deleting storageclass")
	err = ops.SCClient.Delete(scName, &metav1.DeleteOptions{})
	Expect(err).To(BeNil(), "while deleting storageclass {%s}", scObj.Name)

	By("Cleanup")
	deleteArtifact(jobArtifact, "job")
	deleteArtifact(runtaskArtifact, "")
	deleteArtifact(crArtifact, "")
	deleteArtifact(rbacArtifact, "")
	deleteArtifact(openebsArtifact, "")
	podList, err := ops.PodClient.
		WithNamespace("default").
		List(metav1.ListOptions{})
	Expect(err).ShouldNot(HaveOccurred())
	for _, po := range podList.Items {
		if po.Status.Phase == "Succeeded" {
			err = ops.PodClient.Delete(po.Name, &metav1.DeleteOptions{})
			Expect(err).To(BeNil(), "while deleting completed pods")
		}
	}
})

func getArtifactFromURL(url string) artifacts.Artifact {
	// Read yaml file from the url
	o, err := fetch(url)
	Expect(err).ShouldNot(HaveOccurred())
	defer o.Close()
	b, err := ioutil.ReadAll(o)
	Expect(err).ShouldNot(HaveOccurred())
	// Connvert the yaml to unstructured objects
	yamlString := string(b)
	return artifacts.Artifact(yamlString)
}

func applyArtifact(artifact artifacts.Artifact, flag string) {
	artifactList, errs := artifacts.GetArtifactsListUnstructured(artifact)
	Expect(errs).Should(HaveLen(0))
	// Applying unstructured objects
	for i, artifact := range artifactList {
		if flag == "job" && i == 0 {
			unstructured.SetNestedStringMap(artifact.Object, data, "data")
		}
		if flag == "job" && i == 1 {
			artifact.SetNamespace("default")
		}
		cu := k8s.CreateOrUpdate(
			k8s.GroupVersionResourceFromGVK(artifact),
			artifact.GetNamespace(),
		)
		_, err := cu.Apply(artifact)
		Expect(err).ShouldNot(HaveOccurred())
	}
}

func deleteArtifact(artifact artifacts.Artifact, flag string) {
	artifactList, errs := artifacts.GetArtifactsListUnstructured(artifact)
	Expect(errs).Should(HaveLen(0))
	// Applying unstructured objects
	for i, artifact := range artifactList {
		if flag == "job" && i == 1 {
			artifact.SetNamespace("default")
		}
		del := k8s.DeleteResource(
			k8s.GroupVersionResourceFromGVK(artifact),
			artifact.GetNamespace(),
		)
		err := del.Delete(artifact)
		Expect(err).ShouldNot(HaveOccurred())
	}
}

func fetch(url string) (io.ReadCloser, error) {
	httpClient := *http.DefaultClient
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
