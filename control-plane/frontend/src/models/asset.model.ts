/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Scope } from "@/swagger-models/assets-service-client";

// HEADS-UP: for ALL  (tenant scope) - DO NOT SEND SCOPE (undefined)
//  for project scope DO NOT SEND SCOPE - only projectId
export interface IAssetNameCountOptions {
  projectId?: number | null;
  departmentId?: string | null;
  clusterId?: string | null;
}

export interface IAssetNameValidationOptions {
  scopeType: Scope;
  scopeId: string;
}
