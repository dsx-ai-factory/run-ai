import { AssetKind } from "@/swagger-models/assets-service-client";
import { DATA_VOLUME_ROUTE_NAMES } from "../data-volume.routes/data-volume.routes.names";

export enum DATA_SOURCE_ROUTE_NAMES {
  DATA_SOURCE_INDEX = "data-source-index",
  DATA_SOURCE_NEW = "data-source-new",
  DATA_SOURCE_ASSETS_EDIT = "data-source-assets-edit",
  DATA_SOURCE_ENVIRONMENT_NEW = "data-source-environment-new",
  DATA_SOURCE_HOST_PATH_NEW = "data-source-host-path-new",
  DATA_SOURCE_NFS_NEW = "data-source-nfs-new",
  DATA_SOURCE_S3_NEW = "datasource-s3-new",
  DATA_SOURCE_GIT_NEW = "data-source-git-new",
  DATA_SOURCE_PVC_NEW = "data-source-pvc-new",
  DATA_SOURCE_CONFIG_MAP_NEW = "data-source-config-map-new",
  DATA_SOURCE_SECRET_VOLUME_NEW = "data-source-secret-volume-new",
  DATA_SOURCE_DATA_VOLUME_NEW = "data-source-data-volume-new",
}

export const DATA_SOURCE_CREATION_ROUTE_MAP = {
  [AssetKind.Nfs]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_NFS_NEW,
  [AssetKind.Pvc]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_PVC_NEW,
  [AssetKind.S3]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_S3_NEW,
  [AssetKind.Git]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_GIT_NEW,
  [AssetKind.HostPath]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_HOST_PATH_NEW,
  [AssetKind.ConfigMap]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_CONFIG_MAP_NEW,
  [AssetKind.SecretVolume]: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_SECRET_VOLUME_NEW,
  [AssetKind.DataVolume]: DATA_VOLUME_ROUTE_NAMES.DATA_VOLUME_NEW,
};
