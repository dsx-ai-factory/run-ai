import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType } from "@/models/filter.model";
import {
  ECustomCell,
  USER_DISPLAY_NAME_COL_DEFAULTS,
  type IStatusColOptions,
  type ITableColumn,
} from "@/models/table.model";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import { ScopeType } from "@/swagger-models/authorization-client";
import {
  PolicyType,
  type PolicyListEntry,
  type PolicySyncStatus,
  type PolicySyncStatusOfClusterStatusEnum,
} from "@/swagger-models/policy-service-client";
import { dateUtil } from "@/utils/date.util";
import { policyUtil, policyStatusMap } from "@/utils/policy.util/policy.util";
import { policyModelUtil } from "@/utils/policy.util/policy-model.util";
import { capitalizeString } from "@/utils/string.util/string.util";

export const allPreviewPolicyColumnsMap: Record<string, ITableColumn> = {
  parameter: {
    name: "parameter",
    label: "Parameter",
    field: (row) => row.name,
    align: "left",
  },
  default: {
    name: "default",
    label: "Default",
    field: (row) => row.default,
    align: "left",
    format: (val) => policyModelUtil.policyDefaultToPreviewString(val),
    customCell: ECustomCell.HTML_COL,
  },
  rule: {
    name: "rule",
    label: "Rule",
    field: (row) => row.rules,
    align: "left",
    format: (val) => policyModelUtil.policyRulesToPreviewString(val),
    customCell: ECustomCell.HTML_COL,
  },
  source: {
    name: "source",
    label: "Source",
    field: (row) => row.sourceOfRule,
    align: "left",
    customCell: ECustomCell.SOURCE_NAME_COL,
  },
  type: {
    name: "type",
    label: "Type",
    field: (row) => row.type,
    align: "left",
  },
  replica: {
    name: "replica",
    label: "Replica",
    field: (row) => row.replica,
    align: "left",
    format: (val) => capitalizeString(val),
  },
};

export const workloadPreviewPolicyColumns: Array<ITableColumn> = [
  { ...allPreviewPolicyColumnsMap.parameter, display: true },
  { ...allPreviewPolicyColumnsMap.default, display: true },
  { ...allPreviewPolicyColumnsMap.rule, display: true },
  { ...allPreviewPolicyColumnsMap.source, display: true },
];

export const workloadPreviewPolicyColumnsDataSourceTable: Array<ITableColumn> = [
  { ...allPreviewPolicyColumnsMap.parameter, display: true },
  { ...allPreviewPolicyColumnsMap.type, display: true },
  { ...allPreviewPolicyColumnsMap.default, display: true },
  { ...allPreviewPolicyColumnsMap.rule, display: true },
  { ...allPreviewPolicyColumnsMap.source, display: true },
];

export const distributedPreviewPolicyColumns: Array<ITableColumn> = [
  { ...allPreviewPolicyColumnsMap.parameter, display: true },
  { ...allPreviewPolicyColumnsMap.replica, display: true },
  { ...allPreviewPolicyColumnsMap.default, display: true },
  { ...allPreviewPolicyColumnsMap.rule, display: true },
  { ...allPreviewPolicyColumnsMap.source, display: true },
];

export const distributedPreviewPolicyColumnsDataSourceTable: Array<ITableColumn> = [
  { ...allPreviewPolicyColumnsMap.parameter, display: true },
  { ...allPreviewPolicyColumnsMap.replica, display: true },
  { ...allPreviewPolicyColumnsMap.type, display: true },
  { ...allPreviewPolicyColumnsMap.default, display: true },
  { ...allPreviewPolicyColumnsMap.rule, display: true },
  { ...allPreviewPolicyColumnsMap.source, display: true },
];

export const gvkPreviewPolicyColumns: Array<ITableColumn> = [
  { name: "field", label: "Field", field: (row) => row.field, align: "left", display: true },
  { name: "specSelector", label: "Spec selector", field: (row) => row.specSelector, align: "left", display: true },
  { name: "itemSelector", label: "Item selector", field: (row) => row.itemSelector, align: "left", display: true },
  {
    name: "rule",
    label: "Rule",
    field: (row) => row.rule,
    align: "left",
    display: true,
    customCell: ECustomCell.HTML_COL,
  },
];

type IGvkAwarePolicyRow = PolicyListEntry & { isGvk?: boolean; gvkKind?: string };

export const allPolicyColumnsMap: Record<string, ITableColumn> = {
  policy: {
    name: "name",
    label: "Policy",
    field: (row: PolicyListEntry) => row.meta?.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.policies((p) => p.meta?.name),
    },
  },
  type: {
    name: "type",
    label: "Type",
    field: (row: PolicyListEntry) => row.type,
    sortable: true,
    align: "left",
    format: (type: PolicyType, policy?: IGvkAwarePolicyRow) => {
      if (policy?.isGvk) return policy.gvkKind ?? String(type ?? "");
      return type === PolicyType.Distributed ? capitalizeString(PolicyType.Training) : capitalizeString(type);
    },
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.policies((p) => {
        const row = p as IGvkAwarePolicyRow;
        if (row.isGvk) return row.gvkKind ?? "";
        return p.type === PolicyType.Distributed
          ? capitalizeString(PolicyType.Training)
          : capitalizeString(p.type || "");
      }),
    },
  },
  architecture: {
    name: "architecture",
    label: "Policy architecture",
    field: (row: PolicyListEntry) => row.type,
    sortable: true,
    align: "left",
    format: (type: PolicyType, policy?: IGvkAwarePolicyRow) => {
      if (policy?.isGvk) return "N/A";
      if (type === PolicyType.Distributed) {
        return capitalizeString(PolicyType.Distributed);
      }
      return "Standard";
    },
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.policies((p) =>
        (p as IGvkAwarePolicyRow).isGvk
          ? "N/A"
          : p.type === PolicyType.Distributed
            ? capitalizeString(PolicyType.Distributed)
            : "Standard",
      ),
    },
  },
  status: {
    name: "status",
    label: "Status",
    field: (row) => row.status,
    sortable: true,
    align: "left",
    format: (status: PolicySyncStatus | undefined): IStatusColOptions => {
      // Backend-verified: native + GVK rows share the unified per-cluster sync status.
      const currentClusterStatus: PolicySyncStatusOfClusterStatusEnum | undefined = status?.clusters?.[0]?.status;
      const message: string | undefined = status?.clusters?.[0]?.errorMessage || undefined;
      return policyUtil.getPolicyStatusColOptions(currentClusterStatus, message);
    },
    filterKey: "status",
    customCell: ECustomCell.STATUS_COL,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.policies((p) => {
        const clusterStatus = p.status?.clusters?.[0]?.status;
        if (!clusterStatus) return "-";
        return policyStatusMap[clusterStatus]?.status;
      }),
    },
  },
  scope: {
    name: "scope",
    label: "Scope",
    field: (row: PolicyListEntry) => row.meta?.scope,
    sortable: true,
    align: "left",
    format: (scope: ScopeType, policy: PolicyListEntry): string | number => {
      switch (scope) {
        case ScopeType.Project:
          return Number(policy.meta?.projectId);
        case ScopeType.Department:
          return String(policy.meta?.departmentId);
        case ScopeType.Cluster:
          return String(policy.meta?.clusterId);
        default:
          return String(policy.meta?.tenantId);
      }
    },
    customCell: ECustomCell.SCOPE_COL,
    hideFilter: true,
  },
  createdBy: {
    name: "created-by",
    label: "Created by",
    field: (row: PolicyListEntry) => row.meta?.createdBy,
    sortable: true,
    align: "left",
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.policies((p) => p.meta?.createdBy),
    },
  },
  createdAt: {
    name: "created-at",
    label: "Creation time",
    field: (row: PolicyListEntry) => row.meta?.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string) => dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
  lastUpdated: {
    name: "lastUpdated",
    label: "Last updated",
    field: (row: PolicyListEntry) => row.meta?.updatedAt,
    sortable: true,
    align: "left",
    display: true,
    format: (createdAt: string) => createdAt && dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: "Cluster",
    field: (row) => row.meta?.clusterId,
    sortable: true,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
  },
};

export const policyIndexColumns: Array<ITableColumn> = [
  { ...allPolicyColumnsMap.policy, display: true, mandatory: true },
  { ...allPolicyColumnsMap.type, display: true },
  { ...allPolicyColumnsMap.architecture, display: false },
  { ...allPolicyColumnsMap.status, display: true },
  { ...allPolicyColumnsMap.scope, display: true },
  { ...allPolicyColumnsMap.createdBy, display: true },
  { ...allPolicyColumnsMap.createdAt, display: true },
  { ...allPolicyColumnsMap.lastUpdated, display: true },
  { ...allPolicyColumnsMap.cluster, display: true },
];
