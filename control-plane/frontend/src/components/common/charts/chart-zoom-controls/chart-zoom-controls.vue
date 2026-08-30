<template>
  <div class="chart-zoom-controls" role="group" aria-label="Chart zoom controls">
    <runai-tooltip-wrapper tooltip-text="Zoom in" anchor="center left" self="center right">
      <q-btn
        flat
        dense
        round
        :icon="zoomInIcon"
        size="sm"
        class="zoom-btn"
        :disable="isMaxZoom"
        aria-label="Zoom in"
        @click="zoomIn"
        @keydown="handleKeydown"
      />
    </runai-tooltip-wrapper>
    <runai-tooltip-wrapper tooltip-text="Zoom out" anchor="center left" self="center right">
      <q-btn
        flat
        dense
        round
        :icon="zoomOutIcon"
        size="sm"
        class="zoom-btn"
        :disable="isMinZoom"
        aria-label="Zoom out"
        @click="zoomOut"
        @keydown="handleKeydown"
      />
    </runai-tooltip-wrapper>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { runaiIcons, toQuasarIconName } from "@/plugins/runai-icons.catalog";

const ZOOM_STEP = 20;
const MIN_ZOOM_PERCENT = 0;
const MAX_ZOOM_PERCENT = 50;
const zoomInIcon = toQuasarIconName(runaiIcons.action.zoomIn);
const zoomOutIcon = toQuasarIconName(runaiIcons.action.zoomOut);

const emit = defineEmits<{
  (e: "zoom-change", payload: { start: number; end: number }): void;
}>();

const currentStart = ref(MIN_ZOOM_PERCENT);
const currentEnd = ref(100 - MIN_ZOOM_PERCENT);

const isMaxZoom = computed(() => {
  const range = currentEnd.value - currentStart.value;
  return range <= MAX_ZOOM_PERCENT;
});

const isMinZoom = computed(() => {
  return currentStart.value <= MIN_ZOOM_PERCENT && currentEnd.value >= 100 - MIN_ZOOM_PERCENT;
});

function zoomIn(): void {
  if (isMaxZoom.value) return;

  const range = currentEnd.value - currentStart.value;
  const newRange = Math.max(MAX_ZOOM_PERCENT, range - ZOOM_STEP);
  const center = (currentStart.value + currentEnd.value) / 2;

  currentStart.value = Math.max(0, center - newRange / 2);
  currentEnd.value = Math.min(100, center + newRange / 2);

  emit("zoom-change", { start: currentStart.value, end: currentEnd.value });
}

function zoomOut(): void {
  if (isMinZoom.value) return;

  const range = currentEnd.value - currentStart.value;
  const newRange = Math.min(100, range + ZOOM_STEP);
  const center = (currentStart.value + currentEnd.value) / 2;

  currentStart.value = Math.max(0, center - newRange / 2);
  currentEnd.value = Math.min(100, center + newRange / 2);

  emit("zoom-change", { start: currentStart.value, end: currentEnd.value });
}

function handleKeydown(event: KeyboardEvent): void {
  if (event.key === "+" || event.key === "=") {
    event.preventDefault();
    zoomIn();
  } else if (event.key === "-") {
    event.preventDefault();
    zoomOut();
  }
}

function resetZoom(): void {
  currentStart.value = MIN_ZOOM_PERCENT;
  currentEnd.value = 100 - MIN_ZOOM_PERCENT;
}

function syncFromExternal(start: number, end: number): void {
  currentStart.value = start;
  currentEnd.value = end;
}

defineExpose({ resetZoom, syncFromExternal });
</script>

<style lang="scss" scoped>
.chart-zoom-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  position: absolute;
  top: 4px;
  right: 8px;
  z-index: 3;

  &:focus {
    outline: none;
  }

  .zoom-btn {
    width: 24px;
    height: 24px;
    min-width: 24px;
    min-height: 24px;
    color: var(--icon-default, rgba(0, 0, 0, 0.54));

    :deep(.q-icon) {
      font-size: 20px;
    }
  }
}
</style>
