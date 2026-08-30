<template>
  <q-btn
    :disable="isButtonDisabled"
    color="primary"
    class="btn-icon-sm runai-icon-btn--default"
    icon="ph-bold ph-plus"
    aid="primary-cta-btn"
    label="NEW WORKLOAD"
    :flat="flat"
  >
    <q-menu fit :persistent="isGuidedTourActive">
      <q-list>
        <q-item
          v-close-popup
          v-for="option in newWorkloadOptions"
          :clickable="!option.disable"
          :key="option.label"
          :class="{ 'yaml-option-separator': option.value === 'Via YAML' }"
          @click="onOptionClick(option)"
          :aid="option.label + '-btn'"
          :disable="option.disable"
        >
          <q-item-section class="min-w-0 col-grow">
            {{ option.label }}
            <q-tooltip v-if="option.disable && option.disabledTooltip" class="q-pa-sm" max-width="300px">
              {{ option.disabledTooltip }}
            </q-tooltip>
          </q-item-section>
          <q-item-section v-if="option.toolTip" side class="items-center">
            <runai-tooltip :tooltip-text="option.toolTip" />
          </q-item-section>
        </q-item>
      </q-list>
    </q-menu>
    <!--  when the entire button is disabled, show the tooltip -->
    <q-tooltip v-if="isButtonDisabled && tooltipText">
      {{ tooltipText }}
    </q-tooltip>
  </q-btn>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// routes
import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";
import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

// store
import { usePermissionStore } from "@/stores/permissions.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAppStore } from "@/stores/app.store";

// models
import type { ISelectOption } from "@/models/global.model";
import { EClusterErrorMessage } from "@/models/cluster.model";
import { EWorkloadErrorMessage, EWorkloadTypeTooltip } from "@/models/workload.model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

// components
import { RunaiTooltip } from "@/components/common/runai-tooltip";

// utils
import { clusterUtil } from "@/utils/cluster.util";
import { useGuidedTourStore } from "@/stores/guided-tour.store";

export enum ENewWorkloadBtnOptions {
  Training = "Training",
  Workspace = "Workspace",
  Inference = "Inference",
  ViaYAML = "Via YAML",
}
export default defineComponent({
  name: "create-workload-dropdown",
  components: {
    RunaiTooltip,
  },
  props: {
    flat: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    noExistingProjects: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isClusterConnected: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isYamlSubmissionSupported: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      clusterStore: useClusterStore(),
      appStore: useAppStore(),
      guidedTourStore: useGuidedTourStore(),
    };
  },
  computed: {
    isGuidedTourActive(): boolean {
      return this.guidedTourStore.isActive;
    },
    clusterId(): string {
      return this.clusterStore.selectedClusterId;
    },
    action(): typeof Action {
      return Action;
    },
    resourceType(): typeof ResourceType {
      return ResourceType;
    },
    newWorkloadOptions(): ISelectOption[] {
      const hasInferenceCreatePermission = this.permissionStore.hasPermission(ResourceType.Inferences, Action.Create);
      const hasWorkloadCreatePermission = this.permissionStore.hasPermission(ResourceType.Workloads, Action.Create);
      const workloadOptions: ISelectOption[] = [
        {
          label: ENewWorkloadBtnOptions.Workspace,
          value: ENewWorkloadBtnOptions.Workspace,
          disable: !hasWorkloadCreatePermission,
          disabledTooltip: EWorkloadErrorMessage.NoActionPermission,
          toolTip: EWorkloadTypeTooltip.Workspace,
        },
        {
          label: ENewWorkloadBtnOptions.Training,
          value: ENewWorkloadBtnOptions.Training,
          disable: !hasWorkloadCreatePermission,
          disabledTooltip: EWorkloadErrorMessage.NoActionPermission,
          toolTip: EWorkloadTypeTooltip.Training,
        },
        {
          label: ENewWorkloadBtnOptions.Inference,
          value: ENewWorkloadBtnOptions.Inference,
          disable: !hasInferenceCreatePermission || this.isInferenceBlockedByClusterDependencies,
          disabledTooltip: this.inferenceOptionDisabledTooltip,
          toolTip: EWorkloadTypeTooltip.Inference,
        },
      ];

      // Only show Via YAML option if cluster version supports it
      if (this.isYamlSubmissionSupported) {
        workloadOptions.push({
          label: ENewWorkloadBtnOptions.ViaYAML,
          value: ENewWorkloadBtnOptions.ViaYAML,
          disable: !hasWorkloadCreatePermission,
          disabledTooltip: EWorkloadErrorMessage.NoActionPermission,
          toolTip: "Submission of a range of workload types using a standard Kubernetes YAML",
        });
      }

      return workloadOptions;
    },
    canCreateSomeWorkload(): boolean {
      return this.permissionStore.hasSomePermittedActions(
        [ResourceType.Workloads, ResourceType.Inferences],
        [Action.Create],
      );
    },
    canCreateProject(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Project, Action.Create);
    },
    isButtonDisabled(): boolean {
      if (!this.canCreateSomeWorkload) {
        return true;
      }
      return this.noExistingProjects || !this.isClusterConnected;
    },
    isInferenceBlockedByClusterDependencies(): boolean {
      const cluster = this.clusterStore.getClusterById(this.clusterId);
      if (!cluster) {
        return false;
      }
      const isKnativeInstalled = clusterUtil.isKnativeInstalledInCluster(cluster);
      const isLwsInstalled = clusterUtil.isLwsInstalledInCluster(cluster);
      return !isKnativeInstalled && !isLwsInstalled;
    },
    inferenceOptionDisabledTooltip(): string | undefined {
      if (!this.permissionStore.hasPermission(ResourceType.Inferences, Action.Create)) {
        return EWorkloadErrorMessage.NoActionPermission;
      }
      if (this.isInferenceBlockedByClusterDependencies) {
        return EWorkloadErrorMessage.InferenceCantBeCreatedKnativeAndLwsMissing;
      }
      return undefined;
    },
    tooltipText(): string {
      if (!this.isClusterConnected) {
        return EClusterErrorMessage.NotConnected;
      } else if (this.canCreateProject && this.canCreateSomeWorkload) {
        return "To create a new workload, first create at least one project";
      } else if (this.canCreateSomeWorkload) {
        return "Workloads can only be created under a project. Contact your administrator to add a project.";
      }
      return "You don't have permission to perform this action.";
    },
  },
  methods: {
    onOptionClick(option: ISelectOption): void {
      switch (option.value) {
        case ENewWorkloadBtnOptions.Workspace:
          this.$router.push({ name: WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW, query: { clusterId: this.clusterId } });
          break;
        case ENewWorkloadBtnOptions.Training:
          this.$router.push({ name: TRAINING_ROUTE_NAMES.TRAINING_NEW, query: { clusterId: this.clusterId } });
          break;
        case ENewWorkloadBtnOptions.Inference:
          this.$router.push({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW, query: { clusterId: this.clusterId } });
          break;
        case ENewWorkloadBtnOptions.ViaYAML:
          this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_YAML_NEW, query: { clusterId: this.clusterId } });
          break;
      }
    },
  },
});
</script>

<style scoped lang="scss">
.yaml-option-separator {
  border-top: 1px solid $black-12;
}
</style>
