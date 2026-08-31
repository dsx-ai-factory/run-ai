<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <hybrid-section-wrapper
    aid="hybrid-nim-model-section"
    class="hybrid-nim-model-section"
    label="Model"
    :section-disabled="sectionDisabled"
    @clear="$emit(ENimModelEvents.Clear)"
    @reset="$emit(ENimModelEvents.Reset)"
    default-opened
    :show-content-header="false"
  >
    <template #subheader>
      <span>{{ summary }}</span>
    </template>

    <section class="main-section">
      <nim-info-box v-if="showInfoBox" @close="hideInfoBox" />
      <model-source-selector v-model="selectedModelSource" class="q-mb-lg" @update:model-value="onModelSourceChanged" />
      <nim-model-and-access
        v-if="selectedModelSource === nimSourceValue"
        :model-value="nimModelAndAccessModel"
        @update:model-value="onNimModelAndAccessChanged"
        @is-name-valid="isNameValid = $event"
        @is-profile-valid="isProfileValid = $event"
        :cluster-id="clusterId"
        :scope-model="scopeModel"
        :is-profile-required="requiredFields?.profile"
        :is-access-required="requiredFields?.accessModel"
        :show-create-new-personal-credentials-btn="showCreateNewPersonalCredentialsBtn"
        :show-personal-credentials-option="showPersonalCredentialsOption"
        :supporting-ngc-api-key="supportingNgcApiKey"
        :supports-docker-registry="supportsDockerRegistry"
      />
      <template v-if="selectedModelSource === huggingFaceSourceValue">
        <hugging-face-model-and-access
          :model-value="hfSpec"
          :cluster-id="clusterId"
          :scope-model="scopeModel"
          :is-access-required="requiredFields?.accessModel"
          :show-create-new-personal-credentials-btn="showCreateNewPersonalCredentialsBtn"
          :show-personal-credentials-option="showPersonalCredentialsOption"
          @update="onHfUpdate"
        />

        <nim-image-pull-credential-selector
          v-if="supportingNgcApiKey && showImagePullCredentials && hfModelName"
          :cluster-id="clusterId"
          :show-create-new-credential-btn="showCreateNewPersonalCredentialsBtn"
          :scope-model="scopeModel"
          :model-value="nimImagePullCredentialModel"
          @update:model-value="onNimImagePullCredentialChanged"
        />
        <docker-registry-selector
          v-if="!supportingNgcApiKey && showImagePullCredentials && hfModelName"
          :model-value="dockerRegistrySelectorModel"
          :cluster-id="clusterId"
          :scope-model="scopeModel"
          @update:model-value="onDockerRegistryChanged"
        />
      </template>

      <section v-if="isModelNameFilled" class="serving-port-section q-mt-md">
        <serving-endpoint-section
          title="Set who can access the serving endpoint"
          :model-value="servingPortModel"
          @update:model-value="onServingPortChanged"
          show-protocol-and-port
          :policy-rules="policy?.rules?.servingPort"
          :is-port-required="requiredFields?.port"
          :cluster-id="clusterId"
        />
      </section>
    </section>
  </hybrid-section-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { HybridSectionWrapper } from "@/components/workload-creation/hybrid-section-wrapper";
import { ServingEndpointSection } from "@/components/section/inference/serving-endpoint-section";
import { NimModelAndAccess } from "./nim-model-and-access";
import { ModelSourceSelector } from "./model-source-selector";
import { HuggingFaceModelAndAccess } from "@/components/workload-creation/hybrid-hugging-face-model-section/hugging-face-model-and-access";
import { DockerRegistrySelector } from "./nim-access-model/docker-registry-selector";
import {
  NimImagePullCredentialSelector,
  ENimImagePullCredentialType,
  type INimImagePullCredentialSelectorModel,
} from "./nim-access-model/nim-image-pull-credential-selector";
import NimInfoBox from "./nim-info-box.vue";
import { EModelSource } from "./model-source-selector/model-source-selector.model";

// models
import type { IServingEndpointAccessData } from "@/components/section/inference/serving-endpoint-section/serving-endpoint-access.model";
import { storageUtil } from "@/utils/storage.util";
import { STORAGE_KEYS } from "@/models/storage.model";
import type { IUISuperWorkloadSpec } from "@/models/workload.model";
import type { UnifiedPolicyInfoPerReplicaType } from "@/swagger-models/workloads-client";
import {
  EHuggingFaceEnvVariableNames,
  ENimEnvVariableNames,
  NIM_IMAGE_FOR_HF_MODELS,
  NIM_IMAGE_PREFIX,
  SERVING_PORT_SERVER_PORT,
} from "@/models/inference.model";
import type { IHuggingFaceModelAndAccessModel } from "@/components/workload-creation/hybrid-hugging-face-model-section/hugging-face-model-and-access/hugging-face-model-and-access.model";
import { ETokenProvisionMethods } from "@/components/workload-creation/hybrid-hugging-face-model-section/hugging-face-access-section/hugging-face-access-section.model";
import type { ServingPortProtocol } from "@/swagger-models/assets-service-client";
import type { INimAccessModelModel } from "./nim-access-model/nim-access-model.model";
import type { INimModelAndAccessModel } from "./nim-model-and-access/nim-model-and-access.model";
import type { IScopeModel } from "@/models/global.model";
import type { IInferenceFormUIState } from "@/stores/inference.store";
import type { IDockerRegistrySelectorModel } from "./nim-access-model/docker-registry-selector/docker-registry-selector.model";

// utils
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";

import { MIN_CLUSTER_VERSION_FOR_NGC_API_KEY } from "@/common/version.constant";
import { useClusterStore } from "@/stores/cluster.store";
import { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";

export enum ENimModelEvents {
  Update = "update",
  Clear = "clear",
  Reset = "reset",
  UpdateUIState = "update:ui-state",
}

const envInNim = [
  ENimEnvVariableNames.OutlinesCacheDir,
  ENimEnvVariableNames.NimServerPort,
  ENimEnvVariableNames.NimJsonlLogging,
  ENimEnvVariableNames.NimLogLevel,
  ENimEnvVariableNames.NimCachePath,
  ENimEnvVariableNames.NimModelProfile,
  ENimEnvVariableNames.NgcApiKey,
];
const envInHf = [
  EHuggingFaceEnvVariableNames.HF_TOKEN,
  ENimEnvVariableNames.NimModelName,
  ENimEnvVariableNames.NimHttpApiPort,
];

const NIM_DEFAULT_ENV_VARS: Array<{ name: string; value: string }> = [
  { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
  { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
  { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
  { name: ENimEnvVariableNames.NimCachePath, value: "/home" },
];

export default defineComponent({
  components: {
    HybridSectionWrapper,
    ModelSourceSelector,
    NimModelAndAccess,
    HuggingFaceModelAndAccess,
    DockerRegistrySelector,
    NimImagePullCredentialSelector,
    ServingEndpointSection,
    NimInfoBox,
  },
  emits: {
    [ENimModelEvents.Update]: (spec: IUISuperWorkloadSpec) => spec,
    [ENimModelEvents.Clear]: () => true,
    [ENimModelEvents.Reset]: () => true,
    [ENimModelEvents.UpdateUIState]: (_: IInferenceFormUIState) => true,
  },
  props: {
    workloadSpec: {
      type: Object as PropType<IUISuperWorkloadSpec>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
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
    supportsDockerRegistry: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    forceHideInfoBox: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      isNameValid: true as boolean,
      isProfileValid: true as boolean,
      showInfoBox: true as boolean,
      ENimModelEvents,
      selectedModelSource: EModelSource.Nim as EModelSource | "",
      nimSourceValue: EModelSource.Nim,
      huggingFaceSourceValue: EModelSource.HuggingFace,
    };
  },
  created() {
    if (this.forceHideInfoBox || this.supportsDockerRegistry) {
      this.showInfoBox = false;
    } else {
      this.showInfoBox = !storageUtil.get<boolean>(STORAGE_KEYS.NIM_INFO_BOX_CLOSED);
    }
    const savedSource = this.uiState?.selectedModelSource;
    if (savedSource === EModelSource.Nim || savedSource === EModelSource.HuggingFace) {
      this.selectedModelSource = savedSource;
    }
  },
  computed: {
    isModelNameFilled(): boolean {
      const nimFilled = !!this.nimModelName;
      const hfFilled = !!this.hfModelName;
      if (this.selectedModelSource === this.nimSourceValue) return nimFilled;
      if (this.selectedModelSource === this.huggingFaceSourceValue) return hfFilled;
      return false;
    },
    supportingNgcApiKey(): boolean {
      return this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_NGC_API_KEY);
    },
    showImagePullCredentials(): boolean {
      return this.supportsDockerRegistry ?? false;
    },
    dockerRegistrySelectorModel(): IDockerRegistrySelectorModel {
      return {
        dockerRegistryName: this.workloadSpec.imagePullSecrets?.[0]?.name || "",
      };
    },
    nimImagePullCredentialModel(): INimImagePullCredentialSelectorModel {
      const credentialName = this.workloadSpec.imagePullSecrets?.[0]?.name || "";
      const credentialType =
        this.uiState?.nimHfImagePullCredentialType ??
        (credentialName ? ENimImagePullCredentialType.DockerRegistry : ENimImagePullCredentialType.NgcApiKey);
      return { credentialType, credentialName };
    },
    summary(): string {
      if (this.selectedModelSource === this.huggingFaceSourceValue) {
        return this.hfModelName || "None";
      }
      return this.nimModelName || "None";
    },
    hfSpec(): IHuggingFaceModelAndAccessModel {
      const envVars = this.workloadSpec.environmentVariables ?? [];
      const modelName = this.hfModelName;
      const hfTokenEnv = envVars.find((e) => e.name === EHuggingFaceEnvVariableNames.HF_TOKEN);
      let token: string | undefined;
      let selectedCredentials: string | undefined;
      let selectedTokenProvisionMethod = ETokenProvisionMethods.Manually;
      if (hfTokenEnv?.value) {
        token = hfTokenEnv.value;
        selectedTokenProvisionMethod = ETokenProvisionMethods.Manually;
      } else if (hfTokenEnv?.secret?.name || hfTokenEnv?.userCredential?.name) {
        selectedCredentials = hfTokenEnv.secret?.name ?? hfTokenEnv.userCredential?.name ?? undefined;
        selectedTokenProvisionMethod = ETokenProvisionMethods.FromSecret;
      } else if (this.uiState?.hfSelectedTokenProvisionMethod === ETokenProvisionMethods.FromSecret) {
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
    nimModelName(): string {
      return (this.workloadSpec.image || "").replace(`${NIM_IMAGE_PREFIX}/`, "");
    },
    hfModelName(): string {
      const rawModelName = inferenceWorkloadUtil.getEnvironmentVariableValue(
        ENimEnvVariableNames.NimModelName,
        this.workloadSpec.environmentVariables ?? [],
      );
      return rawModelName?.startsWith("hf://") ? rawModelName.slice(5) : rawModelName || "";
    },
    servingPortModel(): IServingEndpointAccessData {
      return {
        protocol: this.workloadSpec.servingPort?.protocol?.toString() || "",
        port: this.workloadSpec.servingPort?.container || undefined,
        clusterLocalAccessOnly: this.workloadSpec.servingPort?.clusterLocalAccessOnly,
        access: inferenceWorkloadUtil.servingPortInternalToolInfo(this.workloadSpec.servingPort),
      };
    },
    profileNameOrHashModelValue(): string {
      return (
        this.workloadSpec.environmentVariables?.find((env) => env.name === ENimEnvVariableNames.NimModelProfile)
          ?.value || ""
      );
    },
    accessModel(): INimAccessModelModel {
      const envVars = this.workloadSpec.environmentVariables || [];
      const ngcEnv = envVars.find((env) => env.name === ENimEnvVariableNames.NgcApiKey);
      const retVal: INimAccessModelModel = {
        apiKey: undefined,
        selectedCredentials: undefined,
        accessFromStorage: this.uiState?.accessFromStorage,
        credentialsSource:
          this.uiState?.credentialsSource ?? inferenceWorkloadUtil.inferCredentialSourceFromEnvVar(ngcEnv),
        credentialType: this.uiState?.credentialType,
        dockerRegistry: {
          dockerRegistryName: this.workloadSpec.imagePullSecrets?.[0]?.name || "",
        },
      };

      if (ngcEnv?.value) {
        retVal.apiKey = ngcEnv.value;
      } else if (ngcEnv?.secret?.name) {
        retVal.selectedCredentials = ngcEnv.secret.name;
      } else if (ngcEnv?.userCredential?.name) {
        retVal.selectedCredentials = ngcEnv.userCredential.name;
      }

      return retVal;
    },
    nimModelAndAccessModel(): INimModelAndAccessModel {
      return {
        modelName: this.nimModelName,
        profileNameOrHash: this.profileNameOrHashModelValue,
        accessModel: this.accessModel,
      };
    },
  },
  methods: {
    onModelSourceChanged(selectedSource: EModelSource): void {
      const servingPort = this.workloadSpec.servingPort?.container ?? SERVING_PORT_SERVER_PORT;
      if (selectedSource === this.huggingFaceSourceValue) {
        let envVars = inferenceWorkloadUtil.excludeEnvironmentVariablesByName(
          envInNim,
          this.workloadSpec.environmentVariables ?? [],
        );
        envVars = inferenceWorkloadUtil.replaceEnvironmentVariable(
          envVars,
          ENimEnvVariableNames.NimServerPort,
          String(servingPort),
        );
        envVars = inferenceWorkloadUtil.replaceEnvironmentVariable(
          envVars,
          ENimEnvVariableNames.NimHttpApiPort,
          String(servingPort),
        );
        this.emitModelData({
          ...this.workloadSpec,
          image: NIM_IMAGE_FOR_HF_MODELS,
          environmentVariables: envVars,
          imagePullSecrets: undefined,
        });
      } else if (selectedSource === this.nimSourceValue) {
        let environmentVariables = inferenceWorkloadUtil.excludeEnvironmentVariablesByName(
          envInHf,
          this.workloadSpec.environmentVariables ?? [],
        );
        for (const { name, value } of NIM_DEFAULT_ENV_VARS) {
          environmentVariables = inferenceWorkloadUtil.replaceEnvironmentVariable(environmentVariables, name, value);
        }
        environmentVariables = inferenceWorkloadUtil.replaceEnvironmentVariable(
          environmentVariables,
          ENimEnvVariableNames.NimServerPort,
          String(servingPort),
        );
        this.emitModelData({ ...this.workloadSpec, environmentVariables, imagePullSecrets: undefined });
      }
      this.$emit(ENimModelEvents.UpdateUIState, { ...this.uiState, selectedModelSource: selectedSource });
    },
    onHfUpdate(payload: IHuggingFaceModelAndAccessModel): void {
      let envVars = this.workloadSpec.environmentVariables ?? [];
      if (payload.modelName) {
        const nimModelName = payload.modelName.startsWith("hf://") ? payload.modelName : `hf://${payload.modelName}`;
        envVars = inferenceWorkloadUtil.replaceEnvironmentVariable(
          envVars,
          ENimEnvVariableNames.NimModelName,
          nimModelName,
        );
      } else {
        envVars = inferenceWorkloadUtil.excludeEnvironmentVariablesByName(ENimEnvVariableNames.NimModelName, envVars);
      }
      envVars = this.applyHfAccessToEnv(envVars, payload);
      this.emitModelData({ ...this.workloadSpec, environmentVariables: envVars });
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
        this.$emit(ENimModelEvents.UpdateUIState, uiStateUpdate);
      }
    },
    applyHfAccessToEnv(
      envVars: NonNullable<IUISuperWorkloadSpec["environmentVariables"]>,
      payload: IHuggingFaceModelAndAccessModel,
    ): NonNullable<IUISuperWorkloadSpec["environmentVariables"]> {
      if (payload.accessFromStorage) {
        return inferenceWorkloadUtil.excludeEnvironmentVariablesByName(EHuggingFaceEnvVariableNames.HF_TOKEN, envVars);
      }
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
    onDockerRegistryChanged(dockerRegistry: IDockerRegistrySelectorModel): void {
      const imagePullSecrets = this.updateImagePullSecrets({
        ...this.accessModel,
        dockerRegistry,
      });
      this.emitModelData({
        ...this.workloadSpec,
        imagePullSecrets,
      });
    },
    onNimImagePullCredentialChanged(model: INimImagePullCredentialSelectorModel): void {
      const imagePullSecrets = model.credentialName ? [{ name: model.credentialName, userCredential: true }] : undefined;
      this.emitModelData({
        ...this.workloadSpec,
        imagePullSecrets,
      });
      this.$emit(ENimModelEvents.UpdateUIState, {
        ...this.uiState,
        nimHfImagePullCredentialType: model.credentialType,
      });
    },
    onServingPortChanged(servingPortAccess: IServingEndpointAccessData): void {
      const port = servingPortAccess?.port ?? SERVING_PORT_SERVER_PORT;
      let environmentVariables = this.workloadSpec.environmentVariables ?? [];
      environmentVariables = inferenceWorkloadUtil.replaceEnvironmentVariable(
        environmentVariables,
        ENimEnvVariableNames.NimServerPort,
        String(port),
      );
      if (this.selectedModelSource === this.huggingFaceSourceValue) {
        environmentVariables = inferenceWorkloadUtil.replaceEnvironmentVariable(
          environmentVariables,
          ENimEnvVariableNames.NimHttpApiPort,
          String(port),
        );
      }
      this.emitModelData({
        ...this.workloadSpec,
        servingPort: {
          protocol: servingPortAccess?.protocol as ServingPortProtocol | undefined,
          container: port,
          clusterLocalAccessOnly: servingPortAccess?.clusterLocalAccessOnly,
          ...inferenceWorkloadUtil.servingPortAccessAuthorization(servingPortAccess?.access),
        },
        environmentVariables,
      });
    },
    onNimModelAndAccessChanged(modelAndAccess: INimModelAndAccessModel): void {
      const { modelName, profileNameOrHash, accessModel } = modelAndAccess;
      const image = modelName ? `${NIM_IMAGE_PREFIX}/${modelName}` : "";

      let environmentVariables = this.workloadSpec.environmentVariables ?? [];
      if (profileNameOrHash) {
        environmentVariables = inferenceWorkloadUtil.replaceEnvironmentVariable(
          environmentVariables,
          ENimEnvVariableNames.NimModelProfile,
          profileNameOrHash,
        );
      } else {
        environmentVariables = inferenceWorkloadUtil.excludeEnvironmentVariablesByName(
          ENimEnvVariableNames.NimModelProfile,
          environmentVariables,
        );
      }
      environmentVariables = this.updateEnvironmentVariables(accessModel, environmentVariables);
      const imagePullSecrets = this.updateImagePullSecrets(accessModel);

      this.emitModelData({
        ...this.workloadSpec,
        image,
        environmentVariables,
        imagePullSecrets,
      });

      this.$emit(ENimModelEvents.UpdateUIState, {
        ...this.uiState,
        accessFromStorage: accessModel.accessFromStorage,
        credentialsSource: accessModel.credentialsSource,
        credentialType: accessModel.credentialType,
      });
    },
    emitModelData(spec: IUISuperWorkloadSpec): void {
      this.$emit(ENimModelEvents.Update, spec);
    },
    hideInfoBox(): void {
      this.showInfoBox = false;
      storageUtil.save(STORAGE_KEYS.NIM_INFO_BOX_CLOSED, true);
    },
    updateEnvironmentVariables(
      accessModel: INimAccessModelModel,
      baseEnv?: IUISuperWorkloadSpec["environmentVariables"],
    ) {
      const { apiKey, selectedCredentials, credentialsSource } = accessModel;
      const environmentVariables = baseEnv ?? this.workloadSpec.environmentVariables;

      if (apiKey) {
        return inferenceWorkloadUtil.replaceEnvironmentVariable(
          environmentVariables,
          ENimEnvVariableNames.NgcApiKey,
          apiKey,
        );
      }
      if (selectedCredentials) {
        const isMyCredentials = credentialsSource === ERunaiCredentialSource.MyCredentials;
        const credentialConfig = {
          name: selectedCredentials,
          key: ENimEnvVariableNames.NgcApiKey,
        };
        return inferenceWorkloadUtil.replaceEnvironmentVariable(
          environmentVariables,
          ENimEnvVariableNames.NgcApiKey,
          undefined,
          isMyCredentials ? undefined : credentialConfig,
          isMyCredentials ? credentialConfig : undefined,
        );
      }
      return inferenceWorkloadUtil.excludeEnvironmentVariablesByName(
        ENimEnvVariableNames.NgcApiKey,
        environmentVariables ?? [],
      );
    },
    updateImagePullSecrets(accessModel: INimAccessModelModel) {
      const dockerRegistryName = accessModel.dockerRegistry?.dockerRegistryName;

      return dockerRegistryName ? [{ name: dockerRegistryName, userCredential: true }] : undefined;
    },
  },
});
</script>

<style scoped lang="scss"></style>
