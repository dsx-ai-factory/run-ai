<template>
  <div class="org-unit-legend">
    <div v-for="(item, index) in items" :key="index" class="legend-item">
      <div class="legend-indicator" :style="{ backgroundColor: item.color }" />
      <span class="text-b-body1 text-b-sub">{{ item.text }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ELegendItem } from "@/models/org-unit-forms.model";

export interface Props {
  items: ELegendItem[];
}

defineProps<Props>();
</script>

<style scoped lang="scss">
.org-unit-legend {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: 26px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.legend-indicator {
  width: 9.793px;
  height: 9.793px;
  border-radius: 100px;
  flex-shrink: 0;
}
</style>
