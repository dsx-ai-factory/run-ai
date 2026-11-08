/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { computed, getCurrentScope, onScopeDispose, ref, watch, type Ref } from "vue";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { useAuthStore } from "@/stores/auth.store";
import type { PermittedScopes } from "@/swagger-models/authorization-client";
import type { SchedulingTotals } from "./node-pool-scheduling.model";
import { usePermissionStore } from "@/stores/permissions.store";
import { useSettingStore } from "@/stores/setting.store";
import { orgUnitSettingsService } from "@/services/control-plane/org-unit-settings.service/org-unit-settings.service";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";
import type { INodePoolSummaryResources } from "../node-pool-wizard-summary/node-pool-wizard-summary.model";
import {
  loadSchedulingDepartmentSnapshot,
  loadSchedulingPool,
  loadDepartmentPage,
  loadSchedulingTotals,
  loadSchedulingResources,
  schedulingFilters,
} from "./node-pool-scheduling.service";
import { extractErrorMessages } from "@/utils/error-alert.util";

export function useNodePoolSchedulingData(poolId: Ref<string>) {
  const permissionStore = usePermissionStore();
  const settingStore = useSettingStore();
  const authStore = useAuthStore();
  const departmentScopes = ref<Partial<Record<Action, PermittedScopes>>>({});
  function coversCluster(scopes?: PermittedScopes): boolean {
    return !!(
      scopes?.system ||
      (authStore.currentUser.tenantId != null && scopes?.tenants?.includes(String(authStore.currentUser.tenantId))) ||
      (snapshot.value && scopes?.clusters?.includes(snapshot.value.clusterId))
    );
  }
  const totalsComplete = computed(() => coversCluster(departmentScopes.value[Action.Read]));
  function canEditDepartment(id: string): boolean {
    const scopes = departmentScopes.value[Action.Update];
    return permissions.value.edit && (coversCluster(scopes) || !!scopes?.departments?.includes(id));
  }
  const permissions = computed(() => ({
    view: permissionStore.hasPermission(ResourceType.Department, Action.Read),
    edit: permissionStore.hasPermission(ResourceType.Department, Action.Update),
  }));
  const snapshot = ref<Nodepool>();
  const pages = ref<DepartmentNodePoolResources[]>([]);
  const baselineRows = ref<DepartmentNodePoolResources[]>([]);
  const aggregates = ref<Record<string, SchedulingTotals>>({});
  const rankCounts = ref<Record<string, number>>({});
  const resources = ref<INodePoolSummaryResources>();
  const quotaPolicy = ref<boolean>();
  const policyError = ref("");
  const loadError = ref("");
  const telemetryError = ref("");
  const poolLoading = ref(false);
  const rowsLoading = ref(false);
  const loading = computed(() => poolLoading.value || rowsLoading.value);
  const search = ref("");
  const rank = ref<string>();
  const count = ref(0);
  const loadingRanks = new Map<string, string>();
  const globalOffset = ref(0);
  const queryKey = computed(() => JSON.stringify([poolId.value, search.value, rank.value]));
  const completedQuery = ref<string>();
  const hasMore = computed(() => completedQuery.value === queryKey.value && globalOffset.value < count.value);
  let scopeVersion = 0;
  let rowsVersion = 0;
  let appending = false;
  async function loadRows(append = false) {
    const pool = snapshot.value;
    if (
      !pool ||
      !permissions.value.view ||
      (append && (appending || rowsLoading.value || completedQuery.value !== queryKey.value))
    )
      return;
    const requestQuery = queryKey.value;
    if (!append) {
      completedQuery.value = undefined;
      globalOffset.value = 0;
      count.value = 0;
      pages.value = [];
    }
    const current = ++rowsVersion;
    const scopeRequest = scopeVersion;
    if (append) appending = true;
    rowsLoading.value = true;
    loadError.value = "";
    const scope = { poolId: pool.id, clusterId: pool.clusterId, rank: rank.value };
    try {
      const [rows, total] = await Promise.all([
        loadDepartmentPage({ ...scope, search: search.value, offset: append ? globalOffset.value : 0, limit: 50 }),
        orgUnitService.countDepartmentNodePoolResources(schedulingFilters(scope), search.value),
      ]);
      if (rowsVersion !== current || scopeVersion !== scopeRequest) return;
      globalOffset.value = (append ? globalOffset.value : 0) + rows.length;
      completedQuery.value = requestQuery;
      pages.value = append
        ? [
            ...pages.value,
            ...rows.filter((row) => !pages.value.some((existing) => existing.orgUnitId === row.orgUnitId)),
          ]
        : rows;
      count.value = total;
    } catch (error) {
      if (rowsVersion === current && scopeVersion === scopeRequest)
        loadError.value = extractErrorMessages(error, "Failed to load departments.").join(" ");
    } finally {
      if (rowsVersion === current) {
        rowsLoading.value = false;
        appending = false;
      }
    }
  }
  async function loadRank(rankValue: string) {
    const pool = snapshot.value;
    const requestIdentity = `${scopeVersion}:${rowsVersion}`;
    if (
      !pool ||
      !permissions.value.view ||
      rowsLoading.value ||
      completedQuery.value !== queryKey.value ||
      loadingRanks.get(rankValue) === requestIdentity ||
      (rank.value && rank.value !== rankValue)
    )
      return;
    if (rank.value === rankValue) return loadRows(true);
    const scopeRequest = scopeVersion;
    const rowRequest = rowsVersion;
    const offset = pages.value.filter((row) => row.rank === rankValue).length;
    loadingRanks.set(rankValue, requestIdentity);
    try {
      const rows = await loadDepartmentPage({
        poolId: pool.id,
        clusterId: pool.clusterId,
        rank: rankValue,
        search: search.value,
        offset,
        limit: 50,
      });
      if (scopeVersion !== scopeRequest || rowsVersion !== rowRequest) return;
      pages.value = [
        ...pages.value,
        ...rows.filter((row) => !pages.value.some((existing) => existing.orgUnitId === row.orgUnitId)),
      ];
    } catch (error) {
      if (scopeVersion === scopeRequest && rowsVersion === rowRequest)
        loadError.value = extractErrorMessages(error, "Failed to load rank departments.").join(" ");
    } finally {
      if (loadingRanks.get(rankValue) === requestIdentity) loadingRanks.delete(rankValue);
    }
  }
  async function refresh() {
    completedQuery.value = undefined;
    rowsLoading.value = false;
    const current = ++scopeVersion;
    rowsVersion++;
    appending = false;
    loadingRanks.clear();
    const id = poolId.value;
    globalOffset.value = 0;
    if (snapshot.value?.id !== id) {
      snapshot.value = undefined;
      search.value = "";
      rank.value = undefined;
    }
    pages.value = [];
    aggregates.value = {};
    rankCounts.value = {};
    resources.value = undefined;
    quotaPolicy.value = undefined;
    departmentScopes.value = {};
    loadError.value = "";
    telemetryError.value = "";
    policyError.value = "";
    poolLoading.value = true;
    try {
      const pool = await loadSchedulingPool(id);
      if (scopeVersion !== current) return;
      snapshot.value = pool;
      const scope = { poolId: id, clusterId: pool.clusterId };
      const telemetry = async () => {
        try {
          const result = await loadSchedulingResources(pool);
          if (scopeVersion === current) resources.value = result;
        } catch (error) {
          if (scopeVersion === current)
            telemetryError.value = extractErrorMessages(error, "Capacity is unavailable.").join(" ");
        }
      };
      const policy = async () => {
        try {
          const result = await orgUnitSettingsService.getLimitQuotaOverSubscription();
          if (scopeVersion === current) {
            settingStore.limitQuotaOverSubscription = result;
            quotaPolicy.value = settingStore.isLimitQuotaOverSubscriptionEnabled;
          }
        } catch (error) {
          if (scopeVersion === current)
            policyError.value = extractErrorMessages(error, "Quota policy is unavailable.").join(" ");
        }
      };
      const departments = async () => {
        if (!permissions.value.view) return;
        const [totals, counts, departmentSnapshot, scopes] = await Promise.all([
          Promise.all([
            loadSchedulingTotals(scope),
            ...ENTITY_PRIORITY_SELECT_OPTIONS.map((option) => loadSchedulingTotals({ ...scope, rank: option.value })),
          ]),
          Promise.all(
            ENTITY_PRIORITY_SELECT_OPTIONS.map((option) =>
              orgUnitService.countDepartmentNodePoolResources(schedulingFilters({ ...scope, rank: option.value })),
            ),
          ),
          loadSchedulingDepartmentSnapshot(scope, pool.schedulingCapabilities?.adjustableWeights === true),
          permissionService.getPermittedScopesForAction(ResourceType.Department),
        ]);
        if (scopeVersion !== current) return;
        departmentScopes.value = scopes;
        const weightTotals = departmentSnapshot.totals;
        baselineRows.value = departmentSnapshot.rows;
        aggregates.value = Object.fromEntries([
          ["all", { ...totals[0], ...weightTotals.all }],
          ...ENTITY_PRIORITY_SELECT_OPTIONS.map((option, i) => [
            option.value,
            {
              ...totals[i + 1],
              ...(weightTotals[option.value] ?? {
                rawWeightsSum: pool.schedulingCapabilities?.adjustableWeights === true ? 0 : undefined,
                automaticWeightsSum: 0,
                finiteQuotaSum: 0,
                unlimitedQuotaCount: 0,
                deservedSum: 0,
              }),
            },
          ]),
        ]);
        rankCounts.value = Object.fromEntries(
          ENTITY_PRIORITY_SELECT_OPTIONS.map((option, i) => [option.value, counts[i]]),
        );
        await loadRows();
      };
      await Promise.all([telemetry(), policy(), departments()]);
    } catch (error) {
      if (scopeVersion === current)
        loadError.value = extractErrorMessages(error, "Failed to load scheduling data.").join(" ");
    } finally {
      if (scopeVersion === current) poolLoading.value = false;
    }
  }
  watch(poolId, refresh, { immediate: true });
  watch([search, rank], () => {
    appending = false;
    void loadRows();
  });
  if (getCurrentScope())
    onScopeDispose(() => {
      scopeVersion++;
      rowsVersion++;
    });
  return {
    snapshot,
    pages,
    baselineRows,
    hasMore,
    aggregates,
    rankCounts,
    resources,
    permissions,
    totalsComplete,
    canEditDepartment,
    quotaPolicy,
    policyError,
    loadError,
    telemetryError,
    loading,
    search,
    rank,
    count,
    refresh,
    loadRank,
    loadMore: () => loadRows(true),
  };
}
