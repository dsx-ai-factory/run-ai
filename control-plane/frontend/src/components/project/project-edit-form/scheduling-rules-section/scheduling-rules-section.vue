<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    label="Scheduling rules"
    :section-invalid="!sectionValid"
    :loading="loading"
    :default-opened="defaultOpened"
    :disable-closing="disableClosing"
    body-no-padding
    :prevent-top-padding-when-open="preventTopPaddingWhenOpen"
    :hide-loading-skeleton="hideLoadingSkeleton"
  >
    <template #subheader>{{ expansionSubHeader }} </template>
    <q-card flat>
      <q-card-section :class="{ 'no-padding': bodyNoPadding }">
        <div class="q-my-md">
          <span class="block q-my-md">
            Set rules to control utilization of the {{ entity + "'s" }} compute resources
          </span>
          <span class="text-italic">
            For more information, see the
            <a target="_blank" :href="DOCS.PLATFORM_MANAGEMENT.SCHEDULING_RULES"> Scheduling Rules </a>
            guide
          </span>
        </div>
        <!-- workload max idle duration -->
        <workload-duration-configurator
          v-if="showWorkloadMaxIdleDurationSection"
          header="Idle GPU timeout"
          subheader="Limit the duration of a workload whose GPU is idle"
          add-duration-button-label="IDLE GPU TIMEOUT"
          :workloads-duration="workloadsMaxIdleDuration"
          :workload-options="workloadMaxIdleDurationOptions"
          :read-only="readOnly"
          @add-duration="addWorkloadMaxIdleDuration"
          @remove-duration="removeWorkloadMaxIdleDuration"
          @duration-changed="onWorkloadMaxIdleDurationChanged"
          @close="hideSection($options.ERulesSection.IdleGpuTimeout)"
        />

        <!-- workload time limit duration -->
        <workload-duration-configurator
          v-if="showWorkloadTimeLimitDurationSection"
          header="Workload time limit"
          subheader="Set a time limit for workspaces regardless of their activity (e.g., stop the workspace after 1 day of work)"
          add-duration-button-label="WORKLOAD TIME LIMIT"
          :workloads-duration="workloadTimeLimitDuration"
          :workload-options="workloadTimeLimitDurationOptions"
          :read-only="readOnly"
          @add-duration="addWorkloadTimeLimitDuration"
          @remove-duration="removeWorkloadTimeLimitDuration"
          @duration-changed="onWorkloadTimeLimitDurationChanged"
          @close="hideSection($options.ERulesSection.WorkloadTimeLimit)"
        />
        <node-affinity-section
          v-if="showNodeAffinitySection"
          :cluster-id="clusterId"
          :workloads="nodeAffinityWorkloads"
          :parent-node-types="parentNodeTypes"
          :read-only="readOnly"
          @add-workload="addNodeAffinity"
          @remove-workload="removeNodeAffinity"
          @workload-changed="onNodeAffinityChanged"
          @add-new-node-affinity-type="addNewNodeAffinityType"
          @update-workload-loading="updateNodeAffinityLoading"
          @close="hideSection($options.ERulesSection.NodeAffinity)"
        />
        <runai-button-with-menu
          v-if="showRulesButton"
          :disable="readOnly"
          button-label="RULE"
          aid="add-scheduling-rule-button"
          :options="sectionOptions"
          @item-clicked="showSection"
        />
      </q-card-section>
    </q-card>
    <template v-if="shouldShowSaveButton">
      <q-separator />
      <q-card-actions align="right" class="q-px-lg">
        <q-btn color="primary" flat label="Save" @click="$emit('save')" />
      </q-card-actions>
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiButtonWithMenu } from "@/components/common/runai-button-with-menu";
import { NodeAffinitySection } from "@/components/project/project-edit-form/scheduling-rules-section/node-affinity-section";
import { WorkloadDurationConfigurator } from "@/components/project/project-edit-form/scheduling-rules-section/workload-duration-configurator/";
// models
import type { ISelectOption } from "@/models/global.model";
import { EWorkloadTimeLimit, type IWorkloadDurationOption } from "@/models/workload.model";
import { EIdleWorkloadMaxIdleDuration } from "@/models/workload.model";
import type { INodeAffinitySelectOption } from "@/models/project.model";
import { EWorkloadNodeAffinity } from "@/models/project.model";
//utils
import type { NodeTypesPerWorkload, SchedulingRules } from "@/swagger-models/org-unit-service-client";
import type { EQuotaEntity } from "@/models/resource.model";
import { DOCS } from "@/constants/docs";

export enum ERulesSection {
  IdleGpuTimeout = "idleGpuTimeout",
  WorkloadTimeLimit = "workloadTimeLimit",
  NodeAffinity = "nodeAffinity",
}

export enum ESectionLabel {
  IdleGpuTimeout = "Idle GPU timeout",
  WorkloadTimeLimit = "Workload time limit",
  NodeAffinity = "Node type (Affinity)",
}

export default defineComponent({
  components: {
    WorkloadDurationConfigurator,
    NodeAffinitySection,
    RunaiButtonWithMenu,
    RunaiExpansionItem,
  },
  emits: ["update:scheduling-rules", "update:node-types", "is-section-invalid", "save"],
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    schedulingRules: {
      type: [Object, null] as PropType<SchedulingRules | null>,
      required: false,
    },
    parentSchedulingRules: {
      type: [Object, null] as PropType<SchedulingRules | null>,
      required: false,
    },
    nodeTypes: {
      type: Object as PropType<NodeTypesPerWorkload>,
      required: false,
    },
    parentNodeTypes: {
      type: Object as PropType<NodeTypesPerWorkload>,
      required: false,
    },
    entity: {
      type: String as PropType<EQuotaEntity>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    defaultOpened: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableClosing: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showSaveButton: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    preventTopPaddingWhenOpen: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    bodyNoPadding: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideLoadingSkeleton: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  ERulesSection: ERulesSection,
  data() {
    const sectionOptions = [
      {
        value: ERulesSection.IdleGpuTimeout,
        label: `+ ${ESectionLabel.IdleGpuTimeout}`,
        disable: false,
      },
      {
        value: ERulesSection.WorkloadTimeLimit,
        label: `+ ${ESectionLabel.WorkloadTimeLimit}`,
        disable: false,
      },
      {
        value: ERulesSection.NodeAffinity,
        label: `+ ${ESectionLabel.NodeAffinity}`,
        disable: false,
      },
    ] as ISelectOption[];
    return {
      showWorkloadMaxIdleDurationSection: false as boolean,
      showWorkloadTimeLimitDurationSection: false as boolean,
      showNodeAffinitySection: false as boolean,
      workloadsMaxIdleDuration: [] as IWorkloadDurationOption[],
      workloadTimeLimitDuration: [] as IWorkloadDurationOption[],
      nodeAffinityWorkloads: [] as INodeAffinitySelectOption[],
      sectionOptions: sectionOptions,
      DOCS,
    };
  },
  created() {
    this.initWorkloadMaxIdleDurationSection();
    this.initWorkloadTimeLimitDurationSection();
    this.initNodeAffinitySection();
  },
  computed: {
    sectionValid(): boolean {
      if (!this.isSectionOpen) {
        return true;
      }

      return (
        [
          this.isWorkloadMaxIdleSectionValid,
          this.isWorkloadTimeLimitSectionValid,
          this.isNodeAffinitySectionValid,
        ].every((isValid) => isValid) && !this.isProjectExceedsInheritedDuration
      );
    },
    isProjectExceedsInheritedDuration(): boolean {
      return this.workloadsMaxIdleDuration.some(
        (workload: IWorkloadDurationOption) =>
          workload.inheritedDuration && workload.duration && workload.duration > workload.inheritedDuration,
      );
    },
    isWorkloadMaxIdleSectionValid(): boolean {
      if (!this.showWorkloadMaxIdleDurationSection) return true;
      return !this.workloadsMaxIdleDuration.some(
        (workload: IWorkloadDurationOption) => !workload.value || workload.duration === null,
      );
    },
    isWorkloadTimeLimitSectionValid(): boolean {
      if (!this.showWorkloadTimeLimitDurationSection) return true;
      return !this.workloadTimeLimitDuration.some(
        (workload: IWorkloadDurationOption) => !workload.value || workload.duration === null,
      );
    },
    isNodeAffinitySectionValid(): boolean {
      if (!this.showNodeAffinitySection) return true;
      return !this.nodeAffinityWorkloads.some(
        (workload: INodeAffinitySelectOption) => !workload.value || workload.selectedTypes.length === 0,
      );
    },
    isSectionOpen(): boolean {
      return (
        this.showWorkloadMaxIdleDurationSection ||
        this.showWorkloadTimeLimitDurationSection ||
        this.showNodeAffinitySection
      );
    },
    showRulesButton(): boolean {
      return this.sectionOptions.some((item: ISelectOption) => !item.disable);
    },
    shouldShowSaveButton(): boolean {
      if (!this.showSaveButton) {
        return false;
      }
      return this.isSectionOpen || !this.isAllSectionsEmpty;
    },
    isAllSectionsEmpty(): boolean {
      return (
        !this.schedulingRules?.trainingJobTimeLimitSeconds &&
        !this.schedulingRules?.interactiveJobTimeLimitSeconds &&
        !this.schedulingRules?.trainingJobMaxIdleDurationSeconds &&
        !this.schedulingRules?.interactiveJobPreemptIdleDurationSeconds &&
        !this.schedulingRules?.interactiveJobMaxIdleDurationSeconds &&
        !this.nodeTypes?.workspace?.length &&
        !this.nodeTypes?.training?.length
      );
    },
    expansionSubHeader(): string {
      if (this.isAllSectionsEmpty) {
        return "None";
      }
      const sections = [];
      if (
        this.getWorkloadDuration(this.workloadsMaxIdleDuration, EIdleWorkloadMaxIdleDuration.training) ||
        this.getWorkloadDuration(this.workloadsMaxIdleDuration, EIdleWorkloadMaxIdleDuration.interactive) ||
        this.getWorkloadDuration(this.workloadsMaxIdleDuration, EIdleWorkloadMaxIdleDuration.interactivePreemptible)
      ) {
        sections.push(ESectionLabel.IdleGpuTimeout);
      }

      if (
        this.getWorkloadDuration(this.workloadTimeLimitDuration, EWorkloadTimeLimit.training) ||
        this.getWorkloadDuration(this.workloadTimeLimitDuration, EWorkloadTimeLimit.interactive)
      ) {
        sections.push(ESectionLabel.WorkloadTimeLimit);
      }

      if (this.nodeTypes?.workspace?.length || this.nodeTypes?.training?.length) {
        sections.push(ESectionLabel.NodeAffinity);
      }
      return sections.join(" / ");
    },
    //********* workload max idle duration  *********
    workloadMaxIdleDurationOptions(): ISelectOption[] {
      const isOptionDisabled = (value: EIdleWorkloadMaxIdleDuration): boolean =>
        this.workloadsMaxIdleDuration.some((workload: IWorkloadDurationOption) => workload.value === value);
      return [
        {
          value: EIdleWorkloadMaxIdleDuration.training,
          label: "Training",
          disable: isOptionDisabled(EIdleWorkloadMaxIdleDuration.training),
        },
        {
          value: EIdleWorkloadMaxIdleDuration.interactivePreemptible,
          label: "Preemptible workspaces",
          disable: isOptionDisabled(EIdleWorkloadMaxIdleDuration.interactivePreemptible),
        },
        {
          value: EIdleWorkloadMaxIdleDuration.interactive,
          label: "Non-preemptible workspaces",
          disable: isOptionDisabled(EIdleWorkloadMaxIdleDuration.interactive),
        },
      ] as ISelectOption[];
    },
    //********* workload time limit duration *********
    workloadTimeLimitDurationOptions(): ISelectOption[] {
      const isOptionDisabled = (value: EWorkloadTimeLimit): boolean =>
        this.workloadTimeLimitDuration.some((workload) => workload.value === value);

      return [
        {
          value: EWorkloadTimeLimit.training,
          label: "Training",
          disable: isOptionDisabled(EWorkloadTimeLimit.training),
        },
        {
          value: EWorkloadTimeLimit.interactive,
          label: "Workspace",
          disable: isOptionDisabled(EWorkloadTimeLimit.interactive),
        },
      ];
    },
  },
  methods: {
    //********* workload max idle duration  *********
    initWorkloadMaxIdleDurationSection() {
      const workloadsMaxIdleDuration: IWorkloadDurationOption[] = [
        {
          label: "Training",
          value: EIdleWorkloadMaxIdleDuration.training,
          duration: this.schedulingRules?.trainingJobMaxIdleDurationSeconds,
          inheritedDuration: this.parentSchedulingRules?.trainingJobMaxIdleDurationSeconds,
        },
        {
          label: "Preemptible workspaces",
          value: EIdleWorkloadMaxIdleDuration.interactivePreemptible,
          duration: this.schedulingRules?.interactiveJobPreemptIdleDurationSeconds,
          inheritedDuration: this.parentSchedulingRules?.interactiveJobPreemptIdleDurationSeconds,
        },
        {
          label: "Non-preemptible workspaces",
          value: EIdleWorkloadMaxIdleDuration.interactive,
          duration: this.schedulingRules?.interactiveJobMaxIdleDurationSeconds,
          inheritedDuration: this.parentSchedulingRules?.interactiveJobMaxIdleDurationSeconds,
        },
      ];

      workloadsMaxIdleDuration.forEach((workloadDuration: IWorkloadDurationOption) => {
        if (
          workloadDuration.duration &&
          !this.workloadsMaxIdleDuration.some((workload) => workload.value === workloadDuration.value)
        ) {
          this.workloadsMaxIdleDuration.push(workloadDuration);
          this.showSection(ERulesSection.IdleGpuTimeout);
        }
      });
    },
    addWorkloadMaxIdleDuration(): void {
      this.workloadsMaxIdleDuration.push({
        label: "",
        value: "",
        duration: 0,
      });
    },
    removeWorkloadMaxIdleDuration(workloadIndex: number): void {
      this.workloadsMaxIdleDuration.splice(workloadIndex, 1);
      this.updateWorkloadMaxIdleSchedulingRules();
    },
    onWorkloadMaxIdleDurationChanged(workload: IWorkloadDurationOption, workloadIndex: number): void {
      this.workloadsMaxIdleDuration.splice(workloadIndex, 1, workload);
      this.updateWorkloadMaxIdleSchedulingRules();
    },
    updateWorkloadMaxIdleSchedulingRules(): void {
      this.$emit("update:scheduling-rules", {
        ...this.schedulingRules,
        trainingJobMaxIdleDurationSeconds: this.getWorkloadDuration(
          this.workloadsMaxIdleDuration,
          EIdleWorkloadMaxIdleDuration.training,
        ),
        interactiveJobPreemptIdleDurationSeconds: this.getWorkloadDuration(
          this.workloadsMaxIdleDuration,
          EIdleWorkloadMaxIdleDuration.interactivePreemptible,
        ),
        interactiveJobMaxIdleDurationSeconds: this.getWorkloadDuration(
          this.workloadsMaxIdleDuration,
          EIdleWorkloadMaxIdleDuration.interactive,
        ),
      });
    },

    //********* workload time limit duration *********
    initWorkloadTimeLimitDurationSection() {
      const workloadsTimeLimit: IWorkloadDurationOption[] = [
        {
          label: "Training",
          value: EWorkloadTimeLimit.training,
          duration: this.schedulingRules?.trainingJobTimeLimitSeconds,
          inheritedDuration: this.parentSchedulingRules?.trainingJobTimeLimitSeconds,
        },
        {
          label: "Workspace",
          value: EWorkloadTimeLimit.interactive,
          duration: this.schedulingRules?.interactiveJobTimeLimitSeconds,
          inheritedDuration: this.parentSchedulingRules?.interactiveJobTimeLimitSeconds,
        },
      ];

      workloadsTimeLimit.forEach((workloadDuration: IWorkloadDurationOption) => {
        if (
          workloadDuration.duration &&
          !this.workloadTimeLimitDuration.some((workload) => workload.value === workloadDuration.value)
        ) {
          this.workloadTimeLimitDuration.push(workloadDuration);
          this.showSection(ERulesSection.WorkloadTimeLimit);
        }
      });
    },
    addWorkloadTimeLimitDuration(): void {
      this.workloadTimeLimitDuration.push({
        label: "",
        value: "",
        duration: 0,
      });
    },
    removeWorkloadTimeLimitDuration(workloadIndex: number): void {
      this.workloadTimeLimitDuration.splice(workloadIndex, 1);
      this.updateWorkloadTimeLimitSchedulingRules();
    },
    onWorkloadTimeLimitDurationChanged(workload: IWorkloadDurationOption, workloadIndex: number): void {
      this.workloadTimeLimitDuration.splice(workloadIndex, 1, workload);
      this.updateWorkloadTimeLimitSchedulingRules();
    },
    updateWorkloadTimeLimitSchedulingRules(): void {
      this.$emit("update:scheduling-rules", {
        ...this.schedulingRules,
        trainingJobTimeLimitSeconds: this.getWorkloadDuration(
          this.workloadTimeLimitDuration,
          EWorkloadTimeLimit.training,
        ),
        interactiveJobTimeLimitSeconds: this.getWorkloadDuration(
          this.workloadTimeLimitDuration,
          EWorkloadTimeLimit.interactive,
        ),
      });
    },

    //********* node affinity *********
    initNodeAffinitySection(): void {
      const nodeTypesNames = this.nodeTypes?.names || {};
      const pushNodeAffinityWorkload = (nodeTypes: string[], label: string, value: EWorkloadNodeAffinity) => {
        const selectedTypes = nodeTypes.map((nodeTypeId: string) => ({
          label: nodeTypesNames[nodeTypeId],
          value: nodeTypeId,
        }));
        if (selectedTypes.length > 0) {
          this.nodeAffinityWorkloads.push({
            label: label,
            value: value,
            loading: false,
            selectedTypes,
          });
          this.showSection(ERulesSection.NodeAffinity);
        }
      };
      pushNodeAffinityWorkload(this.nodeTypes?.training || [], "Training", EWorkloadNodeAffinity.Train);
      pushNodeAffinityWorkload(this.nodeTypes?.workspace || [], "Workspace", EWorkloadNodeAffinity.Interactive);
    },
    addNodeAffinity(): void {
      this.nodeAffinityWorkloads.push({
        label: "",
        value: "",
        loading: false,
        selectedTypes: [],
      });
    },
    addNewNodeAffinityType(workloadIndex: number, newType: ISelectOption): void {
      this.nodeAffinityWorkloads[workloadIndex].selectedTypes.push(newType);
    },
    updateNodeAffinityLoading(workloadIndex: number, isLoading: boolean): void {
      this.nodeAffinityWorkloads[workloadIndex].loading = isLoading;
    },
    onNodeAffinityChanged(workload: INodeAffinitySelectOption, workloadIndex: number): void {
      this.nodeAffinityWorkloads.splice(workloadIndex, 1, workload);
    },
    removeNodeAffinity(workloadIndex: number): void {
      this.nodeAffinityWorkloads.splice(workloadIndex, 1);
    },

    getSectionIndexByName(optionValue: string): number {
      return this.sectionOptions.findIndex((option: ISelectOption) => option.value === optionValue);
    },
    resetSection(section: ERulesSection): void {
      switch (section) {
        case ERulesSection.IdleGpuTimeout:
          this.workloadsMaxIdleDuration = [];
          this.updateWorkloadMaxIdleSchedulingRules();
          break;
        case ERulesSection.WorkloadTimeLimit:
          this.workloadTimeLimitDuration = [];
          this.updateWorkloadTimeLimitSchedulingRules();
          break;
        case ERulesSection.NodeAffinity:
          this.nodeAffinityWorkloads = [];
      }
    },
    resetAllSections(): void {
      [ERulesSection.IdleGpuTimeout, ERulesSection.WorkloadTimeLimit].forEach((section) => this.resetSection(section));
    },
    hideAllSections(): void {
      [ERulesSection.IdleGpuTimeout, ERulesSection.WorkloadTimeLimit].forEach((section) => this.hideSection(section));
    },
    showSection(section: ERulesSection): void {
      const optionIndex = this.getSectionIndexByName(section);
      this.sectionOptions[optionIndex].disable = true;
      this.toggleSectionVisibility(section, true);
      this.addOrResetSection(section, true);
    },
    hideSection(section: ERulesSection): void {
      const optionIndex = this.getSectionIndexByName(section);
      this.sectionOptions[optionIndex].disable = false;
      this.toggleSectionVisibility(section, false);
      this.addOrResetSection(section, false);
    },
    toggleSectionVisibility(section: ERulesSection, visible: boolean): void {
      switch (section) {
        case ERulesSection.IdleGpuTimeout:
          this.showWorkloadMaxIdleDurationSection = visible;
          break;
        case ERulesSection.WorkloadTimeLimit:
          this.showWorkloadTimeLimitDurationSection = visible;
          break;
        case ERulesSection.NodeAffinity:
          this.showNodeAffinitySection = visible;
      }
    },
    addOrResetSection(section: ERulesSection, visible: boolean): void {
      switch (section) {
        case ERulesSection.IdleGpuTimeout:
          if (!visible) {
            this.resetSection(ERulesSection.IdleGpuTimeout);
          } else if (this.workloadsMaxIdleDuration.length === 0) {
            this.addWorkloadMaxIdleDuration();
          }
          break;
        case ERulesSection.WorkloadTimeLimit:
          if (!visible) {
            this.resetSection(ERulesSection.WorkloadTimeLimit);
          } else if (this.workloadTimeLimitDuration.length === 0) {
            this.addWorkloadTimeLimitDuration();
          }
          break;
        case ERulesSection.NodeAffinity:
          if (!visible) {
            this.resetSection(ERulesSection.NodeAffinity);
          } else if (this.nodeAffinityWorkloads.length === 0) {
            this.addNodeAffinity();
          }
      }
    },
    getWorkloadDuration(
      workloads: IWorkloadDurationOption[],
      value: EIdleWorkloadMaxIdleDuration | EWorkloadTimeLimit,
    ): number | null {
      const idleWorkload = workloads.find((workload: IWorkloadDurationOption) => workload.value === value);

      if (idleWorkload && idleWorkload.duration) {
        return idleWorkload.duration;
      } else {
        return null;
      }
    },
  },
  watch: {
    parentSchedulingRules: {
      handler() {
        this.hideAllSections();
        this.resetAllSections();
        this.initWorkloadMaxIdleDurationSection();
        this.initWorkloadTimeLimitDurationSection();
      },
    },
    nodeAffinityWorkloads: {
      handler(newVal: INodeAffinitySelectOption[]) {
        const getNodeTypes = (workloadType: EWorkloadNodeAffinity) =>
          newVal
            .filter(({ value }: INodeAffinitySelectOption) => value === workloadType)
            .flatMap(({ selectedTypes }: INodeAffinitySelectOption) =>
              selectedTypes.map(({ value }) => value?.toString()),
            );

        const nodeTypes: NodeTypesPerWorkload = {};
        const trainingNodeTypes = getNodeTypes(EWorkloadNodeAffinity.Train);
        const workspaceNodeTypes = getNodeTypes(EWorkloadNodeAffinity.Interactive);

        if (trainingNodeTypes.length) nodeTypes.training = trainingNodeTypes as string[];
        if (workspaceNodeTypes.length) nodeTypes.workspace = workspaceNodeTypes as string[];
        const nodeNames: Record<string, string> = {};

        newVal.forEach((workload) => {
          workload.selectedTypes.forEach((type) => {
            nodeNames[type.value as string] = type.label;
          });
        });
        nodeTypes.names = nodeNames;

        this.$emit("update:node-types", nodeTypes);
      },
      deep: true,
    },
  },
});
</script>
