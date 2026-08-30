<template>
  <section class="nim-access-model">
    <runai-section title="Set the access to the NGC catalog">
      <q-checkbox
        aid="access-from-storage"
        v-model="accessFromStorage"
        @update:model-value="onAccessFromStorageChanged"
        label="Do not access NGC. Load the model profile from a local model-store"
        class="offset-left-8"
      />
      <runai-tooltip tooltip-text="Go to the Model store section to set this" tooltip-position="right" />
    </runai-section>

    <div v-if="!accessFromStorage" class="access-method q-mt-lg">
      <runai-section title="Set how to access NGC catalog">
        <runai-radio-options
          :model-value="selectedAccessMethod"
          :options="accessOptions"
          @update:model-value="onAccessMethodChanged"
          tooltip-width="400px"
        />

        <div v-if="isTokenProvidedManually" class="row q-mt-md">
          <policy-string-field
            class="col-9"
            aid="ngc-api-key-input"
            @update:model-value="onApiKeyChanged"
            :model-value="modelValue.apiKey"
            label="API key"
            stack-label
            input-class="placeholder-italic"
            :rules="apiKeyValidationRules"
            placeholder="e.g., nvapi-nFCM...MxGUgyU109fjduery7"
          />
        </div>

        <runai-credential-selector-by-source
          v-else
          :model-value="credentialSelectorModel"
          :cluster-id="clusterId"
          @update:model-value="onCredentialChanged"
          :scope-model="scopeModel"
          :source-select-config="{
            disable: !showPersonalCredentialsOption,
            showPersonalCredentialsOption: showPersonalCredentialsOption,
          }"
          :type-select-config="{
            allowedTypes: [ERunaiCredentialType.NgcApiKey, ERunaiCredentialType.GenericSecret],
            disabledTypes: !supportingNgcApiKey
              ? {
                  [ERunaiCredentialType.NgcApiKey]: {
                    tooltip:
                      'This credential type is supported only for NIM inference workloads created on clusters version 2.23 or higher',
                  },
                }
              : undefined,
          }"
          :credential-select-config="{
            required: isAccessRequired,
            showCreateNewBtn: showCreateNewPersonalCredentialsBtn,
            defaultGenericSecretItems: defaultGenericSecretItems,
          }"
          :name-resolver="credentialNameResolver"
        />
      </runai-section>
    </div>

    <transition name="fade">
      <div v-if="showDockerRegistry" class="docker-registry-selector-container q-mt-lg">
        <docker-registry-selector
          :model-value="dockerRegistrySelectorModel"
          :cluster-id="clusterId"
          :scope-model="scopeModel"
          @update:model-value="onDockerRegistryChanged"
        />
      </div>
    </transition>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { RunaiSection } from "@/components/common/runai-section";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiCredentialSelectorBySource } from "@/components/common/runai-credential-selector-by-source";
import { DockerRegistrySelector } from "./docker-registry-selector";

// models
import { ENimAccessModelMethod, ENimAccessModelEvents, type INimAccessModelModel } from "./nim-access-model.model";
import type { IScopeModel, ISelectOption } from "@/models/global.model";
import { ENimEnvVariableNames } from "@/models/inference.model";
import {
  ERunaiCredentialType,
  ERunaiCredentialSource,
  type ICredentialSelectorModel,
  createResourceNameResolver,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import { AssetKind } from "@/swagger-models/assets-service-client";
import type { ValidationRule } from "quasar";
import type { IDockerRegistrySelectorModel } from "./docker-registry-selector/docker-registry-selector.model";

// others
import { errorMessages } from "@/common/error-message.constant";
import { isNotEmpty, nimProfileValidator } from "@/common/form.validators";
import type { KeyValuePair } from "@/swagger-models/assets-service-client";

const credentialNameResolver = createResourceNameResolver(AssetKind.GenericSecret);

export default defineComponent({
  name: "nim-access-model",
  components: {
    RunaiSection,
    RunaiRadioOptions,
    PolicyStringField,
    RunaiTooltip,
    RunaiCredentialSelectorBySource,
    DockerRegistrySelector,
  },
  emits: {
    [ENimAccessModelEvents.UpdateModelValue]: (modelValue: INimAccessModelModel) => modelValue,
  },
  props: {
    modelValue: {
      type: Object as PropType<INimAccessModelModel>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    isAccessRequired: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
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
    supportingNgcApiKey: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    supportsDockerRegistry: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  data() {
    return {
      localApiKey: "" as string,
      accessOptions: [
        {
          value: ENimAccessModelMethod.FromSecret,
          label: "Select credential",
        },
        {
          value: ENimAccessModelMethod.Manually,
          label: "Provide a key",
        },
      ] as Array<ISelectOption>,
      accessFromStorage: false,
      selectedAccessMethod: ENimAccessModelMethod.FromSecret as ENimAccessModelMethod,
      credentialTooltip: `The credential must contain a key named "${ENimEnvVariableNames.NgcApiKey}"` as string,
      localSelectedCredentials: undefined as string | undefined,
      localCredentialsSource: undefined as ERunaiCredentialSource | undefined,
      apiKeyValidationRules: [] as Array<ValidationRule>,
      defaultGenericSecretItems: [{ key: ENimEnvVariableNames.NgcApiKey, value: "" }] as KeyValuePair[],
      ERunaiCredentialType,
      credentialNameResolver,
    };
  },
  created() {
    this.accessFromStorage = this.modelValue.accessFromStorage ?? false;
    this.initializeLocalCredentialsSource();

    if (!this.accessFromStorage) {
      this.selectedAccessMethod = this.modelValue.apiKey
        ? ENimAccessModelMethod.Manually
        : ENimAccessModelMethod.FromSecret;
    }

    if (this.isAccessRequired) {
      this.apiKeyValidationRules = [this.apiKeyNotEmpty, this.apiKeyValid];
    } else {
      this.apiKeyValidationRules = [this.apiKeyNotRequired];
    }
  },
  computed: {
    selectedCredentials(): Array<string> {
      return this.modelValue.selectedCredentials ? [this.modelValue.selectedCredentials] : [];
    },
    isTokenProvidedManually(): boolean {
      return this.selectedAccessMethod === ENimAccessModelMethod.Manually;
    },
    credentialSelectorModel(): ICredentialSelectorModel {
      return {
        source: this.localCredentialsSource,
        type: this.modelValue.credentialType,
        credentialName: this.modelValue.selectedCredentials || "",
      };
    },
    dockerRegistrySelectorModel(): IDockerRegistrySelectorModel {
      return {
        dockerRegistryName: this.modelValue.dockerRegistry?.dockerRegistryName || "",
      };
    },
    showDockerRegistry(): boolean {
      if (!this.supportsDockerRegistry) return false;
      if (this.selectedAccessMethod === ENimAccessModelMethod.Manually) return true;
      if (this.modelValue.credentialType === ERunaiCredentialType.NgcApiKey || !this.modelValue.credentialType) {
        return false;
      }

      return true;
    },
  },
  methods: {
    initializeLocalCredentialsSource(): void {
      if (this.modelValue.credentialsSource !== undefined) {
        this.localCredentialsSource = this.modelValue.credentialsSource;
      } else {
        this.localCredentialsSource = this.showPersonalCredentialsOption
          ? ERunaiCredentialSource.MyCredentials
          : ERunaiCredentialSource.SharedSecret;
      }
    },
    onDockerRegistryChanged(dockerRegistry: IDockerRegistrySelectorModel): void {
      this.emitChanges({
        ...this.modelValue,
        dockerRegistry: dockerRegistry,
      });
    },
    onAccessMethodChanged(accessModelMethod: ENimAccessModelMethod): void {
      this.selectedAccessMethod = accessModelMethod;
      if (accessModelMethod === ENimAccessModelMethod.Manually) {
        this.localSelectedCredentials = this.modelValue.selectedCredentials;
        this.localCredentialsSource = this.modelValue.credentialsSource;
        this.emitChanges({
          ...this.modelValue,
          apiKey: this.localApiKey,
          selectedCredentials: undefined,
          credentialsSource: undefined,
        });
      } else {
        this.localApiKey = this.modelValue.apiKey || "";
        this.emitChanges({
          ...this.modelValue,
          apiKey: undefined,
          selectedCredentials: this.localSelectedCredentials,
          credentialsSource: this.localCredentialsSource,
        });
      }
    },
    onApiKeyChanged(apiKey: string): void {
      this.localApiKey = apiKey;
      this.emitChanges({
        ...this.modelValue,
        apiKey,
        selectedCredentials: undefined,
        credentialsSource: undefined,
      });
    },
    apiKeyNotEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.KEY_NOT_EMPTY;
    },
    apiKeyValid(val: string): boolean | string {
      return nimProfileValidator(val) || errorMessages.ENTER_A_VALID_NIM_API_KEY;
    },
    apiKeyNotRequired(val: string): boolean | string {
      return val ? this.apiKeyValid(val) : true;
    },
    onAccessFromStorageChanged(accessFromStorage: boolean): void {
      const accessModelMethod: ENimAccessModelMethod | undefined = !accessFromStorage
        ? this.selectedAccessMethod
        : undefined;

      let apiKey = undefined;
      let selectedCredentials = undefined;
      let credentialsSource = undefined;
      if (!accessModelMethod) {
        if (this.selectedAccessMethod === ENimAccessModelMethod.FromSecret) {
          this.localSelectedCredentials = this.modelValue.selectedCredentials;
          this.localCredentialsSource = this.modelValue.credentialsSource;
        } else if (this.selectedAccessMethod === ENimAccessModelMethod.Manually) {
          this.localApiKey = this.modelValue.apiKey || "";
        }
      } else if (accessModelMethod === ENimAccessModelMethod.FromSecret) {
        selectedCredentials = this.localSelectedCredentials;
        credentialsSource = this.localCredentialsSource;
      } else if (accessModelMethod === ENimAccessModelMethod.Manually) {
        apiKey = this.localApiKey;
      }

      this.emitChanges({
        ...this.modelValue,
        apiKey,
        selectedCredentials,
        credentialsSource,
        accessFromStorage,
      });
    },
    onCredentialChanged(credential: ICredentialSelectorModel): void {
      this.localSelectedCredentials = credential.credentialName;
      this.localCredentialsSource = credential.source;
      let dockerRegistryName = "";
      if (
        credential.type === ERunaiCredentialType.NgcApiKey &&
        credential.source === ERunaiCredentialSource.MyCredentials
      ) {
        dockerRegistryName = credential.credentialName;
      }
      this.emitChanges({
        ...this.modelValue,
        selectedCredentials: credential.credentialName,
        credentialsSource: credential.source,
        credentialType: credential.type,
        dockerRegistry: {
          dockerRegistryName: dockerRegistryName,
        },
      });
    },
    emitChanges(modelValue: INimAccessModelModel): void {
      this.$emit(ENimAccessModelEvents.UpdateModelValue, modelValue);
    },
  },
});
</script>

<style scoped lang="scss"></style>
