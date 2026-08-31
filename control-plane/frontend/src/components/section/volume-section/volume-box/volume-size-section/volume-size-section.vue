<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <template v-if="policyRules?.options?.length">
    <div class="row q-gutter-lg">
      <policy-string-field
        aid="claim-size-select"
        class="col-12"
        :model-value="modelValue"
        :disable="disable"
        :policy-rules="policyRules"
        stack-label
        label="Claim size"
        :rules="[isSelected, ...sizeFieldValidations()]"
        @update:model-value="updateSize"
      />
    </div>
  </template>
  <template v-else>
    <section class="volume-size-section">
      <section class="inputs-wrapper row">
        <q-input
          aid="claim-size-input"
          type="number"
          class="claim-size-input"
          :model-value="sizeNum"
          no-error-icon
          label="Claim size"
          min="0"
          :step="inputStep"
          stack-label
          @update:model-value="handleSizeChange(Number($event))"
          :disable="disable || disabledByPolicy"
          :loading="loading"
          :error="showSizeError"
          hide-bottom-space
        />
        <q-tooltip class="explanatory-tooltip" target=".claim-size-input" v-if="disabledByPolicy" max-width="350px">
          This can't be modified based on the policy your administrator set.
        </q-tooltip>

        <runai-select
          aid="unit-select"
          class="claim-size-unit-select"
          standard
          :model-value="selectedUnit"
          label="Units"
          option-label="label"
          map-options
          emit-value
          :options="unitOptions"
          @update:model-value="handleUnitChange"
          :disable="disable || disabledByPolicy"
          unclearable
          :error="showSizeError"
          :loading="loading"
          hide-bottom-space
        />
        <q-tooltip
          class="explanatory-tooltip"
          target=".claim-size-unit-select"
          v-if="disabledByPolicy"
          max-width="350px"
        >
          This can't be modified based on the policy your administrator set.
        </q-tooltip>
      </section>
      <q-field class="hidden-field q-field--with-bottom" :model-value="modelValue" :rules="sizeFieldValidations()" />
    </section>
  </template>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSelect } from "@/components/common/runai-select";
import { PolicyStringField } from "@/components/common/policy-string-field";
// Constants
import { errorMessages } from "@/common/error-message.constant";
// Utils
import { isNotEmpty, isNumberGreaterThanZero } from "@/common/form.validators";
// Models
import { type ISelectOption, allMemoryUnitOptions } from "@/models/global.model";
import { separateDigitsAndLetters } from "@/utils/common.util";
import type { StringRules } from "@/swagger-models/policy-service-client";
import type {
  ClaimSizeCustomization,
  ClaimSizeCustomizationSupportedUnitsEnum,
} from "@/swagger-models/k8s-objects-tracker-client";
import { resourceUtil } from "@/utils/resource.util";

const DEFAULT_MIN_SIZE = "1M";

export default defineComponent({
  name: "volume-size-section",
  components: { RunaiSelect, PolicyStringField },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<StringRules | null>,
      required: false,
      default: () => ({}),
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    claimSizeCustomization: {
      type: Object as PropType<ClaimSizeCustomization>,
      required: false,
    },
  },
  computed: {
    disabledByPolicy(): boolean {
      return this.policyRules?.canEdit === false;
    },
    sizeNum(): number {
      const [sizeNum] = separateDigitsAndLetters(this.modelValue);
      return +sizeNum || 0;
    },
    selectedUnit(): string {
      const [, selectedUnit] = separateDigitsAndLetters(this.modelValue);
      return selectedUnit || "";
    },
    unitOptions(): Array<ISelectOption> {
      if (!this.claimSizeCustomization?.supportedUnits) return allMemoryUnitOptions;

      return allMemoryUnitOptions.map((option) => {
        if (
          this.claimSizeCustomization?.supportedUnits &&
          this.claimSizeCustomization.supportedUnits.includes(option.label as ClaimSizeCustomizationSupportedUnitsEnum)
        ) {
          return option;
        }
        return { ...option, disable: true, disabledTooltip: "This can't be selected based on the storage class setup" };
      });
    },
    minBytesSize(): number {
      return resourceUtil.parseK8sResourceSizeToBytes(this.claimSizeCustomization?.min || DEFAULT_MIN_SIZE);
    },
    maxBytesSize(): number {
      return resourceUtil.parseK8sResourceSizeToBytes(this.claimSizeCustomization?.max || "");
    },
    stepBytesSize(): number {
      if (!this.claimSizeCustomization?.step) return 0;
      return resourceUtil.parseK8sResourceSizeToBytes(this.claimSizeCustomization.step);
    },
    inputStep(): number | undefined {
      if (!this.stepBytesSize || !this.selectedUnit) return undefined;
      try {
        return resourceUtil.convertBytesToK8sResourceSizeUnit(this.stepBytesSize, this.selectedUnit);
      } catch {
        return undefined;
      }
    },
    stepErrorMessage(): string {
      const step = this.claimSizeCustomization?.step || "";
      const min = this.claimSizeCustomization?.min || DEFAULT_MIN_SIZE;
      const max = this.claimSizeCustomization?.max;

      if (min && max) {
        return errorMessages.ENTER_A_VALUE_BETWEEN_WITH_STEP.replace("${step}", step)
          .replace("${min}", `${min}B`)
          .replace("${max}", `${max}B`);
      }
      if (min) {
        return errorMessages.VALUE_MUST_FOLLOW_STEPS_STARTING_MIN.replace("${step}", step).replace("${min}", `${min}B`);
      }
      return errorMessages.VALUE_MUST_FOLLOW_STEPS_OF.replace("${step}", step);
    },
    showSizeError(): boolean {
      const validationRules: Array<(x: string) => boolean | string> = this.sizeFieldValidations();
      return validationRules.some((rule) => rule(this.modelValue) !== true);
    },
  },
  methods: {
    handleSizeChange(val: number): void {
      this.updateSize(val + this.selectedUnit);
    },
    handleUnitChange(val: string): void {
      this.updateSize(this.sizeNum + val);
    },
    updateSize(size: string): void {
      this.$emit("update:model-value", size);
    },
    // Rules
    sizeNotEmpty(val: number): boolean | string {
      return (val !== undefined && !Number.isNaN(val) && val !== null) || errorMessages.VALUE_NOT_EMPTY;
    },
    largerThanZero(val: number): boolean | string {
      return isNumberGreaterThanZero(val) || errorMessages.NUMBER_GREATER_THAN_ZERO;
    },
    isSelected(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.SELECT_A_VALUE;
    },
    sizeFieldValidations(): Array<(x: string) => string | boolean> {
      if (!this.claimSizeCustomization) return [this.isUnderMin];

      const validations: Array<(x: string) => string | boolean> = [];

      if (this.claimSizeCustomization.max) {
        validations.push(this.isBetweenMinMax);
      } else if (this.claimSizeCustomization.min) {
        validations.push(this.isUnderMin);
      }

      if (this.claimSizeCustomization.step) {
        validations.push(this.isValidStep);
      }

      return validations;
    },
    isUnderMin(val: string): boolean | string {
      let valBytes = 0;
      try {
        valBytes = resourceUtil.parseK8sResourceSizeToBytes(val);
      } catch (err) {
        errorMessages.ENTER_A_VALUE_ABOVE_MIN.replace(
          "${min}",
          `${this.claimSizeCustomization?.min || DEFAULT_MIN_SIZE}B`,
        );
      }

      return (
        valBytes >= this.minBytesSize ||
        errorMessages.ENTER_A_VALUE_ABOVE_MIN.replace(
          "${min}",
          `${this.claimSizeCustomization?.min || DEFAULT_MIN_SIZE}B`,
        )
      );
    },
    isAboveMax(val: string): boolean | string {
      const valBytes = resourceUtil.parseK8sResourceSizeToBytes(val);
      return (
        valBytes <= this.maxBytesSize ||
        errorMessages.ENTER_A_VALUE_BELOW_MAX.replace("${max}", `${this.claimSizeCustomization?.max}B`)
      );
    },
    isBetweenMinMax(val: string): boolean | string {
      const valBytes = resourceUtil.parseK8sResourceSizeToBytes(val);
      return (
        (valBytes >= this.minBytesSize && valBytes <= this.maxBytesSize) ||
        errorMessages.ENTER_A_VALUE_BETWEEN.replace(
          "${min}",
          `${this.claimSizeCustomization?.min || DEFAULT_MIN_SIZE}B`,
        ).replace("${max}", `${this.claimSizeCustomization?.max}B`)
      );
    },
    isValidStep(val: string): boolean | string {
      if (!this.stepBytesSize) return true;

      let valBytes = 0;
      try {
        valBytes = resourceUtil.parseK8sResourceSizeToBytes(val);
      } catch {
        return this.stepErrorMessage;
      }

      const isValid = (valBytes - this.minBytesSize) % this.stepBytesSize === 0;
      return isValid || this.stepErrorMessage;
    },
  },
});
</script>

<style lang="scss" scoped>
.volume-size-section {
  .inputs-wrapper {
    gap: 30px;
  }
  .claim-size-input,
  .claim-size-unit-select {
    flex: 1;
  }
}
</style>
