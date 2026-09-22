<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="general-section" label="General" aid="general-section">
    <template #subheader>
      <span>{{ summary }}</span>
    </template>

    <preemptibility-section
      v-if="features.supportPreemptibility"
      :model-value="model.preemptibility"
      :policy-rules="config.policyRules?.preemptibility || null"
      :disabled="false"
      :support-semi-preemptibility="config.workloadPriorityData?.supportSemiPreemptibility"
      @update:model-value="onPreemptibilityChanged"
      :show-quota-warning="config.showQuotaWarning"
    />

    <priority-section
      v-if="features.supportPriorityClass"
      :model-value="model.priorityClass"
      :policy-rules="config.policyRules?.priorityClass"
      :disabled="false"
      @update:model-value="onPriorityChanged"
      :workload-priority-data="config.workloadPriorityData"
      :show-quota-warning="config.showQuotaWarning"
      :support-preemptibility="features.supportPreemptibility"
    />

    <workers-startup-policy-section
      :model-value="model.startupPolicy"
      :disable="false"
      :policy-rules="config.policyRules?.startupPolicy"
      @update:model-value="onStartupPolicyChanged"
    />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { PreemptibilitySection } from "@/components/section/general-section/preemptibility-section";
import { PrioritySection } from "@/components/section/general-section/priority-section";
import { WorkersStartupPolicySection } from "@/components/section/general-section/workers-startup-policy-section";

// models
import type { Preemptibility, DistributedInferenceStartupPolicy } from "@/swagger-models/workloads-client";
import type {
  IGeneralSectionModel,
  IGeneralSectionFeatures,
  IGeneralSectionConfig,
} from "./workload-setup-sections.types";

export enum EEmitName {
  UpdateModel = "update:model",
}

export default defineComponent({
  name: "workload-setup-general-section",
  components: {
    RunaiExpansionItem,
    PreemptibilitySection,
    PrioritySection,
    WorkersStartupPolicySection,
  },
  emits: {
    [EEmitName.UpdateModel]: (model: IGeneralSectionModel) => !!model,
  },
  props: {
    model: {
      type: Object as PropType<IGeneralSectionModel>,
      required: true,
    },
    features: {
      type: Object as PropType<IGeneralSectionFeatures>,
      required: true,
    },
    config: {
      type: Object as PropType<IGeneralSectionConfig>,
      default: () => ({ showQuotaWarning: false }),
    },
  },
  computed: {
    summary(): string {
      const parts: string[] = [];

      if (this.features.supportPriorityClass) {
        parts.push(`Priority: ${this.model.priorityClass || "None"}`);
      }

      if (this.features.supportPreemptibility) {
        parts.push(`Preemptibility: ${this.model.preemptibility || "None"}`);
      }

      return parts.length > 0 ? parts.join(" / ") : "";
    },
  },
  methods: {
    onPriorityChanged(priorityClass: string | null): void {
      this.$emit(EEmitName.UpdateModel, { ...this.model, priorityClass });
    },
    onPreemptibilityChanged(preemptibility: Preemptibility | null): void {
      this.$emit(EEmitName.UpdateModel, { ...this.model, preemptibility });
    },
    onStartupPolicyChanged(startupPolicy: DistributedInferenceStartupPolicy | null): void {
      this.$emit(EEmitName.UpdateModel, { ...this.model, startupPolicy });
    },
  },
});
</script>
