``
<template>
  <runai-section class="q-mb-lg" title="Set the workload priority" :tooltip-text="tooltipText">
    <div class="row items-center q-mt-lg q-gutter-x-md">
      <runai-select
        class="priority-select col-3"
        aid="priority-select"
        :class="{ 'q-pb-none': !hint }"
        :model-value="selectedPriorityOption"
        @update:model-value="updatePriority"
        :options="workloadPrioritiesToDisplay"
        :policy-rules="selectGenericPolicyRules"
        :disable="isSectionDisabled"
        label="Priority"
        :hint="hint"
      />
      <notification-alert-bar
        v-if="shouldShowQuotaWarningBanner"
        :variant="ENotificationBarVariant.Warning"
        class="flex-1 q-mt-md"
      >
        <span v-dompurify-html="quotaWarningText"></span>
      </notification-alert-bar>
    </div>
    <q-tooltip v-if="policyRules?.canEdit === false" target=".priority-select" :offset="[0, -15]">
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
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
// Services
import { workloadService } from "@/services/cluster/workload.service/workload.service";
// Models
import type { ClusterPriority } from "@/swagger-models/workloads-client";
import type { StringRules, StringOption } from "@/swagger-models/policy-service-client";
import type { ISelectOption } from "@/models/global.model";
import type { IGenericSelectPolicyRules } from "@/models/policy.model";
import type { IWorkloadPriorityData } from "../general-section.model";
import { SelectType } from "@/components/common/runai-select/runai-select.vue";
export default defineComponent({
  name: "priority-section",
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
    policyRules: {
      type: [Object, null] as PropType<StringRules | null>,
      required: false,
    },
    workloadPriorityData: {
      type: Object as PropType<IWorkloadPriorityData>,
      required: false,
    },
    showQuotaWarning: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportPreemptibility: {
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
      tooltipText: `Higher-priority workloads receive scheduling preference in their project.`,
      nonPreemptibleTooltip: `This priority makes the workload non-preemptible, which means it will use the project's available GPU quota and won't be interrupted once it starts running.`,
      ENotificationBarVariant,
    };
  },

  computed: {
    workloadPriorities(): ClusterPriority[] | null | undefined {
      return this.workloadPriorityData?.workloadPriorities;
    },
    workloadPrioritiesToDisplay(): ISelectOption[] {
      if (!this.workloadPriorities) return [];
      return this.workloadPriorities
        .slice()
        .sort((a: ClusterPriority, b: ClusterPriority) => a.value - b.value)
        .map((priority: ClusterPriority) => {
          const policyOption = this.policyRules?.options?.find((option: StringOption) => option.value === priority.name);
          const isDisabled = this.policyRules?.options ? !policyOption : false;

          // ONLY show non-preemptible tooltip for clusters < 2.24
          const shouldShowNonPreemptibleTooltip = !this.supportPreemptibility && priority.value >= 100 && !isDisabled;

          // ONLY show quota warnings for clusters < 2.24
          const showQuotaWarning =
            !this.supportPreemptibility &&
            workloadService.exceedsAvailableQuota({
              priorityValue: priority.value,
              currProject: this.workloadPriorityData?.currProject,
              selectedNodePoolsResources: this.workloadPriorityData?.selectedNodePoolsResources,
              totalGpuRequest: this.workloadPriorityData?.totalGpuRequest,
            }) &&
            !isDisabled;

          // Determine tooltip values based on conditions
          const { toolTip, warningTooltip } = this.determineTooltipValues(
            showQuotaWarning,
            shouldShowNonPreemptibleTooltip,
          );

          return {
            label: policyOption?.displayed?.trim() || priority.name.trim(),
            value: priority.name,
            toolTip,
            warningTooltip,
            type: warningTooltip ? SelectType.Warning : undefined,
            isHtmlTooltipText: !!warningTooltip,
          };
        });
    },
    selectedPriorityOption(): ISelectOption | null {
      if (!this.modelValue) return null;
      return this.workloadPrioritiesToDisplay.find((option) => option.value === this.modelValue) || null;
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

    shouldShowQuotaWarningBanner(): boolean {
      return !this.supportPreemptibility && this.showQuotaWarning;
    },

    quotaWarningText(): string {
      return `This priority will cause the workload to be <span class="text-bold">pending</span> since the project doesn't have enough available GPU quota to run non-preemptible workloads. The workload will run without interruption when quota becomes available.`;
    },
  },
  methods: {
    updatePriority(selectedOption: ISelectOption | null): void {
      this.$emit("update:model-value", selectedOption?.value || null);
    },
    getTooltipText(showQuotaWarning: boolean, shouldShowNonPreemptibleTooltip: boolean): string | undefined {
      if (showQuotaWarning) {
        return this.quotaWarningText;
      }

      if (shouldShowNonPreemptibleTooltip) {
        return this.nonPreemptibleTooltip;
      }

      return undefined;
    },
    determineTooltipValues(
      showQuotaWarning: boolean,
      shouldShowNonPreemptibleTooltip: boolean,
    ): { toolTip: string | undefined; warningTooltip: string | undefined } {
      if (showQuotaWarning) {
        return {
          toolTip: undefined,
          warningTooltip: this.quotaWarningText,
        };
      }

      if (shouldShowNonPreemptibleTooltip) {
        return {
          toolTip: this.nonPreemptibleTooltip,
          warningTooltip: undefined,
        };
      }

      return {
        toolTip: undefined,
        warningTooltip: undefined,
      };
    },
  },
});
</script>
