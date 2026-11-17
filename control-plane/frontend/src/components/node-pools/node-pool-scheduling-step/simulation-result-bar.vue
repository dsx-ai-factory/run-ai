<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <div class="simulation-values" :class="{ 'simulation-values-outside': outside }">
    <q-resize-observer @resize="measure" />
    <div class="simulation-segments" @transitionend="measure">
      <span ref="allocatedSegment" class="allocated" :style="{ width: `${width(allocated)}%`, background }">
        <span ref="allocatedLabel" class="segment-value" :aria-hidden="outside">{{ number(allocated) }}</span>
      </span>
      <span ref="unmetSegment" class="unmet" :style="{ width: `${width(unmet)}%` }">
        <span ref="unmetLabel" class="segment-value" :aria-hidden="outside">{{ number(unmet) }}</span>
      </span>
      <img class="allocation-marker" :src="weightLegend" alt="" :style="{ left: `${width(allocated)}%` }" />
    </div>
    <div v-if="outside" class="simulation-outside-labels">
      <span
        >Allocated <strong>{{ number(allocated) }}</strong></span
      >
      <span
        >Unmet <strong>{{ number(unmet) }}</strong></span
      >
    </div>
  </div>
</template>
<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from "vue";
import weightLegend from "@/assets/icons/scheduling/weight-legend.png";
const props = defineProps<{ allocated: number; unmet: number; capacity: number; background: string }>();
const allocatedSegment = ref<HTMLElement>();
const unmetSegment = ref<HTMLElement>();
const allocatedLabel = ref<HTMLElement>();
const unmetLabel = ref<HTMLElement>();
const outside = ref(false);
const number = (value: number) => Number(value.toFixed(2));
const width = (value: number) => Math.min(100, Math.max(0, (value / (props.capacity || 1)) * 100));
function measure() {
  if (!allocatedSegment.value?.parentElement?.clientWidth) return;
  const barWidth = allocatedSegment.value.parentElement.clientWidth;
  outside.value = [
    { segment: allocatedSegment.value, label: allocatedLabel.value, value: props.allocated },
    { segment: unmetSegment.value, label: unmetLabel.value, value: props.unmet },
  ].some(
    ({ segment, label, value }) =>
      segment &&
      label &&
      label.getBoundingClientRect().width + 8 > Math.min(segment.clientWidth, (barWidth * width(value)) / 100),
  );
}
watch(
  () => [props.allocated, props.unmet, props.capacity],
  () => void nextTick(measure),
  { flush: "post" },
);
onMounted(() => {
  measure();
  void document.fonts?.ready.then(measure);
});
</script>
<style scoped lang="scss">
.simulation-values {
  position: relative;
  width: 100%;
  height: 100%;
}
.simulation-segments {
  position: relative;
  display: flex;
  width: 100%;
  height: 100%;
  overflow: hidden;
  border-radius: 2px;
}
.allocated,
.unmet {
  height: 100%;
  text-align: center;
  -webkit-text-stroke: 1.4px white;
  paint-order: stroke fill;
  transition: width 180ms ease-out;
}
.allocated {
  flex: 0 0 auto;
  background: #90caf9;
  overflow: hidden;
}
.unmet {
  flex: 0 1 auto;
  min-width: 0;
  background: #ffe082;
}
.segment-value {
  display: inline-block;
  white-space: nowrap;
}
.simulation-values-outside .segment-value {
  visibility: hidden;
}
.simulation-outside-labels {
  position: absolute;
  top: calc(100% + 12px);
  left: 0;
  width: 100%;
  display: flex;
  justify-content: space-between;
  gap: 8px;
  font-size: 11px;
  line-height: 16px;
  white-space: nowrap;
  pointer-events: none;
}
.allocation-marker {
  position: absolute;
  top: 50%;
  width: 4px;
  height: 38px;
  transform: translate(-3px, -50%);
  transition: left 180ms ease-out;
}
@media (prefers-reduced-motion: reduce) {
  .allocated,
  .unmet,
  .allocation-marker {
    transition: none;
  }
}
</style>
