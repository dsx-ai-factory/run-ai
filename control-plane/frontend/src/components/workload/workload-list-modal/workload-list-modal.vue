<template>
  <section class="workload-modal">
    <runai-base-dialog
      class="workload-list-dialog"
      :model-value="true"
      @close="$emit('close')"
      size="dynamic"
      no-padding
    >
      <template #header>
        <span class="modal-header"> {{ modalHeader }}</span>
      </template>

      <template #body>
        <runai-list-modal-search v-model="filterBy.searchTerm" placeholder="Search workloads" />
        <q-separator />
        <runai-table
          class="workloads-list-table"
          :rows="workloads"
          :columns="modalOptions.columns"
          :loading="loading"
          :filter-by="filterBy"
          modal-view
          @update-filters="updateFilterBy"
          no-shadow
          :get-row-icon="getWorkloadIcon"
          disable-selection
          is-server-side-pagination
          :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25]"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loading"
              :filter-by="filterBy"
              entity-name="workload"
              :show-create-btn="false"
              :show-filter-icon-and-button="false"
            >
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
      <template #footer-left>
        <q-btn
          flat
          :ripple="false"
          color="primary"
          class="btn-link"
          label="Go to workloads to view more details"
          @click="redirectToWorkloads"
        />
      </template>
      <template #footer>
        <q-btn aid="close-workloads-modal-btn" label="Close" color="primary" @click="$emit('close')" />
      </template>
    </runai-base-dialog>
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiTable } from "@/components/common";
//model
import {
  EColumnFilterType,
  EFilterOperator,
  type IFilterBy,
  type IFreeTextFilterModel,
  type IPaginationFilter,
} from "@/models/filter.model";
import { EWorkloadErrorMessage, type IWorkloadListModalOptions } from "@/models/workload.model";
import { workloadIndexColumns } from "@/table-models/workload.table-model";
import { type Workload, WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
//util
import { workloadUtil } from "@/utils/workload.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { filterService } from "@/services/filter.service/filter.service";
import { alertUtil } from "@/utils/alert.util";
import RunaiListModalSearch from "@/components/common/runai-list-modal-search/runai-list-modal-search.vue";
import { ETableFilters } from "@/models/table.model";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { filterUtil } from "@/utils/filter.util/filter.util";

export default defineComponent({
  name: "workload-list-modal",
  components: { RunaiListModalSearch, RunaiTable, RunaiBaseDialog, RunaiTableNoData },
  emits: ["close"],
  props: {
    modalOptions: {
      type: Object as PropType<IWorkloadListModalOptions>,
      required: true,
    },
  },
  data() {
    return {
      filterBy: {} as IFilterBy,
      workloads: [] as Workload[],
      loading: false as boolean,
    };
  },
  created() {
    const defaultFilters = filterService.getEmptyFilterByModel({
      sortBy: WorkloadSortFilterFields.Name,
      rowsPerPage: 10,
    });
    this.updateFilterBy(defaultFilters);
  },
  computed: {
    modalHeader(): string {
      return `Workloads Associated with ${this.modalOptions.entityType} ${this.modalOptions.entityName}`;
    },
  },
  methods: {
    getWorkloadIcon(workload: Workload): string {
      return workloadUtil.getWorkloadIcon(workload);
    },
    async updateFilterBy(filterBy: IFilterBy): Promise<void> {
      this.filterBy = filterBy;
      const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy, undefined, true);
      filters.filterBy = [
        this.modalOptions.entityFilter,
        filterUtil.getEqualsFilterString(WorkloadSortFilterFields.ClusterId, this.modalOptions.clusterId),
        ...(filters.filterBy || []),
      ];
      try {
        this.setLoading(true);
        await this.loadWorkloadsCount(filters.filterBy, filters.search);
        await this.loadWorkloads(filters);
      } catch (error: unknown) {
        this.handleLoadWorkloadsError(error);
      } finally {
        this.setLoading(false);
      }
    },
    async loadWorkloadsCount(filterBy: Array<string>, search?: string): Promise<void> {
      this.filterBy.rowsNumber = await workloadService.getWorkloadsCount(filterBy, undefined, search);
    },
    async loadWorkloads(filters: IPaginationFilter): Promise<void> {
      this.workloads = await workloadService.getWorkloads(
        false,
        filters.offset,
        filters.limit,
        filters.sortOrder,
        (filters.sortBy || undefined) as WorkloadSortFilterFields | undefined,
        filters.filterBy,
        filters.search,
      );
    },
    handleLoadWorkloadsError(error: unknown): void {
      console.error(error);
      this.$q.notify(alertUtil.getError(EWorkloadErrorMessage.FailedToLoadWorkloads));
    },
    setLoading(isLoading: boolean): void {
      this.loading = isLoading;
    },
    getWorkloadsFilterBy(): IFilterBy {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        this.modalOptions.entityFilter,
        workloadIndexColumns,
      );
      return filterService.loadFilters(window.location, ETableFilters.WORKLOAD, defaultFilters, true);
    },
    setColumnFilter(workloadsFilterBy: IFilterBy): void {
      const entityColumnFilter: IFreeTextFilterModel = {
        term: this.modalOptions.entityName,
        name: this.modalOptions.filterName,
        value: EFilterOperator.Equals,
        type: EColumnFilterType.FreeText,
        label: this.modalOptions.entityType,
      };
      filterService.setColumnAdvancedFilter(workloadsFilterBy, entityColumnFilter, ETableFilters.WORKLOAD);
    },
    saveAndNavigateToWorkloads(workloadsFilterBy: IFilterBy): void {
      filterService.saveFilters(ETableFilters.WORKLOAD, {
        ...workloadsFilterBy,
        sortBy: WorkloadSortFilterFields.Name,
      });

      this.$router.push({
        name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      });
    },
    redirectToWorkloads(): void {
      const workloadsFilterBy: IFilterBy = this.getWorkloadsFilterBy();
      this.setColumnFilter(workloadsFilterBy);

      this.saveAndNavigateToWorkloads(workloadsFilterBy);
    },
  },
});
</script>
