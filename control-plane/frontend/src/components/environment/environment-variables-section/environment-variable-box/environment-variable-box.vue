<template>
  <section class="environment-variable-box q-pa-md row items-center">
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
      class="col-2 q-mr-md"
      :options="sourceOptions"
      label="Source"
      emit-value
      stack-label
      unclearable
      :model-value="selectedSource"
      @update:model-value="updateSelectedSource"
      :rules="[() => true]"
      :disable="readOnly"
      selected-ellipsis
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
    <template v-else-if="selectedSource === 'credentials'">
      <runai-select
        aid="environment-variable-credential"
        :style="{ width: '150px' }"
        class="q-mr-md"
        label="Credential name"
        placeholder="Select..."
        stack-label
        emit-value
        :options="credentialOptions"
        :model-value="environmentVariable.credential?.assetId || null"
        @update:model-value="updateEnvironmentVariableCredentialId"
        :rules="[isNameSelected]"
        no-option-text="No credentials found"
        :disable="readOnly"
        selected-ellipsis
      />

      <template v-if="!!environmentVariable.credential?.assetId">
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
          :model-value="environmentVariable.credential?.key || null"
          @update:model-value="updateEnvironmentVariableKey"
          no-option-text="Invalid credentials"
          selected-ellipsis
        />
        <q-input
          v-else-if="!isDockerRegistrySelected"
          aid="environment-variable-key-input"
          :style="{ width: '150px' }"
          stack-label
          placeholder="Enter a key"
          label="Secret key"
          :model-value="environmentVariable.credential?.key"
          @update:model-value="updateEnvironmentVariableKey"
          no-error-icon
          :rules="[isValidKey]"
          :disable="readOnly"
        />
      </template>
    </template>
    <template v-else-if="selectedSource === 'configMap'">
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
        selected-ellipsis
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
        :rules="[isValidKey]"
        :disable="readOnly"
      />
    </template>
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

    <q-space />
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
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiSelect } from "@/components/common/runai-select";
// Models
import type { ISelectOption } from "@/models/global.model";
import { AssetKind, type EnvironmentVariableOfAsset } from "@/swagger-models/assets-service-client";
// Utils
import { errorMessages } from "@/common/error-message.constant";
import { isNotEmpty } from "@/common/form.validators";
import { type TEnvVarNoneCustomSourceOption } from "@/models/environment.model";
// cmps
import { RunaiInstructionsWindowWithPoupOver } from "@/components/common/runai-instructions-window/with-popover";

export enum ESourceType {
  Custom = "custom",
  Credentials = "credentials",
  ConfigMap = "configMap",
}

export enum EAccessKeySecretKey {
  Key = "AccessKeyId",
  Secret = "SecretKey",
}
export enum EUserPassSecretKey {
  Pass = "Password",
  User = "User",
}

export default defineComponent({
  name: "environment-variable-box",
  components: {
    RunaiInstructionsWindowWithPoupOver,
    RunaiSelect,
  },
  emits: ["update-environment-variable", "remove-environment-variable"],
  props: {
    environmentVariable: {
      type: Object as PropType<EnvironmentVariableOfAsset>,
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
  },
  data() {
    return {
      selectedSource: ESourceType.Custom as ESourceType,
    };
  },
  created() {
    this.selectedSource = this.environmentVariable.credential
      ? ESourceType.Credentials
      : this.environmentVariable.configMap
        ? ESourceType.ConfigMap
        : ESourceType.Custom;
  },
  computed: {
    sourceOptions(): ISelectOption[] {
      const options: ISelectOption[] = [
        {
          value: ESourceType.Custom,
          label: "Custom",
          toolTip: "Enter a name and value for the environment variable",
        },
        {
          value: ESourceType.Credentials,
          label: "Credentials",
          toolTip: "Select an existing credential as the environment variable",
        },
      ];

      if (this.supportConfigMap) {
        options.push({
          value: ESourceType.ConfigMap,
          label: "ConfigMap",
          toolTip: "Select a predefined ConfigMap",
        });
      }

      return options;
    },
    showKeySelect(): boolean {
      const kind = this.sourceValueOptions.find((c) => c.value === this.environmentVariable.credential?.assetId)?.type;
      return kind === AssetKind.AccessKey || kind === AssetKind.Password;
    },
    isDockerRegistrySelected(): boolean {
      return (
        this.sourceValueOptions.find((c) => c.value === this.environmentVariable.credential?.assetId)?.type ===
        AssetKind.DockerRegistry
      );
    },
    showInstructionLogo(): boolean {
      // we will show the instruction btn if the user is in create/edit asset
      // or
      // if there is a description that we need to show to the user
      return this.editInstructionText || !!this.environmentVariable.description;
    },
    secretKeyOptions(): EAccessKeySecretKey[] | EUserPassSecretKey[] {
      const kind = this.sourceValueOptions.find((c) => c.value === this.environmentVariable.credential?.assetId)?.type;
      switch (kind) {
        case AssetKind.AccessKey:
          return [EAccessKeySecretKey.Key, EAccessKeySecretKey.Secret];
        case AssetKind.Password:
          return [EUserPassSecretKey.Pass, EUserPassSecretKey.User];
        default:
          return [];
      }
    },
    configMapOptions(): ISelectOption[] {
      return this.sourceValueOptions
        .filter((c) => c.type === AssetKind.ConfigMap)
        .map((c) => ({
          value: c.value,
          label: c.label,
        }));
    },
    credentialOptions(): ISelectOption[] {
      return this.sourceValueOptions
        .filter((c) => c.type !== AssetKind.ConfigMap)
        .map((c) => ({
          value: c.value,
          label: c.label,
        }));
    },
    getEnvironmentVariableHint(): string {
      return this.editInstructionText ? "Enter a value, or leave it empty and provide instructions" : "";
    },
    getInstructionIcon(): string {
      return "ph-fill ph-note";
    },
  },
  methods: {
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
    updateSelectedSource(selectedSource: ESourceType): void {
      this.selectedSource = selectedSource;
      if (selectedSource === ESourceType.Custom) {
        this.$emit("update-environment-variable", {
          ...this.environmentVariable,
          credential: undefined,
          configMap: undefined,
        });
      } else if (selectedSource === ESourceType.Credentials) {
        this.$emit("update-environment-variable", {
          ...this.environmentVariable,
          value: undefined,
          configMap: undefined,
          credential: {
            assetId: null,
            key: "",
          },
        });
      } else if (selectedSource === ESourceType.ConfigMap) {
        this.$emit("update-environment-variable", {
          ...this.environmentVariable,
          value: undefined,
          credential: undefined,
          configMap: {
            name: null,
          },
        });
      }
    },
    updateEnvironmentVariableValue(value: string | number | null): void {
      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        value,
      });
    },
    updateEnvironmentVariableCredentialId(assetId: string): void {
      const kind = this.sourceValueOptions.find((c) => c.value === assetId)?.type;
      const key = kind === AssetKind.DockerRegistry ? ".dockerconfigjson" : null;
      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        credential: {
          key,
          assetId,
        },
      });
    },
    updateEnvironmentVariableConfigMapName(configMapName: string): void {
      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        configMap: {
          name: configMapName,
        },
      });
    },
    updateEnvironmentVariableKey(key: string | number | null): void {
      this.$emit("update-environment-variable", {
        ...this.environmentVariable,
        credential: {
          ...this.environmentVariable.credential,
          key,
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

  .close-button {
    color: $black-54;
    font-size: 12px;
  }
}
</style>
