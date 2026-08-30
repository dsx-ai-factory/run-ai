<template>
  <div class="runai-details-panel-row text-b-body1 row no-wrap" :aid="aid">
    <div class="label-cell">{{ label }}</div>
    <div class="value-cell">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

defineOptions({ name: "runai-details-panel-row" });

const props = defineProps<{
  label: string;
}>();

const aid = computed(() => `details-panel-row-${props.label.toLowerCase().replace(/\s+/g, "-")}`);
</script>

<style lang="scss" scoped>
.runai-details-panel-row {
  gap: var(--space-l);
  align-items: flex-start;

  .label-cell {
    /* Fix the label cell to 23px so its text vertically centers within that box. For tall
       value cells the label keeps the same 23px height, anchored at the top of the row —
       which visually puts the label at the top with natural padding above. */
    min-height: 23px;
    max-height: 23px;
    width: 170px;
    flex-shrink: 0;
    color: var(--b-text-sub);
    display: flex;
    align-items: center;
  }

  .value-cell {
    /* Always vertically centered in the row, regardless of the row's height. For a single-line
       value the cell is centered inside the 23px row driven by the label. For a multi-line
       value the cell takes the full row height, so centering is a no-op. */
    align-self: center;
    flex: 1 1 auto;
    min-width: 0;
    color: var(--b-text-main);
    word-break: break-word;
  }
}
</style>
