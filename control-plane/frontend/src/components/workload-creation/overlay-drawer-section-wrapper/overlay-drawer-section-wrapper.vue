<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="overlay-drawer-section-wrapper full-height column no-wrap">
    <section class="section-wrapper-header">
      <q-item class="header-title row items-center q-pa-md full-width" :style="headerStyle">
        <div class="q-mr-auto row items-center">
          <slot name="header-title"></slot>
        </div>
        <q-btn
          class="btn-icon-lg runai-icon-btn--default"
          icon="ph-bold ph-x"
          flat
          round
          size="12px"
          @click="$emit('close')"
          color="black-54"
          aria-label="Close"
        >
          <q-tooltip>Close</q-tooltip>
        </q-btn>
      </q-item>
      <slot name="header-main"></slot>
    </section>

    <q-scroll-area class="main-content" style="flex: 1 1">
      <slot name="main"></slot>
    </q-scroll-area>

    <section class="section-wrapper-footer">
      <slot name="footer"></slot>
    </section>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "overlay-drawer-section-wrapper",
  emits: ["close"],
  props: {
    headerStyle: {
      type: Object as () => Record<string, string>,
      default: () => ({}),
    },
  },
});
</script>
<style lang="scss" scoped>
.header-title {
  height: 60px;
  font-size: 1.14285714286em;
  font-weight: 500;
  border-bottom: 1px $black-10 solid;
}
</style>
