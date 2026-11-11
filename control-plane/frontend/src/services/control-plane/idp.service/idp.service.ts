/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

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

// A dry run rethrows the raw error: the 400 body carries the parsed config.
async function addIdp(idp: IdpCreationRequest, dryRun = false): Promise<IdpPostResponse> {
  try {
    const response = await identityManagerApi.idpsApi.createIdp(dryRun ? { ...idp, dryRun } : idp);
    return response.data;
  } catch (err: unknown) {
    throw dryRun ? err : httpService.handleHttpError(err);
  }
}

async function updateIdp(
  idpAlias: string,
  idpCreationRequest: IdpCreationRequest,
  dryRun = false,
): Promise<IdpPostResponse | undefined> {
  try {
    const request = dryRun ? { ...idpCreationRequest, dryRun } : idpCreationRequest;
    const response = await identityManagerApi.idpsApi.updateIdp(idpAlias, request);
    // The generated client types the update response as void, but a dry run returns the parsed config.
    return response.data as unknown as IdpPostResponse | undefined;
  } catch (err: unknown) {
    throw dryRun ? err : httpService.handleHttpError(err);
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
