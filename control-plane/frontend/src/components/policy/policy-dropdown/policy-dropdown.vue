<template>
  <q-menu class="policy-dropdown" anchor="bottom middle" self="top middle">
    <q-list class="column">
      <q-item
        clickable
        v-close-popup
        v-for="opt in optionsForDisplay"
        :key="opt.value"
        @click="$emit('policy-selected', opt.value)"
        :aid="opt.value + '-btn'"
        :disable="opt.disable"
        class="row justify-between items-center"
      >
        <span class="q-mr-sm"
          >{{ opt.label }}
          <q-tooltip v-if="opt.disable">There is no policy applied for the selected {{ entityName }}</q-tooltip></span
        >
        <div>
          <runai-tooltip :tooltip-text="opt.toolTip" tooltip-position="right" />
        </div>
      </q-item>
    </q-list>
  </q-menu>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { EWorkloadType } from "@/models/workload.model";
import type { PropType } from "vue";
import type { ISelectOption } from "@/models/global.model";
import { PolicyType } from "@/swagger-models/policy-service-client";
export interface IPolicySelectOptions extends Omit<ISelectOption, "value" | "disable" | "toolTip"> {
  label: string;
  // A native PolicyType, or an any-workload (GVK) policy id (UUID) for the dynamic GVK entries.
  value: PolicyType | string;
  toolTip: string;
  disable?: boolean;
}

const getToolTip = (policyType: string) => {
  return `This policy defines ${policyType} submission. Its rules derive from both the project-level policy and higher-level policies.`;
};

export default defineComponent({
  name: "policy-dropdown",
  components: {
    RunaiTooltip,
  },
  emits: ["policy-selected"],
  props: {
    existingPolicies: {
      type: Array as PropType<Array<PolicyType>>,
      default: () => [PolicyType.Workspace, PolicyType.Training],
    },
    // Dynamic any-workload (GVK) entries appended after the native options; always enabled.
    additionalOptions: {
      type: Array as PropType<IPolicySelectOptions[]>,
      default: () => [],
    },
    entityName: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      options: [
        {
          label: EWorkloadType.Workspace,
          value: PolicyType.Workspace,
          toolTip: getToolTip("workspace"),
        },
        {
          label: "Training (Standard)",
          value: PolicyType.Training,
          toolTip: getToolTip("standard training"),
        },
        {
          label: "Training (Distributed)",
          value: PolicyType.Distributed,
          toolTip: getToolTip("distributed training"),
        },
        {
          label: EWorkloadType.Inference,
          value: PolicyType.Inference,
          toolTip: getToolTip("inference"),
        },
      ] as IPolicySelectOptions[],
    };
  },
  computed: {
    optionsForDisplay(): IPolicySelectOptions[] {
      const nativeOptions = this.options.map((opt: IPolicySelectOptions) => ({
        ...opt,
        disable: !this.existingPolicies.includes(opt.value as PolicyType),
      }));
      return [...nativeOptions, ...this.additionalOptions];
    },
  },
});
</script>
