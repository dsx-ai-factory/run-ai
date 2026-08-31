/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { credentialTextForUI, type TCredentialKinds } from "@/models/credential.model";
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
  type AssetClusterStatusInfo,
  type AssetRef,
  type CredentialsListResponseEntry,
} from "@/swagger-models/assets-service-client";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { dateUtil } from "@/utils/date.util";
import { getFieldContent } from "@/utils/table-format.util";

export const allCredentialColumnsMap: Record<string, ITableColumn> = {
  credential: {
    name: "credentials",
    label: "Credential",
    field: (row) => row.meta.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.credentials((cred) => cred.meta.name),
    },
  },
  description: {
    name: "description",
    label: "Description",
    field: (row: CredentialsListResponseEntry) => row.meta.description,
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
    field: (row: CredentialsListResponseEntry) => row.meta.kind,
    sortable: true,
    align: "left",
    format: (val: TCredentialKinds): string => {
      return credentialTextForUI[val];
    },
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.credentials(
        (cred) => credentialTextForUI[cred.meta.kind as TCredentialKinds],
      ),
    },
  },
  status: {
    name: "status",
    label: "Status",
    field: (row: CredentialsListResponseEntry): AssetClusterStatusInfo | undefined => row.status || undefined,
    format: (statusList: AssetClusterStatusInfo | undefined, row: CredentialsListResponseEntry) => {
      return assetsUtil.getStatusColOptions(statusList, row.meta.createdAt);
    },
    sort: (a, b, rowA, rowB) => {
      const valA: IStatusColOptions = getFieldContent(allCredentialColumnsMap.status, rowA);
      const valB: IStatusColOptions = getFieldContent(allCredentialColumnsMap.status, rowB);
      return valA.status.toLowerCase().localeCompare(valB.status.toLowerCase());
    },
    customCell: ECustomCell.STATUS_COL,
    align: "left",
    sortable: true,
  },
  dataSource: {
    name: "data-sources",
    label: "Data source(s)",
    field: (row: CredentialsListResponseEntry) => row.usedBy?.assets?.datasources,
    format: (dataSources: Array<AssetRef>): string[] => {
      if (!dataSources?.length) return [];
      return dataSources.map((ds: AssetRef) => ds.name);
    },
    sortable: true,
    align: "left",
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "data-source-clicked" },
  },
  environment: {
    name: "environments",
    label: "Environment(s)",
    field: (row: CredentialsListResponseEntry) => row.usedBy?.assets?.environments,
    format: (environments: Array<AssetRef>): string[] => {
      if (!environments?.length) return [];
      return environments.map((env: AssetRef) => env.name);
    },
    sortable: true,
    align: "left",
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "environments-clicked" },
  },
  createdBy: {
    name: "created-by",
    label: "Created by",
    field: (row: CredentialsListResponseEntry) => row.meta.createdBy,
    sortable: true,
    align: "left",
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.credentials((cred) => cred.meta.createdBy),
    },
  },
  clusterResourceName: {
    name: "cluster-resource-name",
    label: "Kubernetes name",
    field: (row: CredentialsListResponseEntry) => row.clusterInfo?.resources?.[0]?.name || "-",
    sortable: true,
    align: "left",
  },
  scope: {
    name: "scope",
    label: "Scope",
    field: (row: CredentialsListResponseEntry) => row.meta,
    sortable: true,
    align: "left",
    format: (scope: Scope, credential: CredentialsListResponseEntry): string | number => {
      switch (scope) {
        case Scope.Project:
          return Number(credential.meta.projectId);
        case Scope.Department:
          return String(credential.meta.departmentId);
        case Scope.Cluster:
          return String(credential.meta.clusterId);
        default:
          return String(credential.meta.tenantId);
      }
    },
    customCell: ECustomCell.SCOPE_COL,
    hideFilter: true,
  },
  createdAt: {
    name: "created-at",
    label: "Creation time",
    field: (row: CredentialsListResponseEntry) => row.meta.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string): string => dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: "Cluster",
    field: (row: CredentialsListResponseEntry) => row.meta.clusterId,
    sortable: true,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
    ignoreFilter: (row: CredentialsListResponseEntry) => {
      // ignoring this column filter if the scope is tenant (as clusterId is undefined)
      return !row.meta.clusterId && row.meta.scope === Scope.Tenant;
    },
  },
};

export const allCredentialColumns: Array<ITableColumn> = [
  allCredentialColumnsMap.credential,
  allCredentialColumnsMap.description,
  allCredentialColumnsMap.type,
  allCredentialColumnsMap.status,
  allCredentialColumnsMap.dataSource,
  allCredentialColumnsMap.environment,
  allCredentialColumnsMap.createdBy,
  allCredentialColumnsMap.scope,
  allCredentialColumnsMap.createdAt,
  allCredentialColumnsMap.cluster,
];

export const credentialIndexColumns: Array<ITableColumn> = [
  { ...allCredentialColumnsMap.credential, display: true, mandatory: true },
  { ...allCredentialColumnsMap.description, display: false },
  { ...allCredentialColumnsMap.type, display: true },
  { ...allCredentialColumnsMap.status, display: true },
  { ...allCredentialColumnsMap.scope, display: true },
  { ...allCredentialColumnsMap.clusterResourceName, display: false },
  { ...allCredentialColumnsMap.dataSource, display: true },
  { ...allCredentialColumnsMap.environment, display: true },
  { ...allCredentialColumnsMap.createdBy, display: true },
  { ...allCredentialColumnsMap.createdAt, display: true },
  { ...allCredentialColumnsMap.cluster, display: true },
];
