<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-pods-tab column no-wrap">
    <runai-list-modal-search
      v-model="filterBy.searchTerm"
      placeholder="Search pods"
      :debounce="300"
      aid="structure-pods-search"
    />
    <q-separator />
    <runai-table
      class="pods-table"
      :rows="pods"
      :columns="columns"
      :loading="loading"
      :filter-by="filterBy"
      :get-row-key="getRowKey"
      :bordered="false"
      modal-view
      disable-selection
      :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25]"
      @update-filters="onFiltersUpdate"
    >
      <template #no-data>
        <runai-table-no-data
          v-if="!loading"
          :filter-by="filterBy"
          entity-name="pod"
          :show-create-btn="false"
          :show-filter-icon-and-button="false"
        />
      </template>
    </runai-table>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { RunaiTable } from "@/components/common/runai-table";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiListModalSearch } from "@/components/common";
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import type { WorkloadServicePod } from "@/models/workload.model";
import { allWorkloadPodsModalColumnsMap } from "@/table-models/workload.table-model";

interface Props {
  pods: WorkloadServicePod[];
  loading?: boolean;
}

withDefaults(defineProps<Props>(), {
  loading: false,
});

// rowsPerPage drives client-side pagination. We deliberately do NOT set rowsNumber:
// when rowsNumber is set, q-table switches to server-side mode and stops applying
// the :filter prop, breaking the search input above.
const filterBy = ref<IFilterBy>({
  rowsPerPage: 10,
  searchTerm: "",
});

const columns = computed<ITableColumn[]>(() => [
  { ...allWorkloadPodsModalColumnsMap.name, display: true },
  { ...allWorkloadPodsModalColumnsMap.status, display: true },
  { ...allWorkloadPodsModalColumnsMap.nodeName, display: true },
  { ...allWorkloadPodsModalColumnsMap.nodePools, display: true },
  { ...allWorkloadPodsModalColumnsMap.gpuComputeAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.gpuMemoryAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.cpuAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.cpuMemoryAllocation, display: true },
  { ...allWorkloadPodsModalColumnsMap.image, display: true },
  { ...allWorkloadPodsModalColumnsMap.ip, display: true },
  { ...allWorkloadPodsModalColumnsMap.created, display: true },
]);

const getRowKey = (pod: WorkloadServicePod): string => pod.id || pod.name;

// Q-table emits pagination/sort changes via runai-table's @update-filters. We must
// reassign filterBy so the rows-per-page selector and any other in-table filters
// actually take effect — otherwise the dropdown updates visually but the table
// keeps using the original rowsPerPage.
const onFiltersUpdate = (next: IFilterBy): void => {
  filterBy.value = next;
};
</script>

<style scoped lang="scss">
.workload-pods-tab {
  height: 100%;
  overflow: auto;
}

// We deliberately do NOT use `flex: 1` on the table. Letting the table size to its
// own content keeps the layout stable when the search filter produces zero matches:
// the body collapses naturally and the "No pods match your filters" message renders
// right below the header instead of being pushed to the bottom of the panel.
.pods-table {
  min-height: 0;
}
</style>
