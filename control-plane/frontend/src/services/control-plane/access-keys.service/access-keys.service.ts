/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { httpService } from "@/services/infra/https.service/http.service";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";
import type {
  AccessKey,
  AccessKeyCreationRequest,
  AccessKeyPostResponse,
  RegenerateAccessKeySecret200Response,
} from "@/swagger-models/identity-manager-client";

export const accessKeysService = {
  getAccessKeys,
  getAccessKey,
  createAccessKey,
  removeAccessKey,
  generateAccessKeySecret,
};

async function getAccessKeys(): Promise<AccessKey[]> {
  try {
    const response = await identityManagerApi.accessKeysApi.getAccessKeys();
    return response.data.accessKeys;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getAccessKey(accessKeyId: string): Promise<AccessKey> {
  try {
    const response = await identityManagerApi.accessKeysApi.getAccessKeyById(accessKeyId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createAccessKey(accessKey: AccessKeyCreationRequest): Promise<AccessKeyPostResponse> {
  try {
    const response = await identityManagerApi.accessKeysApi.createAccessKey(accessKey);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function removeAccessKey(accessKeyId: string): Promise<void> {
  try {
    await identityManagerApi.accessKeysApi.deleteAccessKeyById(accessKeyId);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function generateAccessKeySecret(accessKeyId: string): Promise<RegenerateAccessKeySecret200Response> {
  try {
    const response = await identityManagerApi.accessKeysApi.regenerateAccessKeySecret(accessKeyId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
