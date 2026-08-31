/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TrainingPolicyV2 } from "@/swagger-models/policy-service-client";

import { type IProjectResources } from "@/models/project.model";
import {
  type ComputeAsset,
  type DatasourceListResponseEntry,
  type EnvironmentAsset,
  type Training,
  type WorkloadCreationRequest,
  type WorkloadTemplate,
  GpuRequestType,
  type TrainingV2,
} from "@/swagger-models/assets-service-client";
import type { Project, ProjectMetadata } from "@/swagger-models/org-unit-service-client";

const trainingAssetsEditMockedData = {
  loadedEnvironments: [
    {
      meta: {
        name: "jupyter-lab",
        scope: "tenant",
        id: "ad947a50-55c3-4498-892c-2aec5ad1cefe",
        kind: "environment",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-03-07T08:34:32.966669Z",
        updatedBy: "runai",
        updatedAt: "2023-03-07T08:34:32.966669Z",
      },
      spec: {
        command: "start-notebook.sh",
        args: "--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''",
        environmentVariables: [],
        image: "jupyter/scipy-notebook",
        imagePullPolicy: "IfNotPresent",
        connections: [
          {
            name: "Jupyter",
            internalToolInfo: {
              toolType: "jupyter-notebook",
              connectionType: "ExternalUrl",
              containerPort: 8888,
              externalUrlInfo: {},
            },
          },
        ],
        uidGidSource: "fromTheImage",
        capabilities: [],
        sharedMemory: false,
      },
      compliance: {
        compliance: true,
      },
    },
    {
      meta: {
        name: "jupyter-tensorboard",
        scope: "tenant",
        id: "f1fca23e-56f4-4e88-b2be-2c7df30a28b2",
        kind: "environment",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-03-07T08:34:32.967399Z",
        updatedBy: "runai",
        updatedAt: "2023-03-07T08:34:32.967399Z",
      },
      spec: {
        environmentVariables: [
          {
            name: "LOG_DIR",
            value: "./",
          },
        ],
        image: "gcr.io/run-ai-demo/jupyter-tensorboard",
        imagePullPolicy: "IfNotPresent",
        connections: [
          {
            name: "Jupyter",
            internalToolInfo: {
              toolType: "jupyter-notebook",
              connectionType: "ExternalUrl",
              containerPort: 8888,
              externalUrlInfo: {},
            },
          },
          {
            name: "TensorBoard",
            internalToolInfo: {
              toolType: "tensorboard",
              connectionType: "ExternalUrl",
              containerPort: 6006,
              externalUrlInfo: {},
            },
          },
        ],
        uidGidSource: "fromTheImage",
        capabilities: [],
        sharedMemory: false,
      },
      compliance: {
        compliance: true,
      },
    },
    {
      meta: {
        name: "tensorboard",
        scope: "tenant",
        id: "0353c1fd-563d-4b4d-bbba-8bb54735e816",
        kind: "environment",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-03-07T08:34:32.965269Z",
        updatedBy: "runai",
        updatedAt: "2023-03-07T08:34:32.965269Z",
      },
      spec: {
        command: "tensorboard",
        args: "--logdir ./ --host 0.0.0.0 --path_prefix /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}",
        environmentVariables: [],
        image: "tensorflow/tensorflow:latest",
        imagePullPolicy: "IfNotPresent",
        connections: [
          {
            name: "TensorBoard",
            internalToolInfo: {
              toolType: "tensorboard",
              connectionType: "ExternalUrl",
              containerPort: 6006,
              externalUrlInfo: {},
            },
          },
        ],
        uidGidSource: "fromTheImage",
        capabilities: [],
      },
      compliance: {
        compliance: true,
      },
    },
  ] as EnvironmentAsset[],
  loadedDataSources: [
    {
      meta: {
        name: "test-hp",
        scope: "tenant",
        id: "a4e1499f-559e-45f4-b81e-46cb809db13f",
        kind: "hostPath",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-03-08T08:56:09.616193Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-03-08T08:56:09.616193Z",
      },
      spec: {
        hostPath: {
          path: "/test-path",
          readOnly: false,
          mountPath: "/test-container",
        },
      },
      compliance: {
        compliance: true,
      },
    },
    {
      meta: {
        name: "test-nfs",
        scope: "tenant",
        id: "d2792d6a-bd20-4e76-a60e-e2b2b347d468",
        kind: "nfs",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-03-08T08:55:41.342147Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-03-08T08:55:41.342147Z",
      },
      spec: {
        nfs: {
          path: "/test-path",
          readOnly: false,
          server: "test-server",
          mountPath: "/test-container-path",
        },
      },
      compliance: {
        compliance: true,
      },
    },
  ] as DatasourceListResponseEntry[],
  loadedComputeResources: [
    {
      meta: {
        name: "cpu-only",
        scope: "tenant",
        id: "ed644ac3-1a72-4419-9dc5-e424950d02ef",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-03-07T08:34:32.967182Z",
        updatedBy: "runai",
        updatedAt: "2023-03-07T08:34:32.967182Z",
      },
      spec: {
        cpuCoreRequest: 0.1,
        cpuMemoryRequest: "100M",
      },
      compliance: {
        compliance: true,
      },
    },
    {
      meta: {
        name: "hopa",
        scope: "project",
        clusterId: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
        projectId: 1,
        id: "8734ec10-9a51-4d7d-81b0-c8a110fe5386",
        kind: "compute",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-03-07T12:51:19.511123Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-03-07T12:51:19.511123Z",
        projectName: "rotem-test-1",
      },
      spec: {
        cpuCoreRequest: 0,
        cpuMemoryRequest: "0M",
        largeShmRequest: false,
      },
      compliance: {
        compliance: true,
      },
    },
    {
      meta: {
        name: "popa",
        scope: "project",
        clusterId: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
        projectId: 1,
        id: "875f1ee3-abdb-4ef6-897a-e816fe15e5a3",
        kind: "compute",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-03-07T12:51:28.24909Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-03-07T12:51:28.24909Z",
        projectName: "rotem-test-1",
      },
      spec: {
        cpuCoreRequest: 0,
        cpuMemoryRequest: "0M",
        largeShmRequest: false,
      },
      compliance: {
        compliance: true,
      },
    },
    {
      meta: {
        name: "small-fraction",
        scope: "tenant",
        id: "5ac6fc47-45d8-4b06-a82d-7c5324b49851",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-03-07T08:34:32.967989Z",
        updatedBy: "runai",
        updatedAt: "2023-03-07T08:34:32.967989Z",
      },
      spec: {
        gpuRequestType: GpuRequestType.Portion,
        gpuPortion: 0.1,
        cpuCoreRequest: 0.1,
        cpuMemoryRequest: "100M",
      },
      compliance: {
        compliance: true,
      },
    },
    {
      meta: {
        name: "half-gpu",
        scope: "tenant",
        id: "751115c8-235f-4b35-af99-3d367a51baaf",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-03-07T08:34:32.967604Z",
        updatedBy: "runai",
        updatedAt: "2023-03-07T08:34:32.967604Z",
      },
      spec: {
        gpuRequestType: GpuRequestType.Portion,
        gpuPortion: 0.5,
        cpuCoreRequest: 0.1,
        cpuMemoryRequest: "100M",
      },
      compliance: {
        compliance: true,
      },
    },
    {
      meta: {
        name: "one-gpu",
        scope: "tenant",
        id: "1d47ae37-0ae2-4d13-a879-716472108ed6",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-03-07T08:34:32.966949Z",
        updatedBy: "runai",
        updatedAt: "2023-03-07T08:34:32.966949Z",
      },
      spec: {
        gpuRequestType: GpuRequestType.Portion,
        gpuPortion: 1,
        cpuCoreRequest: 0.1,
        cpuMemoryRequest: "100M",
      },
      compliance: {
        compliance: true,
      },
    },
    {
      meta: {
        name: "two-gpus",
        scope: "tenant",
        id: "92e209ea-9a2d-40db-ade3-408e4eb8c9aa",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-03-07T08:34:32.967809Z",
        updatedBy: "runai",
        updatedAt: "2023-03-07T08:34:32.967809Z",
      },
      spec: {
        gpuRequestType: GpuRequestType.Portion,
        gpuPortion: 2,
        cpuCoreRequest: 0.2,
        cpuMemoryRequest: "200M",
      },
      compliance: {
        compliance: true,
      },
    },
  ] as ComputeAsset[],
  loadedProjects: [
    {
      phase: "Ready",
      clusterUuid: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
      createdAt: "2023-03-07T08:34:28.662Z",
      deservedGpus: 0,
      maxAllowedGpus: -1,
      id: 1,
      name: "project-test-1",
      tenantId: 1,
      departmentId: 1,
      interactiveJobTimeLimitSecs: null,
      interactiveJobMaxIdleDurationSecs: null,
      interactivePreemptibleJobMaxIdleDurationSecs: null,
      trainingJobMaxIdleDurationSecs: null,
      swapEnabled: false,
      gpuOverQuotaWeight: 0,
      nodeAffinity: {
        train: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
        interactive: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
      },
      departmentName: "default",
      status: {
        namespace: "runai-project-test-1",
      },
      permissions: {
        users: ["48dfef96-11e7-4660-8f1d-9d05e7a4a994"],
        groups: [],
        applications: [],
      },
      resources: {
        gpu: {
          deserved: 0,
          overQuotaWeight: 0,
          maxAllowed: -1,
        },
        cpu: {
          deserved: null,
          overQuotaWeight: null,
          maxAllowed: -1,
        },
        memory: {
          deserved: null,
          overQuotaWeight: null,
          maxAllowed: -1,
        },
      },
      nodePoolsResources: [
        {
          nodePool: {
            id: 1,
            name: "default",
          },
          gpu: {
            deserved: 0,
            overQuotaWeight: 0,
            maxAllowed: -1,
          },
          cpu: {
            deserved: null,
            overQuotaWeight: null,
            maxAllowed: -1,
          },
          memory: {
            deserved: null,
            overQuotaWeight: null,
            maxAllowed: -1,
          },
        },
      ],
    },
    {
      phase: "Ready",
      clusterUuid: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
      createdAt: "2023-03-07T08:34:28.662Z",
      deservedGpus: 0,
      maxAllowedGpus: -1,
      id: 2,
      name: "project-test-2",
      tenantId: 1,
      departmentId: 1,
      interactiveJobTimeLimitSecs: null,
      interactiveJobMaxIdleDurationSecs: null,
      interactivePreemptibleJobMaxIdleDurationSecs: null,
      trainingJobMaxIdleDurationSecs: null,
      swapEnabled: false,
      gpuOverQuotaWeight: 0,
      nodeAffinity: {
        train: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
        interactive: {
          affinityType: "only_selected",
          selectedTypes: [
            {
              id: 1,
              name: "test-aff-1",
            },
            {
              id: 2,
              name: "test-aff-2",
            },
            {
              id: 3,
              name: "test-aff-3",
            },
          ],
        },
      },
      departmentName: "default",
      status: {
        namespace: "runai-project-test-2",
      },
      permissions: {
        users: ["48dfef96-11e7-4660-8f1d-9d05e7a4a994"],
        groups: [],
      },
      resources: {
        gpu: {
          deserved: 0,
          overQuotaWeight: 0,
          maxAllowed: -1,
        },
        cpu: {
          deserved: null,
          overQuotaWeight: null,
          maxAllowed: -1,
        },
        memory: {
          deserved: null,
          overQuotaWeight: null,
          maxAllowed: -1,
        },
      },
      nodePoolsResources: [
        {
          id: 4,
          nodePool: {
            id: 1,
            name: "default",
          },
          gpu: {
            deserved: 0,
            overQuotaWeight: 0,
            maxAllowed: -1,
          },
          cpu: {
            deserved: null,
            overQuotaWeight: null,
            maxAllowed: -1,
          },
          memory: {
            deserved: null,
            overQuotaWeight: null,
            maxAllowed: -1,
          },
        },
      ],
    },
  ],
  loadedProjectResources: [
    {
      id: 1,
      tenantId: 1,
      clusterUuid: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
      name: "test-1",
      deservedGpus: 0,
      allocatedGpus: 0,
    },
    {
      id: 2,
      tenantId: 1,
      clusterUuid: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
      name: "test-2",
      deservedGpus: 0,
      allocatedGpus: 0,
    },
  ] as IProjectResources[],
  loadedTemplates: [
    {
      meta: {
        name: "test-template",
        scope: "project",
        clusterId: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
        projectId: 1,
        id: "d3481435-996e-4ca2-9e52-00f8d8203725",
        kind: "workload-template",
        tenantId: 1,
        createdBy: "test@run.ai",
        createdAt: "2023-03-15T15:29:13.340953Z",
        updatedBy: "test@run.ai",
        updatedAt: "2023-03-15T15:29:13.340953Z",
        projectName: "rotem-test-1",
      },
      spec: {
        assets: {
          environment: {
            id: "3c6fc0e5-effa-4a55-85f9-5adb6278ce50",
            name: "test-template-1",
            toolTypes: ["jupyter-notebook", "mlflow", "rstudio"],
          },
          compute: {
            id: "ed644ac3-1a72-4419-9dc5-e424950d02ef",
            name: "cpu-only",
          },
        },
        specificEnv: {
          command: "cmd-1",
          args: "arg-1",
          environmentVariables: [
            {
              name: "var-1",
              value: "12",
            },
          ],
          connections: [],
          runAsUid: 5,
          runAsGid: 3,
          supplementalGroups: "5,5,5",
          allowOverQuota: false,
        },
      },
    },
  ] as WorkloadTemplate[],
  loadedPolicy: {
    effective: {
      defaults: {
        environmentVariables: {
          instances: [
            {
              name: "var1",
              value: "key1",
            },
            {
              name: "var2",
              value: "key2",
            },
          ],
        },
      },
      rules: {
        environmentVariables: {
          instances: {
            canAdd: true,
            locked: ["var2"],
          },
        },
      },
    },
  } as TrainingPolicyV2,
  loadedProjectsMetasV2: [
    {
      id: "4500001",
      name: "proj-1",
      namespace: "runai-proj-1",
      phase: "Ready",
      clusterId: "62d19e09-38cd-455e-afb3-62716c8c5d44",
      createdAt: "2024-10-09T06:55:08.890662Z",
      allocated: {
        cpu: 300,
        memory: 300,
      },
      deserved: {
        gpu: 5,
      },
    },
    {
      id: "4500002",
      name: "np-proj",
      namespace: "runai-np-proj",
      phase: "Ready",
      clusterId: "62d19e09-38cd-455e-afb3-62716c8c5d44",
      createdAt: "2024-10-09T06:57:27.530839Z",
      allocated: {},
      deserved: {
        gpu: 6,
      },
    },
    {
      id: "4500003",
      name: "a-mess-of-a-project",
      namespace: "runai-a-mess-of-a-project",
      phase: "Ready",
      clusterId: "62d19e09-38cd-455e-afb3-62716c8c5d44",
      createdAt: "2024-10-09T12:00:07.255551Z",
      allocated: {},
      deserved: {},
    },
    {
      id: "4500004",
      name: "messy-2",
      namespace: "runai-messy-2",
      phase: "Ready",
      clusterId: "62d19e09-38cd-455e-afb3-62716c8c5d44",
      createdAt: "2024-10-09T12:19:11.21641Z",
      allocated: {},
      deserved: {},
    },
    {
      id: "4500008",
      name: "proj-affinity",
      namespace: "runai-proj-affinity",
      phase: "Ready",
      clusterId: "62d19e09-38cd-455e-afb3-62716c8c5d44",
      createdAt: "2024-10-10T10:07:25.689525Z",
      allocated: {},
      deserved: {
        gpu: 2,
      },
    },
  ] as ProjectMetadata[],
  loadedProjectsV2: [
    {
      nodePoolsResources: [
        {
          id: 4500000,
          nodePool: {
            id: 100000,
            name: "default",
          },
          gpu: {
            deserved: 0,
            maxAllowed: -1,
            overQuotaWeight: 0,
          },
          cpu: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
          memory: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
        },
        {
          id: 4500000,
          nodePool: {
            id: 100001,
            name: "and-anotha",
          },
          gpu: {
            deserved: 0,
            maxAllowed: -1,
            overQuotaWeight: 0,
          },
          cpu: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
          memory: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
        },
      ],
      deservedGpus: 0,
      maxAllowedGpus: -1,
      gpuOverQuotaWeight: 0,
      defaultNodePools: ["default"],
      nodeAffinity: {
        train: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
        interactive: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
      },
      permissions: {
        users: [],
        groups: [],
      },
      resources: {
        gpu: {
          deserved: 0,
          maxAllowed: -1,
          overQuotaWeight: 0,
        },
        cpu: {
          deserved: null,
          maxAllowed: null,
          overQuotaWeight: null,
        },
        memory: {
          deserved: null,
          maxAllowed: null,
          overQuotaWeight: null,
        },
      },
      name: "messy-2",
      namespace: "",
      id: "4500004",
      departmentId: 4500000,
      tenantId: 1001,
      clusterUuid: "62d19e09-38cd-455e-afb3-62716c8c5d44",
      departmentName: "default",
      createdAt: "2024-10-09T12:19:11.21641Z",
      status: {
        namespace: "runai-messy-2",
        phase: "Ready",
        conditions: [
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:19:12Z",
            status: "True",
            type: "NamespaceReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:19:12Z",
            status: "True",
            type: "RoleBindingsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:19:12Z",
            status: "True",
            type: "ServiceMonitorReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:19:12Z",
            status: "True",
            type: "SecretsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:19:12Z",
            status: "True",
            type: "QueuesReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:19:12Z",
            status: "True",
            type: "PvcsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:19:12Z",
            status: "True",
            type: "ConfigMapsReady",
          },
        ],
        quotaStatuses: [
          {
            nodePoolName: "and-anotha",
            allocated: {},
            allocatedNonPreemptible: {},
            requested: {},
          },
          {
            nodePoolName: "default",
            allocated: {},
            allocatedNonPreemptible: {},
            requested: {},
          },
        ],
      },
      phase: "Ready",
    },
    {
      nodePoolsResources: [
        {
          id: 4500000,
          nodePool: {
            id: 100000,
            name: "default",
          },
          gpu: {
            deserved: 0,
            maxAllowed: -1,
            overQuotaWeight: 0,
          },
          cpu: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
          memory: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
        },
        {
          id: 4500000,
          nodePool: {
            id: 100001,
            name: "and-anotha",
          },
          gpu: {
            deserved: 0,
            maxAllowed: -1,
            overQuotaWeight: 0,
          },
          cpu: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
          memory: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
        },
      ],
      deservedGpus: 0,
      maxAllowedGpus: -1,
      gpuOverQuotaWeight: 0,
      defaultNodePools: ["default"],
      nodeAffinity: {
        train: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
        interactive: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
      },
      permissions: {
        users: [],
        groups: [],
      },
      resources: {
        gpu: {
          deserved: 0,
          maxAllowed: -1,
          overQuotaWeight: 0,
        },
        cpu: {
          deserved: null,
          maxAllowed: null,
          overQuotaWeight: null,
        },
        memory: {
          deserved: null,
          maxAllowed: null,
          overQuotaWeight: null,
        },
      },
      name: "a-mess-of-a-project",
      namespace: "",
      id: "4500003",
      departmentId: 4500000,
      tenantId: 1001,
      clusterUuid: "62d19e09-38cd-455e-afb3-62716c8c5d44",
      departmentName: "default",
      createdAt: "2024-10-09T12:00:07.255551Z",
      status: {
        namespace: "runai-a-mess-of-a-project",
        phase: "Ready",
        conditions: [
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:00:12Z",
            status: "True",
            type: "NamespaceReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:00:12Z",
            status: "True",
            type: "RoleBindingsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:00:12Z",
            status: "True",
            type: "ServiceMonitorReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:00:12Z",
            status: "True",
            type: "SecretsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:00:12Z",
            status: "True",
            type: "PvcsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:00:12Z",
            status: "True",
            type: "ConfigMapsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T12:00:12Z",
            status: "True",
            type: "QueuesReady",
          },
        ],
        quotaStatuses: [
          {
            nodePoolName: "default",
            allocated: {},
            allocatedNonPreemptible: {},
            requested: {},
          },
          {
            nodePoolName: "and-anotha",
            allocated: {},
            allocatedNonPreemptible: {},
            requested: {},
          },
        ],
      },
      phase: "Ready",
    },
    {
      nodePoolsResources: [
        {
          id: 4500000,
          nodePool: {
            id: 100000,
            name: "default",
          },
          gpu: {
            deserved: 4,
            maxAllowed: -1,
            overQuotaWeight: 4,
          },
          cpu: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
          memory: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
        },
        {
          id: 4500000,
          nodePool: {
            id: 100001,
            name: "and-anotha",
          },
          gpu: {
            deserved: 2,
            maxAllowed: -1,
            overQuotaWeight: 2,
          },
          cpu: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
          memory: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
        },
      ],
      deservedGpus: 6,
      maxAllowedGpus: -1,
      gpuOverQuotaWeight: 6,
      defaultNodePools: ["default"],
      nodeAffinity: {
        train: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
        interactive: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
      },
      permissions: {
        users: [],
        groups: [],
      },
      resources: {
        gpu: {
          deserved: 6,
          maxAllowed: -1,
          overQuotaWeight: 6,
        },
        cpu: {
          deserved: null,
          maxAllowed: null,
          overQuotaWeight: null,
        },
        memory: {
          deserved: null,
          maxAllowed: null,
          overQuotaWeight: null,
        },
      },
      name: "np-proj",
      namespace: "",
      id: "4500002",
      departmentId: 4500000,
      tenantId: 1001,
      clusterUuid: "62d19e09-38cd-455e-afb3-62716c8c5d44",
      departmentName: "default",
      createdAt: "2024-10-09T06:57:27.530839Z",
      status: {
        namespace: "runai-np-proj",
        phase: "Ready",
        conditions: [
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:57:32Z",
            status: "True",
            type: "NamespaceReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:57:32Z",
            status: "True",
            type: "RoleBindingsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:57:32Z",
            status: "True",
            type: "ServiceMonitorReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:57:32Z",
            status: "True",
            type: "SecretsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:57:32Z",
            status: "True",
            type: "PvcsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:57:32Z",
            status: "True",
            type: "ConfigMapsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:57:32Z",
            status: "True",
            type: "QueuesReady",
          },
        ],
        quotaStatuses: [
          {
            nodePoolName: "and-anotha",
            allocated: {},
            allocatedNonPreemptible: {},
            requested: {},
          },
          {
            nodePoolName: "default",
            allocated: {},
            allocatedNonPreemptible: {},
            requested: {},
          },
        ],
      },
      phase: "Ready",
    },
    {
      nodePoolsResources: [
        {
          id: 4500000,
          nodePool: {
            id: 100000,
            name: "default",
          },
          gpu: {
            deserved: 5,
            maxAllowed: -1,
            overQuotaWeight: 5,
          },
          cpu: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
          memory: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
        },
        {
          id: 4500000,
          nodePool: {
            id: 100001,
            name: "and-anotha",
          },
          gpu: {
            deserved: 0,
            maxAllowed: -1,
            overQuotaWeight: 0,
          },
          cpu: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
          memory: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
        },
      ],
      deservedGpus: 5,
      maxAllowedGpus: -1,
      gpuOverQuotaWeight: 5,
      defaultNodePools: ["default"],
      nodeAffinity: {
        train: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
        interactive: {
          affinityType: "no_limit",
          selectedTypes: [],
        },
      },
      permissions: {
        users: [],
        groups: [],
      },
      resources: {
        gpu: {
          deserved: 5,
          maxAllowed: -1,
          overQuotaWeight: 5,
        },
        cpu: {
          deserved: null,
          maxAllowed: null,
          overQuotaWeight: null,
        },
        memory: {
          deserved: null,
          maxAllowed: null,
          overQuotaWeight: null,
        },
      },
      name: "proj-1",
      namespace: "",
      id: "4500001",
      departmentId: 4500000,
      tenantId: 1001,
      clusterUuid: "62d19e09-38cd-455e-afb3-62716c8c5d44",
      departmentName: "default",
      createdAt: "2024-10-09T06:55:08.890662Z",
      status: {
        namespace: "runai-proj-1",
        phase: "Ready",
        conditions: [
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:55:12Z",
            status: "True",
            type: "NamespaceReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:55:12Z",
            status: "True",
            type: "RoleBindingsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:55:13Z",
            status: "True",
            type: "ServiceMonitorReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:55:13Z",
            status: "True",
            type: "SecretsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:55:12Z",
            status: "True",
            type: "QueuesReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:55:13Z",
            status: "True",
            type: "PvcsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-09T06:55:13Z",
            status: "True",
            type: "ConfigMapsReady",
          },
        ],
        quotaStatuses: [
          {
            nodePoolName: "and-anotha",
            allocated: {},
            allocatedNonPreemptible: {},
            requested: {},
          },
          {
            nodePoolName: "default",
            allocated: {
              cpu: 300,
              memory: 286.102,
            },
            allocatedNonPreemptible: {
              cpu: 200,
              memory: 190.735,
            },
            requested: {
              cpu: 300,
              memory: 286.102,
            },
          },
        ],
      },
      phase: "Ready",
    },
    {
      nodePoolsResources: [
        {
          id: 4500000,
          nodePool: {
            id: 100000,
            name: "default",
          },
          gpu: {
            deserved: 1,
            maxAllowed: -1,
            overQuotaWeight: 1,
          },
          cpu: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
          memory: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
        },
        {
          id: 4500000,
          nodePool: {
            id: 100001,
            name: "and-anotha",
          },
          gpu: {
            deserved: 1,
            maxAllowed: -1,
            overQuotaWeight: 1,
          },
          cpu: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
          memory: {
            deserved: null,
            maxAllowed: null,
            overQuotaWeight: null,
          },
        },
      ],
      deservedGpus: 2,
      maxAllowedGpus: -1,
      gpuOverQuotaWeight: 2,
      defaultNodePools: ["default"],
      nodeAffinity: {
        train: {
          affinityType: "only_selected",
          selectedTypes: [
            {
              id: 4500005,
              name: "some-affinity",
            },
          ],
        },
        interactive: {
          affinityType: "only_selected",
          selectedTypes: [
            {
              id: 4500006,
              name: "aff-1",
            },
            {
              id: 4500007,
              name: "aff-2",
            },
          ],
        },
      },
      permissions: {
        users: [],
        groups: [],
      },
      resources: {
        gpu: {
          deserved: 2,
          maxAllowed: -1,
          overQuotaWeight: 2,
        },
        cpu: {
          deserved: null,
          maxAllowed: null,
          overQuotaWeight: null,
        },
        memory: {
          deserved: null,
          maxAllowed: null,
          overQuotaWeight: null,
        },
      },
      name: "proj-affinity",
      namespace: "",
      id: "4500008",
      departmentId: 4500000,
      tenantId: 1001,
      clusterUuid: "62d19e09-38cd-455e-afb3-62716c8c5d44",
      departmentName: "default",
      createdAt: "2024-10-10T10:07:25.689525Z",
      status: {
        namespace: "runai-proj-affinity",
        phase: "Ready",
        conditions: [
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-10T10:07:27Z",
            status: "True",
            type: "NamespaceReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-10T10:07:27Z",
            status: "True",
            type: "RoleBindingsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-10T10:07:27Z",
            status: "True",
            type: "ServiceMonitorReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-10T10:07:27Z",
            status: "True",
            type: "SecretsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-10T10:07:27Z",
            status: "True",
            type: "PvcsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-10T10:07:27Z",
            status: "True",
            type: "ConfigMapsReady",
          },
          {
            lastProbeTime: "",
            lastTransitionTime: "2024-10-10T10:07:27Z",
            status: "True",
            type: "QueuesReady",
          },
        ],
        quotaStatuses: [
          {
            nodePoolName: "default",
            allocated: {},
            allocatedNonPreemptible: {},
            requested: {},
          },
          {
            nodePoolName: "and-anotha",
            allocated: {},
            allocatedNonPreemptible: {},
            requested: {},
          },
        ],
      },
      phase: "Ready",
    },
  ] as unknown as Project[],
};

const validModel: WorkloadCreationRequest = {
  name: "valid-name",
  projectId: 1,
  clusterId: "",
  namespace: "",
  assets: {
    environment: "ad947a50-55c3-4498-892c-2aec5ad1cefe",
    compute: "ed644ac3-1a72-4419-9dc5-e424950d02ef",
    datasources: [],
  },
  specificEnv: {
    args: "--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''",
    command: "start-notebook.sh",
    environmentVariables: [],
    connections: [],
    allowOverQuota: false,
  },
};

const loadedTrainingById: Training = {
  meta: {
    id: "e82667b2-22cd-4204-92c9-1a76c09a5056",
    tenantId: 1,
    projectId: 1,
    clusterId: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
    name: "test-mock-ws",
    desiredState: "Submitted",
    currentState: "Submitted",
    createdBy: "test@run.ai",
    updatedBy: "test@run.ai",
  },
  spec: {
    assets: {
      environment: {
        id: "f1fca23e-56f4-4e88-b2be-2c7df30a28b2",
        name: "jupyter-tensorboard",
      },
      compute: {
        id: "8734ec10-9a51-4d7d-81b0-c8a110fe5386",
        name: "test-compute",
      },
      datasources: [
        {
          id: "a4e1499f-559e-45f4-b81e-46cb809db13f",
          name: "test-hp",
          kind: "hostPath",
        },
        {
          id: "d2792d6a-bd20-4e76-a60e-e2b2b347d468",
          name: "test-nfs",
          kind: "nfs",
        },
      ],
    },
    specificEnv: {
      args: "some-args",
      command: "some-command",
      runAsUid: 3,
      environmentVariables: [
        {
          name: "LOG_DIR",
          value: "./",
        },
      ],
      connections: [{ name: "custom-tool", externalUrl: "http://domain.com" }],
      allowOverQuota: false,
      annotations: [
        {
          name: "ann1",
          value: "ann1-value",
        },
      ],
      labels: [
        {
          name: "lbl1",
          value: "lbl1-value",
        },
      ],
    },
  },
  workload: {
    metadata: {
      annotations: {
        clusterId: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
        workloadId: "e82667b2-22cd-4204-92c9-1a76c09a5056",
      },
      creationTimestamp: null,
      name: "test-mock-ws",
      namespace: "runai-test-1",
    },
    spec: {
      annotations: {
        items: {
          clusterId: {
            value: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
          },
          workloadId: {
            value: "e82667b2-22cd-4204-92c9-1a76c09a5056",
          },
        },
      },
      cpu: {
        value: "0",
      },
      environment: {
        items: {
          LOG_DIR: {
            value: "./",
          },
          RUNAI_JOB_NAME: {
            value: "${RUNAI_JOB_NAME}",
          },
          RUNAI_PROJECT: {
            value: "${RUNAI_PROJECT}",
          },
        },
      },
      exposedUrls: {
        items: {
          "url-0": {
            value: {
              containerPort: 8888,
              customUrl: false,
            },
          },
          "url-1": {
            value: {
              containerPort: 6006,
              customUrl: false,
            },
          },
        },
      },
      gpu: {
        value: "0",
      },
      image: {
        value: "gcr.io/run-ai-demo/jupyter-tensorboard",
      },
      imagePullPolicy: {
        value: "IfNotPresent",
      },
      largeShm: {
        value: false,
      },
      memory: {
        value: "0M",
      },
      name: {
        value: "test-mock-ws",
      },
      username: {
        value: "test@run.ai",
      },
      volumes: {
        items: {
          "hostPath-test-hp": {
            value: {
              readOnly: false,
              sourcePath: "/test-path",
              targetPath: "/test-container",
            },
          },
          "nfs-test-nfs": {
            value: {
              nfsServer: "test-server",
              readOnly: false,
              sourcePath: "/test-path",
              targetPath: "/test-container-path",
            },
          },
        },
      },
    },
    status: {},
  },
  job: {},
};

const loadedTrainingByIdV2: TrainingV2 = {
  meta: {
    name: "tr-1",
    requestedName: "tr-1",
    workloadId: "ee04a2d0-3b27-42f8-87d0-6fff0e679cdc",
    projectId: "4500004",
    clusterId: "924f8015-d65f-4236-ac88-6bebe8481192",
    createdBy: "test@run.ai",
    createdAt: "2024-06-20T10:51:28.751473Z",
    desiredPhase: "Running",
  },
  assets: {
    environment: {
      id: "f1fca23e-56f4-4e88-b2be-2c7df30a28b2",
      name: "jupyter-tensorboard",
    },
    compute: {
      id: "8734ec10-9a51-4d7d-81b0-c8a110fe5386",
      name: "test-compute",
    },
    datasources: [
      {
        id: "a4e1499f-559e-45f4-b81e-46cb809db13f",
        name: "test-hp",
        kind: "hostPath",
      },
      {
        id: "d2792d6a-bd20-4e76-a60e-e2b2b347d468",
        name: "test-nfs",
        kind: "nfs",
      },
    ],
  },
  specificRunParams: {
    command: "tensorboard",
    args: "--logdir ./ --host 0.0.0.0 --path_prefix /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}",
    environmentVariables: [],
    nodePools: ["default"],
    autoDeletionTimeAfterCompletionSeconds: 2592000,
    backoffLimit: 6,
    connections: [
      {
        name: "TensorBoard",
      },
    ],
    allowOverQuota: false,
  },
};
const loadedTrainingMergedWithPolicy: TrainingV2 = {
  meta: {
    name: "",
    requestedName: "",
    workloadId: "",
    projectId: "",
    clusterId: "",
    createdBy: "",
    createdAt: "",
    desiredPhase: "Running",
  },
  assets: {
    environment: {
      id: "f1fca23e-56f4-4e88-b2be-2c7df30a28b2",
      name: "jupyter-tensorboard",
    },
    compute: {
      id: "8734ec10-9a51-4d7d-81b0-c8a110fe5386",
      name: "test-compute",
    },
    datasources: [
      {
        id: "a4e1499f-559e-45f4-b81e-46cb809db13f",
        name: "test-hp",
        kind: "hostPath",
      },
      {
        id: "d2792d6a-bd20-4e76-a60e-e2b2b347d468",
        name: "test-nfs",
        kind: "nfs",
      },
    ],
  },
  specificRunParams: {
    command: "tensorboard",
    args: "--logdir ./ --host 0.0.0.0 --path_prefix /${RUNAI_PROJECT}/${RUNAI_JOB_NAME}",
    environmentVariables: [],
    nodePools: ["default"],
    autoDeletionTimeAfterCompletionSeconds: 2592000,
    backoffLimit: 6,
    connections: [
      {
        name: "TensorBoard",
      },
    ],
    allowOverQuota: false,
  },
};

const loadedWorkloadTemplateById: WorkloadTemplate = {
  meta: {
    name: "temp-to-copy",
    scope: "project",
    clusterId: "4d724d30-3ae0-4771-8e19-4421c0ea9ab0",
    projectId: 3,
    id: "20071a79-9f07-42a6-9ddc-390f2ab9e1ac",
    kind: "workload-template",
    tenantId: 1,
    createdBy: "test@run.ai",
    createdAt: "2023-03-14T08:44:34.240673Z",
    updatedBy: "test@run.ai",
    updatedAt: "2023-03-14T08:44:34.240673Z",
    projectName: "project-1",
  },
  spec: {
    assets: {
      environment: {
        id: "75eb8c42-d8b8-47c6-9644-b3c17985ef8d",
        name: "jupyter-lab",
      },
      compute: {
        id: "3d617a5a-a905-4e20-83d3-05694eab9cd4",
        name: "cpu-only",
      },
    },
    specificEnv: {
      command: "start-notebook.sh",
      args: "--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''",
      environmentVariables: [],
      connections: [],
      allowOverQuota: false,
    },
  },
};

const trainingCreateMock: WorkloadCreationRequest = {
  name: "test-mock-ws",
  clusterId: "",
  projectId: 1,
  namespace: "",
  assets: {
    environment: "",
  },
};

export {
  trainingAssetsEditMockedData,
  validModel,
  loadedTrainingById,
  loadedTrainingByIdV2,
  loadedTrainingMergedWithPolicy,
  loadedWorkloadTemplateById,
  trainingCreateMock,
};
