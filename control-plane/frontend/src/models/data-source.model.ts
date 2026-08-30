import {
  AssetKind,
  type HostPathSpec,
  type NFSAssetSpec,
  type ClaimInfo,
  type GitAssetSpec,
  type PVCAssetSpec,
  type S3AssetSpec,
  type ConfigMapAssetSpec,
  type SecretAssetSpec,
  type DatasourceListResponseAssetSpec,
} from "@/swagger-models/assets-service-client";
import type { WorkspaceSpecSpecAllOfStorage } from "@/swagger-models/workloads-client";

import type { IUIAssetCreationMeta } from "./global.model";
import { ResourceType } from "@/swagger-models/authorization-client";
import type { SharedScopeBase } from "@/swagger-models/datavolumes-service-client";
import type { IUIStorageFields } from "./workload.model";

export type TDataSourceKinds =
  | typeof AssetKind.HostPath
  | typeof AssetKind.Nfs
  | typeof AssetKind.Git
  | typeof AssetKind.S3
  | typeof AssetKind.Pvc
  | typeof AssetKind.ConfigMap
  | typeof AssetKind.SecretVolume
  | typeof AssetKind.DataVolume;

export type TUIDataSourceInstancesKinds =
  | typeof AssetKind.HostPath
  | typeof AssetKind.Nfs
  | typeof AssetKind.Git
  | typeof AssetKind.S3
  | typeof AssetKind.Pvc
  | typeof AssetKind.ConfigMap
  | typeof AssetKind.SecretVolume
  | typeof AssetKind.DataVolume
  | typeof UI_VOLUME_KIND
  | typeof UI_EMPTY_DIR_KIND;

export const UI_VOLUME_KIND = `pvc-volume`;
export const UI_EMPTY_DIR_KIND = `empty-dir`;

export const NONE_STORAGE_CLASS = "runai-none";
export const NONE_STORAGE_CLASS_LABEL = "None";

export const DATA_SOURCE_TYPE: Record<TDataSourceKinds, string> = {
  [AssetKind.HostPath]: "Host path",
  [AssetKind.Nfs]: "NFS",
  [AssetKind.Git]: "Git",
  [AssetKind.S3]: "S3 Bucket",
  [AssetKind.Pvc]: "PVC",
  [AssetKind.ConfigMap]: "ConfigMap",
  [AssetKind.SecretVolume]: "Secret",
  [AssetKind.DataVolume]: "Data volume",
};

export const DATA_SOURCE_KIND_TO_SPEC_KEY: Record<TDataSourceKinds, string> = {
  [AssetKind.HostPath]: AssetKind.HostPath,
  [AssetKind.Nfs]: AssetKind.Nfs,
  [AssetKind.Git]: AssetKind.Git,
  [AssetKind.S3]: AssetKind.S3,
  [AssetKind.Pvc]: AssetKind.Pvc,
  [AssetKind.ConfigMap]: "config_map",
  [AssetKind.SecretVolume]: "secret",
  [AssetKind.DataVolume]: "dataVolume",
};

export const DATA_SOURCE_KIND_TO_STORAGE_FIELD_KEY: Record<TDataSourceKinds, keyof IUIStorageFields> = {
  [AssetKind.HostPath]: "hostPath",
  [AssetKind.Nfs]: "nfs",
  [AssetKind.Git]: "git",
  [AssetKind.S3]: "s3",
  [AssetKind.Pvc]: "pvc",
  [AssetKind.ConfigMap]: "configMapVolume",
  [AssetKind.SecretVolume]: "secretVolume",
  [AssetKind.DataVolume]: "dataVolume",
};

export const DATA_SOURCE_KIND_TO_CONTAINER_KEY: Record<TDataSourceKinds, string> = {
  [AssetKind.HostPath]: "mountPath",
  [AssetKind.Nfs]: "mountPath",
  [AssetKind.Git]: "path",
  [AssetKind.S3]: "path",
  [AssetKind.Pvc]: "path",
  [AssetKind.ConfigMap]: "mountPath",
  [AssetKind.SecretVolume]: "mountPath",
  [AssetKind.DataVolume]: "mountPath",
};

export type TDataSourceResourceTypes =
  | typeof ResourceType.NfsAssets
  | typeof ResourceType.S3Assets
  | typeof ResourceType.GitAssets
  | typeof ResourceType.PvcAssets
  | typeof ResourceType.HostPathAssets
  | typeof ResourceType.CmVolumeAssets
  | typeof ResourceType.SecretVolumeAssets
  | typeof ResourceType.Datavolumes;

export const DS_RESOURCE_TYPE_TO_ASSET_KIND: Record<TDataSourceResourceTypes, TDataSourceKinds> = {
  [ResourceType.NfsAssets]: AssetKind.Nfs,
  [ResourceType.S3Assets]: AssetKind.S3,
  [ResourceType.GitAssets]: AssetKind.Git,
  [ResourceType.PvcAssets]: AssetKind.Pvc,
  [ResourceType.HostPathAssets]: AssetKind.HostPath,
  [ResourceType.CmVolumeAssets]: AssetKind.ConfigMap,
  [ResourceType.SecretVolumeAssets]: AssetKind.SecretVolume,
  [ResourceType.Datavolumes]: AssetKind.DataVolume,
};

export const DS_KIND_TO_RESOURCE_TYPE: Record<TDataSourceKinds, TDataSourceResourceTypes> = {
  [AssetKind.Nfs]: ResourceType.NfsAssets,
  [AssetKind.S3]: ResourceType.S3Assets,
  [AssetKind.Git]: ResourceType.GitAssets,
  [AssetKind.Pvc]: ResourceType.PvcAssets,
  [AssetKind.HostPath]: ResourceType.HostPathAssets,
  [AssetKind.ConfigMap]: ResourceType.CmVolumeAssets,
  [AssetKind.SecretVolume]: ResourceType.SecretVolumeAssets,
  [AssetKind.DataVolume]: ResourceType.Datavolumes,
};

export const DS_KIND_TO_DS_LIST_KEY: Record<TDataSourceKinds, keyof DatasourceListResponseAssetSpec> = {
  [AssetKind.HostPath]: "hostPath",
  [AssetKind.Nfs]: "nfs",
  [AssetKind.Git]: "git",
  [AssetKind.S3]: "s3",
  [AssetKind.Pvc]: "pvc",
  [AssetKind.ConfigMap]: "config_map",
  [AssetKind.SecretVolume]: "secret",
  [AssetKind.DataVolume]: "dataVolume",
};

export const DS_KIND_TO_STORAGE_FIELDS_KEY: Record<TUIDataSourceInstancesKinds, keyof WorkspaceSpecSpecAllOfStorage> = {
  [AssetKind.HostPath]: "hostPath",
  [AssetKind.Nfs]: "nfs",
  [AssetKind.Git]: "git",
  [AssetKind.S3]: "s3",
  [AssetKind.Pvc]: "pvc",
  [AssetKind.ConfigMap]: "configMapVolume",
  [AssetKind.SecretVolume]: "secretVolume",
  [AssetKind.DataVolume]: "dataVolume",
  [UI_VOLUME_KIND]: "pvc",
  [UI_EMPTY_DIR_KIND]: "emptyDirVolume",
};

export const DATA_SOURCES_TYPES_TO_NAMES = {
  NFS: "NFS",
  PVC: "PVC",
  S3: "S3 Bucket",
  GIT: "Git",
  HOST_PATH: "Host path",
  CONFIG_MAP: "ConfigMap",
  SECRET_VOLUME: "Secret",
  DATA_VOLUME: "Data volume",
};

export const DATA_SOURCE_ASSET_KINDS_TO_NAMES: Record<TDataSourceKinds, string> = {
  [AssetKind.Nfs]: DATA_SOURCES_TYPES_TO_NAMES.NFS,
  [AssetKind.Pvc]: DATA_SOURCES_TYPES_TO_NAMES.PVC,
  [AssetKind.S3]: DATA_SOURCES_TYPES_TO_NAMES.S3,
  [AssetKind.Git]: DATA_SOURCES_TYPES_TO_NAMES.GIT,
  [AssetKind.HostPath]: DATA_SOURCES_TYPES_TO_NAMES.HOST_PATH,
  [AssetKind.ConfigMap]: DATA_SOURCES_TYPES_TO_NAMES.CONFIG_MAP,
  [AssetKind.SecretVolume]: DATA_SOURCES_TYPES_TO_NAMES.SECRET_VOLUME,
  [AssetKind.DataVolume]: DATA_SOURCES_TYPES_TO_NAMES.DATA_VOLUME,
};
export interface IUIVolume {
  // id: string;
  claimName: string;
  path: string;
  claimInfo: ClaimInfo;
  ephemeral: boolean;
}
export interface IUIHostPathCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: HostPathSpec;
}
export interface IUINFSCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: NFSAssetSpec;
}
export interface IUIS3CreationRequest {
  meta: IUIAssetCreationMeta;
  spec: S3AssetSpec;
}
export interface IUIGitCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: GitAssetSpec;
}
export interface IUIPVCCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: PVCAssetSpec;
}
export interface IUIConfigMapCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: ConfigMapAssetSpec;
}
export interface IUISecretCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: SecretAssetSpec;
}
export interface IDataSourceModalOptions {
  dataSourceIds: Array<string>;
  header: string;
}

export interface IDataVolumeCreationRequest {
  meta: {
    name: string;
    projectId: string;
    description?: string;
  };
  spec: {
    originPvcName: string;
    sharedScopes: Array<SharedScopeBase>;
  };
}
