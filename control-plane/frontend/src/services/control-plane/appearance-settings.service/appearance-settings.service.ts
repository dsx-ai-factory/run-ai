import { bffServiceApi } from "@/services/infra/client-apis/bff-service-api/bff-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  AppearanceSettingsKeyEnum,
  type GetBannerConfig,
  type GetBannerConfigValue,
  type UpdateBannerConfig,
} from "@/swagger-models/bff-service-client";

export const appearanceSettingsService = {
  getBannerConfig,
  updateBannerConfig,
};

async function getBannerConfig(): Promise<GetBannerConfigValue | undefined> {
  try {
    const response = await bffServiceApi.settingsApi.getAppearanceSettingByKey(AppearanceSettingsKeyEnum.BannerConfig);
    const data = response.data as GetBannerConfig;
    return data?.value;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateBannerConfig(config: UpdateBannerConfig): Promise<void> {
  try {
    await bffServiceApi.settingsApi.updateAppearanceSettingByKey(AppearanceSettingsKeyEnum.BannerConfig, config);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
