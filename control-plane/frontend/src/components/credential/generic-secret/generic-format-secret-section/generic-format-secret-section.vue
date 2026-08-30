<template>
  <runai-expansion-item class="generic-format-secret-section" default-opened label="Secret" should-validate>
    <template #subheader>
      <section class="summary row">
        {{ summary }}
      </section>
    </template>

    <runai-section title="Set how the credential is created">
      <runai-radio-options
        aid="secret-format-radio-options"
        :options="radioOptions"
        :model-value="isExistingSecret"
        @update:model-value="updateFormState"
        horizontal
      />
    </runai-section>

    <section class="existing-secret-section row" v-if="isExistingSecret">
      <runai-select
        class="secret-select col-5"
        standard
        :model-value="secretDetails.existingSecretName || null"
        :options="existingSecretOptions"
        @update:model-value="updateExistingSecret"
        :rules="[validationRules.validateValueSelected]"
        aid="existing-secret-select"
        :loading="loading"
        no-option-text="No secrets found"
        placeholder="Select a secret..."
      />
    </section>

    <section class="new-secret-section row q-mt-md" v-else>
      <runai-section title="Set the secret format">
        <div class="row">
          <runai-select
            class="col-4"
            aid="secret-format-select"
            :options="supportedFormatOptions"
            :model-value="selectedFormat"
            emit-value
            placeholder="Select a format..."
            :rules="[validationRules.validateValueSelected]"
            @update:model-value="updateSelectedFormat"
          />
        </div>
      </runai-section>
      <runai-section>
        <generic-secret-key-value-section
          v-if="selectedFormat === AssetKind.GenericSecret"
          :secret-details="secretDetails"
          @update-secret-details="$emit('update-secret-details', $event)"
        />

        <div v-else-if="selectedFormat === TYPE_SSH_KEY">
          <q-input
            class="col-5"
            :model-value="secretDetails.keyValuePairs?.[0]?.value || ''"
            no-error-icon
            label="SSH Key"
            :rules="[validationRules.validateKeyNotEmpty]"
            aid="generic-secret-ssh-key-input"
            @update:model-value="updateSshKeyValue"
          />
        </div>

        <generic-secret-input-pair
          :key="selectedFormat"
          v-else-if="selectedFormat === AssetKind.AccessKey || selectedFormat === AssetKind.Password"
          :secret-pair-details="secretPairDetailsProps"
          @update="updateSecretPairDetails"
        />
      </runai-section>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSelect } from "@/components/common/runai-select";
import { GenericSecretKeyValueSection } from "../generic-secret-secret-section/generic-secret-key-value-section";
import { RunaiSection } from "@/components/common/runai-section";
import { GenericSecretInputPair } from "@/components/credential/generic-secret/generic-secret-input-pair";
// types
import type { ISelectOption } from "@/models/global.model";
import { errorMessages } from "@/common/error-message.constant";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { TYPE_SSH_KEY, type IUIGenericSecretSpecDetails, type TGenericFormats } from "@/models/credential.model";

import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { validationRules } from "@/common/validation-rules";
import type { ValidationRule } from "quasar";
import type { GenericSecretInputPairProps } from "../generic-secret-input-pair/generic-secret-input-pair.vue";
import type { SecretListInfo } from "@/swagger-models/k8s-objects-tracker-client";

const GENERIC_FORMAT_LABELS: Record<TGenericFormats, string> = {
  [AssetKind.GenericSecret]: "Key(s) & value(s)",
  [AssetKind.AccessKey]: "Access key",
  [AssetKind.Password]: "Username & password",
  [TYPE_SSH_KEY]: "SSH key",
};

const SECRET_FORMAT_OPTIONS: Array<{ label: (typeof GENERIC_FORMAT_LABELS)[TGenericFormats]; value: TGenericFormats }> =
  [
    {
      label: GENERIC_FORMAT_LABELS[AssetKind.GenericSecret],
      value: AssetKind.GenericSecret,
    },
    {
      label: GENERIC_FORMAT_LABELS[AssetKind.AccessKey],
      value: AssetKind.AccessKey,
    },
    {
      label: GENERIC_FORMAT_LABELS[AssetKind.Password],
      value: AssetKind.Password,
    },
    {
      label: GENERIC_FORMAT_LABELS[TYPE_SSH_KEY],
      value: TYPE_SSH_KEY,
    },
  ];

export default defineComponent({
  name: "generic-format-secret-section",
  components: {
    RunaiExpansionItem,
    RunaiSelect,
    RunaiRadioOptions,
    GenericSecretKeyValueSection,
    RunaiSection,
    GenericSecretInputPair,
  },
  emits: ["update-secret-details"],
  props: {
    secretDetails: {
      type: Object as PropType<IUIGenericSecretSpecDetails>,
      required: true,
    },
    existingSecrets: {
      type: [Array, null] as PropType<SecretListInfo[] | null>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    disabledFormats: {
      type: Array as PropType<TGenericFormats[]>,
      required: false,
      default: () => [],
    },
  },
  data() {
    return {
      isExistingSecret: true as boolean,
      errorMessages,
      validationRules,
      AssetKind,
      TYPE_SSH_KEY,
      SECRET_FORMAT_OPTIONS,

      selectedFormat: null as TGenericFormats | null,
    };
  },
  computed: {
    supportedFormatOptions(): Array<{ label: (typeof GENERIC_FORMAT_LABELS)[TGenericFormats]; value: TGenericFormats }> {
      return this.SECRET_FORMAT_OPTIONS.filter((option) => !this.disabledFormats.includes(option.value));
    },
    radioOptions(): Array<ISelectOption> {
      return [
        {
          label: "Existing secret",
          value: true,
        },
        {
          label: "New secret",
          value: false,
        },
      ];
    },
    summary(): string {
      if (this.isExistingSecret) return this.secretDetails.existingSecretName || "None";
      return this.SECRET_FORMAT_OPTIONS.find((option) => option.value === this.selectedFormat)?.label || "None";
    },
    identifierInputValue(): string {
      if (this.selectedFormat === AssetKind.AccessKey) return this.secretDetails.accessKeyId || "";
      if (this.selectedFormat === AssetKind.Password) return this.secretDetails.user || "";
      return "";
    },
    secretInputValue(): string {
      if (this.selectedFormat === AssetKind.AccessKey) return this.secretDetails.secretAccessKey || "";
      if (this.selectedFormat === AssetKind.Password) return this.secretDetails.password || "";
      return "";
    },
    identifierInputLabel(): string {
      if (this.selectedFormat === AssetKind.AccessKey) return "Access key";
      if (this.selectedFormat === AssetKind.Password) return "Username";
      return "";
    },
    secretInputLabel(): string {
      if (this.selectedFormat === AssetKind.AccessKey) return "Access secret";
      if (this.selectedFormat === AssetKind.Password) return "Password";
      return "";
    },
    identifierInputRules(): Array<ValidationRule> {
      if (this.selectedFormat === AssetKind.AccessKey) return [validationRules.validateKeyNotEmpty];
      if (this.selectedFormat === AssetKind.Password) return [validationRules.validateUsernameNotEmpty];
      return [];
    },
    secretInputRules(): Array<ValidationRule> {
      if (this.selectedFormat === AssetKind.AccessKey) return [validationRules.validateSecretNotEmpty];
      if (this.selectedFormat === AssetKind.Password) return [validationRules.validatePasswordNotEmpty];
      return [];
    },
    secretPairDetailsProps(): GenericSecretInputPairProps {
      return {
        identifierInput: {
          value: this.identifierInputValue,
          label: this.identifierInputLabel,
          rules: this.identifierInputRules,
        },
        secretInput: {
          value: this.secretInputValue,
          label: this.secretInputLabel,
          rules: this.secretInputRules,
        },
      };
    },
    existingSecretOptions(): string[] {
      return this.existingSecrets?.map((s) => s.name) || [];
    },
  },
  methods: {
    updateExistingSecret(existingSecret: string): void {
      this.$emit("update-secret-details", { ...this.secretDetails, existingSecretName: existingSecret });
    },
    updateFormState(isExisting: boolean): void {
      this.isExistingSecret = isExisting;
      this.selectedFormat = null;

      const updatedModel: IUIGenericSecretSpecDetails = {};
      this.$emit("update-secret-details", updatedModel);
    },
    updateSelectedFormat(format: TGenericFormats): void {
      this.selectedFormat = format;
      const updatedModel: IUIGenericSecretSpecDetails = {};
      if (format === AssetKind.GenericSecret) {
        updatedModel.keyValuePairs = [{ key: "", value: "" }];
      } else if (format === AssetKind.AccessKey) {
        updatedModel.accessKeyId = "";
        updatedModel.secretAccessKey = "";
      } else if (format === AssetKind.Password) {
        updatedModel.user = "";
        updatedModel.password = "";
      } else if (format === TYPE_SSH_KEY) {
        updatedModel.keyValuePairs = [{ key: "ssh", value: "" }];
      }
      this.$emit("update-secret-details", updatedModel);
    },
    updateSecretPairDetails(updatedSecretPairDetails: { identifier: string; secret: string }): void {
      const updatedModel: IUIGenericSecretSpecDetails = {};
      if (this.selectedFormat === AssetKind.AccessKey) {
        updatedModel.accessKeyId = updatedSecretPairDetails.identifier || "";
        updatedModel.secretAccessKey = updatedSecretPairDetails.secret;
      } else if (this.selectedFormat === AssetKind.Password) {
        updatedModel.user = updatedSecretPairDetails.identifier || "";
        updatedModel.password = updatedSecretPairDetails.secret;
      }

      this.$emit("update-secret-details", updatedModel);
    },
    updateSshKeyValue(value: string | null | number): void {
      this.$emit("update-secret-details", {
        ...this.secretDetails,
        keyValuePairs: [{ key: "ssh", value: value || "" }],
      });
    },
  },
});
</script>
