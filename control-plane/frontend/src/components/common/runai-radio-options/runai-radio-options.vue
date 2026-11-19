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
        <!-- QOptionGroup nests its label slot inside QRadio, so separate help controls need sibling rows. -->
        <div
          v-if="separateTooltips && optionsWithPolicyApplied.length > 0"
          class="q-option-group q-gutter-x-sm runai-option-group"
          :class="{ 'horizontal-options row': horizontal, 'scrollable-y': scrollableY }"
          role="radiogroup"
          :aria-label="title || 'Options'"
          :aria-disabled="isGroupDisabled || undefined"
        >
          <div
            v-for="(opt, index) in optionsWithPolicyApplied"
            :key="index"
            class="row items-center no-wrap"
            :class="{ 'q-mr-md': !horizontal }"
            :aid="getOptionAid(opt)"
          >
            <q-radio
              ref="optionRadios"
              :model-value="modelValue"
              :val="opt.value"
              :disable="isGroupDisabled || opt.disable"
              :dense="dense"
              :tabindex="index === tabbableOptionIndex ? 0 : -1"
              :aria-label="opt.label"
              :aria-describedby="opt.toolTip ? `${radioDescriptionId}-${index}` : undefined"
              @update:model-value="$emit('update:model-value', $event)"
              @keydown="onRadioKeydown($event, index)"
            >
              <div
                class="row items-center no-wrap"
                :class="{ ellipsis: !!optionMaxWidth }"
                :style="{ 'max-width': optionMaxWidth, height: '40px' }"
              >
                <runai-svg-icon v-if="opt.icon" class="q-mr-sm" :name="opt.icon" size="30" />
                <span class="ellipsis-label ellipsis text-b-subtitle2" :class="{ 'disabled-label': opt.disable }">
                  {{ opt.label }} <q-tooltip v-if="isTextOverflow(opt.label)">{{ opt.label }}</q-tooltip>
                </span>
                <runai-svg-icon v-if="opt.svg" class="q-ml-xs" :name="opt.svg" size="14" />
                <q-tooltip
                  v-if="opt.disable && opt.disabledTooltip"
                  class="q-pa-sm"
                  :max-width="opt.disabledTooltipWidth"
                >
                  {{ opt.disabledTooltip }}
                </q-tooltip>
              </div>
            </q-radio>
            <runai-tooltip
              v-if="opt.toolTip"
              :aria-label="`More information about ${opt.label}`"
              :tooltip-text="opt.toolTip"
              tooltip-position="right"
              :width="opt.toolTipWidth ?? tooltipWidth"
              @revealed="innerToolTipShown = true"
              @hidden="innerToolTipShown = false"
            />
            <span v-if="opt.toolTip" :id="`${radioDescriptionId}-${index}`" hidden>{{ opt.toolTip }}</span>
          </div>
        </div>
        <q-option-group
          v-else-if="optionsWithPolicyApplied.length > 0"
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
import { defineComponent, useId, type PropType } from "vue";
import type { QRadio } from "quasar";
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
    separateTooltips: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      innerToolTipShown: false,
      searchTerm: "" as string,
    };
  },
  setup() {
    return { radioDescriptionId: `radio-help-${useId()}` };
  },
  computed: {
    isGroupDisabled(): boolean {
      return this.disable || this.policyRules?.canEdit === false;
    },
    tabbableOptionIndex(): number {
      if (this.isGroupDisabled) return -1;
      const selectedIndex = this.optionsWithPolicyApplied.findIndex(
        (option) => !option.disable && option.value === this.modelValue,
      );
      return selectedIndex >= 0 ? selectedIndex : this.optionsWithPolicyApplied.findIndex((option) => !option.disable);
    },
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
    async onRadioKeydown(event: KeyboardEvent, index: number): Promise<void> {
      const direction = ["ArrowDown", "ArrowRight"].includes(event.key)
        ? 1
        : ["ArrowUp", "ArrowLeft"].includes(event.key)
          ? -1
          : 0;
      const currentOption = this.optionsWithPolicyApplied.at(index);
      if (!direction || this.isGroupDisabled || !currentOption || currentOption.disable) return;

      event.preventDefault();
      const enabledOptions = this.optionsWithPolicyApplied.filter((option) => !option.disable);
      const currentIndex = enabledOptions.indexOf(currentOption);
      const nextOption = enabledOptions.at((currentIndex + direction + enabledOptions.length) % enabledOptions.length);
      if (!nextOption) return;
      if (nextOption.value !== this.modelValue) this.$emit("update:model-value", nextOption.value);
      await this.$nextTick();
      const radios = this.$refs.optionRadios as QRadio[];
      radios.find((radio) => radio.val === nextOption.value)?.$el.focus();
    },
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
