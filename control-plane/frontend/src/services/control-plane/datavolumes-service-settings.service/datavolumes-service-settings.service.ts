import { dataVolumesServiceApi } from "@/services/infra/client-apis/data-volume-service-api/data-volume-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  DatavolumesSettingsKeyEnum,
  type DatavolumesSettingsGetResponse,
  type GetEnableDataVolumes,
} from "@/swagger-models/datavolumes-service-client";
import type { IEnableDataVolumesSetting } from "@/models/setting.model";

export interface IDatavolumesServiceSettings {
  enableDataVolumes: IEnableDataVolumesSetting;
}

export const datavolumesServiceSettingsService = {
  getAll,
  getEnableDataVolumes,
  updateEnableDataVolumes,
};

async function getAll(): Promise<IDatavolumesServiceSettings> {
  try {
    const response = await dataVolumesServiceApi.settingsApi.getDatavolumesSettings();
    const data = response.data as DatavolumesSettingsGetResponse;

    return {
      enableDataVolumes: {
        value: Boolean(data.enableDataVolumes?.value?.enabled),
        displayName: data.enableDataVolumes?.displayName,
        description: data.enableDataVolumes?.description,
        stage: data.enableDataVolumes?.stage,
      },
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getEnableDataVolumes(): Promise<IEnableDataVolumesSetting> {
  try {
    const response = await dataVolumesServiceApi.settingsApi.getDatavolumesSettingByKey(
      DatavolumesSettingsKeyEnum.EnableDataVolumes,
    );
    const data = response.data as GetEnableDataVolumes;
    return {
      value: Boolean(data.value?.enabled),
      displayName: data.displayName,
      description: data.description,
      stage: data.stage,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateEnableDataVolumes(enabled: boolean): Promise<void> {
  try {
    await dataVolumesServiceApi.settingsApi.updateDatavolumesSettingByKey(DatavolumesSettingsKeyEnum.EnableDataVolumes, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
