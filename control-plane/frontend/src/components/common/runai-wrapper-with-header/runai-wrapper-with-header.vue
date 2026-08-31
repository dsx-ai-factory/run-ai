<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card flat bordered square class="q-mb-md runai-wrapper-with-header">
    <q-card-section class="card-header q-mb-md">
      <span class="text-subtitle2">{{ header }}</span>
      <q-btn
        :disable="disable"
        @click="$emit('close')"
        aid="wrapper-x-btn"
        round
        flat
        class="runai-icon-btn--default"
        icon="ph-bold ph-x"
      ></q-btn>
    </q-card-section>
    <q-card-section class="card-subheader no-padding">
      <div class="text-body2 q-mb-md q-px-md">{{ subheader }}</div>
      <slot name="body"></slot>
    </q-card-section>
    <q-card-actions class="q-mx-sm q-pt-none">
      <slot name="actions"></slot>
    </q-card-actions>
  </q-card>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

export default defineComponent({
  emits: ["close"],
  props: {
    header: {
      type: String as PropType<string>,
      default: "",
    },
    subheader: {
      type: String as PropType<string>,
      default: "",
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
});
</script>
<style scoped lang="scss">
.card-header {
  background-color: $body-background-color;
  height: 50px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
