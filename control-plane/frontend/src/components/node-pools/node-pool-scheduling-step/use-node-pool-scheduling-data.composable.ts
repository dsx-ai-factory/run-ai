/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { computed, ref, watch, type Ref } from "vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import type { DepartmentNodePoolResources, DepartmentSumResponse } from "@/swagger-models/org-unit-service-client";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";
import type { INodePoolSummaryResources } from "../node-pool-wizard-summary/node-pool-wizard-summary.model";
import {
  loadSchedulingPool,
  loadDepartmentPage,
  loadSchedulingTotals,
  loadSchedulingResources,
  schedulingFilters,
} from "./node-pool-scheduling.service";
import { extractErrorMessages } from "@/utils/error-alert.util";

export function useNodePoolSchedulingData(poolId: Ref<string>) {
  const permissionStore = usePermissionStore();
  const permissions = computed(() => ({
    view: permissionStore.hasPermission(ResourceType.Department, Action.Read),
    edit: permissionStore.hasPermission(ResourceType.Department, Action.Update),
  }));
  const snapshot = ref<Nodepool>();
  const pages = ref<DepartmentNodePoolResources[]>([]);
  const aggregates = ref<Record<string, DepartmentSumResponse>>({});
  const resources = ref<INodePoolSummaryResources>();
  const loadError = ref("");
  const telemetryError = ref("");
  const loading = ref(false);
  const search = ref("");
  const rank = ref<string>();
  const count = ref(0);
  let version = 0;
  async function loadRows(append = false) {
    const pool = snapshot.value;
    if (!pool || !permissions.value.view) return;
    const current = ++version;
    loading.value = true;
    loadError.value = "";
    const scope = { poolId: pool.id, clusterId: pool.clusterId, rank: rank.value };
    try {
      const [rows, total] = await Promise.all([
        loadDepartmentPage({ ...scope, search: search.value, offset: append ? pages.value.length : 0, limit: 50 }),
        orgUnitService.countDepartmentNodePoolResources(schedulingFilters(scope), search.value),
      ]);
      if (version !== current) return;
      pages.value = append ? [...pages.value, ...rows] : rows;
      count.value = total;
    } catch (error) {
      if (version === current) loadError.value = extractErrorMessages(error, "Failed to load departments.").join(" ");
    } finally {
      if (version === current) loading.value = false;
    }
  }
  async function refresh() {
    const current = ++version;
    const id = poolId.value;
    if (snapshot.value?.id !== id) snapshot.value = undefined;
    pages.value = [];
    aggregates.value = {};
    resources.value = undefined;
    loadError.value = "";
    telemetryError.value = "";
    loading.value = true;
    try {
      const pool = await loadSchedulingPool(id);
      if (version !== current) return;
      snapshot.value = pool;
      if (!permissions.value.view) {
        loading.value = false;
        return;
      }
      const scope = { poolId: id, clusterId: pool.clusterId };
      const totals = await Promise.all([
        loadSchedulingTotals(scope),
        ...ENTITY_PRIORITY_SELECT_OPTIONS.map((option) => loadSchedulingTotals({ ...scope, rank: option.value })),
      ]);
      if (version !== current) return;
      aggregates.value = Object.fromEntries([
        ["all", totals[0]],
        ...ENTITY_PRIORITY_SELECT_OPTIONS.map((option, i) => [option.value, totals[i + 1]]),
      ]);
      try {
        const result = await loadSchedulingResources(pool);
        if (version === current) resources.value = result;
      } catch (error) {
        if (version === current)
          telemetryError.value = extractErrorMessages(error, "Capacity is unavailable.").join(" ");
      }
      if (version === current) await loadRows();
    } catch (error) {
      if (version === current) {
        loadError.value = extractErrorMessages(error, "Failed to load scheduling data.").join(" ");
        loading.value = false;
      }
    }
  }
  watch(poolId, refresh, { immediate: true });
  watch([search, rank], () => loadRows());
  return {
    snapshot,
    pages,
    aggregates,
    resources,
    permissions,
    loadError,
    telemetryError,
    loading,
    search,
    rank,
    count,
    refresh,
    loadMore: () => loadRows(true),
  };
}
