<template>
  <section class="gpu-devices-box row">
    <compute-resource-box :texts="texts" title="GPU devices per pod">
      <template #inputs>
        <div class="spacer"></div>
        <policy-number-input
          outlined
          :min-value="0"
          label="Request"
          :model-value="gpuDevicesRequest"
          @update:model-value="updateGpuDevices"
          class="input-field"
          input-class="gpu-devices-input"
          aid="gpu-devices-input"
          bg-color="white"
          :rules="[() => true]"
          :policy-rules="policyRules"
          :disable="disable"
          is-integer
        />
      </template>
    </compute-resource-box>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { ComputeResourceBox } from "../compute-resource-box";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import type { IntegerRules } from "@/swagger-models/policy-service-client";

export default defineComponent({
  name: "gpu-devices-box",
  components: {
    ComputeResourceBox,
    PolicyNumberInput,
  },
  emits: ["update-gpu-devices-request"],
  props: {
    gpuDevicesRequest: {
      type: Number as PropType<number>,
      required: true,
    },
    texts: {
      type: Array as PropType<string[]>,
      default: () => [],
    },
    policyRules: {
      type: [Object, null] as PropType<IntegerRules | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  methods: {
    updateGpuDevices(value: number | string | null): void {
      if (value === null || +value < 0) value = 0;
      this.$emit("update-gpu-devices-request", +value);
    },
  },
});
</script>

<style lang="scss">
.gpu-devices-box {
  .spacer {
    width: 144px;
  }

  .input-field {
    width: 135px;
  }
}
</style>
