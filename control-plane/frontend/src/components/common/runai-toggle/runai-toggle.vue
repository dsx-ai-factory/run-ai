<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-tooltip-wrapper :display-tooltip="disableToggle" :tooltip-text="tooltipText">
    <q-toggle
      :style="{ height: '100%' }"
      :aid="aid"
      :disable="disableToggle"
      :model-value="modelValue"
      @update:model-value="updateValue"
      :left-label="leftLabel"
      :label="label"
    />
  </runai-tooltip-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { BooleanRules } from "@/swagger-models/policy-service-client";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  components: {
    RunaiTooltipWrapper,
  },
  props: {
    modelValue: {
      type: [Boolean, null] as PropType<boolean | null>,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<BooleanRules | null>,
      required: false,
    },
    label: {
      type: String as PropType<string>,
      required: false,
    },
    leftLabel: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    tooltip: {
      type: String as PropType<string>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    aid: {
      type: String as PropType<string>,
      required: false,
    },
  },
  computed: {
    disableToggle(): boolean {
      if (this.policyRules?.canEdit === false) return true;
      return !!this.disable;
    },
    tooltipText(): string {
      if (this.policyRules?.canEdit === false) {
        return errorMessages.CANT_BE_MODIFIED_POLICY;
      }
      return this.tooltip || "";
    },
  },
  methods: {
    updateValue(val: boolean): void {
      this.$emit("update:model-value", val);
    },
  },
});
</script>
<style lang="scss" scoped></style>
