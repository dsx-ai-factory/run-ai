import { AssetKind } from "@/swagger-models/assets-service-client";
import type { IDrawerListItem } from "@/models/workload-creation-drawer-model";
import { DATA_SOURCE_TYPE } from "@/models/data-source.model";
import { MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH } from "@/common/version.constant";
import { useClusterStore } from "@/stores/cluster.store";

export const inferenceDataSourceRestrictionsUtil = {
  getDataSourceRestrictionInfo,
  applyInferenceDataSourceRestrictions,
};

interface IDataSourceRestrictionConfig {
  dataStorageSupported: boolean;
  modelStoreSupported: boolean; // Whether this type is supported in model store context (HF/NIM)
  requiresClusterVersion?: string;
  tooltipForUnsupportedVersion?: (type: string) => string;
  tooltipForUnsupportedModelStore?: (type: string) => string; // Tooltip to show when disabled in model store context
}

const tooltipForUnsupportedVersion = (type: string) =>
  `Inference workloads don't support data sources of type ${type} for the current version of the cluster.`;

const tooltipForUnsupportedModelStore = (type: string) =>
  `This setup cannot be selected because the Inference model store does not support data sources of type ${type}`;

const INFERENCE_DATA_SOURCE_RESTRICTIONS: Partial<Record<AssetKind, IDataSourceRestrictionConfig>> = {
  [AssetKind.Pvc]: {
    dataStorageSupported: true,
    modelStoreSupported: true,
  },
  [AssetKind.ConfigMap]: {
    dataStorageSupported: true,
    modelStoreSupported: false,
    tooltipForUnsupportedModelStore: tooltipForUnsupportedModelStore,
  },
  [AssetKind.Git]: {
    dataStorageSupported: true,
    modelStoreSupported: false, // Git is only supported in data source context, not in model store context (HF/NIM)
    tooltipForUnsupportedModelStore: tooltipForUnsupportedModelStore,
  },
  [AssetKind.HostPath]: {
    dataStorageSupported: true,
    modelStoreSupported: true,
    requiresClusterVersion: MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
    tooltipForUnsupportedVersion: tooltipForUnsupportedVersion,
  },
  [AssetKind.Nfs]: {
    dataStorageSupported: true,
    modelStoreSupported: true,
    requiresClusterVersion: MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
    tooltipForUnsupportedVersion: tooltipForUnsupportedVersion,
  },
  [AssetKind.DataVolume]: {
    dataStorageSupported: true,
    modelStoreSupported: true,
    requiresClusterVersion: MIN_CLUSTER_VERSION_FOR_NFS_AND_HOST_PATH,
    tooltipForUnsupportedVersion: tooltipForUnsupportedVersion,
  },
};

/**
 * Helper function to check if a data source type should be disabled for inference workloads
 * @param kind - The AssetKind to check
 * @param clusterId - cluster ID to check version requirements
 * @param isModelStoreContext - Whether this is a model store context (HF/NIM)
 * @returns Object with disabled flag and optional disabledInfo message
 */
function getDataSourceRestrictionInfo(
  kind: AssetKind,
  clusterId: string,
  isModelStoreContext = false,
): { disabled: boolean; disabledInfo?: string } {
  const clusterStore = useClusterStore();
  const restriction = INFERENCE_DATA_SOURCE_RESTRICTIONS[kind];
  const dataSourceType = DATA_SOURCE_TYPE[kind as keyof typeof DATA_SOURCE_TYPE] || kind;

  // If the data source type is not in our config, it's not supported for inference
  if (!restriction || !restriction.dataStorageSupported) {
    return {
      disabled: true,
      disabledInfo: `This setup cannot be selected because Inference workload does not support data sources of type ${dataSourceType}.`,
    };
  }

  // Check if this data source is not supported in model store context
  if (isModelStoreContext && !restriction.modelStoreSupported) {
    return {
      disabled: true,
      disabledInfo:
        restriction.tooltipForUnsupportedModelStore?.(dataSourceType) || tooltipForUnsupportedModelStore(dataSourceType),
    };
  }

  // Check cluster version requirements
  if (restriction.requiresClusterVersion) {
    const isClusterVersionSufficient = clusterStore.isClusterVersionSufficient(
      clusterId,
      restriction.requiresClusterVersion,
    );

    if (!isClusterVersionSufficient) {
      return {
        disabled: true,
        disabledInfo:
          restriction.tooltipForUnsupportedVersion?.(dataSourceType) || tooltipForUnsupportedVersion(dataSourceType),
      };
    }
  }

  // If all checks pass, return not disabled
  return { disabled: false };
}

/**
 * Applies restrictions to data source items for inference workloads
 * @param items - List of drawer items to process
 * @param clusterId - cluster ID to check version requirements
 * @param isModelStoreContext - Whether this is a model store context (HF/NIM) where Git is not supported (default: false)
 * @returns Modified list of items with disabled flags and tooltips applied
 */
function applyInferenceDataSourceRestrictions(
  items: IDrawerListItem[],
  clusterId: string,
  isModelStoreContext = false,
): IDrawerListItem[] {
  return items.map((item) => {
    // If the item doesn't have a kind, return it as is
    const kind = item.kind;
    if (!kind) {
      return item;
    }

    const restrictionInfo = getDataSourceRestrictionInfo(kind, clusterId, isModelStoreContext);

    if (restrictionInfo.disabled) {
      return {
        ...item,
        disabled: true,
        disabledInfo: restrictionInfo.disabledInfo,
      };
    }

    // If all checks pass, return the item as is
    return item;
  });
}
