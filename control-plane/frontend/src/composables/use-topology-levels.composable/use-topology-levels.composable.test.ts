/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, type Mock } from "vitest";
import { useTopologyLevels } from "./use-topology-levels.composable";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNetworkTopologyLevelsByName: vi.fn().mockResolvedValue([]),
  },
}));

describe("useTopologyLevels", () => {
  it("clears levels immediately and skips the request when name or clusterId is missing", async () => {
    const { topologyLevels, fetchTopologyLevels } = useTopologyLevels();

    await fetchTopologyLevels(undefined, "cluster-123");
    expect(clusterService.getNetworkTopologyLevelsByName).not.toHaveBeenCalled();
    expect(topologyLevels.value).toEqual([]);

    await fetchTopologyLevels("topology-a", undefined);
    expect(clusterService.getNetworkTopologyLevelsByName).not.toHaveBeenCalled();
    expect(topologyLevels.value).toEqual([]);
  });

  it("fetches and applies levels when both name and clusterId are given", async () => {
    (clusterService.getNetworkTopologyLevelsByName as Mock).mockResolvedValue(["region", "zone"]);
    const { topologyLevels, fetchTopologyLevels } = useTopologyLevels();

    await fetchTopologyLevels("topology-a", "cluster-123");

    expect(clusterService.getNetworkTopologyLevelsByName).toHaveBeenCalledWith("topology-a", "cluster-123");
    expect(topologyLevels.value).toEqual(["region", "zone"]);
  });

  it("treats a lookup failure as empty levels instead of throwing (levels are supplementary)", async () => {
    (clusterService.getNetworkTopologyLevelsByName as Mock).mockRejectedValue(new Error("network error"));
    const { topologyLevels, fetchTopologyLevels } = useTopologyLevels();

    await expect(fetchTopologyLevels("topology-a", "cluster-123")).resolves.toBeUndefined();
    expect(topologyLevels.value).toEqual([]);
  });

  it("discards a superseded response so it cannot overwrite the latest call's levels", async () => {
    let resolveFirst: (levels: string[]) => void = () => {};
    (clusterService.getNetworkTopologyLevelsByName as Mock)
      .mockImplementationOnce(() => new Promise((resolve) => (resolveFirst = resolve)))
      .mockResolvedValueOnce(["rack", "host"]);

    const { topologyLevels, fetchTopologyLevels } = useTopologyLevels();

    const first = fetchTopologyLevels("topology-a", "cluster-123");
    const second = fetchTopologyLevels("topology-b", "cluster-123");
    await second;
    expect(topologyLevels.value).toEqual(["rack", "host"]);

    // The stale first request resolves after the second one has already completed.
    resolveFirst(["region", "zone"]);
    await first;

    expect(topologyLevels.value).toEqual(["rack", "host"]);
  });

  it("tracks staleness independently per composable instance", async () => {
    (clusterService.getNetworkTopologyLevelsByName as Mock).mockResolvedValue(["region"]);
    const a = useTopologyLevels();
    const b = useTopologyLevels();

    await a.fetchTopologyLevels("topology-a", "cluster-123");
    await b.fetchTopologyLevels("topology-b", "cluster-123");

    expect(a.topologyLevels.value).toEqual(["region"]);
    expect(b.topologyLevels.value).toEqual(["region"]);
  });
});
