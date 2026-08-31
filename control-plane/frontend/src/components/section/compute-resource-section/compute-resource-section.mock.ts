/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { GpuRequestType, type ComputeAsset } from "@/swagger-models/assets-service-client";

const computeResources: ComputeAsset[] = [
  {
    meta: {
      name: "cpu-only",
      scope: "tenant",
      id: "0ff73d2c-30a9-49b2-97a9-5d3912a582ff",
      kind: "compute",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-03-02T15:20:37.605169Z",
      updatedBy: "runai",
      updatedAt: "2023-03-02T15:20:37.605169Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortionLimit: 0,
      cpuCoreRequest: 0.1,
      cpuMemoryRequest: "100M",
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "small-fraction",
      scope: "tenant",
      id: "a34813d3-1dc1-4111-9fba-8af290ee0de9",
      kind: "compute",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-03-02T15:20:37.60588Z",
      updatedBy: "runai",
      updatedAt: "2023-03-02T15:20:37.60588Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortionLimit: 0.1,
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
      id: "3e104f4b-f666-4c7f-8901-9f36defd54f2",
      kind: "compute",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-03-02T15:20:37.605532Z",
      updatedBy: "runai",
      updatedAt: "2023-03-02T15:20:37.605532Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortionLimit: 0.5,
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
      id: "2e97c721-83f4-45db-9859-d3a7bc2683ba",
      kind: "compute",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-03-02T15:20:37.604964Z",
      updatedBy: "runai",
      updatedAt: "2023-03-02T15:20:37.604964Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortionLimit: 1,
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
      id: "8c989a08-fa81-4a77-ae10-ae5b2245103d",
      kind: "compute",
      tenantId: 1,
      createdBy: "runai",
      createdAt: "2023-03-02T15:20:37.605726Z",
      updatedBy: "runai",
      updatedAt: "2023-03-02T15:20:37.605726Z",
    },
    spec: {
      gpuRequestType: GpuRequestType.Portion,
      gpuPortionLimit: 2,
      cpuCoreRequest: 0.2,
      cpuMemoryRequest: "200M",
    },
    compliance: {
      compliance: true,
    },
  },
];

export const mockComputeResourceSectionData = {
  computeResources,
};
