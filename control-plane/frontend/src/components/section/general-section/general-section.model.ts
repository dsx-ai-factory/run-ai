/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import type { MpiLauncherCreationPolicy, ServingConfiguration } from "@/swagger-models/assets-service-client";
import type { CleanPodPolicy, ClusterPriority, Preemptibility } from "@/swagger-models/workloads-client";
import type { Project } from "@/swagger-models/org-unit-service-client";
import type { MnnvlMode } from "@/swagger-models/workloads-manager-client";

export interface IUIGeneralSectionModel {
  allowOverQuota?: boolean | null;
  annotations?: Array<IItemizedListItem>;
  labels?: Array<IItemizedListItem>;
  autoDeletionTimeAfterCompletionSeconds?: number | null;
  backoffLimit?: number | null;
  completions?: number | null;
  parallelism?: number | null;
  terminationGracePeriodSeconds?: number | null;
  cleanPodPolicy?: CleanPodPolicy;
  sshAuthMountPath?: string | null;
  mpiLauncherCreationPolicy?: MpiLauncherCreationPolicy | null;
  servingConfiguration?: ServingConfiguration | null;
  priorityClass?: string | null;
  preemptibility?: Preemptibility | null;
  mnnvl?: MnnvlMode | null;
}

export interface IGeneralSectionDisableFields {
  annotations?: string;
  labels?: string;
}

export interface IWorkloadPriorityData {
  currProject?: Project;
  selectedNodePoolsResources?: string[];
  totalGpuRequest?: number;
  workloadPriorities?: ClusterPriority[];
  supportPriorityClass?: boolean;
  supportSemiPreemptibility?: boolean;
}

export interface IOverrideHints {
  priority?: boolean;
  preemptibility?: boolean;
}
