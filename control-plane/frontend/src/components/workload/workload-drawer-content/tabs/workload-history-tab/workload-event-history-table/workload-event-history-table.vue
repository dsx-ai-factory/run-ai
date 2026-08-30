<template>
  <section class="runai-event-history">
    <section class="table-top-actions-container">
      <runai-page-filter
        :initial-filters-model="columnFilters"
        :filter-by="filters"
        @update-filter-by="$emit('update:filter-by', $event)"
        :filter-options="filterOptions"
        hide-cluster-filter
      />
      <runai-general-search :search-term="searchTerm" @update-search-term="updateFilterBy('searchTerm', $event)" />
    </section>
    <q-table
      :="tableDefaultProps"
      :rows="historyRecords"
      :columns="columns"
      row-key="creationTimestamp"
      :pagination="filters"
      :rows-per-page-options="[0]"
      virtual-scroll
      hide-pagination
    >
      <template v-slot:body-cell-details="props">
        <q-td style="padding: 7px 16px; white-space: break-spaces">
          <runai-status v-if="isPhaseUpdateRecordType(props.row)" :options="getStatusOptions(props.row)" />
          <div v-else>{{ props.value }}</div>
        </q-td>
      </template>

      <template v-slot:no-data>
        <runai-table-no-data
          :filter-by="filters"
          entity-name="event"
          custom-message="There are no events yet."
          @clear-filters="clearFilters"
          :show-create-btn="false"
          light-icon
        />
      </template>
    </q-table>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

// components
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiPageFilter } from "@/components/common/runai-page-filter";
import { RunaiGeneralSearch } from "@/components/common/runai-general-search";

// models
import type { IFilterBy } from "@/models/filter.model";
import type { HistoryRecord } from "@/swagger-models/workloads-client";
import type { IFilterModel } from "@/models/filter.model";
// services
import type { IStatusColOptions, ITableColumn } from "@/models/table.model";
import type { ExtendedPhase } from "@/models/workload.model";
import { WorkloadPhaseMap } from "@/models/workload.model";
import { eventHistoryColumns } from "@/table-models/workload.table-model";
import { workloadUtil } from "@/utils/workload.util";
import type { IFilterOption } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import { tableDefaultProps, type TTableDefaultProps } from "@/common/table.constant";
import { RunaiStatus } from "@/components/common/runai-status";

export default defineComponent({
  components: { RunaiStatus, RunaiTableNoData, RunaiPageFilter, RunaiGeneralSearch },
  emits: ["update:filter-by"],
  props: {
    historyRecords: {
      type: Array as PropType<HistoryRecord[]>,
      required: true,
    },
    filters: {
      type: Object as PropType<IFilterBy>,
      required: true,
    },
  },
  data() {
    return {
      tableDefaultProps: tableDefaultProps as TTableDefaultProps,
    };
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
    columns(): ITableColumn[] {
      return eventHistoryColumns;
    },
  },
  methods: {
    getStatusOptions(historyRecord: HistoryRecord): IStatusColOptions {
      const statusOptions = WorkloadPhaseMap[(historyRecord.spec.phaseUpdate?.phase as ExtendedPhase) || ""];
      return {
        status: statusOptions.status,
        color: statusOptions.color,
        displayAnimation: false,
      };
    },
    isPhaseUpdateRecordType(historyRecord: HistoryRecord): boolean {
      return workloadUtil.isPhaseUpdateRecordType(historyRecord);
    },
    statusColor(historyRecord: HistoryRecord): string {
      if (!historyRecord.spec.phaseUpdate?.phase) return "";
      const color = WorkloadPhaseMap[historyRecord.spec.phaseUpdate?.phase].color || "";

      return color.toLowerCase() + "-color";
    },
    clearFilters(): void {
      this.$emit("update:filter-by", {
        ...this.filters,
        columnFilters: [],
        searchTerm: "",
      });
    },
    updateFilterBy(key: string, newVal: string | number | []): void {
      this.$emit("update:filter-by", { ...this.filters, [key]: newVal });
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-event-history {
  .table-top-actions-container {
    background-color: $white;
    border-top: 1px solid $black-12;
    border-bottom: 1px solid $black-12;
    padding: 7px 0;
  }

  .table-top-actions-container {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
