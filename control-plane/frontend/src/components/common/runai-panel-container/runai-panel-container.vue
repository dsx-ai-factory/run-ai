<template>
  <aside class="runai-panel-container" :class="{ 'runai-panel-container--transparent': transparent }">
    <slot></slot>
  </aside>
</template>

<script setup lang="ts">
defineOptions({
  name: "runai-panel-container",
});

withDefaults(
  defineProps<{
    width?: string;
    transparent?: boolean;
  }>(),
  {
    width: "220px",
    transparent: false,
  },
);
</script>

<style scoped lang="scss">
.runai-panel-container {
  width: v-bind(width);
  border-radius: 5px;
  align-self: flex-start;
  max-height: calc(100vh - 48px);

  &:not(.runai-panel-container--transparent) {
    background-color: white;
    box-shadow: 1px 0px 7px 0px rgba(0, 0, 0, 0.15);
  }
}
</style>
