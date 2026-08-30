import { bffServiceApi } from "@/services/infra/client-apis/bff-service-api/bff-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  WorkloadsSettingsKeyEnum,
  type GetEnableNIM,
  type GetDeletedWorkloadsView,
  type GetAllowGPT2Model,
  type GetEnableModelCatalog,
  type WorkloadsSettingsGetAllResponse,
} from "@/swagger-models/bff-service-client";
import type {
  IEnableNIMSetting,
  IDeletedWorkloadsViewSetting,
  IAllowGPT2ModelSetting,
  IEnableModelCatalogSetting,
} from "@/models/setting.model";

export interface IWorkloadsSettings {
  enableNIM: IEnableNIMSetting;
  deletedWorkloadsView: IDeletedWorkloadsViewSetting;
  allowGPT2Model: IAllowGPT2ModelSetting;
  enableModelCatalog: IEnableModelCatalogSetting;
}

export const workloadsSettingsService = {
  getWorkloadsSettings,
  getEnableNIM,
  updateEnableNIM,
  getDeletedWorkloadsView,
  getAllowGPT2Model,
  updateAllowGPT2Model,
  getEnableModelCatalog,
  updateEnableModelCatalog,
};

async function getEnableNIM(): Promise<IEnableNIMSetting> {
  try {
    const response = await bffServiceApi.settingsApi.getWorkloadsSettingByKey(WorkloadsSettingsKeyEnum.EnableNim);
    const wrapped = response.data as GetEnableNIM;
    return {
      value: Boolean(wrapped.value?.enabled),
      displayName: wrapped.displayName,
      description: wrapped.description,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateEnableNIM(enabled: boolean): Promise<void> {
  try {
    await bffServiceApi.settingsApi.updateWorkloadsSettingByKey(WorkloadsSettingsKeyEnum.EnableNim, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getWorkloadsSettings(): Promise<IWorkloadsSettings> {
  try {
    const response = await bffServiceApi.settingsApi.getWorkloadsSettings();
    const data = response.data as WorkloadsSettingsGetAllResponse;

    return {
      enableNIM: {
        value: Boolean(data.enableNim?.value?.enabled),
        displayName: data.enableNim?.displayName,
        description: data.enableNim?.description,
      },
      deletedWorkloadsView: {
        value: Boolean(data.deletedWorkloadsView?.value?.enabled),
        displayName: data.deletedWorkloadsView?.displayName,
        description: data.deletedWorkloadsView?.description,
      },
      allowGPT2Model: {
        value: Boolean(data.allowGPT2Model?.value?.enabled),
        displayName: data.allowGPT2Model?.displayName,
        description: data.allowGPT2Model?.description,
        stage: data.allowGPT2Model?.stage,
      },
      enableModelCatalog: {
        value: Boolean(data.enableModelCatalog?.value?.enabled),
        displayName: data.enableModelCatalog?.displayName,
        description: data.enableModelCatalog?.description,
        stage: data.enableModelCatalog?.stage,
      },
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getDeletedWorkloadsView(): Promise<IDeletedWorkloadsViewSetting> {
  try {
    const response = await bffServiceApi.settingsApi.getWorkloadsSettingByKey(
      WorkloadsSettingsKeyEnum.DeletedWorkloadsView,
    );
    const wrapped = response.data as GetDeletedWorkloadsView;
    return {
      value: Boolean(wrapped.value?.enabled),
      displayName: wrapped.displayName,
      description: wrapped.description,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getAllowGPT2Model(): Promise<IAllowGPT2ModelSetting> {
  try {
    const response = await bffServiceApi.settingsApi.getWorkloadsSettingByKey(WorkloadsSettingsKeyEnum.AllowGpt2Model);
    const wrapped = response.data as GetAllowGPT2Model;
    return {
      value: Boolean(wrapped.value?.enabled),
      displayName: wrapped.displayName,
      description: wrapped.description,
      stage: wrapped.stage,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateAllowGPT2Model(enabled: boolean): Promise<void> {
  try {
    await bffServiceApi.settingsApi.updateWorkloadsSettingByKey(WorkloadsSettingsKeyEnum.AllowGpt2Model, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getEnableModelCatalog(): Promise<IEnableModelCatalogSetting> {
  try {
    const response = await bffServiceApi.settingsApi.getWorkloadsSettingByKey(
      WorkloadsSettingsKeyEnum.EnableModelCatalog,
    );
    const wrapped = response.data as GetEnableModelCatalog;
    return {
      value: Boolean(wrapped.value?.enabled),
      displayName: wrapped.displayName,
      description: wrapped.description,
      stage: wrapped.stage,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateEnableModelCatalog(enabled: boolean): Promise<void> {
  try {
    await bffServiceApi.settingsApi.updateWorkloadsSettingByKey(WorkloadsSettingsKeyEnum.EnableModelCatalog, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
