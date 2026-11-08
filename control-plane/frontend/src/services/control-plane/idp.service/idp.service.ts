/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type {
  Idp,
  IdpCreationRequest,
  IdpPostResponse,
  IdpResources,
  SamlData,
} from "@/swagger-models/identity-manager-client";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";
import { httpService } from "@/services/infra/https.service/http.service";

export const idpService = {
  removeIdp,
  getIdps,
  addIdp,
  updateIdp,
  dryRunIdp,
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

// dryRun validates SAML and returns the parsed config without storing it.
// An existing identity provider is validated through update; a new one through create.
// A failed validation is HTTP 400 whose body is still that config, so keep it on the thrown error.
// The generated update client types the body as empty, but the server still returns the config.
async function dryRunIdp(idp: IdpCreationRequest, idpAlias?: string): Promise<IdpPostResponse> {
  try {
    const request = { ...idp, dryRun: true };
    const response = idpAlias
      ? await identityManagerApi.idpsApi.updateIdp(idpAlias, request)
      : await identityManagerApi.idpsApi.createIdp(request);
    return response.data as IdpPostResponse;
  } catch (err: unknown) {
    const samlData = samlDataFromError(err);
    try {
      httpService.handleHttpError(err);
    } catch (handled: unknown) {
      if (samlData && typeof handled === "object" && handled !== null) {
        (handled as { samlData?: SamlData }).samlData = samlData;
      }
      throw handled;
    }
    const failure = new Error("SAML metadata is invalid") as Error & { samlData?: SamlData };
    failure.samlData = samlData;
    throw failure;
  }
}

function samlDataFromError(err: unknown): SamlData | undefined {
  if (typeof err !== "object" || err === null || !("response" in err)) return undefined;
  const data = (err as { response?: { data?: unknown } }).response?.data;
  if (typeof data !== "object" || data === null || !("samlData" in data)) return undefined;
  return (data as IdpPostResponse).samlData ?? undefined;
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
