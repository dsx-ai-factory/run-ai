import { type IDrawerListItem, WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import { computeResourceService } from "@/services/control-plane/compute-resource.service/compute-resource.service";
import { environmentService } from "@/services/control-plane/environment.service/environment.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import type { IAssetsFilter } from "@/models/filter.model";
import { type ISelectOption } from "@/models/global.model";
import { DATA_SOURCES_TYPES_TO_NAMES } from "@/models/data-source.model";
import { DATA_SOURCE_ICONS_NAMES } from "@/common/icons.constant";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { useSettingStore } from "@/stores/setting.store";

interface IDrawerEntityMeta {
  entityLabel: string;
  title: string;
  alternativeTitle?: () => string;
  // Phosphor glyph class for the empty-state icon (e.g. "ph-bold ph-database"). Rendered via <q-icon>.
  defaultIcon: string;
  creationOptions?: ISelectOption[];
  loadFunction: (filter: IAssetsFilter) => Promise<IDrawerListItem[]>;
}

const loadDataSources = async (filter: IAssetsFilter) => {
  const isAddDataVolumes = useSettingStore().isDataVolumeEnabled && !!filter.projectId;
  const dataSources = await dataSourceService.list({ ...filter, dataVolumes: isAddDataVolumes, statusInfo: true });

  const loadedItems = workloadHybridUtil.dataSourceListToDrawerItems(dataSources);

  return loadedItems;
};

const loadEnvironments = async (filter: IAssetsFilter) => {
  const environments = await environmentService.list(filter);
  return workloadHybridUtil.environmentListToDrawerItems(environments);
};

const loadComputeResources = async (filter: IAssetsFilter) => {
  const computeResources = await computeResourceService.list(filter);
  return workloadHybridUtil.computeListToDrawerItems(computeResources);
};

const DATA_SOURCE_CREATION_OPTIONS = {
  NFS: {
    label: DATA_SOURCES_TYPES_TO_NAMES.NFS,
    value: AssetKind.Nfs,
    svg: DATA_SOURCE_ICONS_NAMES.nfs,
    aid: `ds-${AssetKind.Nfs}-option`,
  },
  PVC: {
    label: DATA_SOURCES_TYPES_TO_NAMES.PVC,
    value: AssetKind.Pvc,
    svg: DATA_SOURCE_ICONS_NAMES.pvc,
    aid: `ds-${AssetKind.Pvc}-option`,
  },
  S3: {
    label: DATA_SOURCES_TYPES_TO_NAMES.S3,
    value: AssetKind.S3,
    svg: DATA_SOURCE_ICONS_NAMES.s3,
    aid: `ds-${AssetKind.S3}-option`,
  },
  GIT: {
    label: DATA_SOURCES_TYPES_TO_NAMES.GIT,
    value: AssetKind.Git,
    svg: DATA_SOURCE_ICONS_NAMES.git,
    aid: `ds-${AssetKind.Git}-option`,
  },
  HOST_PATH: {
    label: DATA_SOURCES_TYPES_TO_NAMES.HOST_PATH,
    value: AssetKind.HostPath,
    svg: DATA_SOURCE_ICONS_NAMES.hostPath,
    aid: `ds-${AssetKind.HostPath}-option`,
  },
  CONFIG_MAP: {
    label: DATA_SOURCES_TYPES_TO_NAMES.CONFIG_MAP,
    value: AssetKind.ConfigMap,
    svg: DATA_SOURCE_ICONS_NAMES["config-map"],
    aid: `ds-${AssetKind.ConfigMap}-option`,
  },
  SECRET_VOLUME: {
    label: DATA_SOURCES_TYPES_TO_NAMES.SECRET_VOLUME,
    value: AssetKind.SecretVolume,
    svg: DATA_SOURCE_ICONS_NAMES["secret-volume"],
    aid: `ds-${AssetKind.SecretVolume}-option`,
  },
  DATA_VOLUME: {
    label: DATA_SOURCES_TYPES_TO_NAMES.DATA_VOLUME,
    value: AssetKind.DataVolume,
    svg: DATA_SOURCE_ICONS_NAMES["data-volume"],
    aid: `ds-${AssetKind.DataVolume}-option`,
  },
} as const;

export const workloadCreationDrawerEntityMapper: Record<WorkloadCreationDrawerEntities, IDrawerEntityMeta> = {
  [WorkloadCreationDrawerEntities.Environment]: {
    entityLabel: "environment",
    title: "Environments",
    defaultIcon: "ph-bold ph-x-square",
    loadFunction: loadEnvironments,
  },
  [WorkloadCreationDrawerEntities.Compute]: {
    entityLabel: "compute resource",
    title: "Compute Resources",
    defaultIcon: "ph-bold ph-hard-drives",
    loadFunction: loadComputeResources,
  },
  [WorkloadCreationDrawerEntities.DataSource]: {
    entityLabel: "setup",
    title: "Data Sources",
    alternativeTitle: () => (useSettingStore().isDataVolumeEnabled ? "Data & storage" : ""),
    defaultIcon: "ph-bold ph-database",
    creationOptions: [
      DATA_SOURCE_CREATION_OPTIONS.NFS,
      DATA_SOURCE_CREATION_OPTIONS.PVC,
      DATA_SOURCE_CREATION_OPTIONS.S3,
      DATA_SOURCE_CREATION_OPTIONS.GIT,
      DATA_SOURCE_CREATION_OPTIONS.HOST_PATH,
      DATA_SOURCE_CREATION_OPTIONS.CONFIG_MAP,
      DATA_SOURCE_CREATION_OPTIONS.SECRET_VOLUME,
      DATA_SOURCE_CREATION_OPTIONS.DATA_VOLUME,
    ],
    loadFunction: loadDataSources,
  },
  [WorkloadCreationDrawerEntities.ModelStorage]: {
    entityLabel: "setup",
    title: "Data Sources",
    alternativeTitle: () => (useSettingStore().isDataVolumeEnabled ? "Data & storage" : ""),
    defaultIcon: "ph-bold ph-database",
    creationOptions: [
      DATA_SOURCE_CREATION_OPTIONS.NFS,
      DATA_SOURCE_CREATION_OPTIONS.PVC,
      DATA_SOURCE_CREATION_OPTIONS.GIT,
      DATA_SOURCE_CREATION_OPTIONS.HOST_PATH,
      DATA_SOURCE_CREATION_OPTIONS.CONFIG_MAP,
    ],
    loadFunction: loadDataSources,
  },
};
