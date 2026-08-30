<template>
  <div class="runai-panels-layout-wrapper">
    <section class="runai-panels-layout">
      <!-- Left Panel Slot -->
      <aside v-if="$slots.left" class="panel-left panel-sticky" aria-label="Form summary">
        <slot name="left"></slot>
      </aside>

      <!-- Main Content Slot -->
      <main class="panel-main" :style="mainContentStyle">
        <slot></slot>
      </main>

      <!-- Right Panel Slot -->
      <aside v-if="$slots.right" class="panel-right panel-sticky" aria-label="Additional information">
        <slot name="right"></slot>
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

defineOptions({
  name: "runai-panels-layout",
});

const props = withDefaults(
  defineProps<{
    marginTop?: string;
    width?: string;
  }>(),
  {
    marginTop: "24px",
    width: "820px", // 1300px - 220 * 2( side panels ) - 20 * 2 (gap)  = 820px main panel width
  },
);

// using computed instead of calc() scss method due to some caveats:
// when combining v-bind() and calc() in the same expression as it does not always trigger the correct calculation.
const stickyTop = computed(() => {
  const marginValue = parseInt(props.marginTop);
  // 60px is the height of the app header
  const appHeaderValue = 60;
  return `${marginValue + appHeaderValue}px`;
});

const mainContentStyle = computed(() => {
  return {
    width: props.width,
  };
});
</script>

<style scoped lang="scss">
.runai-panels-layout-wrapper {
  width: 100%;
  display: flex;
  justify-content: center;
}

.runai-panels-layout {
  margin-top: v-bind(marginTop);
  display: grid;
  grid-template-columns: 220px 1fr 220px;
  gap: 20px;
  align-items: flex-start;
  min-width: 1300px;
}

.panel-main {
  min-width: 600px;
}

.panel-sticky {
  top: v-bind(stickyTop);
  position: sticky;
}
</style>
