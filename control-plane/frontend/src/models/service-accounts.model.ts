/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { TRolesTypes } from "./roles.model";
import type { ServiceAccount } from "@/swagger-models/identity-manager-client";

export interface ServiceAccountWithRoles extends ServiceAccount {
  roles?: TRolesTypes[];
}
