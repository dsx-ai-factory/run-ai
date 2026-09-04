<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="row items-center breadcrumbs-wrapper" v-if="breadcrumbs.length > 1">
    <template v-for="(breadcrumb, index) in breadcrumbs" :key="index">
      <div v-if="index !== 0" class="q-mr-xs">/</div>
      <q-btn
        v-if="index !== breadcrumbs.length - 1"
        :label="breadcrumb"
        @click="$emit('breadcrumb-clicked', index)"
        flat
        class="btn-link no-padding q-mr-xs breadcrumb"
      />
      <div v-else class="last-breadcrumb">{{ breadcrumb }}</div>
    </template>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "chart-breadcrumbs",
  emits: ["breadcrumb-clicked"],
  props: {
    breadcrumbs: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
  },
});
</script>

<style scoped lang="scss">
.breadcrumbs-wrapper {
  font-size: 12px;
  padding-left: 60px;
  .breadcrumb {
    color: $primary;
    font-size: 12px;
    font-weight: normal;
  }
  .breadcrumb:hover {
    background-color: rgba(128, 128, 128, 0.15);
    text-decoration: unset;
  }
  .last-breadcrumb {
    color: rgb(0, 0, 0);
    font-weight: bold;
    fill: rgb(0, 0, 0);
  }
}
</style>
