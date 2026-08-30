<template>
  <div>
    <runai-radio-options
      class="q-mb-md"
      title="Set whether Multi-Node NVLink (MNNVL) acceleration is required for this workload"
      :title-tooltip="tooltipText"
      :options="mnnvlOptions"
      :model-value="modelValue"
      :disable="disabled"
      @update:model-value="updateMnnvl"
    />
    <div class="q-mt-md text-italic">
      For more information, see the
      <a :href="DOCS.TOPOLOGIES.GB200" target="_blank" aid="mnnvl-docs-link"
        >Using GB200 NVL72 and Multi-Node NVLink Domains</a
      >
      guide
    </div>
  </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
// Components
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
// Models
import type { ISelectOption } from "@/models/global.model";
import { MnnvlMode } from "@/swagger-models/workloads-manager-client";
// Constants
import DOCS from "@/constants/docs";

export const mnnvlOptions: ISelectOption[] = [
  {
    label: "Not required",
    value: MnnvlMode.None,
    toolTip: "The workload will not use MNNVL acceleration, even if MNNVL-capable nodes are available.",
  },
  {
    label: "Required",
    value: MnnvlMode.Required,
    toolTip:
      "The workload will require MNNVL-capable nodes and will use MNNVL acceleration. Scheduling may be restricted to compatible nodes, and if none are available, the workload may remain pending.",
  },
];

export default defineComponent({
  name: "mnnvl-section",
  components: {
    RunaiRadioOptions,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: [String, null] as PropType<string | null>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      DOCS,
      mnnvlOptions,
      tooltipText: `MNNVL provides high-bandwidth, low-latency communication between GPUs on supported nodes, improving performance for multi-GPU workloads.`,
    };
  },
  methods: {
    updateMnnvl(value: string): void {
      this.$emit("update:model-value", value);
    },
  },
});
</script>
