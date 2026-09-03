<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-section
    title="Set which pods should be deleted after workload completion or failure"
    tooltip-text="Keep pods to access their logs for debugging or remove them to free up unnecessary compute resources"
  >
    <div class="row">
      <runai-select
        :disable="sectionDisabled"
        class="col-2"
        :options="cleanPodPolicyOptions"
        :model-value="modelValue || ''"
        :policy-rules="!!policyRules ? { ...policyRules, isClosedList: true } : undefined"
        @update:model-value="$emit('update:model-value', $event?.value)"
      />
    </div>
  </runai-section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
// Cmps
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiSelect } from "@/components/common/runai-select";
// Models
import type { CleanPodPolicyRules } from "@/swagger-models/workloads-client";
import { CleanPodPolicy } from "@/swagger-models/workloads-client";

export default defineComponent({
  name: "clean-pod-policy-section",
  components: {
    RunaiSection,
    RunaiSelect,
  },
  emits: ["update:model-value"],
  props: {
    modelValue: {
      type: [String, null] as PropType<CleanPodPolicy | null>,
      required: true,
    },
    policyRules: {
      type: [Object, null] as PropType<CleanPodPolicyRules | null>,
      required: false,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      cleanPodPolicyOptions: [
        {
          label: "None",
          toolTip:
            "No pods will be deleted upon the completion or failure of the workload, and compute resources will continue to be consumed",
          value: CleanPodPolicy.None,
        },
        {
          label: "All",
          toolTip: "All pods will be deleted upon the completion or failure of the workload",
          value: CleanPodPolicy.All,
        },
        {
          label: "Running",
          toolTip:
            "Only running pods will be deleted upon the completion or failure of the workload. Completed or failed pods and their logs will remain available.",
          value: CleanPodPolicy.Running,
        },
      ],
    };
  },
});
</script>
