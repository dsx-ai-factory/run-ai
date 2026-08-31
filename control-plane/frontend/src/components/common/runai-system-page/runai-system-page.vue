<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-system-page column bg-white">
    <header class="header row justify-between items-center q-mx-xl q-py-md q-px-sm">
      <runai-brand-logo v-if="appStore.isWhiteLabel" :height="mainLogo.height" :img="mainLogo.src" />
      <runai-distribution-logo
        v-else
        v-bind="DISTRIBUTION_LOGO_PRESETS.systemPage"
        class="system-page-distribution-logo"
      />
      <slot name="header-end" />
    </header>
    <slot />
  </section>
</template>

<script setup lang="ts">
// cmps
import { RunaiBrandLogo } from "@/components/common/runai-brand-logo";
import { DISTRIBUTION_LOGO_PRESETS, RunaiDistributionLogo } from "@/components/common/runai-distribution-logo";

defineOptions({
  name: "runai-system-page",
});

import { computed } from "vue";
import { useAppStore } from "@/stores/app.store";

const appStore = useAppStore();
appStore.setAppLoaded(true);
appStore.setPageLoading(false);

const mainLogo = computed(() => appStore.mainLogoDark);
</script>

<style lang="scss" scoped>
.runai-system-page {
  position: relative;
  min-height: 100vh;

  // The header is taken out of the flow so the slotted content is centered
  // against the full viewport rather than the space below the header.
  .header {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 83px;
    border-bottom: 1px solid var(--border-subtle);

    .system-page-distribution-logo[data-profile="ce"] {
      overflow: visible;
    }
  }
}
</style>
