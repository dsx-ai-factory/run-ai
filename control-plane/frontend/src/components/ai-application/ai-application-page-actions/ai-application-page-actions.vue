<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="row items-center justify-start ai-application-page-actions">
    <!-- Show Details - Always available when an application is selected -->
    <runai-action-button btn-action="showDetails" aid="show-ai-application-details-btn" :disable="!application" />
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import type { IApplicationTable } from "@/models/ai-application.model";

export default defineComponent({
  name: "ai-application-page-actions",
  components: {
    RunaiActionButton,
  },
  props: {
    application: {
      type: [Object, null] as PropType<IApplicationTable | null>,
      default: null,
    },
  },
});
</script>

<style lang="scss" scoped>
.ai-application-page-actions {
  gap: 8px;
}
</style>
