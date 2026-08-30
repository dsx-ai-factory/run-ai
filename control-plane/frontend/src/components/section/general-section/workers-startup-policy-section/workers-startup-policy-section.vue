<template>
  <runai-section class="workers-startup-policy-section" title="Set when the workers' pod should be created">
    <div class="row">
      <runai-select
        aid="workers-startup-policy-select"
        class="startup-policy-select col-4"
        placeholder="Select..."
        emit-value
        :options="options"
        :model-value="modelValue"
        :disable="isSectionDisabled"
        :policy-rules="startupPolicyRules"
        @update:model-value="updateStartupPolicy"
      />
      <q-tooltip v-if="policyRules?.canEdit === false" target=".startup-policy-select" :offset="[0, -15]">
        This can't be modified based on the policy your administrator set.
      </q-tooltip>
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
import { DistributedInferenceStartupPolicy } from "@/swagger-models/workloads-client";
import type { DistributedInferenceStartupPolicyFieldRulesStartupPolicy } from "@/swagger-models/policy-service-client";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";

export default defineComponent({
  name: "workers-startup-policy-section",
  components: {
    RunaiSection,
    RunaiSelect,
  },
  emits: {
    "update:model-value": (_: DistributedInferenceStartupPolicy | null) => true,
  },
  props: {
    modelValue: {
      type: [String, null] as PropType<DistributedInferenceStartupPolicy | null>,
      required: true,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<DistributedInferenceStartupPolicyFieldRulesStartupPolicy | null>,
      required: false,
    },
  },
  computed: {
    options(): ISelectOption[] {
      return [
        {
          label: "When the leader pod is created",
          value: DistributedInferenceStartupPolicy.LeaderCreated,
          toolTip:
            "Create the workers' pods alongside the leader pod to reduce startup time. This may cause the deployment to fail if the leader isn't ready.",
        },
        {
          label: "Once the leader pod is ready",
          value: DistributedInferenceStartupPolicy.LeaderReady,
          toolTip: "Wait for the leader's pod to be ready before creating the workers' pods.",
        },
      ];
    },
    startupPolicyRules(): IGenericSelectPolicyRules | undefined {
      return this.policyRules ? { ...this.policyRules, isClosedList: true } : undefined;
    },
    isSectionDisabled(): boolean {
      return this.disable || this.policyRules?.canEdit === false;
    },
  },
  methods: {
    updateStartupPolicy(value: DistributedInferenceStartupPolicy | null) {
      this.$emit("update:model-value", value);
    },
  },
});
</script>
