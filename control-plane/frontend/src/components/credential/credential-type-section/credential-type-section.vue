<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="credential-type-section" label="Type" hide-expend-icon>
    <template #subheader>
      <div class="row items-center">
        <span>{{ typeName }}</span>
      </div>
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
// models
import type { PropType } from "vue";

export default defineComponent({
  components: {
    RunaiExpansionItem,
  },
  props: {
    typeName: {
      type: String as PropType<string>,
      required: true,
    },
  },
});
</script>
