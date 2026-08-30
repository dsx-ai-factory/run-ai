<template>
  <section class="tools-modal">
    <runai-base-dialog :model-value="true" @close="$emit('close')" size="width-md" :no-padding="true">
      <template #header> Tools Associated with Environment {{ environmentName }} </template>
      <template #body>
        <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search tools" />
        <q-separator />
        <tool-table
          :tools="tools"
          :columns="displayedColumns"
          :filter-by="tableFilter"
          @clear-filters="tableFilter.searchTerm = ''"
          :modal-view="true"
        ></tool-table>
      </template>

      <template #footer>
        <q-btn label="Close" color="primary" @click="$emit('close')" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import type { Connection } from "@/swagger-models/assets-service-client";

import { allToolTableColumnsMap, toolTableColumns } from "@/table-models/environment.table-model";
import { filterService } from "@/services/filter.service/filter.service";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { ToolTable } from "./tool-table";
import type { IFilterBy } from "@/models/filter.model";
import { tableUtil } from "@/utils/table.util";
import type { ITableColumn } from "@/models/table.model";
import { RunaiListModalSearch } from "@/components/common";

export default defineComponent({
  components: {
    RunaiBaseDialog,
    ToolTable,
    RunaiListModalSearch,
  },
  emits: ["close"],
  props: {
    environmentName: {
      type: String as PropType<string>,
      required: true,
    },
    tools: {
      type: Array as PropType<Array<Connection>>,
      required: true,
    },
  },
  data() {
    return {
      tableFilter: {} as IFilterBy,
    };
  },
  created() {
    this.loadFilters();
  },
  computed: {
    displayedColumns(): Array<ITableColumn> {
      return tableUtil.getDisplayColumns(toolTableColumns, this.tableFilter.displayedColumns);
    },
  },
  methods: {
    loadFilters(): void {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        allToolTableColumnsMap.toolName.name,
        toolTableColumns,
      );

      this.tableFilter = filterService.getEmptyFilterByModel(defaultFilters);
    },
    updateFilter(filter: IFilterBy): void {
      this.tableFilter = filter;
    },
  },
});
</script>
