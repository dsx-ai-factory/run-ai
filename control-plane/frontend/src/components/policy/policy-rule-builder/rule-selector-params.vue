<template>
  <div class="rule-selector-params q-mt-md">
    <div v-for="param in params" :key="param.name" class="param-row q-mb-md">
      <div class="text-b-body1 text-b-label field-label">{{ param.name }}</div>
      <enforcement-value-input
        :aid="`selector-param-${param.name}`"
        :model-value="values[param.name] || ''"
        :field-type="param.type"
        :enum-values="param.enumValues || []"
        :rules="paramRules(param)"
        @update:model-value="(value) => emit('update', param.name, value)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import EnforcementValueInput from "./enforcement-value-input.vue";
import type { IMappingParam } from "@/models/gvk-policy.model";
import { isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { gvkPolicyUtil, type TEnforcementValueRule } from "@/utils/gvk-policy.util";

export interface Props {
  params: IMappingParam[];
  values: Record<string, string>;
}

defineProps<Props>();

const emit = defineEmits<{
  (event: "update", name: string, value: string): void;
}>();

function paramRules(param: IMappingParam): TEnforcementValueRule[] {
  return [...(param.required ? [isParamFilled] : []), ...gvkPolicyUtil.mappingTypeValidationRules(param.type)];
}

function isParamFilled(val: string | number | null): boolean | string {
  // Coerce to string or numeric params fail `(n).length` in isNotEmpty.
  const str = val === null || val === undefined ? "" : String(val);
  return isNotEmpty(str) || errorMessages.ENTER_A_PARAMETER;
}
</script>

<style lang="scss" scoped>
.rule-selector-params :deep(.enforcement-value-input) {
  width: 100%;
  max-width: 100%;
}

.field-label {
  margin-bottom: -8px;
}
</style>
