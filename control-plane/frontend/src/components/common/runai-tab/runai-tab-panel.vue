<template>
  <q-tab-panel :name="name">
    <slot />
  </q-tab-panel>
</template>

<script setup lang="ts">
export interface Props {
  name: string;
}

defineProps<Props>();
</script>
