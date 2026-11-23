/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { computed, getCurrentScope, onScopeDispose, ref, type Ref } from "vue";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { useAuthStore } from "@/stores/auth.store/index";
import type { PermittedScopes } from "@/swagger-models/authorization-client/index";
import type { ISchedulingTotals } from "@/models/node-pool-scheduling.model";
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
  loadSchedulingDepartmentSnapshot,
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
  let scopeVersion = 0;
  function resetScheduling(poolChanged: boolean) {
    scopeVersion++;
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
      const [total, departmentSnapshot, scopes, projects] = await Promise.all([
        loadSchedulingTotals(scope),
        loadSchedulingDepartmentSnapshot(scope, pool.schedulingCapabilities?.adjustableWeights === true),
        permissionService.getPermittedScopesForAction(ResourceType.Department),
        permissionService.getPermittedScopesForAction(ResourceType.Project).catch(() => ({})),
      ]);
      if (scopeVersion !== current) return;
      departmentScopes.value = scopes;
      projectScopes.value = projects;
      const complete = coversCluster(scopes[Action.Read]);
      const visibleCounts: Record<string, number> = {};
      for (const row of departmentSnapshot.rows) visibleCounts[row.rank] = (visibleCounts[row.rank] ?? 0) + 1;
      // Complete scoped snapshots already contain the rows behind every rank.
      // Retain authoritative rank reads when visibility is restricted.
      const [rankTotals, counts] = complete
        ? ([
            ENTITY_PRIORITY_SELECT_OPTIONS.map((option) => {
              const totals = departmentSnapshot.totals[option.value];
              return {
                weightsSum:
                  pool.schedulingCapabilities?.adjustableWeights === true
                    ? (totals?.rawWeightsSum ?? 0)
                    : (totals?.automaticWeightsSum ?? 0),
                deservedSum: totals?.deservedSum ?? 0,
              };
            }),
            ENTITY_PRIORITY_SELECT_OPTIONS.map((option) => visibleCounts[option.value] ?? 0),
          ] as const)
        : await Promise.all([
            Promise.all(
              ENTITY_PRIORITY_SELECT_OPTIONS.map((option) => loadSchedulingTotals({ ...scope, rank: option.value })),
            ),
            Promise.all(
              ENTITY_PRIORITY_SELECT_OPTIONS.map((option) =>
                orgUnitService.countDepartmentNodePoolResources(schedulingFilters({ ...scope, rank: option.value })),
              ),
            ),
          ]);
      if (scopeVersion !== current) return;
      const weightTotals = departmentSnapshot.totals;
      baselineRows.value = departmentSnapshot.rows;
      aggregates.value = Object.fromEntries([
        ["all", { ...total, ...weightTotals.all }],
        ...ENTITY_PRIORITY_SELECT_OPTIONS.map((option, i) => [
          option.value,
          {
            ...rankTotals[i],
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
    telemetryError,
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
  };
}
