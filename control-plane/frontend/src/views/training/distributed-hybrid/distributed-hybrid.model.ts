/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IRunaiWizardLayoutStep } from "@/models/runai-wizard-layout.model";
import { EDistributedReplicaType } from "@/models/workload.model";
import type { DistributedFramework, ScopeType } from "@/swagger-models/workloads-client";

export type TDistributedFormMetaInfo = {
  clusterId: string;
  projectId: string;
  sameAsWorker: boolean;
  framework?: DistributedFramework;
  // template exclusive (not in use for workload form)
  scopeType?: ScopeType;
  scopeId?: string;
};

export enum EDistributedHybridStepName {
  Worker = EDistributedReplicaType.Worker,
  Master = EDistributedReplicaType.Master,
  Setup = "Setup",
}

const distributedHybridStepDefinitions: Readonly<IRunaiWizardLayoutStep[]> = [
  {
    name: EDistributedHybridStepName.Worker,
    title: "Workers",
    done: false,
  },
  {
    name: EDistributedHybridStepName.Master,
    title: "Master",
    done: false,
  },
  {
    name: EDistributedHybridStepName.Setup,
    title: "Workload setup",
    done: false,
  },
];

export const distributedHybridSteps = (): IRunaiWizardLayoutStep[] =>
  distributedHybridStepDefinitions.map((step) => ({ ...step }));
