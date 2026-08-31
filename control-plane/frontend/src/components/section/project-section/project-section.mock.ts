/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ProjectMetadata } from "@/swagger-models/org-unit-service-client";

export const projectsMockData: ProjectMetadata[] = [
  {
    id: "1",
    clusterId: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
    name: "test-1",
    createdAt: "2021-08-10T09:00:00Z",
    phase: "Active",
    deserved: {
      gpu: 0,
    },
    allocated: {
      gpu: 0,
    },
  },
  {
    id: "2",
    clusterId: "37b28c5a-d60c-40eb-b4a8-ff9ab4d28b81",
    name: "test-2",
    createdAt: "2021-08-10T09:00:00Z",
    phase: "Active",
    deserved: {
      gpu: 0,
    },
    allocated: {
      gpu: 0,
    },
  },
];
