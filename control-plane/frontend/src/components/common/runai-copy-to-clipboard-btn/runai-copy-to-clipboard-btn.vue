<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-btn
    icon-button
    :size="size"
    :icon="icon"
    :aid="aid"
    :tooltip="customTooltip || 'Copy to clipboard'"
    :tooltip-anchor="tooltipAnchor"
    :tooltip-self="tooltipSelf"
    :tooltip-offset="tooltipOffset"
    :aria-label="customTooltip || 'Copy to clipboard'"
    @click="copyText"
  />
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { copyToClipboard } from "quasar";
import { alertUtil } from "@/utils/alert.util";
import { RunaiBtn } from "@/components/common/runai-btn";
import type { TTooltipLocation } from "@/components/common/runai-tooltip-wrapper/runai-tooltip-wrapper.vue";

export default defineComponent({
  name: "runai-copy-to-clipboard-btn",
  components: { RunaiBtn },
  props: {
    aid: {
      type: String as PropType<string>,
      default: "copy-to-clipboard-btn",
    },
    icon: {
      type: String as PropType<string>,
      default: "ph-fill ph-copy",
    },
    textToCopy: {
      type: String as PropType<string>,
      required: true,
    },
    size: {
      type: String as PropType<string>,
      default: "10px",
    },
    tooltipOffset: {
      type: Array as PropType<number[]>,
      required: false,
    },
    tooltipAnchor: {
      type: String as PropType<TTooltipLocation>,
      required: false,
    },
    tooltipSelf: {
      type: String as PropType<TTooltipLocation>,
      required: false,
    },
    customTooltip: {
      type: String as PropType<string>,
      required: false,
    },
  },
  methods: {
    async copyText(): Promise<void> {
      await copyToClipboard(this.textToCopy);
      this.$q.notify(alertUtil.getSuccess("Copied!"));
    },
  },
});
</script>
