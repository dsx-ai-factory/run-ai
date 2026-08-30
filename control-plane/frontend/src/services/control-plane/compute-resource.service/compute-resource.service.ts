// services
import { pick } from "@/utils/common.util";

// models
import type {
  ComputeAsset,
  ComputeCreationRequest,
  ComputeUpdateRequest,
  HttpResponse,
} from "@/swagger-models/assets-service-client";
import type { IAssetsFilter } from "@/models/filter.model";
import type { IUIComputeResourceCreation } from "@/models/compute-resource.model";
import { assetsServiceApi } from "@/services/infra/client-apis/assets-service-api/assets-service-api";
import { httpService } from "@/services/infra/https.service/http.service";

export const computeResourceService = {
  list,
  getEmptyComputeResourceModel,
  remove,
  create,
  getById,
  update,
};

// api calls
async function list(filterBy: IAssetsFilter = {}, filterByName?: string): Promise<Array<ComputeAsset>> {
  const filters: IAssetsFilter = pick(
    filterBy,
    "projectId",
    "departmentId",
    "scope",
    "usageInfo",
    "complyToProject",
    "complyToWorkloadType",
    "clusterUuid",
    "complyToReplicaType",
  );
  try {
    return assetsServiceApi.computeApi
      .listComputeAssets(
        filterByName || undefined,
        filters.scope,
        filters.projectId ? Number(filters.projectId) : undefined,
        filters.departmentId ? String(filters.departmentId) : undefined,
        filters.clusterUuid ? String(filters.clusterUuid) : undefined,
        filters.usageInfo,
        filters.complyToProject,
        filters.complyToWorkloadType,
        filters.complyToReplicaType,
      )
      .then((res) => res.data.entries);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getById(computeResourceId: string): Promise<ComputeAsset> {
  try {
    const response = await assetsServiceApi.computeApi.getComputeAssetById(computeResourceId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function create(computeResource: ComputeCreationRequest): Promise<ComputeAsset> {
  try {
    const response = await assetsServiceApi.computeApi.createComputeAsset(computeResource);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function remove(computeResourceId: string): Promise<HttpResponse> {
  try {
    const response = await assetsServiceApi.computeApi.deleteComputeAssetById(computeResourceId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function update(computeId: string, computeResource: ComputeUpdateRequest): Promise<ComputeAsset> {
  try {
    const response = await assetsServiceApi.computeApi.updateComputeAssetById(computeId, computeResource);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

function getEmptyComputeResourceModel(): IUIComputeResourceCreation {
  return {
    meta: {
      name: "",
      description: null,
      scope: null,
      projectId: null,
    },
    spec: {
      gpuDevicesRequest: 0,
      largeShmRequest: false,
    },
  };
}
