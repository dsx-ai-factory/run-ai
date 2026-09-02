<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-section">
    <runai-section-title
      v-if="title"
      :title="title"
      :tooltip-text="tooltipText"
      :width="tooltipWidth"
      :tooltip-position="tooltipPosition"
      :flat="headerFlat"
    >
      <template #title-right-content>
        <slot name="title-right-content" />
      </template>
    </runai-section-title>
    <slot />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiSectionTitle } from "@/components/common/runai-section-title";

export default defineComponent({
  name: "runai-section",
  components: {
    RunaiSectionTitle,
  },
  props: {
    title: {
      type: String as PropType<string>,
      required: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
    tooltipWidth: {
      type: String as PropType<string>,
      required: false,
    },
    tooltipPosition: {
      type: String as PropType<string>,
      required: false,
    },
    headerFlat: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
});
</script>

<style scoped lang="scss">
.runai-section {
  width: 100%;
}
</style>
