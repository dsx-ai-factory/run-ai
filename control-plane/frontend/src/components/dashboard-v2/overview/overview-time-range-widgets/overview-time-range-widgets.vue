<template>
  <q-card>
    <div class="widgets-content-wrapper">
      <div class="filters-row element-bottom-margin">
        <template v-if="showOrgUnitFilters">
          <runai-autocomplete-filter-input
            v-if="showDepartmentFilter"
            aid="overtime-department-filter"
            :model-value="selectedDepartment"
            :display-value="selectedDepartmentName"
            @update:model-value="handleDepartmentChange"
            :search-options-fn="departmentSearchFn"
            label="Department"
            dense
            class="org-unit-filter"
          />
          <runai-autocomplete-filter-input
            v-if="showProjectFilter"
            :key="selectedDepartment ?? ''"
            aid="overtime-project-filter"
            :model-value="selectedProject"
            :display-value="selectedProjectName"
            @update:model-value="handleProjectChange"
            :search-options-fn="projectSearchFn"
            label="Project"
            dense
            class="org-unit-filter"
          />
        </template>
        <runai-datepicker-with-predefined
          aid="overtime-widget-date-picker"
          class="date-picker-filter"
          @changed="updateDateRange"
          :range-of-dates="rangeOfDatesCallback"
          info-bar-message="Select a period of up to 30 days within the last 90 days"
          :predefined-options="predefinedOptions"
          :default-option-index="5"
          :default-end-date="filterByDates?.dateEnd"
          :default-start-date="filterByDates?.dateStart"
          :default-option-label="filterByDates?.label"
        />
      </div>
      <template v-if="isReady">
        <resource-utilization-time-range-widget
          v-if="filterByDates"
          class="element-bottom-margin"
          flat
          :filter-by-dates="filterByDates"
          :cluster-id="clusterId"
          :node-pool-name="nodePoolName"
          :org-unit-metric-source="orgUnitMetricSource"
          :zoom-active="connectedZoomActive"
          @register="registerChart(chartIds[1])"
          @unregister="unregisterChart(chartIds[1])"
          @on-zoom="handleZoomIn"
          @on-reset-zoom="handleResetZoom"
          @on-zoom-toggle="handleZoomToggle"
        />
        <resource-allocation-time-range-widget
          v-if="filterByDates"
          class="element-bottom-margin"
          flat
          :filter-by-dates="filterByDates"
          :cluster-id="clusterId"
          :node-pool-name="nodePoolName"
          :org-unit-metric-source="orgUnitMetricSource"
          :zoom-active="connectedZoomActive"
          @register="registerChart(chartIds[0])"
          @unregister="unregisterChart(chartIds[0])"
          @on-zoom="handleZoomIn"
          @on-reset-zoom="handleResetZoom"
          @on-zoom-toggle="handleZoomToggle"
        />
        <workload-by-category-time-range-widget
          v-if="filterByDates"
          class="element-bottom-margin"
          :filter-by-dates="filterByDates"
          :cluster-id="clusterId"
          :node-pool-name="nodePoolName"
          :org-unit-metric-source="orgUnitMetricSource"
          :zoom-active="connectedZoomActive"
          @register="registerChart(chartIds[2])"
          @unregister="unregisterChart(chartIds[2])"
          @on-zoom="handleZoomIn"
          @on-reset-zoom="handleResetZoom"
          @on-zoom-toggle="handleZoomToggle"
        />
      </template>
      <div v-else class="empty-state-message">{{ emptyStateMessage }}</div>
    </div>
  </q-card>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiDatepickerWithPredefined } from "@/components/common/runai-datepicker-with-predefined";
import { RunaiAutocompleteFilterInput } from "@/components/common/runai-autocomplete-filter-input";
import { dateUtil, extendedHoursPredefinedOptions, type IRangeDates, TimeUnit } from "@/utils/date.util";
import type { PredefinedOption } from "@/models/date.model";
import { ResourceAllocationTimeRangeWidget } from "@/components/dashboard-v2/widgets/time-range-widgets/resource-allocation-time-range-widget";
import { ResourceUtilizationTimeRangeWidget } from "@/components/dashboard-v2/widgets/time-range-widgets/resource-utilization-time-range-widget";
import { WorkloadByCategoryTimeRangeWidget } from "@/components/dashboard-v2/widgets/time-range-widgets/workload-by-category-time-range-widget";
import {
  type IConnectedChartsOptions,
  useConnectedCharts,
} from "@/components/dashboard-v2/composables/use-connected-charts.composable";
import {
  type IZoomEvent,
  RESOURCE_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID,
  RESOURCE_UTILIZATION_TIME_RANGE_WIDGET_CHART_ID,
  WORKLOAD_BY_CATEGORY_TIME_RANGE_WIDGET_CHART_ID,
} from "@/models/chart.model";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";
import { useAuthStore } from "@/stores/auth.store";
import { ScopeType, ProjectFilterSortFields } from "@/swagger-models/org-unit-service-client";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { filterUtil } from "@/utils/filter.util/filter.util";

export default defineComponent({
  name: "overview-time-range-widgets",
  components: {
    ResourceUtilizationTimeRangeWidget,
    ResourceAllocationTimeRangeWidget,
    WorkloadByCategoryTimeRangeWidget,
    RunaiDatepickerWithPredefined,
    RunaiAutocompleteFilterInput,
  },
  emits: ["date-changed"],
  props: {
    filterByDates: {
      type: Object as PropType<IRangeDates>,
      required: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    nodePoolName: {
      type: String as PropType<string>,
      required: false,
    },
    hasReadClustersPermission: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hasReadDepartmentsPermission: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hasReadProjectsPermission: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      selectedDepartment: null as string | null,
      selectedProject: null as string | null,
      connectedZoomActive: undefined as boolean | undefined,
    };
  },
  setup() {
    const chartIds: string[] = [
      RESOURCE_ALLOCATION_TIME_RANGE_WIDGET_CHART_ID,
      RESOURCE_UTILIZATION_TIME_RANGE_WIDGET_CHART_ID,
      WORKLOAD_BY_CATEGORY_TIME_RANGE_WIDGET_CHART_ID,
    ];
    const options: IConnectedChartsOptions = {
      chartIds,
    };
    const { zoomIn, resetZoom, setZoomMode, registerChartToCrosshair, unregisterChartFromCrosshair } =
      useConnectedCharts(options);
    return {
      zoomIn,
      resetZoom,
      setZoomMode,
      registerChartToCrosshair,
      unregisterChartFromCrosshair,
      chartIds,
    };
  },
  created() {
    this.autoSelectFirstOptions();
  },
  computed: {
    predefinedOptions(): PredefinedOption[] {
      return extendedHoursPredefinedOptions;
    },
    showOrgUnitFilters(): boolean {
      return !this.hasReadClustersPermission;
    },
    showDepartmentFilter(): boolean {
      return this.hasReadDepartmentsPermission;
    },
    showProjectFilter(): boolean {
      if (!this.hasReadProjectsPermission) return false;
      if (this.projects.length === 0) return false;
      // When user has department permission and departments exist, only show project filter after selecting a department
      if (this.hasReadDepartmentsPermission && this.departments.length > 0) return this.selectedDepartment !== null;
      return true;
    },
    departments(): OrgUnit[] {
      return this.authStore.orgUnits.filter((ou: OrgUnit) => ou.type === ScopeType.Department);
    },
    projects(): OrgUnit[] {
      return this.authStore.orgUnits.filter((ou: OrgUnit) => ou.type === ScopeType.Project);
    },
    selectedDepartmentName(): string {
      if (!this.selectedDepartment) return "";
      const dept = this.departments.find((d: OrgUnit) => d.id === this.selectedDepartment);
      return dept?.name ?? "";
    },
    selectedProjectName(): string {
      if (!this.selectedProject) return "";
      const proj = this.projects.find((p: OrgUnit) => p.id === this.selectedProject);
      return proj?.name ?? "";
    },
    departmentSearchFn(): (searchQuery: string) => Promise<Array<{ label: string; value: string }>> {
      return async (searchQuery: string) => {
        const departments = await orgUnitService.getDepartments();
        const filtered = searchQuery
          ? departments.filter((d) => d.name.toLowerCase().includes(searchQuery.toLowerCase()))
          : departments;
        return filtered.map((d) => ({ label: d.name, value: d.id }));
      };
    },
    projectSearchFn(): (searchQuery: string) => Promise<Array<{ label: string; value: string }>> {
      return async (searchQuery: string) => {
        const filterBy = this.selectedDepartment
          ? [filterUtil.getEqualsFilterString(ProjectFilterSortFields.ParentId, this.selectedDepartment)]
          : undefined;
        const projects = await orgUnitService.getProjects(
          filterBy,
          undefined,
          undefined,
          undefined,
          undefined,
          searchQuery || undefined,
        );
        return projects.map((p) => ({ label: p.name, value: p.id }));
      };
    },
    orgUnitMetricSource(): IOrgUnitMetricSource | null {
      if (!this.showOrgUnitFilters) return null;

      if (this.selectedProject) {
        return { type: EOrgUnitMetricSourceType.Project, id: this.selectedProject };
      }
      if (this.selectedDepartment) {
        return { type: EOrgUnitMetricSourceType.Department, id: this.selectedDepartment };
      }
      return null;
    },
    isReady(): boolean {
      if (!this.showOrgUnitFilters) return true;
      return this.orgUnitMetricSource !== null;
    },
    emptyStateMessage(): string {
      if (this.showDepartmentFilter) return "Select a department to view overtime metrics";
      if (this.hasReadProjectsPermission) return "Select a project to view overtime metrics";
      return "No organization units available";
    },
  },
  methods: {
    autoSelectFirstOptions(): void {
      if (!this.showOrgUnitFilters) return;
      if (this.selectedDepartment === null && this.hasReadDepartmentsPermission && this.departments.length > 0) {
        this.selectedDepartment = this.departments[0].id;
      }
      if (this.selectedProject === null && this.hasReadProjectsPermission && this.projects.length > 0) {
        this.selectedProject = this.projects[0].id;
      }
    },
    handleDepartmentChange(value: string): void {
      this.selectedDepartment = value;
      this.selectedProject = null;
    },
    handleProjectChange(value: string): void {
      this.selectedProject = value;
    },
    updateDateRange(dateRange: Array<Date>, label: string): void {
      this.$emit("date-changed", { startDate: dateRange[0], endDate: dateRange[1], label });
    },
    rangeOfDatesCallback(date: string): boolean {
      const dateToCheck = new Date(date);
      if (Number.isNaN(dateToCheck.getTime())) return false;

      dateToCheck.setHours(0, 0, 0, 0);
      const back90Days = dateUtil.adjustDateBy(TimeUnit.day, new Date(), -90);
      back90Days.setHours(0, 0, 0, 0);

      const today = new Date();
      today.setHours(23, 59, 59, 999);

      return dateToCheck >= back90Days && dateToCheck <= today;
    },
    registerChart(chartId: string): void {
      this.registerChartToCrosshair(chartId);
    },
    unregisterChart(chartId: string): void {
      this.unregisterChartFromCrosshair(chartId);
    },
    handleZoomIn(event: IZoomEvent, currentChartId?: string): void {
      this.zoomIn(event.min, event.max, currentChartId || event.chartId);
      this.connectedZoomActive = false;
      this.setZoomMode(false);
    },
    handleResetZoom(currentChartId: string | IZoomEvent): void {
      const chartId = typeof currentChartId === "string" ? currentChartId : currentChartId.chartId;
      this.resetZoom(chartId);
      this.connectedZoomActive = false;
      this.setZoomMode(false);
    },
    handleZoomToggle(active: boolean): void {
      this.connectedZoomActive = active;
      this.setZoomMode(active);
    },
  },
  watch: {
    "authStore.orgUnits": {
      handler(): void {
        this.autoSelectFirstOptions();
      },
    },
  },
});
</script>

<style scoped lang="scss">
.widgets-content-wrapper {
  padding: 12px 16px;
}
.element-bottom-margin {
  margin-bottom: 13px;
}
.filters-row {
  display: flex;
  align-items: center;
  gap: var(--space-s);
  flex-wrap: wrap;
}
.org-unit-filter {
  min-width: 180px;
  max-width: 250px;
}
.date-picker-filter {
  margin-left: auto;
}
.empty-state-message {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
  color: var(--b-text-sub);
  font-size: 14px;
}
</style>
