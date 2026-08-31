<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-btn
    v-if="linkText"
    flat
    color="primary"
    class="widget-link-btn btn-link no-padding"
    :label="linkText"
    @click="$emit('click')"
    :style="minHeight ? { minHeight } : {}"
  >
    <q-tooltip v-if="entityType" anchor="top left" self="top end"
      >Go to {{ entityType }}s to view more details</q-tooltip
    >
  </q-btn>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { EWidgetEntity } from "@/models/chart.model";

export default defineComponent({
  name: "widget-link-btn",
  emits: ["click"],
  props: {
    entityType: {
      type: String as PropType<EWidgetEntity>,
      required: false,
    },
    linkText: {
      type: String as PropType<string>,
      required: true,
    },
    minHeight: {
      type: String as PropType<string>,
      required: false,
    },
  },
});
</script>

<style scoped lang="scss">
.widget-link-btn {
  padding-right: 0;
  font-size: 12px;
}
</style>
