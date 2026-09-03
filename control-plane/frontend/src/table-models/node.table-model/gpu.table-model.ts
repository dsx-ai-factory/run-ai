/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ITableColumn } from "@/models/table.model";
import { dateUtil } from "@/utils/date.util";
import { formatMemoryValue } from "@/table-models/node.table-model/node.table-model";

/**** Gpus modal ****/
export interface IGpuModalRow {
  index: number;
  allocatedCompute: number;
  allocatedMemory: number;
  usedMemory: number;
  computeUtilization: number;
  memoryUtilization: number;
  idleTime: string;
}

export enum GpuModalColumnName {
  Index = "index",
  Status = "status",
  AllocatedCompute = "allocatedCompute",
  AllocatedMemory = "allocatedMemory",
  UsedMemory = "usedMemory",
  ComputeUtilization = "computeUtilization",
  MemoryUtilization = "memoryUtilization",
  IdleTime = "idleTime",
}

export enum GpuModalColumnLabel {
  Index = "Index",
  Status = "Status",
  AllocatedCompute = "Allocated compute",
  AllocatedMemory = "Allocated memory",
  UsedMemory = "Used memory",
  ComputeUtilization = "Compute utilization",
  MemoryUtilization = "Memory utilization",
  IdleTime = "Idle time",
}

export const gpuModalColumns: ITableColumn[] = [
  {
    name: GpuModalColumnName.Index,
    label: GpuModalColumnLabel.Index,
    field: (row: IGpuModalRow): number => row.index,
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: GpuModalColumnName.AllocatedCompute,
    label: GpuModalColumnLabel.AllocatedCompute,
    field: (row: IGpuModalRow): string => row.allocatedCompute.toFixed(2),
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: GpuModalColumnName.AllocatedMemory,
    label: GpuModalColumnLabel.AllocatedMemory,
    field: (row: IGpuModalRow): number => row.allocatedMemory,
    format: (allocatedMemory: number): string => formatMemoryValue(allocatedMemory),
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: GpuModalColumnName.UsedMemory,
    label: GpuModalColumnLabel.UsedMemory,
    field: (row: IGpuModalRow): number => row.usedMemory,
    format: (usedMemory: number): string => formatMemoryValue(usedMemory),
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: GpuModalColumnName.ComputeUtilization,
    label: GpuModalColumnLabel.ComputeUtilization,
    field: (row: IGpuModalRow): number => row.computeUtilization,
    format: (computeUtilization: number): string => `${computeUtilization.toFixed(2)}%`,
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: GpuModalColumnName.MemoryUtilization,
    label: GpuModalColumnLabel.MemoryUtilization,
    field: (row: IGpuModalRow): number => row.memoryUtilization,
    format: (memoryUtilization: number): string => `${memoryUtilization.toFixed(2)}%`,
    sortable: true,
    align: "left",
  },
  {
    name: GpuModalColumnName.IdleTime,
    label: GpuModalColumnLabel.IdleTime,
    field: (row: IGpuModalRow): string => row.idleTime,
    format: (idleTime: string): string => {
      if (idleTime === "0") {
        return "-";
      }
      return dateUtil.formatDuration(Number(idleTime));
    },
    sortable: true,
    align: "left",
    display: true,
    sort: (a: string, b: string): number => Number(a) - Number(b),
  },
];
