<template>
  <runai-drawer-content>
    <template v-slot:header>
      <section class="graph-header">
        <div class="node-name">
          <span class="graph-header-label">Node:</span>
          <span class="graph-header-value">{{ node?.name || "" }}</span>
        </div>
        <div class="gpu-type" v-if="node?.gpuInfo?.gpuType">
          <span class="graph-header-label">GPU type:</span>
          <span class="graph-header-value">{{ node.gpuInfo.gpuType }}</span>
        </div>
      </section>
    </template>
    <template v-slot:subheader>
      <runai-tab
        :model-value="'metrics'"
        :tabs="[{ name: 'metrics', label: 'metrics', aid: 'node-drawer-tab-metrics' }]"
      />
    </template>

    <template v-slot:body> <node-charts :cluster-id="clusterId" :node="node" /> </template>
  </runai-drawer-content>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiDrawerContent } from "@/components/common/runai-drawer-content";
import { RunaiTab } from "@/components/common/runai-tab";
import type { Node } from "@/swagger-models/cluster-service-client";
import { NodeCharts } from "@/components/node/node-drawer-content/node-charts";

export default defineComponent({
  name: "node-drawer-content",
  components: {
    NodeCharts,
    RunaiDrawerContent,
    RunaiTab,
  },
  props: {
    filterBy: {
      type: Object,
      required: true,
    },
    clusterId: {
      type: String,
      required: true,
    },
    node: {
      type: Object as PropType<Node>,
      required: false,
    },
  },
});
</script>

<style scoped lang="scss">
.graph-header {
  display: flex;
  width: 100%;

  > div {
    flex: 50%;
  }
  .graph-header-label {
    color: $black-54;
    margin-right: 5px;
  }

  .graph-header-value {
    color: $black-87;
  }
}
</style>
