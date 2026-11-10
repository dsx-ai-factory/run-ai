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
import { IdpCreationRequestTypeEnum, IdpTypeEnum } from "@/swagger-models/identity-manager-client";
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

// Alias used to validate metadata for an IdP that does not exist yet: a dry run never looks the IdP up or stores it.
const NEW_IDP_DRY_RUN_ALIAS = IdpTypeEnum.Saml;

/**
 * Validates SAML metadata on the backend without saving anything (dryRun) and returns the parsed
 * signing certificate details, including its expiration date.
 * It always uses the update endpoint: unlike create, a rejected update still returns the parsed config
 * (with the expiration date) in the response body. Pass the alias of the IdP being edited; omit it for a new IdP.
 * A rejected configuration (e.g. expired certificate) is returned as a result, not thrown, so the caller can
 * display the expiration details the backend sends back.
 */
async function validateSamlCertificate(
  samlData: SamlCreationData,
  idpAlias?: string,
): Promise<SamlCertificateValidation> {
  const request: IdpCreationRequest = { type: IdpCreationRequestTypeEnum.Saml, samlData, dryRun: true };
  try {
    // The generated client types the update response as void, but a dry run returns the parsed config.
    const response = await identityManagerApi.idpsApi.updateIdp(idpAlias || NEW_IDP_DRY_RUN_ALIAS, request);
    return { samlData: (response.data as unknown as IdpPostResponse | undefined)?.samlData ?? undefined };
  } catch (err: unknown) {
    if (isAxiosError(err) && err.response?.status === 400) {
      const body = err.response.data as { samlData?: SamlData; message?: string } | undefined;
      return { samlData: body?.samlData ?? undefined, errorMessage: body?.message || "Invalid SAML metadata" };
    }
    throw httpService.handleHttpError(err);
  }
}
