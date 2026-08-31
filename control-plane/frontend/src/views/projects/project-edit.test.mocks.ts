/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Constants
export const CLUSTER_ID = "test-cluster-id";

export const PROJECT_ID = "project-123";
export const DEPARTMENT_ID = "dept-1";

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

export const mockProject = {
  id: PROJECT_ID,
  name: "Test Project",
  description: "Test Project Description",
  clusterId: CLUSTER_ID,
  parentId: DEPARTMENT_ID,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 5 },
      cpu: { deserved: 1000 },
    },
  ],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  createdBy: "admin",
};

// Department with limited CPU quota (deserved is a number)
export const mockDepartmentWithLimitedCpu = {
  id: DEPARTMENT_ID,
  name: "Limited CPU Department",
  description: "",
  clusterId: CLUSTER_ID,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 10 },
      cpu: { deserved: 4000 },
      memory: { deserved: 8192, units: "MB" },
    },
  ],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  createdBy: "admin",
};

// Department with unlimited CPU quota (deserved is null)
export const mockDepartmentWithUnlimitedCpu = {
  id: DEPARTMENT_ID,
  name: "Unlimited CPU Department",
  description: "",
  clusterId: CLUSTER_ID,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 10 },
      cpu: { deserved: null },
      memory: { deserved: null, units: "MB" },
    },
  ],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  createdBy: "admin",
};

// Department with node types (training + workspace)
export const mockDepartmentWithNodeTypes = {
  id: DEPARTMENT_ID,
  name: "NodeTypes Department",
  description: "",
  clusterId: CLUSTER_ID,
  resources: [
    {
      nodePool: { name: "default", id: "pool-1" },
      gpu: { deserved: 10 },
    },
  ],
  nodeTypes: {
    training: ["nt-1", "nt-2"],
    workspace: ["nt-3"],
    names: { "nt-1": "GPU-A100", "nt-2": "GPU-V100", "nt-3": "CPU-Only" },
  },
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

// Department with multiple default node pools (the customer's scenario)
export const mockDepartmentWithMultipleDefaultPools = {
  id: DEPARTMENT_ID,
  name: "Multi-Pool Department",
  description: "",
  clusterId: CLUSTER_ID,
  defaultNodePools: ["default", "dist", "driver-test"],
  resources: [
    { nodePool: { name: "default", id: "pool-1" }, gpu: { deserved: 10 } },
    { nodePool: { name: "dist", id: "pool-2" }, gpu: { deserved: 5 } },
    { nodePool: { name: "driver-test", id: "pool-3" }, gpu: { deserved: 2 } },
  ],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  createdBy: "admin",
};

// Existing project that inherits from department (defaultNodePools: null in API)
// but effective shows the department's pools resolved
export const mockProjectInheritingPools = {
  id: PROJECT_ID,
  name: "Inheriting Project",
  clusterId: CLUSTER_ID,
  parentId: DEPARTMENT_ID,
  defaultNodePools: null,
  effective: {
    defaultNodePools: ["default", "dist", "driver-test"],
  },
  resources: [{ nodePool: { name: "default", id: "pool-1" }, gpu: { deserved: 5 } }],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  createdBy: "admin",
};

// Existing project with an explicit single-pool override already set
export const mockProjectWithExplicitDefaultPool = {
  id: PROJECT_ID,
  name: "Explicit Default Project",
  clusterId: CLUSTER_ID,
  parentId: DEPARTMENT_ID,
  defaultNodePools: ["default"],
  effective: {
    defaultNodePools: ["default"],
  },
  resources: [{ nodePool: { name: "default", id: "pool-1" }, gpu: { deserved: 5 } }],
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString(),
  createdBy: "admin",
};
