<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="row items-center avg-utilization-summary" aid="avg-utilization-summary">
    <div class="avg-item" v-for="(avg, i) in avgData" :key="avg.name">
      <div class="avg-item-content-wrapper">
        <div class="avg-title">{{ avg.name }}</div>
        <div class="avg-value">{{ avg.value }}</div>
      </div>
      <q-separator vertical v-if="i !== avgData.length - 1" />
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
export default defineComponent({
  name: "average-utilization-summary",
  props: {
    avgData: {
      type: Array as PropType<Array<{ name: string; value: string }>>,
      default: () => [],
    },
  },
});
</script>

<style scoped lang="scss">
.avg-utilization-summary {
  .avg-item {
    display: flex;
    margin-right: 30px;
    .avg-item-content-wrapper {
      margin-right: 30px;
    }
    .avg-title {
      font-size: 12px;
      color: $black-54;
    }
    .avg-value {
      font-size: 16px;
      color: $black-87;
      font-weight: 700;
    }
    .q-separator--vertical {
      height: 34px;
      align-self: center;
      background-color: $black-12;
    }
  }
}
</style>
