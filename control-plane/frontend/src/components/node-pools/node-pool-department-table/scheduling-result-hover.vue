<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <div ref="anchor" @pointerenter="trackPointer" @pointermove="trackPointer" @focusin="positionAtValue">
    <slot />
    <q-tooltip
      ref="tooltip"
      class="tooltip-dark scheduling-result-tooltip"
      max-width="380px"
      anchor="top left"
      self="bottom middle"
      :offset="[-anchorX, verticalOffset]"
    >
      <slot name="tooltip" />
    </q-tooltip>
  </div>
</template>
<script setup lang="ts">
import { QTooltip } from "quasar";
import { nextTick, ref, watch } from "vue";
const props = defineProps<{ verticalOffset: number; focusFraction: number }>();
const anchor = ref<HTMLElement>();
const tooltip = ref<QTooltip>();
const anchorX = ref(0);
function position(x: number, width: number) {
  anchorX.value = Math.max(0, Math.min(x, width));
  void nextTick(() => tooltip.value?.updatePosition());
}
function trackPointer(event: PointerEvent) {
  const { left, width } = (event.currentTarget as HTMLElement).getBoundingClientRect();
  position(event.clientX - left, width);
}
function positionAtValue() {
  if (!anchor.value) return;
  const width = anchor.value.getBoundingClientRect().width;
  position(props.focusFraction * width, width);
}
watch(
  () => props.focusFraction,
  () => {
    if (anchor.value?.contains(document.activeElement) && document.activeElement?.matches(":focus-visible"))
      positionAtValue();
  },
);
</script>
