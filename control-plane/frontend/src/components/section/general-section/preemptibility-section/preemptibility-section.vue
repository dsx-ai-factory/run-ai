<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-section class="q-mb-lg" title="Set if the workload may be interrupted" :tooltip-text="tooltipText">
    <div class="row items-center q-mt-lg q-gutter-x-md">
      <runai-select
        class="preemptibility-select col-3"
        aid="preemptibility-select"
        :class="{ 'q-pb-none': !preemptibilityHint }"
        :model-value="selectedPreemptibilityOption"
        @update:model-value="updatePreemptibility"
        :options="preemptibilityOptions"
        :policy-rules="selectGenericPolicyRules"
        :disable="isSectionDisabled"
        label="Preemptibility"
        :hint="preemptibilityHint"
      />
      <notification-alert-bar
        v-if="shouldShowQuotaWarning"
        :variant="ENotificationBarVariant.Warning"
        class="flex-1 q-mt-md"
      >
        <span v-dompurify-html="quotaWarningText"></span>
      </notification-alert-bar>
    </div>
    <q-tooltip v-if="policyRules?.canEdit === false" target=".preemptibility-select" :offset="[0, -15]">
      This can't be modified based on the policy your administrator set.
    </q-tooltip>
  </runai-section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
// Components
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiSelect } from "@/components/common/runai-select";
import { SelectType } from "@/components/common/runai-select/runai-select.vue";
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
// Models
import type { PreemptibilityRules, PreemptibilityOptionsOptionsInner } from "@/swagger-models/workloads-client";
import type { ISelectOption } from "@/models/global.model";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";
import { getWorkloadPreemptibilityOptions, Preemptibility } from "@/models/workload.model";

export default defineComponent({
  name: "preemptibility-section",
  components: {
    RunaiSection,
    RunaiSelect,
    NotificationAlertBar,
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
    supportSemiPreemptibility: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    singlePodWorkload: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PreemptibilityRules | null>,
      required: false,
    },
    showQuotaWarning: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hint: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      tooltipText: `Non-preemptible workloads will use the project's available GPU quota and won't be interrupted once they start running.`,
      wasPreemptibilityReset: false,
      ENotificationBarVariant,
      Preemptibility,
      SelectType,
    };
  },

  computed: {
    preemptibilityOptions(): ISelectOption[] {
      const retainSelectedSemiPreemptibility =
        this.hasSinglePodSemiPreemptibility && !this.shouldResetSinglePodPreemptibility;
      const showSemiPreemptibility =
        this.supportSemiPreemptibility && (!this.singlePodWorkload || retainSelectedSemiPreemptibility);

      return getWorkloadPreemptibilityOptions(showSemiPreemptibility).map((option) => {
        const policyOption = this.policyRules?.options?.find(
          (pOption: PreemptibilityOptionsOptionsInner) => pOption.value === option.value,
        );
        const isDisabled = this.policyRules?.options ? !policyOption : false;

        // ONLY show warning tooltip for non-preemptible when quota is exceeded
        const showQuotaWarning = option.value === Preemptibility.NonPreemptible && this.showQuotaWarning && !isDisabled;

        // Determine tooltip values based on conditions (like priority-section)
        let toolTip: string | undefined;
        let warningTooltip: string | undefined;

        if (showQuotaWarning) {
          warningTooltip =
            "This setting will cause the workload to be pending since the project doesn't have enough available GPU quota to run non-preemptible workloads. The workload will run without interruption when quota becomes available.";
        }
        // Otherwise, both remain undefined (no default tooltips)

        return {
          label: policyOption?.displayed?.trim() || option.label.trim(),
          value: option.value,
          toolTip,
          warningTooltip,
          type: warningTooltip ? SelectType.Warning : undefined,
          isHtmlTooltipText: !!warningTooltip,
          ...(this.singlePodWorkload && option.value === Preemptibility.SemiPreemptible
            ? {
                disable: true,
                disabledTooltip: "Semi-preemptible behaves as non-preemptible for single-pod workloads.",
              }
            : {}),
        };
      });
    },
    hasSinglePodSemiPreemptibility(): boolean {
      return this.singlePodWorkload && this.modelValue === Preemptibility.SemiPreemptible;
    },
    shouldResetSinglePodPreemptibility(): boolean {
      return (
        this.hasSinglePodSemiPreemptibility &&
        !this.isSectionDisabled &&
        (!this.policyRules?.options?.length ||
          this.policyRules.options.some((option) => option.value === Preemptibility.NonPreemptible))
      );
    },
    preemptibilityHint(): string | undefined {
      let singlePodHint: string | undefined;
      if (this.hasSinglePodSemiPreemptibility) {
        singlePodHint = "Semi-preemptible behaves as non-preemptible for single-pod workloads.";
      } else if (
        this.singlePodWorkload &&
        this.wasPreemptibilityReset &&
        this.modelValue === Preemptibility.NonPreemptible
      ) {
        singlePodHint = "Changed to non-preemptible because this workload is configured for a single pod.";
      }
      return [this.hint, singlePodHint].filter(Boolean).join(" ") || undefined;
    },
    selectedPreemptibilityOption(): ISelectOption | null {
      if (!this.modelValue) return null;
      return this.preemptibilityOptions.find((option) => option.value === this.modelValue) || null;
    },
    selectGenericPolicyRules(): IGenericSelectPolicyRules | undefined {
      if (!this.policyRules) return undefined;
      return {
        sourceOfRule: this.policyRules.sourceOfRule,
        required: this.policyRules.required,
        canEdit: this.policyRules.canEdit,
        options: this.policyRules.options,
        isClosedList: true,
      };
    },
    isSectionDisabled(): boolean {
      return this.disabled || this.policyRules?.canEdit === false;
    },

    shouldShowQuotaWarning(): boolean {
      // Only show warning when non-preemptible is selected and quota is exceeded
      return this.showQuotaWarning && this.modelValue === Preemptibility.NonPreemptible;
    },

    quotaWarningText(): string {
      return `This setting will cause the workload to be <span class="text-bold">pending</span> since the project doesn't have enough available GPU quota to run non-preemptible workloads. The workload will run without interruption when quota becomes available.`;
    },
  },
  methods: {
    updatePreemptibility(selectedOption: ISelectOption | null): void {
      this.wasPreemptibilityReset = false;
      this.$emit("update:model-value", selectedOption?.value || null);
    },
  },
  watch: {
    shouldResetSinglePodPreemptibility: {
      immediate: true,
      handler(shouldReset: boolean): void {
        if (!shouldReset) return;
        this.wasPreemptibilityReset = true;
        this.$emit("update:model-value", Preemptibility.NonPreemptible);
      },
    },
  },
});
</script>
