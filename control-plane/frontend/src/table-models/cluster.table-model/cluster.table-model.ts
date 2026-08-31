/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ECustomCell, type IStatusColOptions, type ITableColumn } from "@/models/table.model";
import { EColumnFilterType } from "@/models/filter.model";
import { viewLinkColumn } from "@/table-models/common-columns";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import type { ClusterReportedStatusPlatformTypeEnum } from "@/swagger-models/cluster-service-client";
import {
  ClusterState,
  type ClusterDisplayedStatus,
  type DisplayedCluster,
} from "@/swagger-models/cluster-service-client";
import { clusterUtil } from "@/utils/cluster.util";
import { dateUtil } from "@/utils/date.util";
import { isDateString } from "@/utils/date.util/date.util";
import { splitCamelCaseAndCapitalize } from "@/utils/string.util/string.util";
import { tableUtil } from "@/utils/table.util";
import compareVersions from "semver-compare";

export const allClusterColumnsMap: Record<string, ITableColumn> = {
  name: {
    name: "name",
    label: "Cluster",
    field: (row: DisplayedCluster) => row.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.clusters((c) => c.name),
    },
  },
  distribution: {
    name: "distribution",
    label: "Kubernetes distribution",
    field: (row: DisplayedCluster) => row.status?.platform?.type,
    sortable: true,
    align: "left",
    format: (distribution: ClusterReportedStatusPlatformTypeEnum) => clusterUtil.getDistribution(distribution) || "-",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.clusters((c) =>
        c.status?.platform?.type ? clusterUtil.getDistribution(c.status.platform.type) : undefined,
      ),
    },
  },
  distributionVersion: {
    name: "distributionVersion",
    label: "Kubernetes version",
    field: (row: DisplayedCluster) => row.status?.platform?.kubeVersion,
    sortable: true,
    sort: (a: string, b: string) => {
      return compareVersions(a, b);
    },
    align: "left",
    format: (version: string) => version?.replace("v", "") || "-",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.clusters((c) =>
        c.status?.platform?.kubeVersion ? c.status.platform.kubeVersion.replace("v", "") : undefined,
      ),
    },
  },
  status: {
    name: "status",
    label: "Status",
    field: (row: DisplayedCluster) => {
      return row.status;
    },
    sortable: true,
    align: "left",
    customCell: ECustomCell.STATUS_COL,
    customCellEvent: { emitName: "tooltip-click" },
    format: (status: ClusterDisplayedStatus): IStatusColOptions => {
      return clusterUtil.getStatusColOptions(status);
    },
    sort: (fieldA, fieldB, rowA, rowB) => {
      return tableUtil.customColumnSort(allClusterColumnsMap.status, rowA, rowB, (status: IStatusColOptions) => {
        return status.status;
      });
    },
    exportFormat: (row: DisplayedCluster): string =>
      splitCamelCaseAndCapitalize(row.status?.state as ClusterState) || "NA",
    filterKey: "status",
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: Object.values(ClusterState)
        .map((state) => ({
          label: splitCamelCaseAndCapitalize(state),
          value: splitCamelCaseAndCapitalize(state),
        }))
        .sort((a, b) => a.label.localeCompare(b.label)),
    },
  },
  lastConnected: {
    name: "lastConnected",
    label: "Last connected",
    field: (row: DisplayedCluster) => {
      return row.status?.state === ClusterState.Connected ? "Now" : row.lastLiveness ? row.lastLiveness : "Never";
    },
    sortable: true,
    align: "left",
    format: (lastConnected: string) =>
      isDateString(lastConnected) ? dateUtil.dateAndTimeFormat(new Date(lastConnected)) : lastConnected,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.clusters((c) => {
        if (c.status?.state === ClusterState.Connected) return "Now";
        if (c.lastLiveness)
          return isDateString(c.lastLiveness) ? dateUtil.dateAndTimeFormat(new Date(c.lastLiveness)) : c.lastLiveness;
        return "Never";
      }),
    },
  },
  topologies: viewLinkColumn({
    name: "topologies",
    label: "Network topologies",
    emitName: "topologies-clicked",
  }),
  uuid: {
    name: "uuid",
    label: "Run:ai cluster UUID",
    field: (row: DisplayedCluster) => row.uuid,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.clusters((c) => c.uuid),
    },
  },
  createdAt: {
    name: "created-at",
    label: "Creation time",
    field: (row: DisplayedCluster) => row.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string) => dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
  url: {
    name: "url",
    label: "URL",
    field: (row: DisplayedCluster) => row.domain,
    sortable: true,
    align: "left",
    format: (url: string) => url || "—",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.clusters((c) => c.domain ?? undefined),
    },
  },
  version: {
    name: "version",
    label: "NVIDIA Run:ai cluster version",
    field: (row: DisplayedCluster) => row.version,
    sortable: true,
    sort: (a: string, b: string) => {
      return compareVersions(a, b);
    },
    align: "left",
    format: (version: string) => version || "—",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.clusters((c) => c.version ?? undefined),
    },
  },
};

export const allClusterColumns: Array<ITableColumn> = [
  allClusterColumnsMap.name,
  allClusterColumnsMap.distribution,
  allClusterColumnsMap.distributionVersion,
  allClusterColumnsMap.status,
  allClusterColumnsMap.uuid,
  allClusterColumnsMap.createdAt,
  allClusterColumnsMap.url,
  allClusterColumnsMap.version,
  allClusterColumnsMap.lastConnected,
  allClusterColumnsMap.topologies,
];

export const clusterIndexColumns: Array<ITableColumn> = [
  { ...allClusterColumnsMap.name, display: true, mandatory: true },
  { ...allClusterColumnsMap.distribution, display: true },
  { ...allClusterColumnsMap.distributionVersion, display: true },
  { ...allClusterColumnsMap.status, display: true },
  { ...allClusterColumnsMap.lastConnected, display: true },
  { ...allClusterColumnsMap.topologies, display: false },
  { ...allClusterColumnsMap.createdAt, display: true },
  { ...allClusterColumnsMap.url, display: true },
  { ...allClusterColumnsMap.version, display: true },
  { ...allClusterColumnsMap.uuid, display: false },
];
