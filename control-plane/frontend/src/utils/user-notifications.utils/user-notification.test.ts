import { describe, it, expect } from "vitest";
import { uniqueEmailChannelRunaiId } from "@/models/user-setting.model";
import type { NotificationSettings, CategorizedNotificationSettings } from "@/models/user-setting.model";
import type { Subscription } from "@/swagger-models/notifications-proxy-client";
import { userNotificationsUtil } from "@/utils/user-notifications.utils/user-notifications.util";

const { categorizeAndUpdateNotificationSettings, convertNotificationSettingsToSubscription } = userNotificationsUtil;

describe("User Notification Utilities", () => {
  describe("categorizeAndUpdateNotificationSettings", () => {
    it("should categorize and update notification settings correctly", () => {
      const notificationSettings: NotificationSettings[] = [
        { category: "category1", type: "type1", enabled: false },
        { category: "category1", type: "type2", enabled: true },
        { category: "category2", type: "type1", enabled: true },
      ];

      const subscriptions: Subscription[] = [
        {
          channels: [uniqueEmailChannelRunaiId],
          id: "subscription-123",
          events: [
            { category: "category1", types: ["type1"] },
            { category: "category2", types: ["type1", "type2"] },
          ],
        },
      ];

      const categorizedSettings: CategorizedNotificationSettings = categorizeAndUpdateNotificationSettings(
        notificationSettings,
        subscriptions,
      );

      expect(categorizedSettings).toEqual({
        category1: [
          { category: "category1", type: "type1", enabled: true },
          { category: "category1", type: "type2", enabled: false },
        ],
        category2: [{ category: "category2", type: "type1", enabled: true }],
      });
    });
  });

  describe("convertNotificationSettingsToSubscription", () => {
    it("should convert categorized settings to a subscription object", () => {
      const categorizedSettings: CategorizedNotificationSettings = {
        category1: [
          { category: "category1", type: "type1", enabled: true },
          { category: "category1", type: "type2", enabled: false },
        ],
        category2: [{ category: "category2", type: "type1", enabled: true }],
      };

      const typeState = {
        type1: true,
        type2: false,
      };

      const subscription: Subscription = convertNotificationSettingsToSubscription(
        categorizedSettings,
        [uniqueEmailChannelRunaiId],
        typeState,
      );

      expect(subscription).toEqual({
        events: [
          { category: "category1", types: ["type1"] },
          { category: "category2", types: ["type1"] },
        ],
        channels: [uniqueEmailChannelRunaiId],
      });
    });
  });
});
