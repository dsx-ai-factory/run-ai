import type { IDepartmentTable } from "@/models/department.model";
import { EResourceState, EResourceType } from "@/models/resource.model";
import {
  ECustomCell,
  ECustomHeader,
  type ICustomHeaderRadioButtonsConfig,
  type ITableColumn,
} from "@/models/table.model";
import { viewLinkColumn } from "@/table-models/common-columns";
import { dateUtil } from "@/utils/date.util";
import { convertToBytes, memoryFormat, numberFormat, toPercent } from "@/utils/format.util";
import { resourceUtil } from "@/utils/resource.util";
import { EDepartmentColumnName } from "@/models/department.model";
import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import {
  DepartmentFilterSortFields,
  type DepartmentProjectChild,
  type Resources,
} from "@/swagger-models/org-unit-service-client";
import { orgUnitUtil } from "@/utils/org-unit.util";
import {
  AVG_RESOURCE_TIMEFRAME_FILTER_KEY,
  EOrgUnitOverTimeColumnName,
  ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
} from "@/models/org-unit.model";

export enum EDepartmentColumnLabel {
  Name = "Department",
  Description = "Description",
  Cluster = "Cluster",
  Projects = "Project(s)",
  NodePools = "Node pool(s)",
  AccessRules = "Subject(s)",
  AllocatedGpus = "Allocated GPUs",
  AverageGpuAllocation = "Avg. GPU allocation",
  AverageGpuUtilization = "Avg. GPU utilization",
  AverageGpuMemoryUtilization = "Avg. GPU memory utilization",
  AllocatedCpu = "Allocated CPUs (Cores)",
  AllocatedMemory = "Allocated CPU memory",
  GpuUtilization = "GPU allocation ratio",
  CpuUtilization = "CPU allocation ratio",
  MemoryUtilization = "CPU memory allocation ratio",
  GpuQuota = "GPU quota",
  CpuQuota = "CPU quota (Cores)",
  MemoryQuota = "CPU memory quota",
  CreatedAt = "Creation time",
  Workloads = "Workload(s)",
}

export const allDepartmentColumnsMap: Record<string, ITableColumn> = {
  departmentName: {
    name: DepartmentFilterSortFields.Name,
    label: EDepartmentColumnLabel.Name,
    field: (row: IDepartmentTable) => row.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.departmentNames(),
    },
  },
  description: {
    name: "description",
    label: EDepartmentColumnLabel.Description,
    field: (row: IDepartmentTable) => row.description,
    format: (val: string) => val || "-",
    sortable: false,
    align: "left",
    customCell: ECustomCell.DESCRIPTION_COL,
    hideFilter: true,
    sanitizeCell: true,
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: EDepartmentColumnLabel.Cluster,
    field: (row: IDepartmentTable) => row.clusterId,
    sortable: true,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
  },
  projects: {
    name: EDepartmentColumnName.Projects,
    label: EDepartmentColumnLabel.Projects,
    field: (row: IDepartmentTable) => row.children,
    sortable: false,
    align: "left",
    format: (projects: Array<DepartmentProjectChild>): Array<string> | string => {
      if (!projects) return "-";
      return projects.map((p: DepartmentProjectChild) => p.name);
    },
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "projects-clicked" },
    hideFilter: true,
    hideExport: true,
  },
  nodePools: {
    name: EDepartmentColumnName.NodePools,
    label: EDepartmentColumnLabel.NodePools,
    field: (row: IDepartmentTable) => row.resources,
    sortable: false,
    align: "left",
    format: (resources: Resources[]): string[] => resources.map((resources) => resources.nodePool?.name || ""),
    customCell: ECustomCell.LIST_COL,
    customCellEvent: { emitName: "node-pools-clicked" },
    hideFilter: true,
    hideExport: true,
  },
  accessRules: {
    name: EDepartmentColumnName.AccessRules,
    label: EDepartmentColumnLabel.AccessRules,
    field: (row: IDepartmentTable) => row.rolesNames,
    sortable: false,
    align: "left",
    customCell: ECustomCell.ROLE_ASSIGMENT_COL,
    customCellEvent: { emitName: "access-rules-clicked" },
    hideFilter: true,
    hideExport: true,
  },
  allocatedGpus: {
    name: DepartmentFilterSortFields.GpuAllocated,
    label: EDepartmentColumnLabel.AllocatedGpus,
    field: (row: IDepartmentTable) => row.status?.quotaStatus?.allocated?.gpu,
    sortable: true,
    align: "left",
    format: (allocatedGpus: number | undefined) =>
      allocatedGpus ? resourceUtil.getResourceDisplayValue(allocatedGpus, EResourceType.GPU) : "-",
    filter: {
      type: EColumnFilterType.Numeric,
    },
  },
  averageGpuAllocation: {
    name: EOrgUnitOverTimeColumnName.AverageGpuAllocation,
    label: EDepartmentColumnLabel.AverageGpuAllocation,
    field: (row: IDepartmentTable) => row.averageGpuAllocation,
    sortable: true,
    align: "left",
    format: (averageGpuAllocation: number | null) => (averageGpuAllocation ? numberFormat(averageGpuAllocation) : "-"),
    customHeader: ECustomHeader.RadioButtonsTableHeader,
    customHeaderEventName: "avg-resource-timeframe-changed",
    customHeaderConfig: {
      ...ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
      selectedValueFilterByKey: AVG_RESOURCE_TIMEFRAME_FILTER_KEY,
    } as ICustomHeaderRadioButtonsConfig,
    filter: {
      type: EColumnFilterType.Numeric,
    },
    hideExport: true,
  },
  averageGpuUtilization: {
    name: EOrgUnitOverTimeColumnName.AverageGpuUtilization,
    label: EDepartmentColumnLabel.AverageGpuUtilization,
    field: (row: IDepartmentTable) =>
      row.averageGpuUtilization !== null && row.averageGpuUtilization !== undefined
        ? parseFloat(row.averageGpuUtilization.toFixed(0))
        : null,
    format: (averageGpuUtilization: number | null) => orgUnitUtil.getProgressColOptions(averageGpuUtilization),
    customCell: ECustomCell.PROGRESS_COL,
    sortable: true,
    align: "left",
    customHeader: ECustomHeader.RadioButtonsTableHeader,
    customHeaderEventName: "avg-resource-timeframe-changed",
    customHeaderConfig: {
      ...ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
      selectedValueFilterByKey: AVG_RESOURCE_TIMEFRAME_FILTER_KEY,
    } as ICustomHeaderRadioButtonsConfig,
    filter: {
      type: EColumnFilterType.Numeric,
    },
    hideExport: true,
  },
  averageGpuMemoryUtilization: {
    name: EOrgUnitOverTimeColumnName.AverageGpuMemoryUtilization,
    label: EDepartmentColumnLabel.AverageGpuMemoryUtilization,
    field: (row: IDepartmentTable) =>
      row.averageGpuMemoryUtilization !== null && row.averageGpuMemoryUtilization !== undefined
        ? parseFloat(row.averageGpuMemoryUtilization.toFixed(0))
        : null,
    format: (averageGpuMemoryUtilization: number | null) =>
      orgUnitUtil.getProgressColOptions(averageGpuMemoryUtilization),
    customCell: ECustomCell.PROGRESS_COL,
    sortable: true,
    align: "left",
    customHeader: ECustomHeader.RadioButtonsTableHeader,
    customHeaderEventName: "avg-resource-timeframe-changed",
    customHeaderConfig: {
      ...ORG_UNIT_OVERTIME_CUSTOM_HEADER_CONFIG,
      selectedValueFilterByKey: AVG_RESOURCE_TIMEFRAME_FILTER_KEY,
    } as ICustomHeaderRadioButtonsConfig,
    filter: {
      type: EColumnFilterType.Numeric,
    },
    hideExport: true,
  },
  allocatedCpu: {
    name: EDepartmentColumnName.AllocatedCpu,
    label: EDepartmentColumnLabel.AllocatedCpu,
    field: (row: IDepartmentTable) => row.status?.quotaStatus?.allocated?.cpu,
    sortable: false,
    align: "left",
    format: (allocatedCpu: number | undefined) =>
      allocatedCpu ? resourceUtil.getResourceDisplayValue(allocatedCpu, EResourceType.CPU) : "-",
    hideFilter: true,
  },
  allocatedMemory: {
    name: EDepartmentColumnName.AllocatedMemory,
    label: EDepartmentColumnLabel.AllocatedMemory,
    field: (row: IDepartmentTable) => row.status?.quotaStatus?.allocated?.memory,
    sortable: false,
    align: "left",
    format: (allocatedMemory: number | undefined) =>
      allocatedMemory ? memoryFormat(convertToBytes(allocatedMemory, "MB")) : "-",
    hideFilter: true,
  },
  gpuUtilization: {
    name: EDepartmentColumnName.GpuUtilization,
    label: EDepartmentColumnLabel.GpuUtilization,
    field: (row: IDepartmentTable) => row.status?.quotaStatus?.allocated?.gpu,
    sortable: false,
    align: "left",
    format: (allocatedGpus: number | undefined, department: IDepartmentTable) =>
      orgUnitUtil.getAllocationRatioDisplayValue(allocatedGpus || 0, department?.totalResources),
    hideFilter: true,
  },
  cpuUtilization: {
    name: EDepartmentColumnName.CpuUtilization,
    label: EDepartmentColumnLabel.CpuUtilization,
    field: (row: IDepartmentTable) => row.status?.quotaStatus?.allocated?.cpu,
    sortable: true,
    align: "left",
    format: (allocatedCpu: number | undefined, department: IDepartmentTable) =>
      toPercent(allocatedCpu || 0, department.totalResources?.cpuQuota || 0),
    hideFilter: true,
  },
  memoryUtilization: {
    name: EDepartmentColumnName.MemoryUtilization,
    label: EDepartmentColumnLabel.MemoryUtilization,
    field: (row: IDepartmentTable) => row.status?.quotaStatus?.allocated?.memory,
    sortable: false,
    align: "left",
    format: (allocatedMemory: number | undefined, department: IDepartmentTable) =>
      toPercent(resourceUtil.fromBytesToMiB(allocatedMemory || 0), department.totalResources?.memoryQuota || 0),
    hideFilter: true,
  },
  gpuQuota: {
    name: EDepartmentColumnName.GpuQuota,
    label: EDepartmentColumnLabel.GpuQuota,
    field: (row: IDepartmentTable): IDepartmentTable => row,
    sortable: false,
    align: "left",
    customCell: ECustomCell.DEPARTMENT_QUOTA_COL,
    hideFilter: true,
    exportFormat: (row: IDepartmentTable) => row.totalResources?.gpuQuota,
  },
  cpuQuota: {
    name: EDepartmentColumnName.CpuQuota,
    label: EDepartmentColumnLabel.CpuQuota,
    field: (row: IDepartmentTable) => row.totalResources?.cpuQuota,
    sortable: false,
    align: "left",
    format: (cpuQuota: number | null | undefined) => resourceUtil.getResourceDisplayValue(cpuQuota, EResourceType.CPU),
    hideFilter: true,
    exportFormat: (row: IDepartmentTable) => row.totalResources?.cpuQuota,
  },
  memoryQuota: {
    name: EDepartmentColumnName.MemoryQuota,
    label: EDepartmentColumnLabel.MemoryQuota,
    field: (row: IDepartmentTable) => row.totalResources?.memoryQuota,
    sortable: false,
    align: "left",
    format: (memoryQuota: number | null | undefined) => {
      if (memoryQuota === null || memoryQuota === undefined) return EResourceState.Unlimited;
      const memoryQuotaInMib = resourceUtil.fromMbToMib(memoryQuota);
      return resourceUtil.getResourceDisplayValue(memoryQuotaInMib, EResourceType.MEMORY);
    },
    hideFilter: true,
    exportFormat: (row: IDepartmentTable) => row.totalResources?.memoryQuota,
  },
  createdAt: {
    name: DepartmentFilterSortFields.CreatedAt,
    label: EDepartmentColumnLabel.CreatedAt,
    field: (row: IDepartmentTable) => row.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string) => (createdAt ? dateUtil.dateAndTimeFormat(new Date(createdAt)) : "-"),
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  workloads: viewLinkColumn({
    name: EDepartmentColumnName.Workloads,
    label: EDepartmentColumnLabel.Workloads,
    emitName: "workloads-clicked",
  }),
};

export const departmentIndexColumns: Array<ITableColumn> = [
  { ...allDepartmentColumnsMap.departmentName, display: true, mandatory: true },
  { ...allDepartmentColumnsMap.description, display: false },
  { ...allDepartmentColumnsMap.cluster, display: true },
  { ...allDepartmentColumnsMap.projects, display: true },
  { ...allDepartmentColumnsMap.nodePools, display: true },
  { ...allDepartmentColumnsMap.accessRules, display: true },
  { ...allDepartmentColumnsMap.gpuQuota, display: true },
  { ...allDepartmentColumnsMap.allocatedGpus, display: true },
  { ...allDepartmentColumnsMap.averageGpuAllocation, display: true },
  { ...allDepartmentColumnsMap.averageGpuUtilization, display: true },
  { ...allDepartmentColumnsMap.averageGpuMemoryUtilization, display: true },
  { ...allDepartmentColumnsMap.allocatedCpu, display: false },
  { ...allDepartmentColumnsMap.allocatedMemory, display: false },
  { ...allDepartmentColumnsMap.gpuUtilization, display: false },
  { ...allDepartmentColumnsMap.cpuUtilization, display: false },
  { ...allDepartmentColumnsMap.memoryUtilization, display: false },
  { ...allDepartmentColumnsMap.cpuQuota, display: false },
  { ...allDepartmentColumnsMap.memoryQuota, display: false },
  { ...allDepartmentColumnsMap.workloads, display: true },
  { ...allDepartmentColumnsMap.createdAt, display: true },
];

export const departmentDependentColumns = {
  cpu: [
    allDepartmentColumnsMap.cpuQuota.name,
    allDepartmentColumnsMap.memoryQuota.name,
    allDepartmentColumnsMap.memoryUtilization.name,
    allDepartmentColumnsMap.cpuUtilization.name,
    allDepartmentColumnsMap.allocatedCpu.name,
    allDepartmentColumnsMap.allocatedMemory.name,
  ],
  nodePools: [allDepartmentColumnsMap.nodePools.name],
  accessRules: [allDepartmentColumnsMap.accessRules.name],
  cluster: [allDepartmentColumnsMap.cluster.name],
};
