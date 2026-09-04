<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <component :is="tag">
    <slot></slot>
    <div v-if="toggleWithIcon" class="row">
      <q-btn
        aid="runai-tooltip-wrapper-icon-trigger"
        @click.stop
        class="q-ml-sm q-pa-none runai-icon-btn--default text-icon-default runai-tooltip-wrapper-icon"
        :icon="icon"
        flat
        round
        size="8px"
        aria-label="Show additional information"
      >
        <q-tooltip
          v-if="tooltipText"
          class="q-pa-sm runai-tooltip-wrapper tooltip-content"
          anchor="center right"
          self="center left"
        >
          <div class="tooltip-icon-text">{{ tooltipText }}</div>
        </q-tooltip>
      </q-btn>
    </div>
    <q-tooltip
      ref="wrapperToolTip"
      v-else-if="displayTooltip && tooltipText.length > 0"
      class="runai-tooltip-wrapper tooltip-content"
      :class="{ 'disable-break-line': disableBreakLine }"
      :anchor="anchor"
      :self="self"
      :max-width="maxWidth"
      :offset="offset"
    >
      <slot name="tooltip-content">{{ tooltipText }}</slot>
    </q-tooltip>
  </component>
</template>

<script lang="ts">
import type { QTooltip } from "quasar";
import { defineComponent } from "vue";
import type { PropType } from "vue";

export type TTooltipLocation =
  | "top left"
  | "top middle"
  | "top right"
  | "top start"
  | "top end"
  | "center left"
  | "center middle"
  | "center right"
  | "center start"
  | "center end"
  | "bottom left"
  | "bottom middle"
  | "bottom right"
  | "bottom start"
  | "bottom end";

export default defineComponent({
  props: {
    tag: {
      type: String as PropType<string>,
      default: "span",
    },
    tooltipText: {
      type: String as PropType<string>,
      default: "",
    },
    displayTooltip: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    disableBreakLine: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    useHtml: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    anchor: {
      type: String as PropType<TTooltipLocation>,
      default: "center right",
    },
    self: {
      type: String as PropType<TTooltipLocation>,
      default: "center left",
    },
    toggleTooltip: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    maxWidth: {
      type: String as PropType<string>,
      required: false,
    },
    offset: {
      type: Array as PropType<number[]>,
      default: () => [14, 14],
    },
    toggleWithIcon: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    icon: {
      type: String as PropType<string>,
      default: "ph-bold ph-info",
    },
  },
  data() {
    return {
      showIconTooltip: false as boolean,
    };
  },
  watch: {
    toggleTooltip: {
      handler() {
        this.$refs["wrapperToolTip"] && (this.$refs["wrapperToolTip"] as QTooltip).toggle();
      },
    },
  },
});
</script>

<style lang="scss">
.runai-tooltip-wrapper {
  &.tooltip-content {
    &.disable-break-line {
      white-space: nowrap;
    }
  }
}
.runai-tooltip-wrapper-icon {
  color: var(--icon-default);

  .q-icon {
    color: var(--icon-default);
  }
}
.tooltip-icon-text {
  font-size: 12px;
}
</style>
