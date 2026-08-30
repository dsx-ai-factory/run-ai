<template>
  <section class="request-timeout-section">
    <duration-input-section
      class="initialization-timeout-container"
      :label="INITIALIZATION_DEFAULTS.label"
      :tooltip-text="INITIALIZATION_DEFAULTS.tooltipText"
      :model-value="modelValue.initializationTimeoutSeconds || null"
      @update:model-value="updateInitializationTimeout"
      :disable="disable"
      :policy-rules="policyRules?.initializationTimeoutSeconds || null"
      :min="INITIALIZATION_DEFAULTS.min"
      :max="INITIALIZATION_DEFAULTS.max"
    />

    <duration-input-section
      class="request-timeout-container q-mt-md"
      :label="REQUEST_TIMEOUT_DEFAULTS.label"
      :tooltip-text="REQUEST_TIMEOUT_DEFAULTS.tooltipText"
      :model-value="modelValue.requestTimeoutSeconds || null"
      @update:model-value="updateRequestTimeout"
      :disable="disable"
      :policy-rules="policyRules?.requestTimeoutSeconds || null"
      :min="REQUEST_TIMEOUT_DEFAULTS.min"
      :max="REQUEST_TIMEOUT_DEFAULTS.max"
    />
  </section>
</template>

<script lang="ts">
import { type PropType, defineComponent } from "vue";

// components
import { DurationInputSection } from "./duration-input";

// utils
import type { ServingConfigurationRules } from "@/swagger-models/policy-service-client";
import type { ServingConfiguration } from "@/swagger-models/assets-service-client";

export enum EEmitRequestTimeoutSection {
  UpdateModelValue = "update:model-value",
}

export const INITIALIZATION_DEFAULTS = {
  label: "Set the workload initialization timeout",
  tooltipText:
    "The maximum time allowed for a workload to initialize and become ready. If the workload does not start within this time, it will be moved to failed state.",
  min: 5,
  max: 43200, // 12 hours
};

export const REQUEST_TIMEOUT_DEFAULTS = {
  label: "Set the request timeout",
  tooltipText:
    "The maximum time allowed to process an end-user request. If no response is returned within this time, the request will be ignored.",
  min: 5,
  max: 600, // 10 minutes
};

export default defineComponent({
  name: "request-timeout-section",
  components: {
    DurationInputSection,
  },
  emits: {
    [EEmitRequestTimeoutSection.UpdateModelValue]: (_: ServingConfiguration) => {
      return _;
    },
  },
  props: {
    modelValue: {
      type: [Object, null] as PropType<ServingConfiguration>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<ServingConfigurationRules | null>,
      required: false,
    },
  },
  data() {
    return {
      INITIALIZATION_DEFAULTS,
      REQUEST_TIMEOUT_DEFAULTS,
    };
  },
  methods: {
    updateInitializationTimeout(time: number | null): void {
      this.$emit(EEmitRequestTimeoutSection.UpdateModelValue, {
        ...this.modelValue,
        initializationTimeoutSeconds: time,
      });
    },
    updateRequestTimeout(time: number | null): void {
      this.$emit(EEmitRequestTimeoutSection.UpdateModelValue, { ...this.modelValue, requestTimeoutSeconds: time });
    },
  },
});
</script>

<style lang="scss" scoped></style>
