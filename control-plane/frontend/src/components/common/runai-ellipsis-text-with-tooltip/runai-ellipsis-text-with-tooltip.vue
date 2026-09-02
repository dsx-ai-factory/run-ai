<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <span ref="wrapperRef" class="runai-ellipsis-text-with-tooltip">
    {{ text }}
    <q-tooltip v-if="isTruncated" :anchor="anchor" :self="self" :offset="[0, 4]" max-width="400px">
      {{ text }}
    </q-tooltip>
    <q-resize-observer @resize="checkTruncation" />
  </span>
</template>

<script setup lang="ts">
import { nextTick, ref } from "vue";

type TTooltipAnchor =
  | "top left"
  | "top middle"
  | "top right"
  | "center left"
  | "center middle"
  | "center right"
  | "bottom left"
  | "bottom middle"
  | "bottom right";

withDefaults(
  defineProps<{
    text: string;
    anchor?: TTooltipAnchor;
    self?: TTooltipAnchor;
  }>(),
  { anchor: "top middle", self: "bottom middle" },
);

const wrapperRef = ref<HTMLElement | null>(null);
const isTruncated = ref(false);

const checkTruncation = async (): Promise<void> => {
  await nextTick();
  const el = wrapperRef.value;
  if (el) isTruncated.value = el.scrollWidth > el.clientWidth;
};
</script>

<style scoped lang="scss">
.runai-ellipsis-text-with-tooltip {
  display: block;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
