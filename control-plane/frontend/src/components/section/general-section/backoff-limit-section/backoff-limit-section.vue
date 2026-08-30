<template>
  <section class="backoff-limit-section">
    <div class="row items-center">
      <span>Set the backoff limit before workload failure</span>
      <runai-tooltip class="q-ml-md" :tooltip-text="tooltipText" />
    </div>
    <div class="row">
      <policy-number-input
        class="col-3"
        :max-value="100"
        :min-value="0"
        :model-value="modelValue"
        @update:model-value="updateBackoffLimit"
        label="Attempts"
        stack-label
        no-error-icon
        :disable="disabled"
        :policy-rules="policyRules"
      />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Cmps
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import type { NumberRules } from "@/swagger-models/policy-service-client";

export default defineComponent({
  name: "backoff-limit-section",
  components: {
    RunaiTooltip,
    PolicyNumberInput,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: [Number, null, undefined] as PropType<number | null | undefined>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<NumberRules | null>,
      required: false,
    },
  },
  data() {
    return {
      tooltipText: `The backoff limit is the maximum number of retry attempts for failed workloads. After reaching the limit, the workload's status will change to "Failed."`,
    };
  },
  methods: {
    updateBackoffLimit(val: string | number | null): void {
      this.$emit("update:model-value", val ? +val : 0);
    },
  },
});
</script>
