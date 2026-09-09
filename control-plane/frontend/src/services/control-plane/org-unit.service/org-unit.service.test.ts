/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";

const batchDeleteReportsMock = vi.fn();

vi.mock("@/services/infra/client-apis/org-unit-service-api/org-unit-service-api", () => ({
  orgUnitServiceApi: {
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
