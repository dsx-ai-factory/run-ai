/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export const mockGpuAllocationRatio = {
  data: [
    {
      metadata: {
        departmentId: 1,
        departmentName: "default",
        clusterId: "b13c3cf9-4cfe-47f6-a6ca-15e2950668d7",
        clusterName: "barak-1-1711864663",
      },
      current: {
        resources: [
          {
            numberOfPendingWorkloads: 1,
            gpu: {
              quota: 100,
              allocated: 87,
            },
            cpu: {
              quota: 0,
              allocated: 8700,
            },
            memory: {
              quota: 0,
              allocated: 9122.6112,
            },
            nodepoolName: "default",
          },
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 100,
              allocated: 2,
            },
            cpu: {
              quota: 0,
              allocated: 200,
            },
            memory: {
              quota: 0,
              allocated: 200,
            },
            nodepoolName: "node-pool-2",
          },
        ],
        projectResources: [
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 16,
              allocated: 12,
            },
            cpu: {
              quota: 0,
              allocated: 1200,
            },
            memory: {
              quota: 0,
              allocated: 1258.2912000000001,
            },
            projectName: "griffin",
            nodepoolName: "default",
          },
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 32,
              allocated: 2,
            },
            cpu: {
              quota: 0,
              allocated: 200,
            },
            memory: {
              quota: 0,
              allocated: 200,
            },
            projectName: "griffin",
            nodepoolName: "node-pool-2",
          },
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 16,
              allocated: 12,
            },
            cpu: {
              quota: 0,
              allocated: 1200,
            },
            memory: {
              quota: 0,
              allocated: 1258.2912000000001,
            },
            projectName: "suricata",
            nodepoolName: "default",
          },
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 32,
              allocated: 0,
            },
            cpu: {
              quota: 0,
              allocated: 0,
            },
            memory: {
              quota: 0,
              allocated: 0,
            },
            projectName: "suricata",
            nodepoolName: "node-pool-2",
          },
          {
            numberOfPendingWorkloads: 1,
            gpu: {
              quota: 16,
              allocated: 63,
            },
            cpu: {
              quota: 0,
              allocated: 6300,
            },
            memory: {
              quota: 0,
              allocated: 6606.028799999999,
            },
            projectName: "lemur",
            nodepoolName: "default",
          },
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 32,
              allocated: 0,
            },
            cpu: {
              quota: 0,
              allocated: 0,
            },
            memory: {
              quota: 0,
              allocated: 0,
            },
            projectName: "lemur",
            nodepoolName: "node-pool-2",
          },
        ],
      },
    },
  ],
};

export const mockGpuAllocationRatioUnlimitedQuota = {
  data: [
    {
      metadata: {
        departmentId: 1,
        departmentName: "default",
        clusterId: "b13c3cf9-4cfe-47f6-a6ca-15e2950668d7",
        clusterName: "barak-1-1711864663",
      },
      current: {
        resources: [
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: -1,
              allocated: 10,
            },
            cpu: { quota: 0, allocated: 0 },
            memory: { quota: 0, allocated: 0 },
            nodepoolName: "default",
          },
        ],
        projectResources: [
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: -1,
              allocated: 5,
            },
            cpu: { quota: 0, allocated: 0 },
            memory: { quota: 0, allocated: 0 },
            projectName: "proj-a",
            nodepoolName: "default",
          },
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: -1,
              allocated: 5,
            },
            cpu: { quota: 0, allocated: 0 },
            memory: { quota: 0, allocated: 0 },
            projectName: "proj-b",
            nodepoolName: "default",
          },
        ],
      },
    },
  ],
};

export const mockGpuAllocationRatioMixedQuota = {
  data: [
    {
      metadata: {
        departmentId: 1,
        departmentName: "default",
        clusterId: "b13c3cf9-4cfe-47f6-a6ca-15e2950668d7",
        clusterName: "barak-1-1711864663",
      },
      current: {
        resources: [
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 100,
              allocated: 50,
            },
            cpu: { quota: 0, allocated: 0 },
            memory: { quota: 0, allocated: 0 },
            nodepoolName: "default",
          },
        ],
        projectResources: [
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 50,
              allocated: 30,
            },
            cpu: { quota: 0, allocated: 0 },
            memory: { quota: 0, allocated: 0 },
            projectName: "proj-limited",
            nodepoolName: "default",
          },
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: -1,
              allocated: 20,
            },
            cpu: { quota: 0, allocated: 0 },
            memory: { quota: 0, allocated: 0 },
            projectName: "proj-unlimited",
            nodepoolName: "default",
          },
        ],
      },
    },
  ],
};

export const mockGpuAllocationRatioFilterByNodepool = {
  data: [
    {
      metadata: {
        departmentId: 1,
        departmentName: "default",
        clusterId: "b13c3cf9-4cfe-47f6-a6ca-15e2950668d7",
        clusterName: "barak-1-1711864663",
      },
      current: {
        resources: [
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 100,
              allocated: 2,
            },
            cpu: {
              quota: 0,
              allocated: 200,
            },
            memory: {
              quota: 0,
              allocated: 200,
            },
            nodepoolName: "node-pool-2",
          },
        ],
        projectResources: [
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 32,
              allocated: 2,
            },
            cpu: {
              quota: 0,
              allocated: 200,
            },
            memory: {
              quota: 0,
              allocated: 200,
            },
            projectName: "griffin",
            nodepoolName: "node-pool-2",
          },
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 32,
              allocated: 0,
            },
            cpu: {
              quota: 0,
              allocated: 0,
            },
            memory: {
              quota: 0,
              allocated: 0,
            },
            projectName: "suricata",
            nodepoolName: "node-pool-2",
          },
          {
            numberOfPendingWorkloads: 0,
            gpu: {
              quota: 32,
              allocated: 0,
            },
            cpu: {
              quota: 0,
              allocated: 0,
            },
            memory: {
              quota: 0,
              allocated: 0,
            },
            projectName: "lemur",
            nodepoolName: "node-pool-2",
          },
        ],
      },
    },
  ],
};
