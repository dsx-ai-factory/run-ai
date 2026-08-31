/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type {
  Datavolume,
  DatavolumeCreationFields,
  DatavolumeNoSharedScopes,
  DatavolumePatchFields,
  SharedScope,
  SharedScopeBase,
  SharedScopesPatchRequest,
} from "@/swagger-models/datavolumes-service-client";
import { dataVolumesServiceApi } from "@/services/infra/client-apis/data-volume-service-api/data-volume-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import type { IDataVolumeCreationRequest } from "@/models/data-source.model";
import { errorMessages } from "@/common/error-message.constant";
import type { HttpError, IValidationRule } from "@/models/global.model";

export const dataVolumeService = {
  listOriginatedDataVolumes,
  createDataVolume,
  patchDataVolumeSharedScopes,
  deleteDataVolume,
  updateDataVolume,
  updateDataVolumeSharedScopes,
  getSharedScopesById,
  getDataVolumeById,
  getEmptyDataVolumeModel,
  patchDataVolume,
  countDataVolumes,
  checkNameAvailability,
  getNameValidationFunc,
  listUsedPvcs,
};

async function listOriginatedDataVolumes(
  offset?: number,
  limit?: number,
  sortBy?: "name" | "projectId" | "clusterId" | "departmentId" | "projectName" | "createdAt" | "createdBy" | "phase",
  sortOrder?: "asc" | "desc",
  filterBy?: Array<string>,
  search?: string,
): Promise<DatavolumeNoSharedScopes[]> {
  const res = await dataVolumesServiceApi.dataVolumesApi.getDatavolumes(
    "Originated",
    undefined,
    offset,
    limit,
    sortBy,
    sortOrder,
    filterBy,
    search,
  );
  return res.data.datavolumes;
}

async function createDataVolume(dataVolumeCreationRequest: DatavolumeCreationFields): Promise<Datavolume> {
  try {
    const response = await dataVolumesServiceApi.dataVolumesApi.createDatavolume(dataVolumeCreationRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function patchDataVolume(id: string, dataVolume: DatavolumePatchFields): Promise<Datavolume> {
  try {
    const response = await dataVolumesServiceApi.dataVolumesApi.patchDatavolume(id, dataVolume);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function patchDataVolumeSharedScopes(
  datavolumeId: string,
  sharedScopes: SharedScopesPatchRequest,
): Promise<Array<SharedScopeBase>> {
  try {
    const response = await dataVolumesServiceApi.dataVolumesApi.patchDatavolumeSharedScopes(datavolumeId, sharedScopes);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteDataVolume(dataVolumeId: string): Promise<void> {
  try {
    await dataVolumesServiceApi.dataVolumesApi.deleteDatavolume(dataVolumeId);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateDataVolume(dataVolumeId: string, dataVolume: DatavolumePatchFields): Promise<Datavolume> {
  try {
    const response = await dataVolumesServiceApi.dataVolumesApi.patchDatavolume(dataVolumeId, dataVolume);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateDataVolumeSharedScopes(
  dataVolumeId: string,
  sharedScopes: SharedScopesPatchRequest,
): Promise<Array<SharedScopeBase>> {
  try {
    const response = await dataVolumesServiceApi.dataVolumesApi.patchDatavolumeSharedScopes(dataVolumeId, sharedScopes);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getSharedScopesById(dataVolumeId: string): Promise<SharedScope[]> {
  try {
    const response = await dataVolumesServiceApi.dataVolumesApi.getDatavolumeSharedScopes(dataVolumeId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getDataVolumeById(dataVolumeId: string): Promise<Datavolume> {
  try {
    const response = await dataVolumesServiceApi.dataVolumesApi.getDatavolume(dataVolumeId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function countDataVolumes(filterBy?: Array<string>, search?: string): Promise<number> {
  try {
    const response = await dataVolumesServiceApi.dataVolumesApi.countDatavolumes(
      "Originated",
      undefined,
      filterBy,
      search,
    );
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function checkNameAvailability(clusterId: string, dvName: string): Promise<boolean | string> {
  try {
    await dataVolumesServiceApi.dataVolumesApi.datavolumeNameAvailability(dvName, clusterId);
    return true;
  } catch (error: unknown) {
    const httpError: HttpError = error as HttpError;
    if (httpError.response && httpError.response.status === 409) {
      return errorMessages.NAME_ALREADY_EXIST;
    }
    throw error;
  }
}

async function listUsedPvcs(projectId: string): Promise<Array<string>> {
  try {
    const response = await dataVolumesServiceApi.dataVolumesApi.getUsedPvcsFromDatavolumes(projectId);
    return response.data.usedPVCs || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

function getNameValidationFunc(clusterId: string): IValidationRule {
  return (val: string) => checkNameAvailability(clusterId, val);
}

function getEmptyDataVolumeModel(): IDataVolumeCreationRequest {
  return {
    meta: {
      name: "",
      projectId: "",
    },
    spec: {
      originPvcName: "",
      sharedScopes: [],
    },
  };
}
