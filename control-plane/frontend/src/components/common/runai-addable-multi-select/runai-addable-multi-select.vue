<template>
  <div class="multi-select-wrapper" :class="{ 'dense-chip': denseChip }">
    <q-select
      ref="multiSelect"
      aid="addable-multi-select"
      :class="customClass || ''"
      use-input
      use-chips
      multiple
      clearable
      no-error-icon
      reactive-rules
      :options-dense="optionsDense"
      clear-icon="ph-bold ph-x"
      :options="filterOptions"
      :label="label"
      :model-value="selectedOptions"
      :loading="loading"
      :rules="[multiSelectValidation, requiredByPolicyValidation, allowedByPolicyValidation]"
      @new-value="createNewValue"
      @filter="filterFn"
      @clear="onClear"
      @update:model-value="onSelect"
      :disable="disable || disabledState"
      options-selected-class="test-class"
      :max-values="maxValues"
      :hint="hint"
      :counter="counter"
    >
      <template v-slot:selected-item="scope">
        <!-- Quasar's slot scope does not include the parent's disable state, so we forward it
             explicitly as `isDisabled` for consumers that override this slot. -->
        <slot name="selected-item" v-bind="{ ...scope, isDisabled: disable || disabledState }">
          <q-chip
            :text-color="selectedOptionTextColor(scope.opt)"
            :color="selectedOptionBackgroundColor(scope.opt)"
            dense
            :removable="!disable && !disabledState"
            @remove="scope.removeAtIndex(scope.index)"
            :tabindex="scope.tabindex"
          >
            {{ scope.opt.label }}
          </q-chip>
        </slot>
      </template>
      <template v-if="$slots.option" v-slot:option="scope">
        <slot name="option" v-bind="scope" />
      </template>
    </q-select>
    <q-tooltip v-if="disabledState">{{ disabledTooltip }}</q-tooltip>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

//models
import type { ISelectOption } from "@/models/global.model";
import { errorMessages } from "@/common/error-message.constant";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";
import type { QSelect } from "quasar";
// Quasar renders the synthetic "+ new" option as lowercase "(new ...)", so compare using lowercase.
const NEW_LABEL_MARKER = "(new";
export default defineComponent({
  name: "runai-addable-multi-select",
  emits: ["add-option", "options-select"],
  props: {
    options: {
      type: Array as PropType<ISelectOption[]>,
      default: () => [],
    },
    selectedOptions: {
      type: Array as PropType<ISelectOption[]>,
      default: () => [],
    },
    label: {
      type: String as PropType<string>,
      default: "",
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    newOptionLabel: {
      type: String as PropType<string>,
      default: "option",
    },
    required: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    errorMessage: {
      type: String as PropType<string>,
      default: errorMessages.SELECT_A_VALUE,
    },
    disable: {
      type: Boolean as PropType<boolean>,
    },
    maxValues: {
      type: Number as PropType<number>,
      required: false,
    },
    hint: {
      type: String as PropType<string>,
      required: false,
    },
    counter: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    denseChip: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportAddingNewValue: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    canSearch: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    optionsDense: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    policyRules: {
      type: Object as PropType<IGenericSelectPolicyRules | null>,
      required: false,
    },
    disabledInfo: {
      type: String as PropType<string>,
      required: false,
    },
    customClass: {
      type: String as PropType<string>,
      required: false,
    },
    serverSideFilter: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      filterOptions: this.options as ISelectOption[],
      currentFilterValue: "" as string,
    };
  },
  computed: {
    allowedOptions(): ISelectOption[] {
      const policyOptions = this.policyRules?.options;
      if (!policyOptions || !Array.isArray(policyOptions)) return this.options;

      const values = policyOptions.map((c) => c.value);
      return this.options.filter((c) => values.includes(c.value as string));
    },
    disabledTooltip(): string {
      return this.disabledInfo ? this.disabledInfo : errorMessages.CANT_BE_MODIFIED_POLICY;
    },
    disabledState(): boolean {
      if (this.disabledInfo) return true;
      return this.policyRules?.canEdit === false;
    },
    allowedOptionsMap(): Record<string, ISelectOption> {
      return this.allowedOptions.reduce((acc: Record<string, ISelectOption>, option: ISelectOption) => {
        acc[option.value as string] = option;
        return acc;
      }, {});
    },
  },
  methods: {
    multiSelectValidation(options: ISelectOption[]): boolean | string {
      return !this.required || options.length > 0 || this.errorMessage;
    },
    clearActiveOption(): void {
      const selectRef = this.$refs.multiSelect as { setOptionIndex: (index: number) => void };
      selectRef.setOptionIndex(-1);
    },
    resetInput() {
      const selectRef = this.$refs.multiSelect as {
        updateInputValue: (value: string) => void;
        hidePopup: () => void;
      };
      selectRef.updateInputValue("");
      this.clearActiveOption();
      selectRef.hidePopup();
    },
    onClear(): void {
      this.$emit("options-select", []);
    },
    onSelect(options: ISelectOption[]): void {
      let updatedOptions: ISelectOption[] = [];
      if (options) {
        updatedOptions = options.filter((option: ISelectOption) => {
          if (option.label.toLowerCase().includes(NEW_LABEL_MARKER)) {
            this.createNewValue(option.value as string);
            return false;
          } else {
            return true;
          }
        });
      }
      this.$emit("options-select", updatedOptions);
    },
    createNewValue(val: string) {
      if (!this.supportAddingNewValue) return;
      if (val.length > 0) {
        this.$emit("add-option", val, () => {
          this.resetInput();
        });
      }
    },
    filterFn(val: string, update: (func: () => void, afterFn?: () => void) => void) {
      this.currentFilterValue = val;
      update(
        () => {
          if (this.serverSideFilter) {
            this.filterOptions = [...this.allowedOptions];
            if (val !== "" && this.supportAddingNewValue && this.filterOptions.length === 0) {
              this.filterOptions = this.canSearch
                ? [{ label: `${val} (new ${this.newOptionLabel.toLowerCase()})`, value: val }]
                : [{ label: `${val}`, value: val }];
            }
            return;
          }
          if (val === "") {
            this.filterOptions = [...this.allowedOptions];
          } else {
            const needle = val.toLowerCase();
            this.filterOptions = this.allowedOptions.filter(
              (option: ISelectOption) => option.label.toLowerCase().indexOf(needle) > -1,
            );
            if (this.supportAddingNewValue && this.filterOptions.length === 0) {
              if (this.canSearch) {
                this.filterOptions = [{ label: `${val} (new ${this.newOptionLabel.toLowerCase()})`, value: val }];
              } else {
                this.filterOptions = [{ label: `${val}`, value: val }];
              }
            }
          }
        },
        () => this.clearActiveOption(),
      );
    },
    selectedOptionTextColor(selectedOption: ISelectOption): string {
      if (!this.policyRules?.options) return "";
      if (this.allowedOptionsMap[selectedOption.value as string]) return "";
      return "white";
    },
    selectedOptionBackgroundColor(selectedOption: ISelectOption): string {
      if (!this.policyRules?.options) return "";
      if (this.allowedOptionsMap[selectedOption.value as string]) return "";
      return "negative";
    },
    requiredByPolicyValidation(val: ISelectOption[]): boolean | string {
      if (!this.policyRules?.required) return true;
      return !!val.length || errorMessages.REQUIRED_FIELD;
    },
    allowedByPolicyValidation(val: ISelectOption[]): boolean | string {
      if (!this.policyRules?.options || !val?.length) return true;
      return (
        val.every((option: ISelectOption) =>
          this.allowedOptions.some((allowedOption: ISelectOption) => allowedOption.value === option.value),
        ) || errorMessages.CANT_BE_SELECTED_POLICY
      );
    },
  },
  watch: {
    options() {
      (this.$refs.multiSelect as QSelect)?.filter(this.currentFilterValue);
    },
  },
});
</script>
<style lang="scss">
.multi-select-wrapper {
  .q-field--auto-height.q-field--labeled .q-field__native {
    max-height: 125px;
    overflow-y: scroll;

    &::-webkit-scrollbar {
      -webkit-appearance: none;
      height: 2px;
      width: 5px;
    }

    &::-webkit-scrollbar-thumb {
      border-radius: 4px;
      background-color: $black-54;
    }
    .q-field--labeled .q-field__native {
      padding-bottom: 0;
    }
  }
}
</style>
<style lang="scss" scoped></style>
