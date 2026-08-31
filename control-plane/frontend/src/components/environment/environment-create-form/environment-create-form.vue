<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="environment-create-form column items-center q-pt-md">
    <runai-form-wrapper class="environment-form" :form-state="environmentModel">
      <q-form ref="elEnvironmentForm">
        <runai-scope-section
          kind-of-asset="environment"
          :read-only="!isScopeSelectable || isEditMode"
          :scope-model="scopeModel"
          @changed-scope="updateScope"
          :allowed-scopes="allowedScopes"
          :key="selectedCluster"
          :selected-cluster-id="selectedCluster"
        />
        <runai-name-section
          label="Environment"
          asset-kind="environment"
          :original-name="originalAssetName"
          v-model="environmentModel.meta.name"
          :scope-model="scopeModel"
          support-description
          :description="environmentModel.meta.description"
          @update-description="updateDescription"
        />
        <transition-group name="fade">
          <template v-if="isScopeSelected">
            <form-image-section :image-model="imageModel" @update="updateImageModel" />
            <form-workload-types-section
              :read-only="isEditMode"
              :show-distributed-section="isDistributedEnabled"
              :workload-types-section-model="workloadTypesSectionModel"
              @update-workload-types-section-model="updateWorkloadTypes"
              :locked-workload-types="lockedWorkloadTypes"
              :cluster-id="parentClusterId || selectedCluster"
            />
            <form-endpoint-section
              v-if="showEndpointSection"
              :connection="endpointConnectionData"
              @update="updateEndpointConnectionData"
              :protocol-locked-to-http="isDistributedInference"
            />
            <form-tools-section v-if="!isDistributedInference" :tools="tools" @update="updateToolsModel" />
            <form-runtime-section
              :runtime-settings="runtimeSettingsModel"
              :environment-variables-source-value-options="environmentVariablesSourceValueOptions"
              :support-config-map-as-env-var="supportConfigMapAsEnvVar"
              @update="updateRuntimeSettingsModel"
            />
            <form-security-section :security-model="securityModel" :is-s-s-o="isSSO" @update="updateSecurityModel" />
          </template>
        </transition-group>

        <section class="row items-center q-mt-md">
          <q-field
            class="col-4 form-hint no-padding"
            :model-value="displayFormHint"
            :rules="[isFormIncomplete]"
          ></q-field>
          <div class="buttons q-ml-auto">
            <runai-btn
              class="q-mr-md"
              @click="onCancel"
              aid="cancel-environment-btn"
              flat
              color="primary"
              label="cancel"
            />
            <runai-btn
              color="primary"
              type="submit"
              @click="save"
              :loading="validating || submitting"
              aid="create-environment-btn"
              :label="`${isEditMode ? 'save' : 'create'} environment`"
            />
          </div>
        </section>
      </q-form>
    </runai-form-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiNameSection } from "@/components/common/runai-name-section";
import { FormImageSection } from "@/components/environment/form-image-section";
import { FormWorkloadTypesSection } from "@/components/environment/form-workload-types-section";
import { FormEndpointSection } from "@/components/environment/form-endpoint-section";
import { FormToolsSection } from "@/components/environment/form-tools-section";
import { FormRuntimeSection } from "@/components/environment/form-runtime-section";
import { FormSecuritySection } from "@/components/environment/form-security-section";
import { RunaiScopeSection } from "@/components/common/runai-scope-section";
import { RunaiBtn } from "@/components/common/runai-btn";
// Models
import type {
  TEnvVarNoneCustomSourceOption,
  IImage,
  IRuntimeSettings,
  ISecurityModel,
  IUIConnection,
  IUIEnvironmentAsset,
} from "@/models/environment.model";
import {
  ImagePullPolicy,
  InternalConnectionType,
  ServingPortInfoProtocolEnum,
  ToolType,
  Scope,
  type EnvironmentAssetSpec,
  type WorkloadSupportedTypes,
} from "@/swagger-models/assets-service-client";
import { type PermittedScopes } from "@/swagger-models/authorization-client";
import type { IScopeModel } from "@/models/global.model";
// Utils
import { deepCopy } from "@/utils/common.util";
import { scopeUtil } from "@/utils/scope.util";
// Stores
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAppStore } from "@/stores/app.store";

import { MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR } from "@/common/version.constant";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { assetsUtil } from "@/utils/assets.util/assets.util";
import { useEnvironmentVariablesSection } from "@/composables/use-environment-variables-section.composable";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiNameSection,
    FormImageSection,
    FormWorkloadTypesSection,
    FormToolsSection,
    FormRuntimeSection,
    FormSecuritySection,
    FormEndpointSection,
    RunaiScopeSection,
    RunaiBtn,
  },
  props: {
    initialEnvironment: {
      type: Object as PropType<IUIEnvironmentAsset>,
      required: true,
    },
    submitting: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isScopeSelectable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    lockedWorkloadTypes: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: false,
    },
    isEditMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    originalAssetName: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      clusterStore: useClusterStore(),
      appStore: useAppStore(),
      environmentModel: deepCopy(this.initialEnvironment) as IUIEnvironmentAsset,
      environmentForm: null as HTMLFormElement | null,
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
      selectedCluster: "" as string,
      environmentVariableSourceValueOptions: [] as Array<TEnvVarNoneCustomSourceOption>,
      validating: false as boolean,
    };
  },
  setup() {
    const { loadSourceValueOptions, validateEnvironmentVariablesOfAsset, environmentVariablesSourceValueOptions } =
      useEnvironmentVariablesSection();
    return {
      loadSourceValueOptions,
      validateEnvironmentVariablesOfAsset,
      environmentVariablesSourceValueOptions,
    };
  },
  async created() {
    await this.loadSourceValueOptions(
      false,
      this.environmentModel.meta.projectId ? String(this.environmentModel.meta.projectId) : undefined,
      this.environmentModel.meta.departmentId || undefined,
      this.environmentModel.meta.clusterId || undefined,
    );
    this.validateEnvironmentVariablesOfAsset(this.environmentModel.spec.environmentVariables);
  },
  mounted() {
    this.environmentForm = this.$refs.elEnvironmentForm as HTMLFormElement;
  },
  computed: {
    tools(): Array<IUIConnection> {
      return this.environmentModel.spec.connections.filter(
        (connection: IUIConnection) => connection.connectionType !== InternalConnectionType.ServingPort,
      );
    },
    parentClusterId(): string | null {
      if (!this.environmentModel.meta.scope || this.environmentModel.meta.scope === Scope.Tenant) return null;

      const scopeId = assetsUtil.getAssetScopeIdFromAssetMeta(this.environmentModel.meta);

      return orgTreeService.getParentClusterUuid(useAuthStore().orgUnitList, scopeId, this.environmentModel.meta.scope);
    },
    isDistributedEnabled(): boolean {
      return (
        this.clusterStore.isVersionSupportDistributedTraining(this.parentClusterId || this.selectedCluster) &&
        !this.appStore.isWhiteLabel
      );
    },
    isScopeSelected(): boolean {
      return !!this.scopeModel.scope;
    },
    isSSO(): boolean {
      return this.authStore.isSSO;
    },
    scopeModel(): IScopeModel {
      return {
        scope: this.environmentModel.meta.scope,
        projectId: this.environmentModel.meta.projectId,
        departmentId: this.environmentModel.meta.departmentId,
        clusterId: this.environmentModel.meta.clusterId,
      };
    },
    workloadTypesSectionModel(): WorkloadSupportedTypes {
      return this.environmentModel.meta.workloadSupportedTypes || {};
    },
    imageModel(): IImage {
      const { spec }: { spec: EnvironmentAssetSpec } = this.environmentModel;
      return {
        image: spec.image || null,
        imagePullPolicy: spec.imagePullPolicy || ImagePullPolicy.Always,
      };
    },
    securityModel(): ISecurityModel {
      return {
        uidGidSource: this.environmentModel.spec.uidGidSource,
        runAsUid: this.environmentModel.spec.runAsUid,
        runAsGid: this.environmentModel.spec.runAsGid,
        overrideUidGidInWorkspace: this.environmentModel.spec.overrideUidGidInWorkspace,
        supplementalGroups: this.environmentModel.spec.supplementalGroups,
        capabilities: this.environmentModel.spec.capabilities,
      };
    },
    runtimeSettingsModel(): IRuntimeSettings {
      const { spec }: { spec: EnvironmentAssetSpec } = this.environmentModel;
      return {
        command: spec.command || null,
        args: spec.args || null,
        assetEnvironmentVariables: spec.environmentVariables || [],
        workingDir: spec.workingDir || null,
      };
    },
    showEndpointSection(): boolean {
      return !!this.environmentModel.meta.workloadSupportedTypes?.inference;
    },
    isDistributedInference(): boolean {
      const w = this.environmentModel.meta.workloadSupportedTypes;
      return !!(w?.distributed && w?.inference && !w?.training);
    },
    connectionForTools(): Array<IUIConnection> {
      return this.environmentModel.spec.connections.filter(
        (connection: IUIConnection) => connection.connectionType === InternalConnectionType.ServingPort,
      );
    },
    endpointConnectionData(): IUIConnection {
      const servingPortConnection: IUIConnection = {
        id: InternalConnectionType.ServingPort, // since this section is not an array we dont need the id
        name: InternalConnectionType.ServingPort, // since this section is not an array we dont need the name
        toolType: ToolType.Custom,
        connectionType: InternalConnectionType.ServingPort,
        protocol: ServingPortInfoProtocolEnum.Http,
      };

      const connection: IUIConnection | undefined = this.environmentModel.spec.connections.find(
        (connection: IUIConnection) => connection.connectionType === InternalConnectionType.ServingPort,
      );

      return connection || servingPortConnection;
    },
    supportConfigMapAsEnvVar(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.parentClusterId || this.selectedCluster,
        MIN_CLUSTER_VERSION_FOR_CONFIG_MAP_ENV_VAR,
      );
    },
  },
  methods: {
    updateDescription(description: string): void {
      this.environmentModel.meta.description = description;
    },
    updateImageModel(imageModel: IImage): void {
      this.environmentModel.spec = { ...this.environmentModel.spec, ...imageModel };
    },
    updateNameModel(nameModel: string): void {
      this.environmentModel.meta.name = nameModel;
    },
    isDistributedInferenceSelected(workloadSupportedTypes: WorkloadSupportedTypes): boolean {
      return !!(
        workloadSupportedTypes.distributed &&
        workloadSupportedTypes.inference &&
        !workloadSupportedTypes.training
      );
    },
    setInferenceDefaultProtocol(): void {
      const servingIndex = this.environmentModel.spec.connections.findIndex(
        (c) => c.connectionType === InternalConnectionType.ServingPort,
      );
      if (
        servingIndex !== -1 &&
        this.environmentModel.spec.connections[servingIndex].protocol !== ServingPortInfoProtocolEnum.Http
      ) {
        this.environmentModel.spec.connections[servingIndex] = {
          ...this.environmentModel.spec.connections[servingIndex],
          protocol: ServingPortInfoProtocolEnum.Http,
        };
      }
    },
    updateWorkloadTypes(workloadSupportedTypes: WorkloadSupportedTypes): void {
      if (this.environmentModel.meta.workloadSupportedTypes?.inference !== workloadSupportedTypes.inference) {
        this.environmentModel.spec.connections = [];
      }

      this.environmentModel.meta.workloadSupportedTypes = workloadSupportedTypes;
      if (this.isDistributedInferenceSelected(workloadSupportedTypes)) {
        this.setInferenceDefaultProtocol();
      }
    },
    updateSecurityModel(securityModel: ISecurityModel): void {
      this.environmentModel.spec = { ...this.environmentModel.spec, ...securityModel };
    },
    updateToolsModel(toolsModel: Array<IUIConnection>): void {
      const servingPortConnection: IUIConnection | undefined = this.environmentModel.spec.connections.find(
        (connection: IUIConnection) => connection.connectionType === InternalConnectionType.ServingPort,
      );
      this.environmentModel.spec.connections = toolsModel;
      if (servingPortConnection) {
        this.environmentModel.spec.connections.push(servingPortConnection);
      }
    },
    updateRuntimeSettingsModel(runtimeSettings: IRuntimeSettings): void {
      this.environmentModel.spec = {
        ...this.environmentModel.spec,
        command: runtimeSettings.command,
        args: runtimeSettings.args,
        environmentVariables: runtimeSettings.assetEnvironmentVariables || [],
        workingDir: runtimeSettings.workingDir,
      };
    },
    updateScope(updatedScope: IScopeModel): void {
      this.environmentModel.meta = scopeUtil.updateScope(this.environmentModel.meta, updatedScope);
      this.loadSourceValueOptions(
        false,
        this.environmentModel.meta.projectId ? String(this.environmentModel.meta.projectId) : undefined,
        this.environmentModel.meta.departmentId || undefined,
        this.environmentModel.meta.clusterId || undefined,
      );
      this.validateEnvironmentVariablesOfAsset(this.environmentModel.spec.environmentVariables);
    },
    updateEndpointConnectionData(endpointConnection: IUIConnection): void {
      const endpointIndex = this.environmentModel.spec.connections.findIndex((c) => c.id === endpointConnection.id);
      if (endpointIndex !== -1) {
        this.environmentModel.spec.connections[endpointIndex] = endpointConnection;
      } else {
        this.environmentModel.spec.connections.push(endpointConnection);
      }
    },
    onCancel(): void {
      this.$emit("cancel");
    },
    async save(): Promise<void> {
      try {
        this.validating = true;
        this.displayFormHint = false;
        const success: boolean = await (this.environmentForm as HTMLFormElement).validate();
        if (!success) {
          this.showHint();
          return;
        }

        this.$emit("submit", this.environmentModel);
      } catch (error) {
        console.error("Error while validating environment form", error);
      } finally {
        this.validating = false;
      }
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : "Please review and fix the issues in the form";
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});
</script>
