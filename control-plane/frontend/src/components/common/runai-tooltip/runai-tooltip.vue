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
  >
    <q-tooltip
      class="q-pa-sm runai-tooltip"
      :class="{ 'dark-mode': darkMode }"
      :max-width="width"
      :anchor="anchorPosition"
      :self="selfPosition"
      @show="$emit('revealed')"
      @hide="$emit('hidden')"
    >
      <slot name="content">{{ tooltipText }} </slot>
    </q-tooltip>
  </q-btn>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
export type TPosition = "bottom middle" | "top middle" | "center left" | "center right";
export default defineComponent({
  name: "runai-tooltip",
  props: {
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
