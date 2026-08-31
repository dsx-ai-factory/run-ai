/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { InferenceV2 } from "@/swagger-models/assets-service-client";
import type { DisplayedCluster } from "@/swagger-models/cluster-service-client";
import type { ProjectMetadata } from "@/swagger-models/org-unit-service-client";
import { WorkloadDesiredPhase } from "@/swagger-models/workloads-client";

export const clusters: Array<DisplayedCluster> = [
  {
    uuid: "5b7f1e0a-b267-44fe-a7b1-16cd8cf3648f",
    tenantId: 1001,
    name: "mdev3-1754230200",
    createdAt: "2025-08-03T14:10:00.521936Z",
    domain: "https://mdev3.runailabs.com",
    version: "0.0.0-10001470.master-63e91c5",
    status: {
      dependencies: {
        required: {},
        optional: {
          knative: {
            available: true,
            components: {
              hpa: {
                available: true,
              },
              knative: {
                available: true,
                components: {
                  "enable-scale-to-zero": {
                    available: true,
                  },
                  "kubernetes.podspec-affinity": {
                    available: true,
                  },
                  "kubernetes.podspec-init-containers": {
                    available: true,
                  },
                  "kubernetes.podspec-nodeselector": {
                    available: true,
                  },
                  "kubernetes.podspec-persistent-volume-claim": {
                    available: true,
                  },
                  "kubernetes.podspec-persistent-volume-write": {
                    available: true,
                  },
                  "kubernetes.podspec-schedulername": {
                    available: true,
                  },
                  "kubernetes.podspec-securitycontext": {
                    available: true,
                  },
                  "kubernetes.podspec-tolerations": {
                    available: true,
                  },
                  "kubernetes.podspec-volumes-emptydir": {
                    available: true,
                  },
                  "multi-container": {
                    available: true,
                  },
                  "multi-container-probing": {
                    available: false,
                    reason:
                      "Feature multi-container-probing in knative config-features config-map from knative-serving namespace, has actual value \u003c\u003e expected value \u003cenabled\u003e",
                  },
                },
              },
              kourier: {
                available: true,
              },
            },
          },
        },
      },
      state: "Connected",
    },
    updatedAt: "2025-08-03T14:10:00.521936Z",
    lastLiveness: "2025-08-07T12:13:00.686Z",
  },
];

export const projects: Array<ProjectMetadata> = [
  {
    id: "4500096",
    name: "p100",
    namespace: "runai-p100",
    enforceRunaiScheduler: true,
    phase: "Ready",
    clusterId: "5b7f1e0a-b267-44fe-a7b1-16cd8cf3648f",
    createdAt: "2025-08-07T10:50:24.414043Z",
    allocated: {},
    deserved: {},
  },
];

export const emptyMergedWorkload: InferenceV2 = {
  meta: {
    name: "",
    requestedName: "",
    workloadId: "",
    projectId: "",
    clusterId: "",
    createdBy: "",
    createdAt: "0001-01-01T00:00:00Z",
    desiredPhase: WorkloadDesiredPhase.Running,
  },
  assets: {
    environment: {
      id: "",
      name: "",
    },
  },
  specificRunParams: {},
};
