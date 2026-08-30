import {
  ECustomCell,
  USER_DISPLAY_NAME_COL_DEFAULTS,
  type IStatusColOptions,
  type ITableColumn,
} from "@/models/table.model";
import type { ITrainingList } from "@/models/training.model";
import { EWorkloadNodeArchitecture } from "@/models/workload.model";
import type { AssetRef, DisplayedJob, Distributed } from "@/swagger-models/assets-service-client";
import { dateUtil } from "@/utils/date.util";
import { memoryFormat } from "@/utils/format.util";
import { tableNumberFormat } from "@/utils/table-format.util";
import { workloadUtil } from "@/utils/workload.util";

export const allTrainingColumnsMap: Record<string, ITableColumn> = {
  name: {
    name: "name",
    label: "Training",
    field: (row: ITrainingList) => row.meta.name,
    sortable: true,
    align: "left",
    customCell: ECustomCell.NAME_COL,
  },
  status: {
    name: "training-status",
    label: "Status",
    field: (row: ITrainingList) => row.job,
    sortable: true,
    align: "left",
    customCell: ECustomCell.STATUS_COL,
    format: (job: DisplayedJob): IStatusColOptions | string =>
      workloadUtil.getWorkspaceTrainingStatusColOptions(job, job?.statusTooltip),
    sort: (jobA: DisplayedJob, jobB: DisplayedJob): number => {
      if (jobA.status && jobB.status) {
        return jobA.status.localeCompare(jobB.status);
      }
      return -1;
    },
    exportFormat: (job: DisplayedJob) => job.status || "NA",
    filterKey: "status",
  },
  createdBy: {
    name: "createdBy",
    label: "Created by",
    field: (row: ITrainingList) => row.meta.createdBy,
    sortable: true,
    align: "left",
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
  },
  project: {
    name: "project",
    label: "Project",
    field: (row: ITrainingList) => row.job?.project,
    sortable: true,
    align: "left",
  },
  environment: {
    name: "environment",
    label: "Environment",
    field: (row: ITrainingList) => row.spec.assets.environment,
    sortable: true,
    align: "left",
    format: (env: AssetRef) => env?.name || "",
  },
  multiNode: {
    name: "multi-node",
    label: "Workload architecture",
    field: (row: ITrainingList) => row.spec.distributed,
    sortable: true,
    align: "left",
    format: (val: Distributed | undefined) =>
      val ? EWorkloadNodeArchitecture.Distributed : EWorkloadNodeArchitecture.Standard,
  },
  imageName: {
    name: "imageName",
    label: "Image",
    field: (row: ITrainingList) => row.job?.imageName,
    sortable: true,
    align: "left",
  },
  dataSource: {
    name: "data-sources",
    label: "Data source(s)",
    field: (row: ITrainingList) => row.spec.assets.datasources,
    sortable: true,
    align: "left",
    format: (dataSources: Array<AssetRef>): Array<string> => {
      if (!dataSources?.length) return [];
      return dataSources.map((ds: AssetRef) => ds.name);
    },
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "data-source-clicked" },
  },
  node: {
    name: "node",
    label: "Node",
    field: (row: ITrainingList) => row.job?.nodeId,
    sortable: true,
    align: "left",
  },
  nodePool: {
    name: "nodePool",
    label: "Node pool",
    field: (row: ITrainingList) => row.job?.nodePool,
    sortable: true,
    align: "left",
  },
  creationTime: {
    name: "creationTime",
    label: "Creation time",
    field: (row: ITrainingList) => row.job?.creationTime,
    sortable: true,
    align: "left",
    format: (val: number | undefined) => {
      if (!val) return "";
      return dateUtil.dateAndTimeFormat(new Date(val));
    },
  },
  currentRequestedGPUs: {
    name: "currentRequestedGPUs",
    label: "Requested GPUs",
    field: (row: ITrainingList) => row.job?.currentRequestedGPUs,
    sortable: true,
    align: "left",
    format: tableNumberFormat(),
  },
  currentAllocatedGPUs: {
    name: "currentAllocatedGPUs",
    label: "Allocated GPUs",
    field: (row: ITrainingList) => row.job?.currentAllocatedGPUs,
    sortable: true,
    align: "left",
    format: tableNumberFormat(),
  },
  currentAllocatedGPUsMemory: {
    name: "currentAllocatedGPUsMemory",
    label: "Allocated GPU memory",
    field: (row: ITrainingList) => row.extendJobValues?.currentAllocatedGPUsMemoryBytes,
    sortable: true,
    align: "left",
    format: memoryFormat,
  },
  usedGpuMemory: {
    name: "usedGpuMemory",
    label: "Used GPU memory",
    field: (row: ITrainingList) => row.extendJobValues?.usedGpuMemory,
    sortable: false,
    align: "left",
    format: memoryFormat,
  },
  gpusUtilization: {
    name: "gpusUtilization",
    label: "GPU compute utilization",
    field: (row: ITrainingList) => row.extendJobValues?.gpusUtilization,
    sortable: false,
    align: "left",
    format: (val: number) => {
      if (val || +val === 0) {
        return tableNumberFormat("%")(val);
      } else {
        return "-";
      }
    },
  },
  usedCPUs: {
    name: "usedCPUs",
    label: "Used CPU (Cores)",
    field: (row: ITrainingList) => row.extendJobValues?.usedCPUs,
    sortable: false,
    align: "left",
    format: tableNumberFormat(),
  },
  usedMemory: {
    name: "usedMemory",
    label: "Used CPU memory",
    field: (row: ITrainingList) => row.extendJobValues?.usedMemory,
    sortable: false,
    align: "left",
    format: memoryFormat,
  },
};

export const allTrainingColumns: Array<ITableColumn> = [
  allTrainingColumnsMap.name,
  allTrainingColumnsMap.status,
  allTrainingColumnsMap.createdBy,
  allTrainingColumnsMap.project,
  allTrainingColumnsMap.environment,
  allTrainingColumnsMap.multiNode,
  allTrainingColumnsMap.imageName,
  allTrainingColumnsMap.dataSource,
  allTrainingColumnsMap.node,
  allTrainingColumnsMap.nodePool,
  allTrainingColumnsMap.creationTime,
  allTrainingColumnsMap.currentRequestedGPUs,
  allTrainingColumnsMap.currentAllocatedGPUs,
  allTrainingColumnsMap.currentAllocatedGPUsMemory,
  allTrainingColumnsMap.usedGpuMemory,
  allTrainingColumnsMap.gpusUtilization,
  allTrainingColumnsMap.usedCPUs,
  allTrainingColumnsMap.usedMemory,
];

export const trainingIndexColumns: Array<ITableColumn> = [
  { ...allTrainingColumnsMap.name, display: true, mandatory: true },
  { ...allTrainingColumnsMap.status, display: true },
  { ...allTrainingColumnsMap.createdBy, display: false },
  { ...allTrainingColumnsMap.project, display: true },
  { ...allTrainingColumnsMap.environment, display: false },
  { ...allTrainingColumnsMap.multiNode, display: false },
  { ...allTrainingColumnsMap.imageName, display: true },
  { ...allTrainingColumnsMap.dataSource, display: true },
  { ...allTrainingColumnsMap.node, display: false },
  { ...allTrainingColumnsMap.nodePool, display: false },
  { ...allTrainingColumnsMap.creationTime, display: false },
  { ...allTrainingColumnsMap.currentRequestedGPUs, display: true },
  { ...allTrainingColumnsMap.currentAllocatedGPUs, display: true },
  { ...allTrainingColumnsMap.currentAllocatedGPUsMemory, display: false },
  { ...allTrainingColumnsMap.usedGpuMemory, display: false },
  { ...allTrainingColumnsMap.gpusUtilization, display: false },
  { ...allTrainingColumnsMap.usedCPUs, display: false },
  { ...allTrainingColumnsMap.usedMemory, display: false },
];

export const trainingMiniTableColumns: Array<ITableColumn> = [
  { ...allTrainingColumnsMap.name, display: true },
  { ...allTrainingColumnsMap.status, display: true },
];
