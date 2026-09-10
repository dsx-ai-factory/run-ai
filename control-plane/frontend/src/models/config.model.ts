/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export type IConfig = {
  authClientID: string;
  tenantName: string;
  authHost: string;
  isMultiTenant: boolean;
  redirectUrl: string;
};
