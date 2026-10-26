<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="installation-command-box q-pa-md q-my-md">
    <div class="q-mb-md">
      <slot name="top"></slot>
    </div>
    <section class="middle row no-wrap items-center">
      <div class="commands col-11 q-pa-sm">
        <pre
          class="text-b-mono"
          ref="elCommands"
          :style="{ 'max-height': maxHeight ? maxHeight + 'px' : 'unset', 'overflow-y': 'auto' }"
        >
           <slot name="commands"/>
        </pre>
      </div>
      <runai-btn
        class="copy-btn runai-icon-btn--default"
        flat
        round
        icon="ph-fill ph-copy"
        aid="copy-installation-command"
        @click="copyText"
      >
        <q-tooltip class="tool-tip-dark">Copy to clipboard</q-tooltip>
      </runai-btn>
    </section>
    <div class="q-mt-md">
      <slot name="bottom"></slot>
    </div>
  </section>
</template>
<script lang="ts">
import { copyToClipboard } from "quasar";
import type { PropType } from "vue";
import { RunaiBtn } from "@/components/common/runai-btn";

export default {
  components: {
    RunaiBtn,
  },
  props: {
    maxHeight: {
      type: Number as PropType<number>,
      required: false,
    },
  },
  methods: {
    async copyText(): Promise<void> {
      const text: string = (this.$refs.elCommands as HTMLElement).innerText;
      await copyToClipboard(text);
    },
  },
};
</script>
<style lang="scss">
.installation-command-box {
  border: 1px solid #e0e0e0;
  .middle {
    background: #f1f3f4;
    pre {
      white-space: pre-line;
      word-break: break-word;
      overflow: hidden;
      text-overflow: ellipsis;
      overflow-wrap: break-word;
    }
  }
  .copy-btn {
    color: $black-54;
  }
  .tool-tip-dark {
    background-color: $positive;
    color: $white;
  }
}
</style>
