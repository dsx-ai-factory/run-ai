<template>
  <runai-section title="Set the number of workers per segment" :tooltip-text="segmentSizeTooltip" tooltip-width="300px">
    <div class="row">
      <policy-number-input
        class="col-3"
        type="number"
        :model-value="props.segmentSize ?? null"
        @update:model-value="updateSegmentSize"
        stack-label
        :min-value="1"
        :disable="props.sectionDisabled"
        custom-ref="segmentSizeInput"
        aid="segment-size-input"
        is-integer
      />
    </div>
  </runai-section>
</template>

<script setup lang="ts">
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { RunaiSection } from "@/components/common/runai-section";

interface Props {
  segmentSize?: number | null;
  sectionDisabled: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{ (event: "update", segmentSize: number | null): void }>();

const segmentSizeTooltip =
  "Workers in the same segment will be placed on nodes that are as close to each other as possible, reducing latency between them. Set this to match your application's parallel segment size. Leave it blank to disable segmentation. The total workers should be divisible by this value.";

function updateSegmentSize(value: string | number | null): void {
  const segmentSize = value === "" || value === null ? null : Number(value);
  emit("update", segmentSize);
}
</script>
