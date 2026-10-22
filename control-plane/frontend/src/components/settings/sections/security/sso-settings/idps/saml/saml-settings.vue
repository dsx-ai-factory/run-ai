<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

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
  <div v-else class="idp-card">
    <setting-editable-field
      resource-type="SecuritySettings"
      @edit="onEdit"
      @delete="onDelete"
      :delete-disabled="isSSOUser"
      :delete-tool-tip="settingBoxMessages.removeToolTip"
      :edit-tool-tip="settingBoxMessages.editToolTip"
      :label="settingBoxMessages.headerTitle"
      :delete-sub-title="settingBoxMessages.confirmRemoveSubTitle"
      use-larger-label
      hide-border
    >
      <template v-slot:start>
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
    <q-separator class="idp-card-separator" />
    <notification-alert-bar
      v-if="certificateWarningVariant"
      class="q-mx-sm q-mb-sm certificate-warning"
      :variant="certificateWarningVariant"
      icon-override="ph-bold ph-warning-circle"
      show-h1-content
      :title="certificateWarningTitle"
      message="Edit the identity provider and replace the metadata XML with an updated file."
    />
    <notification-alert-bar
      v-if="signingDisabled"
      class="q-mx-sm q-mb-sm certificate-warning"
      :variant="signingDisabledVariant"
      icon-override="ph-bold ph-warning-circle"
      show-h1-content
      :title="signingDisabledTitle"
      :message="signingDisabledMessage"
    />
    <div v-if="hasIdpAttributes" class="redirect-url-field">
      <div>Use the following if required by the identity provider</div>
      <setting-readable-field label="Redirect URL">
        <div class="row no-wrap items-center">
          <div class="col-12 q-pt-md q-pl-md">
            {{ idp?.redirectUri }}
          </div>
        </div>
      </setting-readable-field>
    </div>
    <slot name="box-footer" />
  </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import type { IIdpMapperItem, IWrapperBoxSettingMessages, SamlConfigSettings } from "@/models/setting.model";
import { settingsUtil } from "@/utils/settings.util";
import { useAuthStore } from "@/stores/auth.store";
import { idpCommon, SAML_SIGNING_DISABLED_MESSAGE, SAML_SIGNING_DISABLED_TITLE } from "../idp-common/idp-common";
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
import { ENotificationBarVariant, NotificationAlertBar } from "@/components/common/notification-alert-bar";

export default defineComponent({
  name: "saml-settings",
  components: {
    RunaiBtn,
    SamlFields,
    SettingEditableField,
    IdpMappers,
    SettingReadableField,
    SettingWrapperBox,
    NotificationAlertBar,
  },
  emits: ["remove", "add-idp", "update-idp", "cancel-edit-idp", "edit-mode-changed", "certificate-status"],
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
      certificateExpirationDate: null as Date | null,
      certificateRequestToken: 0,
      signingDisabled: false,
      idpStore: useIdpsStore(),
      idpResources: {} as IdpResources,
      DOCS,
    };
  },
  async created(): Promise<void> {
    this.isEditMode = this.isNewIdp;
    await Promise.all([this.refreshCertificateExpiration(), this.refreshSigningDisabled()]);
  },
  computed: {
    isExpiredCertificate(): boolean {
      if (!this.idp?.samlData?.signingCertificate) return false;
      return !this.certificateExpirationDate || !idpCommon.iValidCertificateDate(this.certificateExpirationDate);
    },
    isCertificateExpiringSoon(): boolean {
      return !!this.certificateExpirationDate && idpCommon.isCertificateExpiringSoon(this.certificateExpirationDate);
    },
    certificateWarningVariant(): ENotificationBarVariant | null {
      if (this.isExpiredCertificate) return ENotificationBarVariant.Error;
      if (this.isCertificateExpiringSoon) return ENotificationBarVariant.Warning;
      return null;
    },
    certificateWarningTitle(): string {
      return this.isExpiredCertificate ? "The certificate expired." : "The certificate is expiring soon.";
    },
    signingDisabledVariant(): ENotificationBarVariant {
      return ENotificationBarVariant.Error;
    },
    signingDisabledTitle(): string {
      return SAML_SIGNING_DISABLED_TITLE;
    },
    signingDisabledMessage(): string {
      return SAML_SIGNING_DISABLED_MESSAGE;
    },
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
      return (this.idp?.redirectUri || "") !== "";
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
        if (this.hasChanges && this.isValidIdpFields) {
          this.idpFields.entityId = this.idp?.samlData?.entityId;
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
      void this.refreshSigningDisabled();
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
    async refreshSigningDisabled(): Promise<void> {
      if (this.isNewIdp) {
        this.signingDisabled = false;
        return;
      }
      const resources = await this.idpStore.loadIdpResources();
      const resource = resources?.resource;
      this.signingDisabled = typeof resource === "string" && idpCommon.isSamlSigningDisabled(resource);
    },
    async refreshCertificateExpiration(): Promise<void> {
      const signingCertificate = this.idp?.samlData?.signingCertificate;
      const requestToken = ++this.certificateRequestToken;
      const expirationDate = signingCertificate
        ? await idpCommon.getCertificateExpirationDate(signingCertificate)
        : null;
      if (requestToken !== this.certificateRequestToken) return;
      this.certificateExpirationDate = expirationDate;
      this.$emit("certificate-status", {
        expired: this.isExpiredCertificate,
        expiringSoon: this.isCertificateExpiringSoon,
      });
    },
  },
  watch: {
    isEditMode(newValue: boolean): void {
      this.$emit("edit-mode-changed", newValue);
    },
    "idp.samlData.signingCertificate"(): void {
      this.refreshCertificateExpiration();
    },
  },
});
</script>
<style lang="scss" scoped>
.idp-card {
  border: 1px solid $black-12;
  border-radius: 4px;
  overflow: hidden;
}
.certificate-warning,
.redirect-url-field {
  margin: 12px;
}
.idp-card-separator {
  margin: 0 12px;
}
:deep(.download-icon) {
  color: var(--icon-default);

  .q-icon {
    font-size: 20px;
  }
}
</style>
