/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { type Node, NodeFilterSortFields, NodeStatusEnum } from "@/swagger-models/cluster-service-client";
import { ECustomCell, type IStatusColOptions, type ITableColumn } from "@/models/table.model";
import { viewLinkColumn } from "@/table-models/common-columns";
import { memoryFormat } from "@/utils/format.util";
import { tableFormatUtil } from "@/utils/table-format.util/table-format.util";
import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType } from "@/models/filter.model";
import { nodeUtil } from "@/utils/node.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";

export const formatMemoryValue = (val: number | undefined): string =>
  memoryFormat(tableFormatUtil.getFormattedNumberOrDefault(val));

const formatReadyTotalGpu = (readyGpus?: number, totalGpus?: number): string[] | string => {
  if (readyGpus === undefined || totalGpus === undefined) {
    return "-";
  }
  return [`${readyGpus}/${totalGpus}`];
};

export interface IAdditionalNodeTableFields {
  totalGpus?: number;
  readyGpus?: number;
  freeGpus?: number;
  totalGpuMemoryBytes?: number;
  allocatedGpus?: number;
  totalCpuCores?: number;
  totalCpuMemoryBytes?: number;
  allocatedCpuCores?: number;
  allocatedCpuMemoryBytes?: number;
}

export interface INodeTable extends Node {
  additionalFields?: IAdditionalNodeTableFields;
}

export const nodeColumnName = {
  name: NodeFilterSortFields.Name,
  status: NodeFilterSortFields.Status,
  nodePool: NodeFilterSortFields.NodePool,
  cluster: CLUSTER_COLUMN_FILTER_NAME,
  nvLinkDomainUid: NodeFilterSortFields.NvLinkDomainUid,
  nvLinkCliqueId: NodeFilterSortFields.NvLinkCliqueId,
  topologyManagerPolicy: "topologyManagerPolicy",
  gpuType: "gpuType",
  gpuCount: "gpuCount",
  freeGpu: "freeGpu",
  totalGpuMemoryBytes: "totalGpuMemoryBytes",
  allocatedGpus: "allocatedGpus",
  totalCpuCores: "totalCpuCores",
  totalCpuMemoryBytes: "TotalCpuMemoryBytes",
  allocatedCpuCores: "allocatedCpuCores",
  allocatedCpuMemoryBytes: "allocatedCpuMemoryBytes",
  pods: "pods",
  readyTotalGPU: "readyTotalGPU",
  numaNodes: "numaNodes",
};

enum ENodeColumnLabel {
  Name = "Node",
  GpuType = "GPU type",
  GpuCount = "Total GPU devices",
  TotalGpuMemoryBytes = "GPU memory",
  AllocatedGpus = "Allocated GPUs",
  TotalCpuCores = "CPU (Cores)",
  TotalCpuMemoryBytes = "CPU memory",
  AllocatedCpuCores = "Allocated CPU (Cores)",
  AllocatedCpuMemoryBytes = "Allocated CPU memory",
  Status = "Status",
  NodePool = "Node Pool",
  Cluster = "Cluster",
  NvLinkDomainUid = "MNNVL Domain UID",
  NvLinkCliqueId = "MNNVL domain clique ID",
  TopologyManagerPolicy = "Topology Manager policy",
  ReadyTotalGPU = "Ready / total GPU devices",
  FreeGpu = "Fully free GPU devices",
  AllocatedGpu = "Allocated GPU",
  Pods = "Pod(s)",
  NumaNodes = "NUMA nodes",
}

const nodeV2IndexColumns: Record<string, ITableColumn> = {
  /*** clusterServiceApi.getNodes columns ***/
  name: {
    name: nodeColumnName.name,
    label: ENodeColumnLabel.Name,
    field: (node: INodeTable): string => node.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.nodes(NodeFilterSortFields.Name, (n) => n.name),
    },
  },
  nodePool: {
    name: nodeColumnName.nodePool,
    label: ENodeColumnLabel.NodePool,
    field: (node: INodeTable): string => node.nodePool || "-",
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.nodes(NodeFilterSortFields.NodePool, (n) => n.nodePool),
    },
  },
  status: {
    name: nodeColumnName.status,
    label: ENodeColumnLabel.Status,
    field: (node: INodeTable): Node => node,
    sortable: true,
    align: "left",
    exportFormat: (node: INodeTable) => node.status || "-",
    customCell: ECustomCell.STATUS_COL,
    format: (node: INodeTable): IStatusColOptions => {
      return nodeUtil.getStatusColOptions(node);
    },
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: filterUtil.parseEnumToSelectOptions(NodeStatusEnum),
    },
  },
  gpuType: {
    name: nodeColumnName.gpuType,
    label: ENodeColumnLabel.GpuType,
    field: (node: INodeTable): string => node.gpuInfo?.gpuType || "-",
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.nodes(NodeFilterSortFields.GpuType, (n) => n.gpuInfo?.gpuType),
    },
  },
  totalGpus: {
    name: nodeColumnName.gpuCount,
    label: ENodeColumnLabel.GpuCount,
    field: (node: INodeTable): string => tableFormatUtil.getFormattedNumberOrDefault(node.gpuInfo?.gpuCount),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  nvLinkDomainUid: {
    name: nodeColumnName.nvLinkDomainUid,
    label: ENodeColumnLabel.NvLinkDomainUid,
    field: (node: INodeTable): string => node.nvLinkDomainUid || "-",
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.nodes(
        NodeFilterSortFields.NvLinkDomainUid,
        (n) => n.nvLinkDomainUid ?? undefined,
      ),
    },
  },
  nvLinkCliqueId: {
    name: nodeColumnName.nvLinkCliqueId,
    label: ENodeColumnLabel.NvLinkCliqueId,
    field: (node: INodeTable): string => node.nvLinkCliqueId || "-",
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.nodes(
        NodeFilterSortFields.NvLinkCliqueId,
        (n) => n.nvLinkCliqueId ?? undefined,
      ),
    },
  },
  topologyManagerPolicy: {
    name: nodeColumnName.topologyManagerPolicy,
    label: ENodeColumnLabel.TopologyManagerPolicy,
    field: (node: INodeTable): string => node.numa?.topologyManagerPolicy || "-",
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  cluster: {
    name: nodeColumnName.cluster,
    label: ENodeColumnLabel.Cluster,
    field: (row: Node): string => row.clusterUuid,
    sortable: false,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
  },
  /*** nodes telemetry columns ***/
  totalGpuMemoryBytes: {
    name: nodeColumnName.totalGpuMemoryBytes,
    label: ENodeColumnLabel.TotalGpuMemoryBytes,
    field: (node: INodeTable): string => formatMemoryValue(node.additionalFields?.totalGpuMemoryBytes),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedGpus: {
    name: nodeColumnName.allocatedGpus,
    label: ENodeColumnLabel.AllocatedGpus,
    field: (node: INodeTable): string =>
      tableFormatUtil.getFormattedNumberOrDefault(node.additionalFields?.allocatedGpus),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  totalCpuCores: {
    name: nodeColumnName.totalCpuCores,
    label: ENodeColumnLabel.TotalCpuCores,
    field: (node: INodeTable): string =>
      tableFormatUtil.getFormattedNumberOrDefault(node.additionalFields?.totalCpuCores),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  totalCpuMemoryBytes: {
    name: nodeColumnName.totalCpuMemoryBytes,
    label: ENodeColumnLabel.TotalCpuMemoryBytes,
    field: (node: INodeTable): string => formatMemoryValue(node.additionalFields?.totalCpuMemoryBytes),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedCpuCores: {
    name: nodeColumnName.allocatedCpuCores,
    label: ENodeColumnLabel.AllocatedCpuCores,
    field: (node: INodeTable): string =>
      tableFormatUtil.getFormattedNumberOrDefault(node.additionalFields?.allocatedCpuCores),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedCpuMemoryBytes: {
    name: nodeColumnName.allocatedCpuMemoryBytes,
    label: ENodeColumnLabel.AllocatedCpuMemoryBytes,
    field: (node: INodeTable): string => formatMemoryValue(node.additionalFields?.allocatedCpuMemoryBytes),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  readyTotalGPU: {
    name: nodeColumnName.readyTotalGPU,
    label: ENodeColumnLabel.ReadyTotalGPU,
    field: (node: INodeTable): string[] | string =>
      formatReadyTotalGpu(node.additionalFields?.readyGpus, node.additionalFields?.totalGpus),
    sortable: false,
    align: "left",
    hideFilter: true,
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "gpus-clicked" },
    exportFormat: (node: INodeTable): string => {
      const readyGpus = node.additionalFields?.readyGpus;
      const totalGpus = node.additionalFields?.totalGpus;
      if (readyGpus === undefined || totalGpus === undefined) {
        return "-";
      }
      return `${readyGpus}/${totalGpus}`;
    },
  },
  freeGpu: {
    name: nodeColumnName.freeGpu,
    label: ENodeColumnLabel.FreeGpu,
    field: (node: INodeTable): string | number => node.additionalFields?.freeGpus ?? "-",
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  pods: viewLinkColumn({
    name: nodeColumnName.pods,
    label: ENodeColumnLabel.Pods,
    emitName: "pods-clicked",
  }),
  numaNodes: viewLinkColumn({
    name: nodeColumnName.numaNodes,
    label: ENodeColumnLabel.NumaNodes,
    emitName: "numa-nodes-clicked",
  }),
};

export const nodeModalColumns: ITableColumn[] = [
  { ...nodeV2IndexColumns.name },
  { ...nodeV2IndexColumns.gpuType },
  { ...nodeV2IndexColumns.totalGpus },
  { ...nodeV2IndexColumns.totalGpuMemoryBytes },
  { ...nodeV2IndexColumns.allocatedGpus },
  { ...nodeV2IndexColumns.totalCpuCores },
  { ...nodeV2IndexColumns.totalCpuMemoryBytes },
  { ...nodeV2IndexColumns.allocatedCpuCores },
  { ...nodeV2IndexColumns.allocatedCpuMemoryBytes },
];

export const nodeIndexColumns: ITableColumn[] = [
  { ...nodeV2IndexColumns.name, display: true, mandatory: true },
  { ...nodeV2IndexColumns.status, display: true },
  { ...nodeV2IndexColumns.nodePool, display: true },
  { ...nodeV2IndexColumns.cluster, display: true },
  { ...nodeV2IndexColumns.nvLinkDomainUid, display: false },
  { ...nodeV2IndexColumns.nvLinkCliqueId, display: false },
  { ...nodeV2IndexColumns.topologyManagerPolicy, display: false },
  { ...nodeV2IndexColumns.numaNodes, display: false },
  { ...nodeV2IndexColumns.gpuType, display: true },
  { ...nodeV2IndexColumns.readyTotalGPU, display: true },
  { ...nodeV2IndexColumns.totalGpuMemoryBytes, display: true },
  { ...nodeV2IndexColumns.allocatedGpus, display: true },
  { ...nodeV2IndexColumns.freeGpu, display: true },
  { ...nodeV2IndexColumns.totalCpuCores, display: true },
  { ...nodeV2IndexColumns.totalCpuMemoryBytes, display: true },
  { ...nodeV2IndexColumns.allocatedCpuCores, display: true },
  { ...nodeV2IndexColumns.allocatedCpuMemoryBytes, display: true },
  { ...nodeV2IndexColumns.pods, display: true },
];

export const nodeDependentColumns = {
  nodePool: new Set([nodeColumnName.nodePool]), //only shown when there more than one node pool
  cluster: new Set([nodeColumnName.cluster]), //only shown when there more than one cluster
  numa: new Set([nodeColumnName.topologyManagerPolicy, nodeColumnName.numaNodes]), //only shown when at least one cluster supports NUMA-aware scheduling
};
