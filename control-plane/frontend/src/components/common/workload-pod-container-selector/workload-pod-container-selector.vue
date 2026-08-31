<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="workload-pod-selector-header row items-center justify-between">
    <section v-if="pods.length" class="row items-center q-pa-sm q-gutter-sm">
      <runai-select
        outlined
        label="Pod"
        class="selector-select"
        dense
        hide-bottom-space
        options-dense
        selected-ellipsis
        emit-value
        :options="podOptions"
        :model-value="selectedPodName"
        :virtual-scroll-slice-size="50"
        @update:model-value="handlePodSelected"
      />

      <runai-select
        v-if="showContainerSelect"
        outlined
        label="Container"
        class="selector-select"
        dense
        hide-bottom-space
        selected-ellipsis
        options-dense
        :options="containerNames"
        :model-value="selectedContainerName"
        @update:model-value="handleContainerSelected"
      />

      <slot name="extra-selectors" />
    </section>

    <!-- Actions slot for custom buttons (like download) -->
    <slot name="actions" :selected-pod="selectedPod" :selected-container="selectedContainerName" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { WorkloadServicePod } from "@/models/workload.model";
import type { ISelectOption } from "@/models/global.model";
import { RunaiSelect } from "@/components/common/runai-select";
import { phaseColors } from "@/utils/workload.util/workload-chart.util";
import type { Phase } from "@/swagger-models/workloads-client";

export interface Props {
  pods: WorkloadServicePod[];
  selectedPodName: string;
  selectedContainerName: string;
  showPodStatus?: boolean;
  showContainerSelect?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showPodStatus: true,
  showContainerSelect: true,
});

const emit = defineEmits<{
  "pod-selected": [podName: string];
  "container-selected": [containerName: string];
}>();

const podOptions = computed((): ISelectOption[] | string[] => {
  if (props.showPodStatus) {
    return props.pods.map((pod) => ({
      label: pod.name,
      value: pod.name,
      subLabel: pod.k8sPhase,
      subLabelColor: getPodStatusColor(pod.k8sPhase),
    }));
  }
  return props.pods.map((pod) => pod.name);
});

const getPodStatusColor = (phase: string): string => {
  return phaseColors[phase as Phase] ?? "";
};

const selectedPod = computed((): WorkloadServicePod | undefined => {
  return props.pods.find((pod) => pod.name === props.selectedPodName);
});

const containerNames = computed((): string[] => {
  if (!selectedPod.value?.containers) return [];
  return selectedPod.value.containers.map((container) => container.name);
});

const handlePodSelected = (podName: string): void => {
  emit("pod-selected", podName);
};

const handleContainerSelected = (containerName: string): void => {
  emit("container-selected", containerName);
};
</script>

<style lang="scss" scoped>
.workload-pod-selector-header {
  width: 100%;
  font-size: 16px;
  font-weight: 500;
  border-bottom: 1px solid $black-10;

  .selector-select {
    width: 280px;
  }
}
</style>
