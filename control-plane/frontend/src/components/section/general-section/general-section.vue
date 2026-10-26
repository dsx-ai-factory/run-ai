<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    :prevent-top-padding-when-open="preventTopPaddingWhenOpen"
    class="general-section"
    label="General"
    aid="general-section"
    :default-opened="defaultOpened"
    :section-invalid="sectionInvalid"
    :should-validate="shouldAutoValidate"
  >
    <template #subheader>
      <div class="general-section__subheader row no-wrap items-center q-pr-md">
        <span class="general-section__summary">{{ summary }}</span>
        <q-icon
          v-if="shouldShowQuotaWarning"
          name="ph-bold ph-warning-circle"
          size="20px"
          color="warning"
          class="general-section__quota-warning-icon q-ml-sm"
        />
      </div>
    </template>
    <section class="row no-wrap justify-between q-mb-md" v-if="showOverQuotaSection">
      <runai-toggle
        aid="allow-over-quota-input"
        :model-value="generalModel.allowOverQuota || false"
        :label="`Allow the workload to exceed the project's quota`"
        left-label
        @update:model-value="allowOverQuotaChanged"
        :disable="sectionDisabled"
      />
      <!--is this relevant?: no policy for this property yet -->
      <q-banner class="bg-orange-1 q-ml-sm col-5">
        <div class="row no-wrap items-center">
          <q-icon name="ph-bold ph-warning" size="sm" color="warning" />
          <div class="q-ml-md text-caption">Note: Workloads running over quota may stop at any time</div>
        </div>
      </q-banner>
    </section>

    <template v-if="!distributedReplicaType">
      <preemptibility-section
        v-if="supportPreemptibility"
        :model-value="generalModel?.preemptibility || null"
        :policy-rules="policyRules?.preemptibility || null"
        :disabled="sectionDisabled"
        :support-semi-preemptibility="workloadPriorityData?.supportSemiPreemptibility"
        :single-pod-workload="isSinglePodWorkload"
        @update:model-value="updateWorkloadPreemptibility"
        :show-quota-warning="shouldShowQuotaWarning"
        :hint="overrideHints.preemptibility ? 'Overrides YAML' : undefined"
      />
      <priority-section
        v-if="workloadPriorityData?.supportPriorityClass"
        :model-value="generalModel?.priorityClass || null"
        :policy-rules="policyRules?.priorityClass"
        :disabled="sectionDisabled"
        @update:model-value="updateWorkloadPriority"
        :workload-priority-data="workloadPriorityData"
        :show-quota-warning="shouldShowQuotaWarning"
        :support-preemptibility="supportPreemptibility"
        :hint="overrideHints.priority ? 'Overrides YAML' : undefined"
      />
      <mnnvl-section
        v-if="supportMnnvl"
        :model-value="generalModel?.mnnvl || null"
        :disabled="sectionDisabled"
        @update:model-value="updateMnnvl"
      />
      <div
        v-if="workloadPriorityData?.supportPriorityClass || supportPreemptibility"
        class="q-mb-lg section-divider"
      ></div>
      <termination-grace-period-section
        v-if="showTerminateGracePeriod"
        :read-only="sectionDisabled"
        :model-value="generalModel.terminationGracePeriodSeconds"
        :policy-rules="policyRules?.terminationGracePeriodSeconds"
        @update:model-value="updateTerminationGracePeriodSeconds"
      />

      <runs-and-parallelism-section
        v-if="showRunsAndParallelism"
        class="q-mb-lg"
        :model-value="{ completions: generalModel.completions, parallelism: generalModel.parallelism }"
        @update:model-value="updateRunsAndParallelism"
        :disabled="sectionDisabled"
        :policy-rules="runsAndParallelismRules"
      />

      <backoff-limit-section
        class="q-mb-lg"
        v-if="showBackoffLimit"
        :model-value="generalModel.backoffLimit"
        @update:model-value="updateBackoffLimit"
        :disabled="sectionDisabled"
        :policy-rules="policyRules?.backoffLimit"
      />

      <auto-delete-time-section
        v-if="showAutoDeletion"
        :model-value="generalModel.autoDeletionTimeAfterCompletionSeconds"
        @update:model-value="autoDeletionTimeoutChanged"
        :disable="sectionDisabled"
        :policy-rules="policyRules?.autoDeletionTimeAfterCompletionSeconds"
      />

      <ssh-authorization
        v-if="showSshAuthMountPath"
        :model-value="generalModel.sshAuthMountPath || ''"
        @update:model-value="updateSshAuthMountPath"
        :section-disabled="sectionDisabled"
        :policy-rules="policyRules?.sshAuthMountPath"
      />

      <clean-pod-policy-section
        v-if="supportCleanPodPolicy"
        :disable="sectionDisabled"
        :model-value="generalModel.cleanPodPolicy || null"
        :policy-rules="policyRules?.cleanPodPolicy"
        @update:model-value="updateCleanPodPolicy"
      />

      <request-timeout-section
        v-if="showRequestTimeout"
        :model-value="generalModel.servingConfiguration || {}"
        @update:model-value="updateRequestTimeout"
        :disable="sectionDisabled"
        :policy-rules="policyRules?.servingConfiguration"
      />
    </template>

    <launcher-policy-section
      class="q-mt-md"
      v-if="showMpiLauncherCreationPolicy"
      :disable="sectionDisabled"
      :model-value="generalModel.mpiLauncherCreationPolicy || null"
      :policy-rules="policyRules?.mpiLauncherCreationPolicy"
      @update:model-value="updateLauncherPolicy"
    />

    <template v-if="showAnnotationsAndLabels">
      <section class="annotations-section">
        <runai-itemized-list
          :padding-top="distributedReplicaType ? false : true"
          :items-list="annotationsList"
          title="Set annotation(s)"
          title-tooltip="Kubernetes annotations are key-value pairs attached to the workload. They are used for storing additional descriptive metadata to enable documentation, monitoring and automation."
          add-item-btn="annotation"
          @changed="onAnnotationUpdated"
          :disable="sectionDisabled || !!disableFields?.annotations"
          :disable-info="disableFields?.annotations"
          :policy-defaults="itemizeDefaults?.annotations"
          :policy-rules="policyRules?.annotations"
        />
      </section>
      <section class="labels-section">
        <runai-itemized-list
          :items-list="labelsList"
          title="Set label(s)"
          title-tooltip="Kubernetes labels are key-value pairs attached to the workload. They are used for categorizing to enable querying."
          add-item-btn="label"
          @changed="onLabelUpdated"
          :disable="sectionDisabled || !!disableFields?.labels"
          :disable-info="disableFields?.labels"
          :policy-defaults="itemizeDefaults?.labels"
          :policy-rules="policyRules?.labels"
        />
      </section>
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiItemizedList } from "@/components/common";
import { AutoDeleteTimeSection } from "./auto-delete-time-section";
import { BackoffLimitSection } from "./backoff-limit-section";
import { RunaiToggle } from "@/components/common";
import { TerminationGracePeriodSection } from "./termination-grace-period-section";
import { CleanPodPolicySection } from "./clean-pod-policy-section";
import { SshAuthorization } from "./ssh-authorization";
import { LauncherPolicySection } from "./launcher-policy-section";
import { RequestTimeoutSection } from "./request-timeout-section";
import { PrioritySection } from "./priority-section";
import { PreemptibilitySection } from "./preemptibility-section";
import { MnnvlSection } from "./mnnvl-section";

// Models
import type {
  IGeneralSectionDisableFields,
  IOverrideHints,
  IUIGeneralSectionModel,
  IWorkloadPriorityData,
} from "./general-section.model";
import type { Preemptibility } from "@/swagger-models/workloads-client";
import type { MnnvlMode } from "@/swagger-models/workloads-manager-client";
import { EDistributedReplicaType, EWorkloadFormType } from "@/models/workload.model";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import type {
  Annotation,
  CleanPodPolicy,
  MpiLauncherCreationPolicy,
  Label,
  ServingConfiguration,
} from "@/swagger-models/assets-service-client";
import type { ILoadedItemizeDefaults } from "@/models/policy.model";
import { RunsAndParallelismSection } from "./runs-and-parallelism-section";
import type {
  ICompletionsAndParallelism,
  IRunsAndParallelismRules,
} from "./runs-and-parallelism-section/runs-and-parallelism-section.vue";
import type { UnifiedPolicyInfoPerReplicaTypeRules } from "@/swagger-models/workloads-client";

// Services
import { workloadService } from "@/services/cluster/workload.service/workload.service";

export enum EGeneralSectionEvents {
  GeneralModelChanged = "general-model-changed",
}

export default defineComponent({
  name: "general-section",
  components: {
    RunaiExpansionItem,
    RunaiItemizedList,
    AutoDeleteTimeSection,
    BackoffLimitSection,
    RunaiToggle,
    RunsAndParallelismSection,
    TerminationGracePeriodSection,
    CleanPodPolicySection,
    SshAuthorization,
    LauncherPolicySection,
    RequestTimeoutSection,
    PrioritySection,
    PreemptibilitySection,
    MnnvlSection,
  },
  emits: [EGeneralSectionEvents.GeneralModelChanged],
  props: {
    workloadFormType: {
      type: String as PropType<EWorkloadFormType>,
      required: true,
    },
    generalModel: {
      type: Object as PropType<IUIGeneralSectionModel>,
      required: true,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    showBackoffLimit: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    showTerminationGracePeriod: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    showRunsAndParallelism: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<UnifiedPolicyInfoPerReplicaTypeRules | null>,
      required: false,
    },
    itemizeDefaults: {
      type: [Object, null] as PropType<ILoadedItemizeDefaults | null>,
      required: false,
    },
    shouldAutoValidate: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    distributedReplicaType: {
      type: String as PropType<EDistributedReplicaType>,
      required: false,
    },
    disableFields: {
      type: Object as PropType<IGeneralSectionDisableFields>,
      required: false,
    },
    supportCleanPodPolicy: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showSshAuthMountPath: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showMpiLauncherCreationPolicy: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showRequestTimeout: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    workloadPriorityData: {
      type: Object as PropType<IWorkloadPriorityData>,
      required: false,
    },
    supportPreemptibility: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    singlePodWorkload: {
      type: Boolean as PropType<boolean>,
      default: undefined,
    },
    supportMnnvl: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showAnnotationsAndLabels: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    showAutoDeletionSection: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    overrideHints: {
      type: Object as PropType<IOverrideHints>,
      default: () => ({}),
    },
    defaultOpened: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    preventTopPaddingWhenOpen: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      EDistributedReplicaType,
    };
  },
  computed: {
    isSinglePodWorkload(): boolean {
      return (
        this.singlePodWorkload ??
        (this.workloadFormType === EWorkloadFormType.Workspace || this.workloadFormType === EWorkloadFormType.Training)
      );
    },
    defaultSummary(): string {
      if (!this.showAnnotationsAndLabels) {
        return "";
      }
      const annotations = this.generalModel.annotations?.length || 0;
      const labels = this.generalModel.labels?.length || 0;
      return `Annotations: ${annotations} / Labels: ${labels}`;
    },
    priorityPreemptibilitySummary(): string {
      const parts: string[] = [];

      if (this.workloadPriorityData?.supportPriorityClass) {
        const priority = this.generalModel.priorityClass || "None";
        parts.push(`Priority: ${priority}`);
      }

      if (this.supportPreemptibility) {
        const preemptibility = this.generalModel.preemptibility || "None";
        parts.push(`Preemptibility: ${preemptibility}`);
      }

      return parts.join(" / ");
    },
    summary(): string {
      if (this.showOverQuotaSection) {
        const prefix: string = this.generalModel.allowOverQuota ? "Allowed" : "Not allowed";
        return prefix + " to exceed the project's quota";
      }

      const parts: string[] = [];

      if (this.showBackoffLimit) {
        parts.push(`Backoff limit (${this.generalModel.backoffLimit || 0}) attempts`);
      }

      if (this.priorityPreemptibilitySummary) {
        parts.push(this.priorityPreemptibilitySummary);
      }

      if (!parts.length) {
        return this.defaultSummary;
      }

      return parts.join(" / ");
    },
    showOverQuotaSection(): boolean {
      return this.workloadFormType === EWorkloadFormType.Workspace && !this.workloadPriorityData?.supportPriorityClass;
    },
    showTerminateGracePeriod(): boolean {
      if (!this.showTerminationGracePeriod) return false;

      switch (this.workloadFormType) {
        case EWorkloadFormType.Workspace:
          return !!this.workloadPriorityData?.supportPriorityClass || !!this.generalModel.allowOverQuota;
        case EWorkloadFormType.Training:
        case EWorkloadFormType.Distributed:
          return true;
        default:
          return false;
      }
    },
    annotationsList(): Array<IItemizedListItem> {
      return this.generalModel.annotations || [];
    },
    labelsList(): Array<IItemizedListItem> {
      return this.generalModel.labels || [];
    },
    invalidAnnotations(): boolean {
      return !!this.generalModel.annotations?.length && this.generalModel.annotations.some((a) => !a.name);
    },
    invalidLabels(): boolean {
      return !!this.generalModel.labels?.length && this.generalModel.labels.some((l) => !l.name);
    },
    sectionInvalid(): boolean {
      const hasInvalidBackoffLimit =
        this.showBackoffLimit &&
        "backoffLimit" in this.generalModel &&
        this.generalModel.backoffLimit !== 0 &&
        !this.generalModel.backoffLimit;
      const hasInvalidAutoDeletion =
        this.showAutoDeletion &&
        "autoDeletionTimeAfterCompletionSeconds" in this.generalModel &&
        typeof this.generalModel.autoDeletionTimeAfterCompletionSeconds !== "number";
      const hasInvalidAnnotationsOrLabels =
        this.showAnnotationsAndLabels && (this.invalidAnnotations || this.invalidLabels);

      return hasInvalidBackoffLimit || hasInvalidAutoDeletion || hasInvalidAnnotationsOrLabels;
    },
    runsAndParallelismRules(): IRunsAndParallelismRules {
      return {
        parallelism: this.policyRules?.parallelism || null,
        completions: this.policyRules?.completions || null,
      };
    },
    showAutoDeletion(): boolean {
      return this.showAutoDeletionSection && this.workloadFormType !== EWorkloadFormType.Inference;
    },
    shouldShowQuotaWarning(): boolean {
      // Only check if priority class is supported
      if (!this.workloadPriorityData?.supportPriorityClass) return false;

      if (this.supportPreemptibility) {
        if (!this.generalModel?.preemptibility) return false;

        return workloadService.exceedsAvailableQuota({
          preemptibility: this.generalModel.preemptibility as Preemptibility,
          currProject: this.workloadPriorityData.currProject,
          selectedNodePoolsResources: this.workloadPriorityData.selectedNodePoolsResources,
          totalGpuRequest: this.workloadPriorityData.totalGpuRequest,
        });
      } else {
        if (!this.generalModel?.priorityClass || !this.workloadPriorityData?.workloadPriorities) return false;

        const selectedPriority = this.workloadPriorityData.workloadPriorities.find(
          (priority) => priority.name === this.generalModel.priorityClass,
        );

        if (!selectedPriority) return false;

        return workloadService.exceedsAvailableQuota({
          priorityValue: selectedPriority.value,
          currProject: this.workloadPriorityData.currProject,
          selectedNodePoolsResources: this.workloadPriorityData.selectedNodePoolsResources,
          totalGpuRequest: this.workloadPriorityData.totalGpuRequest,
        });
      }
    },
  },
  methods: {
    allowOverQuotaChanged(allowOverQuota: boolean): void {
      if (this.workloadFormType === EWorkloadFormType.Workspace && !allowOverQuota) {
        this.$emit(EGeneralSectionEvents.GeneralModelChanged, {
          ...this.generalModel,
          allowOverQuota,
          terminationGracePeriodSeconds: null,
        });
      } else {
        this.$emit(EGeneralSectionEvents.GeneralModelChanged, { ...this.generalModel, allowOverQuota });
      }
    },
    onAnnotationUpdated(annotations: Array<Annotation>) {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, { ...this.generalModel, annotations });
    },
    onLabelUpdated(labels: Array<Label>) {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, { ...this.generalModel, labels });
    },
    autoDeletionTimeoutChanged(autoDeletionTimeAfterCompletionSeconds: number | null): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, {
        ...this.generalModel,
        autoDeletionTimeAfterCompletionSeconds,
      });
    },
    updateBackoffLimit(backoffLimit: number | null): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, { ...this.generalModel, backoffLimit });
    },
    updateRunsAndParallelism(runsAndParallelism: ICompletionsAndParallelism): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, {
        ...this.generalModel,
        completions: runsAndParallelism.completions,
        parallelism: runsAndParallelism.parallelism,
      });
    },
    updateTerminationGracePeriodSeconds(terminationGracePeriodSeconds: number | null): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, { ...this.generalModel, terminationGracePeriodSeconds });
    },
    updateCleanPodPolicy(cleanPodPolicy: CleanPodPolicy | null | undefined): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, {
        ...this.generalModel,
        cleanPodPolicy: cleanPodPolicy || undefined,
      });
    },
    updateSshAuthMountPath(sshAuthMountPath: string | null): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, { ...this.generalModel, sshAuthMountPath });
    },
    updateLauncherPolicy(launcherPolicy: MpiLauncherCreationPolicy | null): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, {
        ...this.generalModel,
        mpiLauncherCreationPolicy: launcherPolicy,
      });
    },
    updateRequestTimeout(servingConfiguration: ServingConfiguration): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, { ...this.generalModel, servingConfiguration });
    },
    updateWorkloadPriority(priorityClass: string | null): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, { ...this.generalModel, priorityClass });
    },
    updateWorkloadPreemptibility(preemptibility: Preemptibility | null): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, { ...this.generalModel, preemptibility });
    },
    updateMnnvl(mnnvl: MnnvlMode | null): void {
      this.$emit(EGeneralSectionEvents.GeneralModelChanged, { ...this.generalModel, mnnvl });
    },
  },
});
</script>

<style lang="scss" scoped>
.general-section__subheader {
  min-width: 0;
  width: 100%;
}

.general-section__summary {
  flex: 1 1 auto;
  min-width: 0;
}

.general-section__quota-warning-icon {
  flex: 0 0 auto;
}

.section-divider {
  border-bottom: 1px solid $black-12;
  margin-top: 48px;
}
</style>
