import { orgUnitServiceApi } from "@/services/infra/client-apis/org-unit-service-api/org-unit-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  OrgUnitSettingsKeyEnum,
  type OrgUnitSettingsGetResponse,
  type OrgUnitSetting,
} from "@/swagger-models/org-unit-service-client";
import type { ILimitQuotaOverSubscriptionSetting } from "@/models/setting.model";

export interface IOrgUnitServiceSettings {
  limitQuotaOverSubscription: ILimitQuotaOverSubscriptionSetting;
}

export const orgUnitSettingsService = {
  getAll,
  getLimitQuotaOverSubscription,
  updateLimitQuotaOverSubscription,
};

async function getAll(): Promise<IOrgUnitServiceSettings> {
  try {
    const response = await orgUnitServiceApi.settingsApi.getOrgUnitSettings();
    const data = response.data as OrgUnitSettingsGetResponse;

    return {
      limitQuotaOverSubscription: {
        value: Boolean(data.limitQuotaOverSubscription?.enabled),
        displayName: data.limitQuotaOverSubscription?.displayName,
        description: data.limitQuotaOverSubscription?.description,
        stage: data.limitQuotaOverSubscription?.stage,
      },
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getLimitQuotaOverSubscription(): Promise<ILimitQuotaOverSubscriptionSetting> {
  try {
    const response = await orgUnitServiceApi.settingsApi.getOrgUnitSettingByKey(
      OrgUnitSettingsKeyEnum.LimitQuotaOverSubscription,
    );
    const data = response.data as OrgUnitSetting;
    return {
      value: Boolean(data.enabled),
      displayName: data.displayName,
      description: data.description,
      stage: data.stage,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateLimitQuotaOverSubscription(enabled: boolean): Promise<void> {
  try {
    await orgUnitServiceApi.settingsApi.updateOrgUnitSettingByKey(OrgUnitSettingsKeyEnum.LimitQuotaOverSubscription, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
