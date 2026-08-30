import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import { Scope, type AssetClusterStatusInfo } from "@/swagger-models/assets-service-client";
import type { ITableColumn, IStatusColOptions } from "@/models/table.model";
import { ALL_CLUSTER_FILTER_LABEL } from "@/models/filter.model";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { previewStrings } from "@/utils/common.util";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type { Tenant } from "@/swagger-models/tenants-manager-client";

interface IScopeEntity {
  meta: {
    scope?: Scope;
    scopeType?: Scope;
    projectId?: number | string | null;
    departmentId?: string | null;
    clusterId?: string | null;
  };
}

export const tableDisplayFormatters = {
  getScopeName(row: IScopeEntity, tenant: Tenant, orgUnitList: Array<OrgUnit>): string {
    const scope = row.meta.scope || row.meta.scopeType;
    const scopeId = this.getScopeId(row, tenant);
    const isSystemScope = (type: string) => scope === Scope.System && type === Scope.Tenant;

    return (
      orgUnitList.find(
        (orgUnit: OrgUnit) => orgUnit.id === scopeId && (orgUnit.type === scope || isSystemScope(orgUnit.type)),
      )?.name || ""
    );
  },

  getScopeId(row: IScopeEntity, tenant: Tenant): string {
    const scope = row.meta.scope || row.meta.scopeType;

    switch (scope) {
      case Scope.Department:
        return String(row.meta.departmentId);
      case Scope.Project:
        return String(row.meta.projectId);
      case Scope.Cluster:
        return String(row.meta.clusterId);
      default:
        return String(tenant.id);
    }
  },

  getClusterName<T extends object = object>(
    clusterUuid: string | undefined | null,
    clusterList: Array<ClustersMinimalData>,
    column?: ITableColumn,
    row?: T,
  ): string {
    if (!clusterUuid) {
      if (column?.ignoreFilter && row && column.ignoreFilter(row)) {
        return ALL_CLUSTER_FILTER_LABEL;
      }
      return "-";
    }

    const clusterName = clusterList.find((c) => c.uuid === clusterUuid)?.name;
    return clusterName || "-";
  },

  getStatusText(statusOptions: IStatusColOptions): string {
    return statusOptions.status || "-";
  },

  getStatusOptions(status: AssetClusterStatusInfo | undefined | null, createdAt: string): IStatusColOptions {
    return assetsUtil.getStatusColOptions(status || undefined, createdAt);
  },

  getListPreview(list: Array<string> | string): string {
    return typeof list === "string" ? list : previewStrings(list);
  },
};
