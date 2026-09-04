<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="topology-labels-modal">
    <runai-base-dialog model-value @close="$emit('close')" size="width-md" :no-padding="true">
      <template #header>Network Topology {{ modalOptions.topologyName || topology?.name || "" }}</template>

      <template #body>
        <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search labels" />
        <q-separator />
        <runai-table
          :rows="labels"
          :columns="columns"
          :loading="loading"
          :filter-by="tableFilter"
          @update-filters="updateFilter"
          no-shadow
          modal-view
          disable-selection
          hide-pagination
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loading"
              :filter-by="tableFilter"
              entity-name="label"
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
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";

// models
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import type { NetworkTopology } from "@/swagger-models/cluster-service-client";
import type { ITopologyLabelPreview } from "@/table-models/topology.table-model";
import {
  nodePoolTopologyLabelsTableColumns,
  type ITopologyLabelsModalOptions,
} from "@/table-models/topology.table-model";

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
      type: Object as PropType<ITopologyLabelsModalOptions>,
      required: true,
    },
  },
  data() {
    return {
      loading: false as boolean,
      tableFilter: {} as IFilterBy,
      columns: nodePoolTopologyLabelsTableColumns as Array<ITableColumn>,
      topology: null as NetworkTopology | null,
    };
  },
  created() {
    this.setFilter();
    this.loadTopologyById();
  },
  computed: {
    labels(): Array<ITopologyLabelPreview> {
      return this.topology?.levels.map((level) => ({ name: level })) || [];
    },
  },
  methods: {
    setFilter(): void {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters("", nodePoolTopologyLabelsTableColumns);
      this.tableFilter = { ...filterService.getEmptyFilterByModel(defaultFilters), sortBy: null };
    },
    async loadTopologyById(): Promise<void> {
      try {
        this.loading = true;
        this.topology = await clusterService.getNetworkTopologyById(this.modalOptions.topologyId);
      } catch (error: unknown) {
        console.error(error);
        this.$q.notify(alertUtil.getError("Failed to load topology"));
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
