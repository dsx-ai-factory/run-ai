<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-drawer-content>
    <template v-slot:header v-if="workload && selectedWorkloadsAmount === 1">
      <runai-svg-icon :name="getWorkloadIcon" size="34" :key="getWorkloadIcon"></runai-svg-icon>
      <div class="ellipsis">{{ workloadName }}</div>
      <runai-status v-if="statusOptions && !hideStatus" :options="statusOptions" />
    </template>

    <template v-if="workload && selectedWorkloadsAmount === 1" v-slot:subheader>
      <runai-tab v-model="drawerTab" :tabs="workloadTabs" />
    </template>

    <template v-slot:body>
      <runai-drawer-no-data
        v-if="!workload || !selectedWorkloadsAmount || selectedWorkloadsAmount > 1"
        :messages="[emptyStateMessage]"
        icon-name="ph-bold ph-file"
        icon-class="text-icon-default"
      />
      <runai-tab-panels v-else v-model="drawerTab">
        <runai-tab-panel class="q-pa-none" :name="drawerTabsOptions.EventHistory">
          <workload-history-tab
            :drawer-width="drawerWidth"
            :workload="workload"
            :show-element-empty-state="Boolean(selectedStructureElementId)"
          />
        </runai-tab-panel>
        <runai-tab-panel class="q-pa-none" :name="drawerTabsOptions.Metrics">
          <workload-metrics-tab
            :workload="workload"
            :drawer-width="drawerWidth"
            :selected-element-id="selectedStructureElementId"
          />
        </runai-tab-panel>
        <runai-tab-panel class="q-pa-none" :name="drawerTabsOptions.Pods">
          <workload-pods-tab :pods="podsToDisplay" :loading="loadingPods" />
        </runai-tab-panel>
        <runai-tab-panel class="q-pa-none" :name="drawerTabsOptions.Logs">
          <workload-logs-tab
            :workload="workload"
            :pods="podsToDisplay"
            :loading-pods="loadingPods"
            :selected-pod-name="selectedPodName"
            :selected-container-name="selectedContainerName"
            @pod-selected="handlePodSelected"
            @container-selected="handleContainerSelected"
          />
        </runai-tab-panel>
        <runai-tab-panel class="q-pa-none" :name="drawerTabsOptions.Terminal">
          <workload-terminal
            :workload="workload"
            :pods="podsToDisplay"
            :loading-pods="loadingPods"
            :selected-pod-name="selectedPodName"
            :selected-container-name="selectedContainerName"
            @pod-selected="handlePodSelected"
            @container-selected="handleContainerSelected"
          />
        </runai-tab-panel>
        <runai-tab-panel class="q-pa-none" :name="drawerTabsOptions.Yaml">
          <workload-yaml-tab :workload="workload" />
        </runai-tab-panel>
        <runai-tab-panel class="q-pa-none" :name="drawerTabsOptions.Details">
          <workload-details-tab
            v-if="workload"
            :workload="workload"
            :is-deleted-view-type="isDeletedViewType"
            :show-element-empty-state="Boolean(selectedStructureElementId)"
          />
        </runai-tab-panel>
        <runai-tab-panel v-if="!hideArchitectureTab" class="q-pa-none" :name="drawerTabsOptions.Structure">
          <workload-architecture-tab
            v-if="workload"
            :workload="workload"
            :selected-element-id="selectedStructureElementId"
            show-expand-actions
            @element-selected="handleStructureElementSelected"
            @expand-node="openExpandedStructure"
          />
        </runai-tab-panel>
      </runai-tab-panels>
    </template>
  </runai-drawer-content>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// cmps
import { RunaiDrawerContent } from "@/components/common/runai-drawer-content";
import { RunaiTab, RunaiTabPanels, RunaiTabPanel, type ITab } from "@/components/common/runai-tab";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";
import { WorkloadMetricsTab } from "@/components/workload/workload-drawer-content/tabs/workload-metrics-tab/";
import { WorkloadHistoryTab } from "@/components/workload/workload-drawer-content/tabs/workload-history-tab/";
import { WorkloadLogsTab } from "@/components/workload/workload-drawer-content/tabs/workload-logs-tab/";
import { WorkloadTerminal } from "@/components/workload/workload-drawer-content/tabs/workload-terminal/";
import { WorkloadYamlTab } from "@/components/workload/workload-drawer-content/tabs/workload-yaml-tab/";
import { WorkloadArchitectureTab } from "@/components/workload/workload-drawer-content/tabs/workload-architecture-tab/";
import { WorkloadPodsTab } from "@/components/workload/workload-drawer-content/tabs/workload-pods-tab";
// models
import { toolTypeIconsMap } from "@/common/icons.constant";
import { type Workload } from "@/swagger-models/workloads-client";
import type { IWorkloadDrawerNavigationPayload, WorkloadServicePod } from "@/models/workload.model";
import { EIntervalLabels } from "@/models/interval.model";

// utils
import { workloadUtil } from "@/utils/workload.util";
import { intervalUtil } from "@/utils/interval.util";
import { RunaiStatus } from "@/components/common/runai-status/";
import type { IStatusColOptions } from "@/models/table.model";
import { useClusterStore } from "@/stores/cluster.store";
import WorkloadDetailsTab from "@/components/workload/workload-drawer-content/tabs/workload-details-tab/workload-details-tab.vue";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
// services
import { workloadService } from "@/services/cluster/workload.service/workload.service";

export enum EWorkloadDrawerTabs {
  Details = "details",
  EventHistory = "event history",
  Metrics = "metrics",
  Pods = "pods",
  Logs = "logs",
  Terminal = "terminal",
  Yaml = "yaml",
  Structure = "structure",
}

export default defineComponent({
  components: {
    WorkloadDetailsTab,
    WorkloadArchitectureTab,
    WorkloadPodsTab,
    RunaiStatus,
    WorkloadHistoryTab,
    WorkloadLogsTab,
    RunaiSvgIcon,
    RunaiDrawerNoData,
    RunaiDrawerContent,
    WorkloadMetricsTab,
    WorkloadTerminal,
    WorkloadYamlTab,
    RunaiTab,
    RunaiTabPanels,
    RunaiTabPanel,
  },
  emits: ["navigation-handled"],
  props: {
    drawerWidth: {
      type: String as PropType<string>,
      required: true,
    },
    hideLogsTab: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideTerminalTab: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideYamlTab: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideDetailsTab: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideStatus: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideArchitectureTab: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isDeletedViewType: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    selectedWorkloadsAmount: {
      type: Number as PropType<number>,
      required: true,
    },
    workload: {
      type: [Object, null] as PropType<Workload | null>,
      required: true,
    },
    navigationPayload: {
      type: [Object, null] as PropType<IWorkloadDrawerNavigationPayload | null>,
      default: null,
    },
  },
  data() {
    return {
      drawerTabsOptions: EWorkloadDrawerTabs,
      drawerTab: null as EWorkloadDrawerTabs | null,
      clusterStore: useClusterStore(),
      // Pod state (shared between logs and terminal tabs)
      pods: [] as WorkloadServicePod[],
      loadingPods: false as boolean,
      selectedPodName: "" as string,
      selectedContainerName: "" as string,
      // Navigation state
      initialPodName: null as string | null,
      selectedStructureElementId: null as string | null,
    };
  },
  created() {
    this.drawerTab = this.drawerTabsOptions.EventHistory;
  },
  computed: {
    workloadTabs(): ITab[] {
      return [
        {
          name: this.drawerTabsOptions.EventHistory,
          label: this.drawerTabsOptions.EventHistory,
          hidden: !!this.selectedStructureElementId,
          aid: "workload-drawer-tab-event-history",
        },
        {
          name: this.drawerTabsOptions.Metrics,
          label: this.drawerTabsOptions.Metrics,
          aid: "workload-drawer-tab-metrics",
        },
        { name: this.drawerTabsOptions.Pods, label: this.drawerTabsOptions.Pods, aid: "workload-drawer-tab-pods" },
        {
          name: this.drawerTabsOptions.Logs,
          label: this.drawerTabsOptions.Logs,
          hidden: this.hideLogsTab,
          aid: "workload-drawer-tab-logs",
        },
        {
          name: this.drawerTabsOptions.Terminal,
          label: this.drawerTabsOptions.Terminal,
          hidden: this.hideTerminalTab,
          aid: "workload-drawer-tab-terminal",
        },
        {
          name: this.drawerTabsOptions.Yaml,
          label: this.drawerTabsOptions.Yaml,
          hidden: this.hideYamlTab,
          aid: "workload-drawer-tab-yaml",
        },
        {
          name: this.drawerTabsOptions.Details,
          label: this.drawerTabsOptions.Details,
          hidden: this.hideDetailsTab || !!this.selectedStructureElementId,
          aid: "workload-drawer-tab-details",
        },
        {
          name: this.drawerTabsOptions.Structure,
          label: "structure",
          hidden: this.hideArchitectureTab,
          aid: "workload-drawer-tab-structure",
        },
      ];
    },
    getWorkloadIcon(): string {
      if (!this.workload) return "";
      return workloadUtil.getWorkloadIcon(this.workload);
    },
    currentWorkloadId(): string {
      return this.workload?.id || "";
    },
    workloadName(): string {
      return this.workload?.name || "";
    },
    statusOptions(): IStatusColOptions | null {
      if (!this.workload) return null;

      return workloadUtil.getWorkloadStatusColOptions(this.workload);
    },
    workloadId(): string | null {
      return this.workload?.id || null;
    },
    // Sorted pods for display
    podsToDisplay(): WorkloadServicePod[] {
      return workloadUtil.sortPods(this.pods);
    },
    // Check if the currently active tab needs pods data
    activeTabNeedsPods(): boolean {
      return this.tabNeedsPods(this.drawerTab);
    },
    // Get selected pod object
    selectedPod(): WorkloadServicePod | undefined {
      return this.podsToDisplay.find((pod) => pod.name === this.selectedPodName);
    },
    emptyStateMessage(): string {
      if (this.selectedWorkloadsAmount > 1) {
        return "Select a single workload to view its details.";
      }
      return "Select a workload to view its details.";
    },
  },
  methods: {
    getIcon(iconName: string): string {
      return toolTypeIconsMap[iconName];
    },
    // Tabs that consume the shared pods data (and therefore need pods loaded + polled)
    tabNeedsPods(tab: EWorkloadDrawerTabs | null): boolean {
      return (
        tab === this.drawerTabsOptions.Pods ||
        tab === this.drawerTabsOptions.Logs ||
        tab === this.drawerTabsOptions.Terminal
      );
    },
    // Pod management methods
    async loadPods(showLoading = false): Promise<void> {
      if (!this.workload?.id) return;

      try {
        if (showLoading) {
          this.loadingPods = true;
        }
        const filterBy = this.selectedStructureElementId
          ? [filterUtil.getEqualsFilterString("elementId", this.selectedStructureElementId)]
          : undefined;

        this.pods = (await workloadService.getWorkloadPods(
          this.workload.id,
          false,
          undefined,
          undefined,
          undefined,
          true, // fetchAll
          filterBy,
        )) as WorkloadServicePod[];

        // Set default selection if needed
        if (this.shouldSetDefaultPod()) {
          this.setDefaultPodSelection();
        }
      } catch (error: unknown) {
        console.error("Failed to load pods:", error);
        this.pods = [];
      } finally {
        this.loadingPods = false;
      }
    },
    shouldSetDefaultPod(): boolean {
      // Need to set default if no pod selected or selected pod no longer exists
      return (
        this.podsToDisplay.length > 0 &&
        (!this.selectedPodName || !this.podsToDisplay.find((pod) => pod.name === this.selectedPodName))
      );
    },
    setDefaultPodSelection(): void {
      // If initialPodName is set and exists in pods, use it
      if (this.initialPodName) {
        const initialPod = this.podsToDisplay.find((pod) => pod.name === this.initialPodName);
        if (initialPod) {
          this.handlePodSelected(this.initialPodName);
          return;
        }
      }
      // Otherwise use first pod
      const firstPod = this.podsToDisplay[0];
      if (firstPod) {
        this.handlePodSelected(firstPod.name);
      }
    },
    handlePodSelected(podName: string): void {
      this.selectedPodName = podName;
      // Reset container to first container of new pod
      const pod = this.podsToDisplay.find((p) => p.name === podName);
      this.selectedContainerName = pod?.containers?.[0]?.name ?? "";
    },
    handleContainerSelected(containerName: string): void {
      this.selectedContainerName = containerName;
    },
    // Interval management
    startPodsInterval(): void {
      intervalUtil.stopInterval(EIntervalLabels.WorkloadPods);
      intervalUtil.startInterval(EIntervalLabels.WorkloadPods, () => this.loadPods(false));
    },
    stopPodsInterval(): void {
      intervalUtil.stopInterval(EIntervalLabels.WorkloadPods);
    },
    resetPodState(): void {
      this.stopPodsInterval();
      this.pods = [];
      this.selectedPodName = "";
      this.selectedContainerName = "";
      this.loadingPods = false;
    },
    handleStructureElementSelected(elementId: string | null): void {
      this.selectedStructureElementId = elementId;
      this.resetPodState();

      if (elementId) {
        const childIncompatibleTabs = [this.drawerTabsOptions.EventHistory, this.drawerTabsOptions.Details];
        if (this.drawerTab && childIncompatibleTabs.includes(this.drawerTab)) {
          this.drawerTab = this.drawerTabsOptions.Metrics;
        }
      }

      if (this.workload && this.activeTabNeedsPods) {
        this.loadPods(true);
        this.startPodsInterval();
      }
    },
    openExpandedStructure(elementId: string | null): void {
      if (!this.workload) return;

      const query: Record<string, string> = {
        name: this.workload.name,
        type: this.workload.type,
        phase: this.workload.phase,
        phaseUpdatedAt: this.workload.phaseUpdatedAt,
        openDrawer: "true",
        workloadId: this.workload.id,
        drawerTab: EWorkloadDrawerTabs.Structure,
      };
      if (elementId) {
        query.elementId = elementId;
      }

      this.$router.push({
        name: WORKLOAD_ROUTE_NAMES.WORKLOAD_STRUCTURE,
        params: { workloadId: this.workload.id },
        query,
      });
    },
  },
  watch: {
    workload: {
      handler(newWorkload: Workload | null, prevWorkload: Workload | null): void {
        // Handle tab visibility
        if (this.hideDetailsTab && this.drawerTab === this.drawerTabsOptions.Details) {
          this.drawerTab = this.drawerTabsOptions.EventHistory;
        }
        if (this.hideTerminalTab && this.drawerTab === this.drawerTabsOptions.Terminal) {
          this.drawerTab = this.drawerTabsOptions.EventHistory;
        }
        if (this.hideYamlTab && this.drawerTab === this.drawerTabsOptions.Yaml) {
          this.drawerTab = this.drawerTabsOptions.EventHistory;
        }
        if (this.hideArchitectureTab && this.drawerTab === this.drawerTabsOptions.Structure) {
          this.drawerTab = this.drawerTabsOptions.EventHistory;
        }

        // Handle workload change for pods
        const workloadChanged = newWorkload?.id !== prevWorkload?.id;
        if (workloadChanged) {
          this.selectedStructureElementId = null;
          this.resetPodState();

          // If we have a workload and we're on a pods tab, load pods
          if (newWorkload && this.activeTabNeedsPods) {
            this.loadPods(true);
            this.startPodsInterval();
          }
        }
      },
      deep: true,
    },
    drawerTab: {
      handler(newTab: EWorkloadDrawerTabs | null, oldTab: EWorkloadDrawerTabs | null): void {
        const wasPodsTab = this.tabNeedsPods(oldTab);
        const isPodsTab = this.tabNeedsPods(newTab);

        if (!wasPodsTab && isPodsTab && this.workload) {
          // Switched to a pods tab - load pods and start interval
          this.loadPods(true);
          this.startPodsInterval();
        } else if (wasPodsTab && !isPodsTab) {
          // Switched away from pods tab - stop interval (keep pods in memory)
          this.stopPodsInterval();
        }
      },
    },
    navigationPayload: {
      handler(payload: IWorkloadDrawerNavigationPayload | null): void {
        if (!payload) {
          // Don't clear initialPodName here - it will be used for pod selection
          return;
        }

        // Map tab type to drawer tab enum
        const allTabs = Object.values(EWorkloadDrawerTabs);
        if (allTabs.includes(payload.tab as EWorkloadDrawerTabs)) {
          this.drawerTab = payload.tab as EWorkloadDrawerTabs;
        }

        this.initialPodName = payload.podName ?? null;

        // If pods are already loaded, apply the selection immediately
        if (this.podsToDisplay.length > 0 && payload.podName) {
          const podExists = this.podsToDisplay.find((pod) => pod.name === payload.podName);
          if (podExists) {
            this.handlePodSelected(payload.podName);
          }
        }

        // Emit after next tick to ensure the tab change is processed
        this.$nextTick(() => {
          this.$emit("navigation-handled");
        });
      },
      immediate: true,
    },
  },
  beforeUnmount() {
    this.stopPodsInterval();
  },
});
</script>
