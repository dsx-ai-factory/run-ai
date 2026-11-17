<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <q-btn
    class="runai-icon-btn--default"
    flat
    round
    size="sm"
    color="icon-default"
    :icon="icon"
    :aria-label="'More information: ' + tooltipText"
    @mouseenter="pointerIn"
    @mouseleave="pointerOut"
    @focusin="focusIn"
    @focusout="focusOut"
    @click="open"
    @keydown.esc="visible = false"
  >
    <q-tooltip
      v-model="visible"
      class="scheduling-help-popup"
      :class="[popupClass, { 'all-pointer-events': interactive }]"
      :no-parent-event="interactive"
      :max-width="width"
      :anchor="anchor"
      :self="self"
      :offset="offset"
      @mouseenter="pointerIn"
      @mouseleave="pointerOut"
      @focusin="focusIn"
      @focusout="focusOut"
      @keydown.esc="visible = false"
    >
      <slot name="content">{{ tooltipText }}</slot>
    </q-tooltip>
  </q-btn>
</template>
<script setup lang="ts">
import { onBeforeUnmount, ref } from "vue";
import type { QTooltipProps } from "quasar";
const props = withDefaults(
  defineProps<{
    tooltipText: string;
    icon?: string;
    width?: string;
    interactive?: boolean;
    popupClass?: string;
    anchor?: QTooltipProps["anchor"];
    self?: QTooltipProps["self"];
    offset?: number[];
  }>(),
  { icon: "ph-bold ph-question", width: "300px", anchor: "center right", self: "center left" },
);
const visible = ref(false);
let hideTimer: ReturnType<typeof setTimeout> | undefined;
let focused = false;
let hovered = false;
function open() {
  if (!props.interactive) return;
  clearTimeout(hideTimer);
  visible.value = true;
}
function scheduleClose() {
  if (!props.interactive) return;
  clearTimeout(hideTimer);
  // Allow the pointer or focus to cross the gap to a documentation link.
  hideTimer = setTimeout(() => {
    if (!focused && !hovered) visible.value = false;
  }, 200);
}
function pointerIn() {
  hovered = true;
  open();
}
function pointerOut() {
  hovered = false;
  scheduleClose();
}
function focusIn() {
  focused = true;
  open();
}
function focusOut() {
  focused = false;
  scheduleClose();
}
onBeforeUnmount(() => clearTimeout(hideTimer));
</script>
<style lang="scss">
.scheduling-help-popup {
  white-space: normal;
  padding: 10px;
  border-radius: 3px;
  background: white;
  color: rgba(0, 0, 0, 0.87);
  box-shadow: -1px 0 7px rgba(0, 0, 0, 0.15);
  font-family: Roboto, sans-serif;
  font-size: 12px;
  font-weight: 400;
  line-height: 16.8px;
}
</style>
