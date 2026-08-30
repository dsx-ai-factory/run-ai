<template>
  <div class="idp-setup-fields col-12">
    <div>
      <div>Add your identity provider metadata</div>
      <div class="row col-12">
        <setting-readonly-input class="col-12" :resource-type="securityResourceType">
          <template v-slot:default="{ readonly }">
            <!-- eslint-disable vuejs-accessibility/no-autofocus -- this whole field set only renders once the user clicks Edit on the IDP settings box (v-if="isEditMode" in ocp-settings.vue), never on a bare page load -->
            <q-input
              class="col-12 q-mb-lg"
              :model-value="ocpSettingsFields.idpBaseUrl"
              placeholder="Enter URL"
              debounce="300"
              :rules="[isValidUrl]"
              no-error-icon
              @update:model-value="updateIdpBaseUrl"
              stack-label
              label="Base URL"
              ref="ocpUrlInput"
              autofocus
              :hint="urlHint"
              :disable="readonly"
            />
            <!-- eslint-enable vuejs-accessibility/no-autofocus -->
          </template>
        </setting-readonly-input>
        <div class="row col-12">
          <setting-readonly-input class="col-5" :resource-type="securityResourceType">
            <template v-slot:default="{ readonly }">
              <q-input
                class="col-5"
                :model-value="ocpSettingsFields.clientId"
                debounce="300"
                :rules="[isNotEmptyClientID]"
                no-error-icon
                @update:model-value="updateClientId"
                stack-label
                label="Client ID"
                ref="ocpClientIdInput"
                :disable="readonly"
                lazy-rules
              />
            </template>
          </setting-readonly-input>
          <div class="col-1" />
          <setting-readonly-input class="col-6" :resource-type="securityResourceType">
            <template v-slot:default="{ readonly }">
              <q-input
                class="col-6 items-end"
                :model-value="ocpSettingsFields.clientSecret"
                debounce="300"
                :rules="[isNotEmptySecret]"
                no-error-icon
                @update:model-value="updateClientSecret"
                stack-label
                label="Client secret"
                ref="ocpClientSecretInput"
                :disable="readonly"
                lazy-rules
              />
            </template>
          </setting-readonly-input>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { isValidUrl } from "@/common/form.validators";
import type { QInput } from "quasar";
import type { OcpCreationData } from "@/swagger-models/identity-manager-client";
import { urlHintMessage } from "@/models/setting.model";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";
import { errorMessages } from "@/common/error-message.constant";
import type { ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "ocp-fields",
  components: { SettingReadonlyInput },
  emits: ["updated"],
  props: {
    ocpSettings: {
      type: Object as PropType<OcpCreationData>,
    },
    isEditMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      ocpSettingsFields: {} as OcpCreationData,
    };
  },
  created() {
    if (this.ocpSettings) {
      this.ocpSettingsFields = {
        idpBaseUrl: this.ocpSettings.idpBaseUrl,
        clientId: this.ocpSettings.clientId,
        clientSecret: this.ocpSettings.clientSecret,
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
      const urlValid = await (this.$refs.ocpUrlInput as QInput).validate();
      const clientIdValid = await (this.$refs.ocpClientIdInput as QInput).validate();
      const clientSecretValid = await (this.$refs.ocpClientSecretInput as QInput).validate();

      // Only proceed if all fields are valid
      if (!urlValid || !clientIdValid || !clientSecretValid) {
        return false;
      }

      return this.isIdpFieldsValid();
    },
    isNotEmptyClientID(value: string | number | null | undefined): boolean | string {
      const strValue = String(value || "");
      if (!strValue) return errorMessages.ENTER_CLIENT_ID;
      if (strValue.includes(" ")) return errorMessages.CLIENT_ID_NO_SPACES;
      return true;
    },
    isNotEmptySecret(value: string | number | null | undefined): boolean | string {
      const strValue = String(value || "");
      if (!strValue) return errorMessages.SECRET_NOT_EMPTY;
      if (strValue.includes(" ")) return errorMessages.CLIENT_SECRET_NO_SPACES;
      return true;
    },
    isValidUrl(url: string | number | null | undefined): boolean | string {
      const strUrl = String(url || "");
      if (this.isEditMode && !strUrl) return true;
      if (!strUrl) return errorMessages.INVALID_URL;
      return isValidUrl(strUrl) || errorMessages.INVALID_URL;
    },
    updateClientId(clientId: string | number | null): void {
      this.ocpSettingsFields.clientId = String(clientId || "");
      this.onUpdated();
    },
    updateClientSecret(clientSecret: string | number | null): void {
      this.ocpSettingsFields.clientSecret = String(clientSecret || "");
      this.onUpdated();
    },
    updateIdpBaseUrl(idpBaseUrl: string | number | null): void {
      this.ocpSettingsFields.idpBaseUrl = String(idpBaseUrl || "");
      this.onUpdated();
    },
    onUpdated(): void {
      this.$emit("updated", this.ocpSettingsFields, this.isIdpFieldsValid());
    },
    isIdpFieldsValid(): boolean {
      const idpBaseUrl: string = this.ocpSettingsFields.idpBaseUrl || "";
      const clientId: string = this.ocpSettingsFields.clientId || "";
      const clientSecret: string = this.ocpSettingsFields.clientSecret || "";

      // In edit mode, empty URL is allowed
      if (this.isEditMode && !idpBaseUrl) {
        return !!clientId && !clientId.includes(" ") && !!clientSecret && !clientSecret.includes(" ");
      }

      // Check all fields
      return (
        isValidUrl(idpBaseUrl) && !!clientId && !clientId.includes(" ") && !!clientSecret && !clientSecret.includes(" ")
      );
    },
  },
});
</script>
