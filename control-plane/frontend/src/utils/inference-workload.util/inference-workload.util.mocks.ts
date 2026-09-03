/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { EAccessOptions } from "@/components/environment/connection-access/connection-access-modal/connection-access-modal.model";
import { EInferenceArchitecture, EInferenceType, EServerType } from "@/models/inference.model";
import type { IUIInferenceCreation, IUIWorkloadCreation } from "@/models/workload.model";
import type { TWorkloadsToConvert } from "../workload.util/workload.util";
import { ClusterState, type ClustersMinimalData } from "@/swagger-models/cluster-service-client";

export const mockInferenceWorkloadUI: IUIInferenceCreation = {
  name: "hf31",
  namespace: "runai-p100",
  clusterId: "754e1142-1b5d-4f7d-9083-093921306723",
  projectId: 4500001,
  assets: {
    environment: "a50ef834-4008-4a3a-9ac8-bfc9c177f5e4",
    compute: "3ab6826a-2123-4b34-a686-64cdf47388ac",
  },
  specificEnv: {
    runAsUid: null,
    runAsGid: null,
    nodeType: null,
    allowOverQuota: null,
    autoDeletionTimeAfterCompletionSeconds: 2592000,
    environmentVariables: [
      {
        name: "RUNAI_MODEL",
        value: "openai-community/gpt2",
      },
    ],
    backoffLimit: 6,
    autoScaleData: {
      minReplicas: 1,
      maxReplicas: 1,
      thresholdMetric: undefined,
      thresholdValue: undefined,
      scaleToZeroRetentionSeconds: undefined,
    },
    servingPortAccess: {
      access: { accessOption: EAccessOptions.PUBLIC },
    },
  },
  architecture: EInferenceArchitecture.Standard,
  inferenceType: EInferenceType.AssetsBased,
  serverType: EServerType.Custom,
};

export const mockBaseInferenceWorkload: IUIWorkloadCreation = {
  name: "hf3",
  namespace: "",
  clusterId: "754e1142-1b5d-4f7d-9083-093921306723",
  projectId: 4500001,
  assets: {
    environment: "a50ef834-4008-4a3a-9ac8-bfc9c177f5e4",
    compute: "3ab6826a-2123-4b34-a686-64cdf47388ac",
    datasources: [],
  },
  specificEnv: {
    args: undefined,
    command: undefined,
    runAsUid: null,
    runAsGid: null,
    supplementalGroups: undefined,
    nodeType: null,
    allowOverQuota: null,
    nodePools: null,
    autoDeletionTimeAfterCompletionSeconds: null,
    connections: undefined,
    environmentVariables: [
      {
        name: "RUNAI_MODEL",
        value: "openai-community/gpt2",
      },
    ],
    annotations: undefined,
    labels: undefined,
    backoffLimit: 6,
    tolerations: undefined,
    podAffinity: undefined,
    parallelism: undefined,
    completions: undefined,
  },
};

export const mockInferenceWorkload: TWorkloadsToConvert = {
  meta: {
    name: "hf3",
    requestedName: "hf3",
    workloadId: "18f31c94-2df5-4adb-9202-35f1ae72d1a0",
    projectId: "4500001",
    clusterId: "754e1142-1b5d-4f7d-9083-093921306723",
    createdBy: "test@run.ai",
    createdAt: "2024-11-07T14:08:09.766903Z",
    desiredPhase: "Running",
  },
  assets: {
    environment: {
      id: "a50ef834-4008-4a3a-9ac8-bfc9c177f5e4",
      name: "llm-server",
    },
    compute: {
      id: "3ab6826a-2123-4b34-a686-64cdf47388ac",
      name: "gpu-memory-16g",
    },
  },
  specificRunParams: {
    environmentVariables: [
      {
        name: "RUNAI_MODEL",
        value: "openai-community/gpt2",
      },
    ],
    autoDeletionTimeAfterCompletionSeconds: null,
    autoScaling: {
      minReplicas: 1,
      maxReplicas: 1,
    },
  },
};

export const mockCluster: ClustersMinimalData = {
  uuid: "422f51ca",
  name: "mdev2-173609",
  tenantId: 1001,
  domain: "https://some.domain.com",
  version: "0.0.0-10000595.master-942f955",
  platform: {
    type: "vanilla",
    kubeVersion: "v1.28.0",
  },
  state: ClusterState.Connected,
  dependenciesStatus: {
    required: {
      "gpu-stack": {
        available: true,
        components: {
          amd: {
            available: false,
            reason: "daemonset amd-gpu-device-plugin-daemonset not found in the cluster",
          },
          nvidia: {
            available: true,
            components: {
              "nvidia-metrics": {
                available: true,
              },
            },
          },
        },
      },
      prometheus: {
        available: true,
        components: {
          Oauth2: {
            available: true,
          },
        },
      },
    },
    optional: {
      knative: {
        available: true,
        components: {
          hpa: {
            available: false,
            reason: "hpa autoscaler not found in the cluster",
          },
          knative: {
            available: true,
          },
          kourier: {
            available: true,
          },
        },
      },
      mpi: {
        available: true,
        components: {
          suspend: {
            available: true,
          },
        },
      },
      pytorch: {
        available: true,
        components: {
          suspend: {
            available: true,
          },
        },
      },
      tensorflow: {
        available: true,
        components: {
          suspend: {
            available: true,
          },
        },
      },
      xgboost: {
        available: true,
        components: {
          suspend: {
            available: true,
          },
        },
      },
    },
  },
};
