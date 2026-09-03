<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="notifications-settings-container" label="Notifications" :subheader="notificationSummary">
    <runai-form-wrapper :form-state="{}" :init-dirty="hasChanges" full-width>
      <notification-channels-selector
        :options="notificationChannelOptions"
        @channel-toggled="onNotificationChannelToggled"
      />
      <template v-if="showNotificationSettings">
        <div class="q-mt-lg">Notify me about my workloads when:</div>
        <div v-for="(categoryNotifications, categoryName) in notificationCategories" :key="categoryName">
          <div class="row items-center q-mb-sm q-mt-lg">
            <span class="q-mr-md text-b-subtitle1">{{ getCategoryDisplayName(String(categoryName)) }}:</span>
            <runai-tooltip
              aid="tool-tip-system-notification"
              :tooltip-text="getCategoryTooltip(String(categoryName))"
              width="400px"
              tooltip-position="right"
            />
          </div>
          <div class="row items-center q-py-sm">
            <q-checkbox
              :toggle-indeterminate="false"
              indeterminate-value="indeterminate"
              size="md"
              v-model="categoryStatus[categoryName]"
              dense
              label="All"
              @update:model-value="toggleCategory(String(categoryName))"
            />
          </div>
          <div class="q-py-sm" v-for="notification in categoryNotifications" :key="notification.type">
            <q-checkbox
              :toggle-indeterminate="false"
              size="md"
              v-model="notificationStatus[notification.type]"
              dense
              :label="getNotificationTypeDisplayName(notification.type)"
              @update:model-value="checkForIndeterminate(String(categoryName))"
            />
          </div>
        </div>
        <div class="row footer">
          <div class="justify-start validation-message col-9 q-pt-sm">
            <span v-if="hasChanges">Unsaved changes</span>
          </div>
          <div class="justify-end no-wrap q-gutter-sm col-3">
            <runai-btn
              aid="email-notifications-cancel-button"
              label="revert"
              flat
              color="primary"
              @click="cancelChanges"
            />
            <runai-btn
              aid="email-notifications-save-button"
              label="save"
              flat
              color="primary"
              :loading="isSaving"
              @click="saveChanges"
            />
          </div>
        </div>
      </template>
    </runai-form-wrapper>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import {
  uniqueEmailChannelRunaiId,
  NotificationChannelLabels,
  type CategorizedNotificationSettings,
} from "@/models/user-setting.model";
import RunaiTooltip from "@/components/common/runai-tooltip/runai-tooltip.vue";
import { userNotificationsUtil } from "@/utils/user-notifications.utils";
import { useAuthStore } from "@/stores/auth.store";
import type {
  DetailedNotificationStateSlack,
  NotificationStateStatus,
  Subscription,
} from "@/swagger-models/notifications-proxy-client";
import { RunaiBtn } from "@/components/common/runai-btn";
import { NotificationChannelsSelector, type NotificationChannelOption } from "./notification-channels-selector";
import { isEqual, haveSameElements } from "@/utils/common.util";
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";

export default defineComponent({
  name: "user-notifications-section",
  components: {
    RunaiBtn,
    RunaiTooltip,
    RunaiExpansionItem,
    NotificationChannelsSelector,
    RunaiFormWrapper,
  },
  emits: ["save"],
  props: {
    notificationCategories: {
      type: Object as PropType<CategorizedNotificationSettings>,
      required: true,
    },
    slackNotificationState: {
      type: Object as PropType<DetailedNotificationStateSlack>,
      required: true,
    },
    emailNotificationState: {
      type: Object as PropType<NotificationStateStatus>,
      required: true,
    },
    currentSubscription: {
      type: [Object, null] as PropType<Subscription | null>,
      required: true,
    },
    isSaving: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      notificationStatus: {} as Record<string, boolean>,
      originalNotificationsStatus: {} as Record<string, boolean>,
      selectedNotificationChannels: [] as string[],
      authStore: useAuthStore(),
      categoryStatus: {} as Record<string, boolean | string>,
    };
  },
  created() {
    this.initializeSettings();
    this.initializeNotificationChannels();
  },
  computed: {
    notificationSummary(): string {
      if (!this.notificationsEnabled || !this.selectedNotificationChannels.length) {
        return "None";
      }
      return `${Object.keys(this.notificationStatus).length} notification types`;
    },
    hasChanges(): boolean {
      return (
        !isEqual(this.notificationStatus, this.originalNotificationsStatus) ||
        !haveSameElements(this.currentSubscription?.channels || [], this.selectedNotificationChannels)
      );
    },
    notificationsEnabled(): boolean {
      return !!(this.slackNotificationState?.enabled || this.emailNotificationState?.enabled);
    },
    slackChannelId(): string {
      return userNotificationsUtil.generateSlackChannelId(this.authStore.tenant.id);
    },
    notificationChannelOptions(): NotificationChannelOption[] {
      return [
        {
          id: uniqueEmailChannelRunaiId,
          label: NotificationChannelLabels.EMAIL,
          isSelected: this.selectedNotificationChannels.includes(uniqueEmailChannelRunaiId),
          isDisabled: !this.emailNotificationState?.enabled,
          disabledTooltipText:
            "To enable email notifications, contact your administrator about configuring your organization's email server",
        },
        {
          id: this.slackChannelId,
          label: NotificationChannelLabels.SLACK,
          isSelected: this.selectedNotificationChannels.includes(this.slackChannelId),
          isDisabled: !this.slackNotificationState?.enabled,
          disabledTooltipText:
            "To enable Slack notifications, contact your administrator about configuring your organization's Slack workspace",
        },
      ];
    },
    showNotificationSettings(): boolean {
      return (
        this.notificationsEnabled &&
        !!(this.selectedNotificationChannels.length || this.currentSubscription?.channels?.length)
      );
    },
    hasZeroSelectedNotificationTypes(): boolean {
      return Object.values(this.notificationStatus).every((enabled) => !enabled);
    },
  },
  methods: {
    cancelChanges() {
      if (this.isSaving || !this.hasChanges) {
        return;
      }
      this.initializeSettings();
    },
    saveChanges(): void {
      if (this.isSaving || !this.hasChanges) {
        return;
      }
      const result: Subscription = userNotificationsUtil.convertNotificationSettingsToSubscription(
        this.notificationCategories,
        this.selectedNotificationChannels,
        this.notificationStatus,
      );
      this.$emit("save", result);
      this.originalNotificationsStatus = { ...this.notificationStatus };
    },
    initializeSettings(): void {
      Object.entries(this.notificationCategories).forEach(([categoryName, settings]) => {
        this.checkForIndeterminate(categoryName);
        settings.forEach((setting) => {
          this.notificationStatus[setting.type] = setting.enabled;
        });
        this.checkForIndeterminate(categoryName);
      });
      this.originalNotificationsStatus = { ...this.notificationStatus };
      this.selectedNotificationChannels = [...(this.currentSubscription?.channels || [])];
    },
    getCategoryTooltip(categoryName: string): string {
      if (categoryName === "Workload status change") {
        return "A notification will be sent when a workload's status changes to the selected status(es)";
      } else if (categoryName === "Workload timeout") {
        return "A notification will be sent when half the time remains before a workload is set to time out, as defined in the project scheduling rules";
      } else {
        return "";
      }
    },
    getNotificationTypeDisplayName(type: string): string {
      if (type === "Degraded") {
        return "Running with issues";
      }
      return type;
    },
    getCategoryDisplayName(categoryName: string): string {
      if (categoryName === "Workload status change") {
        return "The status changes to";
      } else if (categoryName === "Workload timeout") {
        return "One of these is about to stop due to timeout";
      } else {
        return categoryName;
      }
    },
    toggleCategory(categoryName: string) {
      const isEnabled = this.categoryStatus[categoryName];
      this.notificationCategories[categoryName].forEach((notification) => {
        this.notificationStatus[notification.type] = isEnabled as boolean;
      });
    },
    checkForIndeterminate(categoryName: string) {
      const isAllChecked = this.notificationCategories[categoryName].every(
        (notification) => this.notificationStatus[notification.type],
      );
      const isAllNotChecked = this.notificationCategories[categoryName].every(
        (notification) => !this.notificationStatus[notification.type],
      );
      this.categoryStatus[categoryName] = isAllChecked ? true : isAllNotChecked ? false : "indeterminate";
    },
    onNotificationChannelToggled(id: string): void {
      if (this.selectedNotificationChannels.includes(id)) {
        this.selectedNotificationChannels = this.selectedNotificationChannels.filter((channelId) => channelId !== id);
        return;
      }
      this.selectedNotificationChannels.push(id);

      if (this.hasZeroSelectedNotificationTypes) {
        this.autoSelectAllCheckboxes();
      }
    },
    initializeNotificationChannels(): void {
      this.selectedNotificationChannels = [...(this.currentSubscription?.channels || [])];
    },
    autoSelectAllCheckboxes(): void {
      Object.entries(this.notificationCategories).forEach(([categoryName]) => {
        this.categoryStatus[categoryName] = true;
        this.toggleCategory(categoryName);
      });
    },
  },
});
</script>
<style lang="scss" scoped>
.notifications-settings-container {
  margin: auto;
  width: 800px;

  .footer {
    border-top: #dddddd 1px solid;
    margin: 10px -20px 0 -20px;
    padding-top: 10px;

    .validation-message {
      color: $negative;
      font-size: 12px;
      padding-left: 20px;
    }
  }
}
</style>
