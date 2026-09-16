<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="element-details-tab">
    <details-panel-view :sections="accordionSections" :loading="podsCountLoading" />
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { WorkloadElementResponse } from "@/swagger-models/workloads-client";
import type { IDetailsPanelSectionDef } from "@/models/details-panel.model";
import { workloadElementDetailsUtil } from "@/utils/workload-element-details.util";
import { topologyLevelsFetcherService } from "@/services/topology-levels-fetcher.service/topology-levels-fetcher.service";
import DetailsPanelView from "@/components/workload/workload-drawer-content/tabs/workload-details-tab/details-panel-view.vue";

defineOptions({ name: "element-details-tab" });

const props = withDefaults(
  defineProps<{
    element: WorkloadElementResponse | null;
    elementPodsCount?: number;
    podsCountLoading?: boolean;
    clusterId?: string;
  }>(),
  {
    element: null,
    elementPodsCount: 0,
    podsCountLoading: false,
    clusterId: "",
  },
);

const topologyLevels = ref<string[]>([]);

const fetchTopologyLevels = topologyLevelsFetcherService.createTopologyLevelsFetcher();

watch(
  () => [props.element?.id, props.element?.topology?.name, props.clusterId] as const,
  async ([, topologyName, clusterId]) => {
    topologyLevels.value = [];
    const levels = await fetchTopologyLevels(topologyName, clusterId);
    if (levels !== undefined) topologyLevels.value = levels;
  },
  { immediate: true },
);

const accordionSections = computed<IDetailsPanelSectionDef[]>(() => {
  if (!props.element) return [];
  return workloadElementDetailsUtil.buildElementDetailsSections(props.element, {
    topologyLevels: topologyLevels.value,
    podsCount: props.elementPodsCount,
  });
});
</script>

<style scoped lang="scss">
.element-details-tab {
  height: 100%;
}
</style>
