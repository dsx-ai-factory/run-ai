/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util";
import {
  ECustomCell,
  USER_DISPLAY_NAME_COL_DEFAULTS,
  type IStatusColOptions,
  type ITableColumn,
} from "@/models/table.model";
import { EWorkloadNodeArchitecture } from "@/models/workload.model";
import {
  Scope,
  type EnvironmentAsset,
  type WorkloadSupportedTypes,
  type Connection,
  InternalConnectionType,
  type AssetClusterStatusInfo,
} from "@/swagger-models/assets-service-client";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { dateUtil } from "@/utils/date.util";
import { getFieldContent } from "@/utils/table-format.util";
import { workloadUtil } from "@/utils/workload.util";
import { tableDisplayFormatters } from "@/utils/table-display-formatters.util";

export const allEnvironmentColumnsMap: Record<string, ITableColumn> = {
  environment: {
    name: "environment",
    label: "Environment",
    field: (row: EnvironmentAsset) => row.meta.name,
    sortable: true,
    align: "left",
    display: true,
    customCell: ECustomCell.NAME_COL,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.environments((env) => env.meta.name),
    },
  },
  description: {
    name: "description",
    label: "Description",
    field: (row: EnvironmentAsset) => row.meta.description,
    format: (val: string | null | undefined) => val || "-",
    sortable: true,
    align: "left",
    display: false,
    customCell: ECustomCell.DESCRIPTION_COL,
    sanitizeCell: true,
  },
  status: {
    name: "status",
    label: "Status",
    field: (row: EnvironmentAsset): AssetClusterStatusInfo | undefined => row.status || undefined,
    format: (statusList: AssetClusterStatusInfo | undefined, row: EnvironmentAsset) => {
      return assetsUtil.getStatusColOptions(statusList, row.meta.createdAt);
    },
    exportFormat: (row: EnvironmentAsset): string => {
      const statusOptions = tableDisplayFormatters.getStatusOptions(row.status, row.meta.createdAt);
      return tableDisplayFormatters.getStatusText(statusOptions);
    },
    sort: (a, b, rowA, rowB) => {
      const valA: IStatusColOptions = getFieldContent(allEnvironmentColumnsMap.status, rowA);
      const valB: IStatusColOptions = getFieldContent(allEnvironmentColumnsMap.status, rowB);
      return valA.status.toLowerCase().localeCompare(valB.status.toLowerCase());
    },
    customCell: ECustomCell.STATUS_COL,
    align: "left",
    sortable: true,
  },
  scope: {
    name: "scope",
    label: "Scope",
    field: (row: EnvironmentAsset) => row.meta.scope,
    sortable: true,
    align: "left",
    display: true,
    format: (scope: Scope, env: EnvironmentAsset): string | number => {
      switch (scope) {
        case Scope.Project:
          return Number(env.meta.projectId);
        case Scope.Department:
          return String(env.meta.departmentId);
        case Scope.Cluster:
          return String(env.meta.clusterId);
        default:
          return String(env.meta.tenantId);
      }
    },
    exportFormat: (row: EnvironmentAsset): string => {
      return row.meta?.scope || "-";
    },
    customCell: ECustomCell.SCOPE_COL,
    hideFilter: true,
  },
  image: {
    name: "image",
    label: "Image",
    field: (row: EnvironmentAsset) => row.spec.image,
    sortable: true,
    align: "left",
    display: true,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.environments((env) => env.spec.image || ""),
    },
  },
  multiNode: {
    name: "multi-node",
    label: "Workload architecture",
    field: (row: EnvironmentAsset) => row.meta.workloadSupportedTypes?.distributed,
    sortable: true,
    align: "left",
    format: (val: boolean | undefined) =>
      val ? EWorkloadNodeArchitecture.Distributed : EWorkloadNodeArchitecture.Standard,
    display: true,
  },
  tools: {
    name: "tools",
    label: "Tool(s)",
    field: (row: EnvironmentAsset) =>
      row.spec.connections?.filter((connection: Connection) => connection.name !== InternalConnectionType.ServingPort),
    sortable: true,
    align: "left",
    display: true,
    format: (connections: Array<Connection>): string[] => {
      if (!connections?.length) return [];
      return connections.map((con) => con.name);
    },
    exportFormat: (row: EnvironmentAsset): string => {
      const connections = row.spec.connections?.filter(
        (connection: Connection) => connection.name !== InternalConnectionType.ServingPort,
      );
      if (!connections?.length) return "-";
      const toolNames = connections.map((con) => con.name);
      return tableDisplayFormatters.getListPreview(toolNames);
    },
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "tool-clicked" },
  },
  workloadTypes: {
    name: "workload-types",
    label: "Workload types",
    field: (row) => row.meta.workloadSupportedTypes,
    sortable: true,
    align: "left",
    format: (workloadTypes: WorkloadSupportedTypes | undefined): string => {
      if (!workloadTypes) return "-";

      return workloadUtil.convertWorkloadTypesToString(workloadTypes);
    },
  },
  createdBy: {
    name: "created-by",
    label: "Created by",
    field: (row: EnvironmentAsset) => row.meta.createdBy,
    sortable: true,
    align: "left",
    display: false,
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.environments((env) => env.meta.createdBy),
    },
  },
  createdAt: {
    name: "created-at",
    label: "Creation time",
    field: (row: EnvironmentAsset) => row.meta.createdAt,
    sortable: true,
    align: "left",
    display: true,
    format: (createdAt: string) => dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
  lastUpdated: {
    name: "last-updated",
    label: "Last updated",
    field: (row: EnvironmentAsset) => row.meta.updatedAt,
    sortable: true,
    align: "left",
    display: false,
    format: (updatedAt: string) => dateUtil.dateAndTimeFormat(new Date(updatedAt)),
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: "Cluster",
    field: (row) => row.meta.clusterId,
    sortable: true,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
    ignoreFilter: (row: EnvironmentAsset) => {
      // ignoring this column filter if the scope is system or tenant (as clusterId is undefined)
      return !row.meta.clusterId && (row.meta.scope === Scope.System || row.meta.scope === Scope.Tenant);
    },
  },
};

export const allEnvironmentColumns: Array<ITableColumn> = [
  allEnvironmentColumnsMap.environment,
  allEnvironmentColumnsMap.description,
  allEnvironmentColumnsMap.status,
  allEnvironmentColumnsMap.scope,
  allEnvironmentColumnsMap.image,
  allEnvironmentColumnsMap.multiNode,
  allEnvironmentColumnsMap.tools,
  allEnvironmentColumnsMap.workloadTypes,
  allEnvironmentColumnsMap.createdBy,
  allEnvironmentColumnsMap.createdAt,
  allEnvironmentColumnsMap.lastUpdated,
  allEnvironmentColumnsMap.cluster,
];

export const environmentIndexColumns: Array<ITableColumn> = [
  { ...allEnvironmentColumnsMap.environment, display: true, mandatory: true },
  { ...allEnvironmentColumnsMap.description, display: false },
  { ...allEnvironmentColumnsMap.status, display: true },
  { ...allEnvironmentColumnsMap.scope, display: true },
  { ...allEnvironmentColumnsMap.image, display: true },
  { ...allEnvironmentColumnsMap.multiNode, display: false },
  { ...allEnvironmentColumnsMap.tools, display: true },
  { ...allEnvironmentColumnsMap.workloadTypes, display: false },
  { ...allEnvironmentColumnsMap.createdBy, display: true },
  { ...allEnvironmentColumnsMap.createdAt, display: true },
  { ...allEnvironmentColumnsMap.lastUpdated, display: false },
  { ...allEnvironmentColumnsMap.cluster, display: true },
];

export const environmentMiniTableColumns: Array<ITableColumn> = [
  { ...allEnvironmentColumnsMap.environment, display: true },
  { ...allEnvironmentColumnsMap.image, display: true },
];

export const allToolTableColumnsMap: Record<string, ITableColumn> = {
  toolName: {
    name: "tool-name",
    label: "Tool name",
    field: (row) => row.name,
    sortable: true,
    align: "left",
  },
  connectionType: {
    name: "connection-type",
    label: "Connection type",
    field: (row) => row,
    format: (c: Connection) => {
      return c.internalToolInfo?.connectionType || InternalConnectionType.ExternalUrl;
    },
    sortable: true,
    align: "left",
  },
};

export const toolTableColumns: Array<ITableColumn> = [
  {
    ...allToolTableColumnsMap.toolName,
    display: true,
  },
  {
    ...allToolTableColumnsMap.connectionType,
    display: true,
  },
];
