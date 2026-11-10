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

/**
 * Validates SAML metadata on the backend without saving anything (dryRun) and returns the parsed
 * signing certificate details, including its expiration date.
 * The create endpoint is used for both new and existing IdPs: in dry-run mode it only validates and never
 * touches the stored configuration, so no IdP alias is needed.
 * A rejected configuration (e.g. expired certificate) is returned as a result, not thrown, so the caller can
 * display the expiration details the backend sends back.
 */
async function validateSamlCertificate(samlData: SamlCreationData): Promise<SamlCertificateValidation> {
  try {
    const response = await identityManagerApi.idpsApi.createIdp({
      type: IdpCreationRequestTypeEnum.Saml,
      samlData,
      dryRun: true,
    });
    return { samlData: response.data.samlData };
  } catch (err: unknown) {
    if (isAxiosError(err) && err.response?.status === 400) {
      const body = err.response.data as { samlData?: SamlData; message?: string } | undefined;
      return { samlData: body?.samlData, errorMessage: body?.message || "Invalid SAML metadata" };
    }
    throw httpService.handleHttpError(err);
  }
}
