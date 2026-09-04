<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-layout view="lHr Lpr lfr" container style="height: 100vh">
    <runai-header v-if="!fullPage" :breadcrumbs-full-path="sideNavMiniState"></runai-header>
    <div class="layout-navigation">
      <runai-side-nav
        v-if="showSideNav"
        :initial-state="sideNavMiniState"
        @drawer-state-changed="onSideNavStateChanged"
      ></runai-side-nav>
    </div>
    <q-page-container class="main-content">
      <slot />
    </q-page-container>
  </q-layout>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiSideNav } from "../runai-side-nav";
import { RunaiHeader } from "../runai-header";

import { storageUtil } from "@/utils/storage.util";

export default defineComponent({
  name: "new-layout",
  components: {
    RunaiSideNav,
    RunaiHeader,
  },
  data() {
    return {
      sideNavMiniState: false,
    };
  },
  created() {
    this.sideNavMiniState = storageUtil.get<boolean>("sideNaveState") ?? false;
  },
  computed: {
    showSideNav(): boolean {
      return !this.$route.meta.fullWidth;
    },
    fullPage(): boolean {
      return !!this.$route.meta.fullPage;
    },
  },
  methods: {
    onSideNavStateChanged(miniState: boolean): void {
      this.sideNavMiniState = miniState;
      storageUtil.save<boolean>("sideNaveState", miniState);
    },
  },
});
</script>
<style scoped lang="scss">
.layout-navigation {
  position: relative;
  z-index: 10000;
}

.main-content {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}
</style>
