/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export interface ILegacyRoleTransitionRow {
  id: string;
  legacyRole: string;
  roleId: number;
  accessRuleCount: number;
  newRole: string | undefined;
}
