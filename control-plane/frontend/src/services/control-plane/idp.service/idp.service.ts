/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type {
  Idp,
  IdpCreationRequest,
  IdpPostResponse,
  IdpResources,
  SamlCreationData,
  SamlData,
} from "@/swagger-models/identity-manager-client";
import { IdpCreationRequestTypeEnum } from "@/swagger-models/identity-manager-client";
import { isAxiosError } from "axios";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";
import { httpService } from "@/services/infra/https.service/http.service";

export const idpService = {
  removeIdp,
  getIdps,
  addIdp,
  updateIdp,
  getIdpResource,
  validateSamlCertificate,
};

export interface SamlCertificateValidation {
  samlData?: SamlData;
  errorMessage?: string;
}

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

// Dry run: nothing is saved. A rejected config (e.g. an expired certificate) is returned, not thrown,
// because the 400 body still carries the parsed certificate expiration date to show.
async function validateSamlCertificate(
  samlData: SamlCreationData,
  idpAlias?: string,
): Promise<SamlCertificateValidation> {
  const request: IdpCreationRequest = { type: IdpCreationRequestTypeEnum.Saml, samlData, dryRun: true };
  try {
    if (idpAlias) {
      // The generated client types the update response as void, but a dry run returns the parsed config.
      const response = await identityManagerApi.idpsApi.updateIdp(idpAlias, request);
      return { samlData: (response.data as unknown as IdpPostResponse | undefined)?.samlData ?? undefined };
    }
    const response = await identityManagerApi.idpsApi.createIdp(request);
    return { samlData: response.data.samlData ?? undefined };
  } catch (err: unknown) {
    if (isAxiosError(err) && err.response?.status === 400) {
      const body = err.response.data as { samlData?: SamlData; message?: string } | undefined;
      return { samlData: body?.samlData ?? undefined, errorMessage: body?.message || "Invalid SAML metadata" };
    }
    throw httpService.handleHttpError(err);
  }
}
