<template>
  <section class="pvc-event-history-tab">
    <section class="table-filters-bar row q-pa-xs items-center justify-between">
      <runai-page-filter
        :filter-options="filterOptions"
        @update-filter-by="filters = $event"
        :initial-filters-model="columnFilters"
        hide-cluster-filter
      />
      <runai-general-search :search-term="searchTerm" @update-search-term="filters.searchTerm = $event" />
    </section>
    <runai-table
      :="tableDefaultProps"
      :columns="columns"
      :rows="filteredRows"
      :filter-by="filters"
      @update-filters="filters = $event"
      :loading="loading"
      :bordered="false"
      disable-selection
      hide-pagination
    >
      <template #no-data>
        <runai-table-no-data
          v-if="!loading"
          :filter-by="filters"
          entity-name="event"
          :show-create-btn="false"
          @clear-filters="filters = getDefaultFilters()"
        />
      </template>
    </runai-table>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// COMPONENTS
import { RunaiTable } from "@/components/common/runai-table";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiGeneralSearch } from "@/components/common/runai-general-search";
import { RunaiPageFilter } from "@/components/common/runai-page-filter";
// MODELS
import { type ITableColumn } from "@/models/table.model";
import type { IFilterBy, IFilterModel, IFilterOption } from "@/models/filter.model";
import type { HistoryRecord } from "@/swagger-models/assets-service-client";
import { EIntervalLabels } from "@/models/interval.model";
// UTILS
import { dateUtil } from "@/utils/date.util";
import { intervalUtil } from "@/utils/interval.util";
// CONSTANTS
import { tableDefaultProps } from "@/common/table.constant";
// SERVICES
import { filterService } from "@/services/filter.service/filter.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
import { assetsUtil } from "@/utils/assets.util/assets.util";

const eventHistoryColumns: ITableColumn[] = [
  {
    name: "created-at",
    label: "Date & Time",
    field: (row: HistoryRecord) => row.meta.creationTimestamp,
    align: "left",
    sortable: true,
    format: (createdAt: string | undefined) => (createdAt ? dateUtil.dateAndTimeFormat(new Date(createdAt)) : "-"),
    display: true,
  },
  {
    name: "event",
    label: "Event",
    field: (row: HistoryRecord) => (assetsUtil.isPhaseUpdateRecordType(row) ? "Status changed" : row.spec.event?.reason),
    align: "left",
    sortable: true,
    display: true,
    format: (event: string | undefined) => event || "-",
  },
  {
    name: "event-type",
    label: "Type",
    field: (row: HistoryRecord) => row.spec.event?.type || (row.spec.phaseUpdate?.phase ? "Normal" : undefined),
    align: "left",
    sortable: true,
    display: true,
    format: (eventType: string | undefined) => eventType || "-",
  },
  {
    name: "details",
    label: "Details",
    field: (row: HistoryRecord) => row.spec.event?.message || row.spec.phaseUpdate?.phase,
    align: "left",
    sortable: true,
    display: true,
    format: (details: string | undefined) => details || "-",
  },
];

const getDefaultFilters = (): IFilterBy => {
  return {
    displayedColumns: eventHistoryColumns
      .filter((col: ITableColumn) => col.display)
      .map((col: ITableColumn) => col.name),
    sortBy: "createdAt",
    descending: true,
    searchTerm: "",
  };
};

export default defineComponent({
  name: "pvc-event-history-tab",
  components: { RunaiTable, RunaiTableNoData, RunaiGeneralSearch, RunaiPageFilter },
  props: {
    pvcId: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      tableDefaultProps,
      filters: getDefaultFilters() as IFilterBy,
      rows: [] as HistoryRecord[],
      loading: false,
      columns: eventHistoryColumns,
    };
  },
  created() {
    this.resetComponentState();
  },
  computed: {
    filterOptions(): Array<IFilterOption> {
      return this.columns ? filterService.getFilterOptions(this.columns) : [];
    },
    columnFilters(): Array<IFilterModel> {
      return this.filters.columnFilters || ([] as Array<IFilterModel>);
    },
    searchTerm(): string {
      return this.filters.searchTerm || "";
    },
    filteredRows(): HistoryRecord[] {
      const events: HistoryRecord[] = filterService.filterBySearchTermAndByColumns(
        this.rows,
        this.filters,
        eventHistoryColumns,
      );
      return events;
    },
  },
  methods: {
    getDefaultFilters,
    updateSearchTerm(newVal: string): void {
      this.filters.searchTerm = newVal;
    },
    async loadEvents(pvcId: string): Promise<void> {
      try {
        this.loading = true;
        this.rows = await dataSourceService.loadPvcEvents(pvcId);
      } catch (e) {
        console.error(e);
      } finally {
        this.loading = false;
      }
    },
    async resetComponentState(): Promise<void> {
      intervalUtil.stopInterval(EIntervalLabels.PvcEventHistoryTab);
      await this.loadEvents(this.pvcId);
      intervalUtil.startInterval(EIntervalLabels.PvcEventHistoryTab, async () => {
        await this.loadEvents(this.pvcId);
      });
    },
  },
  watch: {
    pvcId(): void {
      this.resetComponentState();
    },
  },
  unmounted() {
    intervalUtil.stopInterval(EIntervalLabels.PvcEventHistoryTab);
  },
});
</script>

<style lang="scss" scoped>
.table-filters-bar {
  border-bottom: 1px solid $black-12;
}
</style>
