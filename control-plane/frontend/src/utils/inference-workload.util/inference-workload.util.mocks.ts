/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ClusterState, type ClustersMinimalData } from "@/swagger-models/cluster-service-client";

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
