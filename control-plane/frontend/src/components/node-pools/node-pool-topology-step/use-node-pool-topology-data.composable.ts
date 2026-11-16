/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { getCurrentScope, onScopeDispose, ref, watch, type Ref } from "vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import type { INodePoolSummaryResources } from "@/models/node-pool.model";
import { extractErrorMessages } from "@/utils/error-alert.util";
import { loadSchedulingResources } from "../node-pool-scheduling-step/node-pool-scheduling.service";

export function useNodePoolTopologyData(poolId: Ref<string>) {
  const pool = ref<Nodepool>();
  const resources = ref<INodePoolSummaryResources>();
  const loading = ref(false);
  const loadError = ref("");
  const telemetryError = ref("");
  let version = 0;
  let disposed = false;

  async function refresh() {
    if (disposed) return;
    const request = ++version;
    const id = poolId.value;
    const current = () => !disposed && request === version;
    pool.value = undefined;
    resources.value = undefined;
    loadError.value = "";
    telemetryError.value = "";
    loading.value = !!id;
    if (!id) return;
    try {
      let snapshot: Nodepool;
      try {
        snapshot = await clusterService.getNodepool(id);
      } catch (error) {
        if (current()) loadError.value = extractErrorMessages(error, "Failed to load node pool.").join(" ");
        return;
      }
      if (!current()) return;
      pool.value = snapshot;
      try {
        const capacity = await loadSchedulingResources(snapshot);
        if (current()) resources.value = capacity;
      } catch (error) {
        if (current())
          telemetryError.value = extractErrorMessages(error, "Node pool capacity is unavailable.").join(" ");
      }
    } finally {
      if (current()) loading.value = false;
    }
  }

  watch(poolId, refresh, { immediate: true, flush: "sync" });
  if (getCurrentScope())
    onScopeDispose(() => {
      disposed = true;
      version++;
    });
  return { pool, resources, loading, loadError, telemetryError, refresh };
}
