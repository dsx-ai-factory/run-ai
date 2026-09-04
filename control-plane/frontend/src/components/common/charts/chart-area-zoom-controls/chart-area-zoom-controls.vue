<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div
    class="chart-area-zoom-controls"
    :style="{ transform: `translateY(-${verticalOffset}px)` }"
    role="group"
    aria-label="Chart zoom controls"
  >
    <runai-tooltip-wrapper tooltip-text="Zoom" anchor="center left" self="center right">
      <q-btn
        flat
        dense
        round
        :icon="zoomIcon"
        size="sm"
        class="zoom-btn"
        :class="{ 'zoom-btn--active': isZoomActive }"
        aria-label="Zoom"
        :aria-pressed="isZoomActive"
        @click="toggleZoom"
      />
    </runai-tooltip-wrapper>
    <runai-tooltip-wrapper tooltip-text="Zoom reset" anchor="center left" self="center right" disable-break-line>
      <q-btn
        flat
        dense
        round
        :icon="resetIcon"
        size="sm"
        class="zoom-btn"
        aria-label="Reset chart zoom"
        @click="handleReset"
      />
    </runai-tooltip-wrapper>
    <span class="visually-hidden" aria-live="polite" aria-atomic="true">{{ announcement }}</span>
  </div>
</template>

<script lang="ts" setup>
import { ref, watch } from "vue";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";

const zoomIcon = toQuasarIconName(runaiIcons.action.zoomIn);
const resetIcon = toQuasarIconName(runaiIcons.action.reset);

const emit = defineEmits<{
  (e: "zoom-toggle", active: boolean): void;
  (e: "reset"): void;
}>();

const props = withDefaults(defineProps<{ externalActive?: boolean; verticalOffset?: number }>(), {
  verticalOffset: 0,
});

const isZoomActive = ref(false);
const announcement = ref("");

watch(
  () => props.externalActive,
  (active) => {
    if (active !== undefined) isZoomActive.value = active;
  },
);

function toggleZoom(): void {
  isZoomActive.value = !isZoomActive.value;
  announcement.value = isZoomActive.value ? "Chart zoom enabled." : "Chart zoom disabled.";
  emit("zoom-toggle", isZoomActive.value);
}

function handleReset(): void {
  isZoomActive.value = false;
  announcement.value = "Chart zoom reset.";
  emit("reset");
}

function deactivateZoom(): void {
  isZoomActive.value = false;
}

defineExpose({ deactivateZoom });
</script>

<style lang="scss" scoped>
.chart-area-zoom-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  position: absolute;
  top: 2px;
  right: 8px;
  z-index: 3;

  .zoom-btn {
    width: 32px;
    height: 32px;
    min-width: 32px;
    min-height: 32px;
    color: var(--icon-default, rgba(0, 0, 0, 0.54));

    &:hover:not(.zoom-btn--active) {
      background-color: var(--bg-primary-hover, #eeeeee);
    }

    &:focus-visible {
      outline: 2px solid var(--border-strong, #707070);
      outline-offset: 2px;
    }

    &--active {
      color: var(--interactive-secondary-default-icon, #65a800);
      background-color: var(--interactive-secondary-pressed-bg, #e0e0e0);

      &:hover {
        background-color: var(--interactive-secondary-hover-bg, #eeeeee);
      }
    }

    :deep(.q-icon) {
      font-size: 20px;
    }
  }
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
