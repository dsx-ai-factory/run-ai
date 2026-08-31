/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Department } from "@/swagger-models/org-unit-service-client";
import type { IFilterBy } from "@/models/filter.model";
import type { AVG_RESOURCE_TIMEFRAME_FILTER_KEY } from "@/models/org-unit.model";

export const DEFAULT_DEPARTMENT_NAME = "default";

export interface IDepartmentTable extends Department {
  rolesNames?: string[];
  averageGpuAllocation?: number;
  averageGpuUtilization?: number;
  averageGpuMemoryUtilization?: number;
}

export enum EDepartmentColumnName {
  DepartmentName = "name",
  NodePools = "node-pools",
  GpuQuota = "gpu-quota",
  AccessRules = "accessRules",
  Projects = "projects",
  CpuQuota = "cpu-quota",
  MemoryQuota = "memory-quota",
  AllocatedCpu = "allocated-cpu",
  AllocatedMemory = "allocated-memory",
  AllocatedGpus = "allocated-gpus",
  GpuUtilization = "gpu-utilization",
  CpuUtilization = "cpu-utilization",
  MemoryUtilization = "memory-utilization",
  CreatedAt = "createdAt",
  Workloads = "workloads",
}

export interface IDepartmentTableFilterBy extends IFilterBy {
  [AVG_RESOURCE_TIMEFRAME_FILTER_KEY]: string;
}
