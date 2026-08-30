<template>
  <runai-segmented-button
    class="policy-config-mode-toggle"
    aid="config-mode-segmented-button"
    aria-label="Set how the policy should be configured"
    :model-value="modelValue"
    :options="options"
    @update:model-value="$emit('update:model-value', $event)"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiSegmentedButton, type ISegmentedButtonOption } from "@/components/common/runai-segmented-button";

export interface Props {
  modelValue: string;
  isAnyWorkload: boolean;
}

const props = defineProps<Props>();
defineEmits<{
  (event: "update:model-value", value: string | number | boolean | null): void;
}>();

const options = computed<ISegmentedButtonOption[]>(() => [
  {
    value: "rule-builder",
    label: "Rule builder",
    disable: !props.isAnyWorkload,
    disabledTooltip: props.isAnyWorkload ? undefined : "This can't be modified for the selected workload type",
  },
  {
    value: "yaml",
    label: "YAML editor",
  },
]);
</script>

<style lang="scss" scoped>
.policy-config-mode-toggle :deep(.runai-segment) {
  min-width: 120px;
}
</style>
