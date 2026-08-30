import { ECustomCell, type IStatusColOptions, type ITableColumn } from "@/models/table.model";
import { viewLinkColumn } from "@/table-models/common-columns";

import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType } from "@/models/filter.model";
import {
  type Nodepool,
  NodepoolGPUNetworkAccelerationDetection,
  NodepoolPhase,
  NodepoolSortFilterFields,
} from "@/swagger-models/cluster-service-client";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { tableFormatUtil } from "@/utils/table-format.util/table-format.util";
import { convertToBytes, memoryFormat } from "@/utils/format.util";
import { allWorkloadColumnsMap } from "@/table-models/workload.table-model";
import type { INodePoolsResourcesRow } from "@/models/project.model";
import { resourceUtil } from "@/utils/resource.util";
import { EResourceType } from "@/models/resource.model";
import { tableUtil } from "@/utils/table.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { dateUtil } from "@/utils/date.util";
import { NODE_POOL_NETWORK_DISCOVERY_OPTIONS } from "@/models/node-pool.model";

export interface IAdditionalNodePoolTableFields {
  totalGpus?: number;
  totalGpuMemoryBytes?: number;
  allocatedGpus?: number;
  allocatedGpuMemory?: number;
  totalCpuCores?: number;
  totalCpuMemoryBytes?: number;
  allocatedCpuCores?: number;
  allocatedCpuMemoryBytes?: number;
}

export interface INodePoolTable extends Nodepool {
  additionalFields?: IAdditionalNodePoolTableFields;
}

const _formatMemoryValue = (val: number | undefined): string =>
  memoryFormat(tableFormatUtil.getFormattedNumberOrDefault(val));

export const nodePoolColumnName = {
  name: NodepoolSortFilterFields.Name,
  phase: NodepoolSortFilterFields.Phase,
  cluster: CLUSTER_COLUMN_FILTER_NAME,
  labelKey: "labelKey",
  labelValue: "labelValue",
  nodes: "nodes",
  gpuNetworkAccelerationDetection: "gpuNetworkAccelerationDetection",
  gpuNetworkAccelerationDetected: "gpuNetworkAccelerationDetected",
  gpuNetworkAccelerationLabelKey: "gpuNetworkAccelerationLabelKey",
  minGuaranteedRuntime: "minGuaranteedRuntime",
  gpuPlacementStrategy: "gpuPlacementStrategy",
  gpuDevicePlacementStrategy: "gpuDevicePlacementStrategy",
  cpuPlacementStrategy: "cpuPlacementStrategy",
  updatedAt: NodepoolSortFilterFields.UpdatedAt,
  createdAt: NodepoolSortFilterFields.CreatedAt,
  workloads: "workloads",
  totalGpus: "totalGpus",
  totalGpuMemoryBytes: "totalGpuMemoryBytes",
  allocatedGpus: "allocatedGpus",
  totalCpuCores: "totalCpuCores",
  totalCpuMemoryBytes: "totalCpuMemoryBytes",
  allocatedCpuCores: "allocatedCpuCores",
  allocatedCpuMemoryBytes: "allocatedCpuMemoryBytes",
  topology: "topology",
};

enum ENodePoolColumnLabel {
  Name = "Node pool",
  Phase = "Status",
  Cluster = "Cluster",
  LabelKey = "Label key",
  LabelValue = "Label value",
  Nodes = "Node(s)",
  Topology = "Network topology",
  GpuNetworkAccelerationDetection = "MNNVL",
  GpuNetworkAccelerationDetected = "MNNVL nodes",
  GpuNetworkAccelerationLabelKey = "MNNVL label key",
  MinGuaranteedRuntime = "Minimum guaranteed runtime",
  GpuNodePlacementStrategy = "GPU node placement strategy",
  GpuDevicePlacementStrategy = "GPU device placement strategy",
  CpuNodePlacementStrategy = "CPU node placement strategy",
  UpdatedAt = "Last updated",
  CreatedAt = "Creation time",
  Workloads = "Workload(s)",
  TotalGpus = "Total GPU devices",
  TotalGpuMemoryBytes = "Total GPU memory",
  AllocatedGpus = "Allocated GPUs",
  TotalCpuCores = "Total CPU (Cores)",
  TotalCpuMemoryBytes = "Total CPU memory",
  AllocatedCpuCores = "Allocated CPU (Cores)",
  AllocatedCpuMemoryBytes = "Allocated CPU memory",
}

export const allNodePoolsColumnsMap: Record<keyof typeof nodePoolColumnName, ITableColumn> = {
  /*** clusterServiceApi.getNodepools columns ***/
  name: {
    name: nodePoolColumnName.name,
    label: ENodePoolColumnLabel.Name,
    field: (nodePool: INodePoolTable): string => nodePool.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.nodePoolsMinimal(),
    },
  },
  cluster: {
    name: nodePoolColumnName.cluster,
    label: ENodePoolColumnLabel.Cluster,
    field: (nodePool: INodePoolTable): string => nodePool.clusterId,
    sortable: false,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
  },
  phase: {
    name: nodePoolColumnName.phase,
    label: ENodePoolColumnLabel.Phase,
    field: (nodePool: INodePoolTable): INodePoolTable => nodePool,
    sortable: true,
    align: "left",
    customCell: ECustomCell.STATUS_COL,
    format: (nodePool: INodePoolTable): IStatusColOptions => nodePoolUtil.getStatusColOptions(nodePool),
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: filterUtil.parseEnumToSelectOptions(NodepoolPhase),
    },
    exportFormat: (nodePool: INodePoolTable): string => nodePool.phase || "-",
  },
  labelKey: {
    name: nodePoolColumnName.labelKey,
    label: ENodePoolColumnLabel.LabelKey,
    field: (nodePool: INodePoolTable): string => nodePool.labelKey,
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  labelValue: {
    name: nodePoolColumnName.labelValue,
    label: ENodePoolColumnLabel.LabelValue,
    field: (nodePool: INodePoolTable): string => nodePool.labelValue,
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  nodes: {
    name: nodePoolColumnName.nodes,
    label: ENodePoolColumnLabel.Nodes,
    field: (nodePool: INodePoolTable): string[] => nodePool.status?.nodes || [],
    sortable: false,
    align: "left",
    hideFilter: true,
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "nodes-clicked" },
  },
  topology: {
    name: nodePoolColumnName.topology,
    label: ENodePoolColumnLabel.Topology,
    field: (nodePool: INodePoolTable): string[] => (nodePool.networkTopologyName ? [nodePool.networkTopologyName] : []),
    sortable: false,
    align: "left",
    hideFilter: true,
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "topology-clicked" },
  },
  gpuNetworkAccelerationDetection: {
    name: nodePoolColumnName.gpuNetworkAccelerationDetection,
    label: ENodePoolColumnLabel.GpuNetworkAccelerationDetection,
    field: (nodePool: INodePoolTable): string => {
      return (
        NODE_POOL_NETWORK_DISCOVERY_OPTIONS.find((option) => option.value === nodePool.gpuNetworkAccelerationDetection)
          ?.label || "-"
      );
    },
    sortable: true,
    align: "left",
    filter: {
      maxWidth: "290px",
      type: EColumnFilterType.EnumBased,
      inputType: "radio",
      selectOptions: [
        {
          label: "Automatic",
          value: NodepoolGPUNetworkAccelerationDetection.Auto,
        },
        {
          label: "Manual",
          value: NodepoolGPUNetworkAccelerationDetection.Use,
        },
      ],
    },
  },
  gpuNetworkAccelerationLabelKey: {
    name: nodePoolColumnName.gpuNetworkAccelerationLabelKey,
    label: ENodePoolColumnLabel.GpuNetworkAccelerationLabelKey,
    field: (nodePool: INodePoolTable): string => nodePool?.gpuNetworkAccelerationLabelKey || "-",
    sortable: false,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.nodePools(
        NodepoolSortFilterFields.GpuNetworkAccelerationLabelKey,
        (np) => np.gpuNetworkAccelerationLabelKey,
      ),
    },
  },
  gpuNetworkAccelerationDetected: {
    name: nodePoolColumnName.gpuNetworkAccelerationDetected,
    label: ENodePoolColumnLabel.GpuNetworkAccelerationDetected,
    field: (nodePool: INodePoolTable): string =>
      nodePool?.status?.gpuNetworkAccelerationDetected ? "Detected" : "Not detected",
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.EnumBased,
      inputType: "radio",
      selectOptions: [
        {
          label: "Detected",
          value: true,
        },
        {
          label: "Not detected",
          value: false,
        },
      ],
    },
  },
  minGuaranteedRuntime: {
    name: nodePoolColumnName.minGuaranteedRuntime,
    label: ENodePoolColumnLabel.MinGuaranteedRuntime,
    field: (nodePool: INodePoolTable): string | undefined | null =>
      nodePool?.schedulingConfiguration?.minGuaranteedRuntime,
    format: (minGuaranteedRuntime: string): string => {
      if (!minGuaranteedRuntime) return "-";
      return dateUtil.formatDuration(dateUtil.parseDurationStringToSeconds(minGuaranteedRuntime), true);
    },
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  gpuPlacementStrategy: {
    name: nodePoolColumnName.gpuPlacementStrategy,
    label: ENodePoolColumnLabel.GpuNodePlacementStrategy,
    field: (nodePool: INodePoolTable): string => {
      return nodePool?.schedulingConfiguration?.gpuPodsPlacementStrategy?.node || "-";
    },
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  gpuDevicePlacementStrategy: {
    name: nodePoolColumnName.gpuDevicePlacementStrategy,
    label: ENodePoolColumnLabel.GpuDevicePlacementStrategy,
    field: (nodePool: INodePoolTable): string => {
      return nodePool?.schedulingConfiguration?.gpuPodsPlacementStrategy?.device || "-";
    },
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  cpuPlacementStrategy: {
    name: nodePoolColumnName.cpuPlacementStrategy,
    label: ENodePoolColumnLabel.CpuNodePlacementStrategy,
    field: (nodePool: INodePoolTable): string => {
      return nodePool?.schedulingConfiguration?.cpuPodsPlacementStrategy?.node || "-";
    },
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  workloads: {
    ...viewLinkColumn({
      name: nodePoolColumnName.workloads,
      label: ENodePoolColumnLabel.Workloads,
      emitName: "workloads-clicked",
    }),
    display: false,
  },
  updatedAt: {
    name: nodePoolColumnName.updatedAt,
    label: ENodePoolColumnLabel.UpdatedAt,
    field: (nodePool: INodePoolTable): string => tableFormatUtil.getFormattedDateOrDefault(nodePool.updatedAt),
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  createdAt: {
    name: nodePoolColumnName.createdAt,
    label: ENodePoolColumnLabel.CreatedAt,
    field: (nodePool: INodePoolTable): string => tableFormatUtil.getFormattedDateOrDefault(nodePool.createdAt),
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  /*** clusterServiceApi.getNodeTelemetries columns ***/
  totalGpus: {
    name: nodePoolColumnName.totalGpus,
    label: ENodePoolColumnLabel.TotalGpus,
    field: (nodePool: INodePoolTable): string =>
      tableFormatUtil.getFormattedNumberOrDefault(nodePool.additionalFields?.totalGpus),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  totalGpuMemoryBytes: {
    name: nodePoolColumnName.totalGpuMemoryBytes,
    label: ENodePoolColumnLabel.TotalGpuMemoryBytes,
    field: (nodePool: INodePoolTable): string => _formatMemoryValue(nodePool.additionalFields?.totalGpuMemoryBytes),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedGpus: {
    name: nodePoolColumnName.allocatedGpus,
    label: ENodePoolColumnLabel.AllocatedGpus,
    field: (nodePool: INodePoolTable): string =>
      tableFormatUtil.getFormattedNumberOrDefault(nodePool.additionalFields?.allocatedGpus),
    sortable: true,
    align: "left",
    hideFilter: true,
  },
  totalCpuCores: {
    name: nodePoolColumnName.totalCpuCores,
    label: ENodePoolColumnLabel.TotalCpuCores,
    field: (nodePool: INodePoolTable): string =>
      tableFormatUtil.getFormattedNumberOrDefault(nodePool.additionalFields?.totalCpuCores),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  totalCpuMemoryBytes: {
    name: nodePoolColumnName.totalCpuMemoryBytes,
    label: ENodePoolColumnLabel.TotalCpuMemoryBytes,
    field: (nodePool: INodePoolTable): string => _formatMemoryValue(nodePool.additionalFields?.totalCpuMemoryBytes),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedCpuCores: {
    name: nodePoolColumnName.allocatedCpuCores,
    label: ENodePoolColumnLabel.AllocatedCpuCores,
    field: (nodePool: INodePoolTable): string =>
      tableFormatUtil.getFormattedNumberOrDefault(nodePool.additionalFields?.allocatedCpuCores),
    sortable: true,
    align: "left",
    hideFilter: true,
  },
  allocatedCpuMemoryBytes: {
    name: nodePoolColumnName.allocatedCpuMemoryBytes,
    label: ENodePoolColumnLabel.AllocatedCpuMemoryBytes,
    field: (nodePool: INodePoolTable): string => _formatMemoryValue(nodePool.additionalFields?.allocatedCpuMemoryBytes),
    sortable: true,
    align: "left",
    hideFilter: true,
  },
};

export const nodePoolIndexColumns: Array<ITableColumn> = [
  { ...allNodePoolsColumnsMap.name, display: true, mandatory: true },
  { ...allNodePoolsColumnsMap.phase, display: true },
  { ...allNodePoolsColumnsMap.cluster, display: false },
  { ...allNodePoolsColumnsMap.labelKey, display: true },
  { ...allNodePoolsColumnsMap.labelValue, display: true },
  { ...allNodePoolsColumnsMap.nodes, display: true },
  { ...allNodePoolsColumnsMap.topology, display: false },
  { ...allNodePoolsColumnsMap.gpuNetworkAccelerationDetection, display: false },
  { ...allNodePoolsColumnsMap.gpuNetworkAccelerationLabelKey, display: false },
  { ...allNodePoolsColumnsMap.gpuNetworkAccelerationDetected, display: false },
  { ...allNodePoolsColumnsMap.minGuaranteedRuntime, display: false },
  { ...allNodePoolsColumnsMap.gpuPlacementStrategy, display: false },
  { ...allNodePoolsColumnsMap.gpuDevicePlacementStrategy, display: false },
  { ...allNodePoolsColumnsMap.cpuPlacementStrategy, display: false },
  { ...allNodePoolsColumnsMap.totalGpus, display: true },
  { ...allNodePoolsColumnsMap.totalGpuMemoryBytes, display: true },
  { ...allNodePoolsColumnsMap.allocatedGpus, display: true },
  { ...allNodePoolsColumnsMap.totalCpuCores, display: true },
  { ...allNodePoolsColumnsMap.totalCpuMemoryBytes, display: true },
  { ...allNodePoolsColumnsMap.allocatedCpuCores, display: true },
  { ...allNodePoolsColumnsMap.allocatedCpuMemoryBytes, display: true },
  { ...allNodePoolsColumnsMap.updatedAt, display: false },
  { ...allNodePoolsColumnsMap.createdAt, display: true },
  { ...allNodePoolsColumnsMap.workloads, display: true },
];

export const nodePoolsDependentColumns = {
  nodes: new Set([nodePoolColumnName.nodes]), //only shown if the user has the permission to view nodes
  cluster: new Set([nodePoolColumnName.cluster]), //only shown when there more than one cluster
  minGuaranteedRuntime: new Set([nodePoolColumnName.minGuaranteedRuntime]), //only shown when cluster version is 2.23 or higher
  gpuDevicePlacementStrategy: new Set([nodePoolColumnName.gpuDevicePlacementStrategy]), //only shown when cluster version is 2.26 or higher
};

//workloads list modal
export const nodePoolWorkloadListModalColumns: Array<ITableColumn> = [
  { ...allWorkloadColumnsMap.name, display: true },
  { ...allWorkloadColumnsMap.type, display: true },
  { ...allWorkloadColumnsMap.status, display: true },
  { ...allWorkloadColumnsMap.submittedBy, display: true },
  { ...allWorkloadColumnsMap.runningPodsVsRequestedPodsForModal, display: true },
  { ...allWorkloadColumnsMap.creationTime, display: true },
  { ...allWorkloadColumnsMap.createdAt, display: true },
  { ...allWorkloadColumnsMap.allocatedGpu, display: true },
  { ...allWorkloadColumnsMap.allocatedGpuMemory, display: true },
  { ...allWorkloadColumnsMap.allocatedCpu, display: true },
  { ...allWorkloadColumnsMap.allocatedCpuMemory, display: true },
];

//org unit node pools modal
export const allOrgUnitNodePoolsModalColumnsMap: Record<string, ITableColumn> = {
  nodePoolName: {
    name: "node-pool-name",
    label: "Node pool",
    field: (row: INodePoolsResourcesRow) => row?.nodePool?.name || "-",
    sortable: true,
    align: "left",
  },
  gpuQuota: {
    name: "gpu",
    label: "GPU quota",
    field: (row: INodePoolsResourcesRow) => row.gpu.deserved,
    sortable: true,
    align: "left",
    format: (gpu: number) => resourceUtil.getResourceDisplayValue(gpu, EResourceType.GPU),
  },
  nodePoolProjectPriority: {
    name: "project-priority",
    label: "Project priority",
    field: (row: INodePoolsResourcesRow) => row.priority,
    sortable: true,
    align: "left",
    format: (priority: string) => orgUnitUtil.mapStringToPriorityLabel(priority),
  },
  nodePoolDepartmentPriority: {
    name: "nodepool-department-priority",
    label: "Department priority",
    field: (row: INodePoolsResourcesRow) => row.priority,
    sortable: true,
    align: "left",
    format: (priority: string) => orgUnitUtil.mapStringToPriorityLabel(priority),
  },
  overQuotaWeight: {
    name: "over-quota-weight",
    label: "Over-quota weight",
    field: (row: INodePoolsResourcesRow) => row.overQuotaWeightLabel,
    sortable: true,
    align: "left",
  },
  cpuQuota: {
    name: "cpu-quota",
    label: "CPU (Cores)",
    field: (row: INodePoolsResourcesRow) => row.cpu?.deserved,
    sortable: true,
    align: "left",
    format: (cpu: number) => resourceUtil.getResourceDisplayValue(cpu, EResourceType.CPU),
  },
  cpuMemoryQuota: {
    name: "cpuMemory",
    label: "CPU memory",
    field: (row: INodePoolsResourcesRow) => row.memory?.deserved,
    sortable: true,
    align: "left",
    format: (memory: number) => resourceUtil.getResourceDisplayValue(memory, EResourceType.MEMORY),
  },
  allocatedGpus: {
    name: "allocatedGpus",
    label: "Allocated GPUs",
    field: (row: INodePoolsResourcesRow) => row.allocatedGpus,
    sortable: true,
    align: "left",
    format: (allocatedGpus: string | undefined) =>
      allocatedGpus ? resourceUtil.getResourceDisplayValue(parseFloat(allocatedGpus), EResourceType.GPU) : "-",
  },
  allocatedCpu: {
    name: "allocatedCpu",
    label: "Allocated CPU (Cores)",
    field: (row: INodePoolsResourcesRow) => row.allocatedCpu,
    sortable: true,
    align: "left",
    format: (allocatedCpu: string | undefined) =>
      allocatedCpu ? resourceUtil.getResourceDisplayValue(parseFloat(allocatedCpu), EResourceType.CPU) : "-",
    sort: (fieldA, fieldB, rowA, rowB) =>
      tableUtil.customColumnSort(allOrgUnitNodePoolsModalColumnsMap.allocatedCpu, rowA, rowB, parseFloat),
  },
  allocatedMemory: {
    name: "allocatedMemory",
    label: "Allocated CPU memory",
    field: (row: INodePoolsResourcesRow) => row.allocatedMemory,
    sortable: true,
    align: "left",
    format: (allocatedMemory: number | undefined) =>
      allocatedMemory ? memoryFormat(convertToBytes(allocatedMemory, "MB")) : "-",
  },
  priorityOrder: {
    name: "priorityOrder",
    label: "Order of priority",
    field: (row: INodePoolsResourcesRow) => row.priorityOrder,
    sortable: true,
    align: "left",
  },
  limit: {
    name: "limit",
    label: "Max GPU devices allocation",
    field: (row: INodePoolsResourcesRow) => resourceUtil.getResourceDisplayValue(row.gpu.limit, EResourceType.GPU_LIMIT),
    sortable: true,
    align: "left",
  },
};

export const orgUnitNodePoolsModalTableColumns: Array<ITableColumn> = [
  { ...allOrgUnitNodePoolsModalColumnsMap.nodePoolName, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.priorityOrder, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.gpuQuota, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.nodePoolProjectPriority, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.nodePoolDepartmentPriority, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.overQuotaWeight, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.limit, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.cpuQuota, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.cpuMemoryQuota, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.allocatedGpus, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.allocatedCpu, display: true },
  { ...allOrgUnitNodePoolsModalColumnsMap.allocatedMemory, display: true },
];

export const orgUnitNodePoolsModalDependentColumns = {
  cpu: new Set([
    allOrgUnitNodePoolsModalColumnsMap.cpuQuota.name,
    allOrgUnitNodePoolsModalColumnsMap.cpuMemoryQuota.name,
    allOrgUnitNodePoolsModalColumnsMap.allocatedCpu.name,
    allOrgUnitNodePoolsModalColumnsMap.allocatedMemory.name,
  ]),
  overQuotaWeight: new Set([allOrgUnitNodePoolsModalColumnsMap.overQuotaWeight.name]),
  priorityOrder: new Set([allOrgUnitNodePoolsModalColumnsMap.priorityOrder.name]),
  project: new Set([allOrgUnitNodePoolsModalColumnsMap.nodePoolProjectPriority.name]),
  department: new Set([allOrgUnitNodePoolsModalColumnsMap.nodePoolDepartmentPriority.name]),
};
