<template>
  <section class="workload-connections-list-modal">
    <runai-base-dialog model-value @close="$emit('close')" size="width-lg" no-padding>
      <template #header> {{ modalOptions.header }} </template>

      <template #body>
        <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search connections" />
        <q-separator />
        <runai-table
          :rows="modalOptions.connections"
          :columns="columns"
          :filter-by="tableFilter"
          @update-filters="updateFilter"
          :loading="false"
          no-shadow
          modal-view
          :get-row-icon="getRowIcon"
          disable-selection
        >
          <template #no-data>
            <runai-table-no-data
              :filter-by="tableFilter"
              entity-name="connection"
              ph-icon-name="ph-bold ph-plug"
              icon-size="44"
              :show-create-btn="false"
              :show-filter-icon-and-button="false"
            >
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>

      <template #footer>
        <q-btn label="Close" color="primary" @click="$emit('close')" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { computed, defineComponent, type ComputedRef, type PropType } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiListModalSearch } from "@/components/common";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

// services
import { filterService } from "@/services/filter.service/filter.service";

// models

// stores
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { allWorkloadsIconsMap } from "@/common/icons.constant";
import type { IWorkloadConnectionsModalOptions } from "@/models/workload.model";
import { allConnectionsTableColumnsMap, connectionsTableColumns } from "@/table-models/workload.table-model";
import type { Connection } from "@/swagger-models/workloads-client";
import { InternalConnectionType } from "@/swagger-models/assets-service-client";

export default defineComponent({
  components: {
    RunaiBaseDialog,
    RunaiListModalSearch,
    RunaiTable,
    RunaiTableNoData,
  },
  emits: ["close"],
  props: {
    modalOptions: {
      type: Object as PropType<IWorkloadConnectionsModalOptions>,
      required: true,
    },
  },
  provide() {
    return {
      workloadDetails: computed(() => this.modalOptions.workloadDetails) as ComputedRef<
        IWorkloadConnectionsModalOptions["workloadDetails"]
      >,
    };
  },
  data() {
    return {
      tableFilter: {} as IFilterBy,
      columns: connectionsTableColumns as Array<ITableColumn>,
    };
  },
  created() {
    this.loadFilters();
  },
  methods: {
    getRowIcon(connection: Connection): string {
      const toolType =
        connection.connectionType === InternalConnectionType.ServingPort ? "serving-endpoint" : connection.toolType;
      return allWorkloadsIconsMap[toolType] || allWorkloadsIconsMap.NVIDIA_WORKLOAD_ICON;
    },
    loadFilters(): void {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        allConnectionsTableColumnsMap.name.name,
        connectionsTableColumns,
      );
      this.tableFilter = filterService.getEmptyFilterByModel(defaultFilters);
    },
    updateFilter(filter: IFilterBy): void {
      this.tableFilter = filter;
    },
  },
});
</script>
