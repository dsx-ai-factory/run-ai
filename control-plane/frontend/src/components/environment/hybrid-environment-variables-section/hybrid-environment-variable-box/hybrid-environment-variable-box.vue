<template>
  <section class="environment-variable-box q-pa-md row items-center">
    <div class="selects-wrapper row items-center flex-1">
      <q-input
        aid="environment-variable-name"
        class="col-3 q-mr-md"
        label="Name"
        stack-label
        placeholder="Enter a name"
        :model-value="environmentVariable.name"
        @update:model-value="updateEnvironmentVariableName"
        :rules="[isValidName]"
        no-error-icon
        :disable="readOnly"
      />
      <runai-select
        aid="environment-variable-source"
        class="q-mr-md"
        :style="{ width: '130px' }"
        :options="sourceOptions"
        label="Source"
        emit-value
        stack-label
        unclearable
        :model-value="selectedSource"
        @update:model-value="updateSelectedSource"
        :rules="[() => true]"
        :disable="readOnly"
      />
      <q-input
        v-if="selectedSource === 'custom'"
        aid="environment-variable-value"
        class="col-3"
        stack-label
        placeholder="Enter a value"
        label="Value"
        :rules="[isValidValue]"
        :model-value="environmentVariable.value"
        @update:model-value="updateEnvironmentVariableValue"
        no-error-icon
        :disable="readOnly"
        :hint="getEnvironmentVariableHint"
      />
      <template v-else-if="selectedSource === EEnvVarSourceType.Secrets">
        <runai-select
          aid="environment-variable-secret"
          :style="{ width: '150px' }"
          class="q-mr-md"
          label="Secret name"
          placeholder="Select..."
          stack-label
          emit-value
          :options="secretOptions"
          :model-value="selectedSecretName"
          @update:model-value="updateEnvironmentVariableSecretName"
          :rules="[isNameSelected]"
          no-option-text="No secrets found"
          :disable="readOnly"
        />
        <template v-if="isSecretSelected">
          <runai-select
            v-if="showKeySelect"
            aid="environment-variable-key-select"
            :style="{ width: '150px' }"
            stack-label
            placeholder="Select a key"
            label="Secret key"
            no-error-icon
            :rules="[isValidKey]"
            :disable="readOnly"
            :options="secretKeyOptions"
            :model-value="selectedSecretKey"
            @update:model-value="updateEnvironmentVariableSecretKey"
            no-option-text="Invalid secret"
          />
          <q-input
            v-else-if="!isDockerRegistrySelected"
            aid="environment-variable-key-input"
            :style="{ width: '150px' }"
            stack-label
            placeholder="Enter a key"
            label="Secret key"
            :model-value="selectedSecretKey"
            @update:model-value="updateEnvironmentVariableSecretKey"
            no-error-icon
            :rules="[isValidKey]"
            :disable="readOnly"
          />
        </template>
      </template>
      <template v-else-if="selectedSource === EEnvVarSourceType.ConfigMap">
        <runai-select
          aid="environment-variable-config-map-name-select"
          :style="{ width: '150px' }"
          class="q-mr-md"
          stack-label
          placeholder="Select..."
          label="ConfigMap name"
          no-error-icon
          :rules="[isNameSelected]"
          :disable="readOnly"
          :options="configMapOptions"
          :model-value="environmentVariable.configMap?.name || null"
          @update:model-value="updateEnvironmentVariableConfigMapName"
          no-option-text="No ConfigMaps found"
          emit-value
        />
        <q-input
          aid="environment-variable-config-map-key-input"
          :style="{ width: '150px' }"
          stack-label
          placeholder="Enter a value"
          label="ConfigMap key"
          :model-value="environmentVariable.configMap?.key"
          @update:model-value="updateEnvironmentVariableConfigMapKey"
          no-error-icon
          lazy-rules
          :rules="[isValidKey]"
          :disable="readOnly"
        />
      </template>
      <template v-else-if="selectedSource === EEnvVarSourceType.UserCredential">
        <runai-credential-selector-by-source
          :model-value="credentialSelectorModel"
          @update:model-value="onCredentialSelectorChanged"
          :source-select-config="{ hidden: true }"
          :type-select-config="{
            hidden: false,
            allowedTypes: allowedCredentialTypes,
            disabledTypes: !supportingNgcApiKey ? { [ERunaiCredentialType.NgcApiKey]: {} } : undefined,
          }"
          :credential-select-config="{
            required: true,
            showCreateNewBtn: !readOnly,
            supportMultipleGenericSecretItems: true,
          }"
          :disable-all="readOnly"
        />
        <q-input
          v-if="showCredentialKeyInput"
          aid="environment-variable-user-credentials-key-input"
          :style="{ width: '150px' }"
          stack-label
          placeholder="Enter a key"
          label="Credential key"
          no-error-icon
          :model-value="selectedUserCredentialKey"
          @update:model-value="updateEnvironmentVariableUserCredentialKey"
          :rules="[isValidKey]"
          :disable="readOnly"
        />
      </template>
    </div>
    <div>
      <q-btn
        v-if="selectedSource === 'custom' && showInstructionLogo"
        round
        class="q-py-sm q-ml-md text-black-54 runai-icon-btn--default"
        flat
        size="sm"
        :icon="getInstructionIcon"
      >
        <q-tooltip max-width="300px">
          {{ environmentVariable.description || (editInstructionText ? "Add instructions" : "Read instructions") }}
        </q-tooltip>
        <runai-instructions-window-with-poup-over
          @update-instruction-text="updateInstructionText"
          :edit-instruction-text="editInstructionText"
          :environment-variable-description="environmentVariable.description"
        ></runai-instructions-window-with-poup-over>
      </q-btn>
      <q-btn
        aid="remove-environment-variable-button"
        class="close-button"
        icon="ph-bold ph-x"
        flat
        round
        @click="removeEnvironmentVariable"
        :disable="readOnly"
        ><q-tooltip v-if="disabledTooptip">{{ disabledTooptip }}</q-tooltip></q-btn
      >
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiCredentialSelectorBySource } from "@/components/common/runai-credential-selector-by-source";
// Models
import type { ISelectOption } from "@/models/global.model";
import { AssetKind, type EnvironmentVariable } from "@/swagger-models/assets-service-client";
import { SecretType } from "@/swagger-models/k8s-objects-tracker-client";
import {
  type ICredentialSelectorModel,
  ERunaiCredentialSource,
  ERunaiCredentialType,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
// Utils
import { errorMessages } from "@/common/error-message.constant";
import { isNotEmpty } from "@/common/form.validators";
import {
  type TEnvVarNoneCustomSourceOption,
  EEnvVarSourceType,
  EAccessKeySecretKey,
  EUserPassSecretKey,
} from "@/models/environment.model";
// cmps
import { RunaiInstructionsWindowWithPoupOver } from "@/components/common/runai-instructions-window/with-popover";

const DOCKER_REGISTRY_KEY = ".dockerconfigjson";
const NGC_API_KEY = "NGC_API_KEY";

const ASSET_KIND_TO_CREDENTIAL_TYPE: Partial<Record<AssetKind, ERunaiCredentialType>> = {
  [AssetKind.DockerRegistry]: ERunaiCredentialType.DockerRegistry,
  [AssetKind.NgcApiKey]: ERunaiCredentialType.NgcApiKey,
};

const CREDENTIAL_TYPE_TO_KEY: Partial<Record<ERunaiCredentialType, string>> = {
  [ERunaiCredentialType.DockerRegistry]: DOCKER_REGISTRY_KEY,
  [ERunaiCredentialType.NgcApiKey]: NGC_API_KEY,
};

export default defineComponent({
  name: "hybrid-environment-variable-box",
  components: {
    RunaiInstructionsWindowWithPoupOver,
    RunaiSelect,
    RunaiCredentialSelectorBySource,
  },
  emits: ["update-environment-variable", "remove-environment-variable"],
  props: {
    environmentVariable: {
      type: Object as PropType<EnvironmentVariable>,
      required: true,
    },
    sourceValueOptions: {
      type: Array as PropType<TEnvVarNoneCustomSourceOption[]>,
      required: true,
    },
    validateEmptyEnvVars: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    editInstructionText: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disabledTooptip: {
      type: String as PropType<string>,
      required: false,
    },
    supportConfigMap: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportUserCredentials: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportingNgcApiKey: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      selectedSource: EEnvVarSourceType.Custom as EEnvVarSourceType,
      selectedCredentialType: undefined as ERunaiCredentialType | undefined,
      EEnvVarSourceType,
      ERunaiCredentialType,
    };
  },
  created() {
    if (this.environmentVariable.secret) {
      this.selectedSource = EEnvVarSourceType.Secrets;
    } else if (this.environmentVariable.configMap) {
      this.selectedSource = EEnvVarSourceType.ConfigMap;
    } else if (this.environmentVariable.userCredential) {
      this.selectedSource = EEnvVarSourceType.UserCredential;
      const assetKind = this.relevantUserCredential?.type as AssetKind | undefined;
      this.selectedCredentialType = assetKind
        ? (ASSET_KIND_TO_CREDENTIAL_TYPE[assetKind] ?? ERunaiCredentialType.GenericSecret)
        : undefined;
    }
  },
  computed: {
    secretOptions(): ISelectOption[] {
      return this.sourceValueOptions
        .filter((c) => c.type !== AssetKind.ConfigMap && !c.isPrivate)
        .map((c) => ({
          value: c.value,
          label: c.label,
        }));
    },
    configMapOptions(): ISelectOption[] {
      return this.sourceValueOptions
        .filter((c) => c.type === AssetKind.ConfigMap)
        .map((c) => ({
          value: c.value,
          label: c.label,
        }));
    },
    userCredentialsOptions(): ISelectOption[] {
      return this.sourceValueOptions
        .filter((c) => c.isPrivate)
        .map((c) => ({
          value: c.value,
          label: c.label,
        }));
    },
    isSecretSelected(): boolean {
      return !!(this.environmentVariable as EnvironmentVariable).secret?.name;
    },
    isMyCredentialsSelected(): boolean {
      return !!(this.environmentVariable as EnvironmentVariable).userCredential?.name;
    },
    sourceOptions(): ISelectOption[] {
      const options: ISelectOption[] = [
        {
          value: EEnvVarSourceType.Custom,
          label: "Custom",
          toolTip: "Enter a name and value for the environment variable",
        },
        {
          value: EEnvVarSourceType.Secrets,
          label: "Shared secret",
          toolTip: "Select a predefined Kubernetes secret",
        },
      ];
      if (this.supportUserCredentials) {
        options.push({
          value: EEnvVarSourceType.UserCredential,
          label: "My credentials",
          toolTip: "These are only available to you. To add new credentials, go to User settings.",
        });
      }
      if (this.supportConfigMap) {
        options.push({
          value: EEnvVarSourceType.ConfigMap,
          label: "ConfigMap",
          toolTip: "Select a predefined ConfigMap",
        });
      }

      return options;
    },
    showKeySelect(): boolean {
      const kind = this.relevantSecret?.type;
      return kind === SecretType.AccessKey || kind === SecretType.Password;
    },
    isDockerRegistrySelected(): boolean {
      return (
        this.relevantSecret?.type === SecretType.DockerRegistry ||
        this.selectedCredentialType === ERunaiCredentialType.DockerRegistry
      );
    },
    isNgcApiKeySelected(): boolean {
      return this.selectedCredentialType === ERunaiCredentialType.NgcApiKey;
    },
    showCredentialKeyInput(): boolean {
      return this.isMyCredentialsSelected && !this.isDockerRegistrySelected && !this.isNgcApiKeySelected;
    },
    showInstructionLogo(): boolean {
      // we will show the instruction btn if the user is in create/edit asset
      // or
      // if there is a description that we need to show to the user
      return this.editInstructionText || !!this.environmentVariable.description;
    },
    relevantSecret(): TEnvVarNoneCustomSourceOption | undefined {
      return this.sourceValueOptions.find(
        (c) => c.value === (this.environmentVariable as EnvironmentVariable).secret?.name,
      );
    },
    relevantUserCredential(): TEnvVarNoneCustomSourceOption | undefined {
      return this.sourceValueOptions.find(
        (c) => c.value === (this.environmentVariable as EnvironmentVariable).userCredential?.name,
      );
    },
    secretKeyOptions(): EAccessKeySecretKey[] | EUserPassSecretKey[] {
      const kind = this.relevantSecret?.type;

      switch (kind) {
        case SecretType.AccessKey:
          return [EAccessKeySecretKey.Key, EAccessKeySecretKey.Secret];
        case SecretType.Password:
          return [EUserPassSecretKey.Pass, EUserPassSecretKey.User];
        default:
          return [];
      }
    },
    getEnvironmentVariableHint(): string {
      return this.editInstructionText ? "Enter a value, or leave it empty and provide instructions" : "";
    },
    getInstructionIcon(): string {
      return "ph-fill ph-note";
    },
    selectedSecretName(): string | null {
      return (this.environmentVariable as EnvironmentVariable).secret?.name || null;
    },
    selectedSecretKey(): string | null {
      return (this.environmentVariable as EnvironmentVariable).secret?.key || null;
    },
    selectedUserCredential(): string | null {
      return (this.environmentVariable as EnvironmentVariable).userCredential?.name || null;
    },
    selectedUserCredentialKey(): string | null {
      return (this.environmentVariable as EnvironmentVariable).userCredential?.key || null;
    },
    credentialSelectorModel(): ICredentialSelectorModel {
      const userCredential = this.environmentVariable.userCredential;
      const assetKind = this.relevantUserCredential?.type as AssetKind | undefined;
      const type =
        this.selectedCredentialType ??
        (assetKind
          ? (ASSET_KIND_TO_CREDENTIAL_TYPE[assetKind] ?? ERunaiCredentialType.GenericSecret)
          : ERunaiCredentialType.GenericSecret);

      return {
        source: ERunaiCredentialSource.MyCredentials,
        type,
        credentialName: userCredential?.name || "",
      };
    },
    allowedCredentialTypes(): ERunaiCredentialType[] {
      const types = [ERunaiCredentialType.GenericSecret, ERunaiCredentialType.DockerRegistry];
      if (this.supportingNgcApiKey) {
        types.push(ERunaiCredentialType.NgcApiKey);
      }
      return types;
    },
  },
  methods: {
    updateEnvironmentVariableConfigMapName(configMapName: string): void {
      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        configMap: {
          name: configMapName,
        },
      });
    },
    updateEnvironmentVariableConfigMapKey(key: string | number | null): void {
      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        configMap: {
          ...this.environmentVariable.configMap,
          key,
        },
      });
    },
    updateEnvironmentVariableUserCredential(userCredential: string): void {
      const secretType = this.sourceValueOptions.find((c) => c.value === userCredential)?.type;
      const key = secretType === AssetKind.DockerRegistry ? DOCKER_REGISTRY_KEY : undefined;

      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        userCredential: {
          name: userCredential,
          key,
        },
      });
    },
    updateEnvironmentVariableUserCredentialKey(key: string | number | null): void {
      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        userCredential: {
          ...this.environmentVariable.userCredential,
          key,
        },
      });
    },
    onCredentialSelectorChanged(value: ICredentialSelectorModel): void {
      const credentialName = value.credentialName;
      const credentialType = value.type;

      this.selectedCredentialType = credentialType;

      const key = credentialType
        ? (CREDENTIAL_TYPE_TO_KEY[credentialType] ?? this.environmentVariable.userCredential?.key ?? undefined)
        : undefined;

      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        userCredential: credentialName
          ? {
              name: credentialName,
              key,
            }
          : undefined,
      });
    },
    updateInstructionText(instructionText: string): void {
      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        description: instructionText,
      });
    },
    updateEnvironmentVariableName(name: string | number | null): void {
      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        name,
      });
    },
    updateSelectedSource(selectedSource: EEnvVarSourceType): void {
      this.selectedSource = selectedSource;
      this.selectedCredentialType = undefined;
      const updatedEnvVar: EnvironmentVariable = { ...this.environmentVariable };
      if (selectedSource === EEnvVarSourceType.Secrets) {
        updatedEnvVar.value = undefined;
        updatedEnvVar.configMap = undefined;
        updatedEnvVar.userCredential = undefined;
        updatedEnvVar.secret = {
          name: null,
          key: "",
        };
      } else if (selectedSource === EEnvVarSourceType.ConfigMap) {
        updatedEnvVar.value = undefined;
        updatedEnvVar.secret = undefined;
        updatedEnvVar.userCredential = undefined;
        updatedEnvVar.configMap = {
          name: null,
          key: "",
        };
      } else if (selectedSource === EEnvVarSourceType.UserCredential) {
        updatedEnvVar.value = undefined;
        updatedEnvVar.secret = undefined;
        updatedEnvVar.configMap = undefined;
        updatedEnvVar.userCredential = {
          name: null,
          key: "",
        };
      } else {
        updatedEnvVar.secret = undefined;
        updatedEnvVar.configMap = undefined;
        updatedEnvVar.userCredential = undefined;
      }

      this.$emit("update-environment-variable", updatedEnvVar);
    },
    updateEnvironmentVariableValue(value: string | number | null): void {
      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        value,
      });
    },
    updateEnvironmentVariableSecretName(identifier: string): void {
      const updatedEnvVar: EnvironmentVariable = { ...this.environmentVariable };

      const secretType = this.sourceValueOptions.find((c) => c.value === identifier)?.type;
      const key = secretType === SecretType.DockerRegistry ? DOCKER_REGISTRY_KEY : null;

      updatedEnvVar.secret = {
        name: identifier,
        key,
      };

      this.$emit("update-environment-variable", updatedEnvVar);
    },
    updateEnvironmentVariableSecretKey(key: string | number | null): void {
      const updatedEnvVar: EnvironmentVariable = { ...this.environmentVariable };

      updatedEnvVar.secret = {
        ...updatedEnvVar.secret,
        key: key?.toString(),
      };

      this.$emit("update-environment-variable", updatedEnvVar);
    },
    removeEnvironmentVariable(): void {
      this.$emit("remove-environment-variable", null);
    },
    // Rules
    isValidName(value: string): boolean | string {
      return isNotEmpty(value) || errorMessages.NAME_NOT_EMPTY;
    },
    isValidValue(value: string): boolean | string {
      // if we are not validating empty env vars, we can skip this validation
      if (!this.validateEmptyEnvVars) return true;

      return isNotEmpty(value) || errorMessages.VALUE_NOT_EMPTY;
    },
    isValidKey(value: string): boolean | string {
      return isNotEmpty(value) || errorMessages.KEY_NOT_EMPTY;
    },
    isNameSelected(value: string): boolean | string {
      return isNotEmpty(value) || errorMessages.SELECT_A_NAME;
    },
  },
});
</script>

<style lang="scss" scoped>
.environment-variable-box {
  background-color: $body-background-color;

  .selects-wrapper {
    gap: 10px;
  }

  .close-button {
    color: $black-54;
    font-size: 12px;
  }

  .credential-selector {
    flex-grow: 1;
    min-width: 260px;

    :deep(.credential-type-select) {
      flex: 1;
    }
    :deep(.credential-list-select) {
      max-width: 145px;
    }
  }
}
</style>
