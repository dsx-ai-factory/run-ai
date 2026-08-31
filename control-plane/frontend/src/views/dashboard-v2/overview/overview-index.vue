<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-page>
    <dashboard-page-filter
      :filter-options="filterOptions"
      :async-filter-options-promise-map="asyncFilterOptionsPromiseMap"
      @filter-changed="onFilterChanged"
      :filter-by="filterBy"
    />
    <widget-layout
      :widgets-layout="widgetsLayout"
      :cluster-id="clusterId"
      :node-pool-name="nodePoolName"
      :filter-by-dates="filterByDates"
      :filter-by="filterBy"
      :has-departments="hasDepartments"
      :has-read-projects-permission="hasReadProjectsPermission"
      :has-read-departments-permission="hasReadDepartmentsPermission"
      :has-read-clusters-permission="hasReadClustersPermission"
      @date-changed="onDateChanged"
    />
  </q-page>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
// components
import { DashboardPageFilter } from "@/components/dashboard-v2/dashboard-page-filter/";
import { WidgetLayout } from "@/components/dashboard-v2/widgets/common/widget-layout/";
// models
import { type IFilterBy, type IFilterModel, type IFilterOption } from "@/models/filter.model";
import type { ISelectOption } from "@/models/global.model";
import { EWidgetComponent } from "@/models/chart.model";
import { PREDEFINED_DATES } from "@/models/date.model";
// stores
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";
import { useGuidedTourStore } from "@/stores/guided-tour.store";
import { usePermissionStore } from "@/stores/permissions.store";
// utils
import { dateUtil, type IRangeDates } from "@/utils/date.util";
import { storageUtil } from "@/utils/storage.util";
// constants
import { OVERVIEW_DATE_FILTERS, OVERVIEW_FILTERS } from "@/common/storage.constant";
import { STORAGE_KEYS } from "@/models/storage.model";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { type Nodepool, NodepoolSortFilterFields } from "@/swagger-models/cluster-service-client";
import { ScopeType } from "@/swagger-models/org-unit-service-client";

// Widget structure constants
const overviewStructure: EWidgetComponent[][] = [
  [
    EWidgetComponent.ReadyNodesWidget,
    EWidgetComponent.ReadyGpuWidget,
    EWidgetComponent.AllocatedGpuWidget,
    EWidgetComponent.IdleAllocatedGpuWidget,
    EWidgetComponent.RunningWorkloads,
    EWidgetComponent.PendingWorkloads,
    EWidgetComponent.FailedWorkloads,
  ],
  [EWidgetComponent.WorkloadByTypeWidget, EWidgetComponent.WorkloadStatus],
  [EWidgetComponent.OrgUnitRankingWidget],
  [EWidgetComponent.WorkloadPendingTimeDistributionWidget, EWidgetComponent.IdleGpuWorkloadWidget],
  [EWidgetComponent.GpuAllocationRatioWidget, EWidgetComponent.NodePoolFreeResourcesWidget],
  [EWidgetComponent.OverviewTimeRangeWidgets],
  [EWidgetComponent.WorkloadByCategoryTimeRangeWidget],
];

// Widgets that require READ Nodes permission
const nodeWidgets: EWidgetComponent[] = [
  EWidgetComponent.ReadyNodesWidget,
  EWidgetComponent.ReadyGpuWidget,
  EWidgetComponent.NodePoolFreeResourcesWidget,
];

// Overtime widgets - require READ Clusters (or READ NodePools + filter) OR org-unit permissions
const overtimeWidgets: EWidgetComponent[] = [
  EWidgetComponent.OverviewTimeRangeWidgets,
  EWidgetComponent.WorkloadByCategoryTimeRangeWidget,
];

// Widgets that require READ Nodes OR READ Workloads permission (either one)
const nodesOrWorkloadsWidgets: EWidgetComponent[] = [
  EWidgetComponent.AllocatedGpuWidget,
  EWidgetComponent.IdleAllocatedGpuWidget,
];

// Widgets that require READ Workloads permission
const workloadWidgets: EWidgetComponent[] = [
  EWidgetComponent.IdleGpuWorkloadWidget,
  EWidgetComponent.RunningWorkloads,
  EWidgetComponent.PendingWorkloads,
  EWidgetComponent.FailedWorkloads,
  EWidgetComponent.WorkloadByTypeWidget,
  EWidgetComponent.WorkloadStatus,
  EWidgetComponent.WorkloadPendingTimeDistributionWidget,
];

// Widgets that require READ Departments permission
const departmentWidgets: EWidgetComponent[] = [EWidgetComponent.GpuAllocationRatioWidget];

// Widgets that require READ Projects OR READ Departments permission
const orgUnitWidgets: EWidgetComponent[] = [EWidgetComponent.OrgUnitRankingWidget];

// Stores
const appStore = useAppStore();
const clusterStore = useClusterStore();
const authStore = useAuthStore();
const guidedTourStore = useGuidedTourStore();
const permissionStore = usePermissionStore();

// Reactive state
const nodepools = ref<Nodepool[]>([]);
const filterByDates = ref<IRangeDates>({
  dateStart: dateUtil.dateSevenDaysAgo(),
  dateEnd: new Date(),
  label: PREDEFINED_DATES.LAST_7_DAYS,
});
const filterBy = ref<IFilterBy>({} as IFilterBy);

// Permission computed properties
const hasReadNodesPermission = computed(() => permissionStore.hasReadNodesPermission());
const hasReadClustersPermission = computed(() => permissionStore.hasReadClustersPermission());
const hasReadNodePoolsPermission = computed(() => permissionStore.hasReadNodePoolsPermission());
const hasReadWorkloadsPermission = computed(() => permissionStore.hasReadWorkloadsPermission());
const hasReadDepartmentsPermission = computed(() => permissionStore.hasReadDepartmentsPermission());
const hasReadProjectsPermission = computed(() => permissionStore.hasReadProjectsPermission());

// Computed properties
const clusterId = computed(() => clusterStore.selectedClusterId);

const nodePoolName = computed(() => {
  return filterBy.value?.columnFilters?.find((filter: IFilterModel) => filter.field === "nodepool")?.term;
});

const hasDepartments = computed(() => {
  const departments = authStore.orgUnits.filter(
    (orgUnit) => orgUnit.type === ScopeType.Department && orgUnit.name !== "default",
  );
  return departments.length > 0;
});

const filterOptions = computed<IFilterOption[]>(() => {
  const options: IFilterOption[] = [];

  if (hasReadNodePoolsPermission.value) {
    options.push({
      field: "nodepool",
      name: "Node pool",
      label: "Node pool",
    });
  }

  return options;
});

const widgetsLayout = computed<EWidgetComponent[][]>(() => {
  return overviewStructure
    .map((row) =>
      row.filter((widget) => {
        // Hide node widgets if no READ Nodes permission
        if (!hasReadNodesPermission.value && nodeWidgets.includes(widget)) {
          return false;
        }

        // Hide workload widgets if no READ Workloads permission
        if (!hasReadWorkloadsPermission.value && workloadWidgets.includes(widget)) {
          return false;
        }

        // Hide department widgets if no READ Departments permission
        if (!hasReadDepartmentsPermission.value && departmentWidgets.includes(widget)) {
          return false;
        }

        // Hide org unit widgets if no READ Projects AND no READ Departments permission
        if (!hasReadProjectsPermission.value && !hasReadDepartmentsPermission.value && orgUnitWidgets.includes(widget)) {
          return false;
        }

        // Hide overtime widgets if no READ Clusters permission AND no (READ NodePools + filter) AND no org-unit permissions
        if (overtimeWidgets.includes(widget)) {
          const hasClusterAccess =
            hasReadClustersPermission.value || (hasReadNodePoolsPermission.value && nodePoolName.value);
          const hasOrgUnitAccess = hasReadDepartmentsPermission.value || hasReadProjectsPermission.value;
          if (!hasClusterAccess && !hasOrgUnitAccess) {
            return false;
          }
        }

        // Hide widgets that require either READ Nodes OR READ Workloads if user has neither
        if (
          !hasReadNodesPermission.value &&
          !hasReadWorkloadsPermission.value &&
          nodesOrWorkloadsWidgets.includes(widget)
        ) {
          return false;
        }

        return true;
      }),
    )
    .filter((row) => row.length > 0);
});

// Async filter options map
const asyncFilterOptionsPromiseMap = {
  nodepool: getFilteredNodePools,
} as unknown as Record<string, (searchQuery: string) => Promise<Array<ISelectOption>>>;

// Methods
function loadFilters(): void {
  filterBy.value = storageUtil.get<IFilterBy | null>(OVERVIEW_FILTERS) ?? {
    clusterUuid: clusterId.value,
    columnFilters: [],
  };

  loadDateFilters();
}

function loadDateFilters(): void {
  const dateFilters = storageUtil.get<{ startDate: Date; endDate: Date; label: string } | null>(OVERVIEW_DATE_FILTERS);

  if (dateFilters !== null) {
    filterByDates.value = dateUtil.getUpdatedTime({
      dateStart: new Date(dateFilters.startDate),
      dateEnd: new Date(dateFilters.endDate),
      label: dateFilters.label,
    });
  }
}

function onFilterChanged(newFilterBy: IFilterBy): void {
  storageUtil.save(OVERVIEW_FILTERS, newFilterBy);
  filterBy.value = { ...newFilterBy };
}

function onDateChanged(data: { startDate: Date; endDate: Date; label: string }): void {
  storageUtil.save(OVERVIEW_DATE_FILTERS, data);

  filterByDates.value = {
    dateStart: data.startDate,
    dateEnd: data.endDate,
    label: data.label,
  };
}

async function getFilteredNodePools(searchQuery: string): Promise<string[]> {
  if (nodepools.value.length === 0) {
    await fetchNodePools();
  }

  return nodepools.value.map((nodePool: Nodepool) => nodePool.name).filter((name: string) => name.includes(searchQuery));
}

async function fetchNodePools(): Promise<void> {
  try {
    nodepools.value = await clusterService.getNodepools([
      filterUtil.getEqualsFilterString(NodepoolSortFilterFields.ClusterId, clusterId.value),
    ]);
  } catch (error) {
    console.error("Error fetching node pools:", error);
    nodepools.value = [];
  }
}

function triggerOverviewTourIfNeeded(): void {
  if (appStore.isWhiteLabel || !appStore.isOnboardingModalClosed) return;

  const seenTourAlready = storageUtil.get(STORAGE_KEYS.OVERVIEW_TOUR_SEEN);

  if (seenTourAlready) return;

  guidedTourStore.openOverviewGuidedTour();

  storageUtil.save(STORAGE_KEYS.OVERVIEW_TOUR_SEEN, true);
}

// Lifecycle - equivalent to created()
loadFilters();
appStore.setPageLoading(false);

// Lifecycle - equivalent to mounted()
onMounted(() => {
  triggerOverviewTourIfNeeded();
});

// Expose for testing
defineExpose({
  widgetsLayout,
  filterOptions,
});
</script>

<style scoped lang="scss"></style>
