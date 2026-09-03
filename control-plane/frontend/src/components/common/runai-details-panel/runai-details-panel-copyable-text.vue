<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <span class="runai-details-panel-copyable-text row items-center no-wrap">
    <span class="copyable-text">{{ text }}</span>
    <runai-copy-to-clipboard-btn :text-to-copy="text" class="copy-btn" />
  </span>
</template>

<script setup lang="ts">
import { RunaiCopyToClipboardBtn } from "@/components/common/runai-copy-to-clipboard-btn";

defineOptions({ name: "runai-details-panel-copyable-text" });

defineProps<{
  text: string;
}>();
</script>

<style lang="scss" scoped>
.runai-details-panel-copyable-text {
  gap: var(--space-xs);

  .copyable-text {
    word-break: break-all;
  }

  .copy-btn {
    color: var(--icon-default);
  }

  /* Shrink the inner q-btn (default 38x38) so the row stays at text height
     and the URL label aligns with the URL value. */
  :deep(.copy-clipboard-btn) {
    width: auto;
    height: auto;
    min-width: 0;
    min-height: 0;
    padding: 2px;
  }
}
</style>
