/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { reactive, ref, watch, onScopeDispose } from "vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { ErrorAlert } from "@/utils/error-alert.util/error-alert.util";
import { ConflictError } from "@/models/http-response.model";
import { errorMessages } from "@/common/error-message.constant";
import { deepCopy } from "@/utils/common.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { NodepoolSortFilterFields, type Nodepool } from "@/swagger-models/cluster-service-client";
import { createNodePoolDraft } from "@/utils/node-pool.util/node-pool.util";

export function useNodePoolForm(clusterId: string) {
  const draft = reactive(createNodePoolDraft(clusterId));
  const pools = ref<Nodepool[]>([]);
  const loading = ref(false);
  const loadError = ref("");
  const submitError = ref("");
  const submitting = ref(false);
  const nameConflict = ref(false);
  watch(
    () => draft.name,
    () => {
      nameConflict.value = false;
    },
  );
  const created = ref<Nodepool>();
  let requestId = 0;
  let disposed = false;
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
    } catch (error) {
      if (currentRequest === requestId) {
        const fallback = ErrorAlert.failedLoadingMessage("node pools");
        const { message } = new ErrorAlert({ generalMessage: fallback }).getNotification(error);
        loadError.value = `${message || fallback}. ${errorMessages.NODE_POOL_LOADING_RETRY}`;
      }
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
      nameConflict.value = false;
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
    if (created.value) return created.value;
    submitting.value = true;
    submitError.value = "";
    const payload = deepCopy(draft);
    try {
      created.value = await clusterService.createNodepool(payload);
      if (disposed) return;
      return created.value;
    } catch (error) {
      if (disposed) return;
      // Duplicate names currently arrive as a generic server error. Refresh the
      // cluster snapshot after failure so the field can show a name conflict.
      await loadPools();
      nameConflict.value =
        pools.value.some((pool) => pool.name === payload.name) ||
        (error instanceof ConflictError && /name/i.test(error.message));
      const context = ErrorAlert.failedCreateMessage(`node pool ${payload.name}`);
      const { message } = new ErrorAlert({ generalMessage: context }).getNotification(error);
      submitError.value = typeof message === "string" && message !== context ? `${context} ${message}` : context;
    } finally {
      submitting.value = false;
    }
  }
  return { draft, pools, loading, loadError, submitError, submitting, created, nameConflict, loadPools, create };
}
