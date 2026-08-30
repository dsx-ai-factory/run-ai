import { controlPlaneService } from "@/services/control-plane/control-plane.service/control-plane.service";
import { API } from "@/common/api.constant";
import { HttpErrorResponse } from "@/models/http-response.model";
import type { IAssetNameValidationOptions } from "@/models/asset.model";
import type { AssetFlatRequest, AssetFlatResponse } from "@/swagger-models/assets-service-client";
import { assetsServiceApi } from "@/services/infra/client-apis/assets-service-api/assets-service-api";
import type { AxiosResponse } from "axios";
import { httpService } from "@/services/infra/https.service/http.service";

export const assetService = {
  isExist,
  requestAssetFlatFormat,
};

const endpoint: string = API.asset;
async function isExist(kind: string, assetName: string, options: IAssetNameValidationOptions): Promise<boolean> {
  //TODO: is says this is deprecated, verifying with Ofer
  // assetsServiceApi.workloadApi.assetNameConflictValidation(kind, assetName, options.scopeType, options.scopeId);
  try {
    await controlPlaneService.get(`${endpoint}/name-availability`, {
      kind,
      assetName,
      ...options,
    });
    return false;
  } catch (error: unknown) {
    // 409 if its taken
    // other error status are other issues
    if (error instanceof HttpErrorResponse) {
      if (error.statusCode === 409) return true;
      console.error(error.serialize());
    }
    throw error;
  }
}

async function requestAssetFlatFormat(assetFlatRequest: AssetFlatRequest): Promise<AssetFlatResponse> {
  try {
    const response: AxiosResponse = await assetsServiceApi.assetsApi.requestAssetFlatFormat(assetFlatRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
