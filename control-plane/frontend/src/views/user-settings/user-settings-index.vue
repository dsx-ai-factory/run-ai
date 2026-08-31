<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="settings-index q-pt-md">
    <user-profile-section />
    <user-notifications-section
      v-if="notificationsSettingsCategories"
      :notification-categories="notificationsSettingsCategories"
      :slack-notification-state="slackNotificationState"
      :email-notification-state="emailNotificationState"
      :current-subscription="currentSubscription"
      :is-saving="isSavingNotifications"
      @save="saveNotificationsSettings"
    />
    <access-keys-section v-if="showAccessKeys" />
    <personal-credentials-section v-if="showPersonalCredentials" />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// stores
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
// types
import {
  workloadNotificationCategories,
  type CategorizedNotificationSettings,
  type NotificationSettings,
} from "@/models/user-setting.model";
// services
import { notificationService } from "@/services/control-plane/notifications.service/notification.service";
// utils
import { userNotificationsUtil } from "@/utils/user-notifications.utils";
import { alertUtil } from "@/utils/alert.util";

// components
import { AccessKeysSection } from "@/components/user-settings/sections/access-keys-section";
import UserProfileSection from "@/components/user-settings/sections/user-profile-section/user-profile-section.vue";
import UserNotificationsSection from "@/components/user-settings/sections/user-notifications-section/user-notifications-section.vue";
import type {
  NotificationStateStatus,
  NotificationType,
  Subscription,
} from "@/swagger-models/notifications-proxy-client";
import { useClusterStore } from "@/stores/cluster.store";
import {
  MIN_CLUSTER_VERSION_FOR_ACCESS_KEYS,
  MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS,
} from "@/common/version.constant";
import { isNewerVersion } from "@/utils/version.util";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import { PersonalCredentialsSection } from "@/components/user-settings/sections/personal-credentials-section";
import { CLUSTER_PR_TO_MASTER_VERSION_PREFIX } from "@/models/cluster.model";
import { requestToLeave } from "@/services/infra/router.service/router.service";

export default defineComponent({
  name: "user-settings-index",
  components: { UserNotificationsSection, UserProfileSection, AccessKeysSection, PersonalCredentialsSection },
  data() {
    return {
      currentSubscription: null as Subscription | null,
      appStore: useAppStore(),
      authStore: useAuthStore(),
      clusterStore: useClusterStore(),
      slackNotificationState: {} as NotificationStateStatus,
      emailNotificationState: {} as NotificationStateStatus,
      notificationsSettingsCategories: null as CategorizedNotificationSettings | null,
      isSavingNotifications: false,
    };
  },
  async created() {
    try {
      this.slackNotificationState = await notificationService.getSlackNotificationState();
      this.emailNotificationState = await notificationService.getEmailNotificationState();
      this.notificationsSettingsCategories = await this.getNotificationsSettings();
    } catch (error) {
      this.$q.notify(alertUtil.getError("Failed to load user notifications settings"));
      this.notificationsSettingsCategories = null;
    }
    this.appStore.setPageLoading(false);
  },
  beforeRouteLeave() {
    if (!this.notificationsSettingsCategories) return true;
    return requestToLeave();
  },
  computed: {
    showAccessKeys(): boolean {
      // Show access keys section only if all clusters have a version greater than or equal to MIN_CLUSTER_VERSION_FOR_ACCESS_KEYS
      return this.clusterStore.clusterList.every((cluster: ClustersMinimalData) =>
        this.isSufficientAccessKeysVersion(cluster),
      );
    },
    showPersonalCredentials(): boolean {
      return this.clusterStore.isAtleastOneClusterSufficientForVersion(MIN_CLUSTER_VERSION_FOR_USER_CREDENTIALS);
    },
  },
  methods: {
    async getNotificationsSettings(): Promise<CategorizedNotificationSettings> {
      const subscriptions = await notificationService.getUserSubscriptions();
      if (subscriptions?.length) this.currentSubscription = subscriptions[0];

      const notificationTypes: Array<NotificationType> = await notificationService.getUserNotificationTypes();
      const workloadNotificationTypes = notificationTypes.filter((notificationType) =>
        workloadNotificationCategories.includes(notificationType.category),
      );
      const notificationSettings: Array<NotificationSettings> = workloadNotificationTypes.map((notificationType) => ({
        ...notificationType,
        enabled: false,
      }));
      return userNotificationsUtil.categorizeAndUpdateNotificationSettings(notificationSettings, subscriptions);
    },
    async saveNotificationsSettings(subscription: Subscription) {
      if (this.isSavingNotifications) return;
      this.isSavingNotifications = true;

      try {
        if (!this.currentSubscription?.id) {
          await notificationService.createUserSubscription(subscription);
        } else {
          await notificationService.updateUserSubscription(this.currentSubscription.id, subscription);
        }
        this.notificationsSettingsCategories = await this.getNotificationsSettings();
        this.$q.notify(alertUtil.getSuccess("Settings saved"));
      } catch (error) {
        this.$q.notify(alertUtil.getError("Failed to save user notifications settings"));
      } finally {
        this.isSavingNotifications = false;
      }
    },
    isSufficientAccessKeysVersion(cluster: ClustersMinimalData) {
      return (
        cluster.version == "" ||
        cluster.version?.startsWith(CLUSTER_PR_TO_MASTER_VERSION_PREFIX) ||
        (cluster.uuid && cluster.version && isNewerVersion(cluster.version, MIN_CLUSTER_VERSION_FOR_ACCESS_KEYS))
      );
    },
  },
});
</script>
