/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, type Mock } from "vitest";
import { topologyLevelsFetcherService } from "./topology-levels-fetcher.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

const { createTopologyLevelsFetcher } = topologyLevelsFetcherService;

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNetworkTopologyLevelsByName: vi.fn().mockResolvedValue([]),
  },
}));

describe("topologyLevelsFetcherService.createTopologyLevelsFetcher", () => {
  it("resolves to an empty array and skips the request when name or clusterId is missing", async () => {
    const fetch = createTopologyLevelsFetcher();

    await expect(fetch(undefined, "cluster-123")).resolves.toEqual([]);
    await expect(fetch("topology-a", undefined)).resolves.toEqual([]);

    expect(clusterService.getNetworkTopologyLevelsByName).not.toHaveBeenCalled();
  });

  it("fetches and resolves the levels when both name and clusterId are given", async () => {
    (clusterService.getNetworkTopologyLevelsByName as Mock).mockResolvedValue(["region", "zone"]);
    const fetch = createTopologyLevelsFetcher();

    await expect(fetch("topology-a", "cluster-123")).resolves.toEqual(["region", "zone"]);

    expect(clusterService.getNetworkTopologyLevelsByName).toHaveBeenCalledWith("topology-a", "cluster-123");
  });

  it("treats a lookup failure as empty levels instead of throwing (levels are supplementary)", async () => {
    (clusterService.getNetworkTopologyLevelsByName as Mock).mockRejectedValue(new Error("network error"));
    const fetch = createTopologyLevelsFetcher();

    await expect(fetch("topology-a", "cluster-123")).resolves.toEqual([]);
  });

  it("resolves a superseded response to undefined so it can't overwrite the latest call's levels", async () => {
    let resolveFirst: (levels: string[]) => void = () => {};
    (clusterService.getNetworkTopologyLevelsByName as Mock)
      .mockImplementationOnce(() => new Promise((resolve) => (resolveFirst = resolve)))
      .mockResolvedValueOnce(["rack", "host"]);

    const fetch = createTopologyLevelsFetcher();

    const first = fetch("topology-a", "cluster-123");
    const second = fetch("topology-b", "cluster-123");
    await expect(second).resolves.toEqual(["rack", "host"]);

    // The stale first request resolves after the second one has already completed.
    resolveFirst(["region", "zone"]);
    await expect(first).resolves.toBeUndefined();
  });
});
