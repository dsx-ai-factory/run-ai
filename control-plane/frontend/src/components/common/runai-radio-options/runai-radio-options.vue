<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-section :title="title" :tooltip-text="titleTooltip" :tooltip-width="titleTooltipWidth" tooltip-position="right">
    <div v-if="showSearch" class="q-pl-sm q-pr-md q-pb-sm sticky-search w-25">
      <q-input v-model="searchTerm" type="search" dense placeholder="Search...">
        <template v-slot:prepend>
          <q-icon name="ph-bold ph-magnifying-glass" size="20px" />
        </template>
      </q-input>
    </div>
    <section class="runai-radio-options column items-start">
      <runai-tooltip-wrapper
        :display-tooltip="!!disableTooltip"
        :tooltip-text="disableTooltip"
        :toggle-tooltip="innerToolTipShown"
      >
        <q-option-group
          v-if="optionsWithPolicyApplied.length > 0"
          :class="{
            'runai-option-group': true,
            'horizontal-options': horizontal,
            row: horizontal,
            'scrollable-y': scrollableY,
          }"
          :options="optionsWithPolicyApplied"
          :model-value="modelValue"
          @update:model-value="$emit('update:model-value', $event)"
          type="radio"
          :inline="horizontal"
          :disable="disable || policyRules?.canEdit === false"
          :dense="dense"
          :aria-label="title || 'Options'"
        >
          <template v-slot:label="opt">
            <div :class="{ row: !horizontal, 'items-center': !horizontal }" :aid="getOptionAid(opt)">
              <div
                class="row items-center no-wrap"
                :class="{ 'q-mr-md': !horizontal, ellipsis: !!optionMaxWidth }"
                :style="{ 'max-width': optionMaxWidth, height: '40px' }"
              >
                <div class="row items-center min-w-0 col-grow">
                  <runai-svg-icon v-if="opt.icon" class="q-mr-sm" :name="opt.icon" size="30" />
                  <span class="ellipsis-label ellipsis text-b-subtitle2" :class="{ 'disabled-label': opt.disable }"
                    >{{ opt.label }} <q-tooltip v-if="isTextOverflow(opt.label)">{{ opt.label }}</q-tooltip></span
                  >
                  <runai-svg-icon v-if="opt.svg" class="q-ml-xs" :name="opt.svg" size="14" />
                  <q-tooltip
                    v-if="opt.disable && opt.disabledTooltip"
                    class="q-pa-sm"
                    :max-width="opt.disabledTooltipWidth"
                  >
                    {{ opt.disabledTooltip }}
                  </q-tooltip>
                </div>
                <runai-tooltip
                  v-if="opt.toolTip"
                  @revealed="innerToolTipShown = true"
                  @hidden="innerToolTipShown = false"
                  :tooltip-text="opt.toolTip"
                  tooltip-position="right"
                  :width="opt.toolTipWidth ?? tooltipWidth"
                />
              </div>
            </div>
          </template>
        </q-option-group>
        <div v-else class="q-pl-sm" aid="empty-state">No results</div>
      </runai-tooltip-wrapper>
    </section>
    <q-field
      class="col-4 form-hint-translate-y-100 no-padding"
      :model-value="modelValue"
      :rules="[isRequiredByPolicy, isValidSelectionByPolicy]"
    />
  </runai-section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";

// Models
import type { ISelectOption } from "@/models/global.model";
import type { IGenericRadioOptionsPolicyRules } from "@/models/policy.model";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "runai-radio-options",
  components: {
    RunaiSection,
    RunaiTooltip,
    RunaiTooltipWrapper,
    RunaiSvgIcon,
  },
  emits: ["update:model-value"],
  props: {
    title: {
      type: String as PropType<string>,
      required: false,
    },
    options: {
      type: Array as PropType<Array<ISelectOption>>,
      required: true,
    },
    modelValue: {
      type: [String, Number, Boolean, null] as PropType<string | number | boolean | null>,
      required: true,
    },
    horizontal: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    titleTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    tooltipWidth: {
      type: String as PropType<string>,
      required: false,
    },
    optionMaxWidth: {
      type: String as PropType<string>,
      required: false,
    },
    titleTooltipWidth: {
      type: String as PropType<string>,
      required: false,
    },
    scrollableY: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: Object as PropType<IGenericRadioOptionsPolicyRules>,
      required: false,
    },
    minCharsToShowOverflowTooltip: {
      type: Number as PropType<number>,
      required: false,
    },
    showSearch: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    dense: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      innerToolTipShown: false,
      searchTerm: "" as string,
    };
  },
  computed: {
    optionsWithPolicyApplied(): ISelectOption[] {
      if (!this.policyRules?.options) return this.options;

      const allowedOpts = this.policyRules.options.map((o) => o.value);

      return this.options.map((o) => {
        const isAllowed = allowedOpts.includes(o.value as string);
        return {
          ...o,
          disable: !isAllowed,
          disabledTooltip: errorMessages.CANT_BE_SELECTED_POLICY,
        };
      });
    },
  },
  methods: {
    getOptionAid(opt: ISelectOption): string {
      if (opt.aid) return opt.aid;
      return `radio-option-${opt.label}`;
    },
    isTextOverflow(text: string): boolean {
      if (!this.minCharsToShowOverflowTooltip) {
        return false;
      }
      return text.length > this.minCharsToShowOverflowTooltip;
    },
    isRequiredByPolicy(val: string | null): boolean | string {
      if (!this.policyRules?.required) return true;
      return !!val || errorMessages.REQUIRED_FIELD;
    },
    isValidSelectionByPolicy(): boolean | string {
      if (!Array.isArray(this.policyRules?.options)) return true;
      return (
        this.policyRules.options.map((o) => o.value).includes((this.modelValue as string) || null) ||
        errorMessages.OPTION_CANT_BE_SELECTED_POLICY
      );
    },
  },
});
</script>
<style lang="scss" scoped>
.runai-radio-options {
  .scrollable-y {
    max-height: 35vh;
    overflow-y: auto;
    overflow-x: hidden;
    min-width: 258px;
  }
  .horizontal-options {
    gap: var(--space-m);
  }

  :deep(.q-radio__label) {
    padding-left: var(--space-xs);
  }

  :deep(.q-radio.disabled) {
    opacity: 1;
  }

  :deep(.q-radio.disabled .q-radio__inner) {
    color: var(--icon-disabled);
  }

  .disabled-label {
    color: var(--text-tertiary);
  }
}
</style>
