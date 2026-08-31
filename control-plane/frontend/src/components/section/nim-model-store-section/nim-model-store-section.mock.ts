/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";

export const mockDataSources: Array<DatasourceListResponseEntry> = [
  {
    meta: {
      name: "hp100",
      scope: "project",
      clusterId: "9dca9e71-a068-4232-993c-be8875c9c96a",
      projectId: 4500001,
      id: "17264c0b-fcb1-4f9c-99dc-56303e19da62",
      kind: "hostPath",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2024-12-01T08:39:48.827931Z",
      updatedBy: "test@run.ai",
      updatedAt: "2024-12-01T08:39:48.827931Z",
      projectName: "p100",
    },
    spec: {
      hostPath: {
        path: "/usr",
        readOnly: false,
        mountPath: "/home",
      },
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "nfs102",
      scope: "project",
      clusterId: "9dca9e71-a068-4232-993c-be8875c9c96a",
      projectId: 4500001,
      id: "7de192d3-1d8b-48a6-add1-30e2361dd792",
      kind: "nfs",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2024-12-01T09:07:24.471203Z",
      updatedBy: "test@run.ai",
      updatedAt: "2024-12-01T09:07:24.471203Z",
      projectName: "p100",
    },
    spec: {
      nfs: {
        path: "/usr",
        readOnly: false,
        server: "myserver",
        mountPath: "/home",
      },
    },
    compliance: {
      compliance: true,
    },
  },
];

export const unsupportedDatasources: Array<DatasourceListResponseEntry> = [
  {
    meta: {
      name: "git100",
      scope: "project",
      clusterId: "9dca9e71-a068-4232-993c-be8875c9c96a",
      projectId: 4500001,
      id: "17264c0b-fcb1-4f9c-99dc-56303e19da72",
      kind: "git",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2024-12-01T08:39:48.827931Z",
      updatedBy: "test@run.ai",
      updatedAt: "2024-12-01T08:39:48.827931Z",
      projectName: "p100",
    },
    spec: {
      git: {
        path: "/usr",
      },
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "S3100",
      scope: "project",
      clusterId: "9dca9e71-a068-4232-993c-be8875c9c96a",
      projectId: 4500001,
      id: "17264c0b-fcb1-4f9c-99dc-56303e19da73",
      kind: "git",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2024-12-01T08:39:48.827931Z",
      updatedBy: "test@run.ai",
      updatedAt: "2024-12-01T08:39:48.827931Z",
      projectName: "p100",
    },
    spec: {
      s3: {
        path: "/usr",
      },
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "config-map100",
      scope: "project",
      clusterId: "9dca9e71-a068-4232-993c-be8875c9c96a",
      projectId: 4500001,
      id: "17264c0b-fcb1-4f9c-99dc-56303e19da74",
      kind: "git",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2024-12-01T08:39:48.827931Z",
      updatedBy: "test@run.ai",
      updatedAt: "2024-12-01T08:39:48.827931Z",
      projectName: "p100",
    },
    spec: {
      config_map: {},
    },
    compliance: {
      compliance: true,
    },
  },
  {
    meta: {
      name: "secret100",
      scope: "project",
      clusterId: "9dca9e71-a068-4232-993c-be8875c9c96a",
      projectId: 4500001,
      id: "17264c0b-fcb1-4f9c-99dc-56303e19da75",
      kind: "git",
      tenantId: 1001,
      createdBy: "test@run.ai",
      createdAt: "2024-12-01T08:39:48.827931Z",
      updatedBy: "test@run.ai",
      updatedAt: "2024-12-01T08:39:48.827931Z",
      projectName: "p100",
    },
    spec: {
      secret: {},
    },
    compliance: {
      compliance: true,
    },
  },
];
