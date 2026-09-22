<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="inference-distributed-workload-setup column items-center q-pt-md">
    <workload-form-wrapper
      ref="formWrapperRef"
      :form-state="model"
      :submitting="uiState.submitting"
      :submit-btn-text="submitBtnText"
      cancel-btn-text="Back"
      init-dirty
      :blurred="uiState.isDrawerOpen"
      :greedy="isGreedy"
      @unblur="$emit(EEmitName.CloseDrawer)"
      @leave-page="$emit(EEmitName.LeavePage)"
      @canceled="$emit(EEmitName.Back)"
      @submit="$emit(EEmitName.Submit)"
    >
      <workload-setup-compute-section
        :model="computeModel"
        :config="computeConfig"
        @update:model="onComputeModelChanged"
      />

      <workload-setup-general-section
        :model="generalModel"
        :features="generalFeatures"
        :config="generalConfig"
        @update:model="onGeneralModelChanged"
      />
    </workload-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { WorkloadFormWrapper } from "@/components/workload/workload-form-wrapper";
import {
  WorkloadSetupComputeSection,
  WorkloadSetupGeneralSection,
  type IWorkloadSetupModel,
  type IComputeSectionModel,
  type IComputeSectionConfig,
  type IGeneralSectionModel,
  type IGeneralSectionFeatures,
  type IGeneralSectionConfig,
} from "./workload-setup-sections";

// models
import type {
  IInferenceSubmitFormClusterConfig,
  IInferenceSubmitFormFeatures,
  IDistributedInferenceSubmitFormPolicy,
} from "@/components/workload-creation/inference/inference-submit-form/inference-submit-form.types";
import type { IWorkloadPriorityData } from "@/components/section/general-section/general-section.model";
import type { Preemptibility } from "@/swagger-models/workloads-client";
import type { Project } from "@/swagger-models/org-unit-service-client";

// utils
import { workloadService } from "@/services/cluster/workload.service/workload.service";

export interface IWorkloadSetupUIState {
  submitting: boolean;
  isDrawerOpen: boolean;
}

export enum EEmitName {
  UpdateModel = "update:model",
  LeavePage = "leave-page",
  Back = "back",
  Submit = "submit",
  CloseDrawer = "close-drawer",
}

export default defineComponent({
  name: "inference-distributed-workload-setup",
  components: {
    WorkloadFormWrapper,
    WorkloadSetupComputeSection,
    WorkloadSetupGeneralSection,
  },
  emits: {
    [EEmitName.UpdateModel]: (model: IWorkloadSetupModel) => model,
    [EEmitName.LeavePage]: () => true,
    [EEmitName.Back]: () => true,
    [EEmitName.Submit]: () => true,
    [EEmitName.CloseDrawer]: () => true,
  },
  props: {
    model: {
      type: Object as PropType<IWorkloadSetupModel>,
      required: true,
    },
    uiState: {
      type: Object as PropType<IWorkloadSetupUIState>,
      required: true,
    },
    features: {
      type: Object as PropType<IInferenceSubmitFormFeatures>,
      required: true,
    },
    policy: {
      type: Object as PropType<IDistributedInferenceSubmitFormPolicy | undefined>,
      required: false,
    },
    clusterConfig: {
      type: Object as PropType<IInferenceSubmitFormClusterConfig>,
      required: true,
    },
    selectedProject: {
      type: Object as PropType<Project>,
      required: false,
    },
    triggerValidationOnStartUp: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      EEmitName,
      submitBtnText: "Create inference",
      isGreedy: false,
    };
  },
  mounted() {
    if (this.triggerValidationOnStartUp) {
      this.isGreedy = true;
      this.$nextTick(async () => {
        const formWrapperRef = this.$refs.formWrapperRef as typeof WorkloadFormWrapper;
        await formWrapperRef?.validate();
        this.isGreedy = false;
      });
    }
  },
  computed: {
    // Compute section props
    computeModel(): IComputeSectionModel {
      return {
        replicas: this.model.replicas,
        leaderCompute: this.model.leaderCompute,
        workerCompute: this.model.workerCompute,
        workers: this.model.workers,
      };
    },
    computeConfig(): IComputeSectionConfig {
      return {
        replicasPolicyRules: this.policy?.distributedPolicy.worker?.rules?.replicas,
        nodePoolsResources: this.clusterConfig.projectNodePoolsResources,
        nodePoolsPolicy: {
          rules: this.policy?.distributedPolicy.worker?.rules?.nodePools || undefined,
          defaults: this.policy?.distributedPolicy.worker?.defaults?.nodePools || undefined,
        },
      };
    },

    // General section props
    generalModel(): IGeneralSectionModel {
      return {
        priorityClass: this.model.priorityClass,
        preemptibility: this.model.preemptibility,
        startupPolicy: this.model.startupPolicy,
      };
    },
    generalFeatures(): IGeneralSectionFeatures {
      return {
        supportPriorityClass: this.features.supportPriorityClass,
        supportPreemptibility: this.features.supportPreemptibility,
      };
    },
    generalConfig(): IGeneralSectionConfig {
      return {
        workloadPriorityData: this.workloadPriorityData,
        showQuotaWarning: this.shouldShowQuotaWarning,
        policyRules: this.policy?.distributedPolicy.worker?.rules,
      };
    },

    // Shared computed
    totalGpuRequest(): number {
      const leaderGpus = this.model.leaderCompute?.gpuDevicesRequest || 0;
      const workerGpus = this.model.workerCompute?.gpuDevicesRequest || 0;
      return leaderGpus + workerGpus * this.model.workers;
    },
    workloadPriorityData(): IWorkloadPriorityData {
      return {
        currProject: this.selectedProject,
        selectedNodePoolsResources: this.model.nodePools,
        totalGpuRequest: this.totalGpuRequest,
        workloadPriorities: this.clusterConfig.workloadPriorities,
        supportPriorityClass: this.features.supportPriorityClass,
        supportSemiPreemptibility: this.features.supportSemiPreemptibility,
      };
    },
    shouldShowQuotaWarning(): boolean {
      if (!this.features.supportPriorityClass) return false;

      if (this.features.supportPreemptibility) {
        const preemptibility = this.model.preemptibility;
        if (!preemptibility) return false;

        return workloadService.exceedsAvailableQuota({
          preemptibility: preemptibility as Preemptibility,
          currProject: this.selectedProject,
          selectedNodePoolsResources: this.workloadPriorityData.selectedNodePoolsResources,
          totalGpuRequest: this.workloadPriorityData.totalGpuRequest,
        });
      }

      return false;
    },
  },
  methods: {
    emitModelUpdate(updates: Partial<IWorkloadSetupModel>): void {
      this.$emit(EEmitName.UpdateModel, { ...this.model, ...updates });
    },

    onComputeModelChanged(computeModel: IComputeSectionModel): void {
      this.emitModelUpdate({
        replicas: computeModel.replicas || 1,
      });
    },

    onGeneralModelChanged(generalModel: IGeneralSectionModel): void {
      this.emitModelUpdate({
        priorityClass: generalModel.priorityClass,
        preemptibility: generalModel.preemptibility,
        startupPolicy: generalModel.startupPolicy,
      });
    },
  },
});
</script>

<style lang="scss" scoped>
.inference-distributed-workload-setup {
  width: 100%;
}
</style>
