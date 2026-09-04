/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { controlPlaneService } from "@/services/control-plane/control-plane.service/control-plane.service";

import type { IJwtPayload, IRefreshTokenResponse } from "@/models/auth.model";

import { API } from "@/common/api.constant";
import jwt_decode from "jwt-decode";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";

export const authService = {
  refreshToken,
  exchangeCode,
  changePassword,
  isValidSsoEmailMapperToken,
};

async function refreshToken(): Promise<IRefreshTokenResponse> {
  return await controlPlaneService.post(`${API.v1}/token`, {
    grantType: "refresh_token",
    clientID: "runai-admin-ui",
  });
}

function exchangeCode(code: string, redirectUri: string, codeVerifier?: string): Promise<IRefreshTokenResponse> {
  return controlPlaneService.post(`${API.v1}/token`, {
    grantType: "exchange_token",
    clientID: "runai-admin-ui",
    code,
    redirectUri,
    ...(codeVerifier ? { codeVerifier } : {}),
  });
}

function changePassword(newPassword: string, currentPassword: string): Promise<unknown> {
  return identityManagerApi.usersApi.changeUserPassword({ newPassword, currentPassword });
}

function isValidSsoEmailMapperToken(accessToken: string): boolean {
  const tokenPayload: IJwtPayload = jwt_decode(accessToken);
  return "email" in tokenPayload;
}
