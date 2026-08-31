<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="topologies-preview-modal">
    <runai-base-dialog class="topologies-preview-modal" model-value @close="$emit('close')" size="width-xl" no-padding>
      <template #header>
        <span class="modal-header"> {{ modalHeader }}</span>
      </template>

      <template #body>
        <runai-list-modal-search v-model="filterBy.searchTerm" placeholder="Search topologies" :debounce="300" />
        <q-separator />
        <runai-table
          class="topologies-list-table"
          :rows="topologies"
          :columns="modalOptions.columns"
          :loading="loading"
          :filter-by="filterBy"
          modal-view
          @update-filters="updateFilterBy"
          no-shadow
          disable-selection
          hide-pagination
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loading"
              :filter-by="filterBy"
              entity-name="topology"
              :show-create-btn="false"
              :show-filter-icon-and-button="false"
            >
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
      <template #footer>
        <q-btn aid="close-topologies-preview-modal-btn" label="Close" color="primary" @click="$emit('close')" />
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
import { type IFilterBy, type IPaginationFilter } from "@/models/filter.model";
//util
import { filterService } from "@/services/filter.service/filter.service";
import { alertUtil } from "@/utils/alert.util";
import RunaiListModalSearch from "@/components/common/runai-list-modal-search/runai-list-modal-search.vue";
import { filterUtil } from "@/utils/filter.util/filter.util";
import type { ITopologyPreviewModalOptions } from "@/table-models/topology.table-model";
import { type NetworkTopology } from "@/swagger-models/cluster-service-client";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

export default defineComponent({
  name: "topologies-preview-modal",
  components: { RunaiListModalSearch, RunaiTable, RunaiBaseDialog, RunaiTableNoData },
  emits: ["close"],
  props: {
    modalOptions: {
      type: Object as PropType<ITopologyPreviewModalOptions>,
      required: true,
    },
  },
  data() {
    return {
      filterBy: {} as IFilterBy,
      topologies: [] as NetworkTopology[],
      loading: false as boolean,
    };
  },
  created() {
    const defaultFilters = filterService.getEmptyFilterByModel({
      sortBy: this.modalOptions.sortByDefault,
    });
    this.updateFilterBy(defaultFilters);
  },
  computed: {
    modalHeader(): string {
      return `Network Topologies associated with ${this.modalOptions.entityType} ${this.modalOptions.entityName}`;
    },
  },
  methods: {
    async loadTopologies(filters: IPaginationFilter): Promise<void> {
      try {
        this.topologies = await clusterService.getNetworkTopologies([
          filterUtil.getEqualsFilterString("clusterId", this.modalOptions.clusterId || ""),
          ...(filters.filterBy || []),
          ...(filters.search ? [filterUtil.getContainsFilterString("name", filters.search)] : []),
        ]);
      } catch (error) {
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load network topologies"));
      }
    },
    async loadTopologiesCount(filterBy: Array<string>, search?: string): Promise<void> {
      try {
        this.filterBy.rowsNumber = await clusterService.getNetworkTopologyCount([
          filterUtil.getEqualsFilterString("clusterId", this.modalOptions.clusterId || ""),
          ...(filterBy || []),
          ...(search ? [filterUtil.getContainsFilterString("name", search)] : []),
        ]);
      } catch (error) {
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load network topologies count"));
      }
    },
    async updateFilterBy(filterBy: IFilterBy): Promise<void> {
      this.filterBy = filterBy;
      const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy, undefined, true);
      filters.filterBy = [...(filters.filterBy || [])];
      try {
        this.setLoading(true);
        await this.loadTopologiesCount(filters.filterBy, filters.search);
        await this.loadTopologies(filters);
      } catch (error: unknown) {
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load network topologies"));
      } finally {
        this.setLoading(false);
      }
    },
    setLoading(isLoading: boolean): void {
      this.loading = isLoading;
    },
  },
});
</script>
