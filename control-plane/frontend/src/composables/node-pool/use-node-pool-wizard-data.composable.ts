/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { getCurrentScope, onScopeDispose, ref, watch, type Ref } from "vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { ForbiddenError, NotFoundError } from "@/models/http-response.model";
import type { INodePoolSummaryResources } from "@/models/node-pool.model";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import { extractErrorMessages } from "@/utils/error-alert.util";
import { loadSchedulingResources } from "./node-pool-scheduling-data.util";

interface IWizardDataOptions {
  retryCapacity?: boolean;
  resetStep?: (poolChanged: boolean) => void;
  loadStep?: (pool: Nodepool) => Promise<void>;
}

export function useNodePoolWizardData(poolId: Ref<string>, options: IWizardDataOptions = {}) {
  const pool = ref<Nodepool>();
  const resources = ref<INodePoolSummaryResources>();
  const loading = ref(false);
  const loadError = ref("");
  const telemetryError = ref("");
  let version = 0;
  let disposed = false;
  let retry: ReturnType<typeof setTimeout> | undefined;
  const retryCapacity = options.retryCapacity && !!getCurrentScope();

  async function refreshCapacity(snapshot: Nodepool, request: number, reloadInventory = false) {
    try {
      // A telemetry retry can fetch new inventory without replacing the user's editing snapshot.
      const latest = reloadInventory ? await clusterService.getNodepool(snapshot.id) : snapshot;
      if (request !== version) return;
      const capacity = await loadSchedulingResources(latest);
      if (request !== version) return;
      resources.value = capacity;
      telemetryError.value =
        retryCapacity && capacity.gpus == null ? "Capacity is not available yet. Retrying automatically." : "";
    } catch (error) {
      if (request !== version) return;
      telemetryError.value = extractErrorMessages(
        error,
        retryCapacity ? "Capacity is unavailable. Retrying automatically." : "Node pool capacity is unavailable.",
      ).join(" ");
    }
    if (retryCapacity && resources.value?.gpus == null)
      retry = setTimeout(() => void refreshCapacity(snapshot, request, true), 5000);
  }

  async function refresh() {
    if (disposed) return;
    clearTimeout(retry);
    const request = ++version;
    const id = poolId.value;
    const poolChanged = pool.value?.id !== id;
    if (poolChanged) pool.value = undefined;
    resources.value = undefined;
    loadError.value = "";
    telemetryError.value = "";
    options.resetStep?.(poolChanged);
    loading.value = !!id;
    if (!id) return;
    try {
      const snapshot = await clusterService.getNodepool(id);
      if (request !== version) return;
      pool.value = snapshot;
      await Promise.all([refreshCapacity(snapshot, request), options.loadStep?.(snapshot)]);
    } catch (error) {
      if (request === version) {
        if (error instanceof ForbiddenError || error instanceof NotFoundError) pool.value = undefined;
        loadError.value = extractErrorMessages(error, "Failed to load node pool.").join(" ");
      }
    } finally {
      if (request === version) loading.value = false;
    }
  }

  watch(poolId, refresh, { immediate: true, flush: "sync" });
  if (getCurrentScope())
    onScopeDispose(() => {
      disposed = true;
      version++;
      clearTimeout(retry);
    });
  return { pool, resources, loading, loadError, telemetryError, refresh };
}
