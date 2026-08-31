/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export interface IHybridEnvironmentSectionDisableFields {
  servingPort?: string;
  protocol?: boolean;
  tools?: string;
  toolsAccess?: string;
  security?: {
    uidGidSource?: string;
    capabilities?: string;
  };
}
