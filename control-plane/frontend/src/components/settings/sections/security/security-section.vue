<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="security-section settings-container"
    label="Security"
    :subheader="sectionSubHeader"
    :section-invalid="isCertificateExpired"
    :section-warning="isCertificateExpiringSoon"
    @show="() => initializeSecuritySettings(true)"
  >
    <sso-settings v-if="settings" v-model:settings="settings" @certificate-status="onCertificateStatus" />
    <div class="row items-center q-mt-xl">
      <span class="q-mr-md">Set logout redirect URL</span>
      <runai-tooltip
        aid="tooltip-logout-redirect-url"
        tooltip-text="Users will be redirected to this URL after logging out from Run:ai."
        width="400px"
        tooltip-position="right"
      />
    </div>
    <settings-url-field
      :server-url="logoutRedirectUri?.uri || ''"
      @save="saveLogoutRedirectUri"
      delete-tool-tip="Remove redirect URL"
      edit-tool-tip="Edit redirect URL"
      add-button-label="redirect url"
      input-label="Redirect URL"
      delete-sub-title="Users will be directed back to the Run:ai sign-in screen after signing out."
      :resource-type="securityResourceType"
    />
    <div class="row items-center q-mt-xl q-mb-sm">
      <span class="q-mr-md">Set UI inactivity timeout</span>
      <runai-tooltip
        aid="tooltip-session-timeout"
        tooltip-text="Users will be automatically logged out of NVIDIA Run:ai after the set period of inactivity. CLI and API access are not affected. (If left blank, users are logged out after 24 hours of inactivity by default.)"
        width="400px"
        tooltip-position="right"
      />
    </div>
    <div class="input-with-label-wrapper row items-center q-mb-md q-mr-sm">
      <setting-readonly-input :resource-type="securityResourceType">
        <template v-slot:default="{ readonly }">
          <q-input
            aid="security-session-timeout-label"
            label="Minutes"
            :model-value="displaySessionTimeout"
            @update:model-value="updateSessionTimeout"
            class="q-mb-sm session-timeout-input"
            style="width: 200px"
            no-error-icon
            outlined
            type="number"
            :min="0"
            @blur="saveSessionTimeout"
            :disable="readonly"
          />
        </template>
      </setting-readonly-input>
    </div>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// Models
import { settingUpdateSuccessMessage, DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES } from "@/models/setting.model";
import { IdleTimer } from "@/services/infra/idle-timer.service/idle-timer.service";
// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
// Stores
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { useIdpsStore } from "@/stores/idps.store";
import { IdpTypeEnum, SettingsKeyEnum } from "@/swagger-models/identity-manager-client";
import type { SettingsGetResponse, GetLogoutRedirectUriSettingValue } from "@/swagger-models/identity-manager-client";
import { SettingsUrlField } from "@/components/settings/settings-url-field";
import { settingsUtil } from "@/utils/settings.util";
import { SettingReadonlyInput } from "@/components/settings/setting-readonly-input";
import { useSecuritySettingsStore } from "@/stores/security-settings.store";
import { SsoSettings } from "@/components/settings/sections/security/sso-settings";
// Utils
import { alertUtil } from "@/utils/alert.util";
import type { ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "security-section",
  components: {
    SettingReadonlyInput,
    SettingsUrlField,
    SsoSettings,
    RunaiTooltip,
    RunaiExpansionItem,
  },
  props: {},
  data() {
    return {
      idpStore: useIdpsStore(),
      securitySettingsStore: useSecuritySettingsStore(),
      timeoutHasChanged: false,
      browserSessionTimeout: DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES,
      originalBrowserSessionTimeout: DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES,
      idleTimerService: IdleTimer.getInstance(),
      logoutRedirectUri: null as GetLogoutRedirectUriSettingValue | null,
      settings: null as SettingsGetResponse | null,
      isCertificateExpired: false,
      isCertificateExpiringSoon: false,
    };
  },
  async created() {
    await this.initializeSecuritySettings();
    document.addEventListener("visibilitychange", this.handleVisibilityChange);
  },
  beforeUnmount() {
    document.removeEventListener("visibilitychange", this.handleVisibilityChange);
  },
  computed: {
    sectionSubHeader(): string {
      const currentIdp = this.idpStore.currentIdp;
      switch (currentIdp?.type) {
        case IdpTypeEnum.Saml:
          return "SSO - SAML 2.0";
        case IdpTypeEnum.Oidc:
          return "SSO - OpenID Connect";
        case IdpTypeEnum.OpenshiftV4:
          return "SSO - OpenShift V4";
        default:
          return "None";
      }
    },
    securityResourceType(): keyof typeof ResourceType {
      return "SecuritySettings";
    },
    displaySessionTimeout(): string | number {
      return this.browserSessionTimeout === DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES ? "" : this.browserSessionTimeout;
    },
  },
  methods: {
    async initializeSecuritySettings(forceRefresh = false): Promise<void> {
      try {
        this.settings = await this.securitySettingsStore.getSettings(forceRefresh);
        const timeoutSettings = this.settings?.browserSessionTimeout;
        const redirectSettings = this.settings?.logoutRedirectUri;

        const idleTimeout = timeoutSettings?.value?.idleTimeout;
        if (idleTimeout !== undefined && idleTimeout !== null) {
          this.browserSessionTimeout = idleTimeout;
          this.originalBrowserSessionTimeout = idleTimeout;
        }
        if (redirectSettings?.value?.uri) {
          this.logoutRedirectUri = {
            uri: redirectSettings.value.uri,
          };
        }
      } catch (error) {
        console.error("Failed to get security settings:", error);
      }
    },

    onCertificateStatus(status: { expired: boolean; expiringSoon: boolean }): void {
      this.isCertificateExpired = status.expired;
      this.isCertificateExpiringSoon = status.expiringSoon;
    },

    handleVisibilityChange(): void {
      if (!document.hidden) {
        this.initializeSecuritySettings(true);
      }
    },

    updateSessionTimeout(value: string | number | null): void {
      const timeout = value !== null && value !== "" ? +value : 0;
      if (timeout !== this.browserSessionTimeout) {
        this.browserSessionTimeout = timeout;
        this.timeoutHasChanged = true;
      } else {
        this.timeoutHasChanged = false;
      }
    },

    async saveSessionTimeout(): Promise<void> {
      if (this.timeoutHasChanged) {
        try {
          await this.securitySettingsStore.updateSetting(SettingsKeyEnum.BrowserSessionTimeout, {
            idleTimeout: this.browserSessionTimeout,
          });
          this.idleTimerService.updateTimeout(this.browserSessionTimeout);

          this.originalBrowserSessionTimeout = this.browserSessionTimeout;
          this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
        } catch (error) {
          this.$q.notify(settingsUtil.getSettingErrorMessage());
          this.browserSessionTimeout = this.originalBrowserSessionTimeout;
        }
        this.timeoutHasChanged = false;
      }
    },

    async saveLogoutRedirectUri(url: string): Promise<void> {
      try {
        await this.securitySettingsStore.updateSetting(SettingsKeyEnum.LogoutRedirectUri, {
          uri: url || "",
        });
        this.logoutRedirectUri = {
          uri: url || "",
        };
        this.$q.notify(alertUtil.getSuccess(settingUpdateSuccessMessage));
      } catch (error) {
        this.$q.notify(settingsUtil.getSettingErrorMessage());
      }
    },
  },
});
</script>

<style lang="scss" scoped>
.session-timeout-input {
  min-width: 200px;
  :deep(.q-field__native) {
    width: 100%;
  }
}
</style>
