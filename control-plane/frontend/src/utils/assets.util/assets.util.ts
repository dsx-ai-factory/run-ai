import type { IScopeModel, IUIAssetCreationMeta } from "@/models/global.model";
import type { IStatusColOptions } from "@/models/table.model";
import {
  type AssetCreationRequest,
  type AssetMeta,
  type WorkloadRefAndStatus,
  type AssetUsageInfo,
  AssetSyncStatusEnum,
  AssetClusterStatusIssueEnum,
  type AssetClusterStatusInfo,
  Scope,
  type HistoryRecord,
  HistoryRecordType,
} from "@/swagger-models/assets-service-client";
import { type PermittedScopes, ScopeType } from "@/swagger-models/authorization-client";
import { tableUtil } from "../table.util";
import type { PolicyMeta } from "@/swagger-models/policy-service-client";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import DOCS from "@/constants/docs";

export const assetsUtil = {
  filterById,
  prepareAssetMetaToCreationRequest,
  extractWorkloadNames,
  extractNames,
  setAssetScopeFromQuery,
  setAssetMetaFromQuery,
  getStatusColOptions,
  containerPathTitle,
  isPermittedAsset,
  getAssetScopeIdFromAssetMeta,
  hasLiveWorkload,
  getEmptyAssetMeta,
  isPhaseUpdateRecordType,
};

interface IAssetMeta {
  meta: AssetMeta | PolicyMeta;
}

function filterById<T extends IAssetMeta>(list: T[], entityToFilter: T | null): T[] {
  if (!entityToFilter) return list;
  return list.filter((asset: T) => asset.meta.id !== entityToFilter.meta.id);
}

function prepareAssetMetaToCreationRequest(uiModelMeta: IUIAssetCreationMeta): AssetCreationRequest | null {
  if (!uiModelMeta.scope) return null;
  if (uiModelMeta.scope !== Scope.Cluster) delete uiModelMeta.clusterId;
  return {
    ...uiModelMeta,
    scope: uiModelMeta.scope,
  };
}

function extractWorkloadNames(usageInfo: AssetUsageInfo): string[] {
  const workspaces = extractNames(usageInfo.workspaces || []);
  const trainings = extractNames(usageInfo.trainings || []);
  const inferences = extractNames(usageInfo.inferences || []);
  const distributed = extractNames(usageInfo.distributed || []);

  return [...workspaces, ...trainings, ...inferences, ...distributed];
}

function extractNames(workloadsRefs: Array<WorkloadRefAndStatus>): Array<string> {
  return workloadsRefs.map((workload: WorkloadRefAndStatus) => workload.name);
}

function setAssetScopeFromQuery(assetMeta: IUIAssetCreationMeta, query: IScopeModel) {
  assetMeta.scope = query.scope;
  if (query.scope === ScopeType.Cluster && query.clusterId) {
    assetMeta.clusterId = query.clusterId;
  } else if (query.scope === ScopeType.Department && query.departmentId) {
    assetMeta.departmentId = query.departmentId;
  } else if (query.scope === ScopeType.Project && query.projectId) {
    assetMeta.projectId = +query.projectId;
  }
  return assetMeta;
}

function getStatusColOptions(status: AssetClusterStatusInfo | undefined, createdAt: string): IStatusColOptions {
  if (!status) {
    return {
      status: "-",
      tooltipText: "",
      displayAnimation: false,
      filterKey: "status",
    };
  }

  return getAssetStatusOptions(status, createdAt);
}

function getAssetStatusOptions(status: AssetClusterStatusInfo, createdAt: string): IStatusColOptions {
  let colStatusOption: IStatusColOptions;

  if (status.issues) {
    const oneMinuteAgo = Date.now() - 60000;
    const createdAtTimestamp = Date.parse(createdAt);
    if (createdAtTimestamp > oneMinuteAgo) {
      colStatusOption = {
        status: "Creating...",
        displayAnimation: true,
        filterKey: "status",
      };
      return tableUtil.getStatusColOptions(colStatusOption);
    }
    if (status.issues.filter((i) => i.issue === AssetClusterStatusIssueEnum.NotCreated).length) {
      colStatusOption = {
        status: AssetSyncStatusEnum.IssuesFound,
        color: "negative",
        displayAnimation: false,
        filterKey: "status",
        tooltipText: status.message,
      };
    } else {
      colStatusOption = {
        status: AssetSyncStatusEnum.IssuesFound,
        color: "warning",
        displayAnimation: false,
        filterKey: "status",
        tooltipText: status.message,
      };
    }
  } else {
    switch (status.status) {
      case AssetSyncStatusEnum.Creating:
        colStatusOption = {
          status: `${AssetSyncStatusEnum.Creating}...`,
          displayAnimation: true,
          filterKey: "status",
          tooltipText: status.message,
        };
        break;
      case AssetSyncStatusEnum.NoIssuesFound:
        colStatusOption = {
          status: AssetSyncStatusEnum.NoIssuesFound,
          color: "success",
          displayAnimation: false,
          filterKey: "status",
        };
        break;
      case AssetSyncStatusEnum.Updating:
        colStatusOption = {
          status: `${AssetSyncStatusEnum.Updating}...`,
          displayAnimation: true,
          filterKey: "status",
          tooltipText: status.message,
        };
        break;
      case AssetSyncStatusEnum.IssuesFound:
        colStatusOption = {
          status: AssetSyncStatusEnum.IssuesFound,
          color: "negative",
          displayAnimation: false,
          filterKey: "status",
          tooltipText: status.message,
        };
        break;
      case AssetSyncStatusEnum.Deleting:
        colStatusOption = {
          status: `${AssetSyncStatusEnum.Deleting}...`,
          displayAnimation: true,
          filterKey: "status",
          tooltipText: status.message,
        };
        break;
      case AssetSyncStatusEnum.Failed:
        colStatusOption = {
          status: AssetSyncStatusEnum.IssuesFound,
          color: "negative",
          displayAnimation: false,
          filterKey: "status",
          tooltipText: status.message,
        };
        break;
      default:
        colStatusOption = {
          status: AssetSyncStatusEnum.NoStatus,
          displayAnimation: false,
          filterKey: "status",
          tooltipText: status.message,
        };
        break;
    }
  }

  if (status.url && colStatusOption.tooltipText) {
    if (DOCS.SUPPORT.URL) {
      colStatusOption.tooltipText += `. <a href="${DOCS.SUPPORT.URL}" target="_blank">Contact NVIDIA Run:ai support.</a>`;
    } else if (!/[.!?]$/.test(colStatusOption.tooltipText)) {
      colStatusOption.tooltipText += ".";
    }
  }

  return tableUtil.getStatusColOptions(colStatusOption);
}

function setAssetMetaFromQuery(
  assetMeta: IUIAssetCreationMeta,
  query: IScopeModel,
  orgUnits: Array<OrgUnit>,
): IUIAssetCreationMeta {
  const newMeta = setAssetScopeFromQuery(assetMeta, query as unknown as IScopeModel);
  if (!newMeta.clusterId) {
    const scopeId = getAssetScopeIdFromAssetMeta(newMeta);
    const parentClusterId = orgTreeService.getParentClusterUuid(
      orgUnits,
      scopeId,
      newMeta.scope as ScopeType, // assets scope includes "system" scope which isn't compatible with scopeType of authorization
    );
    newMeta.clusterId = parentClusterId;
  }
  return newMeta;
}

function containerPathTitle(showOverrideText?: boolean): string {
  let title = "Set the data target location";
  if (showOverrideText) {
    title += " (The workload creator will be able to override this)";
  }
  return title;
}

function isPermittedAsset(assetMeta: AssetMeta, permissions: PermittedScopes): boolean {
  const { scope, tenantId, clusterId, departmentId, projectId } = assetMeta;
  // Checking if the user has a permission to edit the selected compute resource
  switch (scope) {
    case Scope.Tenant:
      return !!permissions.tenants?.includes(tenantId?.toString() || "");
    case Scope.Cluster:
      return !!permissions.clusters?.includes(clusterId?.toString() || "");
    case Scope.Department:
      return !!permissions.departments?.includes(departmentId?.toString() || "");
    case Scope.Project:
      return !!permissions.projects?.includes(projectId?.toString() || "");
    default:
      return true;
  }
}

function getAssetScopeIdFromAssetMeta(assetMeta: IUIAssetCreationMeta): string {
  switch (assetMeta.scope) {
    case ScopeType.Cluster:
      return assetMeta.clusterId || "";
    case ScopeType.Department:
      return assetMeta.departmentId || "";
    case ScopeType.Project:
      return assetMeta.projectId ? String(assetMeta.projectId) : "";
    default:
      return "";
  }
}

function hasLiveWorkload(workloadRefs: Array<WorkloadRefAndStatus>): boolean {
  return workloadRefs.some((w: WorkloadRefAndStatus) => w.status !== "Failed" && w.status !== "Succeeded");
}

function getEmptyAssetMeta(): IUIAssetCreationMeta {
  return {
    name: "",
    description: null,
    scope: null,
    projectId: null,
    clusterId: null,
  };
}

function isPhaseUpdateRecordType(historyRecord: HistoryRecord): boolean {
  return historyRecord.meta.type === HistoryRecordType.PhaseUpdate;
}
