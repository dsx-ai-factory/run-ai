/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  ECustomCell,
  HIGHLIGHT_COLUMN_CONFIG,
  USER_DISPLAY_NAME_COL_DEFAULTS,
  type IStatusColOptions,
  type ITableColumn,
} from "@/models/table.model";
import {
  EWorkloadNodeArchitecture,
  Preemptibility,
  workloadPhaseFilterOptions,
  workloadPreemptibilityOptions,
  type WorkloadServicePod,
} from "@/models/workload.model";
import {
  type Connection,
  type HistoryRecord,
  type PendingSchedulingMessage,
  Phase,
  PodSortFilterFields,
  type Workload,
  type WorkloadRequestResources,
  WorkloadSortFilterFields,
} from "@/swagger-models/workloads-client";
import { dateUtil } from "@/utils/date.util";
import { stringUtil } from "@/utils/string.util";
import { workloadUtil } from "@/utils/workload.util";
import { formatK8sResourceMemory } from "@/utils/format.util/format.util";
import { InternalConnectionType } from "@/swagger-models/assets-service-client";
import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType } from "@/models/filter.model";
import { isEmpty } from "@/utils/object.util";
import { tableFormatUtil } from "@/utils/table-format.util/table-format.util";
import { EWorkloadStatus } from "@/common/status.constant";
import { autocompleteUtil } from "@/utils/autocomplete.util";

//connections
export const allConnectionsTableColumnsMap: Record<string, ITableColumn> = {
  name: {
    name: "name",
    label: "Name",
    field: (row: Connection) => row.name,
    sortable: true,
    align: "left",
    customCell: ECustomCell.NAME_COL,
  },
  connectionType: {
    name: "connection-type",
    label: "Connection type",
    field: (row: Connection) => row.connectionType,
    format: (val: InternalConnectionType) => {
      switch (val) {
        case InternalConnectionType.ExternalUrl:
          return "External URL";
        case InternalConnectionType.NodePort:
          return "Node port";
        case InternalConnectionType.LoadBalancer:
          return "Load Balancer";
        case InternalConnectionType.ServingPort:
          return "Serving endpoint";
        default:
          return val || "-";
      }
    },
    sortable: true,
    align: "left",
  },
  url: {
    name: "connection-url",
    label: "URL",
    field: (row: Connection) => row.url,
    sortable: true,
    align: "left",
    customCell: ECustomCell.WORKLOAD_CONNECTION_COL,
  },
  port: {
    name: "port",
    label: "Port",
    field: (row: Connection) => row.containerPort || "-",
    sortable: true,
    align: "left",
  },
  access: {
    name: "access",
    label: "Access",
    field: (row: Connection) => row,
    sortable: true,
    align: "left",
    customCell: ECustomCell.TOOL_ACCESS_COL,
    hideFilter: true,
  },
};
export const connectionsTableColumns: Array<ITableColumn> = [
  {
    ...allConnectionsTableColumnsMap.name,
    display: true,
  },
  {
    ...allConnectionsTableColumnsMap.connectionType,
    display: true,
  },
  {
    ...allConnectionsTableColumnsMap.access,
    display: true,
  },
  {
    ...allConnectionsTableColumnsMap.port,
    display: true,
  },
  {
    ...allConnectionsTableColumnsMap.url,
    display: true,
  },
];

//workload events history table
const allEventHistoryColumnsMap = {
  createdAt: {
    name: "createdAt",
    label: "Date & Time",
    field: (historyRecord: HistoryRecord) => historyRecord.meta.creationTimestamp,
    align: "left",
    display: true,
    sortable: true,
    format: (createdAt: string) =>
      createdAt && dateUtil.dateAndTimeFormat(new Date(createdAt), { includeMilliseconds: true, includeSeconds: true }),
  },
  event: {
    name: "event",
    label: "Event",
    field: (historyRecord: HistoryRecord) =>
      workloadUtil.isPhaseUpdateRecordType(historyRecord) ? "Status changed" : historyRecord.spec.event?.reason,
    align: "left",
    sortable: true,
    display: true,
  },
  eventType: {
    name: "eventType",
    label: "Event type",
    field: (historyRecord: HistoryRecord) => historyRecord.spec.event?.type || "-",
    align: "left",
    sortable: true,
    display: true,
  },
  eventIssuer: {
    name: "eventIssuer",
    label: "Event issuer",
    field: (historyRecord: HistoryRecord) => historyRecord.spec.event?.source || "-",
    align: "left",
    display: true,
    sortable: true,
  },
  component: {
    name: "component",
    label: "Component",
    field: (historyRecord: HistoryRecord) => historyRecord.spec.event?.involvedObject?.kind || "",
    align: "left",
    display: true,
    sortable: true,
  },
  details: {
    name: "details",
    label: "Details",
    field: (historyRecord: HistoryRecord) =>
      workloadUtil.isPhaseUpdateRecordType(historyRecord)
        ? historyRecord.spec.phaseUpdate?.phase
        : historyRecord.spec.event?.message,
    align: "left",
    display: true,
    sortable: true,
  },
} as Record<string, ITableColumn>;

export const eventHistoryColumns: Array<ITableColumn> = [
  { ...allEventHistoryColumnsMap.createdAt, display: true },
  { ...allEventHistoryColumnsMap.event, display: true },
  { ...allEventHistoryColumnsMap.eventType, display: true },
  { ...allEventHistoryColumnsMap.eventIssuer, display: true },
  { ...allEventHistoryColumnsMap.component, display: true },
  { ...allEventHistoryColumnsMap.details, display: true },
];

//pods modal
export const allWorkloadPodsModalColumnsMap: Record<string, ITableColumn> = {
  name: {
    name: PodSortFilterFields.Name,
    label: "Pod",
    field: (pod: WorkloadServicePod) => pod.name,
    sortable: true,
    align: "left",
  },
  status: {
    name: "status",
    label: "Status",
    field: (pod: WorkloadServicePod) => pod,
    sortable: false,
    align: "left",
    format: (pod: WorkloadServicePod): IStatusColOptions | string => workloadUtil.getPodStatusColOptions(pod),
    sort: (podA: WorkloadServicePod, podB: WorkloadServicePod): number => podA.k8sPhase.localeCompare(podB.k8sPhase),
    exportFormat: (pod: WorkloadServicePod) => pod.k8sPhase || "NA",
    filterKey: "status",
    customCell: ECustomCell.STATUS_COL,
  },
  logs: {
    name: "logs",
    label: "Logs",
    field: "logs",
    align: "center",
    customCell: ECustomCell.ICON_BUTTON_COL,
    customCellEvent: { emitName: "view-logs" },
    customCellConfig: {
      [ECustomCell.ICON_BUTTON_COL]: {
        icon: "ph-bold ph-list-dashes",
        size: "sm",
        getTooltip: () => "View logs",
      },
    },
  },
  terminal: {
    name: "terminal",
    label: "Terminal",
    field: (row: WorkloadServicePod) => row,
    align: "center",
    customCell: ECustomCell.ICON_BUTTON_COL,
    customCellEvent: { emitName: "connect-terminal" },
    customCellConfig: {
      [ECustomCell.ICON_BUTTON_COL]: {
        icon: "ph-bold ph-terminal-window",
        size: "sm",
        isDisabled: (row: WorkloadServicePod) => row.k8sPhase !== EWorkloadStatus.Running,
        getTooltip: (row: WorkloadServicePod) =>
          row.k8sPhase === EWorkloadStatus.Running ? "Connect via terminal" : "To connect, the pod must be running",
      },
    },
  },
  gpuComputeAllocation: {
    name: "gpuComputeAllocation",
    label: "Allocated GPUs",
    field: (pod: WorkloadServicePod): number | null => {
      return pod.allocatedResources?.gpu || null;
    },
    format: (gpuCompute: number | null): number | string => (gpuCompute || gpuCompute === 0 ? gpuCompute : "-"),
    sortable: false,
    align: "left",
  },
  gpuMemoryAllocation: {
    name: "gpuMemoryAllocation",
    label: "Allocated GPU memory ",
    field: (pod: WorkloadServicePod): string | null | undefined => {
      return pod.allocatedResources?.gpuMemory;
    },
    format: (gpuMemory: string | null): string => (gpuMemory ? formatK8sResourceMemory(gpuMemory) : "-"),
    sortable: false,
    align: "left",
  },
  cpuAllocation: {
    name: "cpuAllocation",
    label: "Allocated CPU (Cores)",
    field: (pod: WorkloadServicePod): number | null | undefined => {
      return pod.allocatedResources?.cpu;
    },
    format: (cpu: number | null | undefined): number | string => (cpu || cpu === 0 ? cpu : "-"),
    sortable: false,
    align: "left",
  },
  cpuMemoryAllocation: {
    name: "cpuMemoryAllocation",
    label: "Allocated CPU memory",
    field: (pod: WorkloadServicePod): string | null | undefined => {
      return pod.allocatedResources?.cpuMemory;
    },
    format: (cpuMemory: string | null | undefined): number | string =>
      cpuMemory ? formatK8sResourceMemory(cpuMemory) : "-",
    sortable: false,
    align: "left",
  },
  nodeName: {
    name: PodSortFilterFields.NodeName,
    label: "Node",
    field: (pod: WorkloadServicePod) => pod.nodeName || "-",
    sortable: true,
    align: "left",
  },
  projectName: {
    name: "projectName",
    label: "Project",
    field: (pod: WorkloadServicePod) => (pod.projectName ? [pod.projectName] : "-"),
    sortable: false,
    align: "left",
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "project-clicked" },
  },
  workloadName: {
    name: "workloadName",
    label: "Workload",
    field: (pod: WorkloadServicePod) => (pod.workloadName ? [pod.workloadName] : "-"),
    sortable: false,
    align: "left",
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "workload-clicked" },
  },
  nodePools: {
    name: PodSortFilterFields.CurrentNodepool,
    label: "Node pool",
    field: (pod: WorkloadServicePod) => pod.currentNodePool || "-",
    sortable: true,
    align: "left",
  },
  image: {
    name: "image",
    label: "Image",
    field: (pod: WorkloadServicePod) => {
      const userContainer = pod.containers?.find((container) => container.name.includes("user-container"));
      return userContainer ? userContainer.image : pod.containers[0].image;
    },
    sortable: false,
    align: "left",
  },
  ip: {
    name: "ip",
    label: "IP",
    field: (pod: WorkloadServicePod) => pod.ip || "-",
    sortable: false,
    align: "left",
  },
  created: {
    name: PodSortFilterFields.CreatedAt,
    label: "Creation time",
    field: (pod: WorkloadServicePod) => pod.createdAt,
    format: (createdAt: string): string => (!createdAt ? "-" : dateUtil.dateAndTimeFormat(new Date(createdAt))),
    sortable: true,
    align: "left",
  },
};

export const workloadPodsModalTableColumns: Array<ITableColumn> = [
  { ...allWorkloadPodsModalColumnsMap.name, display: true },
  { ...allWorkloadPodsModalColumnsMap.status, display: true },
  { ...allWorkloadPodsModalColumnsMap.logs, display: true },
  { ...allWorkloadPodsModalColumnsMap.terminal, display: true },
  { ...allWorkloadPodsModalColumnsMap.nodeName, display: true },
  { ...allWorkloadPodsModalColumnsMap.nodePools, display: true },
  { ...allWorkloadPodsModalColumnsMap.gpuComputeAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.gpuMemoryAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.cpuAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.cpuMemoryAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.image, display: true },
  { ...allWorkloadPodsModalColumnsMap.ip, display: true },
  { ...allWorkloadPodsModalColumnsMap.created, display: true },
];

export const nodePodsModalTableColumns: Array<ITableColumn> = [
  { ...allWorkloadPodsModalColumnsMap.name, display: true },
  { ...allWorkloadPodsModalColumnsMap.status, display: true },
  { ...allWorkloadPodsModalColumnsMap.logs, display: true },
  { ...allWorkloadPodsModalColumnsMap.terminal, display: true },
  { ...allWorkloadPodsModalColumnsMap.projectName, display: true },
  { ...allWorkloadPodsModalColumnsMap.workloadName, display: true },
  { ...allWorkloadPodsModalColumnsMap.gpuComputeAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.gpuMemoryAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.cpuAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.cpuMemoryAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.image, display: true },
  { ...allWorkloadPodsModalColumnsMap.ip, display: true },
  { ...allWorkloadPodsModalColumnsMap.created, display: true },
];

export const workloadPodsModalDependentColumns = {
  // display if there is more than one node pool
  nodePools: new Set([allWorkloadPodsModalColumnsMap.nodePools.name]),
  // display for 2.19 +
  ip: new Set([allWorkloadPodsModalColumnsMap.ip.name]),
};

const workloadTypeMap: Record<string, string> = {
  Training: "Training",
  Interactive: "Workspace",
};

//workloads table data
export const allWorkloadColumnsMap: Record<string, ITableColumn> = {
  name: {
    name: WorkloadSortFilterFields.Name,
    label: "Workload",
    field: (workload: Workload) => workload.name,
    sortable: true,
    align: "left",
    customCell: ECustomCell.NAME_COL,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.workloads(WorkloadSortFilterFields.Name, (w) => w.name),
    },
  },
  type: {
    name: WorkloadSortFilterFields.Type,
    label: "Type",
    field: (workload: Workload) => workload.type,
    format: (workloadType: string) => workloadTypeMap[workloadType] || stringUtil.capitalizeString(workloadType),
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.workloads(WorkloadSortFilterFields.Type, (w) => w.type),
    },
  },
  preemptible: {
    name: WorkloadSortFilterFields.Preemptibility,
    label: "Preemptibility",
    field: (workload: Workload): string | null | undefined => {
      if (workload.preemptibility != null) return workload.preemptibility;
      if (typeof workload.preemptible !== "boolean") return workload.preemptibility;
      return workload.preemptible ? Preemptibility.Preemptible : Preemptibility.NonPreemptible;
    },
    format: (preemptibility: string | null | undefined) =>
      workloadPreemptibilityOptions.find((option) => option.value === preemptibility)?.label ?? "-",
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: workloadPreemptibilityOptions,
    },
  },
  status: {
    name: WorkloadSortFilterFields.Phase,
    label: "Status",
    field: (workload: Workload) => workload,
    sortable: true,
    align: "left",
    customCell: ECustomCell.WORKLOAD_STATUS_COL,
    customCellEvent: { emitName: "workload-status-clicked" },
    format: (workload: Workload): IStatusColOptions | string => {
      if (workload.phase === Phase.Pending || workload.phase === Phase.Completed) {
        return workloadUtil.getWorkloadStatusColOptions(workload, workload.phaseMessage);
      }
      // All other phases: suppress phaseMessage if issues exist
      const phaseMessage = workload.additionalFields?.["issues"] ? "" : workload.phaseMessage;
      return workloadUtil.getWorkloadStatusColOptions(workload, phaseMessage);
    },
    sort: (workloadA: Workload, workloadB: Workload): number => workloadA.phase.localeCompare(workloadB.phase),
    exportFormat: (workload: Workload) => workload.phase || "NA",
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: workloadPhaseFilterOptions,
    },
  },
  submittedBy: {
    name: WorkloadSortFilterFields.SubmittedBy,
    label: "Created by",
    field: (workload: Workload) => workload.submittedBy || "-",
    sortable: true,
    align: "left",
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.workloads(WorkloadSortFilterFields.SubmittedBy, (w) => w.submittedBy),
    },
  },
  departmentName: {
    name: WorkloadSortFilterFields.DepartmentName,
    label: "Department",
    field: (workload: Workload) => workload.departmentName || "-",
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.departments(),
    },
  },
  project: {
    name: WorkloadSortFilterFields.ProjectName,
    label: "Project",
    field: (workload: Workload) => [workload.projectName],
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.projects(),
    },
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "project-clicked" },
  },
  nodePools: {
    name: WorkloadSortFilterFields.Nodepool,
    label: "Node pool",
    field: (workload: Workload): string[] => {
      if (workload?.currentNodePools?.length) {
        return workload.currentNodePools;
      }
      return [];
    },
    format: (currentNodePools: string[]): string => (currentNodePools.length > 0 ? currentNodePools.join(",") : "-"),
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.nodePoolsMinimal(),
    },
  },
  runningPodsVsRequestedPods: {
    name: "runningPodsVsRequestedPods",
    label: "Running / requested pods",
    field: (workload: Workload) => workload,
    format: (workload: Workload): string[] => [workloadUtil.getRunningVsRequestedDisplayedValue(workload)],
    sortable: false,
    align: "left",
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "pods-clicked" },
    hideFilter: true,
  },
  runningPodsVsRequestedPodsForModal: {
    name: "runningPodsVsRequestedPods",
    label: "Running / requested pods",
    field: (workload: Workload) => workload,
    format: (workload: Workload): string => workloadUtil.getRunningVsRequestedDisplayedValue(workload),
    sortable: false,
    align: "left",
    customCell: ECustomCell.LIST_COL,
    hideFilter: true,
  },
  creationTime: {
    name: WorkloadSortFilterFields.CreatedAt,
    label: "Creation time",
    field: (workload: Workload) => workload.createdAt,
    format: (createdAt: string): string => {
      return !createdAt ? "-" : dateUtil.dateAndTimeFormat(new Date(createdAt));
    },
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  completionTime: {
    name: WorkloadSortFilterFields.CompletedAt,
    label: "Completion time",
    field: (workload: Workload) => workload.completedAt,
    format: (completedAt: string): string => {
      return !completedAt ? "-" : dateUtil.dateAndTimeFormat(new Date(completedAt));
    },
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  totalPendingTimeSeconds: {
    name: WorkloadSortFilterFields.TotalPendingTimeSeconds,
    label: "Total pending time",
    field: (workload: Workload) => workload.totalPendingTimeSeconds,
    format: (totalPendingTimeSeconds?: number): string => {
      if (!totalPendingTimeSeconds || !dateUtil.isDurationMoreThanMinute(totalPendingTimeSeconds)) return "-";
      return dateUtil.formatDuration(totalPendingTimeSeconds);
    },
    sortable: true,
    align: "left",
    hideFilter: true,
  },
  totalRuntime: {
    name: WorkloadSortFilterFields.TotalRunningTimeSeconds,
    label: "Total runtime",
    field: (workload: Workload) => workload.totalRunningTimeSeconds,
    format: (totalRunningTimeSeconds?: number | null): string => {
      if (!totalRunningTimeSeconds) return "-";
      return dateUtil.formatDuration(totalRunningTimeSeconds);
    },
    sortable: true,
    align: "left",
    hideFilter: true,
  },
  guaranteedRuntimeEndsAt: {
    name: WorkloadSortFilterFields.GuaranteedRuntimeEndsAt,
    label: "Guaranteed runtime left",
    field: (workload: Workload) => workload.guaranteedRuntimeEndsAt,
    format: (guaranteedRuntimeEndsAt?: string | null): string => {
      if (!guaranteedRuntimeEndsAt) return "-";
      return dateUtil.formatTimeRemaining(guaranteedRuntimeEndsAt);
    },
    sortable: true,
    align: "left",
    hideFilter: true,
  },
  image: {
    name: "image",
    label: "Image",
    field: (workload: Workload) => workload.images,
    sortable: false,
    align: "left",
    format: (imgs: string[]) => {
      if (!imgs?.length) return ["-"];
      return imgs;
    },
    customCell: ECustomCell.STRINGS_LIST_COL,
    hideFilter: true,
  },
  multiNode: {
    name: "multi-node",
    label: "Workload architecture",
    field: (row: Workload) => row.distributedFramework,
    sortable: false,
    align: "left",
    format: (distributedFramework: string | undefined) =>
      distributedFramework ? EWorkloadNodeArchitecture.Distributed : EWorkloadNodeArchitecture.Standard,
    hideFilter: true,
  },
  requestedGPU: {
    name: "requestedGPU",
    label: "GPU compute request",
    field: (workload: Workload): string | number | null | undefined => {
      const requestedResources: WorkloadRequestResources | undefined | null = !isEmpty(workload.podsRequestedResources)
        ? workload.podsRequestedResources
        : workload.workloadRequestedResources;
      return requestedResources?.gpu?.request || "-";
    },
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedGPU: {
    name: WorkloadSortFilterFields.AllocatedGpu,
    label: "GPU compute allocation",
    field: (workload: Workload) => {
      return workload.allocatedResources?.gpu || "-";
    },
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.Numeric,
    },
  },
  requestedGPUMemory: {
    name: "requestedGPUMemory",
    label: "GPU memory request",
    field: (workload: Workload): string | number | null | undefined => {
      const requestedResources: WorkloadRequestResources | undefined | null = !isEmpty(workload.podsRequestedResources)
        ? workload.podsRequestedResources
        : workload.workloadRequestedResources;
      return requestedResources?.gpuMemory?.request || "-";
    },
    format: (gpuMemory: string | null | undefined): string => (gpuMemory ? formatK8sResourceMemory(gpuMemory) : "-"),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedGPUMemory: {
    name: "allocatedGPUMemory",
    label: "GPU memory allocation",
    field: (workload: Workload): string | null | undefined => {
      return workload.allocatedResources?.gpuMemory;
    },
    format: (gpuMemory: string | null | undefined): string => (gpuMemory ? formatK8sResourceMemory(gpuMemory) : "-"),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  requestedCPUMemory: {
    name: "requestedCPUMemory",
    label: "CPU memory request",
    field: (workload: Workload): string | null | undefined => {
      const requestedResources: WorkloadRequestResources | undefined | null = !isEmpty(workload.podsRequestedResources)
        ? workload.podsRequestedResources
        : workload.workloadRequestedResources;
      return requestedResources?.cpuMemory?.request || "-";
    },
    format: (cpuMemory: string | null | undefined): string => (cpuMemory ? formatK8sResourceMemory(cpuMemory) : "-"),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedCPUMemory: {
    name: "allocatedCPUMemory",
    label: "CPU memory allocation",
    field: (workload: Workload): string | null | undefined => {
      return workload.allocatedResources?.cpuMemory;
    },
    format: (cpuMemory: string | null | undefined): string => (cpuMemory ? formatK8sResourceMemory(cpuMemory) : "-"),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  requestedCPUCompute: {
    name: "requestedCPUCompute",
    label: "CPU compute request",
    field: (workload: Workload): string | number | null | undefined => {
      const requestedResources: WorkloadRequestResources | undefined | null = !isEmpty(workload.podsRequestedResources)
        ? workload.podsRequestedResources
        : workload.workloadRequestedResources;
      return requestedResources?.cpu?.request || "-";
    },
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedCPUCompute: {
    name: "allocatedCPUCompute",
    label: "CPU compute allocation",
    field: (workload: Workload): number | string => workload.allocatedResources?.cpu || "-",
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  idleGpus: {
    name: WorkloadSortFilterFields.IdleAllocatedGpus,
    label: "Idle GPU devices",
    field: (workload: Workload) => tableFormatUtil.getFormattedNumberOrDefault(workload.idleAllocatedGpus),
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.Numeric,
    },
  },
  connections: {
    name: "connections",
    label: "Connection(s)",
    field: (workload: Workload) => workloadUtil.getWorkloadConnectionsForGrid(workload),
    sortable: false,
    align: "left",
    format: (connections: Array<Connection>): number[] | string => {
      const amount = connections?.length;

      if (amount === 0) return "-";

      return [amount];
    },
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "connections-clicked" },
    hideFilter: true,
  },
  allocatedGpu: {
    name: "allocatedGpu",
    label: "Allocated GPU compute",
    field: (workload: Workload) => {
      return workload.allocatedResources?.gpu || "-";
    },
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedGpuMemory: {
    name: "allocatedGpuMemory",
    label: "Allocated GPU memory ",
    field: (workload: Workload): string | undefined | null => {
      return workload.allocatedResources?.gpuMemory;
    },
    format: (gpuMemory: string | undefined | null): string => (gpuMemory ? formatK8sResourceMemory(gpuMemory) : "-"),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedCpu: {
    name: "allocatedCpu",
    label: "Allocated CPU compute (Cores)",
    field: (workload: Workload) => workload.allocatedResources?.cpu || "-",
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  allocatedCpuMemory: {
    name: "allocatedCpuMemory",
    label: "Allocated CPU memory",
    field: (workload: Workload): string | undefined | null => {
      return workload.allocatedResources?.cpuMemory;
    },
    format: (cpuMemory: string | undefined | null): string => (cpuMemory ? formatK8sResourceMemory(cpuMemory) : "-"),
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: "Cluster",
    field: (workload: Workload): string | undefined => workload.clusterId,
    sortable: false,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
  },
  requestedRuns: {
    name: "requestedRuns",
    label: "Requested runs",
    field: (workload: Workload): number | undefined | null => workload.requestedPods?.completions,
    format: (completions: number | undefined | null): string => completions?.toString() || "-",
    sortable: false,
    align: "left",
    hideFilter: true,
  },
  category: {
    name: WorkloadSortFilterFields.Category,
    label: "Category",
    field: (workload: Workload): string => workload.category,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: [],
    },
  },
  priority: {
    name: WorkloadSortFilterFields.Priority,
    label: "Priority",
    field: (workload: Workload): string | number => {
      if (!workload.priorityClassName) return "-";
      if (workload.priority === null) return workload.priorityClassName;
      return `${workload.priorityClassName} (${workload.priority})`;
    },
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: [],
    },
  },
  aiApplication: {
    name: WorkloadSortFilterFields.AiApplicationName,
    label: "AI application",
    field: (workload: Workload) => workload.aiApplicationName || "-",
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.workloads(
        WorkloadSortFilterFields.AiApplicationName,
        (w) => w.aiApplicationName,
      ),
    },
  },
  idleTime: {
    name: "idleTime",
    label: "Idle time",
    field: (workload: Workload) => workload.idleSince,
    format: (idleSince: string): string => {
      if (!idleSince) return "-";
      const idleTime = new Date(idleSince).getTime();

      const now = Date.now();
      const durationInSeconds = Math.floor((now - idleTime) / 1000);
      if (durationInSeconds < 0) return "-";
      return dateUtil.formatDuration(durationInSeconds);
    },
    sortable: false,
    align: "left",
  },
};

export const workloadIndexColumns: Array<ITableColumn> = [
  { ...allWorkloadColumnsMap.name, display: true, mandatory: true },
  { ...allWorkloadColumnsMap.type, display: true },
  { ...allWorkloadColumnsMap.status, display: true },
  { ...allWorkloadColumnsMap.project, display: true },
  { ...allWorkloadColumnsMap.departmentName, display: true },
  { ...allWorkloadColumnsMap.category, display: false },
  { ...allWorkloadColumnsMap.priority, display: false },
  { ...allWorkloadColumnsMap.preemptible, display: false },
  { ...allWorkloadColumnsMap.cluster, display: false },
  { ...allWorkloadColumnsMap.aiApplication, display: false },
  { ...allWorkloadColumnsMap.nodePools, display: true },
  { ...allWorkloadColumnsMap.runningPodsVsRequestedPods, display: true },
  { ...allWorkloadColumnsMap.requestedRuns, display: false },
  { ...allWorkloadColumnsMap.submittedBy, display: false },
  { ...allWorkloadColumnsMap.creationTime, display: false },
  { ...allWorkloadColumnsMap.completionTime, display: false },
  { ...allWorkloadColumnsMap.totalPendingTimeSeconds, display: false },
  { ...allWorkloadColumnsMap.totalRuntime, display: false },
  { ...allWorkloadColumnsMap.guaranteedRuntimeEndsAt, display: false },
  { ...allWorkloadColumnsMap.connections, display: false },
  { ...allWorkloadColumnsMap.image, display: false },
  { ...allWorkloadColumnsMap.multiNode, display: false },
  { ...allWorkloadColumnsMap.requestedGPU, display: true },
  { ...allWorkloadColumnsMap.allocatedGPU, display: true },
  { ...allWorkloadColumnsMap.requestedGPUMemory, display: true },
  { ...allWorkloadColumnsMap.allocatedGPUMemory, display: true },
  { ...allWorkloadColumnsMap.idleGpus, display: true },
  { ...allWorkloadColumnsMap.requestedCPUCompute, display: false },
  { ...allWorkloadColumnsMap.allocatedCPUCompute, display: false },
  { ...allWorkloadColumnsMap.requestedCPUMemory, display: false },
  { ...allWorkloadColumnsMap.allocatedCPUMemory, display: false },
];

export const deletedWorkloadIndexColumns: Array<ITableColumn> = [
  { ...allWorkloadColumnsMap.name, display: true, mandatory: true },
  { ...allWorkloadColumnsMap.type, display: true },
  { ...allWorkloadColumnsMap.project, display: true, field: (workload: Workload) => workload.projectName },
  { ...allWorkloadColumnsMap.departmentName, display: true },
  { ...allWorkloadColumnsMap.category, display: false },
  { ...allWorkloadColumnsMap.priority, display: false },
  { ...allWorkloadColumnsMap.preemptible, display: false },
  { ...allWorkloadColumnsMap.aiApplication, display: false },
  { ...allWorkloadColumnsMap.requestedRuns, display: false },
  { ...allWorkloadColumnsMap.submittedBy, display: false },
  { ...allWorkloadColumnsMap.creationTime, display: false },
  { ...allWorkloadColumnsMap.completionTime, display: false },
  { ...allWorkloadColumnsMap.totalPendingTimeSeconds, display: false },
  { ...allWorkloadColumnsMap.totalRuntime, display: false },
  { ...allWorkloadColumnsMap.guaranteedRuntimeEndsAt, display: false },
  { ...allWorkloadColumnsMap.connections, display: false },
  { ...allWorkloadColumnsMap.image, display: false },
  { ...allWorkloadColumnsMap.multiNode, display: false },
  { ...allWorkloadColumnsMap.requestedGPU, display: true },
  { ...allWorkloadColumnsMap.requestedGPUMemory, display: true },
  { ...allWorkloadColumnsMap.idleGpus, display: true },
  { ...allWorkloadColumnsMap.requestedCPUCompute, display: false },
  { ...allWorkloadColumnsMap.requestedCPUMemory, display: false },
  { ...allWorkloadColumnsMap.cluster, display: true },
];

//workloads list modal
export const workloadListModalColumns: Array<ITableColumn> = [
  { ...allWorkloadColumnsMap.name, display: true },
  { ...allWorkloadColumnsMap.type, display: true },
  { ...allWorkloadColumnsMap.status, display: true },
  { ...allWorkloadColumnsMap.submittedBy, display: true },
  { ...allWorkloadColumnsMap.runningPodsVsRequestedPodsForModal, display: true },
  { ...allWorkloadColumnsMap.creationTime, display: true },
  { ...allWorkloadColumnsMap.createdAt, display: true },
  { ...allWorkloadColumnsMap.requestedGPU, display: true },
  { ...allWorkloadColumnsMap.requestedGPUMemory, display: true },
  { ...allWorkloadColumnsMap.requestedCPUMemory, display: true },
  { ...allWorkloadColumnsMap.requestedCPUCompute, display: true },
];

export const workloadDependentColumns = {
  // display only for versions that support distributed workloads
  multiNode: new Set([allWorkloadColumnsMap.multiNode.name]),
  // display if department enabled
  department: new Set([allWorkloadColumnsMap.departmentName.name]),
  // display if there is more than one node pool and on active workloads view only
  nodePools: new Set([allWorkloadColumnsMap.nodePools.name]),
  hiddenOnDeletedView: new Set([
    allWorkloadColumnsMap.allocatedGPU.name,
    allWorkloadColumnsMap.allocatedGPUMemory.name,
    allWorkloadColumnsMap.allocatedCpuMemory.name,
    allWorkloadColumnsMap.allocatedCpu.name,
    allWorkloadColumnsMap.allocatedCPUCompute.name,
    allWorkloadColumnsMap.allocatedCPUMemory.name,
    allWorkloadColumnsMap.nodePools.name,
    allWorkloadColumnsMap.runningPodsVsRequestedPods.name,
    allWorkloadColumnsMap.status.name,
  ]),
  // display if there is more than one cluster
  cluster: new Set([allWorkloadColumnsMap.cluster.name]),
};

//workload pending scheduling messages
export const allPendingSchedulingMessageColumnsMap: ITableColumn[] = [
  {
    name: "nodePool",
    label: "Node Pool",
    field: (message: PendingSchedulingMessage) => message.nodePool,
    align: "left",
  },
  {
    name: "reason",
    label: "Reason",
    field: (message: PendingSchedulingMessage) => message.reason,
    align: "left",
  },
  {
    name: "userMessage",
    label: "Details",
    field: (message: PendingSchedulingMessage) => message.userMessage || "-",
    align: "left",
  },
];

//widgets
export const idleGpuWorkloadColumns: Array<ITableColumn> = [
  {
    ...allWorkloadColumnsMap.name,
    display: true,
    sortable: false,
    customCell: ECustomCell.LINK_COL,
    customCellEvent: { emitName: "workload-clicked" },
  },
  {
    ...allWorkloadColumnsMap.project,
    display: true,
    sortable: false,
    customCell: undefined,
    field: (workload: Workload) => workload.projectName,
  },
  { ...allWorkloadColumnsMap.submittedBy, display: true, sortable: false },
  { ...allWorkloadColumnsMap.allocatedGPU, display: true, sortable: false },
  {
    ...allWorkloadColumnsMap.idleGpus,
    ...HIGHLIGHT_COLUMN_CONFIG,
    display: true,
  },
  {
    ...allWorkloadColumnsMap.idleTime,
    sortable: false,
    display: true,
  },
];
