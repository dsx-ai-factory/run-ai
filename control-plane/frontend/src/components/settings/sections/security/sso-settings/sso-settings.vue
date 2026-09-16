<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="sso-settings">
    <div class="row items-center q-mb-md">
      <span class="q-mr-md">Set SSO (Single Sign-On)</span>
      <runai-tooltip
        aid="tooltip-sso-settings"
        tooltip-text="Anyone given access to the Run:ai application via the identity provider system will be able to join the platform and authenticate using SSO."
        width="400px"
        tooltip-position="right"
      />
    </div>
    <setting-readonly-input v-if="shouldShowAddButton" resource-type="SecuritySettings">
      <template v-slot:default="{ readonly }">
        <runai-btn
          aid="add-idp-btn"
          icon="ph-bold ph-plus"
          label="IDENTITY PROVIDER"
          flat
          color="primary"
          class="idp-button btn-icon-sm"
          :disable="readonly"
        >
          <q-menu touch-position>
            <q-list style="min-width: 100px">
              <q-item clickable v-close-popup key="saml" @click="handleSelectIDP(IdpTypeEnum.Saml)">
                <q-item-section>Custom SAML 2.0</q-item-section>
              </q-item>
              <q-item clickable v-close-popup key="oidc" @click="handleSelectIDP(IdpTypeEnum.Oidc)">
                <q-item-section>Custom OpenID Connect</q-item-section>
              </q-item>
              <q-item clickable v-close-popup key="ocp" @click="handleSelectIDP(IdpTypeEnum.OpenshiftV4)">
                <q-item-section>OpenShift V4</q-item-section>
              </q-item>
            </q-list>
          </q-menu>
        </runai-btn>
      </template>
    </setting-readonly-input>
    <div v-else>
      <saml-settings
        v-if="isSamlIdp"
        :idp="currentIdp"
        :title="getIdpTitle"
        :idp-mappers="mappers"
        :is-new-idp="isNewIdp"
        @remove="deleteIdp"
        @update-idp="updateIdp"
        @add-idp="addIdp"
        @cancel-edit-idp="cancelEditIdp"
        @edit-mode-changed="onEditModeChanged"
        @certificate-status="onCertificateStatus"
        ref="samlSettingsEl"
      >
        <template v-if="shouldShowAutoRedirectSSO || shouldShowHelp" #box-footer>
          <q-separator v-if="shouldShowAutoRedirectSSO" class="idp-card-separator" />
          <div v-if="shouldShowAutoRedirectSSO" class="row items-center q-my-md q-mx-md">
            <toggle-setting
              ref="toggleSetting"
              @update-setting="updateRedirectSSO"
              :setting="autoRedirectSetting"
              emit-only
              resource-type="SecuritySettings"
            />
          </div>
          <div class="row items-center q-mt-md q-mb-lg q-mx-md" v-if="shouldShowHelp">
            <span class="text-italic"
              >For more information, see the
              <a :href="DOCS.AUTHENTICATION.OVERVIEW" target="_blank">Single sign-On guide</a></span
            >
          </div>
        </template>
      </saml-settings>
      <oidc-settings
        v-if="isOidcIdp"
        :idp="currentIdp"
        :title="getIdpTitle"
        :idp-mappers="mappers"
        :is-new-idp="isNewIdp"
        @remove="deleteIdp"
        @update-idp="updateIdp"
        @add-idp="addIdp"
        @cancel-edit-idp="cancelEditIdp"
        @edit-mode-changed="onEditModeChanged"
        ref="oidcSettingsEl"
      />
      <ocp-settings
        v-if="isOcpIdp"
        :idp="currentIdp"
        :title="getIdpTitle"
        :idp-mappers="mappers"
        :is-new-idp="isNewIdp"
        @remove="deleteIdp"
        @update-idp="updateIdp"
        @add-idp="addIdp"
        @cancel-edit-idp="cancelEditIdp"
        @edit-mode-changed="onEditModeChanged"
        ref="ocpSettingsEl"
      />
    </div>
  </div>
  <div v-if="!isSamlIdp && shouldShowAutoRedirectSSO" class="q-ml-sm q-mr-sm">
    <q-separator class="q-mt-lg" />
    <div class="row items-center q-my-md">
      <toggle-setting
        ref="toggleSetting"
        @update-setting="updateRedirectSSO"
        :setting="autoRedirectSetting"
        emit-only
        resource-type="SecuritySettings"
      />
    </div>
    <q-separator class="q-mb-md" />
  </div>
  <div class="row items-center q-my-md" v-if="!isSamlIdp && shouldShowHelp">
    <span class="text-italic"
      >For more information, see the
      <a :href="DOCS.AUTHENTICATION.OVERVIEW" target="_blank">Single sign-On guide</a></span
    >
  </div>
</template>

<script lang="ts">
import type {
  Idp,
  IdpCreationRequest,
  Mappers,
  OidcCreationData,
  SamlCreationData,
  SettingsGetResponse,
} from "@/swagger-models/identity-manager-client";
import { IdpCreationRequestTypeEnum, IdpTypeEnum } from "@/swagger-models/identity-manager-client";
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiBtn } from "@/components/common/runai-btn";
import { useIdpsStore } from "@/stores/idps.store";
import { alertUtil } from "@/utils/alert.util";
import {
  autoRedirectSSOSetting,
  defaultMappers,
  defaultOcpMappers,
  settingUpdateSuccessMessage,
} from "@/models/setting.model";
import { settingsUtil } from "@/utils/settings.util";
import { useAppStore } from "@/stores/app.store";
import { OidcSettings } from "./idps/oidc";
import { SamlSettings } from "./idps/saml";
import { OcpSettings } from "./idps/ocp";
import { SettingReadonlyInput } from "@/components/settings/setting-readonly-input";
import { ToggleSetting } from "@/components/settings/toggle-setting";
import { useAuthStore } from "@/stores/auth.store";
import { useSecuritySettingsStore } from "@/stores/security-settings.store";
import { SettingsKeyEnum } from "@/swagger-models/identity-manager-client/api";
import type { Setting } from "@/models/setting.model";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  name: "sso-settings",
  components: { ToggleSetting, SettingReadonlyInput, OcpSettings, OidcSettings, SamlSettings, RunaiTooltip, RunaiBtn },
  emits: ["update:settings", "certificate-status"],
  props: {
    settings: {
      type: Object as PropType<SettingsGetResponse | null>,
      required: true,
    },
  },
  data() {
    return {
      isEditMode: false,
      idpStore: useIdpsStore(),
      isNewIdp: false,
      idpNewAlias: "" as string,
      appStore: useAppStore(),
      currentType: IdpTypeEnum.Saml,
      authStore: useAuthStore(),
      securitySettingsStore: useSecuritySettingsStore(),
      autoRedirectSSOSetting,
      DOCS,
    };
  },
  computed: {
    IdpTypeEnum(): typeof IdpTypeEnum {
      return IdpTypeEnum;
    },
    currentIdp(): Idp | undefined {
      return this.idpStore.currentIdp
        ? this.idpStore.currentIdp
        : this.getDefaultIdpConfig(this.idpNewAlias as IdpTypeEnum);
    },
    mappers(): Mappers {
      return this.isNewIdp
        ? this.isOcpIdp
          ? defaultOcpMappers
          : defaultMappers
        : (this.idpStore.idpMappers as Mappers);
    },
    getIdpTitle(): string {
      switch (this.currentIdp?.type) {
        case IdpCreationRequestTypeEnum.Saml:
          return "SAML 2.0";
        case IdpCreationRequestTypeEnum.Oidc:
          return "OpenID Connect";
        case IdpCreationRequestTypeEnum.OpenshiftV4:
          return "OpenShift V4";
        default:
          return "";
      }
    },
    shouldShowAddButton(): boolean {
      return this.currentIdp == undefined && !this.isNewIdp;
    },
    isSamlIdp(): boolean {
      return this.currentIdp?.type === IdpTypeEnum.Saml;
    },
    isOidcIdp(): boolean {
      return this.currentIdp?.type === IdpTypeEnum.Oidc;
    },
    isOcpIdp(): boolean {
      return this.currentIdp?.type === IdpTypeEnum.OpenshiftV4;
    },
    shouldShowHelp(): boolean {
      return !this.isNewIdp && !this.isEditMode;
    },
    shouldShowAutoRedirectSSO(): boolean {
      return !this.isEditMode && Boolean(this.currentIdp);
    },
    isSSOUser(): boolean {
      return this.authStore.isCurrentUserSSO;
    },
    redirectSSO(): boolean {
      const idpAlias = this.currentIdp?.alias;
      if (!idpAlias || !this.settings?.autoRedirectSSO) return false;
      return (
        this.settings.autoRedirectSSO?.value?.idpAlias === idpAlias &&
        Boolean(this.settings?.autoRedirectSSO?.value?.enabled)
      );
    },
    autoRedirectSetting(): Setting {
      return {
        ...this.autoRedirectSSOSetting,
        value: String(this.redirectSSO),
      };
    },
  },
  methods: {
    async updateRedirectSSO(key: string, enabled: boolean): Promise<void> {
      const idpAlias = this.currentIdp?.alias;
      if (idpAlias) {
        try {
          await this.securitySettingsStore.updateSetting(SettingsKeyEnum.AutoRedirectSso, { enabled, idpAlias });
          if (this.settings) {
            this.$emit("update:settings", {
              ...this.settings,
              autoRedirectSSO: { value: { enabled, idpAlias } },
            });
          }
          this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
        } catch (error) {
          console.error("Failed to update auto redirect SSO:", error);
          this.$q.notify(settingsUtil.getSettingErrorMessage());
        }
      } else {
        console.error("Identity provider not found.");
        this.$q.notify(settingsUtil.getSettingErrorMessage());
      }
    },
    onEditModeChanged(isEditMode: boolean): void {
      this.isEditMode = isEditMode;
    },
    onCertificateStatus(status: { expired: boolean; expiringSoon: boolean }): void {
      this.$emit("certificate-status", status);
    },
    getDefaultIdpConfig(idpType: IdpTypeEnum): Idp | undefined {
      if (!idpType) return undefined;
      return {
        alias: idpType,
        type: idpType,
        redirectUri: this.getDefaultRedirectUri(),
      };
    },
    getDefaultRedirectUri(): string {
      const authHost = this.appStore.config.authHost;
      return `${authHost}/broker/<idp-name>/endpoint`;
    },
    handleSelectIDP(idp: IdpTypeEnum): void {
      this.isNewIdp = true;
      this.idpNewAlias = idp;
    },
    async deleteIdp(): Promise<void> {
      if (!this.currentIdp) return;
      try {
        await this.idpStore.removeIdp(this.currentIdp.alias);
        this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
      } catch (error: unknown) {
        this.$q.notify(settingsUtil.getSettingErrorMessage());
        console.error(error);
      }
    },
    getIdpSettingsElement(): typeof OidcSettings | typeof SamlSettings | typeof OcpSettings {
      if (this.isOidcIdp) {
        return this.$refs.oidcSettingsEl as typeof OidcSettings;
      } else if (this.isOcpIdp) {
        return this.$refs.ocpSettingsEl as typeof OcpSettings;
      }
      return this.$refs.samlSettingsEl as typeof SamlSettings;
    },
    async updateIdp(idpCreationRequest: IdpCreationRequest): Promise<void> {
      if (!this.currentIdp) return;
      try {
        await this.idpStore.updateIdp(this.currentIdp.alias, idpCreationRequest);
        this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
        this.getIdpSettingsElement().changeToReadMode();
      } catch (error: unknown) {
        this.$q.notify(settingsUtil.getSettingErrorMessage());
        console.error(error);
      }
    },
    async addIdp(idpNewFields: OidcCreationData | SamlCreationData, mappers: Mappers): Promise<void> {
      const idpCreationRequest: IdpCreationRequest | null = settingsUtil.createIdpCreationRequest(
        this.idpNewAlias,
        idpNewFields,
        mappers,
      );
      if (idpCreationRequest === null) return;
      try {
        await this.idpStore.addIdp(idpCreationRequest);
        this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
        this.idpNewAlias = "";
        this.isNewIdp = false;
        this.getIdpSettingsElement().changeToReadMode();
      } catch (error: unknown) {
        switch (idpCreationRequest.type) {
          case IdpCreationRequestTypeEnum.Saml:
            this.$q.notify(
              alertUtil.getError("SAML metadata couldn't be retrieved. Validate the metadata URL and try again."),
            );
            break;
          case IdpCreationRequestTypeEnum.Oidc:
            this.$q.notify(
              alertUtil.getError(
                "OpenID Connect metadata couldn't be retrieved. Validate the discovery URL and try again.",
              ),
            );
            break;
          case IdpCreationRequestTypeEnum.OpenshiftV4:
            this.$q.notify(
              alertUtil.getError("OpenShift metadata couldn't be retrieved. Validate the base URL and try again."),
            );
            break;
          default:
            this.$q.notify(settingsUtil.getSettingErrorMessage());
            break;
        }
        console.error(error);
      }
    },
    cancelEditIdp(): void {
      this.isNewIdp = false;
      this.idpNewAlias = "";
    },
  },
});
</script>
<style lang="scss" scoped>
.sso-settings {
  .idp-button {
    margin-bottom: 20px;
  }
}
.idp-card-separator {
  margin: 0 12px;
}
</style>
