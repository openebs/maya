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
	env "github.com/openebs/maya/pkg/env/v1alpha1"
	"strconv"
)

// IsCstorSparsePoolEnabled reads from env variable to check wether cstor sparse pool
// should be created by default or not.
func IsCstorSparsePoolEnabled() (enabled bool) {
	enabled, _ = strconv.ParseBool(env.Get(string(CASDefaultCstorPool)))
	return
}

// CstorPoolSpc070 returns the default storagepoolclaim yaml
// corresponding to version 0.7.0 if cstor sparse pool creation is enabled as a part of
// openebs installation
func CstorPoolSpc070() (list ArtifactList) {
	list.Items = append(list.Items, ParseArtifactListFromMultipleYamlConditional(cstorPoolSpcFor070, IsCstorSparsePoolEnabled)...)
	return
}

// cstorPoolSpcFor070 returns all the yamls related to configuring a stoaragepoolclaim in string
// format
//
// NOTE:
//  This is an implementation of MultiYamlFetcher
func cstorPoolSpcFor070() string {
	return `
---
apiVersion: openebs.io/v1alpha1
kind: StoragePoolClaim
metadata:
  name: cstor-sparse-pool
  annotations:
    cas.openebs.io/create-pool-template: cstor-pool-create-default-0.7.0
    cas.openebs.io/delete-pool-template: cstor-pool-delete-default-0.7.0
    cas.openebs.io/config: |
      - name: PoolResourceLimits
        value: false
      #- name: PoolResourceLimits
      #  value: |-
      #      memory: 2Gi
      #      cpu: 500m
      #- name: AuxResourceLimits
      #  value: |-
      #      memory: 1Gi
      #      cpu: 100m
spec:
  name: cstor-sparse-pool
  type: sparse
  maxPools: 3
  poolSpec:
    poolType: striped
    cacheFile: /tmp/cstor-sparse-pool.cache
    overProvisioning: false
---
`
}
