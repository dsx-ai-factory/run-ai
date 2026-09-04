/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IUIWorkloadCreation } from "@/models/workload.model";
import type { TrainingV2 } from "@/swagger-models/assets-service-client";
import { WorkloadDesiredPhase } from "@/swagger-models/assets-service-client";

const trainingUIMock: IUIWorkloadCreation = {
  name: "dssddsdssd",
  namespace: "",
  clusterId: "1c75d0b6-947a-4526-a989-9b51e350bdd7",
  projectId: 4500003,
  assets: {
    environment: null,
    compute: null,
    datasources: [],
  },
  specificEnv: {
    runAsUid: null,
    runAsGid: null,
    nodeType: null,
    allowOverQuota: false,
    nodePools: ["default"],
    autoDeletionTimeAfterCompletionSeconds: 2592000,
    backoffLimit: 6,
  },
  distributed: {
    numWorkers: 1,
    distFramework: "PyTorch",
    master: null,
  },
};

const loadedTrainingModelWithDefaultsAndPolicyMock: TrainingV2 = {
  meta: {
    name: "example",
    requestedName: "example",
    workloadId: "example-id",
    projectId: "example-project-id",
    clusterId: "example-cluster-id",
    createdBy: "example-user",
    createdAt: new Date().toISOString(),
    desiredPhase: WorkloadDesiredPhase.Deleted,
  },
  assets: {
    environment: {
      id: "asd",
      name: "test",
    },
    compute: null,
    datasources: [],
  },
  distributed: {
    numWorkers: 1,
    master: {
      policyInfo: {
        rules: {
          specificRunParams: {
            command: {
              options: [{ value: "start-notebook.sh" }],
            },
          },
        },
      },
    },
  },
};

export const distributedAssetsEditMockedData = {
  trainingUIMock,
  loadedTrainingModelWithDefaultsAndPolicyMock,
};
