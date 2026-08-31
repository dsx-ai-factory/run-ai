<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-folding-sub-section label="Worker elasticity" aid="worker-elasticity-sub-section">
    <div class="q-my-md">
      Set the minimum and maximum workers for your workload
      <runai-tooltip :tooltip-text="minMaxTooltip" tooltip-position="right" width="450px" />
    </div>
    <div class="min-max-workers q-my-lg">
      <div class="row">
        <policy-number-input
          label="Minimum"
          class="col-4"
          type="number"
          :model-value="minWorkersValue"
          @update:model-value="updateMin"
          stack-label
          :min-value="1"
          :disable="sectionDisabled"
          aid="min-workers-input"
          :policy-rules="policyRulesMin"
          :rules="minWorkersRules"
          is-integer
          reactive-rules
        />

        <span class="col-1 row justify-center items-center q-pt-sm">
          <hr style="width: 10px; display: block" />
        </span>

        <policy-number-input
          class="col-4"
          type="number"
          label="Maximum"
          :model-value="maxWorkersValue"
          @update:model-value="updateMax"
          stack-label
          :min-value="minValueForMaxWorkers"
          :max-value="10000"
          :rules="maxWorkersRules"
          :disable="sectionDisabled"
          aid="max-workers-input"
          :policy-rules="policyRulesMax"
          is-integer
          reactive-rules
        />
      </div>
    </div>
  </runai-folding-sub-section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiFoldingSubSection } from "@/components/common/runai-folding-sub-section";
import { PolicyNumberInput } from "@/components/common/policy-number-input";

// models
import type { IntegerRules } from "@/swagger-models/assets-service-client";

export default defineComponent({
  name: "worker-elasticity-section",
  components: {
    RunaiTooltip,
    RunaiFoldingSubSection,
    PolicyNumberInput,
  },
  emits: {
    "update:min-workers": (_num: number | null) => true,
    "update:max-workers": (_num: number | null) => true,
  },

  props: {
    minWorkers: {
      type: [Number, null, undefined] as PropType<number | null | undefined>,
      required: false,
    },
    maxWorkers: {
      type: [Number, null, undefined] as PropType<number | null | undefined>,
      required: false,
    },
    numWorkers: {
      type: [Number, undefined] as PropType<number | undefined>,
      required: false,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRulesMin: {
      type: [Object, null] as PropType<IntegerRules | null>,
      required: false,
    },
    policyRulesMax: {
      type: [Object, null] as PropType<IntegerRules | null>,
      required: false,
    },
  },
  data() {
    return {
      minMaxTooltip:
        "The workload can use more or fewer workers within this range based on available resources. If they are below the minimum, the workload will stop or remain pending." as string,
    };
  },
  computed: {
    minWorkersValue(): number | null {
      return this.minWorkers !== undefined ? this.minWorkers : null;
    },
    maxWorkersValue(): number | null {
      return this.maxWorkers !== undefined ? this.maxWorkers : null;
    },
    numWorkersValue(): number {
      return this.numWorkers || 1;
    },
    minValueForMaxWorkers(): number {
      if (this.minWorkersValue === null) return 1;
      return this.minWorkersValue;
    },
    minWorkersRules(): Array<(val: number | undefined | null) => boolean | string> {
      return [
        (val: number | undefined | null): boolean | string => {
          if (val === null || val === undefined) return true;

          if (val > this.numWorkersValue || val <= 0) {
            return `Enter a value over 0, 
up to the number of workers`;
          }

          return true;
        },
      ];
    },
    maxWorkersRules(): Array<(val: number | undefined | null) => boolean | string> {
      return [
        (val: number | undefined | null): boolean | string => {
          if (val === null || val === undefined) return true;

          // Max must be up to 10,000 and equal to or above the number of workers
          if (val > 10000 || val < this.numWorkersValue) {
            return "Enter a value equal to or above the number of workers, up to 10,000";
          }

          return true;
        },
      ];
    },
  },
  methods: {
    updateMin(num: string): void {
      const minWorkers = num === "" ? null : Number(num);

      let maxWorkers = this.maxWorkersValue;
      if (
        minWorkers !== null &&
        minWorkers > 0 &&
        (this.maxWorkersValue === null || minWorkers > this.maxWorkersValue)
      ) {
        maxWorkers = minWorkers;
        this.$emit("update:max-workers", maxWorkers);
      }

      this.$emit("update:min-workers", minWorkers);
    },
    updateMax(num: string): void {
      const maxWorkers = num === "" ? null : Number(num);
      this.$emit("update:max-workers", maxWorkers);
    },
  },
});
</script>
