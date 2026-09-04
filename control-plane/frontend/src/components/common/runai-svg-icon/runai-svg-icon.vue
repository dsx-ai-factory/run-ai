<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <component :is="iconComponent" :width="size"></component>
</template>
<script lang="ts">
import { defineAsyncComponent, defineComponent, type PropType } from "vue";

export default defineComponent({
  props: {
    name: {
      type: String as PropType<string>,
      required: true,
    },
    size: {
      type: String as PropType<string>,
      default: "45px",
    },
  },
  computed: {
    iconComponent() {
      return defineAsyncComponent(() => import(`../../../assets/icons/svg/${this.name}.svg`));
    },
  },
});
</script>
