<template>
  <quota-input
    v-bind="$attrs"
    :model-value="modelValue"
    :unit="unit"
    :unit-options="unitOptions"
    :state="state"
    :message="message"
    label="CPU memory"
    placeholder="Unlimited"
    min="0"
    :rules="rules"
    treat-minus-one-as-unlimited
    @update:model-value="$emit('update:model-value', $event)"
    @update:unit="$emit('update:unit', $event)"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { QuotaInput } from "@/components/org-unit/v2/common/quota-input";
import type { UnitOption } from "@/components/org-unit/v2/common/quota-input/quota-input.vue";
import { EQuotaInputState, ECpuMemoryUnit } from "@/models/org-unit.model";
import { Units } from "@/swagger-models/org-unit-service-client";

export interface Props {
  modelValue: number | null;
  unit?: string;
  state?: EQuotaInputState;
  message?: string;
}

const props = withDefaults(defineProps<Props>(), {
  unit: ECpuMemoryUnit.MB,
  state: EQuotaInputState.Default,
  message: "",
});

defineEmits<{
  (e: "update:model-value", value: number | null): void;
  (e: "update:unit", value: Units): void;
}>();

const unitOptions: UnitOption[] = [
  { label: ECpuMemoryUnit.MB, value: Units.Mb },
  { label: ECpuMemoryUnit.GB, value: Units.Gb },
  { label: ECpuMemoryUnit.MiB, value: Units.Mib },
];

const rules = computed(() => {
  const rule = () => {
    if (props.state === EQuotaInputState.Error) {
      return ""; // Return an empty string to indicate an error
    }
    return true;
  };
  return [rule];
});
</script>
