/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Constants
export const CLUSTER_ID = "test-cluster-id";
export const DEPARTMENT_ID = "dept-123";

export const mockClusters = [{ name: "cluster-1", uuid: CLUSTER_ID }];

export const mockDepartment = {
  id: DEPARTMENT_ID,
  name: "Test Department",
  description: "Test Description",
  clusterId: CLUSTER_ID,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 10 },
      cpu: { deserved: 2000 },
    },
  ],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  createdBy: "admin",
};

export const mockCpuEnabledNodePool = {
  id: "pool-1",
  name: "default",
  clusterId: CLUSTER_ID,
  schedulingCapabilities: {
    adjustableWeights: true,
    ranks: true,
    quota: true,
    cpuQuota: true,
  },
};

export const mockCpuDisabledNodePool = {
  id: "pool-2",
  name: "gpu-only",
  clusterId: CLUSTER_ID,
  schedulingCapabilities: {
    adjustableWeights: true,
    ranks: true,
    quota: true,
    cpuQuota: false,
  },
};
