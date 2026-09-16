/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

export type TopologyLevelsFetcher = (
  topologyName: string | null | undefined,
  clusterId: string | null | undefined,
) => Promise<string[] | undefined>;

/**
 * Shared by workload-details-tab (Options API) and element-details-tab (<script setup>). Each call
 * bumps a request token so a superseded response (from a since-changed element/workload) resolves
 * to `undefined` instead of returning stale levels — callers should only apply the result when it's
 * not `undefined`.
 */
function createTopologyLevelsFetcher(): TopologyLevelsFetcher {
  let latestRequestId = 0;

  return async function fetchTopologyLevels(topologyName, clusterId) {
    const requestId = ++latestRequestId;
    if (!topologyName || !clusterId) {
      return requestId === latestRequestId ? [] : undefined;
    }

    // Levels are supplementary to this panel — a lookup failure shouldn't block the rest of the
    // Details tab, so it's treated the same as "no levels" here rather than surfaced to the caller.
    const levels = await clusterService.getNetworkTopologyLevelsByName(topologyName, clusterId).catch(() => []);
    return requestId === latestRequestId ? levels : undefined;
  };
}

export const topologyLevelsFetcherService = {
  createTopologyLevelsFetcher,
};
