/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { computed, reactive, ref, watch, onScopeDispose } from "vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { deepCopy } from "@/utils/common.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { NodepoolSortFilterFields, type Nodepool } from "@/swagger-models/cluster-service-client";
import { isProportional, PROPORTIONAL_CAPABILITIES, type NodePoolCreationDraft } from "./node-pool-create-flow.model";

export function useNodePoolCreation(clusterId: string) {
  const draft = reactive<NodePoolCreationDraft>({
    ...nodePoolUtil.getEmptyNodePool(clusterId),
    description: "",
    schedulingCapabilities: { ...PROPORTIONAL_CAPABILITIES },
  });
  const pools = ref<Nodepool[]>([]);
  const loading = ref(false);
  const loadError = ref("");
  const submitError = ref("");
  const submitting = ref(false);
  const created = ref<Nodepool>();
  const modeMismatch = computed(() => !!created.value && !isProportional(created.value.schedulingCapabilities));
  let requestId = 0;
  let disposed = false;
  let quotasPrepared = false;
  onScopeDispose(() => {
    disposed = true;
    requestId++;
  });
  async function loadPools() {
    const currentRequest = ++requestId;
    pools.value = [];
    loadError.value = "";
    if (!draft.clusterId) {
      loading.value = false;
      return;
    }
    loading.value = true;
    try {
      const result = await clusterService.getNodepools([
        filterUtil.getEqualsFilterString(NodepoolSortFilterFields.ClusterId, draft.clusterId),
      ]);
      if (currentRequest === requestId) pools.value = result;
    } catch {
      if (currentRequest === requestId)
        loadError.value = "Failed to load node pools. Retry before creating a node pool.";
    } finally {
      if (currentRequest === requestId) loading.value = false;
    }
  }
  watch(
    () => draft.clusterId,
    () => {
      draft.labelKey = "";
      draft.labelValue = "";
      submitError.value = "";
      void loadPools();
    },
    { immediate: true },
  );
  watch(
    () => draft.labelKey,
    () => {
      draft.labelValue = "";
    },
  );
  async function create() {
    if (submitting.value || loading.value || loadError.value || !draft.clusterId || disposed) return;
    if (created.value && !modeMismatch.value) return created.value;
    submitting.value = true;
    submitError.value = "";
    const payload = deepCopy(draft);
    try {
      if (!created.value) {
        if (!(await clusterService.isNodePoolNameUnique(payload.clusterId, payload.name))) {
          submitError.value = "This name already exists";
          return;
        }
        if (disposed) return;
        created.value = await clusterService.createNodepool(payload);
      }
      if (disposed) return;
      // Creation currently ignores capabilities. Retain the ID so setup retries cannot create duplicate pools.
      if (modeMismatch.value) {
        // New pools inherit department quotas, which block disabling quota. Scope preparation to this new ID.
        if (created.value.schedulingCapabilities?.quota && !quotasPrepared) {
          await orgUnitService.resetNodePoolQuotas(payload.clusterId, created.value.id);
          quotasPrepared = true;
        }
        if (disposed) return;
        created.value = await clusterService.patchNodepool(created.value.id, {
          schedulingCapabilities: { ...PROPORTIONAL_CAPABILITIES },
        });
        if (modeMismatch.value) {
          submitError.value = `Node pool ${payload.name} was created, but the server did not save Proportional settings. Retry setup.`;
          return;
        }
      }
      return created.value;
    } catch {
      submitError.value = created.value
        ? `Node pool ${payload.name} was created, but Proportional settings couldn't be applied. Retry setup.`
        : `Node pool ${payload.name} couldn't be created`;
    } finally {
      submitting.value = false;
    }
  }
  return { draft, pools, loading, loadError, submitError, submitting, created, modeMismatch, loadPools, create };
}
