<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div
    class="endpoint-card flex items-start no-wrap b-gap-2xs"
    :class="{ 'application-resource-card': variant === 'application' }"
  >
    <handle :id="handleId" type="target" :position="handlePosition" class="endpoint-handle" />

    <div class="endpoint-icon-box flex items-center justify-center">
      <runai-svg-icon name="serving-endpoint" size="16px" />
    </div>

    <span class="endpoint-url text-b-body1" :title="url" :aid="`workload-structure-endpoint-${url}`">
      <runai-highlighted-text :text="url" :search-term="highlightText" />
    </span>

    <runai-copy-to-clipboard-btn
      class="endpoint-copy-btn"
      size="xs"
      :text-to-copy="url"
      custom-tooltip="Copy"
      :tooltip-offset="[0, 8]"
      tooltip-anchor="bottom middle"
      @click.stop
    />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from "@vue-flow/core";
import { RunaiCopyToClipboardBtn } from "@/components/common/runai-copy-to-clipboard-btn";
import { RunaiHighlightedText } from "@/components/common/runai-highlighted-text";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { DEFAULT_ENDPOINT_HANDLE_ID } from "./diagram-layout.util";

export interface Props {
  url: string;
  variant?: "workload" | "application";
  handleId?: string;
  handlePosition?: Position;
  highlightText?: string;
}

withDefaults(defineProps<Props>(), {
  variant: "workload",
  handleId: DEFAULT_ENDPOINT_HANDLE_ID,
  handlePosition: Position.Left,
  highlightText: "",
});
</script>

<style lang="scss" scoped>
.endpoint-card {
  width: 200px;
  padding: 10px;
  border-radius: 3px;
  border: 1px solid var(--border-default);
  background: var(--bg-primary-default);
  color: var(--b-text-main);

  &:hover {
    border-color: var(--border-strong);
  }

  &.application-resource-card {
    width: 230px;
    padding: 10px 7px;
    border-radius: 4px;
    border-color: $black-12;
    background: $white;
    color: $black-87;
  }
}

:deep(.endpoint-handle) {
  opacity: 0;
  pointer-events: none;
}

.endpoint-icon-box {
  width: 24px;
  height: 24px;
  border-radius: 3px;
  flex-shrink: 0;
  background: $diagram-endpoint-icon-bg;
}

.endpoint-url {
  flex: 1;
  min-width: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
}

.endpoint-copy-btn {
  flex-shrink: 0;
}

:deep(.copy-clipboard-btn) {
  width: 28px;
  height: 28px;
  color: var(--icon-default);
}
</style>
