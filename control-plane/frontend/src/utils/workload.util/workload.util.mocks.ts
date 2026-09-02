/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { EWorkloadCategory, type IUIWorkloadCreation } from "@/models/workload.model";
import {
  type Training,
  type WorkloadCreationRequest,
  DistributedFramework,
  type TrainingV2,
} from "@/swagger-models/assets-service-client";
import type { IAdvancedFilterModel, IEnumBasedFilterModel } from "@/models/filter.model";

import { type Workload } from "@/swagger-models/workloads-client";

export const trainingSingleNodeUIModel: IUIWorkloadCreation = {
  name: "tr305",
  projectId: 8,
  namespace: "runai-p100",
  clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
  assets: {
    environment: "55c1bdad-b01f-4635-a883-4a20f9c6763f",
    compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
    uiVolumes: [],
    datasources: [
      { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
      { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
    ],
  },
  specificEnv: {
    args: "",
    runAsUid: null,
    runAsGid: null,
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    connections: [],
    command: "",
    supplementalGroups: null,
    allowOverQuota: false,
    autoDeletionTimeAfterCompletionSeconds: null,
    annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
    labels: [{ name: "label-key-1", value: "label-value-1" }],
  },
};

export const trainingSingleNodeCreationModel: WorkloadCreationRequest = {
  name: "tr305",
  namespace: "runai-p100",
  clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
  projectId: 8,
  assets: {
    environment: "55c1bdad-b01f-4635-a883-4a20f9c6763f",
    compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
    datasources: [
      { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
      { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
    ],
  },
  specificEnv: {
    args: null,
    command: null,
    runAsUid: undefined,
    runAsGid: undefined,
    supplementalGroups: null,
    connections: [],
    allowOverQuota: false,
    autoDeletionTimeAfterCompletionSeconds: null,
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
    labels: [{ name: "label-key-1", value: "label-value-1" }],
  },
};

export const trainingMultiNodeUIModel: IUIWorkloadCreation = {
  name: "dis100",
  projectId: 8,
  namespace: "runai-p100",
  clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
  assets: {
    environment: "c29bed7e-5009-4162-8c4b-ddb296b7e173",
    compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
    datasources: [
      { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
      { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
    ],
  },
  specificEnv: {
    args: "--args2",
    command: "some-command",
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    connections: [],
    runAsGid: null,
    runAsUid: null,
    supplementalGroups: null,
    allowOverQuota: false,
    autoDeletionTimeAfterCompletionSeconds: null,
    annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
    labels: [{ name: "label-key-1", value: "label-value-1" }],
  },
  distributed: {
    distFramework: "MPI",
    numWorkers: 1,
    master: {
      assets: {
        environment: "c29bed7e-5009-4162-8c4b-ddb296b7e173",
        compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
        datasources: [
          { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
          { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
        ],
      },
      specificEnv: {
        args: "--args2",
        command: "some-command",
        environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
        connections: [],
        runAsGid: null,
        runAsUid: null,
        supplementalGroups: null,
        allowOverQuota: false,
        autoDeletionTimeAfterCompletionSeconds: null,
        annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
        labels: [{ name: "label-key-1", value: "label-value-1" }],
      },
    },
  },
};

export const trainingMultiNodeCreationModel: WorkloadCreationRequest = {
  name: "dis100",
  namespace: "runai-p100",
  clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
  projectId: 8,
  assets: {
    environment: "c29bed7e-5009-4162-8c4b-ddb296b7e173",
    compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
    datasources: [
      { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
      { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
    ],
  },
  specificEnv: {
    args: "--args2",
    command: "some-command",
    runAsUid: undefined,
    runAsGid: undefined,
    supplementalGroups: null,
    connections: [],
    allowOverQuota: false,
    autoDeletionTimeAfterCompletionSeconds: null,
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
    labels: [{ name: "label-key-1", value: "label-value-1" }],
  },
  distributed: {
    distFramework: "MPI",
    numWorkers: 1,
    master: {
      specificEnv: {
        args: "--args2",
        command: "some-command",
        runAsUid: undefined,
        runAsGid: undefined,
        supplementalGroups: null,
        connections: [],
        allowOverQuota: false,
        autoDeletionTimeAfterCompletionSeconds: null,
        environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
        annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
        labels: [{ name: "label-key-1", value: "label-value-1" }],
      },
      assets: {
        environment: "c29bed7e-5009-4162-8c4b-ddb296b7e173",
        compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
        datasources: [
          { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
          { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
        ],
      },
    },
  },
};

export const trainingMultiNodeDifferentMasterUIModel: IUIWorkloadCreation = {
  name: "dis100",
  projectId: 8,
  namespace: "runai-p100",
  clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
  assets: {
    environment: "c29bed7e-5009-4162-8c4b-ddb296b7e173",
    compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
    datasources: [
      { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
      { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
    ],
  },
  specificEnv: {
    args: "--args2",
    command: "some-command",
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    connections: [],
    runAsGid: null,
    runAsUid: null,
    supplementalGroups: null,
    allowOverQuota: false,
    autoDeletionTimeAfterCompletionSeconds: null,
    annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
    labels: [{ name: "label-key-1", value: "label-value-1" }],
  },
  distributed: {
    distFramework: "MPI",
    numWorkers: 1,
    master: {
      assets: {
        environment: "c29bed7e-5009-4162-8c4b-ddb296b7e173",
        compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
        datasources: [
          { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
          { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
        ],
      },
      specificEnv: {
        args: "--args2",
        command: "some-command",
        environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
        connections: [],
        runAsGid: null,
        runAsUid: null,
        supplementalGroups: null,
        allowOverQuota: false,
        autoDeletionTimeAfterCompletionSeconds: null,
        annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
        labels: [{ name: "label-key-1", value: "label-value-1" }],
      },
    },
  },
};

export const trainingMultiNodeDifferentMasterCreationModel: WorkloadCreationRequest = {
  name: "dis100",
  namespace: "runai-p100",
  clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
  projectId: 8,
  assets: {
    environment: "c29bed7e-5009-4162-8c4b-ddb296b7e173",
    compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
    datasources: [
      { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
      { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
    ],
  },
  specificEnv: {
    args: "--args2",
    command: "some-command",
    runAsUid: undefined,
    runAsGid: undefined,
    supplementalGroups: null,
    connections: [],
    allowOverQuota: false,
    autoDeletionTimeAfterCompletionSeconds: null,
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
    labels: [{ name: "label-key-1", value: "label-value-1" }],
  },
  distributed: {
    distFramework: "MPI",
    numWorkers: 1,
    master: {
      specificEnv: {
        args: "--args2",
        command: "some-command",
        runAsUid: undefined,
        runAsGid: undefined,
        supplementalGroups: null,
        connections: [],
        allowOverQuota: false,
        autoDeletionTimeAfterCompletionSeconds: null,
        environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
        annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
        labels: [{ name: "label-key-1", value: "label-value-1" }],
      },
      assets: {
        environment: "c29bed7e-5009-4162-8c4b-ddb296b7e173",
        compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
        datasources: [
          { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
          { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
        ],
      },
    },
  },
};

export const trainingMultiNodeNoMasterUIModel: IUIWorkloadCreation = {
  name: "dis100",
  projectId: 8,
  namespace: "runai-p100",
  clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
  assets: {
    environment: "c29bed7e-5009-4162-8c4b-ddb296b7e173",
    compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
    datasources: [
      { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
      { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
    ],
  },
  specificEnv: {
    args: "--args2",
    command: "some-command",
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    connections: [],
    runAsGid: null,
    runAsUid: null,
    supplementalGroups: null,
    allowOverQuota: false,
    autoDeletionTimeAfterCompletionSeconds: null,
    annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
    labels: [{ name: "label-key-1", value: "label-value-1" }],
  },
  distributed: {
    distFramework: "MPI",
    numWorkers: 1,
    noMaster: true,
  },
};

export const trainingMultiNodeNoMasterCreationModel: WorkloadCreationRequest = {
  name: "dis100",
  namespace: "runai-p100",
  clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
  projectId: 8,
  assets: {
    environment: "c29bed7e-5009-4162-8c4b-ddb296b7e173",
    compute: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
    datasources: [
      { id: "fc49d65d-0230-4ace-ab87-5b1c2e3de558", kind: "hostPath" },
      { id: "50bddfe8-4d95-4c6a-9969-0918781398ab", kind: "nfs" },
    ],
  },
  specificEnv: {
    args: "--args2",
    command: "some-command",
    connections: [],
    runAsGid: undefined,
    runAsUid: undefined,
    supplementalGroups: null,
    allowOverQuota: false,
    autoDeletionTimeAfterCompletionSeconds: null,
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    annotations: [{ name: "annotation-key-1", value: "annotation-value-1" }],
    labels: [{ name: "label-key-1", value: "label-value-1" }],
  },
  distributed: { distFramework: "MPI", numWorkers: 1, noMaster: true, master: null },
};

export const trainingSingleNodeResponse: Training = {
  meta: {
    id: "138a7f0d-526a-4a3d-8ad5-7d32051c3012",
    tenantId: 1001,
    name: "w100",
    projectId: 8,
    clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
    desiredState: "Submitted",
    currentState: "Submitted",
    createdBy: "runai",
    updatedBy: "runai",
  },
  spec: {
    specificEnv: {
      environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
      connections: [],
    },
    assets: {
      environment: {
        id: "55c1bdad-b01f-4635-a883-4a20f9c6763",
        name: "env101",
      },
      compute: {
        id: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
        name: "cpu-only",
      },
    },
  },
};

export const trainingSingleNodeResponseV2: TrainingV2 = {
  meta: {
    workloadId: "138a7f0d-526a-4a3d-8ad5-7d32051c3012",
    name: "w100",
    projectId: "8",
    requestedName: "w100",
    clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
    desiredPhase: "Running",
    createdBy: "runai",
    createdAt: "",
  },
  specificRunParams: {
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    connections: [],
  },
  assets: {
    environment: {
      id: "55c1bdad-b01f-4635-a883-4a20f9c6763",
      name: "env101",
    },
    compute: {
      id: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
      name: "cpu-only",
    },
  },
};

export const trainingMultiNodeResponse: Training = {
  meta: {
    id: "138a7f0d-526a-4a3d-8ad5-7d32051c3012",
    tenantId: 1001,
    name: "w100",
    projectId: 8,
    clusterId: "aefb179a-4bee-420f-bb00-1da188a155fd",
    desiredState: "Submitted",
    currentState: "Submitted",
    createdBy: "runai",
    updatedBy: "runai",
  },
  spec: {
    specificEnv: {
      environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
      connections: [],
    },
    assets: {
      environment: {
        id: "55c1bdad-b01f-4635-a883-4a20f9c6763",
        name: "env101",
      },
      compute: {
        id: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
        name: "cpu-only",
      },
    },
    distributed: {
      numWorkers: 1,
      distFramework: DistributedFramework.Mpi,
      master: {
        assets: {
          environment: {
            id: "55c1bdad-b01f-4635-a883-4a20f9c6763",
            name: "env101",
          },
          compute: {
            id: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
            name: "cpu-only",
          },
        },
        specificEnv: {
          environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
          connections: [],
        },
      },
    },
  },
};

export const trainingMultiNodeMasterSameAsWorkerResponseV2: TrainingV2 = {
  meta: {
    name: "dist-1",
    requestedName: "dist-1",
    workloadId: "48c2408f-e00c-4643-8f65-375dbfa78eea",
    projectId: "1",
    clusterId: "924f8015-d65f-4236-ac88-6bebe8481192",
    createdBy: "test@run.ai",
    createdAt: "2024-06-23T09:20:32.215022Z",
    desiredPhase: "Running",
  },
  assets: {
    environment: {
      id: "55c1bdad-b01f-4635-a883-4a20f9c6763",
      name: "env101",
    },
    compute: {
      id: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
      name: "cpu-only",
    },
  },
  specificRunParams: {
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    nodePools: ["default"],
    autoDeletionTimeAfterCompletionSeconds: 2592000,
    backoffLimit: 6,
    connections: [],
    allowOverQuota: false,
  },
  distributed: {
    framework: DistributedFramework.Mpi,
    master: { sameAsWorker: true },
    numWorkers: 1,
  },
};
export const trainingMultiNodeCustomMasterResponseV2: TrainingV2 = {
  meta: {
    name: "dist-1",
    requestedName: "dist-1",
    workloadId: "48c2408f-e00c-4643-8f65-375dbfa78eea",
    projectId: "1",
    clusterId: "924f8015-d65f-4236-ac88-6bebe8481192",
    createdBy: "test@run.ai",
    createdAt: "2024-06-23T09:20:32.215022Z",
    desiredPhase: "Running",
  },
  assets: {
    environment: {
      id: "55c1bdad-b01f-4635-a883-4a20f9c6763",
      name: "env101",
    },
    compute: {
      id: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
      name: "cpu-only",
    },
  },
  specificRunParams: {
    environmentVariables: [{ name: "env-key-1", value: "env-value-1" }],
    nodePools: ["default"],
    autoDeletionTimeAfterCompletionSeconds: 2592000,
    backoffLimit: 6,
    connections: [],
    allowOverQuota: false,
  },
  distributed: {
    numWorkers: 1,
    framework: "MPI",
    master: {
      assets: {
        environment: {
          id: "55c1bdad-b01f-4635-a883-4a20f9c6763",
          name: "env101",
        },
        compute: {
          id: "23d90236-4687-47cc-9d0f-b6cffaa59bfb",
          name: "cpu-only",
        },
      },
      specificRunParams: {
        environmentVariables: [{ name: "env-master-key-1", value: "env-master-value-1" }],
        nodePools: ["default"],
        autoDeletionTimeAfterCompletionSeconds: 200000,
        backoffLimit: 4,
        connections: [],
        allowOverQuota: false,
      },
    },
  },
};

export const advancedFilterPhaseMockRunning: IAdvancedFilterModel[] = [
  {
    selectedValues: ["Running"],
    label: "Status",
    name: "phase",
  } as IEnumBasedFilterModel,
];
export const advancedFilterPhaseMockStopped: IAdvancedFilterModel[] = [
  {
    selectedValues: ["Stopped"],
    label: "Status",
    name: "phase",
  } as IEnumBasedFilterModel,
];
export const advancedFilterPhaseMockPending: IAdvancedFilterModel[] = [
  {
    selectedValues: ["Pending"],
    label: "Status",
    name: "phase",
  } as IEnumBasedFilterModel,
];

export const llmWorkload: Workload = {
  category: EWorkloadCategory.Inference,
  tenantId: 1001,
  runningPods: 0,
  phaseUpdatedAt: "2024-09-26T12:55:09.142896Z",
  k8sPhaseUpdatedAt: "2024-09-26T12:45:07.152105Z",
  updatedAt: "2024-09-30T07:06:06.634875Z",
  source: "Control-plane",
  deletedAt: null,
  type: "Inference",
  name: "hugging-face",
  id: "7f7149a0-8882-455c-9b4c-32a7f7608a47",
  priority: 125,
  priorityClassName: "inference",
  submittedBy: "test@run.ai",
  clusterId: "74b5ddb9-0538-46ad-b343-426f22a99037",
  projectName: "p100",
  projectId: "4500001",
  departmentName: "default",
  departmentId: "4500000",
  namespace: "runai-p100",
  createdAt: "2024-09-26T12:45:08Z",
  workloadRequestedResources: {
    gpuRequestType: "memory",
    gpuMemory: {
      limit: "16G",
      request: "16G",
    },
    cpu: {
      request: 1,
    },
    cpuMemory: {
      request: "16G",
    },
  },
  podsRequestedResources: {},
  allocatedResources: {},
  actionsSupport: {
    delete: true,
  },
  phase: "Failed",
  conditions: [],
  phaseMessage: "Failed...",
  k8sPhase: "",
  requestedPods: {
    min: 1,
    max: 1,
  },
  requestedNodePools: ["default"],
  images: ["runai.jfrog.io/core-llm/runai-vllm:v0.5.5-0.5.0"],
  urls: ["http://hugging-face.runai-p100.svc.cluster.local"],
  environments: [
    {
      name: "llm-server",
      id: "a50ef834-4008-4a3a-9ac8-bfc9c177f5e4",
    },
  ],
  preemptible: false,
  fullySaasWorkload: true,
  environmentVariables: {
    RUNAI_JOB_NAME: "${RUNAI_JOB_NAME}",
    RUNAI_MODEL: "microsoft/DialoGPT-small",
    RUNAI_PROJECT: "${RUNAI_PROJECT}",
  },
};

export const workspaceWorkload: Workload = {
  category: EWorkloadCategory.Workspace,
  tenantId: 1001,
  runningPods: 0,
  phaseUpdatedAt: "2024-09-30T07:14:35.51446Z",
  k8sPhaseUpdatedAt: "2024-09-30T07:14:17.998457Z",
  updatedAt: "2024-09-30T07:14:17.998378Z",
  source: "Control-plane",
  deletedAt: null,
  type: "Workspace",
  name: "w102",
  id: "fa8d3bbd-7d48-4051-a0b4-581ece620097",
  priority: null,
  priorityClassName: "",
  submittedBy: "test@run.ai",
  clusterId: "74b5ddb9-0538-46ad-b343-426f22a99037",
  projectName: "p100",
  projectId: "4500001",
  departmentName: "default",
  departmentId: "4500000",
  namespace: "",
  createdAt: "2024-09-30T07:14:17.99Z",
  workloadRequestedResources: {
    cpu: {
      request: 0.1,
    },
    cpuMemory: {
      request: "100M",
    },
  },
  podsRequestedResources: {},
  allocatedResources: {},
  actionsSupport: {
    delete: true,
    suspend: true,
  },
  phase: "Failed",
  conditions: [],
  phaseMessage:
    "1 error occurred:\n\t* admission webhook \"workload-controller.runai.svc\" denied the request: Nodepool 'default' does not exist, please create it first\n\n",
  k8sPhase: "",
  requestedPods: {
    number: 1,
  },
  requestedNodePools: ["default"],
  images: ["tensorflow/tensorflow:latest"],
  environments: [
    {
      connections: [
        {
          name: "TensorBoard",
          toolType: "tensorboard",
          connectionType: "ExternalUrl",
          containerPort: 6006,
        },
      ],
      name: "tensorboard",
      id: "5971d3e8-8411-44a1-adfa-e5494f92c5c3",
    },
  ],
  fullySaasWorkload: true,
};

export const distributedWorkload: Workload = {
  category: EWorkloadCategory.Training,
  tenantId: 1001,
  runningPods: 0,
  phaseUpdatedAt: "2024-09-30T08:58:25.522629Z",
  k8sPhaseUpdatedAt: "2024-09-30T08:58:15.019747Z",
  updatedAt: "2024-09-30T08:58:15.019691Z",
  source: "Control-plane",
  deletedAt: null,
  type: "Distributed",
  name: "py-workload",
  id: "e7df372d-b973-4ad7-8a84-d8ac701b7134",
  priority: null,
  priorityClassName: "",
  submittedBy: "test@run.ai",
  clusterId: "74b5ddb9-0538-46ad-b343-426f22a99037",
  projectName: "p100",
  projectId: "4500001",
  departmentName: "default",
  departmentId: "4500000",
  namespace: "",
  createdAt: "2024-09-30T08:58:15.005Z",
  workloadRequestedResources: {},
  podsRequestedResources: {},
  allocatedResources: {},
  actionsSupport: {
    delete: true,
    suspend: true,
  },
  phase: "Failed",
  conditions: [],
  phaseMessage:
    "1 error occurred:\n\t* admission webhook \"workload-controller.runai.svc\" denied the request: Nodepool 'default' does not exist, please create it first\n\n",
  k8sPhase: "",
  requestedPods: {
    number: 1,
  },
  requestedNodePools: ["default"],
  images: ["image.com"],
  environments: [
    {
      name: "dis-env",
      id: "4d33ed43-8777-4984-8fba-c53702991403",
    },
  ],
  distributedFramework: "PyTorch",
  fullySaasWorkload: true,
};

export const nimWorkloadWithNimImage: Workload = {
  category: EWorkloadCategory.Inference,
  tenantId: 1001,
  runningPods: 1,
  phaseUpdatedAt: "2024-11-14T11:15:36.798997Z",
  k8sPhaseUpdatedAt: "2024-11-14T11:15:36.798997Z",
  updatedAt: "2024-11-14T11:15:36.798764Z",
  source: "Control-plane",
  deletedAt: null,
  type: "Inference",
  name: "nim-workload",
  id: "nim-workload-id",
  priorityClassName: "inference",
  submittedBy: "test@run.ai",
  clusterId: "fd9a83bb-38b3-4793-87d5-42b846b852f3",
  projectName: "p100",
  projectId: "4500001",
  departmentName: "default",
  departmentId: "4500000",
  namespace: "runai-p100",
  createdAt: "2024-11-14T11:15:07Z",
  workloadRequestedResources: {
    cpu: {
      request: 0.1,
    },
    cpuMemory: {
      request: "100M",
    },
  },
  podsRequestedResources: {},
  allocatedResources: {},
  actionsSupport: {},
  phase: "Running",
  conditions: [],
  k8sPhase: "Ready",
  requestedPods: {
    min: 1,
    max: 1,
  },
  requestedNodePools: ["default"],
  images: ["nvcr.io/nim/mistralai/mistral-7b-instruct:latest"],
  urls: ["https://mdev218.runailabs.com/p100/nim-workload"],
  environments: [],
  preemptible: true,
  priority: null,
};

export const nimWorkloadWithNimWorkloadFlag: Workload = {
  category: EWorkloadCategory.Inference,
  tenantId: 1001,
  runningPods: 1,
  phaseUpdatedAt: "2024-11-14T11:15:36.798997Z",
  k8sPhaseUpdatedAt: "2024-11-14T11:15:36.798997Z",
  updatedAt: "2024-11-14T11:15:36.798764Z",
  source: "Control-plane",
  deletedAt: null,
  type: "Inference",
  name: "nim-workload-flag",
  id: "nim-workload-flag-id",
  priorityClassName: "inference",
  submittedBy: "test@run.ai",
  clusterId: "fd9a83bb-38b3-4793-87d5-42b846b852f3",
  projectName: "p100",
  projectId: "4500001",
  departmentName: "default",
  departmentId: "4500000",
  namespace: "runai-p100",
  createdAt: "2024-11-14T11:15:07Z",
  workloadRequestedResources: {
    cpu: {
      request: 0.1,
    },
    cpuMemory: {
      request: "100M",
    },
  },
  podsRequestedResources: {},
  allocatedResources: {},
  actionsSupport: {},
  phase: "Running",
  conditions: [],
  k8sPhase: "Ready",
  requestedPods: {
    min: 1,
    max: 1,
  },
  requestedNodePools: ["default"],
  images: ["some-other-image:latest"],
  urls: ["https://mdev218.runailabs.com/p100/nim-workload-flag"],
  environments: [],
  preemptible: true,
  priority: null,
  additionalFields: {
    nimWorkload: "true",
  },
};

export const nonNimInferenceWorkload: Workload = {
  tenantId: 1001,
  runningPods: 1,
  phaseUpdatedAt: "2024-11-14T11:15:36.798997Z",
  k8sPhaseUpdatedAt: "2024-11-14T11:15:36.798997Z",
  updatedAt: "2024-11-14T11:15:36.798764Z",
  source: "Control-plane",
  deletedAt: null,
  type: "Inference",
  name: "regular-inference",
  id: "regular-inference-id",
  priorityClassName: "inference",
  submittedBy: "test@run.ai",
  clusterId: "fd9a83bb-38b3-4793-87d5-42b846b852f3",
  projectName: "p100",
  projectId: "4500001",
  departmentName: "default",
  departmentId: "4500000",
  namespace: "runai-p100",
  createdAt: "2024-11-14T11:15:07Z",
  workloadRequestedResources: {
    cpu: {
      request: 0.1,
    },
    cpuMemory: {
      request: "100M",
    },
  },
  podsRequestedResources: {},
  allocatedResources: {},
  actionsSupport: {},
  phase: "Running",
  conditions: [],
  k8sPhase: "Ready",
  requestedPods: {
    min: 1,
    max: 1,
  },
  requestedNodePools: ["default"],
  images: ["some-regular-image:latest"],
  urls: ["https://mdev218.runailabs.com/p100/regular-inference"],
  environments: [],
  preemptible: true,
  priority: null,
  category: EWorkloadCategory.Training,
};

export const nonInferenceWorkload: Workload = {
  tenantId: 1001,
  runningPods: 1,
  phaseUpdatedAt: "2024-11-14T11:15:36.798997Z",
  k8sPhaseUpdatedAt: "2024-11-14T11:15:36.798997Z",
  updatedAt: "2024-11-14T11:15:36.798764Z",
  source: "Control-plane",
  deletedAt: null,
  type: "Training",
  name: "training-workload",
  id: "training-workload-id",
  priorityClassName: "training",
  submittedBy: "test@run.ai",
  clusterId: "fd9a83bb-38b3-4793-87d5-42b846b852f3",
  projectName: "p100",
  projectId: "4500001",
  departmentName: "default",
  departmentId: "4500000",
  namespace: "runai-p100",
  createdAt: "2024-11-14T11:15:07Z",
  workloadRequestedResources: {
    cpu: {
      request: 0.1,
    },
    cpuMemory: {
      request: "100M",
    },
  },
  podsRequestedResources: {},
  allocatedResources: {},
  actionsSupport: {},
  phase: "Running",
  conditions: [],
  k8sPhase: "Ready",
  requestedPods: {
    min: 1,
    max: 1,
  },
  requestedNodePools: ["default"],
  images: ["nvcr.io/nim/some-model:latest"],
  urls: ["https://mdev218.runailabs.com/p100/training-workload"],
  environments: [],
  preemptible: true,
  priority: null,
  category: EWorkloadCategory.Workspace,
};

export const connectionsMocks: Record<string, Workload> = {
  workspace: {
    category: EWorkloadCategory.Workspace,
    tenantId: 1001,
    runningPods: 0,
    phaseUpdatedAt: "2024-11-14T12:25:10.540604Z",
    k8sPhaseUpdatedAt: "2024-11-14T12:25:05.915238Z",
    updatedAt: "2024-11-14T12:25:10.540351Z",
    source: "Control-plane",
    deletedAt: null,
    type: "Workspace",
    name: "w100",
    id: "95416f88-f011-4073-ae55-bfa684f64f24",
    priority: 100,
    priorityClassName: "build",
    submittedBy: "test@run.ai",
    clusterId: "d2e911f5-8a65-42b9-8e7b-29f8c1c20cb1",
    projectName: "p100",
    projectId: "4500030",
    departmentName: "default",
    departmentId: "4500000",
    namespace: "runai-p100",
    createdAt: "2024-11-14T12:25:10Z",
    workloadRequestedResources: {
      gpuRequestType: "portion",
      gpu: {
        request: 0,
      },
      cpu: {
        request: 0,
      },
      cpuMemory: {
        request: "0M",
      },
    },
    podsRequestedResources: {
      cpuMemory: {
        request: "0",
      },
    },
    allocatedResources: {},
    actionsSupport: {
      delete: true,
      suspend: true,
    },
    phase: "Pending",
    conditions: [
      {
        type: "Ready",
        status: "False",
        message: "job is not running",
        lastTransitionTime: "2024-11-14T12:25:10Z",
      },
      {
        type: "LastReconcileSucceeded",
        status: "True",
        lastTransitionTime: "2024-11-14T12:25:10Z",
      },
      {
        type: "Pending",
        status: "True",
        lastTransitionTime: "2024-11-14T12:25:10Z",
      },
    ],
    k8sPhase: "",
    requestedPods: {
      number: 1,
    },
    requestedNodePools: ["default"],
    images: ["tensorflow/tensorflow:latest"],
    urls: ["https://mdev2.runailabs.com/p100/w100"],
    environments: [
      {
        connections: [
          {
            name: "TensorBoard",
            toolType: "tensorboard",
            connectionType: "ExternalUrl",
            url: "https://mdev2.runailabs.com/p100/w100",
            containerPort: 6006,
          },
        ],
        name: "tensorboard",
        id: "3c056f5b-cb13-4be6-8fe4-74f9a1d7ef16",
      },
    ],
    preemptible: false,
    fullySaasWorkload: true,
    environmentVariables: {
      RUNAI_JOB_NAME: "${RUNAI_JOB_NAME}",
      RUNAI_PROJECT: "${RUNAI_PROJECT}",
    },
    command: "tensorboard",
    arguments: "--logdir ./ --host 0.0.0.0 --path_prefix /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}",
    totalPendingTimeSeconds: 1,
  },
  inference: {
    category: EWorkloadCategory.Inference,
    tenantId: 1001,
    runningPods: 1,
    phaseUpdatedAt: "2024-11-14T11:15:36.798997Z",
    k8sPhaseUpdatedAt: "2024-11-14T11:15:36.798997Z",
    updatedAt: "2024-11-14T11:15:36.798764Z",
    source: "Control-plane",
    deletedAt: null,
    type: "Inference",
    name: "i100",
    id: "6ff7248c-19be-4259-8568-61d2a6b8604e",
    priorityClassName: "train",
    submittedBy: "test@run.ai",
    clusterId: "fd9a83bb-38b3-4793-87d5-42b846b852f3",
    projectName: "p100",
    projectId: "4500001",
    departmentName: "default",
    departmentId: "4500000",
    namespace: "runai-p100",
    createdAt: "2024-11-14T11:15:07Z",
    workloadRequestedResources: {
      cpu: {
        request: 0.1,
      },
      cpuMemory: {
        request: "100M",
      },
    },
    podsRequestedResources: {
      cpu: {
        request: 0.13,
      },
      cpuMemory: {
        request: "100M",
      },
    },
    allocatedResources: {
      cpu: 0.13,
      cpuMemory: "100M",
    },
    actionsSupport: {},
    phase: "Running",
    conditions: [
      {
        type: "ConfigurationsReady",
        status: "True",
        lastTransitionTime: "2024-11-14T11:15:36Z",
      },
      {
        type: "Ready",
        status: "True",
        lastTransitionTime: "2024-11-14T11:15:36Z",
      },
      {
        type: "RoutesReady",
        status: "True",
        lastTransitionTime: "2024-11-14T11:15:36Z",
      },
    ],
    k8sPhase: "Ready",
    requestedPods: {
      min: 1,
      max: 1,
    },
    requestedNodePools: ["default"],
    currentNodePools: ["default"],
    urls: ["https://mdev218.runailabs.com/p100/i100", "http://i100.runai-p100.svc.cluster.local"],
    environments: [
      {
        connections: [
          {
            name: "ChatbotUI",
            toolType: "chatbot-ui",
            connectionType: "ExternalUrl",
            url: "https://mdev218.runailabs.com/p100/i100",
          },
        ],
        name: "chatbot-ui",
        id: "715a51c3-93ce-4838-a6c6-4d248e6b9ba0",
      },
    ],
    preemptible: true,
    priority: null,
  },
};
