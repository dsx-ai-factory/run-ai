<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-structure-page">
    <Teleport defer to="#header-left-teleport" v-if="hasHeaderData">
      <div class="structure-header-left row items-center no-wrap">
        <span class="text-b-h6-md header-title">Workload structure</span>

        <div class="row items-center no-wrap workload-identity">
          <runai-svg-icon v-if="headerIcon" :name="headerIcon" size="22" :key="headerIcon" />
          <span class="text-b-h6 workload-name">{{ headerName }}</span>
          <span v-if="headerStatusOptions" class="workload-status-wrap">
            <runai-status :options="headerStatusOptions" />
          </span>
          <template v-if="selectedElementName">
            <i class="ph-bold ph-caret-right breadcrumb-chevron" />
            <span class="selected-element-breadcrumb text-b-h6-md">{{ selectedElementName }}</span>
          </template>
        </div>
      </div>
    </Teleport>

    <runai-default-loader v-if="loading" />
    <runai-drawer-no-data
      v-else-if="!workload"
      :messages="['Workload structure is not available.']"
      icon-name="ph-bold ph-file-text"
    />
    <template v-else>
      <main class="structure-body" :style="{ '--context-panel-width': `${rightPanelWidth}px` }">
        <section class="diagram-panel">
          <workload-architecture-diagram
            :workload="workload"
            :elements="elements"
            :pods="diagramPods"
            :selected-element-id="selectedElementId"
            :element-details-by-id="elementDetailsById"
            selectable
            @node-selected="selectElement"
          />
        </section>

        <aside class="context-panel" :style="{ width: `${rightPanelWidth}px` }">
          <div
            class="context-panel-resizer"
            role="separator"
            aria-orientation="vertical"
            aid="workload-structure-context-resizer"
            v-touch-pan.horizontal.prevent.mouse.preserveCursor="handleResize"
          />
          <runai-tab v-model="activeTab" :tabs="contextTabs" class="context-tabs" />

          <runai-tab-panels v-model="activeTab" class="context-tab-panels">
            <runai-tab-panel class="q-pa-none" :name="tabOptions.EventHistory">
              <workload-history-tab
                :drawer-width="contextPanelWidth"
                :workload="workload"
                :show-element-empty-state="isElementContext"
              />
            </runai-tab-panel>
            <runai-tab-panel class="q-pa-none" :name="tabOptions.Metrics">
              <workload-metrics-tab
                :workload="workload"
                :drawer-width="contextPanelWidth"
                :selected-element-id="selectedElementId"
              />
            </runai-tab-panel>
            <runai-tab-panel class="q-pa-none" :name="tabOptions.Pods">
              <workload-pods-tab :pods="contextPods" :loading="loadingContextPods" />
            </runai-tab-panel>
            <runai-tab-panel class="q-pa-none" :name="tabOptions.Logs">
              <workload-logs-tab
                :workload="workload"
                :pods="contextPods"
                :loading-pods="loadingContextPods"
                :selected-pod-name="selectedPodName"
                :selected-container-name="selectedContainerName"
                @pod-selected="handlePodSelected"
                @container-selected="handleContainerSelected"
              />
            </runai-tab-panel>
            <runai-tab-panel class="q-pa-none" :name="tabOptions.Terminal">
              <workload-terminal
                :workload="workload"
                :pods="contextPods"
                :loading-pods="loadingContextPods"
                :selected-pod-name="selectedPodName"
                :selected-container-name="selectedContainerName"
                @pod-selected="handlePodSelected"
                @container-selected="handleContainerSelected"
              />
            </runai-tab-panel>
            <runai-tab-panel v-if="!hideYamlTab" class="q-pa-none" :name="tabOptions.Yaml">
              <workload-yaml-tab
                :workload="workload"
                :view="EYamlTabView.Expanded"
                :scope-element-id="selectedElementId"
              />
            </runai-tab-panel>
            <runai-tab-panel class="q-pa-none" :name="tabOptions.Details">
              <workload-details-tab :workload="workload" :show-element-empty-state="isElementContext" />
            </runai-tab-panel>
          </runai-tab-panels>
        </aside>
      </main>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { TouchPanValue } from "quasar";
import type { IStatusColOptions } from "@/models/table.model";
import type { WorkloadServicePod } from "@/models/workload.model";
import type { Pod, Workload, WorkloadElement, WorkloadElementResponse } from "@/swagger-models/workloads-client";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import { RunaiTab, RunaiTabPanels, RunaiTabPanel, type ITab } from "@/components/common/runai-tab";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";
import { RunaiStatus } from "@/components/common/runai-status";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { WorkloadHistoryTab } from "@/components/workload/workload-drawer-content/tabs/workload-history-tab";
import { WorkloadMetricsTab } from "@/components/workload/workload-drawer-content/tabs/workload-metrics-tab";
import { WorkloadLogsTab } from "@/components/workload/workload-drawer-content/tabs/workload-logs-tab";
import { WorkloadTerminal } from "@/components/workload/workload-drawer-content/tabs/workload-terminal";
import { WorkloadYamlTab, EYamlTabView } from "@/components/workload/workload-drawer-content/tabs/workload-yaml-tab";
import WorkloadDetailsTab from "@/components/workload/workload-drawer-content/tabs/workload-details-tab/workload-details-tab.vue";
import WorkloadArchitectureDiagram from "@/components/workload/workload-drawer-content/tabs/workload-architecture-tab/workload-architecture-diagram.vue";
import type { WorkloadElementDetailsById } from "@/components/workload/workload-drawer-content/tabs/workload-architecture-tab/workload-architecture-diagram.util";
import { WorkloadPodsTab } from "@/components/workload/workload-drawer-content/tabs/workload-pods-tab";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { workloadUtil } from "@/utils/workload.util";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { useAppStore } from "@/stores/app.store";
import { useClusterStore } from "@/stores/cluster.store";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { MIN_CLUSTER_VERSION_FOR_K8S_OBJECT_API } from "@/common/version.constant";

type TouchPanEvent = Parameters<NonNullable<TouchPanValue>>[0];

const PANEL_DEFAULT_WIDTH = 654;
const PANEL_MIN_WIDTH = 654;
const PANEL_MAX_WIDTH_RATIO = 0.8;

const tabOptions = {
  EventHistory: "event history",
  Metrics: "metrics",
  Pods: "pods",
  Logs: "logs",
  Terminal: "terminal",
  Yaml: "yaml",
  Details: "details",
} as const;

type StructureTab = (typeof tabOptions)[keyof typeof tabOptions];

const route = useRoute();
const router = useRouter();
const appStore = useAppStore();
const clusterStore = useClusterStore();
appStore.setPageLoading(false);

const loading = ref(false);
const loadingContextPods = ref(false);
const workload = ref<Workload | null>(null);
const elements = ref<WorkloadElement[]>([]);
const diagramPods = ref<Pod[]>([]);
const contextPods = ref<WorkloadServicePod[]>([]);

// Token used to discard stale loadContextPods responses. The function is fired
// from element-selection changes AND the 30s refresh tick, so a slow response
// for an old elementId could otherwise overwrite a newer selection's pods.
let contextPodsRequestToken = 0;
const elementDetailsById = ref<WorkloadElementDetailsById>({});
const activeTab = ref<StructureTab>(tabOptions.EventHistory);
const rightPanelWidth = ref(PANEL_DEFAULT_WIDTH);
const originalWidth = ref(PANEL_DEFAULT_WIDTH);
const originalLeft = ref(0);
const selectedPodName = ref("");
const selectedContainerName = ref("");

const workloadId = computed(() => {
  const id = route.params.workloadId;
  return Array.isArray(id) ? id[0] : id;
});

const selectedElementId = computed(() => {
  const elementId = route.query.elementId;
  return typeof elementId === "string" && elementId ? elementId : null;
});

const isElementContext = computed(() => Boolean(selectedElementId.value));

const hideYamlTab = computed<boolean>(() => {
  const clusterId = workload.value?.clusterId;
  if (!clusterId) return true;
  return !clusterStore.isClusterVersionSufficient(clusterId, MIN_CLUSTER_VERSION_FOR_K8S_OBJECT_API);
});

const contextPanelWidth = computed(() => `${rightPanelWidth.value}px`);

const contextTabs = computed<ITab[]>(() => [
  {
    name: tabOptions.EventHistory,
    label: "event history",
    hidden: isElementContext.value,
    aid: "workload-structure-tab-event-history",
  },
  { name: tabOptions.Metrics, label: "metrics", aid: "workload-structure-tab-metrics" },
  { name: tabOptions.Pods, label: "pods", aid: "workload-structure-tab-pods" },
  { name: tabOptions.Logs, label: "logs", aid: "workload-structure-tab-logs" },
  { name: tabOptions.Terminal, label: "terminal", aid: "workload-structure-tab-terminal" },
  { name: tabOptions.Yaml, label: "yaml", hidden: hideYamlTab.value, aid: "workload-structure-tab-yaml" },
  { name: tabOptions.Details, label: "details", hidden: isElementContext.value, aid: "workload-structure-tab-details" },
]);

const queryName = computed(() => (typeof route.query.name === "string" ? route.query.name : ""));
const queryType = computed(() => (typeof route.query.type === "string" ? route.query.type : ""));
const queryPhase = computed(() => (typeof route.query.phase === "string" ? route.query.phase : ""));
const queryPhaseUpdatedAt = computed(() =>
  typeof route.query.phaseUpdatedAt === "string" ? route.query.phaseUpdatedAt : "",
);

const headerName = computed(() => workload.value?.name ?? queryName.value);

const headerIcon = computed(() => {
  if (workload.value) return workloadUtil.getWorkloadIcon(workload.value);
  if (queryType.value) return workloadUtil.getWorkloadIcon({ type: queryType.value } as Workload);
  return "";
});

const headerStatusOptions = computed<IStatusColOptions | null>(() => {
  if (workload.value) return workloadUtil.getWorkloadStatusColOptions(workload.value);
  if (queryPhase.value) {
    return workloadUtil.getWorkloadStatusColOptions({
      phase: queryPhase.value as Workload["phase"],
      phaseUpdatedAt: queryPhaseUpdatedAt.value,
    } as Workload);
  }
  return null;
});

const hasHeaderData = computed(() => Boolean(headerName.value));

const selectedElementName = computed(() => {
  if (!selectedElementId.value) return "";
  return (
    elementDetailsById.value[selectedElementId.value]?.name ??
    elements.value.find((element) => element.id === selectedElementId.value)?.name ??
    ""
  );
});

const childIncompatibleTabs: StructureTab[] = [tabOptions.EventHistory, tabOptions.Details];

const getRouteTab = (): StructureTab => {
  const tab = route.query.tab;
  const tabs = Object.values(tabOptions);
  const resolved =
    typeof tab === "string" && tabs.includes(tab as StructureTab) ? (tab as StructureTab) : tabOptions.EventHistory;
  if (isElementContext.value && childIncompatibleTabs.includes(resolved)) {
    return tabOptions.Metrics;
  }
  return resolved;
};

const replaceQuery = (nextQuery: Record<string, string | undefined>): void => {
  router.replace({
    name: WORKLOAD_ROUTE_NAMES.WORKLOAD_STRUCTURE,
    params: { workloadId: workloadId.value },
    query: {
      ...route.query,
      ...nextQuery,
    },
  });
};

const selectElement = (elementId: string | null): void => {
  replaceQuery({ elementId: elementId ?? undefined });
};

const handleResize = (ev: TouchPanEvent): void => {
  const left = ev.position?.left ?? 0;
  if (ev.isFirst) {
    originalWidth.value = rightPanelWidth.value;
    originalLeft.value = left;
    return;
  }

  const delta = left - originalLeft.value;
  rightPanelWidth.value = Math.max(
    PANEL_MIN_WIDTH,
    Math.min(window.innerWidth * PANEL_MAX_WIDTH_RATIO, originalWidth.value - delta),
  );
};

const fetchElementDetails = async (fetchedElements: WorkloadElement[]): Promise<WorkloadElementDetailsById> => {
  const currentWorkloadId = workloadId.value;
  const entries = await Promise.all(
    fetchedElements.map(async (element): Promise<[string, WorkloadElementResponse | null]> => {
      try {
        return [element.id, await workloadService.getWorkloadElement(currentWorkloadId, element.id)];
      } catch {
        return [element.id, null];
      }
    }),
  );

  return entries.reduce<WorkloadElementDetailsById>((acc, [elementId, elementDetails]) => {
    if (elementDetails) {
      acc[elementId] = elementDetails;
    }
    return acc;
  }, {});
};

const loadContextPods = async (): Promise<void> => {
  if (!workload.value) return;

  const requestToken = ++contextPodsRequestToken;
  loadingContextPods.value = true;
  try {
    const filterBy = selectedElementId.value
      ? [filterUtil.getEqualsFilterString("elementId", selectedElementId.value)]
      : undefined;
    const fetched = (await workloadService.getWorkloadPods(
      workload.value.id,
      false,
      undefined,
      undefined,
      undefined,
      true,
      filterBy,
    )) as WorkloadServicePod[];

    // Drop stale responses — a newer selection has already kicked off another fetch.
    if (requestToken !== contextPodsRequestToken) return;

    contextPods.value = fetched;
    setDefaultPodSelection();
  } finally {
    if (requestToken === contextPodsRequestToken) {
      loadingContextPods.value = false;
    }
  }
};

const setDefaultPodSelection = (): void => {
  if (!contextPods.value.length) {
    selectedPodName.value = "";
    selectedContainerName.value = "";
    return;
  }

  // Preserve the user's current pod + container if the pod still exists. Re-running
  // handlePodSelected would clobber selectedContainerName back to the first container
  // on every auto-refresh tick, which breaks an in-flight Logs/Terminal session.
  if (selectedPodName.value && contextPods.value.some((pod) => pod.name === selectedPodName.value)) {
    return;
  }

  handlePodSelected(contextPods.value[0].name);
};

const handlePodSelected = (podName: string): void => {
  selectedPodName.value = podName;
  const pod = contextPods.value.find((entry) => entry.name === podName);
  selectedContainerName.value = pod?.containers?.[0]?.name ?? "";
};

const handleContainerSelected = (containerName: string): void => {
  selectedContainerName.value = containerName;
};

const refreshStructureData = async (): Promise<void> => {
  if (!workloadId.value) return;
  try {
    const [fetchedWorkload, fetchedElements, fetchedPods] = await Promise.all([
      workloadService.getWorkloadById(workloadId.value),
      workloadService.getWorkloadElements(workloadId.value).catch(() => [] as WorkloadElement[]),
      workloadService
        .getWorkloadPods(workloadId.value, false, undefined, undefined, undefined, true)
        .catch(() => [] as Pod[]),
    ]);

    workload.value = fetchedWorkload as unknown as Workload;
    elements.value = fetchedElements;
    diagramPods.value = fetchedPods as Pod[];
    elementDetailsById.value = await fetchElementDetails(fetchedElements);
    await loadContextPods();
  } catch (error) {
    console.error("Failed to refresh workload structure:", error);
  }
};

const loadPage = async (): Promise<void> => {
  loading.value = true;
  try {
    await refreshStructureData();
  } finally {
    loading.value = false;
  }
};

const startRefreshInterval = (): void => {
  intervalUtil.stopInterval(EIntervalLabels.WorkloadStructure);
  intervalUtil.startInterval(EIntervalLabels.WorkloadStructure, () => {
    refreshStructureData();
  });
};

watch(
  () => route.query.tab,
  () => {
    activeTab.value = getRouteTab();
  },
  { immediate: true },
);

watch(activeTab, (tab) => {
  if (route.query.tab !== tab) {
    replaceQuery({ tab });
  }
});

watch(selectedElementId, async (newId) => {
  if (newId && childIncompatibleTabs.includes(activeTab.value)) {
    activeTab.value = tabOptions.Metrics;
  }
  selectedPodName.value = "";
  selectedContainerName.value = "";
  await loadContextPods();
});

onMounted(async () => {
  await loadPage();
  startRefreshInterval();
});

onUnmounted(() => {
  intervalUtil.stopInterval(EIntervalLabels.WorkloadStructure);
});
</script>

<style lang="scss">
.structure-header-left {
  gap: var(--space-xl);
  margin-left: 18px;

  .header-title {
    white-space: nowrap;
  }

  .workload-identity {
    .workload-name {
      margin-left: var(--space-xs);
      white-space: nowrap;
    }

    .workload-status-wrap {
      margin-left: var(--space-s);
    }
  }

  .breadcrumb-chevron {
    font-size: 12px;
    color: var(--icon-default);
    margin: 0 var(--space-s);
  }

  .selected-element-breadcrumb {
    display: inline-flex;
    align-items: center;
    height: 24px;
    padding: 0 var(--space-xs);
    border: 1px solid var(--border-default);
    border-radius: var(--space-3xs);
    // Figma's interactive/secondary/pressed/bg renders as #eeeeee (grey-3); our local
    // --interactive-secondary-pressed-bg token maps to grey-4 (#e0e0e0), so we use
    // --bg-primary-hover (#eeeeee) here to match the Figma visual.
    background: var(--bg-primary-hover);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 200px;
  }
}
</style>

<style scoped lang="scss">
$header-height: 60px;

.workload-structure-page {
  position: relative;
  height: calc(100vh - #{$header-height});
  display: flex;
  flex-direction: column;
}

.structure-body {
  position: relative;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
}

.diagram-panel {
  flex: 1;
  min-width: 0;
  height: 100%;
  background: $body-background-color;

  :deep(.vue-flow__minimap) {
    right: calc(var(--context-panel-width) + 10px);
  }
}

.context-panel {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 2;
  display: flex;
  flex-direction: column;
  border-left: 1px solid var(--border-subtle);
  background: var(--bg-primary-default);
  box-shadow: -1px 0 7px 0 rgba(0, 0, 0, 0.15);
}

.context-panel-resizer {
  position: absolute;
  top: 0;
  bottom: 0;
  left: -14px;
  width: 14px;
  cursor: ew-resize;
}

.context-tabs {
  min-height: 50px;
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.context-tab-panels {
  flex: 1;
  min-height: 0;
  overflow: auto;
}
</style>
