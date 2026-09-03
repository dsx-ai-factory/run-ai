<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-history">
    <runai-drawer-no-data
      v-if="showElementEmptyState"
      :messages="elementEmptyStateMessages"
      icon-name="ph-bold ph-file-text"
    />
    <runai-default-loader v-else-if="isPageLoading" />
    <template v-else>
      <workload-phase-history-chart
        class="q-pt-md"
        :drawer-width="drawerWidth"
        :phase-update-records="phaseUpdateRecords"
      />
      <workload-event-history-table
        :filters="workloadHistoryFilterBy"
        :history-records="workloadHistoryRecords"
        @update:filter-by="updateFilterBy"
      />
    </template>
  </section>
</template>

<script lang="ts">
// cmps
import { WorkloadPhaseHistoryChart } from "@/components/workload/workload-drawer-content/tabs/workload-history-tab/workload-phase-history-chart";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";

// types
import type { PropType } from "vue";
// model
// service
import { defineComponent } from "vue";
import { type HistoryRecord, HistoryRecordType, type Workload } from "@/swagger-models/workloads-client";
import WorkloadEventHistoryTable from "@/components/workload/workload-drawer-content/tabs/workload-history-tab/workload-event-history-table/workload-event-history-table.vue";
import { useWorkloadStore } from "@/stores/workload.store";
import type { IJobFilter } from "@/models/filter.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { alertUtil } from "@/utils/alert.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";

export default defineComponent({
  components: {
    WorkloadEventHistoryTable,
    WorkloadPhaseHistoryChart,
    RunaiDefaultLoader,
    RunaiDrawerNoData,
  },
  emits: ["load-workload-history"],
  props: {
    drawerWidth: {
      type: String as PropType<string>,
      required: true,
    },
    workload: {
      type: Object as PropType<Workload>,
      required: true,
    },
    showElementEmptyState: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      workloadStore: useWorkloadStore(),
      isPageLoading: false as boolean,
      isWorkloadHistoryLoading: false as boolean,
    };
  },
  created() {
    this.resetComponentState();
  },
  computed: {
    elementEmptyStateMessages(): string[] {
      return [
        "Event history isn't available for this element.",
        "Event history is available at the workload level only.",
      ];
    },
    workloadHistoryFilterBy(): IJobFilter {
      return this.workloadStore.workloadHistoryFilterBy;
    },
    workloadHistoryRecords(): HistoryRecord[] {
      return this.workloadStore.getWorkloadHistoryList(true);
    },
    phaseUpdateRecords() {
      return this.workloadStore
        .getWorkloadHistoryList()
        .filter((historyRecord: HistoryRecord) => {
          return historyRecord.meta.type === HistoryRecordType.PhaseUpdate;
        })
        .sort((recordA: HistoryRecord, recordB: HistoryRecord) => {
          const creationTimeA = new Date(recordA.meta.creationTimestamp).getTime();
          const creationTimeB = new Date(recordB.meta.creationTimestamp).getTime();
          return creationTimeA - creationTimeB;
        });
    },
  },
  methods: {
    async resetComponentState(): Promise<void> {
      if (this.showElementEmptyState) {
        intervalUtil.stopInterval(EIntervalLabels.WorkloadHistoryTab);
        return;
      }

      this.isPageLoading = true;
      try {
        intervalUtil.stopInterval(EIntervalLabels.WorkloadHistoryTab);

        await this.loadWorkloadHistory();
        intervalUtil.startInterval(EIntervalLabels.WorkloadHistoryTab, async () => {
          await this.loadWorkloadHistory();
        });
      } finally {
        this.isPageLoading = false;
      }
    },
    updateFilterBy(filterBy: IJobFilter): void {
      this.workloadStore.setWorkloadHistoryFilterBy(filterBy);
    },
    async loadWorkloadHistory(): Promise<void> {
      if (this.isWorkloadHistoryLoading) return;

      try {
        this.isWorkloadHistoryLoading = true;
        await this.workloadStore.loadWorkloadHistory(this.workload.id, this.workload.clusterId);
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError("Failed to load workload history"));
        }
      } finally {
        this.isWorkloadHistoryLoading = false;
      }
    },
  },
  watch: {
    "workload.id": {
      handler(): void {
        this.resetComponentState();
      },
    },
    showElementEmptyState(): void {
      this.resetComponentState();
    },
  },
  unmounted() {
    intervalUtil.stopInterval(EIntervalLabels.WorkloadHistoryTab);
  },
});
</script>
