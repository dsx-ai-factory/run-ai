<template>
  <runai-expansion-item :loading="loading" label="Node pool label" default-opened :section-invalid="sectionInvalid">
    <template #subheader>
      {{ expansionSubHeader }}
    </template>
    <runai-section title="Limit resources to specific node pools">
      <div class="row q-mb-lg q-mt-md">
        <div class="col-7">
          <q-input
            :model-value="labelKey"
            @update:model-value="$emit('update:label-key', $event)"
            label="Key"
            stack-label
            :placeholder="getLabelKeyPlaceholder"
            :rules="[keyNotEmpty, isValidLabelKeyFormat]"
            reactive-rules
            no-error-icon
            input-class="placeholder-italic"
            :disable="disabled"
            aid="node-pool-label-key-input"
          >
            <template v-slot:error>
              <div v-dompurify-html="labelKeyValidationError()" />
            </template>
          </q-input>
          <q-tooltip v-if="disabled" anchor="center right">The node pool label key cannot be modified</q-tooltip>
        </div>
      </div>
      <div class="row q-mb-lg">
        <div class="col-7">
          <q-input
            :model-value="labelValue"
            @update:model-value="$emit('update:label-value', $event)"
            label="Value"
            stack-label
            aid="node-pool-label-value-input"
            :placeholder="getLabelValuePlaceholder"
            :rules="[valueNotEmpty, uniqueLabelRule, isValidLabelValueFormat]"
            reactive-rules
            no-error-icon
            input-class="placeholder-italic"
            :disable="disabled"
          >
            <template v-slot:error>
              <div v-dompurify-html="labelValueValidationError()" />
            </template>
          </q-input>
          <q-tooltip v-if="disabled" anchor="center right">The node pool label value cannot be modified</q-tooltip>
        </div>
      </div>
    </runai-section>
  </runai-expansion-item>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { isNotEmpty, isValidRegExp } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import { RunaiSection } from "@/components/common/runai-section";
const nodePoolValidationRegexp = {
  labelKey: "^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?/?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$",
  labelValue: "^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$",
};
export default defineComponent({
  name: "node-pool-label-section",
  components: { RunaiSection, RunaiExpansionItem },
  props: {
    labelKey: {
      type: [String, null] as PropType<string | null>,
      required: true,
    },
    labelValue: {
      type: [String, null] as PropType<string | null>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    nodePools: {
      type: Array as PropType<Nodepool[]>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    expansionSubHeader(): string {
      if (!this.labelKey || !this.labelValue) return "None";
      return `${this.labelKey}: ${this.labelValue}`;
    },
    getLabelKeyPlaceholder(): string {
      if (this.disabled) {
        return "";
      }
      return "e.g., nvidia.com/gpu.product";
    },
    getLabelValuePlaceholder(): string {
      if (this.disabled) {
        return "";
      }
      return "e.g., T4";
    },
    sectionInvalid(): boolean {
      if (this.labelValue === null || this.labelKey === null) return true;
      return (
        !isNotEmpty(this.labelValue) ||
        !isNotEmpty(this.labelKey) ||
        !this.isValidLabelValueFormat(this.labelKey) ||
        !this.isValidLabelValueFormat(this.labelValue) ||
        !this.isUniqueLabel()
      );
    },
  },
  methods: {
    keyNotEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.KEY_NOT_EMPTY;
    },
    isValidLabelKeyFormat(val: string): boolean {
      return isValidRegExp(val, nodePoolValidationRegexp.labelKey);
    },
    labelKeyValidationError(): string {
      return errorMessages.KEY_REGEXP_VALIDATION + "<br/>" + nodePoolValidationRegexp.labelKey;
    },
    valueNotEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.VALUE_NOT_EMPTY;
    },
    uniqueLabelRule(): boolean | string {
      return this.isUniqueLabel() || errorMessages.UNIQUE_LABELS;
    },
    isUniqueLabel(): boolean {
      return !this.nodePools.some(
        (pool: Nodepool) => pool.labelKey === this.labelKey && pool.labelValue === this.labelValue,
      );
    },
    isValidLabelValueFormat(val: string): boolean {
      return isValidRegExp(val, nodePoolValidationRegexp.labelValue);
    },
    labelValueValidationError(): string {
      return errorMessages.VALUE_REGEXP_VALIDATION + "<br />" + nodePoolValidationRegexp.labelValue;
    },
  },
});
</script>
