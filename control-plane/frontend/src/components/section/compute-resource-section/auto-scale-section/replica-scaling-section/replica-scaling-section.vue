<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="replica-scaling-section">
    <div class="q-my-md">
      Set the minimum and maximum number of replicas for your inference
      <runai-tooltip :tooltip-text="minMaxTooltip" tooltip-position="right" width="450px" />
    </div>
    <div class="min-max-replicas q-my-lg">
      <div class="row">
        <policy-number-input
          label="Minimum"
          class="replicas-input"
          type="number"
          :model-value="modelValue.minReplicas !== undefined ? modelValue.minReplicas : null"
          @update:model-value="updateMin"
          stack-label
          :min-value="minReplicasAllowed"
          :disable="disabled"
          aid="min-rep-input"
          :policy-rules="policyRules?.minReplicas"
          :rules="[isMinReplicasAllowedZero, isMinReplicasValid]"
          is-integer
        />

        <span class="col-1 row justify-center items-center q-pt-sm">
          <hr style="width: 10px; display: block" />
        </span>

        <policy-number-input
          class="replicas-input"
          type="number"
          label="Maximum"
          :model-value="modelValue.maxReplicas !== undefined ? modelValue.maxReplicas : null"
          @update:model-value="updateMax"
          stack-label
          :min-value="minValueForMaxReplicas"
          :rules="[isMaxMorThanMin, isMaxMoreThanMaxReplicas]"
          :disable="disabled"
          aid="max-rep-input"
          :policy-rules="policyRules?.maxReplicas"
          is-integer
        />
      </div>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import { errorMessages } from "@/common/error-message.constant";
import { isMaxEqualOrHigherThenMin, isValueLowerOrEqualToMax } from "@/common/form.validators";

// models
import { type AutoScalingRules } from "@/swagger-models/assets-service-client";
import {
  type IReplicaScalingSection,
  MAX_REPLICAS_ALLOWED,
  EReplicaScalingSectionEvents,
} from "./replica-scaling-section.model";

export default defineComponent({
  name: "replica-scaling-section",
  components: {
    RunaiTooltip,
    PolicyNumberInput,
  },
  emits: {
    [EReplicaScalingSectionEvents.UpdateModelValue]: (_data: IReplicaScalingSection) => true,
  },

  props: {
    modelValue: {
      type: Object as PropType<IReplicaScalingSection>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<AutoScalingRules | null>,
      required: false,
    },
    isAutoScaleEnabled: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      minMaxTooltip:
        "Set the minimum and maximum number of replicas to be scaled up and down. If there is a difference between the two, you'll need to set conditions for auto-scaling." as string,
    };
  },
  computed: {
    minReplicasAllowed(): number {
      return this.isAutoScaleEnabled ? 0 : 1;
    },
    minValueForMaxReplicas(): number {
      if (this.modelValue.minReplicas === undefined || this.modelValue.minReplicas === 0) return 1;
      return this.modelValue.minReplicas;
    },
  },
  methods: {
    isMinReplicasValid(val: number | undefined): boolean | string {
      return (
        (typeof val === "number" && val >= this.minReplicasAllowed) ||
        errorMessages.ENTER_A_VALUE_ABOVE_MIN.replace("${min}", `${this.minReplicasAllowed}`)
      );
    },
    isMaxMorThanMin(): boolean | string {
      if (this.modelValue.minReplicas === undefined || this.modelValue.maxReplicas === undefined) return true;
      return (
        isMaxEqualOrHigherThenMin(this.modelValue.maxReplicas, this.modelValue.minReplicas) ||
        errorMessages.EQUAL_TO_OR_HIGHER_THAN_THE_MINIMUM
      );
    },
    isMaxMoreThanMaxReplicas(): boolean | string {
      if (this.modelValue.maxReplicas === undefined) return true;
      return (
        isValueLowerOrEqualToMax(this.modelValue.maxReplicas, MAX_REPLICAS_ALLOWED) ||
        `${errorMessages.ENTER_A_NUMBER_EQUAL_TO_OR_LOWER_THAN} 10,000`
      );
    },
    updateMin(num: string): void {
      const minReplicas = num === "" ? undefined : Number(num);

      let maxReplicas = this.modelValue.maxReplicas;
      if (
        minReplicas !== undefined &&
        minReplicas > 0 &&
        (this.modelValue.maxReplicas === undefined || minReplicas > this.modelValue.maxReplicas)
      ) {
        maxReplicas = minReplicas;
      }

      this.updateModel({
        minReplicas,
        maxReplicas,
      });
    },
    updateMax(num: string): void {
      const maxReplicas = num === "" ? undefined : Number(num);

      this.updateModel({
        ...this.modelValue,
        maxReplicas,
      });
    },
    updateModel(modelValue: IReplicaScalingSection): void {
      this.$emit(EReplicaScalingSectionEvents.UpdateModelValue, modelValue);
    },
    isMinReplicasAllowedZero(val: number | undefined): boolean | string {
      if (this.isAutoScaleEnabled) return true;
      if (val === null) return true;
      return Number(val) > 0 || errorMessages.MIN_REPLICAS_ALLOWED_ZERO;
    },
  },
});
</script>
<style lang="scss" scoped>
.replica-scaling-section {
  .replicas-input {
    width: 160px;
  }
}
</style>
