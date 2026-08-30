import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";
import type { UpdateSecuritySettingByKeyRequest, SettingsGetResponse } from "@/swagger-models/identity-manager-client";
import type { SettingsKeyEnum } from "@/swagger-models/identity-manager-client";
import { httpService } from "@/services/infra/https.service/http.service";

class SecuritySettingsService {
  /**
   * Get all security settings
   */
  async getSettings(): Promise<SettingsGetResponse> {
    try {
      const { data } = await identityManagerApi.settingsApi.getSecuritySettings();
      return data;
    } catch (err: unknown) {
      throw httpService.handleHttpError(err);
    }
  }

  /**
   * Get security setting by key
   * @param key The settings key
   */
  async getSettingByKey(key: SettingsKeyEnum): Promise<object> {
    try {
      const { data } = await identityManagerApi.settingsApi.getSecuritySettingByKey(key);
      return data;
    } catch (err: unknown) {
      throw httpService.handleHttpError(err);
    }
  }

  /**
   * Update security setting by key
   * @param key The settings key
   * @param request The update request
   */
  async updateSettingByKey(key: SettingsKeyEnum, request: UpdateSecuritySettingByKeyRequest): Promise<void> {
    try {
      await identityManagerApi.settingsApi.updateSecuritySettingByKey(key, request);
    } catch (err: unknown) {
      throw httpService.handleHttpError(err);
    }
  }
}

export const securitySettingsService: SecuritySettingsService = new SecuritySettingsService();
