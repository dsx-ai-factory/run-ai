<template>
  <section class="serving-endpoint-section">
    <serving-endpoint-access
      :model-value="modelValue"
      :title="title"
      :title-tooltip="titleTooltip"
      :show-protocol-and-port="showProtocolAndPort"
      @update:model-value="onModelChanged"
      :protocol-disabled="protocolDisabled"
      :port-disabled="portDisabled"
      :disabled="disabled"
      :disabled-info="disabledInfo"
      :policy-rules="policyRules"
      :is-port-required="isPortRequired"
      :cluster-id="clusterId || ''"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import ServingEndpointAccess from "./serving-endpoint-access.vue";

// models
import type { IServingEndpointAccessData } from "./serving-endpoint-access.model";
import type { SupersetServingPortRules } from "@/swagger-models/policy-service-client";

export enum EServingEndpointSectionEvents {
  UpdateModelValue = "update:model-value",
}

export default defineComponent({
  name: "serving-endpoint-section",
  components: {
    ServingEndpointAccess,
  },
  emits: {
    [EServingEndpointSectionEvents.UpdateModelValue]: (modelValue: IServingEndpointAccessData) => modelValue,
  },
  props: {
    modelValue: {
      type: Object as PropType<IServingEndpointAccessData>,
      required: true,
    },
    title: {
      type: String as PropType<string>,
      required: true,
    },
    titleTooltip: {
      type: String as PropType<string>,
      required: false,
    },
    showProtocolAndPort: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    protocolDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    portDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disabledInfo: {
      type: String as PropType<string>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<SupersetServingPortRules | null>,
      required: false,
    },
    isPortRequired: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
  },
  methods: {
    onModelChanged(modelValue: IServingEndpointAccessData): void {
      this.$emit(EServingEndpointSectionEvents.UpdateModelValue, modelValue);
    },
  },
});
</script>
<style lang="scss" scoped></style>
