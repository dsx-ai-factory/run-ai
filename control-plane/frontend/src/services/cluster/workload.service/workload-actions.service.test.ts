/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, describe, expect, it, vi } from "vitest";
import { workloadService } from "./workload.service";
import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import { WorkloadActionsApiAxiosParamCreator } from "@/swagger-models/workloads-client";

afterEach(() => vi.restoreAllMocks());

describe("workload action history", () => {
  it("uses the generated authenticated API with a default page of 50", async () => {
    const data = { actions: [], next: 50 };
    const get = vi
      .spyOn(workloadServiceApi.workloadActionsApi, "getWorkloadActionsHistory")
      .mockResolvedValue({ data } as never);
    expect(await workloadService.getWorkloadActions("deleted-workload")).toEqual(data);
    expect(get).toHaveBeenCalledWith("deleted-workload", 0, 50);
  });

  it("forwards the requested page without fetching subsequent pages", async () => {
    const get = vi
      .spyOn(workloadServiceApi.workloadActionsApi, "getWorkloadActionsHistory")
      .mockResolvedValue({ data: { actions: [], next: 100 } } as never);
    await workloadService.getWorkloadActions("workload", 50, 50);
    expect(get).toHaveBeenCalledExactlyOnceWith("workload", 50, 50);
  });

  it("encodes the workload ID and serializes offset and limit in the generated request", async () => {
    const request = await WorkloadActionsApiAxiosParamCreator().getWorkloadActionsHistory("workload/id", 50, 50);
    expect(request.url).toBe("/api/v1/workloads/workload%2Fid/actions?offset=50&limit=50");
  });
  it("serializes sorting and combined filters into a single server request", async () => {
    const request = await WorkloadActionsApiAxiosParamCreator().getWorkloadActionsHistory(
      "workload",
      0,
      50,
      "asc",
      "subject",
      ["action!=update", "subject=@test@run.ai"],
      "Created & user",
    );
    const url = new URL(request.url, "https://example.invalid");
    expect(url.searchParams.get("search")).toBe("Created & user");
    expect(url.searchParams.get("sortBy")).toBe("subject");
    expect(url.searchParams.get("sortOrder")).toBe("asc");
    expect(url.searchParams.get("filterBy")).toBe("action!=update,subject=@test@run.ai");
  });
});
