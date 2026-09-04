<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section>
    <runai-base-dialog :model-value="true" @close="$emit('close')" size="width-sm" no-padding>
      <template #header> {{ modalOptions.header }} </template>

      <template #body>
        <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search service" />
        <q-separator />
        <runai-table
          :columns="modalOptions.columns"
          :rows="modalOptions.services"
          :loading="modalOptions.loading"
          :filter-by="tableFilter"
          @update-filters="updateFilter"
          modal-view
          disable-selection
          hide-pagination
        >
          <template #no-data>
            <runai-table-no-data
              :filter-by="tableFilter"
              :entity-name="entityName"
              :show-create-btn="false"
              :show-filter-icon-and-button="false"
            />
          </template>
        </runai-table>
      </template>

      <template #footer>
        <runai-btn label="Close" color="primary" aid="close-services-modal-btn" @click="$emit('close')" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiListModalSearch } from "@/components/common";
import { RunaiTable } from "@/components/common/runai-table";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiBtn } from "@/components/common/runai-btn";

// models
import type { IFilterBy } from "@/models/filter.model";
//service
import { filterService } from "@/services/filter.service/filter.service";
import type { IServiceTableModalOptions } from "@/models/service.model";
import { allServiceColumnsMap } from "@/table-models/services.table-model";

export default defineComponent({
  components: {
    RunaiTableNoData,
    RunaiTable,
    RunaiBaseDialog,
    RunaiListModalSearch,
    RunaiBtn,
  },
  emits: ["close"],
  props: {
    modalOptions: {
      type: Object as PropType<IServiceTableModalOptions>,
      required: true,
    },
    entityName: {
      type: String as PropType<string>,
      default: "service",
    },
  },
  data() {
    return {
      tableFilter: {} as IFilterBy,
      search: "" as string,
    };
  },
  created() {
    this.loadFilters();
  },
  methods: {
    loadFilters(): void {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        allServiceColumnsMap.name.name,
        this.modalOptions.columns,
      );
      this.tableFilter = filterService.getEmptyFilterByModel(defaultFilters);
    },
    updateFilter(filter: IFilterBy): void {
      this.tableFilter = filter;
    },
  },
});
</script>
