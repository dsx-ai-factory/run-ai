import { bffServiceApi } from "@/services/infra/client-apis/bff-service-api/bff-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import { BrandingSettingsKeyEnum, type GetLogoValue } from "@/swagger-models/bff-service-client";

export const brandingSettingsService = {
  getLogo,
  updateLogo,
  deleteLogo,
};

async function getLogo(): Promise<GetLogoValue | undefined> {
  try {
    const response = await bffServiceApi.settingsApi.getBrandingSettingByKey(BrandingSettingsKeyEnum.Logo);
    const data = response.data as { value?: GetLogoValue };
    return data?.value;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateLogo(logoDataURL: string): Promise<void> {
  try {
    await bffServiceApi.settingsApi.updateBrandingSettingByKey(BrandingSettingsKeyEnum.Logo, {
      logo: logoDataURL,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteLogo(): Promise<void> {
  return updateLogo("");
}
