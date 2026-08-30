<template>
  <section class="runs-and-parallelism-section">
    <div class="row items-center">
      <span>Set the number of runs the workload must finish to be considered complete</span>
      <runai-tooltip class="q-ml-md" :tooltip-text="tooltipText" />
    </div>
    <div class="row">
      <policy-number-input
        class="col-3 q-mr-lg"
        :min-value="1"
        :model-value="modelValue.completions"
        @update:model-value="updateRuns"
        label="Runs"
        stack-label
        :disable="disabled"
        :policy-rules="policyRules?.completions"
      />
      <policy-number-input
        v-if="showParallelismInput"
        class="col-3"
        :min-value="1"
        :model-value="modelValue.parallelism"
        @update:model-value="updateParallelism"
        label="Parallelism"
        stack-label
        :disable="disabled"
        :policy-rules="policyRules?.parallelism"
        reactive-rules
        :rules="[isValidParallelism]"
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
import { errorMessages } from "@/common/error-message.constant";

export interface ICompletionsAndParallelism {
  completions: number | null | undefined;
  parallelism: number | null | undefined;
}

export interface IRunsAndParallelismRules {
  completions: NumberRules | null;
  parallelism: NumberRules | null;
}

export default defineComponent({
  name: "runs-and-parallelism-section",
  components: {
    RunaiTooltip,
    PolicyNumberInput,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: Object as PropType<ICompletionsAndParallelism>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<IRunsAndParallelismRules>,
      required: false,
    },
  },
  data() {
    return {
      tooltipText: `Multiple runs enhance the reliability and validity of the training results. When the number of runs is above 1, you can set how many may be scheduled in parallel.`,
    };
  },
  computed: {
    showParallelismInput(): boolean {
      return !!this.modelValue.completions && this.modelValue.completions > 1;
    },
  },
  methods: {
    isValidParallelism(val: number): boolean | string {
      return +val <= (this.modelValue.completions || 0) || errorMessages.INVALID_PARALLELISM;
    },
    updateRuns(val: string | number | null): void {
      const completions = Number(val) || null;
      let parallelism = Number(this.modelValue.parallelism) || null;

      if (Number(completions) > 1 && !parallelism) parallelism = 1;
      else if (!completions || completions === 1) parallelism = null;

      this.$emit("update:model-value", { completions, parallelism });
    },
    updateParallelism(val: string | number | null): void {
      const parallelism = Number(val) || null;
      const completions = Number(this.modelValue.completions) || null;
      this.$emit("update:model-value", { completions, parallelism });
    },
  },
});
</script>
