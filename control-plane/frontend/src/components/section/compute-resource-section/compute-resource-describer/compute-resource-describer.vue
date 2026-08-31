<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="compute-resource-describer q-pa-md q-mb-md">
    <span class="q-mr-sm">For {{ amount }} {{ label }}:</span>
    <span class="text-weight-bold">{{ computeResourceName }} X {{ amount }}</span>

    <span class="q-ml-md up-to-section" v-if="maxAmount && amount < maxAmount">
      <span class="q-mr-md">up to</span>
      <span class="q-mr-sm">{{ maxAmount }} {{ maxLabel }}:</span>
      <span class="text-weight-bold">{{ computeResourceName }} X {{ maxAmount }}</span>
    </span>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  props: {
    label: {
      type: String as PropType<string>,
      required: true,
    },
    computeResourceName: {
      type: String as PropType<string>,
      required: true,
    },
    amount: {
      type: Number as PropType<number>,
      required: true,
    },
    maxAmount: {
      type: Number as PropType<number>,
      required: false,
    },
    maxLabel: {
      type: String as PropType<string>,
      required: false,
    },
  },
});
</script>
<style lang="scss" scoped>
.compute-resource-describer {
  color: $black-54;
  border: 1px $black-12 solid;
}
</style>
