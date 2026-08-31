<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="header-wrapper">
    <div class="selects-wrapper q-pl-md q-mr-auto" :style="{ gap: gap + 'px' }">
      <slot />
    </div>
    <slot name="right"></slot>
  </div>
</template>
<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
export default defineComponent({
  name: "widget-header-select-wrapper",
  props: {
    gap: {
      type: Number as PropType<number>,
      default: 20,
    },
  },
});
</script>

<style scoped lang="scss">
.chart-widget-wrapper {
  .header-wrapper {
    display: flex;
    height: 50px;
    .selects-wrapper {
      display: flex;
    }
  }
}
</style>
