/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// @ts-nocheck
import { describe, it, expect } from "vitest";
import { clusterUtil } from "@/utils/cluster.util/cluster.util";
import type { IStatusColOptions } from "@/models/table.model";
import type { ClusterDependenciesStatus } from "@/swagger-models/cluster-service-client";
import { ClusterState, ClusterReportedStatusPlatformTypeEnum } from "@/swagger-models/cluster-service-client";

describe("cluster.util", () => {
  it("should handle status without a corresponding color", () => {
    const status = {
      state: "NonexistentStatus",
    };

    const result = clusterUtil.getStatusColOptions(status);
    expect(result.color).toBe("warning"); // or adjust based on your default behavior
  });

  it("should handle status with missing prerequisites", () => {
    const status = {
      state: ClusterState.MissingPrerequisites,
      dependencies: {
        required: {
          dependency1: { available: false },
          dependency2: { available: false },
        },
      },
    };

    const result: IStatusColOptions = clusterUtil.getStatusColOptions(status);
    expect(result.tooltipText).toContain("dependency1");
    expect(result.tooltipText).toContain("dependency2");
  });

  it("renders the fractioning reason rather than naming the gpu-stack dependency", () => {
    const status = {
      state: ClusterState.MissingPrerequisites,
      dependencies: {
        required: {
          "gpu-stack": {
            available: false,
            reason: "NVIDIA GPU Operator version 26.6.0 is below minimum supported version v26.7.1",
            components: {
              nvidia: { available: true },
              "gpu-fractioning": {
                available: false,
                reason: "NVIDIA GPU Operator version 26.6.0 is below minimum supported version v26.7.1",
              },
            },
          },
        },
      },
    };

    const result: IStatusColOptions = clusterUtil.getStatusColOptions(status);

    expect(result.tooltipText).toContain(
      "NVIDIA GPU Operator version 26.6.0 is below minimum supported version v26.7.1",
    );
    expect(result.tooltipText).toContain("As a result, GPU fractions are not supported.");
    expect(result.tooltipText).not.toContain("gpu-stack");
  });

  it("keeps both sections, and one guide link, when fractioning and another dependency fail together", () => {
    const status = {
      state: ClusterState.MissingPrerequisites,
      dependencies: {
        required: {
          "gpu-stack": {
            available: false,
            reason: "NVIDIA GPU Operator version 26.6.0 is below minimum supported version v26.7.1",
            components: {
              nvidia: { available: true },
              "gpu-fractioning": {
                available: false,
                reason: "NVIDIA GPU Operator version 26.6.0 is below minimum supported version v26.7.1",
              },
            },
          },
          "kube-prometheus-stack": { available: false, reason: "Prometheus is not installed" },
        },
      },
    };

    const result: IStatusColOptions = clusterUtil.getStatusColOptions(status);

    expect(result.tooltipText).toContain("kube-prometheus-stack");
    expect(result.tooltipText).toContain("As a result, some features may be impacted.");
    expect(result.tooltipText).toContain("As a result, GPU fractions are not supported.");
    expect(result.tooltipText).not.toContain("gpu-stack");
    expect(result.tooltipText.match(/As a result,/g)).toHaveLength(2);
  });

  it("keeps gpu-stack named when a nested component under an available one has failed", () => {
    const status = {
      state: ClusterState.MissingPrerequisites,
      dependencies: {
        required: {
          "gpu-stack": {
            available: false,
            components: {
              nvidia: {
                available: true,
                components: {
                  "nvidia-metrics": { available: false, reason: "Metrics are missing in the DCGM-exporter." },
                },
              },
              "gpu-fractioning": { available: false, reason: "fractions are not supported on this driver" },
            },
          },
        },
      },
    };

    const result: IStatusColOptions = clusterUtil.getStatusColOptions(status);

    expect(result.tooltipText).toContain("gpu-stack");
    expect(result.tooltipText).toContain("As a result, GPU fractions are not supported.");
  });

  it("keeps gpu-stack named when fractioning is not its only failing component", () => {
    const status = {
      state: ClusterState.MissingPrerequisites,
      dependencies: {
        required: {
          "gpu-stack": {
            available: false,
            reason: "dcgm-exporter service not found in the cluster",
            components: {
              nvidia: { available: false, reason: "dcgm-exporter service not found in the cluster" },
              "gpu-fractioning": { available: false, reason: "fractions are not supported on this driver" },
            },
          },
        },
      },
    };

    const result: IStatusColOptions = clusterUtil.getStatusColOptions(status);

    expect(result.tooltipText).toContain("gpu-stack");
    expect(result.tooltipText).toContain("fractions are not supported on this driver");
    expect(result.tooltipText).toContain("As a result, GPU fractions are not supported.");
  });

  it("names the dependency when gpu-stack fails for a reason other than fractioning", () => {
    const status = {
      state: ClusterState.MissingPrerequisites,
      dependencies: {
        required: {
          "gpu-stack": {
            available: false,
            reason: "dcgm-exporter service not found in the cluster",
            components: {
              nvidia: { available: false, reason: "dcgm-exporter service not found in the cluster" },
              "gpu-fractioning": { available: true },
            },
          },
        },
      },
    };

    const result: IStatusColOptions = clusterUtil.getStatusColOptions(status);

    expect(result.tooltipText).toContain("gpu-stack");
    expect(result.tooltipText).not.toContain("GPU fractions are not supported.");
  });

  it('should handle "Connected" status and set color to "success"', () => {
    const status = {
      state: ClusterState.Connected,
    };

    const result: IStatusColOptions = clusterUtil.getStatusColOptions(status);
    expect(result.color).toBe("success");
  });

  it('should handle "WaitingToConnect" status and set color to "normal"', () => {
    const status = {
      state: ClusterState.WaitingToConnect,
    };

    const result: IStatusColOptions = clusterUtil.getStatusColOptions(status);
    expect(result.color).toBe("normal");
  });

  it("should handle an invalid status and return default options", () => {
    const status = {
      state: "InvalidStatus",
    };

    const result: IStatusColOptions = clusterUtil.getStatusColOptions(status);
    expect(result).toEqual({
      color: "warning",
      displayAnimation: false,
      filterKey: "status",
      status: "-",
      tooltipText: "-",
      colStyle: {
        "min-width": "150px",
      },
    });
  });

  it("should handle cluster distribution", () => {
    expect(clusterUtil.getDistribution(ClusterReportedStatusPlatformTypeEnum.Gke)).toEqual("GKE");
    expect(clusterUtil.getDistribution(ClusterReportedStatusPlatformTypeEnum.Aks)).toEqual("AKS");
    expect(clusterUtil.getDistribution(ClusterReportedStatusPlatformTypeEnum.Eks)).toEqual("EKS");
    expect(clusterUtil.getDistribution(ClusterReportedStatusPlatformTypeEnum.Vanilla)).toEqual("Vanilla");
    expect(clusterUtil.getDistribution(ClusterReportedStatusPlatformTypeEnum.Openshift)).toEqual("OpenShift");
    expect(clusterUtil.getDistribution("bla")).toEqual("-");
  });

  // Unit tests for getMissingPrerequisites function
  describe("getMissingPrerequisites", () => {
    it("should return missing prerequisites", () => {
      const dependencies: ClusterDependenciesStatus = {
        required: {
          dep1: { available: true },
          dep2: { available: false },
          dep3: { available: true },
        },
      };

      const result = clusterUtil.getMissingPrerequisites(dependencies);

      expect(result).toEqual(["dep2"]);
    });

    it("should return an empty array when all prerequisites are available", () => {
      const dependencies: ClusterDependenciesStatus = {
        required: {
          dep1: { available: true },
          dep2: { available: true },
          dep3: { available: true },
        },
      };

      const result = clusterUtil.getMissingPrerequisites(dependencies);

      expect(result).toEqual([]);
    });
  });

  // Unit tests for findClusterById function
  describe("findClusterById", () => {
    const mockClusters = [
      { uuid: "cluster-1", name: "Cluster 1", version: "1.0.0" },
      { uuid: "cluster-2", name: "Cluster 2", version: "2.0.0" },
      { uuid: "cluster-3", name: "Cluster 3" },
    ];

    it("should find cluster by id", () => {
      const result = clusterUtil.findClusterById(mockClusters, "cluster-2");
      expect(result).toEqual({ uuid: "cluster-2", name: "Cluster 2", version: "2.0.0" });
    });

    it("should return undefined when cluster not found", () => {
      const result = clusterUtil.findClusterById(mockClusters, "non-existent");
      expect(result).toBeUndefined();
    });

    it("should return undefined when clusters array is empty", () => {
      const result = clusterUtil.findClusterById([], "cluster-1");
      expect(result).toBeUndefined();
    });
  });

  // Unit tests for formatClusterNameWithVersion function
  describe("formatClusterNameWithVersion", () => {
    it("should format cluster name with version when version exists", () => {
      const cluster = { name: "Test Cluster", version: "1.0.0" };
      const result = clusterUtil.formatClusterNameWithVersion(cluster);
      expect(result).toBe("Test Cluster (v1.0.0)");
    });

    it("should return only cluster name when version is missing", () => {
      const cluster = { name: "Test Cluster" };
      const result = clusterUtil.formatClusterNameWithVersion(cluster);
      expect(result).toBe("Test Cluster");
    });

    it("should return only cluster name when version is empty string", () => {
      const cluster = { name: "Test Cluster", version: "" };
      const result = clusterUtil.formatClusterNameWithVersion(cluster);
      expect(result).toBe("Test Cluster");
    });

    it("should return only cluster name when version is null", () => {
      const cluster = { name: "Test Cluster", version: null };
      const result = clusterUtil.formatClusterNameWithVersion(cluster);
      expect(result).toBe("Test Cluster");
    });

    it("should return only cluster name when version is undefined", () => {
      const cluster = { name: "Test Cluster", version: undefined };
      const result = clusterUtil.formatClusterNameWithVersion(cluster);
      expect(result).toBe("Test Cluster");
    });

    it("should return empty string when cluster is null", () => {
      const result = clusterUtil.formatClusterNameWithVersion(null);
      expect(result).toBe("");
    });

    it("should return empty string when cluster is undefined", () => {
      const result = clusterUtil.formatClusterNameWithVersion(undefined);
      expect(result).toBe("");
    });
  });

  describe("isClusterReachable", () => {
    it.each([ClusterState.Connected, ClusterState.MissingPrerequisites, ClusterState.ServiceIssues])(
      "treats %s as reachable",
      (state) => {
        expect(clusterUtil.isClusterReachable(state)).toBe(true);
      },
    );

    it.each([ClusterState.WaitingToConnect, ClusterState.Disconnected, ClusterState.Unknown, undefined])(
      "treats %s as unreachable",
      (state) => {
        expect(clusterUtil.isClusterReachable(state)).toBe(false);
      },
    );
  });
});
