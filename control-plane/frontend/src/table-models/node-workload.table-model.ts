/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ITableColumn } from "@/models/table.model";
import { memoryFormat } from "@/utils/format.util";
import { tableNumberFormat } from "@/utils/table-format.util";

export const nodeWorkloadColumns = [
  {
    name: "name",
    label: "Workload",
    field: "name",
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: "type",
    label: "Type",
    field: "type",
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: "project",
    label: "Project",
    field: "project",
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: "allocatedGpu",
    label: "Allocated GPU",
    field: "allocatedGpu",
    sortable: true,
    align: "left",
    display: true,
    format: tableNumberFormat(),
  },
  {
    name: "gpuIndex",
    label: "GPU index",
    field: "gpuIndex",
    sortable: true,
    align: "left",
    display: true,
    format: (val: []) => (val ? val.join(",") : "-"),
  },
  {
    name: "allocatedCpu",
    label: "Allocated CPU (Cores)",
    field: "allocatedCpu",
    sortable: true,
    align: "left",
    display: true,
    format: tableNumberFormat(),
  },
  {
    name: "allocatedMemory",
    label: "Allocated CPU memory",
    field: "allocatedMemory",
    sortable: true,
    align: "left",
    display: true,
    format: memoryFormat,
  },
] as Array<ITableColumn>;

export const nodeWorkloadOldVersionColumns = [
  {
    name: "name",
    label: "Workload",
    field: "name",
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: "type",
    label: "Type",
    field: "type",
    sortable: true,
    align: "left",
    display: true,
  },
  {
    name: "project",
    label: "Project",
    field: "project",
    sortable: true,
    align: "left",
    display: true,
  },
] as Array<ITableColumn>;
