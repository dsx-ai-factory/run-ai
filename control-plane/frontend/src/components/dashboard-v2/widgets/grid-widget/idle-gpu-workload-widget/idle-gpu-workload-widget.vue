<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <chart-widget-wrapper
    class="idle-gpu-workload-widget"
    aid="idle-gpu-workload-widget"
    :options="options"
    :loading="loading"
    :error="error"
    slot-wrapper-class=""
    hide-actions
    @link-clicked="redirectToWorkloads"
    is-grid
  >
    <runai-table
      :rows="workloads"
      :columns="idleGpuWorkloadColumns"
      :filter-by="filterBy"
      :loading="loading"
      disable-selection
      no-shadow
      :hide-pagination="!empty"
      :bordered="false"
      @workload-clicked="redirectToWorkloadsFilterByWorkload"
    >
      <template #no-data>
        <runai-table-no-data
          :filter-by="filterBy"
          entity-name="workload"
          :custom-message="EWidgetEmptyMessages.NoIdleWorkloads"
          :show-create-btn="false"
          :show-filter-icon-and-button="false"
      /></template>
    </runai-table>
  </chart-widget-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType, type Ref, ref, watch } from "vue";
import {
  EWidgetEmptyMessages,
  EWidgetEntity,
  type IWidgetWrapperOptions,
  MIN_IDLE_GPUS_FILTER_VALUE,
} from "@/models/chart.model";
import { type Workload, WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { idleGpuWorkloadColumns, workloadIndexColumns } from "@/table-models/workload.table-model";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { useRouter } from "vue-router";
import { RunaiTable } from "@/components/common";
import { EFilterOperator, type IFilterBy } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import { ETableFilters } from "@/models/table.model";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { widgetUtil } from "@/utils/widget.util";
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { useRefresh } from "@/composables/use-refresh.composable";
import { EIntervalLabels } from "@/models/interval.model";

const NUMBER_OF_WORKLOADS = 5;
export default defineComponent({
  name: "idle-gpu-workload-widget",
  components: { RunaiTableNoData, ChartWidgetWrapper, RunaiTable },
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
  setup(props) {
    const router = useRouter();

    const loading = ref(true);
    const error = ref(false);
    const empty = ref(false);
    const filterBy: Ref<IFilterBy> = ref({ descending: true, sortBy: "idleAllocatedGpus" });
    const workloads: Ref<Workload[]> = ref([]);
    const options: Ref<IWidgetWrapperOptions> = ref({
      linkText: "All workloads",
      title: "Idle workloads",
      entityType: EWidgetEntity.Workload,
      tooltipText: `The ${NUMBER_OF_WORKLOADS} workloads with the highest number of idle GPU devices.<br/>Updated about every 20 minutes, so it may lag behind the idle allocated GPU devices count.`,
    });

    const getWorkloads = async () => {
      try {
        const response = await workloadService.getWorkloads(
          false,
          undefined,
          NUMBER_OF_WORKLOADS,
          "desc",
          WorkloadSortFilterFields.IdleAllocatedGpus,
          getFilters(),
        );
        workloads.value = response;
        empty.value = response.length === 0;
        error.value = false;
      } catch (err) {
        error.value = true;
      } finally {
        loading.value = false;
      }
    };

    const getFilters = (): string[] => {
      const filters = [
        filterUtil.getGreaterThanOrEqualFilterString(
          WorkloadSortFilterFields.IdleAllocatedGpus,
          MIN_IDLE_GPUS_FILTER_VALUE,
        ),
        filterUtil.getEqualsFilterString(WorkloadSortFilterFields.ClusterId, props.clusterId),
      ];
      if (props.nodePoolName) {
        filters.push(filterUtil.getEqualsFilterString(WorkloadSortFilterFields.Nodepool, props.nodePoolName));
      }
      return filters;
    };

    // A row click targets one workload; the idle filter is left out because the table can be ~20 minutes stale.
    const getQueryParams = (workloadName?: string) => {
      return {
        sortBy: WorkloadSortFilterFields.IdleAllocatedGpus,
        descending: "true",
        advancedFilters: workloadName
          ? `[${filterUtil.getFreeTextFilterQuery(
              WorkloadSortFilterFields.Name,
              workloadName,
              "Workload",
              EFilterOperator.Equals,
            )}]`
          : widgetUtil.getIdleWorkloadsAdvancedFilters(props.nodePoolName),
      };
    };

    const redirectToWorkloads = (workloadName?: string) => {
      const query = getQueryParams(workloadName);
      selectIdleGpuColumn();
      router.push({
        name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
        query,
      });
    };

    const redirectToWorkloadsFilterByWorkload = (workload: Workload) => {
      redirectToWorkloads(workload.name);
    };

    const selectIdleGpuColumn = () => {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        WorkloadSortFilterFields.IdleAllocatedGpus,
        workloadIndexColumns,
      );
      const filters = filterService.loadFilters(window.location, ETableFilters.WORKLOAD, defaultFilters);
      if (
        filters.displayedColumns?.length &&
        !filters?.displayedColumns?.includes(WorkloadSortFilterFields.IdleAllocatedGpus)
      ) {
        filters.displayedColumns.push(WorkloadSortFilterFields.IdleAllocatedGpus);
      }
      filterService.saveFilters(ETableFilters.WORKLOAD, filters);
    };

    useRefresh(EIntervalLabels.IdleGpuWorkloadWidget, getWorkloads);

    watch([() => props.nodePoolName, () => props.clusterId], async () => {
      loading.value = true;
      error.value = false;
      await getWorkloads();
    });
    return {
      options,
      workloads,
      loading,
      idleGpuWorkloadColumns,
      redirectToWorkloads,
      redirectToWorkloadsFilterByWorkload,
      empty,
      error,
      filterBy,
    };
  },
  computed: {
    EWidgetEmptyMessages() {
      return EWidgetEmptyMessages;
    },
  },
});
</script>

<style lang="scss"></style>
