<template>
  <stat-widget-wrapper :options="options" @link-clicked="redirectToWorkloads" :loading="loading" :error="error">
    <template #data> {{ workloadsCount }}</template>
  </stat-widget-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//routes
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { workloadService } from "@/services/cluster/workload.service/workload.service";

//models
import { EWidgetEntity, type IWidgetWrapperOptions } from "@/models/chart.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { Phase, WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import {
  EColumnFilterType,
  EFilterOperator,
  type IEnumBasedFilterModel,
  type IFilterBy,
  type IFreeTextFilterModel,
} from "@/models/filter.model";
//cmps
import { StatWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/stat-widget-wrapper";
import { filterService } from "@/services/filter.service/filter.service";
import { workloadIndexColumns } from "@/table-models/workload.table-model";
import { ETableFilters } from "@/models/table.model";
import { workloadPhaseFilterOptions } from "@/models/workload.model";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { filterUtil } from "@/utils/filter.util/filter.util";

export default defineComponent({
  name: "pending-workloads-widget",
  components: { StatWidgetWrapper },
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    nodePoolName: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      options: {
        linkText: "All workloads",
        title: "Running workloads",
        entityType: EWidgetEntity.Workload,
      } as IWidgetWrapperOptions,
      workloadsCount: 0 as number,
      loading: true as boolean,
      error: false as boolean,
    };
  },
  created() {
    this.getWorkloadsCount();
    this.startRefreshLoadData();
  },
  methods: {
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.RunningWorkloadsWidget, this.getWorkloadsCount);
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.RunningWorkloadsWidget);
    },
    async getWorkloadsCount(): Promise<void> {
      try {
        this.workloadsCount = await workloadService.getWorkloadsCount(this.getFilterBy());

        this.error = false;
      } catch (error: unknown) {
        this.handleError(error);
      } finally {
        this.loading = false;
      }
    },
    getFilterBy(): string[] {
      const filterBy = [
        filterUtil.getEqualsFilterString(WorkloadSortFilterFields.Phase, Phase.Running),
        filterUtil.getEqualsFilterString(WorkloadSortFilterFields.ClusterId, this.clusterId),
      ];
      if (this.nodePoolName) {
        filterBy.push(filterUtil.getEqualsFilterString(WorkloadSortFilterFields.Nodepool, this.nodePoolName));
      }
      return filterBy;
    },
    handleError(error: unknown): void {
      this.error = true;
      if (error instanceof HttpErrorResponse) {
        console.error(error.message);
      } else {
        console.error("Error fetching data", error);
      }
    },
    redirectToWorkloads(): void {
      this.setColumnFilter();
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
    },
    getWorkloadsFilterBy(): IFilterBy {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        WorkloadSortFilterFields.Phase,
        workloadIndexColumns,
      );
      return filterService.loadFilters(window.location, ETableFilters.WORKLOAD, defaultFilters);
    },
    setColumnFilter(): void {
      const workloadsFilterBy: IFilterBy = this.getWorkloadsFilterBy();
      this.setRunningColumnFilter(workloadsFilterBy);
      if (this.nodePoolName) {
        this.setNodePoolColumnFilter(workloadsFilterBy);
      } else {
        this.removeNodePoolFilter(workloadsFilterBy);
      }
    },
    setRunningColumnFilter(workloadsFilterBy: IFilterBy): void {
      const runningColumnFilter: IEnumBasedFilterModel = {
        name: WorkloadSortFilterFields.Phase,
        selectedValues: [Phase.Running],
        type: EColumnFilterType.EnumBased,
        label: "Status",
        selectOptions: workloadPhaseFilterOptions,
      };
      filterService.setColumnAdvancedFilter(workloadsFilterBy, runningColumnFilter, ETableFilters.WORKLOAD);
    },
    setNodePoolColumnFilter(workloadsFilterBy: IFilterBy): void {
      const nodePoolColumnFilter: IFreeTextFilterModel = {
        name: WorkloadSortFilterFields.Nodepool,
        term: this.nodePoolName as string,
        value: EFilterOperator.Equals,
        label: "Node pool",
        type: EColumnFilterType.FreeText,
      };
      filterService.setColumnAdvancedFilter(workloadsFilterBy, nodePoolColumnFilter, ETableFilters.WORKLOAD);
    },
    removeNodePoolFilter(workloadsFilterBy: IFilterBy): void {
      filterService.removeAdvancedFilter(workloadsFilterBy, WorkloadSortFilterFields.Nodepool, ETableFilters.WORKLOAD);
    },
  },
  watch: {
    nodePoolName: {
      async handler(): Promise<void> {
        this.loading = true;
        await this.getWorkloadsCount();
      },
    },
    clusterId: {
      async handler(): Promise<void> {
        this.loading = true;
        await this.getWorkloadsCount();
      },
    },
  },
  unmounted() {
    this.stopRefreshLoadData();
  },
});
</script>

<style scoped lang="scss"></style>
