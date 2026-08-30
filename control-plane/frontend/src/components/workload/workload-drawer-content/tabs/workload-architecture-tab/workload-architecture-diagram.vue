<template>
  <div ref="containerRef" class="architecture-graph-container full-width full-height relative-position">
    <vue-flow
      v-model:nodes="flowNodes"
      v-model:edges="flowEdges"
      :default-viewport="{ zoom: 1, x: 20, y: 20 }"
      :min-zoom="0.3"
      :max-zoom="1.5"
      fit-view-on-init
      :fit-view-options="{ padding: 0.15 }"
      :nodes-draggable="false"
      :nodes-connectable="false"
      :elements-selectable="false"
      @node-click="onNodeClick"
    >
      <template #node-element="nodeProps">
        <workload-element-card
          :data-node-id="nodeProps.id"
          :name="nodeProps.data.name"
          :kind="nodeProps.data.kind"
          :element-kind="nodeProps.data.elementKind"
          :kubernetes-name="nodeProps.data.kubernetesName"
          :phase="nodeProps.data.phase"
          :workload-type="nodeProps.data.workloadType"
          :resources="nodeProps.data.resources"
          :pods-display="nodeProps.data.podsDisplay"
          :selected="selectedNodeId === nodeProps.id"
          :show-expand-action="showExpandActions"
          :selectable="selectable"
          @expand-click="emitExpandNode(nodeProps.id)"
        />
      </template>

      <template #node-endpoint="nodeProps">
        <diagram-endpoint-card :data-node-id="nodeProps.id" :url="nodeProps.data.url" />
      </template>

      <template #edge-circle="edgeProps">
        <diagram-connection-edge v-bind="edgeProps" />
      </template>

      <template #edge-horizontal-circle="edgeProps">
        <diagram-connection-edge v-bind="edgeProps" />
      </template>

      <controls position="bottom-left" :show-interactive="false">
        <template #control-zoom-in>
          <control-button aid="workload-structure-zoom-in" aria-label="Zoom in" @click="zoomIn">
            <i class="ph-bold ph-plus control-icon" />
            <q-tooltip anchor="center right" self="center left" :offset="[8, 0]" class="diagram-control-tooltip">
              Zoom in
            </q-tooltip>
          </control-button>
        </template>
        <template #control-zoom-out>
          <control-button aid="workload-structure-zoom-out" aria-label="Zoom out" @click="zoomOut">
            <i class="ph-bold ph-minus control-icon" />
            <q-tooltip anchor="center right" self="center left" :offset="[8, 0]" class="diagram-control-tooltip">
              Zoom out
            </q-tooltip>
          </control-button>
        </template>
        <template #control-fit-view>
          <control-button aid="workload-structure-fit-view" aria-label="Fit to view" @click="handleFitView">
            <i class="ph-bold ph-corners-out control-icon" />
            <q-tooltip anchor="center right" self="center left" :offset="[8, 0]" class="diagram-control-tooltip">
              Fit to view
            </q-tooltip>
          </control-button>
        </template>
      </controls>
      <mini-map position="bottom-right" pannable zoomable />
    </vue-flow>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, computed } from "vue";
import { VueFlow, useVueFlow } from "@vue-flow/core";
import { Controls, ControlButton } from "@vue-flow/controls";
import { MiniMap } from "@vue-flow/minimap";
import type { Node, Edge, NodeMouseEvent } from "@vue-flow/core";
import type { Workload, WorkloadElement, Pod } from "@/swagger-models/workloads-client";
import { DiagramConnectionEdge, DiagramEndpointCard, measureCardHeights } from "@/components/common/diagram-layout";
import WorkloadElementCard from "./workload-element-card.vue";
import { buildDiagram, ENDPOINT_NODE_TYPE, type WorkloadElementDetailsById } from "./workload-architecture-diagram.util";

export interface Props {
  workload: Workload;
  elements?: WorkloadElement[];
  pods?: Pod[];
  selectedElementId?: string | null;
  showExpandActions?: boolean;
  selectable?: boolean;
  elementDetailsById?: WorkloadElementDetailsById;
}

const props = withDefaults(defineProps<Props>(), {
  elements: () => [],
  pods: () => [],
  selectedElementId: null,
  showExpandActions: false,
  selectable: false,
  elementDetailsById: () => ({}),
});

const emit = defineEmits<{
  "node-selected": [nodeId: string | null];
  "expand-node": [nodeId: string | null];
}>();

const containerRef = ref<HTMLElement | null>(null);
const flowNodes = ref<Node[]>([]);
const flowEdges = ref<Edge[]>([]);
const cardHeights = ref<Record<string, number>>({});

const { fitView, updateNodeInternals, zoomIn, zoomOut } = useVueFlow();

const handleFitView = (): void => {
  fitView({ padding: 0.15 });
};

const selectedNodeId = computed(() => props.selectedElementId ?? props.workload.id);

const rebuildDiagram = (): void => {
  const result = buildDiagram(props.workload, props.elements, props.pods, cardHeights.value, props.elementDetailsById);
  flowNodes.value = result.nodes;
  flowEdges.value = result.edges;
};

const updatePositions = (shouldFitView: boolean): void => {
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      cardHeights.value = measureCardHeights(".architecture-graph-container");
      rebuildDiagram();
      const allIds = flowNodes.value.map((n) => n.id);
      if (allIds.length > 0) updateNodeInternals(allIds);
      if (shouldFitView) {
        nextTick(() => fitView({ padding: 0.15 }));
      }
    });
  });
};

const onNodeClick = (event: NodeMouseEvent): void => {
  if (!props.selectable) return;
  if (event.node.type === ENDPOINT_NODE_TYPE) return;
  emit("node-selected", event.node.id === props.workload.id ? null : event.node.id);
};

const emitExpandNode = (nodeId: string): void => {
  emit("expand-node", nodeId === props.workload.id ? null : nodeId);
};

const dataFingerprint = computed(() => {
  const elementIds = props.elements
    .map((e) => e.id)
    .sort()
    .join(",");
  const podIds = props.pods
    .map((p) => p.id ?? p.name)
    .sort()
    .join(",");
  const detailsIds = Object.keys(props.elementDetailsById).sort().join(",");
  return `${props.workload?.id}|${elementIds}|${podIds}|${detailsIds}`;
});

watch(
  dataFingerprint,
  () => {
    cardHeights.value = {};
    rebuildDiagram();
    nextTick(() => updatePositions(true));
  },
  { immediate: true },
);
</script>

<style lang="scss">
@import "@/assets/scss/components/vue-flow-diagram";

// Teleported to body — must be unscoped to reach the portaled tooltip element.
.q-tooltip.diagram-control-tooltip {
  white-space: nowrap;
}
</style>

<style lang="scss" scoped>
.architecture-graph-container {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 500px;

  :deep(.vue-flow) {
    background-color: $body-background-color;
    border-radius: 4px;
  }
}
</style>
