/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";

export const credentialsArrMock: CredentialsListResponseEntry[] = [
  {
    meta: {
      name: "hf",
      scope: "project",
      clusterId: "ee80a811-4eb8-47c1-a83c-3d58ae79695e",
      projectId: 4500001,
      id: "2923400b-22fd-4be2-82d6-22854f225772",
      kind: "genericSecret",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2025-02-04T08:33:16.724453Z",
      updatedBy: "test@run.ai",
      updatedAt: "2025-02-04T08:33:16.724453Z",
      deletedAt: "0001-01-01T00:00:00Z",
      projectName: "p100",
    },
    spec: {
      genericSecret: {
        keyValuePairs: [
          {
            key: "********",
            value: "********",
          },
        ],
        keyId: 0,
      },
    },
    clusterInfo: {
      resources: [
        {
          name: "genericsecret-hf",
        },
      ],
    },
  },
];
