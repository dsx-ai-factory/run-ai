<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <span :class="getStatusClass"
    >{{ options.status }} <span v-if="displayUpdateTime" aid="status-update-time">({{ updateTime }})</span></span
  >
  <q-spinner-hourglass class="on-left" v-if="options.displayAnimation" />
  <template v-if="options.tooltipText">
    <q-btn
      @click.stop
      class="q-ml-sm q-pa-none info-btn runai-icon-btn--default"
      icon="ph-bold ph-info"
      flat
      round
      aria-label="Show status information"
    >
      <q-menu class="q-pa-sm" v-if="options.tooltipText" anchor="center right"
        ><!-- eslint-disable-next-line vuejs-accessibility/click-events-have-key-events, vuejs-accessibility/no-static-element-interactions -->
        <div
          class="tooltip"
          :style="tooltipStyle"
          v-dompurify-html="options.tooltipText"
          @click="handleToolTipClick"
        ></div
      ></q-menu>
    </q-btn>
  </template>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

//utils
import { dateUtil } from "@/utils/date.util";

// models
import type { IStatusColOptions } from "@/models/table.model";

export default defineComponent({
  name: "runai-status",
  emits: ["tooltip-clicked"],
  props: {
    options: {
      type: Object as PropType<IStatusColOptions>,
      required: true,
    },
  },
  computed: {
    displayUpdateTime(): boolean {
      if (!this.updateTime) return false;
      return !this.options.displayAnimation;
    },
    updateTime(): string {
      if (this.options.statusUpdatedTimeInMs === undefined || this.options.statusUpdatedTimeInMs === null) return "";

      const timeAgo = dateUtil.timeAgo(this.options.statusUpdatedTimeInMs);

      if (timeAgo.startsWith("-")) {
        return "";
      }

      return timeAgo;
    },
    getStatusClass(): string {
      if (this.options.color) {
        return this.options.color.toLowerCase() + "-color";
      }
      return "";
    },
    tooltipStyle(): { maxWidth?: string } {
      return this.options.tooltipMaxWidth ? { maxWidth: this.options.tooltipMaxWidth } : {};
    },
  },
  methods: {
    handleToolTipClick(event: Event) {
      this.$emit("tooltip-clicked", event);
    },
  },
});
</script>
<style lang="scss" scoped>
.tooltip {
  font-size: 12px;
}
.success-color {
  color: $success;
}
.negative-color {
  color: $negative;
}
.warning-color {
  color: $warning;
}
.info-btn {
  color: var(--icon-default);
}
</style>
