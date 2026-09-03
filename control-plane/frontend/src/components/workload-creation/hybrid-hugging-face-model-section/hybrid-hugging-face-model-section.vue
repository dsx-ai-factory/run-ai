<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <hybrid-section-wrapper
    aid="hybrid-hugging-face-model-section"
    class="hybrid-hugging-face-model-section"
    label="Model"
    :section-disabled="sectionDisabled"
    default-opened
    :show-content-header="false"
  >
    <template #subheader>
      <span aid="hybrid-hugging-face-model-section-summary">{{ summary }}</span>
    </template>
    <section class="main-section">
      <hugging-face-model-and-access
        :model-value="hfSpec"
        :cluster-id="clusterId"
        :scope-model="scopeModel"
        :disable-model-selection="disableModelSelection"
        :is-access-required="requiredFields?.accessModel"
        :show-create-new-personal-credentials-btn="showCreateNewPersonalCredentialsBtn"
        :show-personal-credentials-option="showPersonalCredentialsOption"
        @update="onHfUpdate"
      />

      <runai-folding-sub-section v-if="hfSpec.modelName" label="Serving endpoint" aid="serving-endpoint-access-btn">
        <serving-endpoint-section
          class="serving-port-access-section"
          title="Set who can access the serving endpoint"
          :model-value="servingPortModel"
          @update:model-value="onServingPortChanged"
          :policy-rules="policy?.rules?.servingPort"
          :is-port-required="requiredFields?.port"
          :cluster-id="clusterId"
        />
      </runai-folding-sub-section>
    </section>
  </hybrid-section-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { HybridSectionWrapper } from "@/components/workload-creation/hybrid-section-wrapper";
import { HuggingFaceModelAndAccess } from "./hugging-face-model-and-access";
import { ServingEndpointSection } from "@/components/section/inference/serving-endpoint-section";
import { RunaiFoldingSubSection } from "@/components/common/runai-folding-sub-section";

// models
import type { ServingPortProtocol } from "@/swagger-models/assets-service-client";
import type { IServingEndpointAccessData } from "@/components/section/inference/serving-endpoint-section/serving-endpoint-access.model";
import type { IUISuperWorkloadSpec } from "@/models/workload.model";
import type { UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";
import type { IScopeModel } from "@/models/global.model";
import type { IInferenceFormUIState } from "@/stores/inference.store";
import type { IHuggingFaceModelAndAccessModel } from "./hugging-face-model-and-access/hugging-face-model-and-access.model";
import { ETokenProvisionMethods } from "./hugging-face-access-section/hugging-face-access-section.model";
import { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { EHuggingFaceEnvVariableNames } from "@/models/inference.model";

// utils
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";

export enum EHybridHuggingFaceModelSectionEvents {
  Update = "update",
  UpdateUIState = "update:ui-state",
}

export default defineComponent({
  components: {
    HybridSectionWrapper,
    HuggingFaceModelAndAccess,
    ServingEndpointSection,
    RunaiFoldingSubSection,
  },
  emits: {
    [EHybridHuggingFaceModelSectionEvents.Update]: (modelValue: IUISuperWorkloadSpec) => modelValue,
    [EHybridHuggingFaceModelSectionEvents.UpdateUIState]: (_: IInferenceFormUIState) => true,
  },
  props: {
    workloadSpec: {
      type: Object as PropType<IUISuperWorkloadSpec>,
      required: true,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    disableModelSelection: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    policy: {
      type: [Object, null] as PropType<UnifiedPolicyInfoPerReplicaType | null>,
      required: false,
    },
    uiState: {
      type: Object as PropType<IInferenceFormUIState>,
      required: false,
    },
    requiredFields: {
      type: Object as PropType<Record<string, boolean>>,
      required: false,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: true,
    },
    showCreateNewPersonalCredentialsBtn: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    showPersonalCredentialsOption: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
  },
  data() {
    return {
      EHybridHuggingFaceModelSectionEvents,
    };
  },
  computed: {
    summary(): string {
      const args = this.workloadSpec.args || "";
      return inferenceWorkloadUtil.getHFModelNameFromArgs(args) || "None";
    },
    hfSpec(): IHuggingFaceModelAndAccessModel {
      const args = this.workloadSpec.args || "";
      const modelName = inferenceWorkloadUtil.getHFModelNameFromArgs(args);
      const envVars = this.workloadSpec.environmentVariables ?? [];
      const hfTokenEnv = envVars.find((e) => e.name === EHuggingFaceEnvVariableNames.HF_TOKEN);
      let token: string | undefined;
      let selectedCredentials: string | undefined;
      let selectedTokenProvisionMethod =
        this.uiState?.hfSelectedTokenProvisionMethod ?? ETokenProvisionMethods.FromSecret;
      if (hfTokenEnv?.value) {
        token = hfTokenEnv.value;
        selectedTokenProvisionMethod = ETokenProvisionMethods.Manually;
      } else if (hfTokenEnv?.secret?.name || hfTokenEnv?.userCredential?.name) {
        selectedCredentials = hfTokenEnv.secret?.name ?? hfTokenEnv.userCredential?.name ?? undefined;
        selectedTokenProvisionMethod = ETokenProvisionMethods.FromSecret;
      }
      return {
        modelName,
        token,
        selectedCredentials,
        credentialsSource:
          this.uiState?.credentialsSource ?? inferenceWorkloadUtil.inferCredentialSourceFromEnvVar(hfTokenEnv),
        selectedTokenProvisionMethod,
        accessFromStorage: this.uiState?.accessFromStorage ?? false,
      };
    },
    servingPortModel(): IServingEndpointAccessData {
      return {
        protocol: this.workloadSpec.servingPort?.protocol?.toString() || "",
        port: this.workloadSpec.servingPort?.container || undefined,
        clusterLocalAccessOnly: this.workloadSpec.servingPort?.clusterLocalAccessOnly,
        access: inferenceWorkloadUtil.servingPortInternalToolInfo(this.workloadSpec.servingPort),
      };
    },
  },
  methods: {
    onHfUpdate(payload: IHuggingFaceModelAndAccessModel): void {
      const args = inferenceWorkloadUtil.getHFArgsForModel(this.workloadSpec.image || "", payload.modelName);
      const environmentVariables = this.applyHfAccessToEnv(this.workloadSpec.environmentVariables ?? [], payload);
      this.emitModelData({ ...this.workloadSpec, args, environmentVariables });
      const hasUiStateUpdate =
        payload.credentialsSource !== undefined ||
        payload.selectedTokenProvisionMethod !== undefined ||
        payload.accessFromStorage !== undefined;
      if (hasUiStateUpdate) {
        const uiStateUpdate: IInferenceFormUIState = {
          ...this.uiState,
          ...(payload.credentialsSource !== undefined && { credentialsSource: payload.credentialsSource }),
          ...(payload.selectedTokenProvisionMethod !== undefined && {
            hfSelectedTokenProvisionMethod:
              payload.selectedTokenProvisionMethod === ETokenProvisionMethods.FromSecret
                ? ETokenProvisionMethods.FromSecret
                : ETokenProvisionMethods.Manually,
          }),
          ...(payload.accessFromStorage !== undefined && { accessFromStorage: payload.accessFromStorage }),
        };
        this.$emit(EHybridHuggingFaceModelSectionEvents.UpdateUIState, uiStateUpdate);
      }
    },
    applyHfAccessToEnv(
      envVars: NonNullable<IUISuperWorkloadSpec["environmentVariables"]>,
      payload: IHuggingFaceModelAndAccessModel,
    ): NonNullable<IUISuperWorkloadSpec["environmentVariables"]> {
      const { token, selectedCredentials, selectedTokenProvisionMethod, credentialsSource } = payload;
      if (token && selectedTokenProvisionMethod === ETokenProvisionMethods.Manually) {
        return inferenceWorkloadUtil.replaceEnvironmentVariable(envVars, EHuggingFaceEnvVariableNames.HF_TOKEN, token);
      }
      if (selectedCredentials && selectedTokenProvisionMethod === ETokenProvisionMethods.FromSecret) {
        const isMyCredentials = credentialsSource === ERunaiCredentialSource.MyCredentials;
        const credentialConfig = { name: selectedCredentials, key: EHuggingFaceEnvVariableNames.HF_TOKEN };
        return inferenceWorkloadUtil.replaceEnvironmentVariable(
          envVars,
          EHuggingFaceEnvVariableNames.HF_TOKEN,
          undefined,
          isMyCredentials ? undefined : credentialConfig,
          isMyCredentials ? credentialConfig : undefined,
        );
      }
      return inferenceWorkloadUtil.excludeEnvironmentVariablesByName(EHuggingFaceEnvVariableNames.HF_TOKEN, envVars);
    },
    onServingPortChanged(servingPortAccess: IServingEndpointAccessData): void {
      this.emitModelData({
        ...this.workloadSpec,
        servingPort: {
          protocol: servingPortAccess?.protocol as ServingPortProtocol | undefined,
          container: servingPortAccess?.port,
          clusterLocalAccessOnly: servingPortAccess?.clusterLocalAccessOnly,
          ...inferenceWorkloadUtil.servingPortAccessAuthorization(servingPortAccess.access),
        },
      });
    },
    emitModelData(spec: IUISuperWorkloadSpec): void {
      this.$emit(EHybridHuggingFaceModelSectionEvents.Update, spec);
    },
  },
});
</script>
