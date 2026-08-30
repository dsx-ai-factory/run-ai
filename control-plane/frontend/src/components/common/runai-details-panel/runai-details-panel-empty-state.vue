<template>
  <div class="runai-details-panel-empty-state column items-center justify-start" aid="details-panel-empty-state">
    <p class="message text-b-subtitle2">No results match your filters</p>
    <i class="empty-icon ph-bold ph-funnel" aria-hidden="true" />
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: "runai-details-panel-empty-state" });
</script>

<style lang="scss" scoped>
.runai-details-panel-empty-state {
  padding: var(--space-2xl) var(--space-m);
  gap: var(--space-m);
  color: var(--b-text-main);

  .message {
    margin: 0;
  }

  .empty-icon {
    font-size: 48px;
    color: var(--b-text-sub);
  }
}
</style>
