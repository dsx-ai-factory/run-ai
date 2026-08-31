<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <workload-pod-container-selector
    :pods="pods"
    :selected-pod-name="selectedPodName"
    :selected-container-name="selectedContainerName"
    :show-container-select="supportClusterApiLogs"
    @pod-selected="setSelectedPod"
    @container-selected="setSelectedContainer"
  >
    <template v-if="supportClusterApiLogs" #extra-selectors>
      <runai-select
        outlined
        label="Container's logs from"
        class="logs-source-select"
        dense
        hide-bottom-space
        options-dense
        :options="logsSourceOptions"
        :model-value="selectedLogsSource"
        emit-value
        map-options
        @update:model-value="setSelectedLogsSource"
      />
    </template>

    <template #actions>
      <runai-tooltip-wrapper disable-break-line :tooltip-text="downloadTooltipText">
        <q-btn
          class="download-button q-ma-sm runai-icon-btn--default"
          icon="ph-bold ph-download-simple"
          :disable="downloadDisabled"
          flat
          round
          @click="downloadLogs"
        >
        </q-btn>
      </runai-tooltip-wrapper>
    </template>
  </workload-pod-container-selector>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import { ELogsSource, type ILogParams } from "@/models/logs.model";
import type { WorkloadServicePod } from "@/models/workload.model";
import type { ISelectOption } from "@/models/global.model";

import type { AxiosResponse } from "axios";

// services
import { researcherService } from "@/services/cluster/researcher.service/researcher.service";
import { createDownloadLink } from "@/utils/common.util";

// Components
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { WorkloadPodContainerSelector } from "@/components/common/workload-pod-container-selector";
import { RunaiSelect } from "@/components/common/runai-select";

// utils
import { alertUtil } from "@/utils/alert.util";
import { clusterApiService } from "@/services/cluster/cluster-api.service/cluster-api.service";

const defaultFilename = "logs.txt";

const LOGS_SOURCE_OPTIONS: ISelectOption[] = [
  { label: "Current instance", value: ELogsSource.CurrentInstance },
  { label: "Previous instance", value: ELogsSource.PreviousInstance },
];

export default defineComponent({
  components: { RunaiTooltipWrapper, WorkloadPodContainerSelector, RunaiSelect },
  emits: ["pod-selected", "container-selected", "logs-source-selected"],
  props: {
    pods: {
      type: Array as PropType<Array<WorkloadServicePod>>,
      required: true,
    },
    disableExport: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    logParams: {
      type: Object as PropType<ILogParams>,
      required: true,
    },
    supportClusterApiLogs: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    selectedLogsSource: {
      type: String as PropType<ELogsSource>,
      default: ELogsSource.CurrentInstance,
    },
    isClusterConnected: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  computed: {
    logsSourceOptions(): ISelectOption[] {
      return LOGS_SOURCE_OPTIONS;
    },
    downloadDisabled(): boolean {
      return this.disableExport || !this.isClusterConnected;
    },
    downloadTooltipText(): string {
      if (!this.isClusterConnected) {
        return "To download as text your cluster needs to be connected.";
      }
      if (this.disableExport) {
        return "There are no logs to download for this workload.";
      }
      return "Download as text";
    },
    selectedPodName(): string {
      return this.logParams.podName;
    },
    selectedPod(): WorkloadServicePod | undefined {
      return this.pods.find((pod: WorkloadServicePod) => pod.name === this.selectedPodName);
    },
    selectedContainerName(): string {
      if (!this.selectedPod?.containers) return "";
      const containerNames = this.selectedPod.containers.map((container) => container.name);
      return this.logParams.container || (containerNames.length > 0 ? containerNames[0] : "");
    },
  },
  methods: {
    setSelectedPod(podName: string): void {
      this.$emit("pod-selected", podName);
    },
    setSelectedContainer(containerName: string): void {
      this.$emit("container-selected", containerName);
    },
    setSelectedLogsSource(logsSource: ELogsSource): void {
      this.$emit("logs-source-selected", logsSource);
    },
    extractFilenameFromHeaders(headers: AxiosResponse["headers"]): string {
      const contentDisposition = headers["content-disposition"];

      if (contentDisposition?.includes("attachment")) {
        const splitHeader = contentDisposition.split("=");
        if (splitHeader?.length === 2) {
          return splitHeader[1].replace(/^["']|["']$/g, "");
        }
      }
      return defaultFilename;
    },
    async downloadLogs(): Promise<void> {
      if (!this.logParams) return;
      try {
        let logsBlob: Blob;
        let filename: string = defaultFilename;
        if (this.supportClusterApiLogs) {
          const logsResponse = await clusterApiService.downloadLogsClusterApi(this.logParams);
          logsBlob = logsResponse.data;
          filename = this.extractFilenameFromHeaders(logsResponse.headers);
        } else {
          logsBlob = await researcherService.downloadLogs(this.logParams);
        }
        createDownloadLink(logsBlob, filename);
      } catch (err: unknown) {
        this.$q.notify(alertUtil.getError("Failed to download logs"));
      }
    },
  },
});
</script>
<style lang="scss" scoped>
.logs-source-select {
  min-width: 180px;
  max-width: 260px;
}
</style>
