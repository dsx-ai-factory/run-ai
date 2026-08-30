<template>
  <runai-section class="launcher-policy-section" title="Set when the launcher pod should be created">
    <div class="row">
      <runai-select
        aid="launcher-policy-select"
        class="col-4"
        placeholder="Select..."
        emit-value
        :options="options"
        :model-value="modelValue"
        :disable="disable"
        :policy-rules="launcherRules"
        @update:model-value="updateLauncherPolicy"
      />
    </div>
  </runai-section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiSelect } from "@/components/common/runai-select";
// Models
import type { ISelectOption } from "@/models/global.model";
import { MpiLauncherCreationPolicy, type MpiLauncherCreationPolicyRules } from "@/swagger-models/workloads-client";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";

export default defineComponent({
  name: "launcher-policy-section",
  components: {
    RunaiSection,
    RunaiSelect,
  },
  emits: {
    "update:model-value": (_: MpiLauncherCreationPolicy | null) => true,
  },
  props: {
    modelValue: {
      type: [String, null] as PropType<MpiLauncherCreationPolicy | null>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<MpiLauncherCreationPolicyRules | null>,
      required: false,
    },
  },
  computed: {
    options(): ISelectOption[] {
      return [
        {
          label: "When workload is created",
          value: MpiLauncherCreationPolicy.AtStartup,
          toolTip:
            "Create the launcher pod alongside the worker pods to reduce startup time. This may cause the launcher to fail if the workers aren't ready.",
        },
        {
          label: "Once all workers are ready",
          value: MpiLauncherCreationPolicy.WaitForWorkersReady,
          toolTip: "Wait for all worker pods to be ready before creating the launcher.",
        },
      ];
    },
    launcherRules(): IGenericSelectPolicyRules | undefined {
      return this.policyRules ? { ...this.policyRules, isClosedList: true } : undefined;
    },
  },
  methods: {
    updateLauncherPolicy(value: MpiLauncherCreationPolicy | null) {
      this.$emit("update:model-value", value);
    },
  },
});
</script>
