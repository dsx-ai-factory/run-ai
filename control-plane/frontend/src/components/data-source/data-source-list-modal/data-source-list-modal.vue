<template>
  <section class="data-source-list-modal">
    <runai-base-dialog :model-value="modelValue" @close="$emit('close')" size="width-md" :no-padding="true">
      <template #header> {{ modalOptions.header }} </template>

      <template #body>
        <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search data sources" />
        <q-separator />
        <runai-table
          :rows="dataSources"
          :columns="columns"
          :loading="loading"
          :filter-by="tableFilter"
          @update-filters="updateFilter"
          no-shadow
          modal-view
          :get-row-icon="getRowIcon"
          disable-selection
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loading"
              :filter-by="tableFilter"
              entity-name="data source"
              ph-icon-name="ph-bold ph-database"
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
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiListModalSearch } from "@/components/common";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

// services
import { alertUtil } from "@/utils/alert.util";
import { filterService } from "@/services/filter.service/filter.service";

// models
import type { IDataSourceModalOptions } from "@/models/data-source.model";
import { allDataSourceColumnsMap, dataSourceMiniTableColumns } from "@/table-models/data-source.table-model";
import type { AssetKind, DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";

// stores
import { useDataSourceStore } from "@/stores/data-source.store";
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { DATA_SOURCE_ICONS_NAMES } from "@/common/icons.constant";

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
      type: Object as PropType<IDataSourceModalOptions>,
      required: true,
    },
  },
  data() {
    return {
      dataSourceStore: useDataSourceStore(),
      modelValue: true as boolean,
      loading: false as boolean,
      tableFilter: {} as IFilterBy,
      dataSourceIdsMap: new Set<string>(this.modalOptions.dataSourceIds) as Set<string>,
      columns: dataSourceMiniTableColumns as Array<ITableColumn>,
    };
  },
  created() {
    this.loadFilters();
    this.loadDataSources();
  },
  computed: {
    dataSources(): Array<DatasourceListResponseEntry> {
      return this.dataSourceStore.dataSourceList.filter((dataSource: DatasourceListResponseEntry) =>
        this.dataSourceIdsMap.has(dataSource.meta.id),
      );
    },
  },
  methods: {
    getRowIcon(dataSource: DatasourceListResponseEntry): string {
      const kind: AssetKind = dataSource.meta.kind;
      return (DATA_SOURCE_ICONS_NAMES[kind as keyof object] as string) || DATA_SOURCE_ICONS_NAMES.default;
    },
    loadFilters(): void {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        allDataSourceColumnsMap.dataSource.name,
        dataSourceMiniTableColumns,
      );
      this.tableFilter = filterService.getEmptyFilterByModel(defaultFilters);
    },
    async loadDataSources(): Promise<void> {
      try {
        this.loading = true;
        if (!this.dataSourceStore.dataSourceList.length) {
          await this.dataSourceStore.loadDataSources({ usageInfo: true });
        }
      } catch (error: unknown) {
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load data sources"));
      } finally {
        this.loading = false;
      }
    },
    updateFilter(filter: IFilterBy): void {
      this.tableFilter = filter;
    },
  },
});
</script>
