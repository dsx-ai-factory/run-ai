/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";

const resetProjectsMock = vi.fn();
const countProjectsMock = vi.fn();
const resetDepartmentsMock = vi.fn();
const batchDeleteReportsMock = vi.fn();

vi.mock("@/services/infra/client-apis/org-unit-service-api/org-unit-service-api", () => ({
  orgUnitServiceApi: {
    projectsApi: {
      countProjects: (...args: unknown[]) => countProjectsMock(...args),
      projectsResetResources: (...args: unknown[]) => resetProjectsMock(...args),
    },
    departmentsApi: { departmentsResetResources: (...args: unknown[]) => resetDepartmentsMock(...args) },
    reportsApi: {
      batchDeleteReports: (...args: unknown[]) => batchDeleteReportsMock(...args),
    },
  },
}));

import { orgUnitService } from "./org-unit.service";

describe("orgUnitService.batchDeleteReports", () => {
  beforeEach(() => {
    batchDeleteReportsMock.mockReset();
    batchDeleteReportsMock.mockResolvedValue({ data: { id: "batch-1", accepted: [], failed: [] } });
  });

  function sentPayload(): { ids: unknown } {
    return batchDeleteReportsMock.mock.calls[0][0] as { ids: unknown };
  }

  // A Set serializes to {} rather than a JSON array, so the payload has to be a
  // real array or the backend receives no ids at all.
  it("sends ids as a JSON array, not a Set", async () => {
    await orgUnitService.batchDeleteReports(["id-1", "id-2"]);

    expect(JSON.parse(JSON.stringify(sentPayload()))).toEqual({ ids: ["id-1", "id-2"] });
  });

  it("drops duplicate ids", async () => {
    await orgUnitService.batchDeleteReports(["id-1", "id-2", "id-1"]);

    expect(JSON.parse(JSON.stringify(sentPayload()))).toEqual({ ids: ["id-1", "id-2"] });
  });

  it("returns the batch response body", async () => {
    batchDeleteReportsMock.mockResolvedValue({
      data: { id: "batch-9", accepted: ["id-1"], failed: [{ id: "id-2", code: 403, message: "nope" }] },
    });

    const response = await orgUnitService.batchDeleteReports(["id-1", "id-2"]);

    expect(response.accepted).toEqual(["id-1"]);
    expect(response.failed).toHaveLength(1);
  });
});

describe("orgUnitService.resetNodePoolQuotas", () => {
  beforeEach(() => {
    countProjectsMock.mockReset();
    countProjectsMock.mockResolvedValue({ data: { count: 1 } });
    resetProjectsMock.mockReset();
    resetDepartmentsMock.mockReset();
  });
  it("resets projects and departments only on the specified pool", async () => {
    await orgUnitService.resetNodePoolQuotas("cluster-a", "new-pool");
    const expected = { clusterId: "cluster-a", nodePoolIds: ["new-pool"], action: "QUOTAS_TO_DEFAULT" };
    expect(resetProjectsMock).toHaveBeenCalledExactlyOnceWith(expected);
    expect(resetDepartmentsMock).toHaveBeenCalledExactlyOnceWith(expected);
  });
  it("refuses an empty pool ID without sending a reset", async () => {
    await expect(orgUnitService.resetNodePoolQuotas("cluster-a", "")).rejects.toThrow();
    expect(resetProjectsMock).not.toHaveBeenCalled();
    expect(resetDepartmentsMock).not.toHaveBeenCalled();
  });
  it("skips project reset in an empty cluster while preparing the department quota", async () => {
    countProjectsMock.mockResolvedValue({ data: { count: 0 } });
    await orgUnitService.resetNodePoolQuotas("cluster-a", "new-pool");
    expect(countProjectsMock).toHaveBeenCalledExactlyOnceWith(["clusterId==cluster-a"]);
    expect(resetProjectsMock).not.toHaveBeenCalled();
    expect(resetDepartmentsMock).toHaveBeenCalledExactlyOnceWith({
      clusterId: "cluster-a",
      nodePoolIds: ["new-pool"],
      action: "QUOTAS_TO_DEFAULT",
    });
  });
});
