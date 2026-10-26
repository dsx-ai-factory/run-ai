/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ref, type Ref } from "vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

export interface UseTopologyLevelsReturn {
  topologyLevels: Ref<string[]>;
  fetchTopologyLevels: (topologyName: string | null | undefined, clusterId: string | null | undefined) => Promise<void>;
}

/**
 * Shared by workload-details-tab and element-details-tab. Owns `topologyLevels` plus the
 * request-staleness guard: each call clears the levels immediately, treats a lookup failure as
 * empty (levels are supplementary to these panels, not critical), and discards a response
 * superseded by a newer call. Callers decide *when* to call `fetchTopologyLevels` — a reactive
 * watch on prop changes for element-details-tab, a step inside a larger fetch orchestration for
 * workload-details-tab — this only owns what happens on each call.
 */
export function useTopologyLevels(): UseTopologyLevelsReturn {
  const topologyLevels = ref<string[]>([]);
  let latestRequestId = 0;

  async function fetchTopologyLevels(
    topologyName: string | null | undefined,
    clusterId: string | null | undefined,
  ): Promise<void> {
    const requestId = ++latestRequestId;
    topologyLevels.value = [];
    if (!topologyName || !clusterId) return;

    const levels = await clusterService.getNetworkTopologyLevelsByName(topologyName, clusterId).catch(() => []);
    if (requestId !== latestRequestId) return;
    topologyLevels.value = levels;
  }

  return { topologyLevels, fetchTopologyLevels };
}
