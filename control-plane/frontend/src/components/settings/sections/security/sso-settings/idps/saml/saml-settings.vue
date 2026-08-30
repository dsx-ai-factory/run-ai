<template>
  <setting-wrapper-box
    v-if="isEditMode"
    :setting-box-messages="settingBoxMessages"
    has-header
    :has-header-buttons="false"
    :has-footer="isEditMode"
    :validation-message="validationMessage"
    @secondary-button-clicked="onCancelEdit"
    @primary-button-clicked="onUpdate"
    :secondary-button-label="isReadOnly ? 'Close' : 'Cancel'"
    :show-primary="!isReadOnly"
    resource-type="SecuritySettings"
  >
    <div class="col-12">
      <div class="q-mb-lg">
        <saml-fields
          :idp-resources="idpResources"
          :is-new-idp="isNewIdp"
          :saml-settings="samlConfigSettings"
          @updated="updateSamlFields"
          ref="samlFieldsEl"
        />
      </div>
      <div class="col-12 q-mb-lg">
        <div v-if="hasIdpAttributes">Use the following if required by the identity provider</div>
        <setting-readable-field label="Redirect URL" v-if="idp?.redirectUri">
          <div class="row no-wrap items-center">
            <div class="col-12 q-pt-md q-pl-md">
              {{ idp.redirectUri }}
            </div>
          </div>
        </setting-readable-field>
        <setting-readable-field label="Entity ID">
          <div class="row no-wrap items-center">
            <div class="col-12 q-pt-md q-pl-md">
              {{ idp?.samlData?.entityId || defaultEntityId }}
            </div>
          </div>
        </setting-readable-field>
      </div>
      <idp-mappers :idp-mappers="idpMappersItems" />
      <div class="row items-center q-mt-xl">
        <span class="text-italic"
          >For more information, see <a :href="DOCS.AUTHENTICATION.SAML" target="_blank">Setup SSO with SAML</a></span
        >
      </div>
    </div>
  </setting-wrapper-box>
  <setting-editable-field
    v-else
    resource-type="SecuritySettings"
    @edit="onEdit"
    @delete="onDelete"
    :delete-disabled="isSSOUser"
    :delete-tool-tip="settingBoxMessages.removeToolTip"
    :edit-tool-tip="settingBoxMessages.editToolTip"
    :label="settingBoxMessages.headerTitle"
    :delete-sub-title="settingBoxMessages.confirmRemoveSubTitle"
    use-larger-label
  >
    <template v-slot:start>
      <div class="expiration-date row" v-if="isExpiredCertificate">
        <div class="col-2 q-ml-lg">
          <q-icon name="ph-bold ph-warning-circle" class="q-mr-xl exclamation-icon" />
        </div>
        <div class="col-8 expiration-text">Certificate expired</div>
      </div>
      <runai-btn
        icon="ph-bold ph-arrow-square-out"
        class="q-pa-sm download-icon"
        flat
        size="sm"
        rounded
        aid="open-metadata-xml"
        @click="openMetadataXml"
        tooltip="Open service provider metadata XML"
      />
    </template>
  </setting-editable-field>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import type { IIdpMapperItem, IWrapperBoxSettingMessages, SamlConfigSettings } from "@/models/setting.model";
import { settingsUtil } from "@/utils/settings.util";
import { useAuthStore } from "@/stores/auth.store";
import { idpCommon } from "../idp-common/idp-common";
import type {
  Idp,
  IdpCreationRequest,
  IdpResources,
  Mappers,
  SamlCreationData,
} from "@/swagger-models/identity-manager-client";
import { IdpCreationRequestTypeEnum } from "@/swagger-models/identity-manager-client";
import { useIdpsStore } from "@/stores/idps.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { ResourceType } from "@/swagger-models/authorization-client";
import { IdpMappers } from "@/components/settings/sections/security/sso-settings/idps/idp-mappers";
import { SettingEditableField } from "@/components/settings/setting-editable-field";
import { SettingWrapperBox } from "@/components/settings/setting-wrapper-box";
import { SettingReadableField } from "@/components/settings/setting-readable-field";
import SamlFields from "@/components/settings/sections/security/sso-settings/idps/saml/saml-fields/saml-fields.vue";
import { DOCS } from "@/constants/docs";
import { RunaiBtn } from "@/components/common/runai-btn";

export default defineComponent({
  name: "saml-settings",
  components: { RunaiBtn, SamlFields, SettingEditableField, IdpMappers, SettingReadableField, SettingWrapperBox },
  emits: ["remove", "add-idp", "update-idp", "cancel-edit-idp", "edit-mode-changed"],
  props: {
    title: {
      type: String as PropType<string>,
      required: true,
    },
    idp: {
      type: Object as PropType<Idp> | null,
    },
    idpMappers: {
      type: Object as PropType<Mappers>,
    },
    isNewIdp: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      isEditMode: false,
      idpMappersItems: this._getIdpMappers() as Record<string, IIdpMapperItem>,
      idpFields: {} as SamlCreationData,
      isValidIdpFields: false,
      authStore: useAuthStore(),
      isIdpFieldsChanged: false,
      isExpiredCertificate: false,
      idpStore: useIdpsStore(),
      idpResources: {} as IdpResources,
      DOCS,
    };
  },
  async created(): Promise<void> {
    this.isEditMode = this.isNewIdp;
    this.isExpiredCertificate = await this.isCertificateExpired();
  },
  computed: {
    defaultEntityId(): string {
      return settingsUtil.getDefaultEntityId(window.location.hostname);
    },
    settingBoxMessages(): IWrapperBoxSettingMessages {
      return idpCommon.getSettingsBoxMessages(this.title, this.isSSOUser);
    },
    isSSOUser(): boolean {
      return this.authStore.isCurrentUserSSO;
    },
    isIdpMapperChanged(): boolean {
      return !settingsUtil.isIdpMappersEqual(this.idpMappers || {}, this.idpMappersItems);
    },
    validationMessage(): string {
      if (this.isEditMode && this.hasChanges) return "Unsaved changes";
      return "";
    },
    hasChanges(): boolean {
      return this.isIdpFieldsChanged || this.isIdpMapperChanged;
    },
    hasIdpAttributes(): boolean {
      return this.idp?.redirectUri !== "";
    },
    samlConfigSettings(): SamlConfigSettings {
      if (this.idp?.samlData) {
        return {
          idpEntityId: this.idp?.samlData?.entityId || "",
          singleSignOnServiceUrl: this.idp?.samlData?.singleSignOnServiceUrl || "",
          signingCertificate: this.idp?.samlData?.signingCertificate || "",
        };
      } else {
        return {} as SamlConfigSettings;
      }
    },
    shouldSendIdpFile(): boolean {
      return !!this.idpFields.metadataXmlFile && this.isIdpFieldsChanged;
    },
    getIdpData(): IdpCreationRequest {
      return {
        name: this.idp?.alias,
        type: IdpCreationRequestTypeEnum.Saml,
        samlData: this.shouldSendIdpFile || this.idpFields.metadataXmlUrl ? this.idpFields : undefined,
        mappers: settingsUtil.getMappersFromIdpMapperItems(this.idpMappersItems),
      };
    },
    isReadOnly(): boolean {
      return usePermissionStore().isReadOnly(ResourceType.SecuritySettings);
    },
  },
  methods: {
    onDelete(): void {
      this.$emit("remove");
    },
    async onEdit(): Promise<void> {
      this.idpResources = await this.idpStore.loadIdpResources();
      this.idpMappersItems = this._getIdpMappers();
      this.isEditMode = true;
    },
    updateSamlFields(newSamlFields: SamlCreationData, isValid: boolean, idpFieldsHasChanges: boolean): void {
      this.isValidIdpFields = isValid;
      this.idpFields = newSamlFields;
      this.isIdpFieldsChanged = idpFieldsHasChanges;
    },
    async onUpdate(): Promise<void> {
      if (this.isNewIdp) {
        await (this.$refs.samlFieldsEl as typeof SamlFields).validate();
        if (!this.isValidIdpFields) return;
        this.idpFields.entityId = this.defaultEntityId;
        this.$emit("add-idp", this.idpFields, settingsUtil.getMappersFromIdpMapperItems(this.idpMappersItems));
      } else {
        if (this.hasChanges && (this.isValidIdpFields || !this.isNewIdp)) {
          this.$emit("update-idp", this.getIdpData);
        }
      }
    },
    onCancelEdit(): void {
      this.isEditMode = false;
      this.$emit("cancel-edit-idp");
    },
    changeToReadMode(): void {
      this.isEditMode = false;
      this.idpFields = {} as SamlCreationData;
    },
    _getIdpMappers(): Record<string, IIdpMapperItem> {
      return this.idpMappers ? settingsUtil.getIdpMappersItems(this.idpMappers) : {};
    },
    async openMetadataXml(): Promise<void> {
      const spMetadataUrl = this.idpStore.currentIdp?.samlData?.serviceProviderMetadataUrl;
      if (spMetadataUrl) {
        window.open(spMetadataUrl, "_blank");
      }
    },
    async isCertificateExpired(): Promise<boolean> {
      if (!this.idp?.samlData?.signingCertificate) return false;
      const expirationDate = await idpCommon.getCertificateExpirationDate(this.idp?.samlData?.signingCertificate);
      return !idpCommon.iValidCertificateDate(expirationDate as Date);
    },
  },
  watch: {
    isEditMode(newValue: boolean): void {
      this.$emit("edit-mode-changed", newValue);
    },
  },
});
</script>
<style lang="scss" scoped>
.expiration-date {
  color: $negative;
  margin-top: 12px;
  .exclamation-icon {
    font-size: 18px;
  }
  .expiration-text {
    margin-top: 4px;
    font-size: 12px;
  }
}
.download-icon {
  color: $black-54;
}
</style>
