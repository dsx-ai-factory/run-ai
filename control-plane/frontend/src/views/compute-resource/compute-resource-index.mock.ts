/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { GpuRequestType, type ComputeAsset } from "@/swagger-models/assets-service-client";

export const computeResourceIndexMock = {
  computelist: [
    {
      meta: {
        name: "cpu-only",
        scope: "tenant",
        id: "20d5e09a-e3cf-43d3-8066-3b506cbd7fab",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-05-04T08:38:21.474935Z",
        updatedBy: "runai",
        updatedAt: "2023-05-04T08:38:21.474935Z",
      },
      spec: {
        gpuRequestType: GpuRequestType.Portion,
        gpuPortion: 0,
        cpuCoreRequest: 0.1,
        cpuMemoryRequest: "100M",
      },
    },
    {
      meta: {
        name: "small-fraction",
        scope: "tenant",
        id: "889ef349-4ed9-464b-97fd-3ff265f71f2e",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-05-04T08:38:21.475645Z",
        updatedBy: "runai",
        updatedAt: "2023-05-04T08:38:21.475645Z",
      },
      spec: {
        gpuRequestType: GpuRequestType.Portion,
        gpuPortion: 0.1,
        cpuCoreRequest: 0.1,
        cpuMemoryRequest: "100M",
      },
      usedBy: {
        workspaces: [
          {
            id: "50e4f51b-d469-4c41-9380-8c473ba9229b",
            name: "d1d1dd11d",
            status: "Pending",
          },
        ],
        assets: {},
      },
    },
    {
      meta: {
        name: "half-gpu",
        scope: "tenant",
        id: "30263424-88eb-4350-a5c2-4c41552015ee",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-05-04T08:38:21.475293Z",
        updatedBy: "runai",
        updatedAt: "2023-05-04T08:38:21.475293Z",
      },
      spec: {
        gpuRequestType: GpuRequestType.Portion,
        gpuPortion: 0.5,
        cpuCoreRequest: 0.1,
        cpuMemoryRequest: "100M",
      },
    },
    {
      meta: {
        name: "one-gpu",
        scope: "tenant",
        id: "c91d1eb4-91f0-4cf5-becf-6fab619c2f47",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-05-04T08:38:21.474722Z",
        updatedBy: "runai",
        updatedAt: "2023-05-04T08:38:21.474722Z",
      },
      spec: {
        gpuRequestType: GpuRequestType.Portion,
        gpuPortion: 1,
        cpuCoreRequest: 0.1,
        cpuMemoryRequest: "100M",
      },
    },
    {
      meta: {
        name: "two-gpus",
        scope: "tenant",
        id: "ebaeef3d-e135-43de-867f-cf47246c6848",
        kind: "compute",
        tenantId: 1,
        createdBy: "runai",
        createdAt: "2023-05-04T08:38:21.475507Z",
        updatedBy: "runai",
        updatedAt: "2023-05-04T08:38:21.475507Z",
      },
      spec: {
        gpuRequestType: GpuRequestType.Portion,
        gpuPortionRequest: 2,
        cpuCoreRequest: 0.2,
        cpuMemoryRequest: "200M",
      },
      usedBy: {
        trainings: [
          {
            id: "8682347b-4565-4ee6-a891-ef4c1135e072",
            name: "training-is-live",
            status: "Running",
          },
        ],
        assets: {},
      },
    },
  ] as ComputeAsset[],
  filterBy: {
    sortBy: "compute-resource",
    descending: false,
    page: 1,
    rowsPerPage: 20,
    columnFilters: [],
    searchTerm: "",
    displayedColumns: [
      "compute-resource",
      "project",
      "created-by",
      "created-at",
      "gpu",
      "cpu-cores",
      "cpu-memory",
      "workspaces",
      "templates",
    ],
    usageInfo: true,
  },
  filterByFromStorage: {
    clusterUuid: "",
    sortBy: "compute-resource",
    descending: false,
    page: 1,
    rowsPerPage: 20,
    columnFilters: [],
    searchTerm: "",
    displayedColumns: ["compute-resource", "gpu", "created-at", "cpu-cores", "cpu-memory"],
    usageInfo: true,
  },
};
