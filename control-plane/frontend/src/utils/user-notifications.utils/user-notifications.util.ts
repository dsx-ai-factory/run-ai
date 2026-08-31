/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { type CategorizedNotificationSettings, type NotificationSettings } from "@/models/user-setting.model";
import type { Subscription, SubscriptionEvent } from "@/swagger-models/notifications-proxy-client";

export const userNotificationsUtil = {
  categorizeAndUpdateNotificationSettings,
  convertNotificationSettingsToSubscription,
  generateSlackChannelId,
};

function extractSubscriptionPairs(subscriptions: Subscription[]): Set<{ category: string | undefined; type: string }> {
  return new Set(
    subscriptions.flatMap((subscription) =>
      subscription.events.flatMap((event) =>
        (event.types || []).map((type) => ({
          category: event.category,
          type,
        })),
      ),
    ),
  );
}

function updateNotificationSetting(
  setting: NotificationSettings,
  subscriptionPairs: Set<{ category: string | undefined; type: string }>,
  categorizedSettings: CategorizedNotificationSettings,
) {
  const isEnabled = Array.from(subscriptionPairs).some(
    (pair) => pair.category === setting.category && pair.type === setting.type,
  );

  if (!categorizedSettings[setting.category]) {
    categorizedSettings[setting.category] = [];
  }

  categorizedSettings[setting.category].push({
    ...setting,
    enabled: isEnabled,
  });
}

function categorizeAndUpdateNotificationSettings(
  notificationSettings: NotificationSettings[],
  subscriptions: Subscription[],
): CategorizedNotificationSettings {
  const subscriptionPairs: Set<{ category: string | undefined; type: string }> = extractSubscriptionPairs(subscriptions);
  const categorizedSettings: CategorizedNotificationSettings = {};
  notificationSettings.forEach((setting) => updateNotificationSetting(setting, subscriptionPairs, categorizedSettings));
  return categorizedSettings;
}

function convertNotificationSettingsToSubscription(
  categorizedSettings: CategorizedNotificationSettings,
  channels: string[],
  typeState: Record<string, boolean>,
): Subscription {
  const events: Array<SubscriptionEvent> = [];

  Object.keys(categorizedSettings).forEach((category) => {
    const types: string[] = [];

    categorizedSettings[category].forEach((setting) => {
      if (typeState[setting.type]) {
        types.push(setting.type);
      }
    });

    if (types.length > 0) {
      events.push({ category, types });
    }
  });

  return {
    events,
    channels,
  };
}

function generateSlackChannelId(tenantId: number): string {
  return `slack-${tenantId}`;
}
