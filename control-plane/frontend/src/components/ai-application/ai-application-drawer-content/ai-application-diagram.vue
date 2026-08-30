<template>
  <div ref="containerRef" class="resource-graph-container full-width full-height relative-position">
    <vue-flow
      v-model:nodes="flowNodes"
      v-model:edges="flowEdges"
      :default-viewport="{ zoom: 1, x: 20, y: 20 }"
      :min-zoom="0.3"
      :max-zoom="1.5"
      fit-view-on-init
      :fit-view-options="{ padding: 0.2 }"
      :nodes-draggable="false"
      :nodes-connectable="false"
      :elements-selectable="false"
      @node-click="onNodeClick"
    >
      <template #node-island>
        <div class="resource-island" />
      </template>

      <template #node-Workload="nodeProps">
        <ai-application-resource-card
          :data-node-id="nodeProps.id"
          :name="nodeProps.data.name"
          :type="EResourceCardType.Workload"
          :phase="nodeProps.data.phase"
          :expanded="isExpanded(nodeProps.id)"
          :search-term="searchTerm"
        >
          <template #details v-if="isExpanded(nodeProps.id) && getWorkloadById(nodeProps.data.workloadId)">
            <ai-application-workload-expanded
              :workload="getWorkloadById(nodeProps.data.workloadId)!"
              @go-to-workload="goToWorkload(nodeProps.data.name)"
            />
          </template>
        </ai-application-resource-card>
      </template>

      <template #node-Secret="nodeProps">
        <ai-application-resource-card
          :data-node-id="nodeProps.id"
          :name="nodeProps.data.name"
          :type="EResourceCardType.Secret"
          :search-term="searchTerm"
        />
      </template>

      <template #node-PersistentVolumeClaim="nodeProps">
        <ai-application-resource-card
          :data-node-id="nodeProps.id"
          :name="nodeProps.data.name"
          :type="EResourceCardType.Pvc"
          :search-term="searchTerm"
        />
      </template>

      <template #node-ConfigMap="nodeProps">
        <ai-application-resource-card
          :data-node-id="nodeProps.id"
          :name="nodeProps.data.name"
          :type="EResourceCardType.ConfigMap"
          :search-term="searchTerm"
        />
      </template>

      <template #node-endpoint="nodeProps">
        <diagram-endpoint-card
          :data-node-id="nodeProps.id"
          :url="nodeProps.data.url"
          variant="application"
          :handle-id="ENDPOINT_BOTTOM_HANDLE_ID"
          :handle-position="Position.Bottom"
          :highlight-text="searchTerm"
        />
      </template>

      <template #edge-connection="edgeProps">
        <diagram-connection-edge v-bind="edgeProps" />
      </template>

      <controls position="bottom-left" :show-interactive="false">
        <template #icon-zoom-in>
          <i class="ph-bold ph-plus control-icon" />
        </template>
        <template #icon-zoom-out>
          <i class="ph-bold ph-minus control-icon" />
        </template>
        <template #icon-fit-view>
          <i class="ph-bold ph-corners-out control-icon" />
        </template>
      </controls>
      <mini-map position="bottom-right" pannable zoomable />
    </vue-flow>
    <div v-if="showNoResults" class="diagram-no-results text-b-body1">
      Application details aren't available right now. If this issue persists, contact your administrator.
    </div>
    <ai-application-diagram-legend />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onMounted, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import { Position, VueFlow, useVueFlow } from "@vue-flow/core";
import { Controls } from "@vue-flow/controls";
import { MiniMap } from "@vue-flow/minimap";
import type { Edge, Node, NodeMouseEvent } from "@vue-flow/core";
import type {
  AIApplicationResource,
  AiApplicationEndpoint,
  AiApplicationResourceConnection,
  Workload,
} from "@/swagger-models/workloads-client";
import { WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";
import { DiagramConnectionEdge, DiagramEndpointCard, measureCardHeights } from "@/components/common/diagram-layout";
import AiApplicationResourceCard from "./ai-application-resource-card.vue";
import AiApplicationWorkloadExpanded from "./ai-application-workload-expanded.vue";
import AiApplicationDiagramLegend from "./ai-application-diagram-legend.vue";
import {
  EResourceCardType,
  EResourceKind,
  ENDPOINT_BOTTOM_HANDLE_ID,
  FILTERABLE_RESOURCE_KINDS,
  createWorkloadMap,
  buildApplicationGraph,
  buildDiagram,
  filterApplicationGraph,
  getGraphStructureFingerprint,
} from "./ai-application-diagram.util";

export interface Props {
  resources?: AIApplicationResource[];
  workloads?: Workload[];
  resourceConnections?: AiApplicationResourceConnection[];
  endpoints?: AiApplicationEndpoint[];
  selectedResourceKinds?: EResourceKind[];
  searchTerm?: string;
}

const props = withDefaults(defineProps<Props>(), {
  resources: () => [],
  workloads: () => [],
  resourceConnections: () => [],
  endpoints: () => [],
  selectedResourceKinds: () => [...FILTERABLE_RESOURCE_KINDS],
  searchTerm: "",
});

const router = useRouter();
const containerRef = ref<HTMLElement | null>(null);
const flowNodes = ref<Node[]>([]);
const flowEdges = ref<Edge[]>([]);
const cardHeights = ref<Record<string, number>>({});
const expandedNodes = ref<Set<string>>(new Set());
const { fitView, updateNodeInternals } = useVueFlow();

const workloadMap = computed(() => createWorkloadMap(props.workloads));
const applicationGraph = computed(() =>
  buildApplicationGraph(props.resources, props.resourceConnections, props.endpoints, props.workloads),
);
const filteredApplicationGraph = computed(() =>
  filterApplicationGraph(applicationGraph.value, {
    selectedKinds: props.selectedResourceKinds,
    searchTerm: props.searchTerm,
  }),
);
const graphStructureFingerprint = computed(() => getGraphStructureFingerprint(applicationGraph.value));
const filterFingerprint = computed(
  () => `${props.selectedResourceKinds.slice().sort().join(",")}|${props.searchTerm.trim().toLocaleLowerCase()}`,
);
const showNoResults = computed(
  () =>
    filteredApplicationGraph.value.islands.length === 0 &&
    filteredApplicationGraph.value.unconnectedResources.length === 0,
);
let previousGraphStructureFingerprint = "";
let resizeObserver: ResizeObserver | null = null;

const isExpanded = (nodeId: string): boolean => expandedNodes.value.has(nodeId);

const rebuildDiagram = (): void => {
  const result = buildDiagram(
    filteredApplicationGraph.value,
    cardHeights.value,
    containerRef.value?.clientWidth || undefined,
  );
  flowNodes.value = result.nodes;
  flowEdges.value = result.edges;
};

const measureAndUpdatePositions = (shouldFitView = true): void => {
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      cardHeights.value = measureCardHeights(".resource-graph-container");
      rebuildDiagram();
      const nodeIds = flowNodes.value.map((node) => node.id);
      if (nodeIds.length > 0) updateNodeInternals(nodeIds);
      if (shouldFitView) nextTick(() => fitView({ padding: 0.15 }));
    });
  });
};

const toggleExpanded = (nodeId: string): void => {
  if (expandedNodes.value.has(nodeId)) {
    expandedNodes.value.delete(nodeId);
  } else {
    expandedNodes.value.add(nodeId);
  }
  expandedNodes.value = new Set(expandedNodes.value);
  nextTick(() => measureAndUpdatePositions(false));
};

const getWorkloadById = (workloadId: string | undefined): Workload | undefined => {
  if (!workloadId) return undefined;
  return workloadMap.value.get(workloadId);
};

const onNodeClick = (event: NodeMouseEvent): void => {
  if (event.node.type === EResourceKind.Workload) toggleExpanded(event.node.id);
};

const goToWorkload = (workloadName: string): void => {
  if (workloadName) {
    router.push({
      name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
      query: {
        advancedFilters: JSON.stringify([
          {
            name: WorkloadSortFilterFields.Name,
            label: "Workload",
            type: EColumnFilterType.FreeText,
            term: workloadName,
            value: EFilterOperator.Contains,
          },
        ]),
      },
    });
  }
};

onMounted(() => {
  if (!containerRef.value) return;
  let previousWidth = containerRef.value.clientWidth;
  resizeObserver = new ResizeObserver(([entry]) => {
    const width = entry.contentRect.width;
    if (width === previousWidth) return;
    previousWidth = width;
    nextTick(() => measureAndUpdatePositions(false));
  });
  resizeObserver.observe(containerRef.value);
});

onUnmounted(() => resizeObserver?.disconnect());

watch(
  [() => props.resources, () => props.workloads, () => props.resourceConnections, () => props.endpoints],
  () => {
    const structureChanged = graphStructureFingerprint.value !== previousGraphStructureFingerprint;
    previousGraphStructureFingerprint = graphStructureFingerprint.value;

    // Keep measured heights on status-only polls so rebuild doesn't flash default-height layout.
    if (structureChanged) {
      cardHeights.value = {};
    }
    rebuildDiagram();
    if (structureChanged || flowNodes.value.length === 0) {
      nextTick(() => measureAndUpdatePositions(true));
    }
  },
  { immediate: true, deep: true },
);

watch(filterFingerprint, () => {
  cardHeights.value = {};
  rebuildDiagram();
  nextTick(() => measureAndUpdatePositions(true));
});
</script>

<style lang="scss">
@import "@/assets/scss/components/vue-flow-diagram";

.vue-flow__node-Workload {
  cursor: pointer;
}
</style>

<style lang="scss" scoped>
.resource-graph-container {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 500px;

  :deep(.vue-flow) {
    background-color: $body-background-color;
    border-radius: 4px;
  }
}

.diagram-no-results {
  position: absolute;
  top: 50%;
  left: 50%;
  z-index: 4;
  transform: translate(-50%, -50%);
  color: var(--b-text-sub);
}

.resource-island {
  width: 100%;
  height: 100%;
  border: 1px solid var(--border-default);
  border-radius: 8px;
  background: var(--bg-secondary-default);
  pointer-events: none;
}
</style>
