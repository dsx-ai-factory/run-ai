/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IWorkloadPriorityData } from "@/components/section/general-section/general-section.model";
import type { IComputeSectionNodePoolsPolicy } from "@/components/section/compute-resource-section/node-pools-resource-section";
import type { IntegerRules, StringOption } from "@/swagger-models/assets-service-client";
import type {
  Preemptibility,
  DistributedInferenceStartupPolicy,
  WorkspaceSpecSpecAllOfCompute,
} from "@/swagger-models/workloads-client";
import type { UnifiedPolicyInfoPerReplicaTypeRules } from "@/swagger-models/workloads-client";

// Workload Setup Model - the main data model for the workload setup step
export interface IWorkloadSetupModel {
  replicas: number;
  workers: number;
  leaderCompute?: WorkspaceSpecSpecAllOfCompute;
  workerCompute?: WorkspaceSpecSpecAllOfCompute;
  priorityClass: string | null;
  preemptibility: Preemptibility | null;
  startupPolicy: DistributedInferenceStartupPolicy | null;
  nodePools?: string[];
}

// Compute Section
export interface IComputeSectionModel {
  replicas: number;
  workers: number;
  leaderCompute?: WorkspaceSpecSpecAllOfCompute;
  workerCompute?: WorkspaceSpecSpecAllOfCompute;
}

export interface IComputeSectionConfig {
  replicasPolicyRules?: IntegerRules | null;
  nodePoolsResources?: StringOption[];
  nodePoolsPolicy?: IComputeSectionNodePoolsPolicy | null;
}

// General Section
export interface IGeneralSectionModel {
  priorityClass: string | null;
  preemptibility: Preemptibility | null;
  startupPolicy: DistributedInferenceStartupPolicy | null;
}

export interface IGeneralSectionFeatures {
  supportPriorityClass: boolean;
  supportPreemptibility: boolean;
}

export interface IGeneralSectionConfig {
  workloadPriorityData?: IWorkloadPriorityData;
  showQuotaWarning: boolean;
  policyRules?: UnifiedPolicyInfoPerReplicaTypeRules | null;
}
