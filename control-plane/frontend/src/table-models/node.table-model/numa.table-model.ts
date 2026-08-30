import { type NumaNodeDetails } from "@/swagger-models/cluster-service-client";
import type { ITableColumn } from "@/models/table.model";
import { tableFormatUtil } from "@/utils/table-format.util/table-format.util";
import { formatMemoryValue } from "@/table-models/node.table-model/node.table-model";

/**** Numa nodes modal ****/
export enum NumaModalColumnName {
  Name = "name",
  TotalGpus = "totalGpus",
  AllocatedGpus = "allocatedGpus",
  TotalCpuCores = "totalCpuCores",
  TotalCpuMemoryBytes = "totalCpuMemoryBytes",
  AllocatedCpuCores = "allocatedCpuCores",
  AllocatedCpuMemoryBytes = "allocatedCpuMemoryBytes",
}

export enum NumaModalColumnLabel {
  Name = "Name",
  TotalGpus = "Total GPU devices",
  AllocatedGpus = "Allocated GPUs",
  TotalCpuCores = "CPU (Cores)",
  TotalCpuMemoryBytes = "CPU memory",
  AllocatedCpuCores = "Allocated CPU (Cores)",
  AllocatedCpuMemoryBytes = "Allocated CPU memory",
}

export const numaModalColumns: ITableColumn[] = [
  {
    name: NumaModalColumnName.Name,
    label: NumaModalColumnLabel.Name,
    field: (row: NumaNodeDetails): string => row.name,
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: NumaModalColumnName.TotalGpus,
    label: NumaModalColumnLabel.TotalGpus,
    field: (row: NumaNodeDetails): number => row.totalGpus,
    format: (totalGpus: number): string => tableFormatUtil.getFormattedNumberOrDefault(totalGpus),
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: NumaModalColumnName.AllocatedGpus,
    label: NumaModalColumnLabel.AllocatedGpus,
    field: (row: NumaNodeDetails): number => row.allocatedGpus,
    format: (allocatedGpus: number): string => tableFormatUtil.getFormattedNumberOrDefault(allocatedGpus),
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: NumaModalColumnName.TotalCpuCores,
    label: NumaModalColumnLabel.TotalCpuCores,
    field: (row: NumaNodeDetails): number => row.totalCpuCores,
    format: (totalCpuCores: number): string => tableFormatUtil.getFormattedNumberOrDefault(totalCpuCores),
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: NumaModalColumnName.TotalCpuMemoryBytes,
    label: NumaModalColumnLabel.TotalCpuMemoryBytes,
    field: (row: NumaNodeDetails): number => row.totalCpuMemoryBytes,
    format: (totalCpuMemoryBytes: number): string => formatMemoryValue(totalCpuMemoryBytes),
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: NumaModalColumnName.AllocatedCpuCores,
    label: NumaModalColumnLabel.AllocatedCpuCores,
    field: (row: NumaNodeDetails): number => row.allocatedCpuCores,
    format: (allocatedCpuCores: number): string => tableFormatUtil.getFormattedNumberOrDefault(allocatedCpuCores),
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: NumaModalColumnName.AllocatedCpuMemoryBytes,
    label: NumaModalColumnLabel.AllocatedCpuMemoryBytes,
    field: (row: NumaNodeDetails): number => row.allocatedCpuMemoryBytes,
    format: (allocatedCpuMemoryBytes: number): string => formatMemoryValue(allocatedCpuMemoryBytes),
    sortable: true,
    align: "left",
    display: true,
  },
];
