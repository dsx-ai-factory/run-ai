import type { Idp, IdpCreationRequest, IdpPostResponse, IdpResources } from "@/swagger-models/identity-manager-client";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";
import { httpService } from "@/services/infra/https.service/http.service";

export const idpService = {
  removeIdp,
  getIdps,
  addIdp,
  updateIdp,
  getIdpResource,
};

async function getIdps(): Promise<Idp[]> {
  try {
    const response = await identityManagerApi.idpsApi.getIdps();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function addIdp(idp: IdpCreationRequest): Promise<IdpPostResponse> {
  try {
    const response = await identityManagerApi.idpsApi.createIdp(idp);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateIdp(idpAlias: string, idpCreationRequest: IdpCreationRequest): Promise<void> {
  try {
    const response = await identityManagerApi.idpsApi.updateIdp(idpAlias, idpCreationRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function removeIdp(idpAlias: string): Promise<void> {
  try {
    const response = await identityManagerApi.idpsApi.deleteIdp(idpAlias);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getIdpResource(idp: string): Promise<IdpResources> {
  try {
    const response = await identityManagerApi.idpsApi.getIdpResources(idp);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
