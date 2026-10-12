<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <hybrid-section-wrapper
    default-opened
    :prevent-top-padding-when-open="preventTopPaddingWhenOpen"
    aid="hybrid-environment-section"
    label="Environment"
    :selected-name="selectedName"
    :support-linking="supportLinking"
    :section-disabled="sectionDisabled"
    :show-content-header="showContentHeader"
    :is-linked="isLinked"
    @load-section-templates="$emit('load-section-templates')"
    @clear="$emit('clear')"
    @reset="$emit('reset')"
    @toggle-linked="$emit('toggle-linked')"
  >
    <template #subheader>
      <span>{{ environmentSummary }}</span>
    </template>

    <form-image-origin-section
      class="q-mb-md form-image-input-section"
      :model-value="workloadSpec.image || null"
      :policy-rules="policy?.rules?.image"
      :disable="sectionDisabled || isLinked"
      :is-required="requiredFields?.image"
      :tooltip-text="isLinked ? linkedDisableText : undefined"
      :support-authenticated-ngc="supportAuthenticatedNgc"
      :authenticated-only="authenticatedNgcOnly"
      :scope-id="scopeId"
      :scope-type="scopeType"
      :cluster-id="clusterId"
      @update:model-value="updateModel('image', $event)"
      @credential-changed="onCredentialChanged"
    />

    <serving-endpoint-section
      v-if="supportingServingPort && servingEndpointModel"
      title="Set an inference serving endpoint"
      title-tooltip="Define how applications or services should communicate inference requests to the workload."
      :model-value="servingEndpointModel"
      @update:model-value="onServingPortChanged"
      show-protocol-and-port
      :disabled="!!disableFields?.servingPort"
      :disabled-info="servingPortDisabledInfo"
      :policy-rules="policy?.rules?.servingPort"
      :is-port-required="requiredFields?.port"
      :port-disabled="isLinked"
      :protocol-disabled="isLinked || !!disableFields?.protocol"
      :cluster-id="clusterId"
      :is-distributed-inference="isDistributedInference"
    />

    <runai-folding-sub-section label="Image pull" aid="image-pull-policy-sub-section">
      <form-image-pull-policy
        :disable="sectionDisabled || isLinked"
        :image-pull-policy="workloadSpec.imagePullPolicy"
        :policy-rules="policy?.rules?.imagePullPolicy"
        :support-image-pull-secrets="supportImagePullSecrets"
        :image-pull-secrets="workloadSpec.imagePullSecrets || null"
        :image-pull-secrets-rules="policy?.rules?.imagePullSecrets"
        :support-ngc-api-key="supportingNgcApiKey"
        :tooltip-text="formImagePullPolicyTooltipText"
        :cluster-id="clusterId"
        @update="updateModel('imagePullPolicy', $event)"
        @update-image-pull-secrets="updateImagePullSecrets"
        @credential-created="$emit('credential-created', $event)"
      />
    </runai-folding-sub-section>

    <runai-folding-sub-section v-if="!hideSpecificSections?.tools" label="Tools" aid="tools-section">
      <form-tools-list
        :section-disabled="sectionDisabled || !!disableFields?.tools || isLinked"
        :tools="workloadSpec.connections"
        @update="updateTools"
        :cluster-id="clusterId"
        :access-disabled="!!disableFields?.toolsAccess"
        :tools-disabled-info="disableFields?.tools"
        :tooltip-text="isLinked ? linkedDisableText : undefined"
        :policy-rules="policy?.rules?.exposedUrls"
        :policy-defaults="policy?.defaults?.exposedUrls"
      />
    </runai-folding-sub-section>

    <runai-folding-sub-section label="Runtime settings" aid="runtime-settings-sub-section">
      <form-runtime-settings
        :disable="sectionDisabled || isLinked"
        :form-type="EEnvVarFormType.Hybrid"
        :render-key="renderKey"
        :runtime-settings="runtimeSettings"
        :environment-variable-source-value-options="environmentVariableSourceValueOptions"
        :policy="policy"
        :support-config-map-as-env-var="supportConfigMapAsEnvVar"
        :support-user-credentials="supportUserCredentials"
        :tooltip-text="isLinked ? linkedDisableText : undefined"
        :validate-empty-env-vars="validateEmptyEnvVars"
        :supporting-ngc-api-key="supportingNgcApiKey"
        @update="updateRuntimeSettings"
        @credential-created="$emit('credential-created', $event)"
      />
    </runai-folding-sub-section>

    <runai-folding-sub-section label="Security" aid="security-sub-section">
      <form-security-settings
        :disable="sectionDisabled || !!disableFields?.security || isLinked"
        :security-model="securityModel"
        :is-s-s-o="authStore.isSSO"
        :policy-rules="policy?.rules"
        :disable-capabilities-info="disableFields?.security?.capabilities"
        :disable-uid-gid-source-info="disableFields?.security?.uidGidSource"
        :tooltip-text="isLinked ? linkedDisableText : undefined"
        @update="updateSecuritySettings"
      />
    </runai-folding-sub-section>
  </hybrid-section-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Stores
import { useAuthStore } from "@/stores/auth.store";
// Cmps
import { HybridSectionWrapper } from "../hybrid-section-wrapper";
import { RunaiFoldingSubSection } from "@/components/common/runai-folding-sub-section";
import { FormImagePullPolicy } from "@/components/environment/form-image-pull-policy";
import { FormRuntimeSettings } from "@/components/environment/form-runtime-section/form-runtime-settings";
import { FormSecuritySettings } from "@/components/environment/form-security-section/form-security-settings";
import { FormToolsList } from "@/components/environment/form-tools-section/form-tools-list";
import { ServingEndpointSection } from "@/components/section/inference/serving-endpoint-section";
import { FormImageOriginSection } from "@/components/environment/form-image-origin-section";

// Models
import type { IUISuperWorkloadSpec } from "@/models/workload.model";
import {
  EEnvVarFormType,
  type TEnvVarNoneCustomSourceOption,
  type IRuntimeSettings,
  type ISecurityModel,
  type IUIConnection,
} from "@/models/environment.model";
import type { UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";
import type { IServingEndpointAccessData } from "@/components/section/inference/serving-endpoint-section/serving-endpoint-access.model";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import type { IHybridEnvironmentSectionDisableFields } from "./hybrid-environment-section.model";
import type { IHideSpecificSectionsEnvironment } from "@/models/workload-hybrid.model";
import type { ImagePullSecret } from "@/swagger-models/assets-service-client";
import { errorMessages } from "@/common/error-message.constant";
import type { ICredentialChangedEvent } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { useClusterStore } from "@/stores/cluster.store";
import { MIN_CLUSTER_VERSION_FOR_NGC_API_KEY } from "@/common/version.constant";

export default defineComponent({
  name: "hybrid-environment-section",
  components: {
    FormToolsList,
    FormImagePullPolicy,
    FormSecuritySettings,
    FormRuntimeSettings,
    HybridSectionWrapper,
    RunaiFoldingSubSection,
    ServingEndpointSection,
    FormImageOriginSection,
  },
  emits: ["load-section-templates", "reset", "clear", "update", "toggle-linked", "credential-created"],
  props: {
    workloadSpec: {
      type: Object as PropType<IUISuperWorkloadSpec>,
      required: true,
    },
    selectedName: {
      type: String as PropType<string>,
      required: false,
    },
    environmentVariableSourceValueOptions: {
      type: Array as PropType<Array<TEnvVarNoneCustomSourceOption>>,
      required: false,
      default: () => [],
    },
    policy: {
      type: [Object, null] as PropType<UnifiedPolicyInfoPerReplicaType | null>,
      required: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
    renderKey: {
      type: String as PropType<string>,
      required: false,
      default: "",
    },
    supportConfigMapAsEnvVar: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportingServingPortAccess: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    authenticatedNgcOnly: { type: Boolean, default: false },
    isDistributedInference: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportUserCredentials: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportImagePullSecrets: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportAuthenticatedNgc: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    scopeId: {
      type: String as PropType<string>,
      required: false,
    },
    scopeType: {
      type: String as PropType<string>,
      required: false,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    showContentHeader: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    disableFields: {
      type: Object as PropType<IHybridEnvironmentSectionDisableFields>,
      required: false,
    },
    supportingServingPort: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    requiredFields: {
      type: Object as PropType<Record<string, boolean>>,
      required: false,
    },
    supportLinking: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    isLinked: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    validateEmptyEnvVars: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    preventTopPaddingWhenOpen: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideSpecificSections: {
      type: Object as PropType<IHideSpecificSectionsEnvironment>,
      required: false,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      clusterStore: useClusterStore(),
      EEnvVarFormType,
      linkedDisableText: "This can't be modified when the Link button is active",
    };
  },
  computed: {
    supportingNgcApiKey(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterId || "", MIN_CLUSTER_VERSION_FOR_NGC_API_KEY);
    },
    environmentSummary(): string {
      return this.workloadSpec.image || "";
    },
    runtimeSettings(): IRuntimeSettings {
      return {
        command: this.workloadSpec.command || null,
        args: this.workloadSpec.args || null,
        workingDir: this.workloadSpec.workingDir || null,
        hybridEnvironmentVariables: this.workloadSpec.environmentVariables || [],
      };
    },
    securityModel(): ISecurityModel {
      return {
        uidGidSource: this.workloadSpec.security?.uidGidSource || undefined,
        runAsUid: this.workloadSpec.security?.runAsUid,
        runAsGid: this.workloadSpec.security?.runAsGid,
        supplementalGroups: this.workloadSpec.security?.supplementalGroups,
        capabilities: this.workloadSpec.security?.capabilities,
      };
    },
    servingEndpointModel(): IServingEndpointAccessData | undefined {
      const servingPortData: IServingEndpointAccessData = {
        protocol: this.workloadSpec.servingPort?.protocol?.toString() || undefined,
        port:
          this.workloadSpec.servingPort?.container !== undefined && this.workloadSpec.servingPort?.container !== null
            ? this.workloadSpec.servingPort.container
            : undefined,
      };

      if (!this.supportingServingPortAccess) return servingPortData;

      return {
        ...servingPortData,
        access: inferenceWorkloadUtil.servingPortInternalToolInfo(this.workloadSpec.servingPort),
        clusterLocalAccessOnly: this.workloadSpec.servingPort?.clusterLocalAccessOnly,
      };
    },
    servingPortDisabledInfo(): string | undefined {
      if (this.isLinked) return this.linkedDisableText;
      if (this.disableFields?.servingPort) return this.disableFields.servingPort;
      return undefined;
    },
    formImagePullPolicyTooltipText(): string | undefined {
      if (this.isLinked) return this.linkedDisableText;
      if (this.policy?.rules?.imagePullPolicy?.canEdit === false) return errorMessages.CANT_BE_MODIFIED_POLICY;
      return undefined;
    },
  },
  methods: {
    updateModel(key: string, val: string | number | null | undefined) {
      this.$emit("update", { ...this.workloadSpec, [key]: val });
    },
    updateImagePullSecrets(imagePullSecrets: ImagePullSecret[]) {
      this.$emit("update", { ...this.workloadSpec, imagePullSecrets });
    },
    updateRuntimeSettings(runtimeSettings: IRuntimeSettings) {
      this.$emit("update", {
        ...this.workloadSpec,
        environmentVariables: runtimeSettings.hybridEnvironmentVariables,
        command: runtimeSettings.command,
        args: runtimeSettings.args,
        workingDir: runtimeSettings.workingDir,
      });
    },
    updateSecuritySettings(securityModel: ISecurityModel) {
      this.$emit("update", { ...this.workloadSpec, security: { ...this.workloadSpec.security, ...securityModel } });
    },
    updateTools(tools: IUIConnection[]) {
      this.$emit("update", { ...this.workloadSpec, connections: tools });
    },
    onServingPortChanged(servingEndpointModel: IServingEndpointAccessData): void {
      const servingPort = {
        protocol: servingEndpointModel.protocol,
        container: servingEndpointModel.port,
        clusterLocalAccessOnly: servingEndpointModel?.clusterLocalAccessOnly,
        ...inferenceWorkloadUtil.servingPortAccessAuthorization(servingEndpointModel.access),
      };

      this.$emit("update", { ...this.workloadSpec, servingPort });
    },
    onCredentialChanged(credentialEvent: ICredentialChangedEvent): void {
      if (!this.supportImagePullSecrets) return;
      const { credentialName, credentialSource } = credentialEvent;
      if (!credentialName) {
        this.updateImagePullSecrets([]);
        return;
      }

      if (credentialSource === ERunaiCredentialSource.MyCredentials) {
        this.updateImagePullSecrets([
          {
            name: credentialName,
            userCredential: true,
          },
        ]);
      } else if (credentialSource === ERunaiCredentialSource.SharedSecret) {
        this.updateImagePullSecrets([
          {
            name: credentialName,
          },
        ]);
      }
    },
  },
});
</script>
