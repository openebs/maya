package v1alpha2

import (
	api "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	zpool "github.com/openebs/maya/pkg/apis/openebs.io/zpool/v1alpha1"
	blockdevice "github.com/openebs/maya/pkg/blockdevice/v1alpha2"
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	zfs "github.com/openebs/maya/pkg/zfs/cmd/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getPathForCSPBdevList(bdevs []api.CStorPoolClusterBlockDevice) ([]string, error) {
	var vdev []string
	var err error

	for _, b := range bdevs {
		path, er := getPathForBDev(b.BlockDeviceName)
		if er != nil {
			err = ErrorWrapf(err, "Failed to fetch path for bdev {%s} {%s}", b.BlockDeviceName, er.Error())
			continue
		}
		vdev = append(vdev, path)
	}
	return vdev, err
}

func getPathForBDev(bdev string) (string, error) {
	bd, err := blockdevice.NewKubeClient().
		WithNamespace(env.Get("NAMESPACE")).
		Get(bdev, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return bd.Spec.Path, nil
}

func checkIfPoolPresent(name string) bool {
	if _, err := zfs.NewPoolGProperty().
		WithParsableMode(true).
		WithScriptedMode(true).
		WithField("name").
		WithProperty("name").
		WithPool(name).
		Execute(); err != nil {
		return false
	}
	return true
}

func isBdevPathChanged(bdev api.CStorPoolClusterBlockDevice) (string, bool, error) {
	var err error
	var isPathChanged bool

	newPath, er := getPathForBDev(bdev.BlockDeviceName)
	if er != nil {
		err = errors.Errorf("Failed to get bdev {%s} path err {%s}", bdev.BlockDeviceName, er.Error())
	}

	if err == nil && newPath != bdev.DevLink {
		isPathChanged = true
	}
	return newPath, isPathChanged, err
}

func compareDisk(path string, d []zpool.Vdev) bool {
	for _, v := range d {
		if path == v.Path {
			return true
		}
		for _, p := range v.Children {
			if path == p.Path {
				return true
			}
			if r := compareDisk(path, p.Children); r {
				return true
			}
		}
	}
	return false
}

func checkIfDeviceUsed(path string, t zpool.Topology) bool {
	var isUsed bool

	if isUsed = compareDisk(path, t.VdevTree.Topvdev); isUsed {
		return isUsed
	}

	if isUsed = compareDisk(path, t.VdevTree.Spares); isUsed {
		return isUsed
	}

	if isUsed = compareDisk(path, t.VdevTree.Readcache); isUsed {
		return isUsed
	}
	return isUsed
}