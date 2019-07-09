/*
Copyright 2018 The OpenEBS Authors.

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

package v1alpha2

import (
	api "github.com/openebs/maya/pkg/apis/openebs.io/v1alpha2"
	"github.com/openebs/maya/pkg/util"

	//	clientset1 "github.com/openebs/maya/pkg/client/generated/clientset/versioned"

	clientset2 "github.com/openebs/maya/pkg/client/generated/openebs.io/v1alpha2/clientset/internalclientset"
	"k8s.io/client-go/kubernetes"
)

// ImportedCStorPools is a map of imported cstor pools API config identified via their UID
var ImportedCStorPools map[string]*api.CStorNPool

//PoolAddEventHandled is a flag representing if the pool has been initially imported or created
var PoolAddEventHandled = false

// RunnerVar the runner variable for executing binaries.
var RunnerVar util.Runner

// KubeClient is for kubernetes CR operation
var KubeClient kubernetes.Interface

// OpenEBSClient2 is for openebs CR operation
var OpenEBSClient2 clientset2.Interface