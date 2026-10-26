/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum EUserType {
  REGULAR_USER = "regular-user",
  SSO_USER = "sso-user",
}

export interface IMe {
  username: string;
  identityProvider: string;
  id: string;
  tenantId: number;
  firstName?: string;
  lastName?: string;
}
// this is how we get the values from the jwt
export interface IAuthProviderConfig {
  email: string;
  identity_provider: string;
  tenant_id: number;
  subject_uuid: string;
}

export interface IUser {
  email: string;
  name: string;
  roles: Array<string>;
  entityType: EUserType;
  needToChangePassword: boolean;
  permitAllClusters: boolean;
  permittedClusters: [];
  username: string;
  tenantId?: number;
  userId?: string;
  lastUpdated?: string;
  password?: string;
  id: string;
  createdAt?: string;
  lastLogin?: string;
  createdBy?: string;
  updatedAt?: string;
  tempPassword?: string;
  isLocal?: boolean;
  identityProvider?: string;
  groups?: Array<string>;
  firstName?: string;
  lastName?: string;
}
