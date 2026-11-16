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
    @mouseenter="openInteractive"
    @mouseleave="closeInteractive"
    @focusin="openInteractive"
    @focusout="closeInteractive"
    @click="openInteractive"
  >
    <q-tooltip
      v-model="visible"
      :no-parent-event="interactive"
      class="q-pa-sm runai-tooltip"
      :class="[tooltipClass, { 'dark-mode': darkMode, 'all-pointer-events': interactive }]"
      :max-width="width"
      :anchor="anchorPosition"
      :self="selfPosition"
      :offset="tooltipOffset"
      @show="$emit('revealed')"
      @hide="$emit('hidden')"
      @mouseenter="openInteractive"
      @mouseleave="closeInteractive"
      @focusin="openInteractive"
      @focusout="closeInteractive"
    >
      <slot name="content">{{ tooltipText }} </slot>
    </q-tooltip>
  </q-btn>
</template>

<script lang="ts">
import { defineComponent, onBeforeUnmount, ref, type PropType } from "vue";
export type TPosition = "bottom middle" | "top middle" | "center left" | "center right";
export default defineComponent({
  name: "runai-tooltip",
  props: {
    interactive: { type: Boolean, default: false },
    tooltipOffset: { type: Array as PropType<number[]>, default: undefined },
    tooltipClass: {
      type: String,
      default: undefined,
    },
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
    function openInteractive() {
      if (!props.interactive) return;
      clearTimeout(hideTimer);
      visible.value = true;
    }
    function closeInteractive() {
      if (!props.interactive) return;
      clearTimeout(hideTimer);
      hideTimer = setTimeout(() => {
        visible.value = false;
      }, 200);
    }
    onBeforeUnmount(() => clearTimeout(hideTimer));
    return { visible, openInteractive, closeInteractive };
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
