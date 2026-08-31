<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="notifications-section settings-container" label="Notifications" :subheader="subheader">
    <email-notifications
      :setting="emailSettings"
      class="q-pb-lg"
      v-if="shouldShowEmailSection"
      @deleted="deleteEmailSettings"
      @updated="updateEmailSettings"
    />
    <template v-if="shouldShowEnableNotificationToggle">
      <div class="q-pb-lg">
        <toggle-setting
          :setting="emailNotificationsEnabledSettings"
          emit-only
          resource-type="NotificationSettings"
          :disabled="isEmailToggleDisabled"
          @update-setting="updateEmailNotificationStateForTenant"
        />
        <q-tooltip v-if="isEmailToggleDisabled">To enable email integration, contact your cloud provider</q-tooltip>
      </div>
    </template>
    <q-separator
      v-if="shouldShowEmailSection || shouldShowEnableNotificationToggle"
      class="q-mb-md hide-if-last-child"
    />
    <slack-notifications
      class="q-pb-lg"
      :has-slack-integration="isSlackEnabled"
      :workspace-name="slackWorkspaceName"
      :app-credentials-exist="slackAppCredentialsExist"
      v-if="shouldShowSlackSection"
      @connect-slack-workspace="connectSlackWorkspace"
      @disconnect-slack-workspace="disconnectSlackWorkspace"
    />
    <q-separator v-if="shouldShowSlackSection" class="q-mb-md hide-if-last-child" />
    <system-notifications :setting="adminMessageSetting" />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { DEFAULT_EMPTY_EMAIL_SETTINGS, EmailErrorMessages, SettingKeys } from "@/models/setting.model";
import type { IAdminMessageSetting, ISettingEmailServer } from "@/models/setting.model";
import { useSettingStore } from "@/stores/setting.store";
import type { Setting } from "@/models/setting.model";
import type {
  NotificationStateStatus,
  DetailedNotificationStateSlack,
} from "@/swagger-models/notifications-proxy-client";
import SystemNotifications from "@/components/settings/sections/notifications/system-notifications/system-notifications.vue";
import EmailNotifications from "@/components/settings/sections/notifications/email-notifications/email-notifications.vue";
import SlackNotifications from "@/components/settings/sections/notifications/slack-notifications/slack-notifications.vue";
import { notificationService } from "@/services/control-plane/notifications.service/notification.service";
import { settingsUtil } from "@/utils/settings.util";
import { useAppStore } from "@/stores/app.store";
import ToggleSetting from "@/components/settings/toggle-setting/toggle-setting.vue";
import { alertUtil } from "@/utils/alert.util";
import { intervalUtil } from "@/utils/interval.util/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { useAuthStore } from "@/stores/auth.store";
import { useNotificationStore } from "@/stores/notifications.store";
import { ForbiddenError, HttpErrorResponse, NotFoundError } from "@/models/http-response.model";

const slackAuthPullingTimeoutMs = 2 * 60 * 1000; // 2 minutes

export default defineComponent({
  name: "notifications-section",
  components: { ToggleSetting, EmailNotifications, SlackNotifications, SystemNotifications, RunaiExpansionItem },
  data() {
    return {
      settingStore: useSettingStore(),
      appStore: useAppStore(),
      authStore: useAuthStore(),
      notificationStore: useNotificationStore(),
      emailSettings: undefined as ISettingEmailServer | undefined,
      emailSettingLoaded: false,
      emailNotificationsEnabledSettings: {} as Setting,
      emailNotificationState: {} as NotificationStateStatus,
      slackNotificationState: {} as DetailedNotificationStateSlack,
      slackNotificationStateLoaded: false,
      slackAuthStartTime: null as number | null,
    };
  },
  created: async function () {
    this.emailNotificationState = await this.getEmailNotificationState();
    this.slackNotificationState = await this.getSlackNotificationState();
    if (this.appStore.isMultiTenant) {
      this.emailNotificationsEnabledSettings = this.notificationsEnableSettings();
      this.notificationStore.loadNotificationChannels();
    } else {
      this.emailSettings = await this.emailServerSetting();
    }
    this.emailSettingLoaded = true;
  },
  beforeUnmount() {
    if (!this.slackAuthStartTime) return;
    this.stopPollingSlackAuth();
  },
  computed: {
    adminMessageSetting(): IAdminMessageSetting | undefined {
      if (!this.settingStore.adminMessage) return undefined;
      return {
        key: SettingKeys.AdminMessage,
        value: this.settingStore.adminMessage,
        label: "Admin message",
        description: "Write a message that will be displayed to all users",
      };
    },
    shouldShowEmailSection(): boolean {
      return this.emailSettings != undefined && this.emailSettingLoaded && !this.appStore.isMultiTenant;
    },
    shouldShowEnableNotificationToggle(): boolean {
      return this.emailNotificationsEnabledSettings?.value != undefined && this.appStore.isMultiTenant;
    },
    shouldShowSlackSection(): boolean {
      return this.slackNotificationStateLoaded;
    },
    isEmailToggleDisabled(): boolean {
      return this.appStore.isMultiTenant && !this.notificationStore.isEmailChannelAvailable;
    },
    tenantSlackChannelName(): string {
      return `slack-${this.authStore.tenant.id}`;
    },
    isEmailEnabled(): boolean {
      return !!this.emailNotificationState?.enabled;
    },
    isSlackEnabled(): boolean {
      return !!this.slackNotificationState?.enabled;
    },
    slackWorkspaceName(): string {
      return this.slackNotificationState?.settings?.workspaceName || "";
    },
    slackAppCredentialsExist(): boolean {
      return !!this.slackNotificationState?.settings?.appCredentialsConfigured;
    },
    subheader(): string {
      const emailTitle =
        this.emailNotificationsEnabledSettings?.value || this.emailSettings?.smtpHost ? "Email notifications" : "";
      const systemTitle =
        this.adminMessageSetting != undefined && this.adminMessageSetting.value?.msg ? "System notification" : "";
      const slackTitle = this.isSlackEnabled ? "Slack notifications" : "";

      const titles = [systemTitle, emailTitle, slackTitle].filter(Boolean);
      return titles.length ? titles.join(" / ") : "None";
    },
  },
  methods: {
    async emailServerSetting(): Promise<ISettingEmailServer | undefined> {
      try {
        const emailChannel = await notificationService.getEmail();
        if (emailChannel?.config) {
          return settingsUtil.smtpConfigToEmailSettings(emailChannel.config);
        }
        return DEFAULT_EMPTY_EMAIL_SETTINGS;
      } catch (error: unknown) {
        if (
          error instanceof HttpErrorResponse &&
          (error.statusCode === NotFoundError.statusCode || error.statusCode === ForbiddenError.statusCode)
        ) {
          return DEFAULT_EMPTY_EMAIL_SETTINGS;
        }
        console.warn("SMTP configuration retrieval failed. Notification service may be down.");
        return undefined;
      }
    },
    notificationsEnableSettings(): Setting {
      return {
        key: SettingKeys.SmtpEnabled,
        value: String(this.isEmailEnabled),
        label: "Enable email notifications",
        description: "Enable users to receive email notifications and set their own preferences via the user settings.",
      };
    },
    async getSlackNotificationState(): Promise<DetailedNotificationStateSlack> {
      try {
        const state = await notificationService.getDetailedSlackNotificationState();
        if (!this.slackNotificationStateLoaded) this.slackNotificationStateLoaded = true;
        return state;
      } catch (error: unknown) {
        console.warn("Failed to get slack notification state", error);
        return { enabled: false };
      }
    },
    async getEmailNotificationState(): Promise<NotificationStateStatus> {
      try {
        const state = await notificationService.getEmailNotificationState();
        return state;
      } catch (error: unknown) {
        console.warn("Failed to get email notification state", error);
        return { enabled: false };
      }
    },
    async deleteEmailSettings(): Promise<void> {
      this.emailSettings = DEFAULT_EMPTY_EMAIL_SETTINGS;
      await this.updateEmailNotificationStateForTenant("", false);
    },
    async updateEmailSettings(emailServerSettings: ISettingEmailServer): Promise<void> {
      this.emailSettings = emailServerSettings;
      await this.updateEmailNotificationStateForTenant("", true);
    },
    async updateEmailNotificationStateForTenant(_key: string, enabled: boolean): Promise<void> {
      try {
        await notificationService.updateEmailNotificationState({ enabled });
        this.emailNotificationState = await this.getEmailNotificationState();
      } catch (error: unknown) {
        this.$q.notify(
          alertUtil.getError(
            enabled
              ? EmailErrorMessages.EMAIL_NOTIFICATION_ENABLE_ERROR
              : EmailErrorMessages.EMAIL_NOTIFICATION_DISABLE_ERROR,
          ),
        );
        return;
      } finally {
        this.emailNotificationsEnabledSettings.value = enabled ? "true" : "";
      }
      return Promise.resolve();
    },
    async connectSlackWorkspace(accessToken?: string): Promise<void> {
      try {
        if (accessToken) {
          await notificationService.createSlackApp(accessToken);
        }
        const slackRedirectUrl = await notificationService.connectSlackWorkspace();
        window.open(slackRedirectUrl, "_blank");
        this.startPollingSlackAuth();
      } catch (error: unknown) {
        this.$q.notify(alertUtil.getError("Failed to connect Slack workspace. Please try again later."));
      }
    },
    startPollingSlackAuth(): void {
      this.slackAuthStartTime = Date.now();
      intervalUtil.startInterval(EIntervalLabels.SlackAuthPolling, this.checkSlackAuthStatus);
    },
    stopPollingSlackAuth(): void {
      this.slackAuthStartTime = null;
      intervalUtil.stopInterval(EIntervalLabels.SlackAuthPolling);
    },
    async checkSlackAuthStatus(): Promise<void> {
      try {
        const elapsedTimeMs = Date.now() - (this.slackAuthStartTime || 0);

        if (elapsedTimeMs >= slackAuthPullingTimeoutMs) {
          this.stopPollingSlackAuth();
          return;
        }

        this.slackNotificationState = await this.getSlackNotificationState();
        if (this.isSlackEnabled) {
          this.stopPollingSlackAuth();
        }
      } catch (error) {
        console.warn("Failed to check Slack auth status");
      }
    },
    async disconnectSlackWorkspace(): Promise<void> {
      try {
        await notificationService.disconnectSlackWorkspace();
        if (!this.appStore.isMultiTenant) {
          await notificationService.deleteSlackApp();
        }
        this.slackNotificationState = await this.getSlackNotificationState();
        this.$q.notify(alertUtil.getSuccess("Slack disconnected"));
      } catch (error) {
        this.$q.notify(alertUtil.getError("Failed to disconnect Slack workspace"));
      }
    },
  },
});
</script>
