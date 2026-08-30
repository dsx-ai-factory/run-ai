import { defineStore } from "pinia";
import { notificationService } from "@/services/control-plane/notifications.service/notification.service";
import type { NotificationChannelMeta } from "@/swagger-models/notifications-proxy-client";

export const useNotificationStore = defineStore("notifications", {
  state: () => ({
    channels: [] as NotificationChannelMeta[],
    isChannelsLoaded: false,
  }),
  getters: {
    isEmailChannelAvailable(): boolean {
      return this.channels.some((channel) => channel.name === "runai-email");
    },
  },
  actions: {
    async loadNotificationChannels(): Promise<void> {
      try {
        this.channels = await notificationService.getNotificationChannels();
        this.isChannelsLoaded = true;
      } catch (err) {
        console.warn("Failed to load notification channels:", err);
        this.channels = [];
      }
    },
  },
});
