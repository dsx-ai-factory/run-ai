import { AssetKind, PvcVolumeMode } from "@/swagger-models/assets-service-client";
import type { Pvc, PVCAsset, PVCCreationRequest, Scope } from "@/swagger-models/assets-service-client";
import type { SupersetRulesAllOfStorage } from "@/swagger-models/policy-service-client";
import { makeId } from "../common.util";
import type { IUIVolume, TDataSourceKinds } from "@/models/data-source.model";
import type { IScopeCredentialsList } from "@/models/credential.model";
import { ScopeType } from "@/swagger-models/authorization-client";
import type { IUIAssetCreationMeta } from "@/models/global.model";

const DEFAULT_CLAIM_INFO_SIZE = "1G";

export const dataSourceUtil = {
  mapPvcsToUiVolumes,
  convertVolumeToPvc,
  getEmptyVolume,
  getRandomClaimName,
  getListScopeByMeta,
  isDataSource,
  isTDataSourceKinds,
  filterDataSourcesByPolicyRules,
};

function mapPvcsToUiVolumes(pvcs: Array<PVCAsset>): Array<IUIVolume> {
  return pvcs.map((volume: PVCAsset) => {
    return {
      claimName: "claim-" + makeId(10),
      path: volume.spec.path || "",
      ephemeral: volume.spec.ephemeral || false,
      claimInfo: volume.spec.claimInfo || {
        size: DEFAULT_CLAIM_INFO_SIZE,
        storageClassName: "",
        accessModes: {},
        volumeMode: PvcVolumeMode.Filesystem,
      },
    };
  });
}

export interface IScopeInfo {
  scope: Scope;
  projectId?: number;
  departmentId?: string;
  clusterId?: string;
}

function isDataSource(kind: string | undefined): boolean {
  if (!kind) return false;

  return new Set([
    AssetKind.HostPath.toString(),
    AssetKind.Nfs.toString(),
    AssetKind.Git.toString(),
    AssetKind.S3.toString(),
    AssetKind.Pvc.toString(),
    AssetKind.ConfigMap.toString(),
    AssetKind.SecretVolume.toString(),
  ]).has(kind);
}

function convertVolumeToPvc(volume: IUIVolume, scopeInfo: IScopeInfo): PVCCreationRequest {
  return {
    meta: {
      name: volume.claimName,
      scope: scopeInfo.scope,
      projectId: scopeInfo.projectId,
      departmentId: scopeInfo.departmentId,
      clusterId: scopeInfo.clusterId,
      autoDelete: true,
    },
    spec: {
      path: volume.path,
      claimName: volume.claimName,
      // Only include ephemeral if true, as false is the default and not supported for inference workloads
      ...(volume.ephemeral && { ephemeral: volume.ephemeral }),
      claimInfo: {
        accessModes: volume.claimInfo.accessModes,
        storageClass: volume.claimInfo.storageClass,
        size: volume.claimInfo.size,
        volumeMode: volume.claimInfo.volumeMode,
      },
    },
  };
}

function getRandomClaimName(prefix: string): string {
  // in k8s labels cannot be longer than 63 characters, because the claimName is used as a label in some instances its critical we limit its length
  const randomSuffix = `-${makeId(5)}`;
  const lengthLimit = 63 - randomSuffix.length;

  if (prefix.length > lengthLimit) prefix = prefix.slice(0, lengthLimit);

  return prefix + randomSuffix;
}

function getEmptyVolume(defaults?: Pvc | null): IUIVolume {
  return {
    claimName: getRandomClaimName("claim"),
    path: defaults?.path || "",
    ephemeral: defaults?.ephemeral || false,
    claimInfo: {
      size: defaults?.claimInfo?.size || DEFAULT_CLAIM_INFO_SIZE,
      storageClass: defaults?.claimInfo?.storageClass || null,
      accessModes: defaults?.claimInfo?.accessModes || {},
      volumeMode: defaults?.claimInfo?.volumeMode || null,
    },
  };
}

function getListScopeByMeta(assetMeta: IUIAssetCreationMeta): IScopeCredentialsList {
  const listScopeFilter: IScopeCredentialsList = {};
  if (assetMeta.scope === ScopeType.Tenant) {
    listScopeFilter.scope = ScopeType.Tenant;
  } else if (assetMeta.scope === ScopeType.Cluster && assetMeta.clusterId) {
    listScopeFilter.clusterUuid = assetMeta.clusterId;
  } else if (assetMeta.scope === ScopeType.Department && assetMeta.departmentId) {
    listScopeFilter.departmentId = assetMeta.departmentId;
  } else if (assetMeta.scope === ScopeType.Project && assetMeta.projectId) {
    listScopeFilter.projectId = assetMeta.projectId;
  }
  return listScopeFilter;
}

function isTDataSourceKinds(kind: AssetKind): kind is TDataSourceKinds {
  return [
    AssetKind.HostPath.toString(),
    AssetKind.Nfs.toString(),
    AssetKind.Git.toString(),
    AssetKind.S3.toString(),
    AssetKind.Pvc.toString(),
    AssetKind.ConfigMap.toString(),
    AssetKind.SecretVolume.toString(),
  ].includes(kind.toString());
}

function filterDataSourcesByPolicyRules<T extends AssetKind>(
  dsTypes: Set<T>,
  datasourcesRules?: SupersetRulesAllOfStorage | null,
): Set<T> {
  if (!datasourcesRules) {
    return dsTypes;
  }

  const allowedDsTypes = new Set(dsTypes);

  const dataSourcesToRulesMap: Partial<Record<AssetKind, keyof SupersetRulesAllOfStorage>> = {
    [AssetKind.Pvc]: "pvc",
    [AssetKind.Nfs]: "nfs",
    [AssetKind.S3]: "s3",
    [AssetKind.Git]: "git",
    [AssetKind.HostPath]: "hostPath",
    [AssetKind.SecretVolume]: "secretVolume",
    [AssetKind.ConfigMap]: "configMapVolume",
  };

  allowedDsTypes.forEach((kind: T) => {
    const ruleKey = dataSourcesToRulesMap[kind];
    const rule = ruleKey ? datasourcesRules[ruleKey] : undefined;
    if (rule && "instances" in rule && rule.instances?.canAdd === false) {
      allowedDsTypes.delete(kind);
    }
  });

  return allowedDsTypes;
}
