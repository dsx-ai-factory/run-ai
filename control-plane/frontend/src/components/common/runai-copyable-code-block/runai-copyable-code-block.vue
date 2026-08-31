<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="copy-command-wrapper">
    <span v-if="label" class="text-b-subtitle2">{{ label }}</span>
    <section class="row items-center copy-command-box">
      <div class="col-11" :class="{ 'code-with-colors': useColors }">
        <pre v-if="codeWithHtml" :style="codeBlockStyle" class="text-b-mono" v-dompurify-html="codeWithHtml" />
        <pre v-else :style="codeBlockStyle" class="text-b-mono" v-text="code" />
      </div>
      <div class="col flex justify-end">
        <runai-copy-to-clipboard-btn :tooltip-anchor="tooltipAnchor" :text-to-copy="code" />
      </div>
    </section>
  </div>
</template>

<script lang="ts">
import { RunaiCopyToClipboardBtn } from "@/components/common/runai-copy-to-clipboard-btn";
import { defineComponent, type PropType, type StyleValue } from "vue";

export default defineComponent({
  name: "runai-copyable-code-block",
  components: { RunaiCopyToClipboardBtn },
  props: {
    code: {
      type: String,
      required: true,
    },
    codeWithHtml: {
      type: String,
      required: false,
    },
    maxHeight: {
      type: Number,
      required: false,
    },
    useColors: {
      type: Boolean,
      default: false,
    },
    tooltipAnchor: {
      type: String as PropType<
        | "top left"
        | "top middle"
        | "top right"
        | "top start"
        | "top end"
        | "center left"
        | "center middle"
        | "center right"
        | "center start"
        | "center end"
        | "bottom left"
        | "bottom middle"
        | "bottom right"
        | "bottom start"
        | "bottom end"
      >,
      required: false,
    },
    label: {
      type: String,
      required: false,
    },
  },
  computed: {
    codeBlockStyle(): StyleValue {
      return {
        "max-height": this.maxHeight ? this.maxHeight + "px" : "unset",
        "overflow-y": "auto",
      };
    },
  },
});
</script>

<style scoped lang="scss">
.copy-command-wrapper {
  display: flex;
  flex-direction: column;
  gap: var(--space-2xs);
}

.copy-command-box {
  background: var(--b-bg-secondary-default);
  padding: 10px var(--space-m);
  border-radius: var(--b-border-radius-radius-xs);
  min-height: 70px;
}

pre {
  white-space: pre-line;
  word-break: break-word;
  overflow: hidden;
  text-overflow: ellipsis;
  overflow-wrap: break-word;
}
</style>
<style lang="scss">
.code-with-colors {
  .yaml-item {
    color: rgba(0, 0, 0, 0.54);
  }

  .yaml-colon {
    color: rgba(0, 0, 0, 0.54);
  }

  .yaml-key {
    color: rgb(63, 110, 198);
  }

  .yaml-value {
    color: rgb(28, 125, 77);
  }
}
</style>
