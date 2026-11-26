<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <node-pool-wizard-load-state v-if="!pool" step="scheduling" :loading="loading" :error="loadError" @retry="refresh" />
  <node-pool-scheduling-step
    v-else
    :key="pool.id"
    :pool="pool"
    :resources="resources"
    :department-pages="departmentPages"
    :rows="baselineRows"
    :baseline-rows="baselineRows"
    :aggregates="aggregates"
    :ranks="ranks"
    :rank-counts="rankCounts"
    :loading="loading"
    :load-error="loadError"
    :can-view-departments="permissions.view"
    :can-edit-departments="permissions.edit"
    :can-save-pool="canSavePool"
    :can-reset-resources="canResetResources"
    :load-has-gpu-limits="loadHasGpuLimits"
    :totals-complete="totalsComplete"
    :can-edit-department="canEditDepartment"
    :search-term="search"
    :rank-filter="rank"
    :quota-policy="quotaPolicy"
    :load-quota-context="loadQuotaContext"
    @refresh="refresh"
    :refresh-saved-scheduling="refreshSavedScheduling"
    @search="search = $event"
    @filter-rank="rank = $event"
    @leave-approved="leaveApproved = true"
    @saving="saving = $event"
    @pool-saved="pool = $event"
    @open-identity="openStep('identity')"
    @open-topology="openStep('topology')"
    @saved="finish"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import { NodePoolSchedulingStep } from "@/components/node-pools/node-pool-scheduling-step";
import { NodePoolWizardLoadState } from "@/components/node-pools/node-pool-wizard-load-state";
import { useNodePoolSchedulingData } from "@/composables/node-pool/use-node-pool-scheduling-data.composable";
import { useNodePoolWizardNavigation } from "@/composables/node-pool/use-node-pool-wizard-navigation.composable";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

const route = useRoute();
const permissionStore = usePermissionStore();
const canSavePool = computed(() => permissionStore.hasPermission(ResourceType.Nodepools, Action.Update));
const poolId = computed(() => (typeof route.params.id === "string" ? route.params.id : ""));
const {
  snapshot: pool,
  baselineRows,
  departmentPages,
  aggregates,
  rankCounts,
  resources,
  permissions,
  totalsComplete,
  canEditDepartment,
  canResetResources,
  loadHasGpuLimits,
  quotaPolicy,
  policyError,
  loadError,
  loading,
  search,
  rank,
  refresh,
  loadChildQuotaSum,
} = useNodePoolSchedulingData(poolId);
const { saving, leaveApproved, openStep, finish } = useNodePoolWizardNavigation(pool);
async function refreshSavedScheduling() {
  await refresh();
  return !!pool.value && !loadError.value;
}
const ranks = computed(() => ENTITY_PRIORITY_SELECT_OPTIONS.map((option) => option.value).reverse());
async function loadQuotaContext(id: string) {
  if (!pool.value || quotaPolicy.value === undefined)
    throw new Error(policyError.value || "Quota policy is unavailable.");
  const childQuotaSum = await loadChildQuotaSum(id);
  return { childQuotaSum, enforceChildQuota: quotaPolicy.value };
}
</script>
