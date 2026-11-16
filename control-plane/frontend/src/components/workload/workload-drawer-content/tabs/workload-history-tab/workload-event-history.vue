<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-event-history" aid="workload-event-history">
    <div v-if="!showElementEmptyState" class="history-type-selector">
      <runai-select
        v-model="historyType"
        :options="workloadHistoryTypes"
        label="Type"
        aid="workload-history-type"
        outlined
        dense
        hide-bottom-space
        stack-label
        emit-value
      />
    </div>
    <div class="history-type-content">
      <workload-user-actions-history
        v-if="historyType === EWorkloadHistoryType.UserActions && !showElementEmptyState"
        :workload-id="workload.id"
      />
      <workload-history-tab
        v-else
        :workload="workload"
        :drawer-width="drawerWidth"
        :show-element-empty-state="showElementEmptyState"
      />
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import { EWorkloadHistoryType, workloadHistoryTypes, type WorkloadHistoryType } from "./workload-event-history.model";
import { RunaiSelect } from "@/components/common/runai-select";
import { WorkloadUserActionsHistory } from "@/components/workload/workload-drawer-content/tabs/workload-history-tab/workload-user-actions-history";
import WorkloadHistoryTab from "./workload-history-tab.vue";
import type { Workload } from "@/swagger-models/workloads-client";

const props = withDefaults(
  defineProps<{
    workload: Workload;
    drawerWidth: string;
    showElementEmptyState?: boolean;
    initialType?: WorkloadHistoryType;
  }>(),
  { showElementEmptyState: false, initialType: EWorkloadHistoryType.Logs },
);
const historyType = ref(props.initialType);
const emit = defineEmits<{ "update:initialType": [value: WorkloadHistoryType] }>();
watch(historyType, (value) => emit("update:initialType", value));
watch(
  () => props.workload.id,
  () => {
    historyType.value = EWorkloadHistoryType.Logs;
  },
);
watch(
  () => props.initialType,
  (value) => {
    historyType.value = value;
  },
);
</script>

<style lang="scss" scoped>
.workload-event-history {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}
.history-type-selector {
  flex-shrink: 0;
  padding: 7px 14px;
  border-bottom: 1px solid var(--border-subtle);

  .runai-select {
    width: 180px;
  }
}
.history-type-content {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
</style>
