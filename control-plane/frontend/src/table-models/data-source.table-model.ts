import { DATA_SOURCE_TYPE, type TDataSourceKinds } from "@/models/data-source.model";
import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util";
import {
  ECustomCell,
  USER_DISPLAY_NAME_COL_DEFAULTS,
  type IStatusColOptions,
  type ITableColumn,
} from "@/models/table.model";
import {
  Scope,
  type DatasourceListResponseEntry,
  type AssetClusterStatusInfo,
} from "@/swagger-models/assets-service-client";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { dateUtil } from "@/utils/date.util";
import { getFieldContent } from "@/utils/table-format.util";

export const allDataSourceColumnsMap: Record<string, ITableColumn> = {
  dataSource: {
    name: "data-source",
    label: "Data source",
    field: (row) => row.meta.name,
    sortable: true,
    align: "left",
    customCell: ECustomCell.NAME_COL,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.dataSources((ds) => ds.meta.name),
    },
  },
  description: {
    name: "description",
    label: "Description",
    field: (row: DatasourceListResponseEntry) => row.meta.description,
    format: (val: string | null | undefined) => val || "-",
    sortable: true,
    align: "left",
    display: false,
    customCell: ECustomCell.DESCRIPTION_COL,
    sanitizeCell: true,
  },
  type: {
    name: "type",
    label: "Type",
    field: (row) => row.meta.kind,
    format: (kind: TDataSourceKinds) => DATA_SOURCE_TYPE[kind] || kind,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.dataSources((ds) => DATA_SOURCE_TYPE[ds.meta.kind as TDataSourceKinds]),
    },
  },
  status: {
    name: "status",
    label: "Status",
    field: (row: DatasourceListResponseEntry): AssetClusterStatusInfo | undefined => row.status || undefined,
    format: (statusList: AssetClusterStatusInfo | undefined, row: DatasourceListResponseEntry) => {
      return assetsUtil.getStatusColOptions(statusList, row.meta.createdAt);
    },
    sort: (a, b, rowA, rowB) => {
      const valA: IStatusColOptions = getFieldContent(allDataSourceColumnsMap.status, rowA);
      const valB: IStatusColOptions = getFieldContent(allDataSourceColumnsMap.status, rowB);
      return valA.status.toLowerCase().localeCompare(valB.status.toLowerCase());
    },
    customCell: ECustomCell.STATUS_COL,
    align: "left",
    sortable: true,
  },
  scope: {
    name: "scope",
    label: "Scope",
    field: (dataSource: DatasourceListResponseEntry): Scope => dataSource.meta.scope,
    sortable: true,
    align: "left",
    format: (scope: Scope, dataSource: DatasourceListResponseEntry): string | number => {
      switch (scope) {
        case Scope.Project:
          return Number(dataSource.meta.projectId);
        case Scope.Department:
          return String(dataSource.meta.departmentId);
        case Scope.Cluster:
          return String(dataSource.meta.clusterId);
        default:
          return String(dataSource.meta.tenantId);
      }
    },
    exportFormat: (scope: Scope): Scope => scope,
    customCell: ECustomCell.SCOPE_COL,
    hideFilter: true,
  },
  clusterResourceName: {
    name: "cluster-resource-name",
    label: "Kubernetes name",
    field: (row: DatasourceListResponseEntry) => row.clusterInfo?.resources?.[0]?.name || "-",
    sortable: true,
    align: "left",
  },
  createdBy: {
    name: "created-by",
    label: "Created by",
    field: (row) => row.meta.createdBy,
    sortable: true,
    align: "left",
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.dataSources((ds) => ds.meta.createdBy),
    },
  },
  createdAt: {
    name: "created-at",
    label: "Creation time",
    field: (row) => row.meta.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string) => dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: "Cluster",
    field: (row) => row.meta.clusterId,
    sortable: true,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
    ignoreFilter: (row: DatasourceListResponseEntry) => {
      // ignoring this column filter if the scope is system or tenant (as clusterId is undefined)
      return !row.meta.clusterId && (row.meta.scope === Scope.System || row.meta.scope === Scope.Tenant);
    },
  },
};

export const allDataSourceColumns: Array<ITableColumn> = [
  allDataSourceColumnsMap.dataSource,
  allDataSourceColumnsMap.description,
  allDataSourceColumnsMap.type,
  allDataSourceColumnsMap.status,
  allDataSourceColumnsMap.scope,
  allDataSourceColumnsMap.createdBy,
  allDataSourceColumnsMap.createdAt,
  allDataSourceColumnsMap.cluster,
];

export const dataSourceIndexColumns: Array<ITableColumn> = [
  { ...allDataSourceColumnsMap.dataSource, display: true, mandatory: true },
  { ...allDataSourceColumnsMap.description, display: false },
  { ...allDataSourceColumnsMap.type, display: true },
  { ...allDataSourceColumnsMap.status, display: true },
  { ...allDataSourceColumnsMap.scope, display: true },
  { ...allDataSourceColumnsMap.clusterResourceName, display: true },
  { ...allDataSourceColumnsMap.createdBy, display: true },
  { ...allDataSourceColumnsMap.createdAt, display: true },
  { ...allDataSourceColumnsMap.cluster, display: true },
];

export const dataSourceMiniTableColumns: Array<ITableColumn> = [
  { ...allDataSourceColumnsMap.dataSource, display: true },
  { ...allDataSourceColumnsMap.type, display: true },
];
