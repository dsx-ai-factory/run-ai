<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <setting-wrapper-box
    v-if="isEditMode"
    :setting-box-messages="settingBoxMessages"
    :remove-disabled="isSSOUser"
    has-header
    :has-header-buttons="false"
    :has-footer="isEditMode"
    :validation-message="validationMessage"
    @secondary-button-clicked="onCancelEdit"
    @primary-button-clicked="onUpdate"
    @edit-clicked="onEdit"
    :show-primary="!isReadOnly"
    :secondary-button-label="isReadOnly ? 'Close' : 'Cancel'"
    resource-type="SecuritySettings"
  >
    <div class="col-12">
      <div class="q-mb-lg">
        <oidc-fields
          @updated="updateOidcFields"
          ref="oidcSetupFieldsEl"
          :oidc-settings="oidcConfigSettings"
          :is-edit-mode="!isNewIdp"
        />
      </div>
    </div>
    <div class="col-12 q-mb-lg">
      <div v-if="hasIdpAttributes">Use the following if required by the identity provider</div>
      <setting-readable-field label="Redirect URL" v-if="idp?.redirectUri">
        <div class="row no-wrap items-center">
          <div class="col-9 q-pt-md q-pl-md">
            {{ idp.redirectUri }}
          </div>
        </div>
      </setting-readable-field>
    </div>
    <div class="col-12 q-my-lg">
      <idp-scope v-if="isEditMode" :scopes="sourceOidcSettings.scopes" idp-name="OIDC" @update-scope="updateScope" />
    </div>
    <div class="col-12">
      <idp-mappers :idp-mappers="idpMappersItems" />
    </div>
    <div class="row items-center q-mt-xl">
      <span class="text-italic"
        >For more information, see
        <a :href="DOCS.AUTHENTICATION.OIDC" target="_blank">Setup SSO with OpenID Connect</a></span
      >
    </div>
  </setting-wrapper-box>
  <setting-editable-field
    v-else
    resource-type="SecuritySettings"
    @edit="onEdit"
    @delete="onRemove"
    :delete-disabled="isSSOUser"
    :delete-tool-tip="settingBoxMessages.removeToolTip"
    :edit-tool-tip="settingBoxMessages.editToolTip"
    :label="settingBoxMessages.headerTitle"
    :delete-sub-title="settingBoxMessages.confirmRemoveSubTitle"
    use-larger-label
  >
  </setting-editable-field>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import SettingWrapperBox from "@/components/settings/setting-wrapper-box/setting-wrapper-box.vue";
import type { Idp, IdpCreationRequest, Mappers, OidcCreationData } from "@/swagger-models/identity-manager-client";
import { IdpCreationRequestTypeEnum } from "@/swagger-models/identity-manager-client";
import SettingReadableField from "@/components/settings/setting-readable-field/setting-readable-field.vue";
import IdpMappers from "@/components/settings/sections/security/sso-settings/idps/idp-mappers/idp-mappers.vue";
import type { IIdpMapperItem, IWrapperBoxSettingMessages } from "@/models/setting.model";
import { settingsUtil } from "@/utils/settings.util";
import { useAuthStore } from "@/stores/auth.store";
import { OidcFields } from "./oidc-fields";
import { DEFAULT_SCOPE_OPTIONS, idpCommon, OPENID_SCOPE } from "../idp-common/idp-common";
import SettingEditableField from "@/components/settings/setting-editable-field/setting-editable-field.vue";
import { isEqual } from "@/utils/common.util";
import type { ISelectOption } from "@/models/global.model";
import IdpScope from "@/components/settings/sections/security/sso-settings/idps/idp-scope/idp-scope.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { ResourceType } from "@/swagger-models/authorization-client";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  name: "oidc-settings",
  components: { IdpScope, SettingEditableField, OidcFields, IdpMappers, SettingReadableField, SettingWrapperBox },
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
      oidcConfigSettings: {
        discoverDocumentUrl: "",
        clientId: "",
        clientSecret: "",
        scopes: [OPENID_SCOPE],
        mandatoryClaim: null,
      } as OidcCreationData,
      sourceOidcSettings: {} as OidcCreationData,
      isValidFields: true,
      authStore: useAuthStore(),
      selectedScopeOptions: [DEFAULT_SCOPE_OPTIONS] as ISelectOption[],
      scopeInputError: false,
      DOCS,
    };
  },
  created(): void {
    this.isEditMode = this.isNewIdp;
    this.sourceOidcSettings = { ...this.oidcConfigSettings };
  },
  computed: {
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
      if (this.isEditMode && (this.isIdpMapperChanged || this.fieldsHasChanges)) return "Unsaved changes";
      return "";
    },
    fieldsHasChanges(): boolean {
      return !isEqual(this.sourceOidcSettings, this.oidcConfigSettings);
    },
    hasIdpAttributes(): boolean {
      return (this.idp?.redirectUri || "") !== "";
    },
    getUpdateIdpData(): IdpCreationRequest {
      return {
        name: this.idp?.alias,
        type: IdpCreationRequestTypeEnum.Oidc,
        oidcData: this.fieldsHasChanges ? this.filterMandatoryClaim(this.oidcConfigSettings) : undefined,
        mappers: settingsUtil.getMappersFromIdpMapperItems(this.idpMappersItems),
      };
    },
    isReadOnly(): boolean {
      return usePermissionStore().isReadOnly(ResourceType.SecuritySettings);
    },
  },
  methods: {
    filterMandatoryClaim(oidcData: OidcCreationData): OidcCreationData {
      if (oidcData.mandatoryClaim) {
        const { claim, values } = oidcData.mandatoryClaim;
        if (!claim || !values || values.length === 0) {
          return { ...oidcData, mandatoryClaim: undefined };
        }
      }
      return oidcData;
    },
    onRemove(): void {
      this.$emit("remove");
    },
    onEdit(): void {
      this.oidcConfigSettings = {
        discoverDocumentUrl: "",
        clientId: this.idp?.oidcData?.clientId || "",
        clientSecret: this.idp?.oidcData?.clientSecret || "",
        scopes: idpCommon.getDefaultScope(this.idp?.oidcData?.scopes, true),
        mandatoryClaim: this.idp?.oidcData?.mandatoryClaim,
      } as OidcCreationData;
      this.sourceOidcSettings = { ...this.oidcConfigSettings };
      this.idpMappersItems = this._getIdpMappers();
      this.isEditMode = true;
    },
    updateOidcFields(newOidcFields: OidcCreationData, isValid: boolean): void {
      this.isValidFields = isValid;
      if (isValid) {
        this.oidcConfigSettings = {
          ...this.oidcConfigSettings,
          ...newOidcFields,
        };
      }
    },
    updateScope(scopes: Array<string>): void {
      this.oidcConfigSettings.scopes = scopes;
      this.updateOidcFields(this.oidcConfigSettings, this.isValidFields);
    },
    async onUpdate(): Promise<void> {
      const isValid = await (this.$refs.oidcSetupFieldsEl as typeof OidcFields).validate();
      this.isValidFields = isValid;

      if (!isValid) {
        return;
      }

      if (this.isNewIdp) {
        if (this.isValidFields) {
          this.$emit(
            "add-idp",
            this.filterMandatoryClaim(this.oidcConfigSettings),
            settingsUtil.getMappersFromIdpMapperItems(this.idpMappersItems),
          );
        }
      } else {
        if (this.shouldUpdateIdp() && this.isValidFields) {
          this.$emit("update-idp", this.getUpdateIdpData);
        }
      }
    },
    shouldUpdateIdp(): boolean {
      if (!this.isValidFields) {
        return false;
      }

      return this.isIdpMapperChanged || this.fieldsHasChanges;
    },
    onCancelEdit(): void {
      this.isEditMode = false;
      this.$emit("cancel-edit-idp");
    },
    changeToReadMode(): void {
      this.isEditMode = false;
      this.oidcConfigSettings = {} as OidcCreationData;
    },
    _getIdpMappers(): Record<string, IIdpMapperItem> {
      return this.idpMappers ? settingsUtil.getIdpMappersItems(this.idpMappers) : {};
    },
  },
  watch: {
    isEditMode(newValue: boolean): void {
      this.$emit("edit-mode-changed", newValue);
    },
  },
});
</script>
