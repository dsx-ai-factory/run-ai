/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, it, expect, vi } from "vitest";
import { workloadService } from "./workload.service";
import { workloadsManagerApi } from "@/services/infra/client-apis/workloads-manager-api/workloads-manager-api";

// models
import { ClusterState, type ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import { EWorkloadType, EExternalWorkloadType, EWorkloadName, Preemptibility } from "@/models/workload.model";
import type { Project } from "@/swagger-models/org-unit-service-client";
import { RunState } from "@/swagger-models/workloads-manager-client";
import { SourceApi, type Workload } from "@/swagger-models/workloads-client";

describe("workload.service", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("run state", () => {
    const workloadV2 = {
      id: "workload-id",
      sourceApi: SourceApi.WorkloadsV2,
      type: "Job",
    } as Workload;

    it("should suspend WorkloadsV2 workloads with the workloads manager API", async () => {
      const setWorkloadRunStateSpy = vi
        .spyOn(workloadsManagerApi.workloadsV2Api, "setWorkloadRunState")
        .mockResolvedValue({} as never);

      await workloadService.stopWorkloadV2(workloadV2);

      expect(setWorkloadRunStateSpy).toHaveBeenCalledWith("workload-id", { value: RunState.Suspend });
    });

    it("should resume WorkloadsV2 workloads with the workloads manager API", async () => {
      const setWorkloadRunStateSpy = vi
        .spyOn(workloadsManagerApi.workloadsV2Api, "setWorkloadRunState")
        .mockResolvedValue({} as never);

      await workloadService.activateWorkloadV2(workloadV2);

      expect(setWorkloadRunStateSpy).toHaveBeenCalledWith("workload-id", { value: RunState.Resume });
    });
  });

  describe("canSubmitWorkload", () => {
    const defaultCluster: ClustersMinimalData = {
      uuid: "12234",
      name: "my-cluster",
      tenantId: 0,
    };

    it("should return null if cluster is connected and has a domain", () => {
      const cluster: ClustersMinimalData = {
        ...defaultCluster,
        state: ClusterState.Connected,
        domain: "example.com",
      };
      const result = workloadService.canSubmitWorkload(cluster);
      expect(result).toBeNull();
    });

    it("should return info if cluster is disconnected", () => {
      const cluster: ClustersMinimalData = {
        ...defaultCluster,
        state: ClusterState.Disconnected,
        domain: "example.com",
      };
      const result = workloadService.canSubmitWorkload(cluster);
      expect(result).toEqual({ info: "Cluster is not connected" });
    });

    it("should return info if cluster is waiting to connect", () => {
      const cluster: ClustersMinimalData = {
        ...defaultCluster,
        state: ClusterState.WaitingToConnect,
        domain: "example.com",
      };
      const result = workloadService.canSubmitWorkload(cluster);
      expect(result).toEqual({ info: "Cluster is not connected" });
    });

    it("should return info if cluster domain is missing", () => {
      const cluster: ClustersMinimalData = {
        ...defaultCluster,
        state: ClusterState.Connected,
        domain: "",
      };
      const result = workloadService.canSubmitWorkload(cluster);
      expect(result).toEqual({ info: "cluster domain is missing or empty" });
    });

    it("should return info if cluster domain is undefined", () => {
      const cluster: ClustersMinimalData = {
        ...defaultCluster,
        state: ClusterState.Connected,
      };
      const result = workloadService.canSubmitWorkload(cluster);
      expect(result).toEqual({ info: "cluster domain is missing or empty" });
    });
  });

  describe("getFrameworkForUrl", () => {
    it("should return correct framework for distributed workload with pytorch", () => {
      const result = workloadService.getFrameworkForUrl(EWorkloadType.Distributed, EExternalWorkloadType.PyTorchJob);
      expect(result).toBe("pytorch");
    });

    it("should return correct framework for distributed workload with tensorflow", () => {
      const result = workloadService.getFrameworkForUrl(EWorkloadType.Distributed, EExternalWorkloadType.TFJob);
      expect(result).toBe("tf");
    });

    it("should return 'runai' for distributed workload without distributedFramework", () => {
      const result = workloadService.getFrameworkForUrl(EWorkloadType.Distributed, undefined);
      expect(result).toBe("runai");
    });

    it("should return 'runai' for distributed workload with unknown framework", () => {
      const result = workloadService.getFrameworkForUrl(EWorkloadType.Distributed, "unknown-framework");
      expect(result).toBe("runai");
    });

    it("should return 'knative' for inference workload", () => {
      const result = workloadService.getFrameworkForUrl(EWorkloadType.Inference);
      expect(result).toBe("knative");
    });

    it("should return 'runai' for training workload", () => {
      const result = workloadService.getFrameworkForUrl(EWorkloadType.Training);
      expect(result).toBe("runai");
    });

    it("should return 'runai' for workspace workload", () => {
      const result = workloadService.getFrameworkForUrl(EWorkloadType.Workspace);
      expect(result).toBe("runai");
    });

    it("should handle different distributed framework enum values", () => {
      const result = workloadService.getFrameworkForUrl(EWorkloadType.Distributed, EExternalWorkloadType.PyTorchJob);
      expect(result).toBe("pytorch");
    });

    it("should return 'runai' for unknown workload type", () => {
      const result = workloadService.getFrameworkForUrl("unknown-type");
      expect(result).toBe("runai");
    });
  });

  describe("getDefaultPreemptibilityByType", () => {
    it("should return Preemptible for Training workload", () => {
      const result = workloadService.getDefaultPreemptibilityByType(EWorkloadName.Training);
      expect(result).toBe(Preemptibility.Preemptible);
    });

    it("should return Preemptible for Distributed workload", () => {
      const result = workloadService.getDefaultPreemptibilityByType(EWorkloadName.Distributed);
      expect(result).toBe(Preemptibility.Preemptible);
    });

    it("should return NonPreemptible for Workspace workload", () => {
      const result = workloadService.getDefaultPreemptibilityByType(EWorkloadName.Workspace);
      expect(result).toBe(Preemptibility.NonPreemptible);
    });

    it("should return NonPreemptible for Inference workload", () => {
      const result = workloadService.getDefaultPreemptibilityByType(EWorkloadName.Inference);
      expect(result).toBe(Preemptibility.NonPreemptible);
    });
  });

  describe("exceedsAvailableQuota", () => {
    const mockProject: Partial<Project> = {
      id: "1",
      name: "test-project",
      resources: [
        {
          nodePool: { id: "np-1", name: "default" },
          gpu: { deserved: 10 },
        },
      ],
      status: {
        nodePoolQuotaStatuses: [
          {
            nodePoolId: "np-1",
            nodePoolName: "default",
            allocatedNonPreemptible: { gpu: 5 },
          },
        ],
      },
    };

    const mockProjectMultiNodePool: Partial<Project> = {
      id: "2",
      name: "test-project-multi",
      resources: [
        {
          nodePool: { id: "np-1", name: "default" },
          gpu: { deserved: 10 },
        },
        {
          nodePool: { id: "np-2", name: "np-secondary" },
          gpu: { deserved: 0 },
        },
      ],
      status: {
        nodePoolQuotaStatuses: [
          {
            nodePoolId: "np-1",
            nodePoolName: "default",
            allocatedNonPreemptible: { gpu: 5 },
          },
          {
            nodePoolId: "np-2",
            nodePoolName: "np-secondary",
            allocatedNonPreemptible: { gpu: 0 },
          },
        ],
      },
    };

    it("should return true when non-preemptible and exceeds quota", () => {
      const result = workloadService.exceedsAvailableQuota({
        currProject: mockProject as Project,
        selectedNodePoolsResources: ["default"],
        totalGpuRequest: 6, // 6 + 5 allocated = 11 > 10 quota
        preemptibility: Preemptibility.NonPreemptible,
      });
      expect(result).toBe(true);
    });

    it("should return false when non-preemptible and within quota", () => {
      const result = workloadService.exceedsAvailableQuota({
        currProject: mockProject as Project,
        selectedNodePoolsResources: ["default"],
        totalGpuRequest: 4, // 4 + 5 allocated = 9 < 10 quota
        preemptibility: Preemptibility.NonPreemptible,
      });
      expect(result).toBe(false);
    });

    it("should return false when preemptible regardless of quota", () => {
      const result = workloadService.exceedsAvailableQuota({
        currProject: mockProject as Project,
        selectedNodePoolsResources: ["default"],
        totalGpuRequest: 15,
        preemptibility: Preemptibility.Preemptible,
      });
      expect(result).toBe(false);
    });

    it("should return false when project is undefined", () => {
      const result = workloadService.exceedsAvailableQuota({
        currProject: undefined,
        selectedNodePoolsResources: ["default"],
        totalGpuRequest: 15,
        preemptibility: Preemptibility.NonPreemptible,
      });
      expect(result).toBe(false);
    });

    it("should return false when selectedNodePoolsResources is empty", () => {
      const result = workloadService.exceedsAvailableQuota({
        currProject: mockProject as Project,
        selectedNodePoolsResources: [],
        totalGpuRequest: 15,
        preemptibility: Preemptibility.NonPreemptible,
      });
      expect(result).toBe(false);
    });

    describe("multi-node-pool scenarios", () => {
      it("should return false when at least one node pool has sufficient quota", () => {
        // default: 10 quota, 5 allocated, request 4 -> 4 + 5 = 9 < 10 (OK)
        // np-secondary: 0 quota, 0 allocated, request 4 -> 4 + 0 = 4 > 0 (exceeds)
        // Since at least one node pool (default) can accommodate, should return false
        const result = workloadService.exceedsAvailableQuota({
          currProject: mockProjectMultiNodePool as Project,
          selectedNodePoolsResources: ["default", "np-secondary"],
          totalGpuRequest: 4,
          preemptibility: Preemptibility.NonPreemptible,
        });
        expect(result).toBe(false);
      });

      it("should return true only when ALL selected node pools exceed quota", () => {
        // default: 10 quota, 5 allocated, request 6 -> 6 + 5 = 11 > 10 (exceeds)
        // np-secondary: 0 quota, 0 allocated, request 6 -> 6 + 0 = 6 > 0 (exceeds)
        // Both exceed, should return true
        const result = workloadService.exceedsAvailableQuota({
          currProject: mockProjectMultiNodePool as Project,
          selectedNodePoolsResources: ["default", "np-secondary"],
          totalGpuRequest: 6,
          preemptibility: Preemptibility.NonPreemptible,
        });
        expect(result).toBe(true);
      });

      it("should return false when only node pool with zero quota is selected but request is zero", () => {
        // np-secondary: 0 quota, 0 allocated, request 0 -> 0 + 0 = 0, 0 < 0 is false (OK)
        const result = workloadService.exceedsAvailableQuota({
          currProject: mockProjectMultiNodePool as Project,
          selectedNodePoolsResources: ["np-secondary"],
          totalGpuRequest: 0,
          preemptibility: Preemptibility.NonPreemptible,
        });
        expect(result).toBe(false);
      });

      it("should return true when only node pool with zero quota is selected and request is positive", () => {
        // np-secondary: 0 quota, 0 allocated, request 1 -> 1 + 0 = 1 > 0 (exceeds)
        const result = workloadService.exceedsAvailableQuota({
          currProject: mockProjectMultiNodePool as Project,
          selectedNodePoolsResources: ["np-secondary"],
          totalGpuRequest: 1,
          preemptibility: Preemptibility.NonPreemptible,
        });
        expect(result).toBe(true);
      });
    });
  });
});
