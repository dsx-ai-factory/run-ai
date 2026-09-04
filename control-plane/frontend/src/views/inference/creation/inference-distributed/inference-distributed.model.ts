/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import type { ScopeType } from "@/swagger-models/workloads-client";

export enum EDistributedInferenceType {
  Leader = "leader",
  Workers = "workers",
  WorkloadSetup = "workload-setup",
}

export type TDistributedInferenceFormMetaInfo = {
  clusterId: string;
  projectId: string;
  // template exclusive (not in use for workload form)
  scopeType?: ScopeType;
  scopeId?: string;
};

export const distributedInferenceSteps: IRunaiWizardLayoutStep[] = [
  {
    name: EDistributedInferenceType.Leader,
    title: "Leader",
    done: false,
  },
  {
    name: EDistributedInferenceType.Workers,
    title: "Workers",
    done: false,
  },
  {
    name: EDistributedInferenceType.WorkloadSetup,
    title: "Workload setup",
    done: false,
  },
];
