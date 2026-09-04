/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export interface IJwtPayload {
  iss?: string;
  sub?: string;
  aud?: string[] | string;
  exp?: number;
  nbf?: number;
  iat?: number;
  jti?: string;
  sid?: string; // Session ID
  subject_uuid?: string; // User ID
}

export interface IRefreshTokenResponse {
  idToken: string;
  refreshToken: string;
  accessToken: string;
}
