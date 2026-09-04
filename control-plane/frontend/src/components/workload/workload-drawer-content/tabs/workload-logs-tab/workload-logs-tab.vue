<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-logs-tab">
    <workload-logs-header
      :pods="pods"
      :log-params="logParams"
      :support-cluster-api-logs="isSupportClusterApiLogs"
      :disable-export="showDrawerNoData"
      :is-cluster-connected="isClusterConnected"
      :selected-logs-source="logsSource"
      @pod-selected="handlePodSelected"
      @container-selected="handleContainerSelected"
      @logs-source-selected="handleLogsSourceSelected"
    />
    <section class="workload-logs-body">
      <runai-default-loader v-if="showLoader" />

      <runai-drawer-no-data v-if="showDrawerNoData" :messages="drawerNoDataMessages" :icon-name="drawerNoDataIcon" />

      <runai-log-reader
        v-else-if="showLogger"
        :cluster-id="workload.clusterId"
        :log-params="logParams"
        :support-cluster-api-logs="isSupportClusterApiLogs"
        @error="handleLoggerError"
        @connecting="setLoggerConnectingState"
        @no-data="noData = $event"
        @close="handleLoggerClose"
      />
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { type PropType } from "vue";

// Components
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import { WorkloadLogsHeader } from "./workload-logs-header";
import { ELogsSource } from "@/models/logs.model";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";
import { RunaiLogReader } from "@/components/common/runai-log-reader";

// models
import type { ILogParams } from "@/models/logs.model";
import { EWorkloadPodPhase, type WorkloadServicePod } from "@/models/workload.model";
import type { Workload } from "@/swagger-models/workloads-client";
import { Phase } from "@/swagger-models/assets-service-client";

// stores
import { useClusterStore } from "@/stores/cluster.store";

// services
import { workloadService } from "@/services/cluster/workload.service/workload.service";

// constants
import { MIN_CLUSTER_VERSION_FOR_WORKLOAD_CLUSTER_API_LOGS } from "@/common/version.constant";

const TAIL_LENGTH = 1000;

export default defineComponent({
  components: { WorkloadLogsHeader, RunaiDefaultLoader, RunaiDrawerNoData, RunaiLogReader },
  emits: ["pod-selected", "container-selected"],
  props: {
    workload: {
      type: Object as PropType<Workload>,
      required: true,
    },
    pods: {
      type: Array as PropType<WorkloadServicePod[]>,
      required: true,
    },
    loadingPods: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    selectedPodName: {
      type: String as PropType<string>,
      default: "",
    },
    selectedContainerName: {
      type: String as PropType<string>,
      default: "",
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      logParams: this.getDefaultLogParams(),
      logsSource: ELogsSource.CurrentInstance as ELogsSource,
      error: false as boolean,
      loadingLogs: true as boolean,
      noData: false as boolean,
      connectionClosed: false as boolean,
      hasLogsAfterClose: false as boolean,
    };
  },
  computed: {
    isSupportClusterApiLogs(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workload.clusterId,
        MIN_CLUSTER_VERSION_FOR_WORKLOAD_CLUSTER_API_LOGS,
      );
    },
    isClusterConnected(): boolean {
      const cluster = this.clusterStore.getClusterById(this.workload.clusterId);
      if (!cluster) return false;
      return workloadService.canSubmitWorkload(cluster) === null;
    },
    showLogs(): boolean {
      const isRelevantPhaseForLogs = [
        Phase.Running,
        Phase.Failed,
        Phase.Completed,
        // Initializing and Degraded can have pods with restarted containers, allowing previous instance logs
        Phase.Initializing,
        Phase.Degraded,
      ].some((phase: Phase) => this.workload.phase === phase);

      return (
        isRelevantPhaseForLogs || this.pods.some((pod: WorkloadServicePod) => pod.k8sPhase === EWorkloadPodPhase.Running)
      );
    },
    showDrawerNoData(): boolean {
      // If pods finished loading and there are no pods, show no data immediately
      // (don't wait for loadingLogs since log reader won't mount without pods)
      const noPodsAfterLoading = !this.loadingPods && !this.podsExist;

      if (this.loadingLogs && !noPodsAfterLoading) return false;

      // If connection closed and we have logs, don't show no-data drawer
      if (this.connectionClosed && this.hasLogsAfterClose) {
        return false;
      }

      return !this.showLogs || noPodsAfterLoading || this.error || this.noData;
    },
    drawerNoDataMessages(): string[] {
      if (this.error) return ["Logs aren't available right now.", "If this issue persists, contact your administrator."];
      else if (!this.showLogs) return [`This workload has no logs because it is not active.`];
      return [`There are no logs yet.`];
    },
    drawerNoDataIcon(): string {
      if (this.error || !this.showLogs) return "ph-bold ph-file-minus";
      return "ph-bold ph-file";
    },
    podsExist(): boolean {
      return this.pods.length > 0;
    },
    showLoader(): boolean {
      // Don't show loader if pods finished loading and there are no pods
      // (log reader won't mount, so we'd be stuck loading forever)
      if (!this.loadingPods && !this.podsExist) {
        return false;
      }
      return this.loadingPods || this.loadingLogs;
    },
    showLogger(): boolean {
      return (this.podsExist && !this.error) || (this.connectionClosed && this.hasLogsAfterClose);
    },
  },
  methods: {
    getDefaultLogParams(): ILogParams {
      return {
        project: this.workload.projectName || "",
        jobName: this.workload.name || "",
        podName: "",
        tail: TAIL_LENGTH,
        type: this.workload.type,
        distributedFramework: this.workload.distributedFramework,
        container: undefined,
        externalWorkloadName: this.workload.externalWorkloadName,
      };
    },
    resetLogState(): void {
      this.error = false;
      this.loadingLogs = true; // Show loading while reconnecting
      this.noData = false;
      this.connectionClosed = false;
      this.hasLogsAfterClose = false;
    },
    isPreviousInstance(): boolean {
      return this.logsSource === ELogsSource.PreviousInstance;
    },
    updateLogParamsFromSelection(): void {
      // Update log params when pod/container selection changes from parent
      if (this.selectedPodName && this.selectedPodName !== this.logParams.podName) {
        // Reset log state when pod changes to allow reconnection
        this.resetLogState();
        this.logParams = {
          project: this.workload.projectName,
          jobName: this.workload.name,
          podName: this.selectedPodName,
          tail: TAIL_LENGTH,
          type: this.workload.type,
          distributedFramework: this.workload.distributedFramework,
          container: this.selectedContainerName || undefined,
          externalWorkloadName: this.workload.externalWorkloadName,
          previous: this.isPreviousInstance(),
        };
      } else if (this.selectedContainerName && this.selectedContainerName !== this.logParams.container) {
        // Reset log state when container changes to allow reconnection
        this.resetLogState();
        this.logParams = {
          ...this.logParams,
          container: this.selectedContainerName,
        };
      }
    },
    handlePodSelected(podName: string): void {
      this.$emit("pod-selected", podName);
    },
    handleContainerSelected(containerName: string): void {
      this.$emit("container-selected", containerName);
    },
    handleLogsSourceSelected(logsSource: ELogsSource): void {
      this.logsSource = logsSource;
      this.resetLogState();
      this.logParams = {
        ...this.logParams,
        previous: logsSource === ELogsSource.PreviousInstance,
      };
    },
    setLoggerConnectingState(isConnecting: boolean): void {
      this.loadingLogs = isConnecting;
    },
    handleLoggerError(hasError: boolean): void {
      // Only set error if we don't have logs to preserve
      if (this.hasLogsAfterClose) return;
      this.error = hasError;
    },
    handleLoggerClose(hasLogs: boolean): void {
      this.connectionClosed = true;
      this.loadingLogs = false;

      if (hasLogs) {
        this.hasLogsAfterClose = true;
        this.error = false;
      } else {
        this.hasLogsAfterClose = false;
      }
    },
  },
  watch: {
    workload: {
      handler(newWorkload: Workload, prevWorkload: Workload): void {
        if (newWorkload.id === prevWorkload.id && newWorkload.phase === prevWorkload.phase) return;

        this.resetLogState();
      },
      deep: true,
    },
    selectedPodName: {
      handler(): void {
        this.updateLogParamsFromSelection();
      },
      immediate: true,
    },
    selectedContainerName: {
      handler(): void {
        this.updateLogParamsFromSelection();
      },
    },
  },
});
</script>
<style lang="scss" scoped>
.workload-logs-tab {
  background-color: $white;
  height: 100%;
  display: flex;
  flex-direction: column;

  .workload-logs-header {
    width: 100%;
    font-size: 16px;
    font-weight: 500;
    border-bottom: 1px solid $black-10;
  }

  .workload-logs-body {
    color: $white;
    flex: 1;
    background-color: var(--bg-code-surface);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
}
</style>
