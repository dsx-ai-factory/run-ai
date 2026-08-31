/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { OVER_QUOTA_DISABLED_VALUE, OVER_QUOTA_ENABLED_VALUE } from "@/models/resource.model";

export const defaultResources = [
  {
    nodePool: {
      id: "2",
      name: "default",
    },
    gpu: {
      deserved: 0,
      limit: -1,
      overQuotaWeight: 2,
    },
  },
];

export const resourcesOverQuotaWeight2 = [
  {
    nodePool: {
      id: "2",
      name: "default",
    },
    gpu: {
      deserved: 0,
      limit: -1,
      overQuotaWeight: OVER_QUOTA_ENABLED_VALUE,
    },
  },
];
export const resourcesOverQuotaWeight0 = [
  {
    nodePool: {
      id: "2",
      name: "default",
    },
    gpu: {
      deserved: 0,
      limit: -1,
      overQuotaWeight: OVER_QUOTA_DISABLED_VALUE,
    },
  },
];
