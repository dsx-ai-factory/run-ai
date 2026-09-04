<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="runai-duration-inputs row items-center b-gap-m">
    <q-input
      v-if="!hideDays"
      class="duration-field"
      aid="days-input"
      :min="0"
      stack-label
      hide-bottom-space
      no-error-icon
      :error="hasValidationError"
      type="number"
      label="Days"
      v-model="days"
      @update:model-value="updateTotalTime"
      :disable="disable"
      @wheel.prevent
    />
    <q-input
      v-if="!hideHours"
      class="duration-field"
      aid="hours-input"
      :min="0"
      stack-label
      hide-bottom-space
      no-error-icon
      :error="hasValidationError"
      type="number"
      :max="23"
      label="Hours"
      v-model="hours"
      @update:model-value="updateTotalTime"
      :disable="disable"
      @wheel.prevent
    />
    <q-input
      class="duration-field"
      aid="minutes-input"
      :min="0"
      stack-label
      hide-bottom-space
      no-error-icon
      :error="hasValidationError"
      type="number"
      :max="59"
      label="Minutes"
      v-model="minutes"
      @update:model-value="updateTotalTime"
      :disable="disable"
      @wheel.prevent
    />
    <q-input
      v-if="!hideSeconds"
      class="duration-field"
      aid="seconds-input"
      :min="0"
      stack-label
      hide-bottom-space
      no-error-icon
      :error="hasValidationError"
      type="number"
      :max="59"
      label="Seconds"
      v-model="seconds"
      @update:model-value="updateTotalTime"
      :disable="disable"
      @wheel.prevent
    />
  </div>
  <q-field
    class="hidden-field"
    aid="runai-duration-input-error-message"
    :rules="combinedRules"
    :model-value="modelValue"
    no-error-icon
  ></q-field>
</template>

<script lang="ts">
import { type PropType, defineComponent } from "vue";
import type { ValidationRule } from "quasar";
import { errorMessages } from "@/common/error-message.constant";

const DAY_IN_SECONDS = 24 * 60 * 60;
const HOUR_IN_SECONDS = 60 * 60;
const MINUTE_IN_SECONDS = 60;

const nonNegativeRule: ValidationRule = (val: number | null | undefined) => {
  if (val === null || val === undefined) return true;
  return val >= 0 ? true : errorMessages.VALID_TIME_FRAME;
};
export default defineComponent({
  name: "runai-duration-input",
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: [Number, null, undefined] as PropType<number | null | undefined>,
      required: true,
    },
    hideSeconds: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideDays: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideHours: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    rules: {
      type: Array as PropType<ValidationRule[]>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      days: null as string | number | null,
      hours: null as string | number | null,
      minutes: null as string | number | null,
      seconds: null as string | number | null,
      combinedRules: [nonNegativeRule] as ValidationRule[],
    };
  },
  created() {
    this.initModelValue();
    this.updateCombinedRules();
  },
  computed: {
    hasValidationError(): boolean {
      return this.combinedRules.some((rule) => {
        if (typeof rule !== "function") return false;
        try {
          const result = (rule as (val: unknown) => boolean | string | Promise<boolean | string>)(this.modelValue);
          if (result instanceof Promise) return false;
          return typeof result === "string" || result === false;
        } catch {
          return true;
        }
      });
    },
  },
  methods: {
    updateCombinedRules(): void {
      this.combinedRules = [...(this.rules ?? []), nonNegativeRule];
    },
    initModelValue(): void {
      if (this.modelValue === null || this.modelValue === undefined) return;
      // Initializing the inputs values
      let remainingSeconds = Math.floor(this.modelValue);
      if (!this.hideDays) {
        this.days = Math.floor(remainingSeconds / DAY_IN_SECONDS) || null;
        remainingSeconds %= DAY_IN_SECONDS;
      }
      if (!this.hideHours) {
        this.hours = Math.floor(remainingSeconds / HOUR_IN_SECONDS) || null;
        remainingSeconds %= HOUR_IN_SECONDS;
      }
      this.minutes = Math.floor(remainingSeconds / MINUTE_IN_SECONDS) || null;
      if (!this.hideSeconds) {
        this.seconds = remainingSeconds % MINUTE_IN_SECONDS || null;
        if (this.modelValue === 0) this.seconds = 0;
      }
    },
    updateTotalTime(): void {
      if ([this.days, this.hours, this.minutes, this.seconds].every((value) => value === null || value === "")) {
        this.$emit("update:model-value", null);
        return;
      }

      let totalTime = 0;
      totalTime += Number(this.seconds) || 0;
      totalTime += Number(this.minutes) * MINUTE_IN_SECONDS || 0;
      totalTime += Number(this.hours) * HOUR_IN_SECONDS || 0;
      totalTime += Number(this.days) * DAY_IN_SECONDS || 0;
      this.$emit("update:model-value", totalTime);
    },
  },
  watch: {
    modelValue() {
      this.initModelValue();
    },
    rules() {
      this.updateCombinedRules();
    },
  },
});
</script>

<style lang="scss" scoped>
.duration-field {
  width: 120px;
}

.runai-duration-inputs {
  :deep(.q-field__label) {
    color: var(--b-text-label);
  }

  :deep(.q-field__native) {
    color: var(--b-text-main);
  }

  :deep(.q-field__control::before) {
    border-color: var(--border-default);
  }

  :deep(.q-field--disabled) {
    .q-field__label,
    .q-field__native {
      color: var(--text-disabled);
    }
  }
}

.hidden-field {
  :deep(.q-field__bottom) {
    color: var(--text-error);
  }
}
</style>
