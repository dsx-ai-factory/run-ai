<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="row items-center justify-start workload-page-actions">
    <!--Run-->
    <runai-tooltip-wrapper
      v-if="!isDeletedViewType"
      :display-tooltip="isRunBtnDisabled"
      :tooltip-text="runBtnTooltip"
      max-width="300px"
      :offset="actionTooltipOffset"
    >
      <runai-action-button
        btn-action="run"
        aid="activate-workload-btn"
        @click="emitAction(EWorkloadAction.run)"
        :disable="isRunBtnDisabled"
      />
    </runai-tooltip-wrapper>

    <!--Stop-->
    <runai-tooltip-wrapper
      v-if="!isDeletedViewType"
      :display-tooltip="isStopBtnDisabled"
      :tooltip-text="stopBtnTooltip"
      max-width="300px"
      :offset="actionTooltipOffset"
    >
      <runai-action-button
        btn-action="stop"
        aid="stop-workload-btn"
        @click="emitAction(EWorkloadAction.stop)"
        :disable="isStopBtnDisabled"
      />
    </runai-tooltip-wrapper>

    <!--Connect-->
    <runai-tooltip-wrapper
      v-if="!isDeletedViewType"
      :display-tooltip="isConnectBtnDisabled"
      :tooltip-text="connectBtnTooltip"
      max-width="300px"
      :offset="actionTooltipOffset"
    >
      <runai-action-button
        btn-action="connect"
        :disable="isConnectBtnDisabled"
        aid="connect-workload-btn"
        @click="handleConnectClick"
      >
        <tools-select v-if="hasMultipleTools" :tool-items="toolItems" @value-changed="openTool" />
      </runai-action-button>
    </runai-tooltip-wrapper>

    <!--Copy-->
    <runai-tooltip-wrapper
      v-if="!isDeletedViewType"
      :display-tooltip="isCopyWorkloadDisabled"
      :tooltip-text="copyBtnTooltip"
      max-width="350px"
      :offset="actionTooltipOffset"
    >
      <runai-action-button
        btn-action="duplicate"
        aid="copy-workload-btn"
        @click="emitAction(EWorkloadAction.clone)"
        :disable="isCopyWorkloadDisabled"
      />
    </runai-tooltip-wrapper>

    <!--Update-->
    <runai-tooltip-wrapper
      v-if="!isDeletedViewType"
      :display-tooltip="isUpdateWorkloadDisabled"
      :tooltip-text="updateBtnTooltip"
      max-width="350px"
      :offset="actionTooltipOffset"
    >
      <runai-action-button
        btn-action="update"
        aid="update-workload-btn"
        @click="emitAction(EWorkloadAction.update)"
        :disable="isUpdateWorkloadDisabled"
      />
    </runai-tooltip-wrapper>

    <!--Delete-->
    <runai-tooltip-wrapper
      v-if="!isDeletedViewType"
      :display-tooltip="isDeleteBtnDisabled"
      :tooltip-text="deleteBtnTooltip"
      max-width="300px"
      :offset="actionTooltipOffset"
    >
      <runai-action-button
        :disable="isDeleteBtnDisabled"
        btn-action="delete"
        aid="delete-workload-btn"
        @click="emitAction(EWorkloadAction.delete)"
      />
    </runai-tooltip-wrapper>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { ToolsSelect } from "@/components/tools/tools-select";
//model
import { Phase, type Workload, type Connection, Source, SourceApi } from "@/swagger-models/workloads-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { EWorkloadAction, EWorkloadEntity, EWorkloadErrorMessage, EWorkloadType } from "@/models/workload.model";
import type { IToolItem } from "@/models/workload.model";
import { InternalConnectionType } from "@/swagger-models/assets-service-client";
import { EClusterErrorMessage } from "@/models/cluster.model";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
//util
import { workloadUtil } from "@/utils/workload.util";
import { openNewWindow } from "@/utils/common.util";

//store
import { usePermissionStore } from "@/stores/permissions.store";
import { useAuthStore } from "@/stores/auth.store";
import { useSettingStore } from "@/stores/setting.store";
import { useClusterStore } from "@/stores/cluster.store";
import {
  MIN_CLUSTER_VERSION_FOR_MULTIPLE_WORKLOAD_ACTIONS,
  MIN_CLUSTER_VERSION_FOR_UPDATE_INFERENCE,
} from "@/common/version.constant";

const multipleWorkloadsSelectedMessage =
  "This action isn't supported when multiple workloads are selected. To proceed, select only one workload.";

export default defineComponent({
  name: "workload-page-actions",
  components: { RunaiTooltipWrapper, RunaiActionButton, ToolsSelect },
  emits: ["action"],
  props: {
    selectedWorkloads: {
      type: Array as PropType<Array<Workload>>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isClusterConnected: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isDeletedViewType: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    selectedWorkloadsAmount: {
      type: Number as PropType<number>,
      required: true,
    },
    workload: {
      type: Object as PropType<Workload>,
      required: true,
    },
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      authStore: useAuthStore(),
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),
      actionTooltipOffset: [-20, 0] as number[],
    };
  },
  computed: {
    Action() {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    EWorkloadAction(): typeof EWorkloadAction {
      return EWorkloadAction;
    },
    isRunBtnDisabled(): boolean {
      return !this.canRunWorkloads || !this.isClusterConnected;
    },
    isStopBtnDisabled(): boolean {
      return !this.canStopWorkloads || !this.isClusterConnected;
    },
    isConnectBtnDisabled(): boolean {
      return !this.canConnectWorkload || !this.isClusterConnected;
    },
    isDeleteBtnDisabled(): boolean {
      return !this.canDeleteWorkloads || !this.isClusterConnected;
    },
    isMultipleWorkloadsSelected(): boolean {
      return this.selectedWorkloadsAmount > 1;
    },
    isMultipleWorkloadsActionSupported(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workload.clusterId,
        MIN_CLUSTER_VERSION_FOR_MULTIPLE_WORKLOAD_ACTIONS,
      );
    },
    isCopyWorkloadDisabled(): boolean {
      if (this.isMultipleWorkloadsSelected) return true;
      return (
        !this.isRunaiWorkload ||
        !this.isCopySupportedWorkloadType ||
        !this.isClusterConnected ||
        !this.canCreateWorkload ||
        !this.isSelectedWorkloadProjectExists ||
        this.requiresFlexibleSubmissionEnabled
      );
    },
    isUpdateWorkloadDisabled(): boolean {
      if (this.isMultipleWorkloadsSelected) return true;
      return (
        this.isSubmittedByWorkloadV2 ||
        !this.isRunaiWorkload ||
        !this.isClusterConnected ||
        !this.canCreateWorkload ||
        !this.isSelectedWorkloadProjectExists ||
        !this.isInferenceWorkload ||
        !this.isFlatWorkload ||
        !this.isWorkloadPhaseSupportingUpdate ||
        this.isCreatedFromCli
      );
    },
    isInferenceWorkload(): boolean {
      return this.workload.type === EWorkloadType.Inference;
    },
    isDistributedInferenceWorkload(): boolean {
      return this.workload.type === EWorkloadType.DistributedInference;
    },
    isCopySupportedWorkloadType(): boolean {
      return [
        EWorkloadType.Workspace,
        EWorkloadType.Training,
        EWorkloadType.Distributed,
        EWorkloadType.Inference,
        EWorkloadType.DistributedInference,
      ].includes(this.workload.type as EWorkloadType);
    },
    canRunWorkloads(): boolean {
      if (this.isMultipleWorkloadsSelected && !this.isMultipleWorkloadsActionSupported) return false;
      if (!this.canCreateWorkload) return false;
      if (this.loading) return false;

      return this.selectedRunnableWorkloads.length > 0;
    },
    runBtnTooltip(): string {
      if (this.isMultipleWorkloadsSelected && !this.isMultipleWorkloadsActionSupported) {
        return multipleWorkloadsSelectedMessage;
      }

      if (!this.isClusterConnected) {
        return EClusterErrorMessage.NotConnected;
      }

      if (!this.isSelectedWorkloadProjectExists) {
        return `A workload whose project no longer exists cannot be activated.`;
      }

      if (!this.workload.actionsSupport?.suspend) {
        return EWorkloadErrorMessage.RunActionNotAuthorized;
      }

      if (this.workload.phase === Phase.Failed) {
        return EWorkloadErrorMessage.WorkloadFailedCantBeRun;
      }

      return this.getTooltipByAction(EWorkloadAction.activate);
    },
    stopBtnTooltip(): string {
      if (this.isMultipleWorkloadsSelected && !this.isMultipleWorkloadsActionSupported) {
        return multipleWorkloadsSelectedMessage;
      }

      if (!this.isClusterConnected) {
        return EClusterErrorMessage.NotConnected;
      }

      if (!this.isSelectedWorkloadProjectExists) {
        return `A workload whose project no longer exists cannot be stopped.`;
      }

      if (!this.workload.actionsSupport?.suspend) {
        return EWorkloadErrorMessage.StopActionNotAuthorized;
      }

      if (this.isInferenceWorkload || this.isDistributedInferenceWorkload || this.isSubmittedByWorkloadV2) {
        return EWorkloadErrorMessage.ActionNotSupported;
      }

      return this.getTooltipByAction(EWorkloadAction.stop);
    },
    connectBtnTooltip(): string {
      if (this.isMultipleWorkloadsSelected) {
        return multipleWorkloadsSelectedMessage;
      }

      const selectedWorkload = this.workload;

      if (this.isDistributedInferenceWorkload || this.isSubmittedByWorkloadV2) {
        return EWorkloadErrorMessage.ActionNotSupported;
      }

      if (selectedWorkload.phase !== Phase.Running) {
        return this.getTooltipByAction(EWorkloadAction.connect);
      }

      if (this.connectionsWithoutServingPort.length === 0 || !selectedWorkload.externalConnections?.length) {
        return `There are no tools for the selected workload to connect to`;
      }
      return this.getTooltipByAction(EWorkloadAction.connect);
    },
    canStopWorkloads(): boolean {
      if (this.isMultipleWorkloadsSelected && !this.isMultipleWorkloadsActionSupported) return false;
      if (!this.canCreateWorkload) return false;
      if (this.loading) return false;

      // If some of the selected workloads are stoppable then we enable the stop button
      return this.selectedStoppableWorkloads.length > 0;
    },
    isSelectedWorkloadProjectExists(): boolean {
      if (this.isMultipleWorkloadsSelected) return false;

      const selectedWorkload = this.workload;
      return !!this.authStore.orgUnits.find(
        (orgUnit: OrgUnit) => orgUnit.type === ResourceType.Project && orgUnit.id === selectedWorkload.projectId,
      );
    },
    connectionsWithoutServingPort(): Array<Connection> {
      const selectedWorkload = this.workload;
      if (!selectedWorkload?.environments) return [];
      return (
        selectedWorkload?.environments[0].connections?.filter(
          (connection: Connection) => connection.name !== InternalConnectionType.ServingPort,
        ) || []
      );
    },
    externalConnections(): Connection[] {
      const selectedWorkload = this.workload;
      return selectedWorkload.externalConnections || [];
    },
    canConnectWorkload(): boolean {
      if (this.isMultipleWorkloadsSelected) return false;

      const selectedWorkload = this.workload;

      if (
        this.isWorkloadPhaseDeleting ||
        selectedWorkload.phase !== Phase.Running ||
        !this.canReadOrCreateWorkload(selectedWorkload)
      ) {
        return false;
      }
      if (this.isDistributedInferenceWorkload) {
        return false;
      }

      if (selectedWorkload.type === EWorkloadType.Inference && selectedWorkload.environments) {
        const connections: IToolItem[] = workloadUtil.convertConnectionsToToolItems(this.connectionsWithoutServingPort);
        return connections.length > 0;
      }

      const urls: string[] | undefined = selectedWorkload.urls;
      return !!urls;
    },
    canDeleteWorkloads(): boolean {
      return this.selectedWorkloads.some((workload: Workload) => {
        if (workload.type === EWorkloadType.Inference) {
          return this.canCreateInference;
        }

        return (
          this.permissionStore.hasPermission(ResourceType.Workloads, Action.Delete) && !!workload.actionsSupport?.delete
        );
      });
    },
    canCreateWorkload(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Workloads, Action.Create);
    },
    canCreateInference(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Inferences, Action.Create);
    },
    canReadWorkload(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Workloads, Action.Read);
    },
    canReadInference(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Inferences, Action.Read);
    },
    isRunaiWorkload(): boolean {
      return workloadUtil.isRunaiWorkload(this.workload);
    },
    isSubmittedByWorkloadV2(): boolean {
      return workloadUtil.isWorkloadSubmittedByV2(this.workload);
    },
    deleteBtnTooltip(): string {
      if (!this.isClusterConnected) {
        return EClusterErrorMessage.NotConnected;
      }
      if (this.isMultipleWorkloadsSelected) {
        return `The selected workloads can't be deleted. You're either not authorized to delete them or need to go to the application in which they were created to delete them from there.`;
      }
      if (this.isWorkloadPhaseDeleting) {
        return EWorkloadErrorMessage.WorkloadIsDeleting;
      }
      if (!this.workload.actionsSupport?.delete) {
        return EWorkloadErrorMessage.DeleteWorkloadIsNotSupported;
      }
      if (!this.canDeleteWorkloads) {
        return EWorkloadErrorMessage.NoDeletePermission;
      }
      return "";
    },
    toolItems(): IToolItem[] {
      const allConnections = [...this.connectionsWithoutServingPort, ...this.externalConnections];
      return workloadUtil.convertConnectionsToToolItems(allConnections);
    },
    hasMultipleTools(): boolean {
      return this.toolItems.length > 1;
    },
    isWorkloadPhaseDeleting(): boolean {
      return this.workload?.phase === Phase.Deleting;
    },
    isWorkloadPhaseSupportingUpdate(): boolean {
      if (!this.workload) return false;

      const phasesNotSupporting = new Set<Phase>([Phase.Creating, Phase.Deleting, Phase.Failed, Phase.Terminating]);
      return !phasesNotSupporting.has(this.workload.phase);
    },
    isCreatedFromCli(): boolean {
      if (!this.workload) return false;
      return this.workload.source === Source.Cli;
    },
    copyBtnTooltip(): string {
      if (this.isMultipleWorkloadsSelected) {
        return multipleWorkloadsSelectedMessage;
      }

      if (!this.isClusterConnected) {
        return EClusterErrorMessage.NotConnected;
      }

      if (!this.isSelectedWorkloadProjectExists) {
        return EWorkloadErrorMessage.ProjectNoLongerExists;
      }

      if (this.requiresFlexibleSubmissionEnabled) {
        return EWorkloadErrorMessage.RequiresFlexibleSubmissionForm;
      }

      if (!this.isRunaiWorkload || !this.isCopySupportedWorkloadType) {
        return EWorkloadErrorMessage.ActionNotSupported;
      }

      return this.getTooltipByAction(EWorkloadAction.clone);
    },
    updateBtnTooltip(): string {
      if (this.isMultipleWorkloadsSelected) {
        return multipleWorkloadsSelectedMessage;
      }

      if (!this.isClusterConnected) {
        return EClusterErrorMessage.NotConnected;
      }

      if (!this.isSelectedWorkloadProjectExists) {
        return EWorkloadErrorMessage.ProjectNoLongerExists;
      }

      if (!this.isInferenceWorkload || !this.isFlatWorkload || !this.isClusterSupported) {
        return EWorkloadErrorMessage.WorkloadsCreatedUsingTheFlexibleAndSpecificCluster;
      }

      if (this.workload.phase === Phase.Creating) {
        return EWorkloadErrorMessage.WorkloadCreatingCantBeUpdated;
      }

      if (this.workload.phase === Phase.Deleting) {
        return EWorkloadErrorMessage.WorkloadDeletingCantBeUpdated;
      }

      if (this.workload.phase === Phase.Terminating) {
        return EWorkloadErrorMessage.WorkloadTerminatingCantBeUpdated;
      }

      if (this.workload.phase === Phase.Failed) {
        return EWorkloadErrorMessage.WorkloadFailedCantBeUpdated;
      }

      if (this.isCreatedFromCli) {
        return EWorkloadErrorMessage.CliWorkloadsCantBeUpdated;
      }

      return this.getTooltipByAction(EWorkloadAction.update);
    },
    isClusterSupported(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.workload.clusterId,
        MIN_CLUSTER_VERSION_FOR_UPDATE_INFERENCE,
      );
    },
    isFlatWorkload(): boolean {
      return !this.workload?.environments?.[0]?.id;
    },
    requiresFlexibleSubmissionEnabled(): boolean {
      if (this.isFlatWorkload && this.workload.type === EWorkloadType.Inference) return false;
      return (
        this.isFlatWorkload &&
        !this.settingStore.isHybridSubmissionEnabled &&
        this.workload.source !== Source.Cli &&
        this.workload.sourceApi !== SourceApi.WorkloadsV2
      );
    },
    selectedStoppableWorkloads(): Array<Workload> {
      return this.selectedWorkloads.filter(
        (workload: Workload) =>
          this.isWorkloadStoppable(workload) &&
          this.isWorkloadSupportSuspend(workload) &&
          this.isWorkloadProjectExists(workload),
      );
    },
    selectedRunnableWorkloads(): Array<Workload> {
      return this.selectedWorkloads.filter(
        (workload: Workload) =>
          workload.phase === Phase.Stopped &&
          this.isWorkloadSupportSuspend(workload) &&
          this.isWorkloadProjectExists(workload),
      );
    },
  },
  methods: {
    isWorkloadStoppable(workload: Workload): boolean {
      const nonStoppablePhases: Phase[] = [
        Phase.Creating,
        Phase.Stopping,
        Phase.Deleting,
        Phase.Stopped,
        Phase.Failed,
        Phase.Completed,
        Phase.Terminating,
      ];
      return !nonStoppablePhases.includes(workload?.phase);
    },
    isWorkloadSupportSuspend(workload: Workload): boolean {
      return !!workload.actionsSupport?.suspend;
    },
    isWorkloadProjectExists(workload: Workload): boolean {
      return !!this.authStore.orgUnits.find(
        (orgUnit: OrgUnit) => orgUnit.type === ResourceType.Project && orgUnit.id === workload.projectId,
      );
    },
    canReadOrCreateWorkload(selectedWorkload: Workload): boolean {
      if (selectedWorkload.type === EWorkloadType.Inference) {
        return this.canCreateInference || this.canReadInference;
      } else {
        return this.canReadWorkload || this.canCreateWorkload;
      }
    },
    openTool(url: string): void {
      openNewWindow(url);
    },
    emitAction(action: EWorkloadAction): void {
      this.$emit("action", action);
    },
    getTooltipByAction(action: EWorkloadAction): string {
      const selectedWorkload = this.workload;
      if (action === EWorkloadAction.delete && !selectedWorkload.actionsSupport?.delete) {
        return EWorkloadErrorMessage.NotRunaiWorkloadMessage;
      }

      if (
        (action === EWorkloadAction.stop || action === EWorkloadAction.run) &&
        !selectedWorkload.actionsSupport?.suspend
      ) {
        return EWorkloadErrorMessage.NotRunaiWorkloadMessage;
      }

      if (action === EWorkloadAction.activate) {
        return workloadUtil.getDisabledToolTipText(EWorkloadAction.run, EWorkloadEntity.workload);
      }

      return workloadUtil.getDisabledToolTipText(action, EWorkloadEntity.workload);
    },
    handleConnectClick() {
      if (this.toolItems.length === 1) {
        this.openTool(this.toolItems[0].url);
        return;
      }
      this.openFirstConnectionIfOldWorkloadModel();
    },
    openFirstConnectionIfOldWorkloadModel() {
      const selectedWorkload = this.workload;

      // Old workload models do not provide environment connections info, therefore in order to connect to its tool it needs to be taken from the URLs property
      const isOldWorkloadModel =
        selectedWorkload.urls?.length && !selectedWorkload.environments && !selectedWorkload.externalConnections;
      if (!isOldWorkloadModel) return;
      // if the cluster is 2.9, URLs are arriving as a single string seperated by a comma, we need to take the first
      const urlString = selectedWorkload?.urls?.[0];
      const firstUrl = urlString?.split(", ")[0];
      !!firstUrl && openNewWindow(firstUrl);
    },
  },
});
</script>

<style scoped lang="scss"></style>
