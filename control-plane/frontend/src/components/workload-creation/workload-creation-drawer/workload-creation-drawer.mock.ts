/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ComputeAsset, DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";

export const workloadCreationDrawerMock = {
  computesItems: [
    {
      name: "cpu-only",
      id: "df92e729-8bb6-4242-875f-c76e8fcac585",
      kind: "compute",
      description: "",
      icons: [
        {
          name: "default",
          icon: "ph-bold ph-hard-drives",
        },
      ],
      policy: {
        compliance: true,
      },
      createdAt: "2024-12-23T09:29:09.392705Z",
      lastUsed: "",
      disabled: false,
      details: [
        {
          title: "GPU devices per pod",
          content: ["GPU devices: 0"],
        },
        {
          title: "CPU compute per pod",
          content: ["Request: 0.1 Cores"],
        },
        {
          title: "CPU memory per pod",
          content: ["Request: 100 MB"],
        },
      ],
    },
    {
      name: "two-gpus",
      id: "5f2e370d-fe41-474e-846d-deceacdbf341",
      kind: "compute",
      description: "",
      icons: [
        {
          name: "default",
          icon: "ph-bold ph-hard-drives",
        },
      ],
      policy: {
        compliance: true,
      },
      createdAt: "2024-12-23T09:29:09.394596Z",
      lastUsed: "",
      disabled: false,
      details: [
        {
          title: "GPU devices per pod",
          content: ["GPU devices: 2"],
        },
        {
          title: "CPU compute per pod",
          content: ["Request: 0.2 Cores"],
        },
        {
          title: "CPU memory per pod",
          content: ["Request: 200 MB"],
        },
      ],
    },
    {
      name: "small-fraction",
      id: "ce5df1d0-7374-4a08-87cd-f1038c4f0752",
      kind: "compute",
      description: "",
      icons: [
        {
          name: "default",
          icon: "ph-bold ph-hard-drives",
        },
      ],
      policy: {
        compliance: true,
      },
      createdAt: "2024-12-23T09:29:09.395172Z",
      lastUsed: "",
      disabled: false,
      details: [
        {
          title: "GPU devices per pod",
          content: ["GPU devices: 1"],
        },
        {
          title: "GPU memory per device",
          content: ["Request: 10%"],
        },
        {
          title: "CPU compute per pod",
          content: ["Request: 0.1 Cores"],
        },
        {
          title: "CPU memory per pod",
          content: ["Request: 100 MB"],
        },
      ],
    },
    {
      name: "half-gpu",
      id: "2eecee2b-09bd-4d4e-96de-69a27d9ebb50",
      kind: "compute",
      description: "",
      icons: [
        {
          name: "default",
          icon: "ph-bold ph-hard-drives",
        },
      ],
      policy: {
        compliance: true,
      },
      createdAt: "2024-12-23T09:29:09.393973Z",
      lastUsed: "",
      disabled: false,
      details: [
        {
          title: "GPU devices per pod",
          content: ["GPU devices: 1"],
        },
        {
          title: "GPU memory per device",
          content: ["Request: 50%"],
        },
        {
          title: "CPU compute per pod",
          content: ["Request: 0.1 Cores"],
        },
        {
          title: "CPU memory per pod",
          content: ["Request: 100 MB"],
        },
      ],
    },
    {
      name: "one-gpu",
      id: "e6ca4eb0-5718-4aea-bb68-6842b1f1d814",
      kind: "compute",
      description: "",
      icons: [
        {
          name: "default",
          icon: "ph-bold ph-hard-drives",
        },
      ],
      policy: {
        compliance: true,
      },
      createdAt: "2024-12-23T09:29:09.392177Z",
      lastUsed: "",
      disabled: false,
      details: [
        {
          title: "GPU devices per pod",
          content: ["GPU devices: 1"],
        },
        {
          title: "GPU memory per device",
          content: ["Request: 100%"],
        },
        {
          title: "CPU compute per pod",
          content: ["Request: 0.1 Cores"],
        },
        {
          title: "CPU memory per pod",
          content: ["Request: 100 MB"],
        },
      ],
    },
  ],
  computeList: [
    {
      meta: {
        name: "cpu-only",
        scope: "tenant",
        id: "df92e729-8bb6-4242-875f-c76e8fcac585",
        kind: "compute",
        tenantId: 1001,
        createdBy: "Run:ai",
        createdAt: "2024-12-23T09:29:09.392705Z",
        updatedBy: "Run:ai",
        updatedAt: "2024-12-23T09:29:09.392705Z",
      },
      spec: {
        gpuDevicesRequest: 0,
        gpuRequestType: "portion",
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
        id: "5f2e370d-fe41-474e-846d-deceacdbf341",
        kind: "compute",
        tenantId: 1001,
        createdBy: "Run:ai",
        createdAt: "2024-12-23T09:29:09.394596Z",
        updatedBy: "Run:ai",
        updatedAt: "2024-12-23T09:29:09.394597Z",
      },
      spec: {
        gpuDevicesRequest: 2,
        gpuRequestType: "portion",
        cpuCoreRequest: 0.2,
        cpuMemoryRequest: "200M",
      },
      compliance: {
        compliance: true,
      },
    },
    {
      meta: {
        name: "small-fraction",
        scope: "tenant",
        id: "ce5df1d0-7374-4a08-87cd-f1038c4f0752",
        kind: "compute",
        tenantId: 1001,
        createdBy: "Run:ai",
        createdAt: "2024-12-23T09:29:09.395172Z",
        updatedBy: "Run:ai",
        updatedAt: "2024-12-23T09:29:09.395172Z",
      },
      spec: {
        gpuDevicesRequest: 1,
        gpuRequestType: "portion",
        gpuPortionRequest: 0.1,
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
        id: "2eecee2b-09bd-4d4e-96de-69a27d9ebb50",
        kind: "compute",
        tenantId: 1001,
        createdBy: "Run:ai",
        createdAt: "2024-12-23T09:29:09.393973Z",
        updatedBy: "Run:ai",
        updatedAt: "2024-12-23T09:29:09.393973Z",
      },
      spec: {
        gpuDevicesRequest: 1,
        gpuRequestType: "portion",
        gpuPortionRequest: 0.5,
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
        id: "e6ca4eb0-5718-4aea-bb68-6842b1f1d814",
        kind: "compute",
        tenantId: 1001,
        createdBy: "Run:ai",
        createdAt: "2024-12-23T09:29:09.392177Z",
        updatedBy: "Run:ai",
        updatedAt: "2024-12-23T09:29:09.392177Z",
      },
      spec: {
        gpuDevicesRequest: 1,
        gpuRequestType: "portion",
        gpuPortionRequest: 1,
        cpuCoreRequest: 0.1,
        cpuMemoryRequest: "100M",
      },
      compliance: {
        compliance: true,
      },
    },
  ] as ComputeAsset[],
  dataSourceList: [
    {
      meta: {
        name: "hp-1",
        scope: "cluster",
        clusterId: "c2d1e6a3-d4d6-4ecb-871a-ab8b1e64b791",
        id: "2b72d044-76c5-459d-ba8b-c64318872057",
        kind: "hostPath",
        tenantId: 1001,
        createdBy: "test@run.ai",
        createdAt: "2024-12-23T14:14:53.104798Z",
        updatedBy: "test@run.ai",
        updatedAt: "2024-12-23T14:14:53.104798Z",
      },
      spec: {
        hostPath: {
          path: "/wallakia",
          readOnly: false,
          mountPath: "/cp111",
        },
      },
    },
    {
      meta: {
        name: "nfs-1",
        scope: "cluster",
        clusterId: "c2d1e6a3-d4d6-4ecb-871a-ab8b1e64b791",
        id: "fa47f7fb-61a7-48a7-8819-8238fc3260fc",
        kind: "nfs",
        tenantId: 1001,
        createdBy: "test@run.ai",
        createdAt: "2024-12-23T14:14:27.307998Z",
        updatedBy: "test@run.ai",
        updatedAt: "2024-12-23T14:14:27.307998Z",
      },
      spec: {
        nfs: {
          path: "/hopa",
          readOnly: false,
          server: "mine",
          mountPath: "/cp",
        },
      },
    },
  ] as DatasourceListResponseEntry[],
};
