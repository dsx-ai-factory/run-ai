/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	trackerclient "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	storagev1 "k8s.io/api/storage/v1"
)

const IsDefaultClassLabel = "storageclass.kubernetes.io/is-default-class"

func ParseStorageClassToServerScheme(storageClass *storagev1.StorageClass) trackerclient.StorageClass {
	return trackerclient.StorageClass{
		ClusterId:        &config.Config().ClusterUuid,
		StorageClassName: storageClass.Name,
		Spec: &trackerclient.StorageClassSpec{
			IsDefault:            CheckIfStorageClassIsDefault(storageClass),
			AllowVolumeExpansion: storageClass.AllowVolumeExpansion,
			Provisioner:          &storageClass.Provisioner,
		},
	}
}

// CheckIfStorageClassIsDefault returns true if the storageClass is the default storageClass in the cluster
func CheckIfStorageClassIsDefault(storageClass *storagev1.StorageClass) bool {
	return storageClass.Annotations != nil && storageClass.Annotations[IsDefaultClassLabel] == "true"
}
