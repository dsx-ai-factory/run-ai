import { defineStore } from "pinia";
import type { SettingsGetResponse, SettingsKeyEnum } from "@/swagger-models/identity-manager-client";
import { securitySettingsService } from "@/services/control-plane/security-settings.service/security-settings.service";

export const useSecuritySettingsStore = defineStore("security-settings", {
  state: () => ({
    settings: null as SettingsGetResponse | null,
  }),

  actions: {
    async getSettings(forceRefresh = false): Promise<SettingsGetResponse | null> {
      if (this.settings !== null && !forceRefresh) {
        return this.settings;
      }
      try {
        this.settings = await securitySettingsService.getSettings();
        return this.settings;
      } catch (error) {
        console.error("Failed to get security settings", error);
        throw error;
      }
    },

    async getSettingByKey<T extends object = object>(key: SettingsKeyEnum): Promise<T | null> {
      if (this.settings?.[key as keyof typeof this.settings]) {
        return this.settings[key as keyof typeof this.settings] as T;
      }

      try {
        const setting = await securitySettingsService.getSettingByKey(key as keyof typeof this.settings);
        this.settings = {
          ...this.settings,
          [key]: setting,
        } as SettingsGetResponse;

        return setting as T;
      } catch (error) {
        console.error(`Failed to get security setting by key ${key}:`, error);
        throw error;
      }
    },

    async updateSetting(key: SettingsKeyEnum, value: object): Promise<void> {
      try {
        await securitySettingsService.updateSettingByKey(key, value);
        if (this.settings) {
          this.settings = {
            ...this.settings,
            [key]: { value },
          };
        }
      } catch (error) {
        console.error(`Failed to update ${key} setting:`, error);
        throw error;
      }
    },
  },
});
