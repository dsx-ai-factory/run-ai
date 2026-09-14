/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//@ts-nocheck
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { IFilterBy } from "@/models/filter.model";
import { environmentColumns } from "@/models/environment.model";

import { filterService } from "@/services/filter.service/filter.service";
import { environmentService } from "@/services/control-plane/environment.service/environment.service";
import { assetsServiceApi } from "@/services/infra/client-apis/assets-service-api/assets-service-api";

vi.mock("@/services/filter.service/filter.service");
vi.mock("@/services/infra/client-apis/assets-service-api/assets-service-api");

const environments = {
  entries: [
    {
      meta: {
        name: "pycharm-3",
        scope: "tenant",
        id: "f67c8072-208c-4d6a-acb4-9058d27379f8",
        kind: "environment",
        tenantId: 1,
        createdBy: "0486471c-7583-4119-a1b4-d779a10f3093",
        createdAt: "2022-10-19T11:39:27.751556Z",
        updatedBy: "0486471c-7583-4119-a1b4-d779a10f3093",
        updatedAt: "2022-10-19T11:39:27.751556Z",
      },
      spec: {
        args: "ENV=puki",
        environmentVariables: [
          {
            name: "token",
            value: "123344",
          },
        ],
        image: "ubuntu",
        connections: [
          {
            name: "pycharmExternal",
            isExternal: false,
            internalToolInfo: {
              toolType: "pycharm",
              connectionType: "ExternalUrl",
              containerPort: 22,
            },
          },
        ],
      },
      usedBy: {
        workspaces: [
          {
            id: "342",
            name: "ws",
          },
        ],
      },
    },
    {
      meta: {
        name: "run-as-user",
        scope: "tenant",
        id: "2e3e97d1-0618-4cbf-ac44-980cbfd8a9a0",
        kind: "environment",
        tenantId: 1,
        createdBy: "0486471c-7583-4119-a1b4-d779a10f3093",
        createdAt: "2022-10-31T11:05:45.180268Z",
        updatedBy: "0486471c-7583-4119-a1b4-d779a10f3093",
        updatedAt: "2022-10-31T11:05:45.180268Z",
      },
      spec: {
        environmentVariables: [],
        image: "sdada",
        imagePullPolicy: "IfNotPresent",
        runAsUser: true,
        capabilities: [],
      },
    },
    {
      meta: {
        name: "run-as-user-2",
        scope: "tenant",
        id: "73a1b78f-6b69-46ca-99e1-8598e4484430",
        kind: "environment",
        tenantId: 1,
        createdBy: "0486471c-7583-4119-a1b4-d779a10f3093",
        createdAt: "2022-10-31T12:29:08.400132Z",
        updatedBy: "0486471c-7583-4119-a1b4-d779a10f3093",
        updatedAt: "2022-10-31T12:29:08.400132Z",
      },
      spec: {
        environmentVariables: [],
        image: "imagos",
        imagePullPolicy: "IfNotPresent",
        connections: [
          {
            name: "pycharmExternal",
            isExternal: false,
            internalToolInfo: {
              toolType: "pycharm",
              connectionType: "ExternalUrl",
              containerPort: 22,
            },
          },
          {
            name: "Tool 1",
            isExternal: false,
            internalToolInfo: {
              toolType: "jupyter-notebook",
              connectionType: "ExternalUrl",
              containerPort: 8888,
            },
          },
        ],
        runAsUser: true,
        capabilities: [],
      },
    },
    {
      meta: {
        name: "tomenvos",
        scope: "tenant",
        id: "089ef9f6-9284-4c2e-bdc1-40407268e696",
        kind: "environment",
        tenantId: 1,
        createdBy: "0486471c-7583-4119-a1b4-d779a10f3093",
        createdAt: "2022-11-03T09:49:13.320033Z",
        updatedBy: "0486471c-7583-4119-a1b4-d779a10f3093",
        updatedAt: "2022-11-03T09:49:13.320033Z",
      },
      spec: {
        environmentVariables: [],
        image: "ubuntu",
        imagePullPolicy: "IfNotPresent",
        connections: [
          {
            name: "Tool 1",
            isExternal: false,
            internalToolInfo: {
              toolType: "jupyter-notebook",
              connectionType: "ExternalUrl",
              containerPort: 8888,
            },
          },
          {
            name: "Tool 2",
            isExternal: false,
            internalToolInfo: {
              toolType: "tensorboard",
              connectionType: "ExternalUrl",
              containerPort: 6006,
            },
          },
        ],
        runAsUser: false,
        capabilities: [],
      },
    },
  ],
};

const filterBy: IFilterBy = {
  searchTerm: "",
  columnFilters: [],
};

describe("environment service", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  describe("environmentService.list", () => {
    it("should return all the environments without filtering", async () => {
      assetsServiceApi.environmentApi.listEnvironmentAssets.mockResolvedValue({ data: environments });
      const res = await environmentService.list(filterBy, environmentColumns);
      expect(res.length).toBe(environments.entries.length);
    });
    it("should not call the filter service function when filterBy is empty", async () => {
      assetsServiceApi.environmentApi.listEnvironmentAssets.mockResolvedValue({ data: environments });
      await environmentService.list(filterBy, environmentColumns);
      expect(filterService.filterByColumns).toBeCalledTimes(0);
      expect(filterService.filterBySearchTerm).toBeCalledTimes(0);
    });
  });
});
