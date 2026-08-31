<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="workload-node-pool-list-item" :aid="itemAid">
    {{ nodePool }}

    <runai-tooltip
      v-if="!disabled"
      icon="ph-bold ph-x"
      @click="$emit('remove')"
      tooltip-text="Remove"
      class="remove-button"
      v-show="removeable"
    />
  </div>
</template>

<script lang="ts">
import { type PropType, defineComponent } from "vue";

// cmps
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { stringUtil } from "@/utils/string.util/string.util";

export default defineComponent({
  name: "workload-node-pool-list-item",
  components: {
    RunaiTooltip,
  },
  emits: ["remove"],
  props: {
    nodePool: {
      type: String as PropType<string>,
      required: true,
    },
    removeable: {
      type: Boolean as PropType<boolean>,
      false: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  computed: {
    itemAid(): string {
      return `node-pool-${stringUtil.slug(this.nodePool)}-list-item`;
    },
  },
});
</script>

<style lang="scss" scoped>
.workload-node-pool-list-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
