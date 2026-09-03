<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="onboarding-sidebar-footer" :class="{ 'support-unavailable': !DOCS.SUPPORT.URL }">
    <div v-if="DOCS.SUPPORT.URL" class="help-text">
      Having trouble?
      <a :href="DOCS.SUPPORT.URL" target="_blank" class="help-link q-pl-xs"> Get help </a>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import DOCS from "@/constants/docs";
export default defineComponent({
  name: "sidebar-footer",
  setup() {
    return {
      DOCS,
    };
  },
});
</script>

<style scoped lang="scss">
.onboarding-sidebar-footer {
  &.support-unavailable::before {
    content: "\00a0";
    visibility: hidden;
    font-size: 12px;
  }

  .help-text {
    font-size: 12px;
    color: $white-70;

    .help-link {
      color: white !important; // Override the default link color
      text-decoration: none;
      &:hover {
        text-decoration: underline;
      }
    }
  }
}
</style>
