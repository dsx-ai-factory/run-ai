/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { INodePoolTotals } from "@/models/org-unit-forms.model";

export const mockProject = {
  resources: [
    {
      nodePool: { name: "default", id: "1" },
      cpu: { deserved: 1000 },
    },
  ],
  defaultNodePools: ["default"],
  status: {
    nodePoolQuotaStatuses: [{ nodePoolName: "default", allocatedNonPreemptible: { cpu: 500 } }],
  },
};

export const mockDepartment = {
  resources: [
    {
      nodePool: { name: "default", id: "1" },
      cpu: { deserved: 2000 },
    },
  ],
};

// Mock data for department context testing
export const mockDepartmentResources = [
  {
    nodePool: { name: "default", id: "1" },
    cpu: { deserved: 5000 }, // 5 cores in millicores
    memory: { deserved: 1024, units: "MB" },
  },
];

export const mockDepartmentNodePoolQuotaStatuses = [
  {
    nodePoolName: "default",
    allocatedNonPreemptible: { cpu: 1000, memory: 256 }, // 1 core, 256 MB
  },
];

export const mockNodePoolTotals: INodePoolTotals = {
  default: {
    cpuCores: 10, // 10 cores total in node pool
    memoryBytes: 10737418240, // 10 GB in bytes
    totalGpus: 0,
  },
};

export const mockSubordinateResources = [
  {
    nodePool: { name: "default", id: "1" },
    cpu: { deserved: 3000 }, // 3 cores in millicores (subordinate projects' total)
    memory: { deserved: 512, units: "MB" },
    gpu: { deserved: 0 },
  },
];
