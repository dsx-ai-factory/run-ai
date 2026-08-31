<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td>
    <div class="row justify-between items-center">
      <div class="url-text ellipsis">
        {{ urlText }}
      </div>
      <div class="flex gap-15">
        <runai-copy-to-clipboard-btn
          :text-to-copy="urlText"
          :custom-tooltip="copyTooltipText"
          :disable="!row.url"
        ></runai-copy-to-clipboard-btn>
        <runai-tooltip
          v-if="isServingPort"
          class="connect-btn"
          :tooltip-text="servingPortTooltipText"
          icon="ph-bold ph-info"
          tooltip-position="bottom"
        ></runai-tooltip>
        <q-btn
          v-else
          class="connect-btn runai-icon-btn--default"
          flat
          round
          icon="ph-bold ph-arrow-square-out"
          @click="connect"
          :disable="isDisabled"
          v-permission="{ resourceType: ResourceType.Workloads, action: Action.Create }"
        >
          <q-tooltip v-if="connectTooltipText" class="tool-tip-dark">{{ connectTooltipText }} </q-tooltip>
        </q-btn>
      </div>
    </div>
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// components
import { RunaiCopyToClipboardBtn } from "@/components/common/runai-copy-to-clipboard-btn";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
// models
import type { ITableColumn } from "@/models/table.model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { IWorkloadConnectionsModalOptions } from "@/models/workload.model";
import { InternalConnectionType } from "@/swagger-models/assets-service-client";
import { Phase } from "@/swagger-models/workloads-client";

// utils
import { copyToClipboard } from "quasar";
import { getFieldContent } from "@/utils/table-format.util";
import { alertUtil } from "@/utils/alert.util";
import { openNewWindow } from "@/utils/common.util";

export default defineComponent({
  components: {
    RunaiCopyToClipboardBtn,
    RunaiTooltip,
  },
  emits: ["custom-cell-event"],
  inject: ["workloadDetails"],
  props: {
    row: {
      type: Object as PropType<any>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
  },
  data() {
    return {
      ResourceType,
      Action,
    };
  },
  computed: {
    urlText(): string {
      return getFieldContent(this.column, this.row) || "";
    },
    servingPortTooltipText(): string {
      if (!this.isServingPort) return "";
      return this.row.url.includes(".svc.cluster.local")
        ? "This URL is for internal cluster use and is not accessible externally."
        : "";
    },
    copyTooltipText(): string {
      if (!this.row.url) return "Connection address is not available";
      return "Copy address to clipboard";
    },
    connectTooltipText(): string {
      if (!this.row.url) return "Connection address is not available";
      if (!this.isAllowed) return "Only the workload creator can connect to this tool";
      if (!this.isWorkloadRunning) return "To connect, the selected workload must be running";

      return "Connect";
    },
    isWorkloadRunning(): boolean {
      if (this.row.workloadPhase) {
        return this.row.workloadPhase === Phase.Running;
      }
      return !!(this.workloadDetails as IWorkloadConnectionsModalOptions["workloadDetails"])?.isRunning;
    },
    isDisabled(): boolean {
      return !this.row.url || this.isServingPort || !this.isWorkloadRunning || !this.isAllowed;
    },
    isServingPort(): boolean {
      return this.row.connectionType === InternalConnectionType.ServingPort;
    },
    isAllowed(): boolean {
      if (!this.row.authorizedUsers?.length) return true;
      const currentUser = (this.workloadDetails as IWorkloadConnectionsModalOptions["workloadDetails"])?.currentUser;
      return this.row.authorizedUsers.includes(currentUser);
    },
  },
  methods: {
    connect(): void {
      if (!this.workloadDetails) return;
      const url: string = this.urlText;
      openNewWindow(url);
    },
    async copyText(): Promise<void> {
      await copyToClipboard(this.urlText);
      this.$q.notify(alertUtil.getSuccess("Copied!"));
    },
  },
});
</script>
<style lang="scss" scoped>
.connect-btn {
  color: $black-54;
}
.url-text {
  flex: 1;
}
</style>
