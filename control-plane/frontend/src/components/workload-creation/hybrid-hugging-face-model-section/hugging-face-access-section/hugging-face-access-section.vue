<template>
  <section class="hugging-face-access-section">
    <runai-radio-options
      title="Set how to access Hugging Face"
      :model-value="modelValue.selectedTokenProvisionMethod"
      :options="tokenProvisionMethods"
      @update:model-value="onTokenProvisionMethodChanged"
    />
    <notification-alert-bar
      v-if="isManually"
      class="q-mt-md"
      :variant="ENotificationBarVariant.Warning"
      aid="hugging-face-plaintext-token-warning"
    >
      The token will be visible in the workload spec and via the API.
      <br />
      For secured storage, choose "Select credential" and enter the Hugging Face token as a new credential.
    </notification-alert-bar>
    <div class="row q-my-md italic">
      Note: Make sure you have been granted access to
      <a :href="huggingFaceModelInfo" target="_blank" class="q-ml-xs">the selected model</a>
    </div>
    <div class="q-mt-md">
      <div v-if="isManually" class="row">
        <policy-string-field
          class="credentials-token-input full-width"
          :ref="tokenInputRef"
          aid="credentials-token-input"
          @update:model-value="onTokenChanged"
          :model-value="modelValue.token"
          label="Access token"
          stack-label
          :debounce="500"
          :rules="[validateToken]"
          placeholder="e.g., hf_OLVVQaWA...NsBZROFHpX"
          input-class="placeholder-italic"
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
        :type-select-config="{ hidden: true, allowedTypes: [ERunaiCredentialType.GenericSecret] }"
        :credential-select-config="{
          required: isAccessRequired,
          showCreateNewBtn: showCreateNewPersonalCredentialsBtn,
          defaultGenericSecretItems: defaultGenericSecretItems,
        }"
        :grid-config="{ credentialListSelect: 'col-6' }"
        :name-resolver="credentialNameResolver"
      />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { QInput } from "quasar";

// components
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { RunaiCredentialSelectorBySource } from "@/components/common/runai-credential-selector-by-source";
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";

// models
import {
  EHuggingFaceAccessSectionEvents,
  ETokenProvisionMethods,
  type IHuggingFaceAccessSectionModel,
} from "./hugging-face-access-section.model";
import type { IScopeModel, ISelectOption } from "@/models/global.model";
import {
  ERunaiCredentialType,
  type ERunaiCredentialSource,
  type ICredentialSelectorModel,
  createResourceNameResolver,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import type { IHuggingFaceModelAndTokenValidationRes } from "./hugging-face-access-section.model";
import type {
  HuggingFaceModelInfo,
  HuggingFaceModelInfoRequest,
  KeyValuePair,
} from "@/swagger-models/assets-service-client";

// services
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";

// others
import { errorMessages } from "@/common/error-message.constant";
import { isEmpty, haggingFaceTokenValid } from "@/common/form.validators";
import { ErrorAlert } from "@/utils/error-alert.util";
import { EHuggingFaceEnvVariableNames } from "@/models/inference.model";

const credentialNameResolver = createResourceNameResolver();

export default defineComponent({
  name: "hugging-face-access-section",
  components: {
    RunaiRadioOptions,
    PolicyStringField,
    RunaiCredentialSelectorBySource,
    NotificationAlertBar,
  },
  emits: {
    [EHuggingFaceAccessSectionEvents.UpdateModelValue]: (modelValue: IHuggingFaceAccessSectionModel) => modelValue,
  },
  props: {
    modelValue: {
      type: Object as PropType<IHuggingFaceAccessSectionModel>,
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
    modelName: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      tokenProvisionMethods: [
        {
          value: ETokenProvisionMethods.FromSecret,
          label: "Select credential",
        },
        {
          value: ETokenProvisionMethods.Manually,
          label: "Provide a token",
        },
      ] as Array<ISelectOption>,
      localCredentialsSource: undefined as ERunaiCredentialSource | undefined,
      localSelectedCredentials: undefined as string | undefined,
      tokenInputRef: "tokenInput" as string,
      isValid: false as boolean,
      defaultGenericSecretItems: [{ key: EHuggingFaceEnvVariableNames.HF_TOKEN, value: "" }] as KeyValuePair[],
      ERunaiCredentialType,
      ENotificationBarVariant,
      credentialNameResolver,
    };
  },
  created() {
    this.localCredentialsSource = this.modelValue.credentialsSource;
    this.localSelectedCredentials = this.modelValue.selectedCredentials;

    if (this.modelValue.token) {
      this.invokeTokenValidation();
    }
  },
  computed: {
    isManually(): boolean {
      return this.modelValue.selectedTokenProvisionMethod === ETokenProvisionMethods.Manually;
    },
    credentialSelectorModel(): ICredentialSelectorModel {
      return {
        source: this.modelValue.credentialsSource || this.localCredentialsSource,
        credentialName: this.modelValue.selectedCredentials || "",
      };
    },
    huggingFaceModelInfo(): string {
      return `https://huggingface.co/${this.modelName}`;
    },
  },
  methods: {
    onTokenProvisionMethodChanged(tokenProvisionMethod: ETokenProvisionMethods): void {
      this.emitChanges({
        ...this.modelValue,
        token: undefined,
        selectedCredentials: undefined,
        selectedTokenProvisionMethod: tokenProvisionMethod,
      });
    },
    async onTokenChanged(token: string): Promise<void> {
      this.isValid = true;
      this.emitChanges({
        ...this.modelValue,
        token,
        selectedCredentials: undefined,
      });
    },
    onCredentialChanged(credential: ICredentialSelectorModel): void {
      this.localSelectedCredentials = credential.credentialName;
      this.localCredentialsSource = credential.source;
      this.emitChanges({
        ...this.modelValue,
        token: undefined,
        selectedCredentials: credential.credentialName,
        credentialsSource: credential.source,
      });
    },
    emitChanges(modelValue: IHuggingFaceAccessSectionModel): void {
      this.$emit(EHuggingFaceAccessSectionEvents.UpdateModelValue, modelValue);
    },
    async verifyModelWithToken(
      huggingFaceModelInfoRequest: HuggingFaceModelInfoRequest,
    ): Promise<IHuggingFaceModelAndTokenValidationRes> {
      const modelValidationRes: IHuggingFaceModelAndTokenValidationRes = {
        modelName: huggingFaceModelInfoRequest.modelName,
        hfToken: huggingFaceModelInfoRequest.hfToken || "",
        gated: true,
      };
      try {
        if (isEmpty(modelValidationRes.hfToken)) {
          modelValidationRes.errorMessage = errorMessages.ENTER_A_TOKEN;
        } else if (!haggingFaceTokenValid(huggingFaceModelInfoRequest.hfToken || "")) {
          modelValidationRes.errorMessage = errorMessages.ENTER_A_VALID_TOKEN;
        } else {
          const huggingFaceModelInfo: HuggingFaceModelInfo =
            await inferenceService.getHuggingFaceModelInfo(huggingFaceModelInfoRequest);
          if (!huggingFaceModelInfo.authorized) {
            modelValidationRes.errorMessage = errorMessages.TOKEN_IS_NOT_AUTHORIZED;
          }
        }
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedToValidateMessage("model or token"),
        });
        this.$q.notify(errorAlert.getNotification(error));
      }
      return modelValidationRes;
    },
    async validateToken(val: string): Promise<string | boolean> {
      if (!this.isAccessRequired) {
        return true;
      }
      const validationResults: IHuggingFaceModelAndTokenValidationRes = await this.verifyModelWithToken({
        modelName: this.modelValue.modelName,
        hfToken: val,
      });
      return validationResults.errorMessage || true;
    },
    invokeTokenValidation(): void {
      this.$nextTick(() => {
        (this.$refs[this.tokenInputRef] as QInput)?.validate();
      });
    },
  },
});
</script>

<style scoped lang="scss"></style>
