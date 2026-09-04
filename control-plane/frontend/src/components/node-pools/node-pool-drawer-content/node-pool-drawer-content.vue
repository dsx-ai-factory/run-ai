<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-drawer-content>
    <template v-slot:header>
      <section class="graph-header">
        <div class="node-pool-name">
          <span class="graph-header-label">Node Pool:</span>
          <span class="graph-header-value">{{ nodePool.name }}</span>
        </div>
      </section>
    </template>

    <template v-slot:subheader>
      <runai-tab
        v-model="drawerTab"
        class="full-width"
        :tabs="[{ name: ENodePoolDrawerTabs.Metrics, label: 'METRICS', aid: 'node-pool-drawer-tab-metrics' }]"
      />
    </template>

    <template v-slot:body>
      <runai-tab-panels v-model="drawerTab" class="node-pool-drawer-tab-panels">
        <runai-tab-panel class="q-pa-none" :name="ENodePoolDrawerTabs.Metrics">
          <node-pool-charts :selected-node-pool="nodePool" :cluster-id="nodePool.clusterId" />
        </runai-tab-panel>
      </runai-tab-panels>
    </template>
  </runai-drawer-content>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// cmps
import { RunaiDrawerContent } from "@/components/common/runai-drawer-content";
import { RunaiTab, RunaiTabPanels, RunaiTabPanel } from "@/components/common/runai-tab";

// types
import type { PropType } from "vue";

// models
import { NodePoolCharts } from "@/components/node-pools/node-pool-drawer-content/node-pool-charts/";
import type { Nodepool } from "@/swagger-models/cluster-service-client";

export enum ENodePoolDrawerTabs {
  Metrics = "metrics",
}

export default defineComponent({
  components: {
    NodePoolCharts,
    RunaiDrawerContent,
    RunaiTab,
    RunaiTabPanels,
    RunaiTabPanel,
  },
  props: {
    nodePool: {
      type: Object as PropType<Nodepool>,
      required: true,
    },
  },
  data() {
    return {
      drawerTab: ENodePoolDrawerTabs.Metrics,
    };
  },
  setup() {
    return { ENodePoolDrawerTabs };
  },
});
</script>
<style lang="scss">
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

.node-pool-drawer-tab-panels {
  flex: 1;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;

  .q-tab-panel {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
}
</style>
