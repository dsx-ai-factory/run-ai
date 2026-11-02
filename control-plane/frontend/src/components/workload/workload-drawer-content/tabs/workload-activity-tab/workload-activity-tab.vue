<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-activity-tab" aid="workload-activity-tab" aria-label="Workload activity">
    <div class="activity-filters">
      <runai-advanced-filters
        :filter-by="tableFilters"
        :initial-filters-model="initialFilters"
        :filter-options="workloadActivityFilterOptions"
        hide-cluster-filter
        @update-filter-by="updateFilters"
      />
      <runai-general-search
        :key="workloadId"
        :search-term="tableFilters.searchTerm || ''"
        tooltip-text="Search events and subjects"
        @update-search-term="updateFilters({ searchTerm: $event })"
      />
    </div>
    <runai-default-loader
      v-if="loading && !hasLoaded"
      aid="workload-activity-loading"
      role="status"
      aria-label="Loading activity"
    />
    <template v-else>
      <div v-if="error" class="activity-error" role="alert">
        <span>Activity could not be loaded.</span>
        <span v-if="hasLoaded">Showing the last loaded activity.</span>
        <runai-btn
          aid="workload-activity-retry"
          flat
          color="primary"
          label="Retry"
          :disable="loading || refreshing"
          @click="retry"
        />
      </div>
      <runai-table-wrapper
        v-if="hasLoaded"
        class="activity-table-wrapper"
        :filters-object="tableFilters"
        :sticky="false"
        role="region"
        aria-label="Workload activity history"
        tabindex="0"
      >
        <template #table>
          <runai-table
            :rows="workloadStore.workloadActionRecords"
            :columns="workloadActivityColumns"
            :filter-by="tableFilters"
            :get-row-key="getRowKey"
            @update-filters="updateFilters"
            :loading="loading"
            :rows-per-page-options="[0]"
            disable-selection
            hide-pagination
            no-shadow
            :bordered="false"
            aria-label="Workload activity history"
          >
            <template #no-data>
              <runai-table-no-data
                :filter-by="tableFilters"
                entity-name="activity"
                custom-message="No activity has been recorded for this workload."
                custom-no-filter-results-message="No results match your filters."
                @clear-filters="updateFilters({ advancedFilters: [], searchTerm: '' })"
                :show-create-btn="false"
                ph-icon-name="ph-bold ph-file"
                icon-size="44"
                light-icon
              >
                <template #secondary-text>Older activity may no longer be available.</template>
              </runai-table-no-data>
            </template>
          </runai-table>
        </template>
      </runai-table-wrapper>
      <nav v-if="hasLoaded" class="activity-pagination" aria-label="Activity pagination">
        <span class="activity-page-size">50 per page</span>
        <span aid="workload-activity-range" aria-live="polite">{{ pageRange }}</span>
        <runai-btn
          aid="workload-activity-previous"
          icon-button
          icon="ph-bold ph-caret-left"
          aria-label="Previous activity page"
          tooltip="Previous page"
          :disable="page === 0 || loading || refreshing"
          @click="loadPage(page - 1, pageOffsets[page - 1])"
        />
        <runai-btn
          aid="workload-activity-next"
          icon-button
          icon="ph-bold ph-caret-right"
          aria-label="Next activity page"
          tooltip="Next page"
          :disable="workloadStore.workloadActionsNext == null || loading || refreshing"
          @click="loadNextPage"
        />
      </nav>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { RunaiBtn } from "@/components/common/runai-btn";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import { RunaiGeneralSearch } from "@/components/common/runai-general-search";
import { RunaiAdvancedFilters } from "@/components/common/runai-advanced-filters";
import type { WorkloadActionsQuery } from "@/services/cluster/workload.service/workload.service";
import { filterService } from "@/services/filter.service/filter.service";
import { RunaiTable } from "@/components/common/runai-table";
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { EIntervalLabels } from "@/models/interval.model";
import type { IFilterBy } from "@/models/filter.model";
import type { WorkloadActionRecord } from "@/swagger-models/workloads-client";
import { useWorkloadStore } from "@/stores/workload.store";
import { deepCopy } from "@/utils/common.util";
import { intervalUtil } from "@/utils/interval.util";
import { workloadActivityColumns, workloadActivityFilterOptions } from "./workload-activity-tab.model";

const props = defineProps<{ workloadId: string }>();
const workloadStore = useWorkloadStore();
// The API supplies a next offset, not a total count. Paginate explicitly rather
// than letting QTable paginate a single server page or invent a total.
const tableFilters = ref<IFilterBy>({
  page: 1,
  rowsPerPage: 0,
  rowsNumber: 0,
  sortBy: "acceptedAt",
  descending: true,
  advancedFilters: [],
});
const initialFilters = computed(() => deepCopy(tableFilters.value.advancedFilters || []));
const queryOptions = computed<WorkloadActionsQuery>(() => {
  const { sortBy, sortOrder, filterBy } = filterService.mapColumnsFilterToFilterParams(tableFilters.value);
  return {
    sortBy: sortBy === "action" || sortBy === "subject" ? sortBy : "acceptedAt",
    sortOrder,
    filterBy,
    ...(tableFilters.value.searchTerm ? { search: tableFilters.value.searchTerm } : {}),
  };
});
const pageOffsets = ref([0]);
const page = ref(0);
const loading = ref(false);
const refreshing = ref(false);
const hasLoaded = ref(false);
const error = ref(false);
let request = 0;
let failedPage = 0;
let failedOffset = 0;

const pageRange = computed(() => {
  const count = workloadStore.workloadActionRecords.length;
  const offset = pageOffsets.value[page.value];
  return count ? `${offset + 1}–${offset + count}` : "0 actions";
});
const getRowKey = (row: WorkloadActionRecord): string => row.id;

async function loadPage(targetPage: number, offset: number, background = false): Promise<void> {
  const current = ++request;
  failedPage = targetPage;
  failedOffset = offset;
  if (background) refreshing.value = true;
  else loading.value = true;
  error.value = false;
  try {
    await workloadStore.loadWorkloadActions(props.workloadId, offset, 50, queryOptions.value);
    if (current !== request) return;
    page.value = targetPage;
    pageOffsets.value = [...pageOffsets.value.slice(0, targetPage), offset];
    hasLoaded.value = true;
  } catch {
    if (current === request) error.value = true;
  } finally {
    if (current === request) {
      loading.value = false;
      refreshing.value = false;
    }
  }
}

function updateFilters(filters: IFilterBy): void {
  if (!filters) return;
  const next = { ...tableFilters.value, ...filters, page: 1, rowsPerPage: 0, rowsNumber: 0 };
  if (JSON.stringify(next) === JSON.stringify(tableFilters.value)) return;
  tableFilters.value = deepCopy(next);
  void resetActivity();
}

function loadNextPage(): void {
  if (workloadStore.workloadActionsNext != null) void loadPage(page.value + 1, workloadStore.workloadActionsNext);
}

function retry(): void {
  void loadPage(failedPage, failedOffset);
}

async function resetActivity(): Promise<void> {
  intervalUtil.stopInterval(EIntervalLabels.WorkloadActivityTab);
  workloadStore.clearWorkloadActions();
  pageOffsets.value = [0];
  page.value = 0;
  hasLoaded.value = false;
  refreshing.value = false;
  const initial = loadPage(0, 0);
  const current = request;
  await initial;
  if (current !== request) return;
  intervalUtil.startInterval(EIntervalLabels.WorkloadActivityTab, () => {
    // Preserve an unsuccessful navigation's Retry destination until the user
    // retries it or chooses a different page.
    if (error.value && failedPage !== page.value) return;
    if (!loading.value && !refreshing.value) return loadPage(page.value, pageOffsets.value[page.value], true);
  });
}

watch(
  () => props.workloadId,
  () => {
    tableFilters.value = {
      page: 1,
      rowsPerPage: 0,
      rowsNumber: 0,
      sortBy: "acceptedAt",
      descending: true,
      advancedFilters: [],
    };
    void resetActivity();
  },
  { immediate: true },
);
onBeforeUnmount(() => {
  request++;
  intervalUtil.stopInterval(EIntervalLabels.WorkloadActivityTab);
  workloadStore.clearWorkloadActions();
});
</script>

<style lang="scss" scoped>
.workload-activity-tab {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  color: var(--b-text-main);
}

.activity-filters {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 7px;
  padding: 7px 0;
  border-bottom: 1px solid var(--border-subtle);

  :deep(.advanced-filters) {
    flex: 1;
    min-width: 0;
  }
}

.activity-table-wrapper {
  flex: 1;
  min-height: 0;
  overflow: auto;

  :deep(table) {
    table-layout: fixed;
    width: 100%;
    min-width: 400px;
  }

  :deep(tbody td) {
    font-size: 12px;
    line-height: 16.8px;
  }

  :deep(tbody tr) {
    height: 50px;
  }

  // Keep scrolling in the named, focusable wrapper instead of Quasar's inner div.
  :deep(.q-table__middle) {
    max-height: none !important;
    overflow: visible;
  }

  :deep(.empty-state-ph-icon) {
    color: var(--icon-default);
    background: transparent;
    border-radius: 0;
    padding: 0;
  }
}

.activity-error {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 7px;
  padding: 10px 16px;
  background: var(--bg-primary-warning);
  font-size: 12px;
}

.activity-pagination {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 7px 16px;
  border-top: 1px solid var(--border-subtle);
  color: var(--b-text-sub);
  font-size: 12px;
}

.activity-page-size {
  margin-right: auto;
}
</style>
