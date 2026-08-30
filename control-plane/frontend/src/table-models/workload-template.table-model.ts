import { CLUSTER_COLUMN_FILTER_NAME } from "@/models/filter.model";
import { ECustomCell, USER_DISPLAY_NAME_COL_DEFAULTS, type ITableColumn } from "@/models/table.model";
import { EWorkloadNodeArchitecture } from "@/models/workload.model";
import { Scope, type AssetRef, type WorkloadTemplate } from "@/swagger-models/assets-service-client";
import { dateUtil } from "@/utils/date.util";

export const allWorkloadTemplateColumnsMap: Record<string, ITableColumn> = {
  template: {
    name: "template",
    label: "Template",
    field: (row) => row.meta.name,
    sortable: true,
    align: "left",
    display: true,
    customCell: ECustomCell.NAME_COL,
  },
  description: {
    name: "description",
    label: "Description",
    field: (row: WorkloadTemplate) => row.meta.description,
    format: (val: string | null | undefined) => val || "-",
    sortable: true,
    align: "left",
    display: false,
    customCell: ECustomCell.DESCRIPTION_COL,
    sanitizeCell: true,
  },
  scope: {
    name: "scope",
    label: "Scope",
    field: (row: WorkloadTemplate) => row.meta.scope,
    sortable: true,
    align: "left",
    display: true,
    format: (scope: Scope, workloadTemplate: WorkloadTemplate): string | number => {
      switch (scope) {
        case Scope.Project:
          return Number(workloadTemplate.meta.projectId);
        case Scope.Department:
          return String(workloadTemplate.meta.departmentId);
        case Scope.Cluster:
          return String(workloadTemplate.meta.clusterId);
        default:
          return String(workloadTemplate.meta.tenantId);
      }
    },
    customCell: ECustomCell.SCOPE_COL,
    hideFilter: true,
  },
  environment: {
    name: "environment",
    label: "Environment",
    field: (row) => row.spec.assets?.environment?.name,
    sortable: true,
    align: "left",
    display: true,
  },
  multiNode: {
    name: "multi-node",
    label: "Workload architecture",
    field: (row: WorkloadTemplate) => row.meta.workloadSupportedTypes?.distributed,
    sortable: true,
    align: "left",
    format: (val: boolean | undefined) =>
      val ? EWorkloadNodeArchitecture.Distributed : EWorkloadNodeArchitecture.Standard,
    display: true,
  },
  computeResource: {
    name: "compute-resource",
    label: "Compute resource",
    field: (row) => row.spec.assets?.compute?.name,
    format: (name: string | undefined) => name || "-",
    sortable: true,
    align: "left",
    display: true,
  },
  dataSource: {
    name: "datasources",
    label: "Data source(s)",
    field: (row) => row.spec.assets?.datasources,
    format: (dataSources: Array<AssetRef>) => {
      if (!dataSources?.length) return [];
      return dataSources.map((ds) => ds.name);
    },
    sortable: true,
    align: "left",
    display: true,
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "data-source-clicked" },
  },
  createdBy: {
    name: "created-by",
    label: "Created by",
    field: (row) => row.meta.createdBy,
    sortable: true,
    align: "left",
    display: true,
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
  },
  createdAt: {
    name: "created-at",
    label: "Creation time",
    field: (row) => row.meta.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string) => dateUtil.dateAndTimeFormat(new Date(createdAt)),
    display: true,
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: "Cluster",
    field: (row) => row.meta.clusterId,
    sortable: true,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
    ignoreFilter: (row: WorkloadTemplate) => {
      // ignoring this column filter if the scope is system or tenant (as clusterId is undefined)
      return !row.meta.clusterId && (row.meta.scope === Scope.System || row.meta.scope === Scope.Tenant);
    },
  },
};

export const allWorkloadTemplateColumns: Array<ITableColumn> = [
  allWorkloadTemplateColumnsMap.template,
  allWorkloadTemplateColumnsMap.description,
  allWorkloadTemplateColumnsMap.scope,
  allWorkloadTemplateColumnsMap.environment,
  allWorkloadTemplateColumnsMap.multiNode,
  allWorkloadTemplateColumnsMap.computeResource,
  allWorkloadTemplateColumnsMap.dataSource,
  allWorkloadTemplateColumnsMap.createdBy,
  allWorkloadTemplateColumnsMap.createdAt,
  allWorkloadTemplateColumnsMap.cluster,
];

export const workloadTemplateIndexColumns: Array<ITableColumn> = [
  { ...allWorkloadTemplateColumnsMap.template, display: true, mandatory: true },
  { ...allWorkloadTemplateColumnsMap.description, display: false },
  { ...allWorkloadTemplateColumnsMap.scope, display: true },
  { ...allWorkloadTemplateColumnsMap.environment, display: true },
  { ...allWorkloadTemplateColumnsMap.multiNode, display: false },
  { ...allWorkloadTemplateColumnsMap.computeResource, display: true },
  { ...allWorkloadTemplateColumnsMap.dataSource, display: true },
  { ...allWorkloadTemplateColumnsMap.createdBy, display: true },
  { ...allWorkloadTemplateColumnsMap.createdAt, display: true },
  { ...allWorkloadTemplateColumnsMap.cluster, display: true },
];
