/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from "vitest";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

const getNetworkTopologies = vi.fn();

vi.mock("@/services/infra/client-apis/cluster-service-api/cluster-service-api", () => ({
  clusterServiceApi: {
    networkTopologiesApi: {
      getNetworkTopologies: (...args: unknown[]) => getNetworkTopologies(...args),
    },
  },
}));

describe("clusterService.getNetworkTopologyLevelsByName", () => {
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

  it("returns an empty array instead of throwing when the lookup fails", async () => {
    getNetworkTopologies.mockRejectedValue(new Error("network error"));

    const levels = await clusterService.getNetworkTopologyLevelsByName("my-gb200-topology", "cluster-123");

    expect(levels).toEqual([]);
  });
});
