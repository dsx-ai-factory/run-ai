<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="large-shm-toggle row items-center">
    <runai-toggle
      label="Increase shared memory size"
      left-label
      :model-value="modelValue"
      @update:model-value="$emit('update:model-value', $event)"
      aid="shm-toggle"
      :policy-rules="policyRules"
      :disable="disable"
    />
    <runai-tooltip
      tooltip-text="When enabled, the shared memory size available to the workload will be increased from the default 64MB to the node's total available memory or the CPU memory limit, if set above"
      tooltip-position="right"
      width="380px"
    />
    <q-tooltip anchor="center middle" v-if="tooltipText">{{ tooltipText }}</q-tooltip>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiToggle } from "@/components/common";
import type { BooleanRules } from "@/swagger-models/policy-service-client";

export default defineComponent({
  name: "large-shm-toggle",
  components: {
    RunaiTooltip,
    RunaiToggle,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<BooleanRules | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
  },
});
</script>
