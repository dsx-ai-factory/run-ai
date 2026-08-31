/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ISelectOption } from "@/models/global.model";

export enum EQuotaEntity {
  project = "project",
  department = "department",
}

export enum EResourceType {
  GPU = "gpu",
  GPU_LIMIT = "gpuLimit",
  CPU = "cpu",
  MEMORY = "memory",
}

export enum EResourceState {
  Unlimited = "Unlimited",
}

export enum EResourceLabel {
  GPU = "GPU devices",
  CPU = "CPUs (Cores)",
  MEMORY = "CPU memory",
}

export const enum EMemoryUnit {
  MB = "MB",
  GB = "GB",
}

export enum EOverQuotaPriority {
  None = -1,
  Lowest = 0,
  Low = 1,
  Medium = 2,
  High = 3,
}

export enum EOverQuotaPriorityLabel {
  None = "None",
  Lowest = "Lowest",
  Low = "Low",
  Medium = "Medium",
  High = "High",
}

export enum EOverQuotaLabel {
  Enabled = "Enabled",
  Disabled = "Disabled",
}

export const DEFAULT_NODE_POOL_NAME = "default";
export const CPU_VALUE_FACTOR = 1000;
export const RESOURCE_MAX_ALLOWED_INFINITE_VALUE = null;
export const RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC = -1;
export const memoryUnitOptions: Array<string> = ["MB", "GB"];
export const OVER_QUOTA_ENABLED_VALUE = 2;
export const OVER_QUOTA_DISABLED_VALUE = 0;
export const EMPTY_PRIORITY_VALUE = "Empty";
export const MIN_PRESENTED_GPU = 100;
export const MIN_PRESENTED_CPU_CORES = 100;
export const MIN_PRESENTED_CPU_MEMORY_MIB = 10000;
export const MAX_PRESENTED_CPU_CORES = 10000;
export const MAX_PRESENTED_CPU_MEMORY_MIB = 9536.74; //10GB

export const overQuotaOptions: Array<ISelectOption> = [
  {
    label: EOverQuotaLabel.Enabled,
    value: OVER_QUOTA_ENABLED_VALUE,
  },
  {
    label: EOverQuotaLabel.Disabled,
    value: OVER_QUOTA_DISABLED_VALUE,
  },
];

export const defaultQuotaOption: ISelectOption = {
  label: EOverQuotaLabel.Enabled,
  value: OVER_QUOTA_ENABLED_VALUE,
};

export const defaultQuotaPriorityOption: ISelectOption = {
  label: EOverQuotaPriorityLabel.Medium,
  value: EOverQuotaPriority.Medium,
};

export const overQuotaPriorityOptions: Array<ISelectOption> = [
  {
    label: EOverQuotaPriorityLabel.None,
    value: EOverQuotaPriority.None,
  },
  {
    label: EOverQuotaPriorityLabel.Lowest,
    value: EOverQuotaPriority.Lowest,
    toolTip:
      "The project will get unused resources only after all other projects with higher priority receive their requests. Any higher-priority projects that are added will claim this project's over quota resources and likely cause its workloads to become pending.",
  },
  {
    label: EOverQuotaPriorityLabel.Low,
    value: EOverQuotaPriority.Low,
  },
  {
    label: EOverQuotaPriorityLabel.Medium,
    value: EOverQuotaPriority.Medium,
  },
  {
    label: EOverQuotaPriorityLabel.High,
    value: EOverQuotaPriority.High,
  },
];

export type NodePoolResourcesSumRecord = Record<string, INodePoolResourcesSum>;
export type DepartmentNodePoolResourcesSumRecord = Record<string, IDepartmentNodePoolResourcesSum>;
export type INodePoolIsOverQuotaRecord = Record<string, INodePoolIsOverQuota>;
export type NodePoolAllocatedNonPreemptibleSumRecord = Record<string, INodePoolAllocatedNonPreemptibleSum>;

export interface IResource {
  deserved: number | null;
  maxAllowed: number | null;
  overQuotaWeight: number | null;
}

export interface IResourcePermissions {
  users: string[];
}

export interface INodePoolResourcesSum {
  gpu: number;
  cpu: number;
  memory: number;
}

export interface IDepartmentNodePoolResourcesSum {
  gpu: number;
  cpu?: number | null;
  memory?: number | null;
}

export interface INodePoolAllocatedNonPreemptibleSum {
  gpu: number;
  cpu: number;
  memory: number;
}
export interface INodePoolIsOverQuota {
  gpu: boolean;
  cpu: boolean;
  memory: boolean;
}

//telemetry
export interface IClusterTotalResources {
  cpuCores: number;
  memoryBytes: number;
  gpu: number;
}

export const k8sUnitMultiplier: Record<string, number> = {
  "": 1, // No suffix
  m: 1e-3, // Milli (CPU)
  k: 1e3, // Kilo
  K: 1e3, // Kilo (alternative)
  M: 1e6, // Mega
  G: 1e9, // Giga
  T: 1e12, // Tera
  P: 1e15, // Peta
  E: 1e18, // Exa
  Ki: 1024, // Kibi
  Mi: 1024 ** 2, // Mebi
  Gi: 1024 ** 3, // Gibi
  Ti: 1024 ** 4, // Tebi
  Pi: 1024 ** 5, // Pebi
  Ei: 1024 ** 6, // Exbi
};
