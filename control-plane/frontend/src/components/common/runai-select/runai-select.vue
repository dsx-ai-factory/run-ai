<template>
  <q-select
    ref="qSelectRef"
    class="runai-select"
    :class="{ 'has-hint': hint, 'runai-select--ellipsis': selectedEllipsis }"
    @popup-show="isPopupOpen = true"
    @popup-hide="isPopupOpen = false"
    :options="optionsForSelect"
    :model-value="modelValue"
    @update:model-value="updateModelValue"
    :emit-value="emitValue"
    :rules="selectRules"
    :outlined="outlined"
    no-error-icon
    :multiple="multiple"
    :hide-bottom-space="hideBottomSpace"
    :disable="isDisabled"
    :label="label"
    :loading="loading"
    :stack-label="stackLabel"
    :dense="dense"
    :clearable="!!clearIcon"
    :clear-icon="clearIcon"
    :virtual-scroll-slice-size="virtualScrollSliceSize"
    :hint="hint"
    :use-input="filterable && (!hasSelection || multiple)"
    :input-debounce="filterable ? 0 : undefined"
    :placeholder="filterable ? placeholder : undefined"
    @filter="onFilter"
  >
    <template v-slot:selected>
      <div class="row items-center selected-value-container" v-if="selectedOption">
        <runai-svg-icon
          v-if="typeof selectedOption === 'object' && selectedOption?.icon"
          :name="selectedOption.icon"
          :key="selectedOption.icon"
          size="20px"
          class="q-mr-sm"
        />
        <div v-if="selectedEllipsis" class="selected-value-wrapper">
          <span ref="selectedTextRef" class="ellipsis selected-text">{{ selectedText }}</span>
          <span
            v-if="typeof selectedOption === 'object' && selectedOption?.subLabel"
            class="q-ml-sm selected-sub-label"
            :style="{ color: selectedOption?.subLabelColor }"
            >{{ selectedOption?.subLabel }}</span
          >
          <!-- Full value on hover only while the select is closed (the open dropdown already shows it). -->
          <q-tooltip v-if="isSelectedTextOverflows && !isPopupOpen">
            {{ selectedText }}
          </q-tooltip>
        </div>
        <span v-else>
          {{ selectedText
          }}<span
            v-if="typeof selectedOption === 'object' && selectedOption?.subLabel"
            class="q-ml-sm"
            :style="{ color: selectedOption?.subLabelColor }"
            >{{ selectedOption?.subLabel }}
          </span>
        </span>
      </div>
      <!-- When filterable, the placeholder is shown by the native input (use-input), not this slot. -->
      <template v-else-if="placeholder && !filterable">
        <span class="select-placeholder">{{ placeholder }}</span>
      </template>
    </template>
    <template v-slot:option="scope">
      <q-item
        v-bind="scope.itemProps"
        :aid="scope.opt.aid"
        class="row items-center justify-between no-wrap runai-select-item"
        :class="{ 'runai-select-item--separator': scope.opt.separator }"
      >
        <div class="row items-center no-wrap runai-select-item__main">
          <runai-svg-icon v-if="scope.opt.icon" :name="scope.opt.icon" size="25px" class="q-mr-md" />
          <q-item-label class="ellipsis runai-select-item__label">{{ scope.opt.label || scope.opt }}</q-item-label>
          <role-stage-badge class="q-ml-sm" v-if="scope.opt.roleStageBadge" :role-name="scope.opt.roleStageBadge" />
        </div>

        <!-- Sub-label (e.g. a status / phase chip) pinned to the right of the row. -->
        <span
          v-if="scope.opt.subLabel"
          class="q-ml-md runai-select-item__sub-label"
          :style="{ color: scope.opt.subLabelColor }"
          >{{ scope.opt.subLabel }}</span
        >

        <!-- display a tooltip when hovering on the row -->
        <q-tooltip
          v-if="scope.opt.disable && scope.opt.disabledTooltip"
          class="q-pa-sm runai-tooltip"
          :max-width="maxWidthItemTooltip"
          anchor="bottom middle"
          self="top middle"
        >
          {{ scope.opt.disabledTooltip }}
        </q-tooltip>

        <!-- display a tooltip when hovering on the icon only -->
        <runai-tooltip
          v-if="scope.opt.warningTooltip || scope.opt.toolTip"
          tooltip-position="right"
          :tooltip-text="getTooltipText(scope.opt)"
          class="q-ml-md"
          :width="iconTooltipWidth"
          :icon="getTypeIcon(scope.opt.type)"
          :color="getTypeColor(scope.opt.type)"
        >
          <template v-slot:content v-if="scope.opt.isHtmlTooltipText">
            <span v-dompurify-html="getTooltipText(scope.opt)"></span>
          </template>
        </runai-tooltip>
      </q-item>
    </template>
    <template v-slot:no-option>
      <q-item>
        <q-item-section class="text-grey">{{ noOptionText }}</q-item-section>
      </q-item>
      <q-separator v-if="customAction?.text" />
      <q-item v-if="customAction?.text" clickable class="custom-action-item" @click="handleCustomAction">
        <q-item-section>
          <div class="row justify-end">
            <q-icon name="ph-bold ph-plus" size="16px" color="primary" />
            <q-item-label class="custom-action-item-text text-primary-clr q-ml-sm">
              {{ customAction.text }}
            </q-item-label>
          </div>
        </q-item-section>
      </q-item>
    </template>

    <!-- Custom action option at the bottom -->
    <template v-slot:after-options v-if="customAction?.text">
      <q-separator />
      <q-item class="custom-action-item">
        <q-item-section>
          <div class="row justify-end">
            <q-btn
              aid="runai-select-custom-btn"
              flat
              color="primary"
              @click="handleCustomAction"
              :label="customAction.text"
              class="btn-icon-sm q-ml-sm"
              icon="ph-bold ph-plus"
            />
          </div>
        </q-item-section>
      </q-item>
    </template>
  </q-select>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RoleStageBadge } from "@/components/common/role-stage-badge";
// Models
import type { ISelectOption, SelectUpdateFilterFunction } from "@/models/global.model";
import type { ValidationRule } from "quasar";
import { policyUtil } from "@/utils/policy.util";
import { isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { type IGenericSelectPolicyRules } from "@/models/policy.model";

// Interfaces
export interface IRunaiSelectCustomAction {
  text: string;
}

export const SELECT_TYPES = {
  warning: {
    icon: "ph-bold ph-warning-circle",
    color: "warning",
  },
  success: {
    icon: "ph-bold ph-check-circle",
    color: "success",
  },
  negative: {
    icon: "ph-bold ph-x-circle",
    color: "negative",
  },
  info: {
    icon: "ph-bold ph-info",
    color: "info",
  },
  positive: {
    icon: "ph-bold ph-check-circle",
    color: "positive",
  },
} as const;

export enum SelectType {
  Warning = "warning",
  Success = "success",
  Negative = "negative",
  Info = "info",
  Positive = "positive",
}

export default defineComponent({
  name: "runai-select",
  components: { RunaiSvgIcon, RunaiTooltip, RoleStageBadge },
  emits: ["update:model-value", "custom-action"],
  props: {
    modelValue: {
      type: [Object, String, Number, Boolean, null, undefined, Array] as PropType<
        ISelectOption | string | number | boolean | null | undefined | Array<string>
      >,
      required: true,
    },
    options: {
      type: Array as PropType<Array<ISelectOption | string | number | boolean>>,
      required: true,
    },
    placeholder: {
      type: String as PropType<string>,
      default: "",
    },
    noOptionText: {
      type: String as PropType<string>,
      default: "",
    },
    rules: {
      type: Array as PropType<Array<ValidationRule>>,
      default: () => [],
    },
    outlined: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    emitValue: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    unclearable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    clearIcon: {
      type: String as PropType<string>,
      required: false,
    },
    hideBottomSpace: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: Object as PropType<IGenericSelectPolicyRules>,
      required: false,
    },
    label: {
      type: String as PropType<string>,
      required: false,
    },
    /**
     * @param {string} iconTooltipWidth Tooltip width when hovering on the icon only.
     */
    iconTooltipWidth: {
      type: String as PropType<string>,
      required: false,
    },
    /**
     * @param {string} maxWidthItemTooltip Tooltip width when hovering a row without an icon.
     */
    maxWidthItemTooltip: {
      type: String as PropType<string>,
      default: "250px",
    },
    multiple: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    selectedEllipsis: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    stackLabel: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    /**
     * @param {IRunaiSelectCustomAction} customAction Configuration for the custom action button
     */
    customAction: {
      type: Object as PropType<IRunaiSelectCustomAction>,
      required: false,
    },
    dense: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    virtualScrollSliceSize: {
      type: Number as PropType<number>,
      required: false,
    },
    hint: {
      type: String as PropType<string>,
      required: false,
    },
    filterable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      searchTerm: "" as string,
      isSelectedTextOverflows: false,
      isPopupOpen: false,
    };
  },
  mounted() {
    if (this.selectedEllipsis) {
      this.checkTextOverflow();
    }
  },
  computed: {
    isDisabled(): boolean {
      return !!this.disable || this.policyRules?.canEdit === false;
    },
    hasSelection(): boolean {
      if (this.multiple) return Array.isArray(this.modelValue) && this.modelValue.length > 0;
      return this.modelValue !== null && this.modelValue !== undefined && this.modelValue !== "";
    },
    optionsForSelect(): Array<ISelectOption | string | number | boolean> {
      if (!this.filterable || !this.searchTerm) return this.displayedOptions;
      const needle = this.searchTerm.toLowerCase();
      return this.displayedOptions.filter((option) => {
        const optionLabel = typeof option === "object" ? (option.label ?? "").toString() : option.toString();
        return optionLabel.toLowerCase().includes(needle);
      });
    },
    displayedOptions() {
      if (!this.policyRules?.options) return this.options;
      const displayedOptions = this.options.map((option) => {
        if (typeof option !== "object") return { label: option.toString(), value: option };
        return option;
      });
      return policyUtil.getOptionsWithRules(
        displayedOptions,
        this.policyRules?.options,
        this.policyRules.isClosedList,
        this.modelValue,
      );
    },
    selectedOption(): ISelectOption | string | number | boolean | null {
      if (!this.multiple) {
        const fromOptions = this.findOption(this.options);
        if (fromOptions) return fromOptions;

        const fromDisplayedOptions = this.findOption(this.displayedOptions);
        if (fromDisplayedOptions) return fromDisplayedOptions;

        return null;
      } else {
        if (!Array.isArray(this.modelValue)) return null;

        return this.options
          .filter((option) =>
            (this.modelValue as Array<string | number | boolean | null | undefined>).includes(
              typeof option === "object" ? option.value : option,
            ),
          )
          .map((option) => {
            return typeof option === "object" ? option.label : option;
          })
          .join(", ");
      }
    },
    selectRules(): ValidationRule[] {
      return this.policyRules?.required
        ? [...this.rules, this.requiredRule, this.isValidSelection]
        : [...this.rules, this.isValidSelection];
    },
    selectedText(): string {
      if (typeof this.selectedOption === "object" && this.selectedOption) {
        return this.selectedOption?.label || "";
      }
      return this.selectedOption?.toString() || "";
    },
  },

  methods: {
    findOption(
      options: Array<ISelectOption | string | number | boolean>,
    ): ISelectOption | string | number | boolean | null {
      return (
        options.find((option) => {
          const value = typeof option === "object" && !Array.isArray(option) ? option.value : option;
          if (this.modelValue && typeof this.modelValue === "object" && !Array.isArray(this.modelValue)) {
            return value === this.modelValue.value;
          }
          return value === this.modelValue;
        }) || null
      );
    },
    updateModelValue(value: ISelectOption | string | number | boolean | null): void {
      if (this.unclearable && value === null) return;
      this.searchTerm = "";
      this.$emit("update:model-value", value);
    },
    onFilter(value: string, update: SelectUpdateFilterFunction): void {
      update(() => {
        this.searchTerm = value || "";
      });
    },
    requiredRule(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.REQUIRED_FIELD;
    },
    isValidSelection(): boolean | string {
      if (!Array.isArray(this.policyRules?.options)) return true;
      const valueToCheck =
        this.modelValue && typeof this.modelValue === "object" && !Array.isArray(this.modelValue)
          ? this.modelValue.value
          : this.modelValue;
      return (
        this.policyRules.options.map((o) => o.value).includes(valueToCheck as string) ||
        errorMessages.CANT_BE_SELECTED_POLICY
      );
    },
    getTooltipText(option: ISelectOption): string {
      return option.warningTooltip || option.toolTip || "";
    },
    checkTextOverflow(): void {
      this.$nextTick(() => {
        const textElement = this.$refs.selectedTextRef as HTMLElement | undefined;
        if (!textElement) return;
        this.isSelectedTextOverflows = textElement.scrollWidth > textElement.clientWidth;
      });
    },
    getTypeConfig(type: SelectType) {
      return SELECT_TYPES[type as keyof typeof SELECT_TYPES] || null;
    },
    getTypeIcon(type: SelectType) {
      return this.getTypeConfig(type)?.icon || undefined;
    },
    getTypeColor(type: SelectType) {
      return this.getTypeConfig(type)?.color || undefined;
    },
    handleCustomAction(): void {
      this.$emit("custom-action");
      // Close the dropdown after emitting the event
      this.$nextTick(() => {
        const qSelect = this.$refs.qSelectRef as { hidePopup?: () => void };
        if (qSelect?.hidePopup) {
          qSelect.hidePopup();
        }
      });
    },
  },
  watch: {
    selectedOption() {
      if (this.selectedEllipsis) {
        this.checkTextOverflow();
      }
    },
  },
});
</script>

<style lang="scss" scoped>
// Option row: main label takes the available width and ellipsizes when too long;
// the sub-label (status) stays on the right.
.runai-select-item__main {
  min-width: 0;
  flex: 1 1 auto;
}

.runai-select-item__label {
  min-width: 0;
}

.runai-select-item__sub-label {
  flex: 0 0 auto;
  white-space: nowrap;
}

// Single-value ellipsis: keep the field on one clipped line in every state
// (incl. focused/open) so the truncation "…" doesn't disappear or wrap on click.
// Scoped to selectedEllipsis so multi-select chip wrapping is untouched.
.runai-select--ellipsis {
  :deep(.q-field__native) {
    flex-wrap: nowrap;
    overflow: hidden;
  }
}

.selected-value-container {
  width: 100%;
  min-width: 0;

  span:first-of-type {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .selected-value-wrapper {
    display: flex;
    align-items: center;
    min-width: 0;
    flex: 1;
  }

  // Name ellipsizes; status sub-label keeps its width on the right.
  .selected-text {
    min-width: 0;
    flex: 1 1 auto;
  }

  .selected-sub-label {
    flex: 0 0 auto;
    white-space: nowrap;
  }
}

.select-placeholder {
  color: var(--text-placeholder);
}

.runai-select-item--separator {
  border-top: 1px solid $black-12;
}

.runai-select {
  :deep(.q-field__label) {
    color: var(--b-text-label);
  }

  :deep(.q-field__native),
  :deep(.q-field__input) {
    color: var(--b-text-main);
  }

  :deep(.q-field__control::before) {
    border-color: var(--border-default);
  }

  :deep(.q-field__control:hover::before) {
    border-color: var(--border-strong);
  }

  // Focus / pressed — underline + label go brand green.
  // For outlined, ::after spans the full control box, so setting `background`
  // would fill the field with the brand color; only do it for the underline variant.
  &.q-field--focused:not(.q-field--outlined) :deep(.q-field__control::after) {
    background: var(--border-pressed);
  }

  // Chained `.q-field` (in addition to `.runai-select`) bumps specificity to
  // 0,5,0 so this beats the global rule in `assets/scss/components/_input.scss`
  // (0,4,0) that sets every q-field label to `var(--b-text-label)` regardless
  // of focus.
  &.q-field.q-field--focused:not(.q-field--error) :deep(.q-field__label) {
    color: var(--text-brand);
  }

  &.q-field--outlined {
    :deep(.q-field__control) {
      border-radius: var(--space-2xs);
    }

    :deep(.q-field__control::before) {
      border-color: var(--border-default);
      border-radius: var(--space-2xs);
    }

    :deep(.q-field__control:hover::before) {
      border-color: var(--border-strong);
    }

    &.q-field--focused :deep(.q-field__control::after) {
      border-color: var(--border-pressed);
      border-radius: var(--space-2xs);
    }
  }

  &.q-field--disabled {
    :deep(.q-field__label),
    :deep(.q-field__native),
    :deep(.q-field__input) {
      color: var(--text-disabled);
    }

    :deep(.q-field__control::before) {
      border-color: var(--border-default);
    }

    .select-placeholder {
      color: var(--text-disabled);
    }
  }

  &.q-field--error {
    :deep(.q-field__label),
    :deep(.q-field__messages) {
      color: var(--text-error);
    }

    :deep(.q-field__control::before) {
      border-color: var(--border-error);
    }

    &:not(.q-field--outlined) :deep(.q-field__control::after) {
      background: var(--border-error);
    }

    &.q-field--outlined :deep(.q-field__control::after) {
      border-color: var(--border-error);
    }
  }
}

.custom-action-item {
  border-top: 1px solid $grey-2;

  .custom-action-item-text {
    text-transform: uppercase;
  }
}
</style>
