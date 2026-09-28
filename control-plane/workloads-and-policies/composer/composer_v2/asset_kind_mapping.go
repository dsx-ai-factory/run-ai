/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer_v2

import (
	workloadApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
)

var workloadStorageCracker = jsonutils.NewJsonCracker(workloadApi.WorkspaceSpecSpecAllOfStorage{})
var datasourcesListCracker = jsonutils.NewJsonCracker(assetsApi.DatasourceListResponseAssetSpec{})

const (
	AssetKind_HOST_PATH     = "hostPath"
	AssetKind_NFS           = "nfs"
	AssetKind_PVC           = "pvc"
	AssetKind_GIT           = "git"
	AssetKind_S3            = "s3"
	AssetKind_CONFIG_MAP    = "config-map"
	AssetKind_SECRET_VOLUME = "secret-volume"
)

// Mapping b/w AssetKind_XX to the payload of the workload

var MapAssetKindToWorkloadSpec = map[string]string{
	AssetKind_HOST_PATH:     workloadStorageCracker.NameOf(&workloadStorageCracker.Struct().HostPath),
	AssetKind_NFS:           workloadStorageCracker.NameOf(&workloadStorageCracker.Struct().Nfs),
	AssetKind_PVC:           workloadStorageCracker.NameOf(&workloadStorageCracker.Struct().Pvc),
	AssetKind_S3:            workloadStorageCracker.NameOf(&workloadStorageCracker.Struct().S3),
	AssetKind_GIT:           workloadStorageCracker.NameOf(&workloadStorageCracker.Struct().Git),
	AssetKind_CONFIG_MAP:    workloadStorageCracker.NameOf(&workloadStorageCracker.Struct().ConfigMapVolume),
	AssetKind_SECRET_VOLUME: workloadStorageCracker.NameOf(&workloadStorageCracker.Struct().SecretVolume),
}

// Mapping b/w AssetKind_XX to the payload of datasources list

var MapAssetKindToDatasourcesList = map[string]string{
	AssetKind_HOST_PATH:     datasourcesListCracker.NameOf(&datasourcesListCracker.Struct().HostPath),
	AssetKind_NFS:           datasourcesListCracker.NameOf(&datasourcesListCracker.Struct().Nfs),
	AssetKind_PVC:           datasourcesListCracker.NameOf(&datasourcesListCracker.Struct().Pvc),
	AssetKind_S3:            datasourcesListCracker.NameOf(&datasourcesListCracker.Struct().S3),
	AssetKind_GIT:           datasourcesListCracker.NameOf(&datasourcesListCracker.Struct().Git),
	AssetKind_CONFIG_MAP:    datasourcesListCracker.NameOf(&datasourcesListCracker.Struct().ConfigMap),
	AssetKind_SECRET_VOLUME: datasourcesListCracker.NameOf(&datasourcesListCracker.Struct().Secret),
}
