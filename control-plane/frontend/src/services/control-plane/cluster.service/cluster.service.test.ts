/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { fileService } from "@/services/file.service/file.service";
import { MetricsType } from "@/swagger-models/cluster-service-client";

const getNetworkTopologies = vi.fn();

vi.mock("@/services/infra/client-apis/cluster-service-api/cluster-service-api", () => ({
  clusterServiceApi: {
    networkTopologiesApi: {
      getNetworkTopologies: (...args: unknown[]) => getNetworkTopologies(...args),
    },
  },
}));

describe("clusterService.getNetworkTopologyLevelsByName", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("returns the levels of the matching topology", async () => {
    getNetworkTopologies.mockResolvedValue({
      data: { networkTopologies: [{ name: "my-gb200-topology", levels: ["region", "zone"] }] },
    });

    const levels = await clusterService.getNetworkTopologyLevelsByName("my-gb200-topology", "cluster-123");

    expect(levels).toEqual(["region", "zone"]);
    expect(getNetworkTopologies).toHaveBeenCalledWith(["name==my-gb200-topology", "clusterId==cluster-123"]);
  });

  it("returns an empty array when no topology matches", async () => {
    getNetworkTopologies.mockResolvedValue({ data: { networkTopologies: [] } });

    const levels = await clusterService.getNetworkTopologyLevelsByName("missing-topology", "cluster-123");

    expect(levels).toEqual([]);
  });

  it("throws a normalized error when the lookup fails, like its siblings in this service", async () => {
    getNetworkTopologies.mockRejectedValue(new Error("network error"));

    await expect(clusterService.getNetworkTopologyLevelsByName("my-gb200-topology", "cluster-123")).rejects.toThrow();
  });
});

describe("clusterService.getNodepoolMetricsCsv", () => {
  it("downloads the node pool metrics as CSV with an encoded node pool name", async () => {
    const downloadFileSpy = vi.spyOn(fileService, "downloadFile").mockResolvedValue();

    await clusterService.getNodepoolMetricsCsv(
      "cluster-1",
      "pool a/b",
      "2026-01-01T00:00:00.000Z",
      "2026-01-02T00:00:00.000Z",
      [MetricsType.TotalGpu],
      10,
      "pool.csv",
    );

    const [path, , filename, query] = downloadFileSpy.mock.calls[0];
    expect(path).toMatch(/\/clusters\/cluster-1\/nodepools\/pool%20a%2Fb\/metrics$/);
    expect(filename).toEqual("pool.csv");
    expect(query).toEqual({
      metricType: [MetricsType.TotalGpu],
      start: "2026-01-01T00:00:00.000Z",
      end: "2026-01-02T00:00:00.000Z",
      numberOfSamples: 10,
    });
  });
});
