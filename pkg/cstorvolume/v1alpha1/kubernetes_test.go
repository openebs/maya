package v1alpha1

import (
	"errors"
	"reflect"
	"testing"

	apis "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha1"
	client "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	clientset "github.com/openebs/maya/pkg/client/generated/clientset/internalclientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fakeGetClientset() (clientset *clientset.Clientset, err error) {
	return &client.Clientset{}, nil
}

func fakeListfn(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.CStorVolumeList, error) {
	return &apis.CStorVolumeList{}, nil
}

func fakeGetfn(cli *clientset.Clientset, name, namespace string, opts metav1.GetOptions) (*apis.CStorVolume, error) {
	return &apis.CStorVolume{}, nil
}

func fakeDelfn(cli *clientset.Clientset, name, namespace string, opts *metav1.DeleteOptions) error {
	return nil
}

func fakeListErrfn(cli *clientset.Clientset, namespace string, opts metav1.ListOptions) (*apis.CStorVolumeList, error) {
	return &apis.CStorVolumeList{}, errors.New("some error")
}

func fakeGetErrfn(cli *clientset.Clientset, name, namespace string, opts metav1.GetOptions) (*apis.CStorVolume, error) {
	return &apis.CStorVolume{}, errors.New("some error")
}

func fakeDelErrfn(cli *clientset.Clientset, name, namespace string, opts *metav1.DeleteOptions) error {
	return errors.New("some error")

}

func fakeSetClientset(k *kubeclient) {
	k.clientset = &client.Clientset{}
}

func fakeSetNilClientset(k *kubeclient) {
	k.clientset = nil
}

func fakeGetNilErrClientSet() (clientset *clientset.Clientset, err error) {
	return nil, nil
}

func fakeGetErrClientSet() (clientset *clientset.Clientset, err error) {
	return nil, errors.New("Some error")
}

func fakeClientSet(k *kubeclient) {}

func TestKubernetesWithDefaults(t *testing.T) {
	tests := map[string]struct {
		expectListFn, expectGetClientset bool
	}{
		"When mockclient is empty":                {true, true},
		"When mockclient contains getClientsetFn": {false, true},
		"When mockclient contains ListFn":         {true, false},
		"When mockclient contains both":           {true, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			fc := &kubeclient{}
			if !mock.expectListFn {
				fc.list = fakeListfn
			}
			if !mock.expectGetClientset {
				fc.getClientset = fakeGetClientset
			}

			fc.withDefaults()
			if mock.expectListFn && fc.list == nil {
				t.Fatalf("test %q failed: expected fc.list not to be empty", name)
			}
			if mock.expectGetClientset && fc.getClientset == nil {
				t.Fatalf("test %q failed: expected fc.getClientset not to be empty", name)
			}
		})
	}
}

func TestKubernetesWithKubeClient(t *testing.T) {
	tests := map[string]struct {
		Clientset             *client.Clientset
		expectKubeClientEmpty bool
	}{
		"Clientset is empty":     {nil, true},
		"Clientset is not empty": {&client.Clientset{}, false},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			h := WithKubeClient(mock.Clientset)
			fake := &kubeclient{}
			h(fake)
			if mock.expectKubeClientEmpty && fake.clientset != nil {
				t.Fatalf("test %q failed expected fake.clientset to be empty", name)
			}
			if !mock.expectKubeClientEmpty && fake.clientset == nil {
				t.Fatalf("test %q failed expected fake.clientset not to be empty", name)
			}
		})
	}
}

func TestKubernetesKubeClient(t *testing.T) {
	tests := map[string]struct {
		expectClientSet bool
		opts            []kubeclientBuildOption
	}{
		"Positive 1": {true, []kubeclientBuildOption{fakeSetClientset}},
		"Positive 2": {true, []kubeclientBuildOption{fakeSetClientset, fakeClientSet}},
		"Positive 3": {true, []kubeclientBuildOption{fakeSetClientset, fakeClientSet, fakeClientSet}},

		"Negative 1": {false, []kubeclientBuildOption{fakeSetNilClientset}},
		"Negative 2": {false, []kubeclientBuildOption{fakeSetNilClientset, fakeClientSet}},
		"Negative 3": {false, []kubeclientBuildOption{fakeSetNilClientset, fakeClientSet, fakeClientSet}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c := KubeClient(mock.opts...)
			if !mock.expectClientSet && c.clientset != nil {
				t.Fatalf("test %q failed expected fake.clientset to be empty", name)
			}
			if mock.expectClientSet && c.clientset == nil {
				t.Fatalf("test %q failed expected fake.clientset not to be empty", name)
			}
		})
	}
}

func TesKubernetestGetClientOrCached(t *testing.T) {
	tests := map[string]struct {
		expectErr  bool
		KubeClient *kubeclient
	}{
		// Positive tests
		"Positive 1": {false, &kubeclient{nil, "", fakeGetNilErrClientSet, fakeGetfn, fakeListfn, fakeDelfn}},
		"Positive 2": {false, &kubeclient{&client.Clientset{}, "", fakeGetNilErrClientSet, fakeGetfn, fakeListfn, fakeDelfn}},
		// Negative tests
		"Negative 1": {true, &kubeclient{nil, "", fakeGetErrClientSet, fakeGetfn, fakeListfn, fakeDelfn}},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			c, err := mock.KubeClient.getClientOrCached()
			if mock.expectErr && err == nil {
				t.Fatalf("test %q failed : expected error not to be nil but got %v", name, err)
			}
			if !reflect.DeepEqual(c, mock.KubeClient.clientset) {
				t.Fatalf("test %q failed : expected clientset %v but got %v", name, mock.KubeClient.clientset, c)
			}
		})
	}
}

func TestWithNamespaceBuildOption(t *testing.T) {
	tests := map[string]struct {
		namespace string
	}{
		"Test 1": {""},
		"Test 2": {"alpha"},
		"Test 3": {"beta"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := KubeClient(WithNamespace(mock.namespace))
			if k.namespace != mock.namespace {
				t.Fatalf("Test %q failed: expected %v got %v", name, mock.namespace, k.namespace)
			}
		})
	}
}

func TestKubenetesList(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		list         listFn
		expectErr    bool
	}{
		"Test 1": {fakeGetErrClientSet, fakeListfn, true},
		"Test 2": {fakeGetClientset, fakeListfn, false},
		"Test 3": {fakeGetClientset, fakeListErrfn, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := kubeclient{getClientset: mock.getClientset, list: mock.list}
			_, err := k.List(metav1.ListOptions{})
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubenetesGet(t *testing.T) {
	tests := map[string]struct {
		getClientset    getClientsetFn
		get             getFn
		expectErr       bool
		name, namespace string
	}{
		"Test 1": {fakeGetErrClientSet, fakeGetfn, true, "testvol", "test-ns"},
		"Test 2": {fakeGetClientset, fakeGetfn, false, "testvol", "test-ns"},
		"Test 3": {fakeGetClientset, fakeGetErrfn, true, "testvol", ""},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := kubeclient{getClientset: mock.getClientset, get: mock.get}
			_, err := k.Get(mock.name, mock.namespace)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}

func TestKubenetesDelete(t *testing.T) {
	tests := map[string]struct {
		getClientset getClientsetFn
		del          delFn
		expectErr    bool
		name         string
	}{
		"Test 1": {fakeGetErrClientSet, fakeDelfn, true, "testvol"},
		"Test 2": {fakeGetClientset, fakeDelfn, false, "testvol"},
		"Test 3": {fakeGetClientset, fakeDelErrfn, true, "testvol"},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			k := kubeclient{getClientset: mock.getClientset, del: mock.del}
			err := k.Delete(mock.name)
			if mock.expectErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.expectErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
		})
	}
}
