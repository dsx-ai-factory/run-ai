<template>
  <div class="row items-center" :class="field.format === 'boolean' ? 'no-wrap' : ''" v-if="componentToRender">
    <component
      :is="componentToRender"
      :class="classes"
      :model-value="field.value"
      :label="field.label"
      :disable="disable"
      :placeholder="field.placeholder"
      :policy-rules="policyRules"
      :rules="[regexRule]"
      stack-label
      custom-required-message="This field is required"
      enable-label-tooltip
      @update:model-value="updateValue"
    />
    <runai-tooltip
      v-if="field.tooltip"
      class="q-ml-md self-center"
      tooltip-position="right"
      :tooltip-text="field.tooltip"
      icon="ph-fill ph-note-pencil"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { PolicyNumberInput } from "../policy-number-input";
import { PolicyStringField } from "../policy-string-field";
import { RunaiToggle } from "../runai-toggle";
import type { NumberRules, StringRules, BooleanRules } from "@/swagger-models/policy-service-client";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export interface ICustomField {
  key: string;
  label: string;
  value: number | string | boolean | null;
  format: "boolean" | "number" | "text";
  tooltip?: string;
  required?: boolean;
  min?: number;
  max?: number;
  step?: number;
  regex?: string;
  options?: string[];
  placeholder?: string;
  invalidMessage?: string;
}

export default defineComponent({
  name: "runai-custom-field",
  components: {
    PolicyNumberInput,
    PolicyStringField,
    RunaiToggle,
    RunaiTooltipWrapper,
    RunaiSection,
    RunaiTooltip,
  },
  emits: ["update"],
  props: {
    field: {
      type: Object as PropType<ICustomField>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  computed: {
    toggleWithIcon(): boolean {
      return !!this.field.tooltip && this.field.format === "boolean";
    },
    classes() {
      if (this.field.format === "boolean") return "";
      return "col-11";
    },
    componentToRender(): string {
      switch (this.field.format) {
        case "number":
          return PolicyNumberInput.name || "runai-number-input";
        case "text":
          return PolicyStringField.name || "runai-string-field";
        case "boolean":
          return RunaiToggle.name || "runai-toggle";
        default:
          return "";
      }
    },
    policyRules() {
      switch (this.field.format) {
        case "number":
          return {
            required: this.field.required,
            min: this.field.min,
            max: this.field.max,
            step: this.field.step,
          } as NumberRules;
        case "text":
          return {
            required: this.field.required,
            options: this.field.options,
          } as StringRules;
        case "boolean":
          return {
            required: this.field.required,
          } as BooleanRules;
        default:
          return null;
      }
    },
  },
  methods: {
    regexRule(val: string): boolean | string {
      if (!this.field.regex || (!this.field.required && !val)) return true;

      const regex = new RegExp(this.field.regex);
      return regex.test(val) || this.field.invalidMessage || "Invalid input";
    },
    updateValue(val: string | number | boolean | null): void {
      if (typeof val === "string" && this.field.format === "number") val = parseInt(val);
      this.$emit("update", val);
    },
  },
});
</script>
