/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Workload } from "@/swagger-models/workloads-client";

export const workloadIndexMock = {
  workloads: [
    {
      priority: 100,
      category: "Workspace",
      tenantId: 1001,
      runningPods: 0,
      phaseUpdatedAt: "2024-06-19T12:48:12.065016Z",
      k8sPhaseUpdatedAt: "2024-06-19T12:48:12.065016Z",
      updatedAt: "2024-06-19T12:48:12.064961Z",
      source: "Control-plane",
      deletedAt: null,
      type: "Workspace",
      name: "first-ws",
      id: "bdbe3b86-7065-4614-b5d3-a8eb41725153",
      priorityClassName: "",
      submittedBy: "test@run.ai",
      clusterId: "924f8015-d65f-4236-ac88-6bebe8481192",
      projectName: "",
      projectId: "1",
      departmentName: "",
      departmentId: "1",
      namespace: "",
      createdAt: "2024-06-19T12:48:12.052Z",
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
      phase: "Creating",
      conditions: [],
      phaseMessage: "Workload creating in Control-Plane",
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
              connectionType: "",
            },
          ],
          name: "tensorboard",
          id: "842f17d5-eb9d-45c7-ba37-ec55d45a9fdb",
        },
      ],
      fullySaasWorkload: true,
    },
  ] as Workload[],
};
