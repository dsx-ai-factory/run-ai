import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  WorkloadsSettingsKeyEnum,
  type WorkloadsSettingsGetResponse,
  type GetEnablePolicies,
  type GetEnableHybridSubmission,
  type GetEnableTemplatesV2,
  type GetGpuOverProvisioning,
} from "@/swagger-models/workloads-client";
import type {
  IEnablePoliciesSetting,
  IEnableHybridSubmissionSetting,
  IEnableTemplatesV2Setting,
  IGpuOverProvisioningSetting,
} from "@/models/setting.model";

export interface IWorkloadsServiceSettings {
  enablePolicies: IEnablePoliciesSetting;
  enableHybridSubmission: IEnableHybridSubmissionSetting;
  enableTemplatesV2: IEnableTemplatesV2Setting;
  gpuOverProvisioning: IGpuOverProvisioningSetting;
}

export const workloadsServiceSettingsService = {
  getAll,
  getEnablePolicies,
  updateEnablePolicies,
  getEnableHybridSubmission,
  updateEnableHybridSubmission,
  getEnableTemplatesV2,
  updateEnableTemplatesV2,
  getGpuOverProvisioning,
  updateGpuOverProvisioning,
};

async function getAll(): Promise<IWorkloadsServiceSettings> {
  try {
    const response = await workloadServiceApi.settingsApi.getWorkloadsServiceSettings();
    const data = response.data as WorkloadsSettingsGetResponse;

    return {
      enablePolicies: {
        value: Boolean(data.enablePolicies?.value?.enabled),
        displayName: data.enablePolicies?.displayName,
        description: data.enablePolicies?.description,
        stage: data.enablePolicies?.stage,
      },
      enableHybridSubmission: {
        value: Boolean(data.enableHybridSubmission?.value?.enabled),
        displayName: data.enableHybridSubmission?.displayName,
        description: data.enableHybridSubmission?.description,
        stage: data.enableHybridSubmission?.stage,
      },
      enableTemplatesV2: {
        value: Boolean(data.enableTemplatesV2?.value?.enabled),
        displayName: data.enableTemplatesV2?.displayName,
        description: data.enableTemplatesV2?.description,
        stage: data.enableTemplatesV2?.stage,
      },
      gpuOverProvisioning: {
        value: Boolean(data.gpuOverProvisioning?.value?.enabled),
        displayName: data.gpuOverProvisioning?.displayName,
        description: data.gpuOverProvisioning?.description,
        stage: data.gpuOverProvisioning?.stage,
      },
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getEnablePolicies(): Promise<IEnablePoliciesSetting> {
  try {
    const response = await workloadServiceApi.settingsApi.getWorkloadsServiceSettingByKey(
      WorkloadsSettingsKeyEnum.EnablePolicies,
    );
    const data = response.data as GetEnablePolicies;
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

async function updateEnablePolicies(enabled: boolean): Promise<void> {
  try {
    await workloadServiceApi.settingsApi.updateWorkloadsServiceSettingByKey(WorkloadsSettingsKeyEnum.EnablePolicies, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getEnableHybridSubmission(): Promise<IEnableHybridSubmissionSetting> {
  try {
    const response = await workloadServiceApi.settingsApi.getWorkloadsServiceSettingByKey(
      WorkloadsSettingsKeyEnum.EnableHybridSubmission,
    );
    const data = response.data as GetEnableHybridSubmission;
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

async function updateEnableHybridSubmission(enabled: boolean): Promise<void> {
  try {
    await workloadServiceApi.settingsApi.updateWorkloadsServiceSettingByKey(
      WorkloadsSettingsKeyEnum.EnableHybridSubmission,
      { enabled },
    );
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getEnableTemplatesV2(): Promise<IEnableTemplatesV2Setting> {
  try {
    const response = await workloadServiceApi.settingsApi.getWorkloadsServiceSettingByKey(
      WorkloadsSettingsKeyEnum.EnableTemplatesV2,
    );
    const data = response.data as GetEnableTemplatesV2;
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

async function updateEnableTemplatesV2(enabled: boolean): Promise<void> {
  try {
    await workloadServiceApi.settingsApi.updateWorkloadsServiceSettingByKey(WorkloadsSettingsKeyEnum.EnableTemplatesV2, {
      enabled,
    });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getGpuOverProvisioning(): Promise<IGpuOverProvisioningSetting> {
  try {
    const response = await workloadServiceApi.settingsApi.getWorkloadsServiceSettingByKey(
      WorkloadsSettingsKeyEnum.GpuOverProvisioning,
    );
    const data = response.data as GetGpuOverProvisioning;
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

async function updateGpuOverProvisioning(enabled: boolean): Promise<void> {
  try {
    await workloadServiceApi.settingsApi.updateWorkloadsServiceSettingByKey(
      WorkloadsSettingsKeyEnum.GpuOverProvisioning,
      { enabled },
    );
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
