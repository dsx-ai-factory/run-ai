<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-btn
    :icon="icon"
    :size="size"
    class="runai-icon-btn--default"
    flat
    round
    :ripple="ripple"
    :color="color"
    :class="{ 'runai-tooltip-icon': !color }"
    :aria-label="ariaLabel || 'More information: ' + tooltipText"
    @mouseenter="pointerIn"
    @mouseleave="pointerOut"
    @focusin="focusIn"
    @focusout="focusOut"
    @click="open"
    @keydown.esc="close"
  >
    <q-tooltip
      class="q-pa-sm runai-tooltip"
      :class="[popupClass, { 'dark-mode': darkMode, 'all-pointer-events': interactive }]"
      :model-value="interactive ? visible : null"
      @update:model-value="visible = $event"
      :no-parent-event="interactive"
      :max-width="width"
      :anchor="anchor || anchorPosition"
      :self="self || selfPosition"
      :offset="offset"
      @mouseenter="pointerIn"
      @mouseleave="pointerOut"
      @focusin="focusIn"
      @focusout="focusOut"
      @keydown.esc="close"
      @show="$emit('revealed')"
      @hide="$emit('hidden')"
    >
      <slot name="content">{{ tooltipText }} </slot>
    </q-tooltip>
  </q-btn>
</template>

<script lang="ts">
import { defineComponent, onBeforeUnmount, ref, type PropType } from "vue";
import type { QTooltipProps } from "quasar";
export type TPosition = "bottom middle" | "top middle" | "center left" | "center right";
export default defineComponent({
  name: "runai-tooltip",
  props: {
    interactive: {
      type: Boolean,
      default: false,
    },
    popupClass: String,
    anchor: String as PropType<QTooltipProps["anchor"]>,
    self: String as PropType<QTooltipProps["self"]>,
    offset: Array as PropType<number[]>,
    tooltipText: {
      type: String as PropType<string>,
      required: true,
    },
    tooltipPosition: {
      type: String as PropType<string>,
      default: "top",
    },
    width: {
      type: String as PropType<string>,
      default: "250px",
      required: false,
    },
    icon: {
      type: String as PropType<string>,
      default: "ph-bold ph-question",
      required: false,
    },
    size: {
      type: String as PropType<string>,
      default: "sm",
      required: false,
    },
    darkMode: {
      type: Boolean as PropType<boolean>,
      default: false,
      required: false,
    },
    ripple: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    color: {
      type: String as PropType<string>,
      required: false,
      default: "icon-default",
    },
    ariaLabel: {
      type: String as PropType<string>,
      required: false,
    },
  },
  setup(props) {
    const visible = ref(false);
    let hideTimer: ReturnType<typeof setTimeout> | undefined;
    let focused = false;
    let hovered = false;
    function open() {
      if (!props.interactive) return;
      clearTimeout(hideTimer);
      visible.value = true;
    }
    function close() {
      clearTimeout(hideTimer);
      visible.value = false;
    }
    function scheduleClose() {
      if (!props.interactive) return;
      clearTimeout(hideTimer);
      // Allow the pointer or focus to cross the gap to interactive popup content.
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
    return { visible, open, close, pointerIn, pointerOut, focusIn, focusOut };
  },
  computed: {
    selfPosition(): TPosition {
      switch (this.tooltipPosition) {
        case "top":
          return "bottom middle";
        case "bottom":
          return "top middle";
        case "right":
          return "center left";
        case "left":
          return "center right";
        default:
          return "bottom middle";
      }
    },
    anchorPosition(): TPosition {
      switch (this.tooltipPosition) {
        case "top":
          return "top middle";
        case "bottom":
          return "bottom middle";
        case "right":
          return "center right";
        case "left":
          return "center left";
        default:
          return "top middle";
      }
    },
  },
});
</script>

<style lang="scss">
.runai-tooltip-icon {
  color: var(--icon-default);

  .q-icon {
    color: var(--icon-default);
  }
}

.runai-tooltip {
  white-space: normal;

  &.dark-mode {
    background-color: var(--bg-inverse-default);
    color: var(--text-inverse);
  }
}
</style>
