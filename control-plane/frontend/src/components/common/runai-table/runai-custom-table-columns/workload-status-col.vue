<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td :style="statusOptions.colStyle">
    <div class="row items-center no-wrap">
      <div class="status-container" :aid="`${row.name}-status`">
        <runai-status :options="statusOptions" />
      </div>
      <q-btn
        @click.stop="openMenu"
        v-if="showExtendedMenu"
        class="q-ml-sm q-pa-none runai-icon-btn--default text-icon-default"
        icon="ph-bold ph-info"
        flat
        round
      >
        <workload-pending-scheduling-messages-menu
          v-if="isPendingSchedulingMessagesMenuOpen"
          @hide="isPendingSchedulingMessagesMenuOpen = false"
          :workload-id="workloadId"
        />
        <workload-phase-messages-menu
          v-if="isPhaseMessagesMenuOpen"
          :columns="allWorkloadPhaseMessageColumnsMap"
          :messages="phaseMessages"
        />
      </q-btn>
    </div>
  </q-td>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

// models
import type { IStatusColOptions, ITableColumn } from "@/models/table.model";
import { RunaiStatus } from "@/components/common/runai-status/";
import { WorkloadPendingSchedulingMessagesMenu } from "@/components/workload/workload-pending-scheduling-messages-menu";
import { Phase, type Workload } from "@/swagger-models/workloads-client";
import { WorkloadPhaseMessagesMenu } from "@/components/workload/workload-phase-messages-menu";
import { allWorkloadPhaseMessageColumnsMap, type Field } from "@/models/workload-phase-messages.model";
import { workloadUtil } from "@/utils/workload.util";
import { tableFormatUtil } from "@/utils/table-format.util/table-format.util";

export default defineComponent({
  name: "workload-status-col",
  components: { WorkloadPhaseMessagesMenu, WorkloadPendingSchedulingMessagesMenu, RunaiStatus },
  emits: ["custom-cell-event"],
  props: {
    row: {
      type: Object as PropType<Workload>,
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
  },
  data() {
    return {
      isPendingSchedulingMessagesMenuOpen: false,
      isPhaseMessagesMenuOpen: false,
      phaseMessages: [] as Field[],
    };
  },
  computed: {
    statusOptions(): IStatusColOptions {
      return tableFormatUtil.getFieldContent(this.column, this.row);
    },
    workloadId(): string {
      return this.row.id;
    },
    showExtendedMenu(): boolean {
      if (this.isPendingPhase) {
        return !!this.row.phaseReason;
      }
      return this.shouldShowIssuesMenu;
    },
    isPendingPhase(): boolean {
      return this.row.phase === Phase.Pending;
    },
    isCompletedPhase(): boolean {
      return this.row.phase === Phase.Completed;
    },
    shouldShowIssuesMenu(): boolean {
      return !this.isPendingPhase && !this.isCompletedPhase && !!this.phaseIssues;
    },
    allWorkloadPhaseMessageColumnsMap(): ITableColumn[] {
      return allWorkloadPhaseMessageColumnsMap;
    },
    phaseIssues(): string {
      return this.row.additionalFields?.["issues"] ?? "";
    },
  },
  methods: {
    openMenu(): void {
      if (this.isPendingPhase) {
        this.isPendingSchedulingMessagesMenuOpen = true;
      } else if (this.shouldShowIssuesMenu) {
        this.phaseMessages = workloadUtil.flattenWorkloadPhaseIssues(this.phaseIssues);
        this.isPhaseMessagesMenuOpen = true;
      }
    },
  },
});
</script>
<style scoped lang="scss">
.status-container {
  min-width: 130px;
}
</style>
