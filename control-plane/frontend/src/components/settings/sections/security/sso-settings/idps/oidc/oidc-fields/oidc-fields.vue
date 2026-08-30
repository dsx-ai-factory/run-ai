<template>
  <div class="idp-setup-fields col-12">
    <div>
      <div>Add your identity provider metadata</div>
      <div class="row col-12">
        <div class="row col-12">
          <setting-readonly-input class="col-12" :resource-type="securityResourceType">
            <template v-slot:default="{ readonly }">
              <!-- eslint-disable vuejs-accessibility/no-autofocus -- this whole field set only renders once the user clicks Edit on the IDP settings box (v-if="isEditMode" in oidc-settings.vue), never on a bare page load -->
              <q-input
                class="col-12 q-mb-lg"
                :model-value="oidcSettingsFields.discoverDocumentUrl"
                placeholder="Enter URL"
                debounce="300"
                :rules="[isValidUrl]"
                no-error-icon
                @update:model-value="updateDiscoveryUrl"
                stack-label
                label="Discovery URL"
                ref="oidcUrlInput"
                autofocus
                :hint="urlHint"
                :disable="readonly || disabled"
              />
              <!-- eslint-enable vuejs-accessibility/no-autofocus -->
            </template>
          </setting-readonly-input>
        </div>
        <div class="row col-12 client-inputs">
          <setting-readonly-input class="col-5" :resource-type="securityResourceType">
            <template v-slot:default="{ readonly }">
              <q-input
                class="col-5"
                :model-value="oidcSettingsFields.clientId"
                debounce="300"
                :rules="[isNotEmptyClientID]"
                no-error-icon
                @update:model-value="updateClientId"
                stack-label
                label="Client ID"
                ref="oidcClientIdInput"
                :disable="readonly || disabled"
              />
            </template>
          </setting-readonly-input>
          <div class="col-1" />
          <setting-readonly-input class="col-6" :resource-type="securityResourceType">
            <template v-slot:default="{ readonly }">
              <q-input
                class="col-6 items-end"
                :model-value="oidcSettingsFields.clientSecret"
                debounce="300"
                :rules="[isNotEmptySecret]"
                no-error-icon
                @update:model-value="updateClientSecret"
                stack-label
                label="Client secret"
                ref="oidcClientSecretInput"
                :disable="readonly || disabled"
              />
            </template>
          </setting-readonly-input>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { isNotEmpty, isValidUrl } from "@/common/form.validators";
import type { QInput } from "quasar";
import type { OidcCreationData } from "@/swagger-models/identity-manager-client";
import { urlHintMessage } from "@/models/setting.model";
import { SettingReadonlyInput } from "@/components/settings/setting-readonly-input";
import { errorMessages } from "@/common/error-message.constant";
import type { ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "oidc-fields",
  components: { SettingReadonlyInput },
  emits: ["updated"],
  props: {
    oidcSettings: {
      type: Object as PropType<OidcCreationData>,
    },
    isEditMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      oidcSettingsFields: {} as OidcCreationData,
    };
  },
  created() {
    if (this.oidcSettings) {
      this.oidcSettingsFields = {
        discoverDocumentUrl: this.oidcSettings.discoverDocumentUrl,
        clientId: this.oidcSettings.clientId,
        clientSecret: this.oidcSettings.clientSecret,
      };
    }
  },
  computed: {
    urlHint(): string {
      return this.isEditMode ? urlHintMessage : "";
    },
    securityResourceType(): keyof typeof ResourceType {
      return "SecuritySettings";
    },
  },
  methods: {
    async validate(): Promise<boolean> {
      const urlValid = await (this.$refs.oidcUrlInput as QInput).validate();
      const clientIdValid = await (this.$refs.oidcClientIdInput as QInput).validate();
      const clientSecretValid = await (this.$refs.oidcClientSecretInput as QInput).validate();

      return urlValid && clientIdValid && clientSecretValid && this.isIdpFieldsValid();
    },
    isNotEmptyClientID(value: string): boolean | string {
      if (!value) return errorMessages.ENTER_CLIENT_ID;
      if (value.includes(" ")) return errorMessages.CLIENT_ID_NO_SPACES;
      return true;
    },
    isNotEmptySecret(value: string): boolean | string {
      if (!value) return errorMessages.SECRET_NOT_EMPTY;
      if (value.includes(" ")) return errorMessages.CLIENT_SECRET_NO_SPACES;
      return true;
    },
    isValidUrl(url: string): boolean | string {
      return (this.isEditMode && url === "") || isValidUrl(url) || errorMessages.INVALID_URL;
    },
    updateClientId(clientId: string | number | null): void {
      this.oidcSettingsFields.clientId = String(clientId || "").trim();
      this.onUpdated();
    },
    updateClientSecret(clientSecret: string | number | null): void {
      this.oidcSettingsFields.clientSecret = String(clientSecret || "").trim();
      this.onUpdated();
    },
    updateDiscoveryUrl(discoveryUrl: string | number | null): void {
      this.oidcSettingsFields.discoverDocumentUrl = String(discoveryUrl || "");
      this.onUpdated();
    },
    onUpdated(): void {
      this.$emit("updated", this.oidcSettingsFields, this.isIdpFieldsValid());
    },
    isIdpFieldsValid(): boolean {
      const discoverDocumentUrl: string = this.oidcSettingsFields.discoverDocumentUrl;
      return (
        (this.isEditMode && !discoverDocumentUrl) ||
        (isValidUrl(discoverDocumentUrl) &&
          isNotEmpty(this.oidcSettingsFields.clientId) &&
          isNotEmpty(this.oidcSettingsFields.clientSecret) &&
          !this.oidcSettingsFields.clientId.includes(" ") &&
          !this.oidcSettingsFields.clientSecret.includes(" "))
      );
    },
  },
});
</script>
