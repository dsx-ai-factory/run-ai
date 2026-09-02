<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="credential-selector row">
    <runai-select
      v-if="!sourceSelectConfig.hidden"
      class="credential-source-select"
      :class="resolvedGrid.sourceSelect"
      :options="credentialSourceOptions"
      v-model="selectedCredentialSource"
      @update:model-value="onCredentialSourceChanged"
      emit-value
      aid="credential-source-select"
      label="Source"
      :disable="sourceSelectConfig.disable || disableAll"
    />

    <runai-select
      v-if="!typeSelectConfig.hidden"
      class="credential-type-select"
      :class="resolvedGrid.typeSelect"
      :options="credentialTypeOptions"
      v-model="selectedCredentialType"
      @update:model-value="onCredentialTypeChanged"
      emit-value
      aid="credential-type-select"
      label="Type"
      stack-label
      :disable="typeSelectConfig.disable || disableAll"
    />

    <runai-select
      ref="credentialListSelectRef"
      class="credential-list-select"
      :class="resolvedGrid.credentialListSelect"
      :options="filteredCredentialsOptions"
      :model-value="selectedCredentialValue"
      @update:model-value="onCredentialChanged"
      aid="credential-list-select"
      :loading="loading"
      :label="credentialSelectLabel"
      placeholder="Select..."
      no-option-text="No credentials found"
      stack-label
      :rules="validationRules"
      :custom-action="customAction"
      @custom-action="showPersonalCredentialModal = true"
      :disable="disableAll"
    />

    <credential-creation-modal
      v-if="showPersonalCredentialModal"
      :credential-type="personalCredentialType"
      :loading="isCreating"
      @cancel="showPersonalCredentialModal = false"
      @create="createCredential"
      :support-multiple-generic-secret-items="credentialSelectConfig.supportMultipleGenericSecretItems || false"
      :default-generic-secret-items="credentialSelectConfig.defaultGenericSecretItems"
      :disabled-keys-list="disabledKeysList"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent, type ComponentPublicInstance, type PropType } from "vue";

// components
import { RunaiSelect } from "@/components/common/runai-select";
import { CredentialCreationModal } from "@/components/user-settings/sections/personal-credentials-section/credential-creation-modal";

// models
import {
  ERunaiCredentialSource,
  ECredentialSelectorEvents,
  type ICredentialSelectorModel,
  ERunaiCredentialType,
  type ICredentialTypeConfig,
  type ISourceSelectConfig,
  type ITypeSelectConfig,
  type ICredentialSelectConfig,
  type IGridConfig,
  type TResolvedCredentialSelectorGrid,
  DEFAULT_CREDENTIAL_SELECTOR_GRID,
  type ICredentialOption,
  type TCredentialNameResolver,
} from "./runai-credential-selector-by-source.model";
import type { IScopeModel, ISelectOption } from "@/models/global.model";
import {
  AssetKind,
  type GenericSecretCreationRequest,
  type KeyValuePair,
  type NgcApiKeyCreationRequest,
  type DockerRegistryCreationRequest,
  type CredentialsListResponseEntry,
  AssetSyncStatusEnum,
} from "@/swagger-models/assets-service-client";
import type { QSelect } from "quasar";
import { type ValidationRule } from "quasar";
import type { TCredentialKinds } from "@/models/credential.model";
import type { IRunaiSelectCustomAction } from "@/components/common/runai-select/runai-select.vue";
import type { IAssetsFilter } from "@/models/filter.model";

// services
import { credentialService } from "@/services/control-plane/credential.service/credential.service";

// others
import { ErrorAlert } from "@/utils/error-alert.util/error-alert.util";
import { errorMessages } from "@/common/error-message.constant";
import { isEmpty } from "@/common/form.validators";

// stores
import { useClusterStore } from "@/stores/cluster.store";

// constants
import { MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS } from "@/common/version.constant";
import { dateUtil } from "@/utils/date.util";

// Credential Type Configuration
const CREDENTIAL_TYPE_CONFIG: Record<ERunaiCredentialType, ICredentialTypeConfig> = {
  [ERunaiCredentialType.NgcApiKey]: {
    label: "NGC API key",
    assetKind: AssetKind.NgcApiKey,
    createFn: async (model, service) => {
      await service.createNgcApiKey(model as NgcApiKeyCreationRequest);
    },
  },
  [ERunaiCredentialType.DockerRegistry]: {
    label: "Docker registry",
    assetKind: AssetKind.DockerRegistry,
    createFn: async (model, service) => {
      await service.createDockerRegistry(model as DockerRegistryCreationRequest);
    },
  },
  [ERunaiCredentialType.GenericSecret]: {
    label: "Generic secret",
    assetKind: AssetKind.GenericSecret,
    createFn: async (model, service) => {
      const prepared = service.prepareGenericSecretForServerRequest(model as GenericSecretCreationRequest);
      await service.createGeneralCredential(prepared);
    },
  },
};

const assetKindToCredentialType: Record<string, ERunaiCredentialType> = {
  [AssetKind.NgcApiKey]: ERunaiCredentialType.NgcApiKey,
  [AssetKind.GenericSecret]: ERunaiCredentialType.GenericSecret,
  [AssetKind.DockerRegistry]: ERunaiCredentialType.DockerRegistry,
};

const supportedTypes: Array<AssetKind> = [AssetKind.NgcApiKey, AssetKind.GenericSecret, AssetKind.DockerRegistry];

export default defineComponent({
  name: "runai-credential-selector-by-source",
  components: {
    RunaiSelect,
    CredentialCreationModal,
  },
  emits: {
    [ECredentialSelectorEvents.UpdateModelValue]: (modelValue: ICredentialSelectorModel) => modelValue,
  },
  props: {
    modelValue: {
      type: Object as PropType<ICredentialSelectorModel>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: false,
    },
    sourceSelectConfig: {
      type: Object as PropType<ISourceSelectConfig>,
      required: false,
      default: () => ({
        hidden: false,
        disable: false,
        showPersonalCredentialsOption: true,
      }),
    },
    typeSelectConfig: {
      type: Object as PropType<ITypeSelectConfig>,
      required: false,
      default: () => ({
        hidden: true,
        disable: false,
        allowedTypes: [
          ERunaiCredentialType.NgcApiKey,
          ERunaiCredentialType.GenericSecret,
          ERunaiCredentialType.DockerRegistry,
        ],
      }),
    },
    credentialSelectConfig: {
      type: Object as PropType<ICredentialSelectConfig>,
      required: false,
      default: () => ({
        required: true,
        showCreateNewBtn: false,
        supportMultipleGenericSecretItems: false,
      }),
    },
    gridConfig: {
      type: Object as PropType<IGridConfig>,
      required: false,
      default: () => ({}),
    },
    disableAll: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    nameResolver: {
      type: Function as PropType<TCredentialNameResolver>,
      default: undefined,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      allCredentialsOptions: [] as Array<ICredentialOption>,
      loading: false as boolean,
      selectedCredentialSource: ERunaiCredentialSource.MyCredentials as ERunaiCredentialSource, // there is a initial value for this
      credentialSourceOptions: [
        {
          value: ERunaiCredentialSource.SharedSecret,
          label: "Shared secret",
          toolTip: "Select a predefined Kubernetes secret",
        },
      ] as Array<ISelectOption>,
      credentialTypeOptions: [] as Array<ISelectOption>,
      validationRules: [] as Array<ValidationRule>,
      showPersonalCredentialModal: false as boolean,
      isCreating: false as boolean,
      selectedCredentialType: null as ERunaiCredentialType | null,
    };
  },
  async created() {
    this.initializeCredentialSource();
    this.initializeCredentialTypeOptions();
    this.initializeCredentialType();
    this.initializePersonalCredentialsOption();
    this.initializeValidationRules();
    await this.loadAllCredentials();
  },
  computed: {
    resolvedGrid(): TResolvedCredentialSelectorGrid {
      return {
        ...DEFAULT_CREDENTIAL_SELECTOR_GRID,
        ...this.gridConfig,
      };
    },
    filteredCredentialsOptions(): Array<ISelectOption> {
      return this.allCredentialsOptions
        .filter((credential: ICredentialOption) => {
          if (this.selectedCredentialSource !== credential.source) return false;
          if (this.selectedCredentialType && this.selectedCredentialType !== credential.type) return false;
          return true;
        })
        .map((credential: ICredentialOption) => {
          return {
            value: credential.id,
            label: credential.name,
            disable: credential.disable,
            disabledTooltip: credential.disabledTooltip,
          };
        });
    },
    personalCredentialType(): TCredentialKinds {
      const type = this.selectedCredentialType || ERunaiCredentialType.GenericSecret;
      return CREDENTIAL_TYPE_CONFIG[type].assetKind;
    },
    customAction(): IRunaiSelectCustomAction | undefined {
      if (
        !this.credentialSelectConfig.showCreateNewBtn ||
        this.selectedCredentialSource !== ERunaiCredentialSource.MyCredentials
      ) {
        return undefined;
      }
      return {
        text: "new credential",
      };
    },
    disabledKeysList(): string[] {
      return (
        this.credentialSelectConfig.defaultGenericSecretItems
          ?.filter((item) => item.key !== null && item.key !== undefined && item.key !== "")
          .map((item: KeyValuePair) => item.key || "") || []
      );
    },
    selectedCredentialValue(): string | null {
      // If we have an ID, use it
      if (this.modelValue.id) {
        return this.modelValue.id;
      }
      // If we only have a name, try to find the ID from the credentials list
      if (this.modelValue.credentialName) {
        const credential = this.allCredentialsOptions.find((cred) => cred.name === this.modelValue.credentialName);
        return (credential?.id as string) || null;
      }
      return null;
    },
    credentialSelectLabel(): string {
      if (!this.nameResolver) return "Credential name";
      return this.selectedCredentialSource === ERunaiCredentialSource.MyCredentials ? "Credential name" : "Secret name";
    },
  },
  methods: {
    async getCredentialByName(name: string): Promise<ICredentialOption | undefined> {
      let credential = this.allCredentialsOptions.find((cred) => cred.name === name);
      // If we don't find the credential, it might be a new one that was created after the component was created,
      // so we try to reload the credentials and find it again.
      if (!credential) {
        // Don't validate the source and type again, just find the credential by name, this is needed to prevent infinite loop
        await this.loadAllCredentials(false);
        credential = this.allCredentialsOptions.find((cred) => cred.name === name);
        if (!credential) {
          return;
        }
      }
      return credential;
    },
    validateSourceAndType(credential: ICredentialOption): void {
      this.selectedCredentialSource = credential.source;
      this.selectedCredentialType = credential.type;
      this.emitChanges({
        source: credential.source,
        type: credential.type,
        credentialName: credential.name,
        id: credential.id,
      });
    },
    // Provide a credential name if it was created recently to check its status.
    async loadAllCredentials(shouldValidate: boolean = true, recentlyCreatedCredentialName?: string): Promise<void> {
      let shouldStopLoading = true;
      try {
        this.loading = true;
        const allCredentials = await credentialService.listCredentials(this.getFilterBy());
        this.allCredentialsOptions = this.mapAndFilterToCredentialOptions(allCredentials);

        // If there is already a selected credential, validate its source and type.
        if (this.modelValue.credentialName && shouldValidate) {
          const credential = await this.getCredentialByName(this.modelValue.credentialName);
          if (credential) {
            this.validateSourceAndType(credential);
          }
        }

        if (recentlyCreatedCredentialName) {
          shouldStopLoading = this.pollCreatedCredentialStatus(recentlyCreatedCredentialName, allCredentials);
        }
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedLoadingMessage("secrets"),
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        shouldStopLoading && (this.loading = false);
      }
    },
    pollCreatedCredentialStatus(credentialName: string, allCredentials: Array<CredentialsListResponseEntry>): boolean {
      const credential = allCredentials.find((cred) => cred.meta.name === credentialName);
      if (!credential) return true;

      const createdAt = credential.meta.createdAt;
      const status = credential.status?.status;
      const isMoreThanMinuteAgo = dateUtil.isDurationMoreThanMinute((Date.now() - Date.parse(createdAt)) / 1000);

      if ((!status || status === AssetSyncStatusEnum.Creating) && !isMoreThanMinuteAgo) {
        setTimeout(() => this.loadAllCredentials(false, credentialName), 1000);
        return false;
      }

      ((this.$refs.credentialListSelectRef as ComponentPublicInstance)?.$refs.qSelectRef as QSelect)?.validate();
      return true;
    },
    mapAndFilterToCredentialOptions(credentials: Array<CredentialsListResponseEntry>): Array<ICredentialOption> {
      return credentials
        .filter((credential: CredentialsListResponseEntry) => {
          return supportedTypes.includes(credential.meta.kind);
        })
        .map((credential: CredentialsListResponseEntry) => {
          const name = this.nameResolver ? this.nameResolver(credential) : credential.meta.name;

          const disableByStatus =
            !credential.status?.status || credential.status?.status === AssetSyncStatusEnum.Creating ? true : false;

          return {
            id: credential.meta.id,
            name,
            source: credential.meta.private ? ERunaiCredentialSource.MyCredentials : ERunaiCredentialSource.SharedSecret,
            type: assetKindToCredentialType[credential.meta.kind],
            disable: disableByStatus,
            disabledTooltip: disableByStatus
              ? "This setup is being created. If it isn't created after a few seconds, contact your administrator."
              : undefined,
          };
        });
    },
    onCredentialSourceChanged(source: ERunaiCredentialSource): void {
      this.selectedCredentialSource = source;
      this.emitChanges({
        source: source,
        type: this.selectedCredentialType || undefined,
        credentialName: "",
        id: undefined,
      });
    },
    onCredentialTypeChanged(type: ERunaiCredentialType): void {
      this.selectedCredentialType = type;
      this.emitChanges({
        source: this.selectedCredentialSource,
        type: type,
        credentialName: "",
        id: undefined,
      });
    },
    getInitialCredentialType(): ERunaiCredentialType | null {
      // 1. Use existing type if already set
      if (this.modelValue.type) {
        return this.modelValue.type;
      }

      const firstNotDisabledType = this.credentialTypeOptions.find((type) => !type.disable);
      if (firstNotDisabledType) {
        return firstNotDisabledType.value as ERunaiCredentialType;
      }
      return ERunaiCredentialType.GenericSecret;
    },
    initializeCredentialSource(): void {
      this.selectedCredentialSource = this.modelValue.source || this.getDefaultCredentialSource();
    },
    initializeCredentialType(): void {
      this.selectedCredentialType = this.getInitialCredentialType();
    },
    initializeCredentialTypeOptions(): void {
      const allowedTypes = this.typeSelectConfig.allowedTypes || [
        ERunaiCredentialType.NgcApiKey,
        ERunaiCredentialType.GenericSecret,
        ERunaiCredentialType.DockerRegistry,
      ];
      this.credentialTypeOptions = allowedTypes.map((type) => {
        const config = CREDENTIAL_TYPE_CONFIG[type];
        const disabledConfig = this.typeSelectConfig.disabledTypes?.[type];
        return {
          value: type,
          label: config.label,
          disable: !!disabledConfig,
          disabledTooltip: disabledConfig?.tooltip,
        };
      });
    },
    initializePersonalCredentialsOption(): void {
      if (
        !this.sourceSelectConfig.hidden &&
        (this.sourceSelectConfig.showPersonalCredentialsOption ?? true) &&
        this.clusterStore.isClusterVersionSufficient(this.clusterId || null, MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS)
      ) {
        this.credentialSourceOptions.push({
          value: ERunaiCredentialSource.MyCredentials,
          label: "My credentials",
          toolTip: "These are only available to you. To add new credentials, go to User settings.",
        });
      }
    },
    initializeValidationRules(): void {
      const isRequired = this.credentialSelectConfig.required !== false;
      this.validationRules = isRequired ? [this.required, this.validCredentialStatus] : [this.validCredentialStatus];
    },
    onCredentialChanged(credential: ISelectOption): void {
      this.emitChanges({
        source: this.selectedCredentialSource,
        type: this.selectedCredentialType || undefined,
        credentialName: credential.label,
        id: credential.value as string,
      });
    },
    getFilterBy(): IAssetsFilter {
      return {
        projectId: this.scopeModel?.projectId,
        clusterUuid: this.scopeModel?.clusterId || undefined,
        departmentId: this.scopeModel?.departmentId,
        statusInfo: true,
        includePrivate: true,
      };
    },
    async afterCredentialCreated(credentialName: string): Promise<void> {
      await this.loadAllCredentials(false, credentialName);
      const credential = this.allCredentialsOptions.find((cred) => cred.name === credentialName);
      if (credential) {
        this.onCredentialChanged({ value: credential.id, label: credential.name });
      }
    },
    async createCredential(
      credentialModel: GenericSecretCreationRequest | NgcApiKeyCreationRequest | DockerRegistryCreationRequest,
    ): Promise<void> {
      try {
        this.isCreating = true;

        const type = this.selectedCredentialType || ERunaiCredentialType.GenericSecret;
        const config = CREDENTIAL_TYPE_CONFIG[type];

        await config.createFn(credentialModel, credentialService);

        this.$q.notify({
          message: `Credential ${credentialModel.meta.name} created`,
          color: "positive",
        });
        await this.afterCredentialCreated(credentialModel.meta.name);
        this.showPersonalCredentialModal = false;
      } catch (e: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: "Failed to create credential",
        });
        this.$q.notify(errorAlert.getNotification(e));
      } finally {
        this.isCreating = false;
      }
    },
    emitChanges(modelValue: ICredentialSelectorModel): void {
      this.$emit(ECredentialSelectorEvents.UpdateModelValue, modelValue);
    },
    required(value: string | null | undefined): string | boolean {
      return !isEmpty(value) || errorMessages.SELECT_CREDENTIAL;
    },
    validCredentialStatus(credentialId: string | null | undefined): boolean | string {
      // Not validating the credential status if no credential is selected.
      if (!credentialId) return true;

      const credential = this.allCredentialsOptions.find((cred) => cred.id === credentialId);
      if (!credential) return true;

      if (credential.disable) return "Loading credential. This may take a few moments.";
      return true;
    },
    getDefaultCredentialSource(): ERunaiCredentialSource {
      return this.sourceSelectConfig.showPersonalCredentialsOption &&
        this.credentialSourceOptions.find((option) => option.value === ERunaiCredentialSource.MyCredentials)
        ? ERunaiCredentialSource.MyCredentials
        : ERunaiCredentialSource.SharedSecret;
    },
  },
  watch: {
    modelValue: {
      async handler(newValue?: ICredentialSelectorModel, oldValue?: ICredentialSelectorModel): Promise<void> {
        if (newValue?.credentialName === oldValue?.credentialName) return;
        if (!newValue?.credentialName) return;
        const credential = await this.getCredentialByName(newValue?.credentialName || "");
        if (credential) {
          this.validateSourceAndType(credential);
          // If the credential is still in "creating" status, poll until it's ready
          if (credential.disable) {
            await this.loadAllCredentials(false, newValue.credentialName);
          }
        }
      },
      deep: true,
    },
  },
});
</script>
