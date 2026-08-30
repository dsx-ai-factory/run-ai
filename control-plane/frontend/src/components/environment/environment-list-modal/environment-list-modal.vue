<template>
  <section class="environment-list-modal">
    <runai-base-dialog :model-value="modelValue" @close="$emit('close')" size="width-md" :no-padding="true">
      <template #header> {{ modalOptions.header }} </template>

      <template #body>
        <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search environments" />
        <q-separator />
        <runai-table
          :rows="environments"
          :columns="columns"
          :loading="loading"
          :filter-by="tableFilter"
          @update-filters="updateFilter"
          no-shadow
          modal-view
          disable-selection
          hide-pagination
          :get-row-icon="getRowIcon"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loading"
              :filter-by="tableFilter"
              entity-name="environment"
              ph-icon-name="ph-bold ph-x-square"
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
import type { IEnvironmentModalOptions } from "@/models/environment.model";
import type { Connection, DistributedFramework, EnvironmentAsset } from "@/swagger-models/assets-service-client";
import { distFrameworkIconsMap, toolTypeIconsMap } from "@/common/icons.constant";

// stores
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { useEnvironmentStore } from "@/stores/environment.store";
import { allEnvironmentColumnsMap, environmentMiniTableColumns } from "@/table-models/environment.table-model";

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
      type: Object as PropType<IEnvironmentModalOptions>,
      required: true,
    },
  },
  data() {
    return {
      environmentStore: useEnvironmentStore(),
      modelValue: true as boolean,
      loading: false as boolean,
      tableFilter: {} as IFilterBy,
      environmentIdsMap: new Set<string>(this.modalOptions.environmentIds) as Set<string>,
      columns: environmentMiniTableColumns as Array<ITableColumn>,
    };
  },
  created() {
    this.loadFilters();
    this.loadEnvironments();
  },
  computed: {
    environments(): Array<EnvironmentAsset> {
      return this.environmentStore.environmentList.filter((env: EnvironmentAsset) =>
        this.environmentIdsMap.has(env.meta.id),
      );
    },
  },
  methods: {
    loadFilters(): void {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        allEnvironmentColumnsMap.environment.name,
        environmentMiniTableColumns,
      );
      this.tableFilter = filterService.getEmptyFilterByModel(defaultFilters);
    },
    async loadEnvironments(): Promise<void> {
      try {
        this.loading = true;
        if (!this.environmentStore.environmentList.length) {
          await this.environmentStore.loadEnvironments({ usageInfo: true });
        }
      } catch (error: unknown) {
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load environments"));
      } finally {
        this.loading = false;
      }
    },
    updateFilter(filter: IFilterBy): void {
      this.tableFilter = filter;
    },
    getRowIcon(environment: EnvironmentAsset): string {
      const distFramework: DistributedFramework | null = environment.meta.workloadSupportedTypes?.distFramework || null;
      if (distFramework) return distFrameworkIconsMap[distFramework];
      const connections: Array<Connection> = environment.spec.connections || [];
      if (!connections?.length) return "";

      const defaultIcon = "custom-tool";
      const firstConnection = connections[0];
      return (
        (firstConnection.internalToolInfo && toolTypeIconsMap[firstConnection.internalToolInfo.toolType]) ||
        (firstConnection.externalToolInfo && toolTypeIconsMap[firstConnection.externalToolInfo.toolType]) ||
        defaultIcon
      );
    },
  },
});
</script>
