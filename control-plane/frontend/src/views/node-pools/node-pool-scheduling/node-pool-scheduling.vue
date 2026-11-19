<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-form-wrapper v-if="!pool" :form-state="{}" full-width>
    <runai-default-loader v-if="loading" />
    <q-banner v-else-if="loadError" class="bg-red-1 text-negative q-ma-lg" role="alert">
      {{ loadError }}
      <template #action
        ><runai-btn aid="retry-scheduling-pool" label="Retry" flat color="primary" @click="refresh"
      /></template>
    </q-banner>
  </runai-form-wrapper>
  <node-pool-scheduling-step
    v-else
    :key="pool.id"
    :pool="pool"
    :resources="resources"
    :rows="pages"
    :baseline-rows="baselineRows"
    :aggregates="aggregates"
    :ranks="ranks"
    :rank-counts="rankCounts"
    :loading="loading"
    :load-error="loadError"
    :telemetry-error="telemetryError"
    :can-view-departments="permissions.view"
    :can-edit-departments="permissions.edit"
    :can-save-pool="canSavePool"
    :can-reset-resources="canResetResources"
    :totals-complete="totalsComplete"
    :can-edit-department="canEditDepartment"
    :has-more="hasMore"
    :search-term="search"
    :rank-filter="rank"
    :quota-policy="quotaPolicy"
    :load-quota-context="loadQuotaContext"
    @refresh="refresh"
    @search="search = $event"
    @filter-rank="rank = $event"
    @load-more="loadMore"
    @load-rank="loadRank"
    @leave-approved="leaveApproved = true"
    @saving="saving = $event"
    @pool-saved="pool = $event"
    @saved="onSaved"
  />
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { onBeforeRouteLeave, onBeforeRouteUpdate, useRoute, useRouter } from "vue-router";
import NodePoolSchedulingStep from "@/components/node-pools/node-pool-scheduling-step/node-pool-scheduling-step.vue";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import RunaiFormWrapper from "@/components/common/runai-form-wrapper/runai-form-wrapper.vue";
import { useNodePoolSchedulingData } from "@/components/node-pools/node-pool-scheduling-step/use-node-pool-scheduling-data.composable";
import { loadChildQuotaSum } from "@/components/node-pools/node-pool-scheduling-step/node-pool-scheduling.service";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";
import { useAppStore } from "@/stores/app.store";
import { requestToLeave } from "@/services/infra/router.service/router.service";

import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
const route = useRoute();
const router = useRouter();
const permissionStore = usePermissionStore();
const canSavePool = computed(() => permissionStore.hasPermission(ResourceType.Nodepools, Action.Update));
const saving = ref(false);
async function onSaved(savedPool: Nodepool, destination: "close" | "continue" = "close") {
  leaveApproved.value = true;
  await router.push(
    destination === "continue"
      ? { name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, params: { id: savedPool.id }, query: { stepName: "topology" } }
      : { name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX },
  );
}
const appStore = useAppStore();
const poolId = computed(() => (typeof route.params.id === "string" ? route.params.id : ""));
const {
  snapshot: pool,
  pages,
  baselineRows,
  hasMore,
  aggregates,
  rankCounts,
  resources,
  permissions,
  totalsComplete,
  canEditDepartment,
  canResetResources,
  quotaPolicy,
  policyError,
  loadError,
  telemetryError,
  loading,
  search,
  rank,
  refresh,
  loadMore,
  loadRank,
} = useNodePoolSchedulingData(poolId);
const leaveApproved = ref(false);
const ranks = computed(() => ENTITY_PRIORITY_SELECT_OPTIONS.map((option) => option.value).reverse());
async function loadQuotaContext(id: string) {
  if (!pool.value || quotaPolicy.value === undefined)
    throw new Error(policyError.value || "Quota policy is unavailable.");
  const selected = pool.value;
  const childQuotaSum = await loadChildQuotaSum({ poolId: selected.id, clusterId: selected.clusterId }, id);
  return { childQuotaSum, enforceChildQuota: quotaPolicy.value };
}
watch(
  pool,
  (value) => {
    appStore.setPageLoading(false);
    appStore.setPageTitleOverride(value ? `Edit node pool: ${value.name}` : null);
  },
  { immediate: true },
);
watch(poolId, () => {
  leaveApproved.value = false;
});
function canLeave() {
  if (saving.value) return false;
  return leaveApproved.value || !pool.value || requestToLeave();
}
onBeforeRouteLeave(canLeave);
onBeforeRouteUpdate(canLeave);
onBeforeUnmount(() => {
  appStore.setPageTitleOverride(null);
});
</script>
