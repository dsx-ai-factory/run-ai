<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <component :is="tag" v-dompurify-html="formatHtml"></component>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { formatMarkdown } from "@/services/infra/markdown.service/markdown.service";

export default defineComponent({
  props: {
    tag: {
      type: String,
      required: false,
      default: "div",
    },
    html: {
      type: String,
      required: false,
    },
  },
  computed: {
    formatHtml() {
      return this.html ? formatMarkdown(this.html) : "";
    },
  },
});
</script>
