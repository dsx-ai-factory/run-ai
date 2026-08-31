<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-menu
    @hide="$emit('hide')"
    model-value
    aid="workload-pending-scheduling-messages-menu"
    anchor="bottom middle"
    self="top middle"
  >
    <div class="q-pa-md">The workload is in queue and waiting to be scheduled, due to the following reasons:</div>
    <q-separator />
    <q-table table-class="runai-table" :columns="columns" :rows="messages" :loading="loading" hide-bottom flat>
      <template v-slot:body="props">
        <q-tr :props="props">
          <q-td v-for="col in columns" :key="col.name" :props="props">
            <div v-dompurify-html="replaceNewLineWithBreaks(props.row[col.name])"></div>
          </q-td>
        </q-tr>
      </template>
    </q-table>
  </q-menu>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

//model
import { EWorkloadErrorMessage } from "@/models/workload.model";
import { allPendingSchedulingMessageColumnsMap } from "@/table-models/workload.table-model";
import type { ITableColumn } from "@/models/table.model";
import type { PendingSchedulingMessage } from "@/swagger-models/workloads-client";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { stringUtil } from "@/utils/string.util";

export default defineComponent({
  name: "workload-pending-scheduling-messages-pop-over",
  emits: ["hide"],
  props: {
    workloadId: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      loading: true as boolean,
      messages: [] as PendingSchedulingMessage[],
    };
  },
  created() {
    this.displayPendingSchedulingMessagesPopOver();
  },
  computed: {
    columns(): ITableColumn[] {
      return allPendingSchedulingMessageColumnsMap;
    },
  },
  methods: {
    replaceNewLineWithBreaks(text: string): string {
      return stringUtil.replaceNewLineWithBreaks(text);
    },
    async displayPendingSchedulingMessagesPopOver(): Promise<void> {
      try {
        this.messages = await workloadService.getWorkloadPendingSchedulingMessages(this.workloadId);
      } catch (error: unknown) {
        this.handlePendingSchedulingMessagesError(error);
      } finally {
        this.loading = false;
      }
    },
    handlePendingSchedulingMessagesError(error: unknown): void {
      const errorAlert = new ErrorAlert({
        generalMessage: EWorkloadErrorMessage.FailedToLoadPendingSchedulingMessages,
      });
      this.$q.notify(errorAlert.getNotification(error));
    },
  },
});
</script>

<style scoped lang="scss"></style>
