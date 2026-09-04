<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<!-- eslint-disable vuejs-accessibility/no-static-element-interactions -->
<template>
  <div
    class="element-card"
    :class="cardClasses"
    :role="selectable ? 'button' : undefined"
    :tabindex="selectable ? 0 : undefined"
    :aria-label="selectable ? `Select ${name}` : undefined"
    :aid="selectable ? `workload-structure-card-${name}` : undefined"
    @click="handleCardClick"
    @keydown.enter="handleCardClick"
    @keydown.space.prevent="handleCardClick"
  >
    <handle id="top" type="target" :position="Position.Top" class="card-handle" />
    <handle id="bottom" type="source" :position="Position.Bottom" class="card-handle" />
    <handle
      v-if="isWorkload"
      :id="WORKLOAD_RIGHT_HANDLE_ID"
      type="source"
      :position="Position.Right"
      class="card-handle"
    />

    <div class="card-header flex items-start no-wrap gap-5">
      <div v-if="showIcon" class="card-icon-box flex items-center justify-center" :class="iconBoxClass">
        <i :class="iconClass"></i>
      </div>
      <div class="flex column flex-1 min-width-0">
        <span class="card-name text-b-subtitle1">{{ name }}</span>
        <workload-status-dot v-if="phase" :status="phase" class="q-mt-xs" />
      </div>
      <q-btn
        v-if="isExpandable"
        flat
        round
        size="xs"
        icon="ph-bold ph-caret-down"
        color="black-54"
        class="expand-chevron-btn"
        :class="{ rotated: expanded }"
        :ripple="false"
      />
    </div>
    <div v-if="isExpandable || showExpandAction" class="card-details">
      <template v-if="typeLabel || kubernetesName || hasResourceDetails">
        <div v-if="typeLabel" class="detail-row text-b-body1">Type: {{ typeLabel }}</div>
        <div v-if="kubernetesName && !isWorkload" class="detail-row text-b-body1">Name: {{ kubernetesName }}</div>
        <div v-for="detail in resourceDetails" :key="detail.label" class="detail-row text-b-body1">
          {{ detail.label }}: {{ detail.value }}
        </div>
      </template>
      <div v-if="podsDisplay" class="detail-row text-b-body1">{{ podsLabel }}: {{ podsDisplay }}</div>
      <div v-if="showExpandAction" class="expand-action-row">
        <q-btn
          flat
          round
          dense
          padding="sm"
          size="sm"
          icon="ph-bold ph-picture-in-picture"
          class="expand-action-btn text-icon-default"
          :ripple="false"
          aria-label="Maximize view"
          :aid="`workload-structure-card-maximize-${name}`"
          @click.stop="$emit('expand-click')"
        >
          <q-tooltip class="tool-tip-dark">Maximize view</q-tooltip>
        </q-btn>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { Handle, Position } from "@vue-flow/core";
import { WorkloadStatusDot } from "@/components/common/workload-status-dot";
import {
  EArchitectureNodeKind,
  WORKLOAD_RIGHT_HANDLE_ID,
  type NodeResourceInfo,
} from "./workload-architecture-diagram.util";
import { formatK8sResourceMemory, numberFormat } from "@/utils/format.util";

export interface Props {
  name: string;
  kind: EArchitectureNodeKind;
  elementKind?: string | null;
  kubernetesName?: string | null;
  phase?: string | null;
  workloadType?: string | null;
  resources?: NodeResourceInfo | null;
  podsDisplay?: string | null;
  selected?: boolean;
  showExpandAction?: boolean;
  selectable?: boolean;
  expanded?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  elementKind: null,
  kubernetesName: null,
  phase: null,
  workloadType: null,
  resources: null,
  podsDisplay: null,
  selected: false,
  showExpandAction: false,
  selectable: false,
  expanded: false,
});

const emit = defineEmits<{
  click: [];
  "expand-click": [];
}>();

const handleCardClick = (): void => {
  if (props.selectable) {
    emit("click");
  }
};

const isWorkload = computed(() => props.kind === EArchitectureNodeKind.Workload);
const hasResourceDetails = computed(() => props.resources !== null);
const hasPodsDisplay = computed(() => props.podsDisplay !== null);
const podsLabel = computed(() => (isWorkload.value ? "Running / requested pods" : "Pods"));
const typeLabel = computed(() => props.workloadType || props.elementKind || null);

const isExpandable = computed(() => {
  if (hasResourceDetails.value) return true;
  if (props.elementKind || props.kubernetesName) return true;
  if (hasPodsDisplay.value) return true;
  return false;
});

const showIcon = computed(() => isWorkload.value);

const iconClass = computed(() => {
  if (isWorkload.value) return "ph-bold ph-laptop";
  return "";
});

const iconBoxClass = computed(() => ({
  "workload-icon": isWorkload.value,
}));

const cardClasses = computed(() => ({
  "workload-card": isWorkload.value,
  selected: props.selected,
  selectable: props.selectable,
}));

const formatNum = (val: number | null | undefined): string => (typeof val === "number" ? numberFormat(val) : "-");

const formatMem = (val: string | null | undefined): string => (val ? formatK8sResourceMemory(val) : "-");

interface DetailItem {
  label: string;
  value: string;
}

const resourceDetails = computed<DetailItem[]>(() => {
  const r = props.resources;
  if (!r) return [];

  const details: DetailItem[] = [];

  if (r.cpuComputeAllocation !== null) {
    details.push({ label: "CPU compute allocation", value: formatNum(r.cpuComputeAllocation) });
  }
  if (r.cpuMemoryAllocation !== null) {
    details.push({ label: "CPU memory allocation", value: formatMem(r.cpuMemoryAllocation) });
  }
  if (r.gpuComputeAllocation !== null) {
    details.push({ label: "GPU compute allocation", value: formatNum(r.gpuComputeAllocation) });
  }
  if (r.gpuMemoryAllocation !== null) {
    details.push({ label: "GPU memory allocation", value: formatMem(r.gpuMemoryAllocation) });
  }

  return details;
});
</script>

<style lang="scss" scoped>
.element-card {
  width: 200px;
  padding: 10px;
  border-radius: 3px;
  border: 1px solid var(--border-default);
  background: var(--bg-primary-default);
  color: var(--b-text-main);
  cursor: default;
  position: relative;

  &.selectable {
    cursor: pointer;
  }

  &.workload-card {
    border-color: var(--border-hover);
  }

  &.selected {
    border-color: var(--border-pressed);
    box-shadow: 0 0 0 1px var(--border-pressed);
  }

  &:hover {
    border-color: var(--border-strong);
  }
}

.card-icon-box {
  width: 24px;
  height: 24px;
  border-radius: 3px;
  flex-shrink: 0;
  background: $body-background-color;

  i {
    font-size: 12px;
    color: var(--b-text-sub);
  }

  &.workload-icon {
    background: $diagram-workload-icon-bg;

    i {
      color: var(--b-text-main);
    }
  }
}

.min-width-0 {
  min-width: 0;
}

.card-details {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--border-default);
}

.detail-row {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;

  & + & {
    margin-top: 3px;
  }
}

.expand-action-row {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--space-xs);
}

// Vue Flow renders <Handle> as a portaled <div>; scoped attrs don't reach it,
// so target the class with :deep to hide the handle dot without removing the
// connection point.
:deep(.card-handle) {
  opacity: 0;
  pointer-events: none;
}
</style>
