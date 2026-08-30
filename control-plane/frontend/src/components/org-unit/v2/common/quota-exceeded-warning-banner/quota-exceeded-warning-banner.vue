<template>
  <q-banner class="quota-exceeded-warning-banner bg-orange-1 text-b-body1">
    <div class="row no-wrap items-center">
      <q-icon class="ph-bold ph-warning-circle" color="warning" size="20px" />
      <div class="text-b-body1 q-ml-md">
        {{ baseMessage }}<span v-if="showExcessAmount"> by {{ excessAmount }} GPUs</span>.
      </div>
    </div>
  </q-banner>
</template>

<script setup lang="ts">
import { computed } from "vue";

export interface Props {
  excessAmount: number;
  parentEntityType?: string; // "department", "node pool", "cluster", etc.
  hideExcessAmount?: boolean; // When true, don't show "by X GPUs" suffix
}

const props = withDefaults(defineProps<Props>(), {
  parentEntityType: "department",
  hideExcessAmount: false,
});

// Whether to show the "by X GPUs" suffix
const showExcessAmount = computed(() => !props.hideExcessAmount);

// The base message without the excess amount suffix
const baseMessage = computed(() => {
  const quotaText = props.parentEntityType === "node pool" ? "available GPU's" : "quota";
  return `The total assigned quota exceeds the ${props.parentEntityType}'s ${quotaText}`;
});
</script>

<style scoped lang="scss"></style>
