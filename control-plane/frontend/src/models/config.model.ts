/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export enum EEnvTypes {
  Local = "local",
  OnPremise = "on-premise",
  Staging = "staging",
  Test = "test",
  CustomerTest = "customer-test",
  CustomerStaging = "customer-staging",
  CustomerCloud = "customer-cloud",
  Cloud = "cloud",
  Unknown = "unknown",
}
export type IConfig = {
  authClientID: string;
  tenantName: string;
  authHost: string;
  isMultiTenant: boolean;
  redirectUrl: string;
};
