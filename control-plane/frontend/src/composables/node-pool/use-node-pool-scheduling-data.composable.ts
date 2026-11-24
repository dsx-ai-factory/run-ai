/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { computed, getCurrentScope, onScopeDispose, ref, type Ref } from "vue";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { useAuthStore } from "@/stores/auth.store/index";
import type { PermittedScopes } from "@/swagger-models/authorization-client/index";
import type { IDepartmentPages, ISchedulingTotals } from "@/models/node-pool-scheduling.model";
import { usePermissionStore } from "@/stores/permissions.store/index";
import { useSettingStore } from "@/stores/setting.store/index";
import { orgUnitSettingsService } from "@/services/control-plane/org-unit-settings.service/org-unit-settings.service";
import { Action, ResourceType } from "@/swagger-models/authorization-client/index";
import type { Nodepool } from "@/swagger-models/cluster-service-client/index";
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client/index";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";
import { useNodePoolWizardData } from "./use-node-pool-wizard-data.composable";
import {
  loadChildQuotaSum,
  loadNodePoolGpuLimits,
  loadSchedulingDepartmentPage,
  schedulingSnapshotTotals,
  loadSchedulingTotals,
  schedulingFilters,
} from "@/composables/node-pool/node-pool-scheduling-data.util";
import { extractErrorMessages } from "@/utils/error-alert.util/index";

export function useNodePoolSchedulingData(poolId: Ref<string>) {
  const permissionStore = usePermissionStore();
  const settingStore = useSettingStore();
  const authStore = useAuthStore();
  const projectScopes = ref<Partial<Record<Action, PermittedScopes>>>({});
  const canResetResources = computed(
    () => coversCluster(departmentScopes.value[Action.Update]) && coversCluster(projectScopes.value[Action.Update]),
  );
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
  const baselineRows = ref<DepartmentNodePoolResources[]>([]);
  const aggregates = ref<Record<string, ISchedulingTotals>>({});
  const rankCounts = ref<Record<string, number>>({});
  const quotaPolicy = ref<boolean>();
  const policyError = ref("");
  const search = ref("");
  const rank = ref<string>();
  const nextOffset = ref<number>();
  const loadingAll = ref(false);
  let pendingPage: Promise<boolean> | undefined;
  let pendingAll: Promise<boolean> | undefined;
  let scopeVersion = 0;
  function applySnapshotTotals(pool: Nodepool) {
    if (!totalsComplete.value) return;
    const totals = schedulingSnapshotTotals(baselineRows.value, pool.schedulingCapabilities?.adjustableWeights === true);
    const empty = schedulingSnapshotTotals([], pool.schedulingCapabilities?.adjustableWeights === true).all;
    for (const [rank, aggregate] of Object.entries(aggregates.value)) {
      Object.assign(aggregate, totals[rank] ?? empty);
    }
  }
  function loadNext(): Promise<boolean> {
    if (pendingPage) return pendingPage;
    const pool = snapshot.value;
    const offset = nextOffset.value;
    if (!pool || offset == null) return Promise.resolve(true);
    const current = scopeVersion;
    pendingPage = loadSchedulingDepartmentPage({ poolId: pool.id, clusterId: pool.clusterId }, offset)
      .then((page) => {
        if (current !== scopeVersion) return false;
        const byId = new Map(baselineRows.value.map((row) => [row.orgUnitId, row]));
        for (const row of page.nodePoolResources) byId.set(row.orgUnitId, row);
        baselineRows.value = [...byId.values()];
        nextOffset.value = page.next;
        if (page.next == null) applySnapshotTotals(pool);
        return true;
      })
      .catch((error) => {
        if (current === scopeVersion)
          loadError.value = extractErrorMessages(error, "Failed to load departments.").join(" ");
        return false;
      })
      .finally(() => {
        if (current === scopeVersion) pendingPage = undefined;
      });
    return pendingPage;
  }
  function loadAll(): Promise<boolean> {
    if (pendingAll) return pendingAll;
    if (nextOffset.value == null) return Promise.resolve(true);
    const current = scopeVersion;
    loadingAll.value = true;
    pendingAll = (async () => {
      while (current === scopeVersion && nextOffset.value != null) {
        if (!(await loadNext())) return false;
      }
      return current === scopeVersion;
    })().finally(() => {
      if (current === scopeVersion) {
        loadingAll.value = false;
        pendingAll = undefined;
      }
    });
    return pendingAll;
  }
  const departmentPages = computed<IDepartmentPages>(() => ({
    nextOffset: nextOffset.value,
    loadingAll: loadingAll.value,
    loadNext,
    loadAll,
  }));
  function resetScheduling(poolChanged: boolean) {
    scopeVersion++;
    nextOffset.value = undefined;
    pendingPage = undefined;
    pendingAll = undefined;
    loadingAll.value = false;
    if (poolChanged) {
      search.value = "";
      rank.value = undefined;
    }
    baselineRows.value = [];
    aggregates.value = {};
    rankCounts.value = {};
    quotaPolicy.value = undefined;
    departmentScopes.value = {};
    projectScopes.value = {};
    policyError.value = "";
  }
  async function loadScheduling(pool: Nodepool) {
    const current = scopeVersion;
    const scope = { poolId: pool.id, clusterId: pool.clusterId };
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
      const [total, page, scopes, projects] = await Promise.all([
        loadSchedulingTotals(scope),
        loadSchedulingDepartmentPage(scope, 0),
        permissionService.getPermittedScopesForAction(ResourceType.Department),
        permissionService.getPermittedScopesForAction(ResourceType.Project).catch(() => ({})),
      ]);
      if (scopeVersion !== current) return;
      departmentScopes.value = scopes;
      projectScopes.value = projects;
      const rawWeightsAvailable = pool.schedulingCapabilities?.adjustableWeights === true;
      const ranks = ENTITY_PRIORITY_SELECT_OPTIONS.map((option) => option.value);
      const rowsComplete = page.next == null && coversCluster(scopes[Action.Read]);
      const snapshotTotals = rowsComplete
        ? schedulingSnapshotTotals(page.nodePoolResources, rawWeightsAvailable)
        : undefined;
      const visibleCounts: Record<string, number> = {};
      for (const row of page.nodePoolResources) visibleCounts[row.rank] = (visibleCounts[row.rank] ?? 0) + 1;
      const perRank = await Promise.all(
        ranks.map(async (rank) => {
          if (snapshotTotals) {
            const totals = snapshotTotals[rank];
            return {
              count: visibleCounts[rank] ?? 0,
              totals: {
                deservedSum: totals?.deservedSum ?? 0,
                weightsSum: (rawWeightsAvailable ? totals?.rawWeightsSum : totals?.automaticWeightsSum) ?? 0,
              },
            };
          }
          const rankScope = { ...scope, rank };
          const count = await orgUnitService.countDepartmentNodePoolResources(schedulingFilters(rankScope));
          const totals = count ? await loadSchedulingTotals(rankScope) : { weightsSum: 0, deservedSum: 0 };
          return { count, totals };
        }),
      );
      if (scopeVersion !== current) return;
      // The server sums weights using the saved capability. Keep the other denominator unknown
      // until a complete peer snapshot or a finite quota sum supplies it.
      function serverTotals(value: typeof total): ISchedulingTotals {
        return {
          ...value,
          rawWeightsSum: rawWeightsAvailable ? value.weightsSum : undefined,
          automaticWeightsSum: rawWeightsAvailable
            ? value.deservedSum >= 0
              ? value.deservedSum
              : undefined
            : value.weightsSum,
          finiteQuotaSum: value.deservedSum >= 0 ? value.deservedSum : undefined,
          unlimitedQuotaCount: value.deservedSum >= 0 ? 0 : undefined,
        };
      }
      baselineRows.value = page.nodePoolResources;
      aggregates.value = Object.fromEntries([
        ["all", serverTotals(total)],
        ...ranks.map((rank, i) => [rank, serverTotals(perRank[i].totals)]),
      ]);
      rankCounts.value = Object.fromEntries(ranks.map((rank, i) => [rank, perRank[i].count]));
      if (page.next == null) applySnapshotTotals(pool);
      nextOffset.value = page.next;
    };
    try {
      await Promise.all([policy(), departments()]);
    } catch (error) {
      if (scopeVersion === current)
        loadError.value = extractErrorMessages(error, "Failed to load scheduling data.").join(" ");
    }
  }
  const {
    pool: snapshot,
    resources,
    loading,
    loadError,
    refresh,
  } = useNodePoolWizardData(poolId, { retryCapacity: true, resetStep: resetScheduling, loadStep: loadScheduling });
  if (getCurrentScope())
    onScopeDispose(() => {
      scopeVersion++;
    });
  return {
    snapshot,
    loadHasGpuLimits: async () => {
      const pool = snapshot.value;
      const version = scopeVersion;
      if (!pool) throw new Error("Node pool is unavailable.");
      const hasLimits = await loadNodePoolGpuLimits({ poolId: pool.id, clusterId: pool.clusterId });
      if (version !== scopeVersion) throw new Error("Node pool changed. Try again.");
      return hasLimits;
    },
    loadChildQuotaSum: async (id: string) => {
      if (!snapshot.value) throw new Error("Node pool is unavailable.");
      return loadChildQuotaSum({ poolId: snapshot.value.id, clusterId: snapshot.value.clusterId }, id);
    },
    baselineRows,
    departmentPages,
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
    loading,
    search,
    rank,
    refresh,
  };
}
