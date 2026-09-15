<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="element-details-tab">
    <details-panel-view :sections="accordionSections" />
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { WorkloadElementResponse } from "@/swagger-models/workloads-client";
import type { IDetailsPanelSectionDef } from "@/models/details-panel.model";
import { buildElementDetailsSections } from "@/utils/workload-element-details.util";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import DetailsPanelView from "@/components/workload/workload-drawer-content/tabs/workload-details-tab/details-panel-view.vue";

defineOptions({ name: "element-details-tab" });

const props = withDefaults(
  defineProps<{
    element: WorkloadElementResponse | null;
    elementPodsCount?: number;
    clusterId?: string;
  }>(),
  {
    element: null,
    elementPodsCount: 0,
    clusterId: "",
  },
);

const topologyLevels = ref<string[]>([]);

let latestRequestId = 0;

async function fetchTopologyLevels(topologyName: string | undefined, clusterId: string): Promise<void> {
  const requestId = ++latestRequestId;
  topologyLevels.value = [];
  if (!topologyName || !clusterId) return;

  const levels = await clusterService.getNetworkTopologyLevelsByName(topologyName, clusterId);
  if (requestId !== latestRequestId) return;
  topologyLevels.value = levels;
}

watch(
  () => [props.element?.id, props.element?.topology?.name, props.clusterId] as const,
  ([, topologyName, clusterId]) => {
    fetchTopologyLevels(topologyName, clusterId);
  },
  { immediate: true },
);

const accordionSections = computed<IDetailsPanelSectionDef[]>(() => {
  if (!props.element) return [];
  return buildElementDetailsSections(props.element, {
    topologyLevels: topologyLevels.value,
    podsCount: props.elementPodsCount,
  });
});

defineExpose({ accordionSections });
</script>

<style scoped lang="scss">
.element-details-tab {
  height: 100%;
}
</style>
