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
        <ocp-fields
          @updated="updateOcpFields"
          ref="ocpSetupFieldsEl"
          :ocp-settings="ocpConfigSettings"
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
      <idp-scope
        v-if="isEditMode"
        :scopes="sourceOcpSettings.scopes"
        idp-name="openshift-v4"
        @update-scope="updateScope"
      />
    </div>
    <div class="col-12">
      <idp-mappers :idp-mappers="idpMappersItems" />
    </div>
    <div class="row items-center q-mt-xl">
      <span class="text-italic"
        >For more information, see
        <a :href="DOCS.AUTHENTICATION.OPENSHIFT" target="_blank">Setup SSO with OpenShift</a></span
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
import type { PropType } from "vue";
import { defineComponent } from "vue";
import type {
  Idp,
  IdpCreationRequest,
  Mappers,
  OcpCreationData,
  OcpData,
} from "@/swagger-models/identity-manager-client";
import { IdpCreationRequestTypeEnum } from "@/swagger-models/identity-manager-client";
import type { IIdpMapperItem, IWrapperBoxSettingMessages } from "@/models/setting.model";
import { idpCommon } from "../idp-common/idp-common";
import { useAuthStore } from "@/stores/auth.store";
import { settingsUtil } from "@/utils/settings.util/settings.util";
import { isEqual } from "@/utils/common.util";
import { OcpFields } from "./ocp-fields";
import { SettingReadableField } from "@/components/settings/setting-readable-field";
import { SettingWrapperBox } from "@/components/settings/setting-wrapper-box";
import { IdpMappers } from "../idp-mappers";
import { usePermissionStore } from "@/stores/permissions.store";
import { ResourceType } from "@/swagger-models/authorization-client";
import type { ISelectOption } from "@/models/global.model";
import { SettingEditableField } from "@/components/settings/setting-editable-field";
import { IdpScope } from "@/components/settings/sections/security/sso-settings/idps/idp-scope";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  name: "ocp-settings",
  components: { IdpScope, SettingEditableField, OcpFields, IdpMappers, SettingReadableField, SettingWrapperBox },
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
      ocpConfigSettings: {
        idpBaseUrl: "",
        clientId: "",
        clientSecret: "",
        scopes: [],
      } as OcpCreationData,
      sourceOcpSettings: {} as OcpData,
      isValidFields: true,
      authStore: useAuthStore(),
      selectedScopeOptions: [] as ISelectOption[],
      DOCS,
    };
  },
  created(): void {
    this.isEditMode = this.isNewIdp;
    this.sourceOcpSettings = { ...this.ocpConfigSettings };
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
      return !isEqual(this.sourceOcpSettings, this.ocpConfigSettings);
    },
    hasIdpAttributes(): boolean {
      return (this.idp?.redirectUri || "") !== "";
    },
    getUpdateIdpData(): IdpCreationRequest {
      return {
        name: this.idp?.alias,
        type: IdpCreationRequestTypeEnum.OpenshiftV4,
        ocpData: this.fieldsHasChanges ? this.ocpConfigSettings : undefined,
        mappers: settingsUtil.getMappersFromIdpMapperItems(this.idpMappersItems),
      };
    },
    isReadOnly(): boolean {
      return usePermissionStore().isReadOnly(ResourceType.SecuritySettings);
    },
  },
  methods: {
    onRemove(): void {
      this.$emit("remove");
    },
    onEdit(): void {
      this.isEditMode = true;
      this.ocpConfigSettings = {
        idpBaseUrl: this.idp?.ocpData?.idpBaseUrl || "",
        clientId: this.idp?.ocpData?.clientId || "",
        clientSecret: this.idp?.ocpData?.clientSecret || "",
        scopes: idpCommon.getDefaultScope(this.idp?.ocpData?.scopes, false),
      } as OcpCreationData;
      this.sourceOcpSettings = { ...this.ocpConfigSettings };
      this.idpMappersItems = this._getIdpMappers();
    },
    onCancelEdit(): void {
      this.isEditMode = false;
      this.$emit("cancel-edit-idp");
    },
    updateScope(scopes: Array<string>): void {
      this.ocpConfigSettings.scopes = scopes;
      this.updateOcpFields(this.ocpConfigSettings, this.isValidFields);
    },
    async onUpdate(): Promise<void> {
      const isValid = await (this.$refs.ocpSetupFieldsEl as typeof OcpFields).validate();
      this.isValidFields = isValid;

      if (!isValid) {
        return;
      }

      if (this.isNewIdp) {
        if (this.isValidFields) {
          this.$emit("add-idp", this.ocpConfigSettings, settingsUtil.getMappersFromIdpMapperItems(this.idpMappersItems));
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
    updateOcpFields(newOcpFields: OcpCreationData, isValid: boolean): void {
      this.isValidFields = isValid;
      if (isValid) {
        this.ocpConfigSettings = {
          ...this.ocpConfigSettings,
          ...newOcpFields,
        };
      }
    },
    changeToReadMode(): void {
      this.isEditMode = false;
      this.ocpConfigSettings = {} as OcpCreationData;
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
