/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// services
import { assetsServiceApi } from "@/services/infra/client-apis/assets-service-api/assets-service-api";
import { k8sObjectTrackerApi } from "@/services/infra/client-apis/k8s-object-tracker-api/k8s-object-tracker-api";
import { httpService } from "@/services/infra/https.service/http.service";
import { makeId, pick } from "@/utils/common.util";

// model
import type { HistoryRecord } from "@/swagger-models/assets-service-client";
import {
  type DatasourceListResponseEntry,
  type HostPathCreationRequest,
  type HostPathAsset,
  type NFSCreationRequest,
  type S3CreationRequest,
  type NFSAsset,
  type S3Asset,
  type GitCreationRequest,
  type GitAsset,
  type PVCCreationRequest,
  type PVCAsset,
  type ClaimInfo,
  type AccessKey,
  type Password,
  type HostPathUpdateRequest,
  type S3UpdateRequest,
  type PVCUpdateRequest,
  type NFSUpdateRequest,
  type GitUpdateRequest,
  PvcVolumeMode,
  AssetKind,
  type HttpResponse,
  type ConfigMapAsset,
  type ConfigMapCreationRequest,
  type ConfigMapUpdateRequest,
  type GenericSecret,
  type SecretAsset,
  type SecretAssetUpdateRequest,
  type SecretAssetCreationRequest,
} from "@/swagger-models/assets-service-client";
import type {
  ConfigMapListInfo,
  PVCListInfo,
  StorageClass,
  StorageClassConfiguration,
  StorageClassV2,
} from "@/swagger-models/k8s-objects-tracker-client";
import type { IAssetsFilter } from "@/models/filter.model";
import {
  type IUINFSCreationRequest,
  type IUIPVCCreationRequest,
  type IUIHostPathCreationRequest,
  type IUIVolume,
  type IUIS3CreationRequest,
  type IUIGitCreationRequest,
  type IUISecretCreationRequest,
} from "@/models/data-source.model";
// constant
import type { IUIConfigMapCreationRequest, TDataSourceKinds } from "@/models/data-source.model";
import { dataSourceUtil } from "@/utils/data-source.util";
import type { IScopeInfo } from "@/utils/data-source.util/data-source.util";
import type { IScopeCredentialsList } from "@/models/credential.model";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { dataVolumesServiceApi } from "@/services/infra/client-apis/data-volume-service-api/data-volume-service-api";

export const dataSourceService = {
  list,
  remove,
  createHostPath,
  createNFS,
  createPVC,
  createS3,
  createGit,
  createConfigMap,
  createSecretVolume,
  listStorageClass,
  listStorageClassV2,
  getStorageClassConfigurationByName,
  listAccessKeysByScope,
  listPasswordsByScope,
  listGenericSecretsByScope,
  getHostPathById,
  getS3ById,
  getPvcById,
  getGitById,
  getNFSById,
  getSecretVolumeById,
  getConfigMapById,
  updateHostPath,
  updateGit,
  updateNFS,
  updatePVC,
  updateS3,
  updateConfigMap,
  updateSecretVolume,
  loadPVCAssets,
  loadPvcEvents,
  createWorkloadVolumes,
  listExistingPVCs,
  listExistingConfigMaps,
  getDataSourceById,
  // helpers
  getEmptyHostPathModel,
  getEmptyNfsModel,
  getEmptyS3Model,
  getEmptyGitModel,
  getEmptyPvcModel,
  getEmptyClaimInfo,
  getEmptyConfigMapModel,
  getEmptySecretVolumeModel,
};

// api calls
async function list(filterBy: IAssetsFilter = {}): Promise<Array<DatasourceListResponseEntry>> {
  const filters: IAssetsFilter = pick(
    filterBy,
    "sortBy",
    "page",
    "rowsPerPage",
    "projectId",
    "departmentId",
    "scope",
    "usageInfo",
    "complyToProject",
    "complyToWorkloadType",
    "clusterUuid",
    "statusInfo",
    "complyToReplicaType",
    "assetIds",
    "dataVolumes",
  );
  try {
    return assetsServiceApi.dataSourceApi
      .listDatasourceAssets(
        undefined,
        filters.scope,
        filters.projectId ? Number(filters.projectId) : undefined,
        filters.departmentId?.toString(),
        filters.clusterUuid?.toString(),
        filters.usageInfo,
        filters.complyToProject,
        filters.complyToWorkloadType,
        filters.statusInfo,
        filters.assetIds,
        filters.complyToReplicaType,
        filters.dataVolumes,
      )

      .then((res) => res.data.entries);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function remove(dataSource: DatasourceListResponseEntry): Promise<HttpResponse> {
  const dataSourceId: string = dataSource.meta.id;
  try {
    switch (dataSource.meta.kind) {
      case AssetKind.Pvc:
        return (await assetsServiceApi.pvcApi.deletePvcAssetById(dataSourceId)).data;
      case AssetKind.Git:
        return (await assetsServiceApi.gitApi.deleteGitAssetById(dataSourceId)).data;
      case AssetKind.Nfs:
        return (await assetsServiceApi.nfsApi.deleteNfsAssetById(dataSourceId)).data;
      case AssetKind.HostPath:
        return (await assetsServiceApi.hostPathApi.deleteHostPathById(dataSourceId)).data;
      case AssetKind.S3:
        return (await assetsServiceApi.s3Api.deleteS3AssetById(dataSourceId)).data;
      case AssetKind.ConfigMap:
        return (await assetsServiceApi.configMapApi.deleteConfigMapAssetById(dataSourceId)).data;
      case AssetKind.SecretVolume:
        return (await assetsServiceApi.secretApi.deleteSecretAssetById(dataSourceId)).data;
      case AssetKind.DataVolume:
        return (await dataVolumesServiceApi.dataVolumesApi.deleteDatavolume(dataSourceId)).data;
      default:
        throw `Wrong data source kind ${dataSource.meta.kind}`;
    }
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listExistingPVCs(
  scopeType: string,
  scopeId: string,
  pvcOrigin: "existing" | "assets" | "all" = "existing",
  includeAncestors = false,
): Promise<PVCListInfo[]> {
  try {
    const response = await k8sObjectTrackerApi.pvcApi.listPvcs(
      scopeType,
      scopeId,
      undefined,
      undefined,
      pvcOrigin,
      includeAncestors,
    );
    return response.data.pvcs;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listExistingConfigMaps(
  scopeType: string,
  scopeId: string,
  includeAncestors = false,
): Promise<ConfigMapListInfo[]> {
  try {
    const response = await k8sObjectTrackerApi.configMapsApi.listConfigMaps(
      scopeType,
      scopeId,
      undefined,
      undefined,
      includeAncestors,
    );
    return response.data.configMaps;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listStorageClass(clusterId: string): Promise<string[]> {
  const data = (await k8sObjectTrackerApi.storageClassesApi.getStorageClasses(clusterId, undefined, true)).data;
  return data?.items?.map((item: StorageClass) => item.storage_class_name) || [];
}

async function listStorageClassV2(clusterId: string, name?: string): Promise<StorageClassV2[]> {
  const data = (await k8sObjectTrackerApi.storageClassesApi.getStorageClassesV2(clusterId, name, true)).data;
  return data?.items || [];
}

async function getStorageClassConfigurationByName(
  clusterId: string,
  name: string,
): Promise<StorageClassConfiguration[]> {
  try {
    const response = await k8sObjectTrackerApi.storageClassConfigurationApi.listStorageClassConfiguration([
      filterUtil.getEqualsFilterString("scopeId", clusterId),
      filterUtil.getEqualsFilterString("name", name),
    ]);
    return response.data.storageClassConfigurations;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listAccessKeysByScope(scope: IScopeCredentialsList): Promise<AccessKey[]> {
  try {
    const response = await assetsServiceApi.credentialApi.listAccessKeys(
      undefined,
      scope.scope,
      Number(scope.projectId) || undefined,
      scope.departmentId || undefined,
      scope.clusterUuid || undefined,
    );
    return response.data.entries;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listPasswordsByScope(scope: IScopeCredentialsList): Promise<Password[]> {
  try {
    const response = await assetsServiceApi.credentialApi.listPasswords(
      undefined,
      scope.scope,
      Number(scope.projectId) || undefined,
      scope.departmentId || undefined,
      scope.clusterUuid || undefined,
    );
    return response.data.entries;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listGenericSecretsByScope(scope: IScopeCredentialsList): Promise<GenericSecret[]> {
  try {
    const response = await assetsServiceApi.credentialApi.listGenericSecret(
      undefined,
      scope.scope,
      Number(scope.projectId) || undefined,
      scope.departmentId || undefined,
      scope.clusterUuid || undefined,
    );
    return response.data.entries;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

interface IDataSourceGetByIdParams {
  usageInfo?: boolean;
  complyToProject?: number;
  complyToWorkloadType?: "Workspace" | "Training";
  statusInfo?: boolean;
  // only for PVC assets, to check if the PVC is being used by data volumes
  dvUsageInfo?: boolean;
}

async function getDataSourceById(
  dataSourceId: string,
  dataSourceKind: TDataSourceKinds,
  options: IDataSourceGetByIdParams = {},
): Promise<HostPathAsset | S3Asset | PVCAsset | NFSAsset | GitAsset | ConfigMapAsset | null> {
  switch (dataSourceKind) {
    case AssetKind.S3:
      return getS3ById(dataSourceId, options);
    case AssetKind.HostPath:
      return getHostPathById(dataSourceId, options);
    case AssetKind.Nfs:
      return getNFSById(dataSourceId, options);
    case AssetKind.Pvc:
      return getPvcById(dataSourceId, options);
    case AssetKind.Git:
      return getGitById(dataSourceId, options);
    case AssetKind.ConfigMap:
      return getConfigMapById(dataSourceId, options);
    default:
      return null;
  }
}
async function getHostPathById(hostPathId: string, options: IDataSourceGetByIdParams = {}): Promise<HostPathAsset> {
  const params = pick(options, "complyToProject", "complyToWorkloadType", "usageInfo");

  try {
    const response = await assetsServiceApi.hostPathApi.getHostPathById(
      hostPathId,
      params.usageInfo,
      params.complyToProject,
      params.complyToWorkloadType,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getS3ById(s3Id: string, options: IDataSourceGetByIdParams = {}): Promise<S3Asset> {
  const params = pick(options, "complyToProject", "complyToWorkloadType", "statusInfo", "usageInfo");

  try {
    const response = await assetsServiceApi.s3Api.getS3AssetById(
      s3Id,
      params.usageInfo,
      params.complyToProject,
      params.complyToWorkloadType,
      params.statusInfo,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function getPvcById(pvcId: string, options: IDataSourceGetByIdParams = {}): Promise<PVCAsset> {
  const params = pick(options, "complyToProject", "complyToWorkloadType", "statusInfo", "usageInfo", "dvUsageInfo");

  try {
    const response = await assetsServiceApi.pvcApi.getPvcAssetById(
      pvcId,
      params.usageInfo,
      params.complyToProject,
      params.complyToWorkloadType,
      params.statusInfo,
      undefined,
      params.dvUsageInfo,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function getNFSById(nfsId: string, options: IDataSourceGetByIdParams = {}): Promise<NFSAsset> {
  const params = pick(options, "complyToProject", "complyToWorkloadType", "usageInfo");

  try {
    const response = await assetsServiceApi.nfsApi.getNfsAssetById(
      nfsId,
      params.usageInfo,
      params.complyToProject,
      params.complyToWorkloadType,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getGitById(gitId: string, options: IDataSourceGetByIdParams = {}): Promise<GitAsset> {
  const params = pick(options, "complyToProject", "complyToWorkloadType", "statusInfo", "usageInfo");

  try {
    const response = await assetsServiceApi.gitApi.getGitAssetById(
      gitId,
      params.usageInfo,
      params.complyToProject,
      params.complyToWorkloadType,
      params.statusInfo,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getConfigMapById(cmId: string, options: IDataSourceGetByIdParams = {}): Promise<ConfigMapAsset> {
  const params = pick(options, "complyToProject", "complyToWorkloadType", "statusInfo", "usageInfo");

  try {
    const response = await assetsServiceApi.configMapApi.getConfigMapAssetById(
      cmId,
      params.usageInfo,
      params.complyToProject,
      params.complyToWorkloadType,
      params.statusInfo,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getSecretVolumeById(secretId: string, options: IDataSourceGetByIdParams = {}): Promise<SecretAsset> {
  const params = pick(options, "complyToProject", "complyToWorkloadType", "statusInfo", "usageInfo");

  try {
    const response = await assetsServiceApi.secretApi.getSecretAssetById(
      secretId,
      params.usageInfo,
      params.complyToProject,
      params.complyToWorkloadType,
      params.statusInfo,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateHostPath(hostPathId: string, hostPathDataSource: HostPathUpdateRequest): Promise<HostPathAsset> {
  try {
    const response = await assetsServiceApi.hostPathApi.updateHostPathById(hostPathId, hostPathDataSource);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateS3(s3Id: string, s3DataSource: S3UpdateRequest): Promise<S3Asset> {
  try {
    const response = await assetsServiceApi.s3Api.updateS3AssetById(s3Id, s3DataSource);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function updatePVC(pvcId: string, pvcDataSource: PVCUpdateRequest): Promise<PVCAsset> {
  try {
    const response = await assetsServiceApi.pvcApi.updatePvcAssetById(pvcId, pvcDataSource);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function updateNFS(nfsId: string, nfsDataSource: NFSUpdateRequest): Promise<NFSAsset> {
  try {
    const response = await assetsServiceApi.nfsApi.updateNfsAssetById(nfsId, nfsDataSource);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateGit(gitId: string, gitDataSource: GitUpdateRequest): Promise<GitAsset> {
  try {
    const response = await assetsServiceApi.gitApi.updateGitAssetById(gitId, gitDataSource);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateConfigMap(cmId: string, cmDataSource: ConfigMapUpdateRequest): Promise<ConfigMapAsset> {
  try {
    const response = await assetsServiceApi.configMapApi.updateConfigMapAssetById(cmId, cmDataSource);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateSecretVolume(secretId: string, secretDataSource: SecretAssetUpdateRequest): Promise<SecretAsset> {
  try {
    const response = await assetsServiceApi.secretApi.updateSecretAssetById(secretId, secretDataSource);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function loadPVCAssets(pvcIds: Array<string>): Promise<Array<PVCAsset>> {
  if (!pvcIds?.length) return [];
  const prms = pvcIds.map((id: string) => getPvcById(id));
  return Promise.all(prms);
}

async function createWorkloadVolumes(
  workloadName: string,
  volumes: Array<IUIVolume>,
  scopeInfo: IScopeInfo,
): Promise<Array<string>> {
  const createdVolumes: PVCAsset[] = await Promise.all(
    volumes.map((volume: IUIVolume) => {
      volume.claimName = `${workloadName}-pvc-${makeId(5)}`;
      const pvc = dataSourceUtil.convertVolumeToPvc(volume, scopeInfo);
      return createPVC(pvc);
    }),
  );
  return createdVolumes.map((volume: PVCAsset) => volume.meta.id);
}

async function loadPvcEvents(pvcId: string): Promise<HistoryRecord[]> {
  const resp = await assetsServiceApi.pvcApi.getPvcHistory(pvcId);
  return resp.data.records;
}

// helpers
function getEmptyHostPathModel(): IUIHostPathCreationRequest {
  return {
    meta: {
      name: "",
      description: null,
      scope: null,
      projectId: null,
    },
    spec: {
      path: "",
      readOnly: false,
      mountPath: "",
    },
  };
}

function getEmptyNfsModel(): IUINFSCreationRequest {
  return {
    meta: {
      name: "",
      description: null,
      scope: null,
      clusterId: null,
      projectId: null,
    },
    spec: {
      path: "",
      readOnly: false,
      server: "",
      mountPath: "",
    },
  };
}
function getEmptyS3Model(): IUIS3CreationRequest {
  return {
    meta: {
      name: "",
      description: null,
      scope: null,
      clusterId: null,
      projectId: null,
    },
    spec: {
      bucket: "",
      path: "",
      url: null,
      accessKeyAssetId: "",
    },
  };
}

function getEmptyGitModel(): IUIGitCreationRequest {
  return {
    meta: {
      name: "",
      description: null,
      scope: null,
      clusterId: null,
      projectId: null,
    },
    spec: {
      repository: "",
      branch: null,
      revision: null,
      path: "",
    },
  };
}

function getEmptyPvcModel(): IUIPVCCreationRequest {
  return {
    meta: {
      name: "",
      description: null,
      scope: null,
      clusterId: null,
      projectId: null,
    },
    spec: {
      path: "",
      claimName: "",
      existingPvc: true,
      claimInfo: null,
      readOnly: false,
    },
  };
}

function getEmptyClaimInfo(): ClaimInfo {
  return {
    size: "",
    storageClass: null,
    accessModes: {
      readWriteOnce: false,
      readOnlyMany: false,
      readWriteMany: false,
    },
    volumeMode: PvcVolumeMode.Filesystem,
  };
}

function getEmptyConfigMapModel(): IUIConfigMapCreationRequest {
  return {
    meta: {
      name: "",
      description: null,
      scope: null,
      clusterId: null,
      projectId: null,
    },
    spec: {
      configMap: "",
      mountPath: "",
    },
  };
}

function getEmptySecretVolumeModel(): IUISecretCreationRequest {
  return {
    meta: {
      name: "",
      description: null,
      scope: null,
      clusterId: null,
      projectId: null,
    },
    spec: {
      mountPath: "",
      credentialAssetId: "",
    },
  };
}

async function createHostPath(hostPath: HostPathCreationRequest): Promise<HostPathAsset> {
  try {
    const response = await assetsServiceApi.hostPathApi.createHostPath(hostPath);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createS3(s3: S3CreationRequest): Promise<S3Asset> {
  try {
    const response = await assetsServiceApi.s3Api.createS3Asset(s3);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createNFS(nfs: NFSCreationRequest): Promise<NFSAsset> {
  try {
    const response = await assetsServiceApi.nfsApi.createNfsAsset(nfs);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createGit(git: GitCreationRequest): Promise<GitAsset> {
  try {
    const response = await assetsServiceApi.gitApi.createGitAsset(git);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createPVC(pvc: PVCCreationRequest): Promise<PVCAsset> {
  try {
    const response = await assetsServiceApi.pvcApi.createPvcAsset(pvc);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createConfigMap(cm: ConfigMapCreationRequest): Promise<ConfigMapAsset> {
  try {
    const response = await assetsServiceApi.configMapApi.createConfigMapAsset(cm);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createSecretVolume(secretVolume: SecretAssetCreationRequest): Promise<SecretAsset> {
  try {
    const response = await assetsServiceApi.secretApi.createSecretAsset(secretVolume);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
