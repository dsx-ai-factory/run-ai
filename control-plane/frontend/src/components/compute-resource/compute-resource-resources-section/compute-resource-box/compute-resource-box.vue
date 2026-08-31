<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="compute-resource-box q-my-xs row">
    <div class="inputs-section q-pa-md bg-grey-2">
      <div class="title">
        {{ title }}
      </div>
      <div class="row items-center no-wrap">
        <slot name="inputs"></slot>
      </div>
    </div>
    <div class="text-section">
      <ul class="info-list">
        <li v-for="message in texts" :key="message">● {{ message }}</li>
      </ul>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "compute-resource-box",
  props: {
    texts: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    title: {
      type: String,
      required: true,
    },
  },
});
</script>

<style lang="scss" scoped>
.compute-resource-box {
  height: 176px;

  .inputs-section {
    display: flex;
    flex-direction: column;
    justify-content: center;
    .title {
      padding-bottom: 20px;
      font-weight: 700;
    }

    width: 541px;
  }

  .text-section {
    border: 1px solid $black-10;
    border-left: none;
    font-size: 11px;
    width: 225px;
    padding: 3px;
    ul {
      margin-left: 6px;
    }
  }
}
</style>
