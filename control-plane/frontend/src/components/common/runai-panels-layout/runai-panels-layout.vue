<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="runai-panels-layout-wrapper" :style="layoutStyle">
    <section
      class="runai-panels-layout"
      :class="{ 'runai-panels-layout--hide-right': hideRight, 'runai-panels-layout--responsive': responsiveLayout }"
    >
      <!-- Track focus from slotted controls so keyboard users can expand the rail too. -->
      <!-- eslint-disable-next-line vuejs-accessibility/no-static-element-interactions -->
      <aside
        v-if="$slots.left"
        class="panel-left panel-sticky"
        :class="{ 'panel-left--collapsed': collapseLeft, 'panel-left--expanded': expandLeft }"
        :tabindex="collapseLeft ? 0 : undefined"
        aria-label="Form summary"
        @mouseenter="leftHovered = true"
        @mouseleave="leftHovered = false"
        @focusin="leftFocused = true"
        @focusout="onLeftFocusOut"
      >
        <div class="panel-left-content">
          <slot name="left" :collapsed="collapseLeft && !expandLeft" :expanded="expandLeft"></slot>
        </div>
      </aside>

      <!-- Main Content Slot -->
      <main class="panel-main">
        <slot></slot>
      </main>

      <!-- Right Panel Slot -->
      <aside
        v-if="$slots.right"
        v-show="!hideRight"
        class="panel-right panel-sticky"
        aria-label="Additional information"
      >
        <slot name="right"></slot>
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useQuasar } from "quasar";
import type { IRunaiPanelsLayoutProps } from "@/models/runai-panels-layout.model";

defineOptions({
  name: "runai-panels-layout",
});

const props = withDefaults(defineProps<IRunaiPanelsLayoutProps>(), {
  marginTop: "24px",
  responsive: false,
  width: "820px", // 1300px - 220 * 2( side panels ) - 20 * 2 (gap)  = 820px main panel width
  leftWidth: "220px",
  rightWidth: "220px",
  collapsedLeftWidth: "56px",
  hideRightOnSmallScreen: false,
  collapseLeftOnSmallScreen: false,
});

// using computed instead of calc() scss method due to some caveats:
// when combining v-bind() and calc() in the same expression as it does not always trigger the correct calculation.
const stickyTop = computed(() => {
  const marginValue = parseInt(props.marginTop);
  // 60px is the height of the app header
  const appHeaderValue = 60;
  return `${marginValue + appHeaderValue}px`;
});

const $q = useQuasar();
const smallScreen = computed(() => $q.screen.width < 1280);
const hideRight = computed(() => props.hideRightOnSmallScreen && smallScreen.value);
const collapseLeft = computed(() => props.collapseLeftOnSmallScreen && smallScreen.value);
const leftHovered = ref(false);
const leftFocused = ref(false);
const expandLeft = computed(() => collapseLeft.value && (leftHovered.value || leftFocused.value));

const onLeftFocusOut = (event: FocusEvent) => {
  leftFocused.value = (event.currentTarget as HTMLElement).contains(event.relatedTarget as Node | null);
};

const responsiveLayout = computed(
  () => props.responsive || props.hideRightOnSmallScreen || props.collapseLeftOnSmallScreen,
);
const layoutStyle = computed(() => ({
  "--layout-width": `calc(${collapseLeft.value ? props.collapsedLeftWidth : props.leftWidth} + ${props.width} + ${hideRight.value ? "20px" : `${props.rightWidth} + 40px`})`,
  "--layout-left-width": collapseLeft.value ? props.collapsedLeftWidth : props.leftWidth,
  "--layout-right-width": props.rightWidth,
  "--layout-left-content-width": collapseLeft.value && !expandLeft.value ? props.collapsedLeftWidth : props.leftWidth,
}));
</script>

<style scoped lang="scss">
.runai-panels-layout-wrapper {
  // Give shrink-to-fit parents the preferred width, while fitting constrained containers.
  width: var(--layout-width);
  max-width: 100%;
  margin-inline: auto;
  min-width: v-bind("responsiveLayout ? '0' : '1300px'");
  display: flex;
  justify-content: center;
}

.runai-panels-layout {
  margin-top: v-bind(marginTop);
  display: grid;
  grid-template-columns: var(--layout-left-width) minmax(0, 1fr) var(--layout-right-width);
  gap: 20px;
  align-items: flex-start;
  width: 100%;
  max-width: var(--layout-width);

  &--hide-right {
    grid-template-columns: var(--layout-left-width) minmax(0, 1fr);
  }
}

.panel-main {
  grid-column: 2;
  min-width: v-bind("responsiveLayout ? '0' : '600px'");
  width: v-bind("responsiveLayout ? 'auto' : props.width");
}

.panel-left {
  grid-column: 1;
  grid-row: 1;
}

.panel-left-content {
  width: var(--layout-left-content-width);
}

.panel-left--expanded {
  // The grid keeps the mini rail's width while the content overlays its neighbour.
  z-index: 2;
}

.panel-right {
  grid-column: 3;
  grid-row: 1;
}

.panel-sticky {
  top: v-bind(stickyTop);
  position: sticky;
}
</style>
