/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";

const batchDeleteReportsMock = vi.fn();
const simulateMock = vi.fn();
const departmentMock = vi.fn();

vi.mock("@/services/infra/client-apis/org-unit-service-api/org-unit-service-api", () => ({
  orgUnitServiceApi: {
    orgUnitApi: { simulateNodePoolFairShare: (...args: unknown[]) => simulateMock(...args) },
    departmentsApi: { getDepartment: (...args: unknown[]) => departmentMock(...args) },
    reportsApi: {
      batchDeleteReports: (...args: unknown[]) => batchDeleteReportsMock(...args),
    },
  },
}));

import { orgUnitService } from "./org-unit.service";
import { createPinia, setActivePinia } from "pinia";
import { CanceledError } from "axios";

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

describe("orgUnitService scheduling API boundaries", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    simulateMock.mockReset();
    departmentMock.mockReset();
  });
  it("preserves simulation validation status and server message", async () => {
    simulateMock.mockRejectedValue({ response: { status: 400, data: { message: "Limit below quota" } } });
    await expect(
      orgUnitService.simulateNodePoolFairShare("p", {
        nodePool: {
          capacity: { gpu: 64 },
          schedulingCapabilities: { quota: false, ranks: false, adjustableWeights: true, limits: false },
        },
        orgUnitType: "department",
        orgUnits: [{ id: "a" }],
      }),
    ).rejects.toMatchObject({ statusCode: 400, message: "Limit below quota" });
  });
  it("preserves cancellation instead of normalizing an intentional abort", async () => {
    const aborted = new CanceledError("canceled");
    simulateMock.mockRejectedValue(aborted);
    await expect(
      orgUnitService.simulateNodePoolFairShare("p", {
        nodePool: {
          capacity: { gpu: 64 },
          schedulingCapabilities: { quota: false, ranks: false, adjustableWeights: true, limits: false },
        },
        orgUnitType: "department",
        orgUnits: [{ id: "a" }],
      }),
    ).rejects.toBe(aborted);
  });
  it("resolves child quota by pool identity and keeps missing data distinct from an empty department", async () => {
    departmentMock.mockResolvedValue({
      data: {
        clusterId: "c",
        children: [],
        projectsAggregatedResources: [{ nodePool: { id: "p", name: "pool" }, gpu: { deserved: 7 } }],
      },
    });
    await expect(
      orgUnitService.getDepartmentSubordinateGpuQuota("a", { nodePoolId: "p", clusterId: "c" }),
    ).resolves.toBe(7);
    await expect(orgUnitService.getDepartmentSubordinateGpuQuota("a", { nodePoolName: "pool" })).resolves.toBe(7);
    departmentMock.mockResolvedValue({ data: { clusterId: "c", children: [] } });
    await expect(orgUnitService.getDepartmentSubordinateGpuQuota("a", { nodePoolId: "p" })).resolves.toBeUndefined();
    await expect(
      orgUnitService.getDepartmentSubordinateGpuQuota("a", { nodePoolId: "p", emptyChildrenAsZero: true }),
    ).resolves.toBe(0);
    departmentMock.mockResolvedValue({ data: { clusterId: "other", children: [] } });
    await expect(
      orgUnitService.getDepartmentSubordinateGpuQuota("a", { nodePoolId: "p", clusterId: "c" }),
    ).rejects.toThrow("Department does not belong to this cluster.");
  });
});
