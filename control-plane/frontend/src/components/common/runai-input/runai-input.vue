<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-tooltip-wrapper :display-tooltip="disableInput" :tooltip-text="tooltipText">
    <q-input
      class="runai-input"
      aid="runai-input"
      :model-value="value"
      @update:model-value="updateValue"
      :disable="disableInput"
      no-error-icon
      :stack-label="stackLabel"
      :placeholder="placeholder"
      :label="label"
    />
  </runai-tooltip-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { StringRules } from "@/swagger-models/policy-service-client";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";

export default defineComponent({
  components: {
    RunaiTooltipWrapper,
  },
  props: {
    modelValue: {
      type: [String, null] as PropType<string | null>,
      required: true,
    },
    defaultValue: {
      type: [String, null] as PropType<string | null>,
      required: true,
    },
    policyRules: {
      type: Object as PropType<StringRules>,
      required: false,
    },
    policyDefault: {
      type: String as PropType<string>,
      required: false,
    },
    label: {
      type: String as PropType<string>,
      required: false,
    },
    tooltip: {
      type: String as PropType<string>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    placeholder: {
      type: String as PropType<string>,
      required: false,
    },
    stackLabel: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  created() {
    // Updating the model if the policy allow to edit.
    if (this.policyRules && this.policyDefault !== undefined && this.policyRules.canEdit) {
      this.updateValue(this.policyDefault);
    } else if (this.policyDefault === undefined && this.modelValue === null) {
      this.updateValue(this.defaultValue);
    }
  },
  computed: {
    value(): string | null {
      // Showing the policy default if it's not allowed to edit. Otherwise, showing the model value.
      if (this.policyRules && this.policyDefault !== undefined && !this.policyRules.canEdit) {
        return this.policyDefault;
      }
      return this.modelValue;
    },
    disableInput(): boolean {
      if (this.policyRules?.canEdit === false) return true;
      return !!this.disable;
    },
    tooltipText(): string {
      if (this.policyRules?.canEdit === false) {
        return "This value has been defined by the administrator";
      }
      return this.tooltip || "";
    },
  },
  methods: {
    updateValue(val: string | number | null): void {
      this.$emit("update:model-value", val);
    },
  },
});
</script>
<style lang="scss" scoped></style>
