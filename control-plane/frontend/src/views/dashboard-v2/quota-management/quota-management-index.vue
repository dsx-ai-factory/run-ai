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
      :has-departments="hasDepartments"
      :has-read-projects-permission="hasReadProjectsPermission"
      :has-read-departments-permission="hasReadDepartmentsPermission"
      :has-read-clusters-permission="hasReadClustersPermission"
      :org-unit-metric-source="orgUnitMetricSource"
    />
  </q-page>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { DashboardPageFilter } from "@/components/dashboard-v2/dashboard-page-filter/";
import { WidgetLayout } from "@/components/dashboard-v2/widgets/common/widget-layout/";
import { type IFilterBy, type IFilterModel, type IFilterOption } from "@/models/filter.model";
import type { ISelectOption } from "@/models/global.model";
import { EWidgetComponent } from "@/models/chart.model";
import { type NodepoolMinimal, NodepoolSortFilterFields } from "@/swagger-models/cluster-service-client";
import { ScopeType, type Department } from "@/swagger-models/org-unit-service-client";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { storageUtil } from "@/utils/storage.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { OVERVIEW_FILTERS } from "@/common/storage.constant";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

const quotaStructure: EWidgetComponent[][] = [
  [EWidgetComponent.QuotaTotalGpuWidget, EWidgetComponent.AllocatedQuotaGpuWidget, EWidgetComponent.PendingWorkloads],
  [
    EWidgetComponent.QuotaByNodepoolWidget,
    EWidgetComponent.AllocationByNodepoolWidget,
    EWidgetComponent.PendingWorkloadsByNodepoolWidget,
  ],
  [EWidgetComponent.DeptAllocationRatioWidget, EWidgetComponent.ProjectAllocationRatioWidget],
  [EWidgetComponent.QuotaAllocationTimeRangeWidget],
];

const widgetPermissions = new Map<EWidgetComponent, () => boolean>([
  [EWidgetComponent.QuotaTotalGpuWidget, () => hasReadProjectsPermission.value && hasReadNodesPermission.value],
  [EWidgetComponent.AllocatedQuotaGpuWidget, () => hasReadProjectsPermission.value],
  [EWidgetComponent.PendingWorkloads, () => hasReadWorkloadsPermission.value],
  [EWidgetComponent.QuotaAllocationTimeRangeWidget, () => hasReadClustersPermission.value],
  [EWidgetComponent.PendingWorkloadsByNodepoolWidget, () => hasReadWorkloadsPermission.value],
  [EWidgetComponent.QuotaByNodepoolWidget, () => hasReadProjectsPermission.value && hasReadNodePoolsPermission.value],
  [
    EWidgetComponent.AllocationByNodepoolWidget,
    () => hasReadProjectsPermission.value && hasReadNodePoolsPermission.value,
  ],
  [EWidgetComponent.DeptAllocationRatioWidget, () => hasReadDepartmentsPermission.value && hasDepartments.value],
  [EWidgetComponent.ProjectAllocationRatioWidget, () => hasReadProjectsPermission.value],
]);

const appStore = useAppStore();
const clusterStore = useClusterStore();
const authStore = useAuthStore();
const permissionStore = usePermissionStore();

const nodepools = ref<NodepoolMinimal[]>([]);
const departmentsCache = ref<Department[]>([]);
const filterBy = ref<IFilterBy>({} as IFilterBy);

const hasReadClustersPermission = computed(() => permissionStore.hasReadClustersPermission());
const hasReadNodesPermission = computed(() => permissionStore.hasReadNodesPermission());
const hasReadNodePoolsPermission = computed(() => permissionStore.hasReadNodePoolsPermission());
const hasReadWorkloadsPermission = computed(() => permissionStore.hasReadWorkloadsPermission());
const hasReadDepartmentsPermission = computed(() => permissionStore.hasReadDepartmentsPermission());
const hasReadProjectsPermission = computed(() => permissionStore.hasReadProjectsPermission());

const clusterId = computed(() => clusterStore.selectedClusterId);

const nodePoolName = computed(() => {
  return filterBy.value?.columnFilters?.find((filter: IFilterModel) => filter.field === "nodepool")?.term;
});

const selectedDepartmentName = computed(() => {
  return filterBy.value?.columnFilters?.find((filter: IFilterModel) => filter.field === "department")?.term;
});

const selectedProjectName = computed(() => {
  return filterBy.value?.columnFilters?.find((filter: IFilterModel) => filter.field === "project")?.term;
});

const hasDepartments = computed(() => {
  return authStore.orgUnits.some((orgUnit) => orgUnit.type === ScopeType.Department && orgUnit.name !== "default");
});

const orgUnitMetricSource = computed<IOrgUnitMetricSource | null>(() => {
  if (selectedProjectName.value) {
    const project = authStore.orgUnits.find(
      (ou) => ou.type === ScopeType.Project && ou.name === selectedProjectName.value,
    );
    if (project) return { type: EOrgUnitMetricSourceType.Project, id: project.id };
  }
  if (selectedDepartmentName.value) {
    const dept = authStore.orgUnits.find(
      (ou) => ou.type === ScopeType.Department && ou.name === selectedDepartmentName.value,
    );
    if (dept) return { type: EOrgUnitMetricSourceType.Department, id: dept.id };
  }
  return null;
});

const filterOptions = computed<IFilterOption[]>(() => {
  const options: IFilterOption[] = [];
  if (hasReadNodePoolsPermission.value) {
    options.push({ field: "nodepool", name: "Node pool", label: "Node pool" });
  }
  if (hasReadDepartmentsPermission.value && hasDepartments.value) {
    options.push({ field: "department", name: "Department", label: "Department" });
  }
  if (hasReadProjectsPermission.value) {
    options.push({ field: "project", name: "Project", label: "Project" });
  }
  return options;
});

const widgetsLayout = computed<EWidgetComponent[][]>(() => {
  return quotaStructure
    .map((row) => row.filter((widget) => widgetPermissions.get(widget)?.() ?? true))
    .filter((row) => row.length > 0);
});

const asyncFilterOptionsPromiseMap = {
  nodepool: getFilteredNodePools,
  department: getFilteredDepartments,
  project: getFilteredProjects,
} as unknown as Record<string, (searchQuery: string) => Promise<Array<ISelectOption>>>;

function loadFilters(): void {
  filterBy.value = storageUtil.get<IFilterBy | null>(OVERVIEW_FILTERS) ?? {
    clusterUuid: clusterId.value,
    columnFilters: [],
  };
}

function onFilterChanged(newFilterBy: IFilterBy): void {
  storageUtil.save(OVERVIEW_FILTERS, newFilterBy);
  filterBy.value = { ...newFilterBy };
}

async function getFilteredNodePools(searchQuery: string): Promise<string[]> {
  if (nodepools.value.length === 0) {
    await fetchNodePools();
  }
  return nodepools.value
    .map((nodePool: NodepoolMinimal) => nodePool.name)
    .filter((name: string) => name.includes(searchQuery));
}

async function getFilteredDepartments(searchQuery: string): Promise<string[]> {
  if (departmentsCache.value.length === 0) {
    departmentsCache.value = await orgUnitService.getDepartments();
  }
  return departmentsCache.value
    .map((d) => d.name)
    .filter((name) => (searchQuery ? name.toLowerCase().includes(searchQuery.toLowerCase()) : true));
}

async function getFilteredProjects(searchQuery: string): Promise<string[]> {
  const projects = await orgUnitService.getProjects(
    undefined,
    undefined,
    undefined,
    undefined,
    undefined,
    searchQuery || undefined,
  );
  return projects.map((p) => p.name);
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

loadFilters();
appStore.setPageLoading(false);

defineExpose({
  widgetsLayout,
  filterOptions,
});
</script>

<style scoped lang="scss"></style>
