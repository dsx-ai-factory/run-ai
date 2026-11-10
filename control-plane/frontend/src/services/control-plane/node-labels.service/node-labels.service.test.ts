/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it, vi } from "vitest";
import { httpService } from "@/services/infra/https.service/http.service";
import { nodeLabelsService } from "./node-labels.service";
vi.mock("@/services/infra/https.service/http.service", () => ({
  httpService: { request: vi.fn().mockResolvedValue([]) },
}));
describe("node label API contract", () => {
  it("uses the keys and values paths with an exact key query parameter", async () => {
    const query = { contains: "H%_", offset: 50, limit: 50 };
    await nodeLabelsService.keys("cluster-id", query);
    expect(httpService.request).toHaveBeenLastCalledWith(
      "",
      "/api/v1/clusters/cluster-id/nodes/labels/keys",
      "GET",
      { responseType: "json" },
      {},
      query,
    );
    await nodeLabelsService.values("cluster-id", "nvidia.com/gpu.product", query);
    expect(httpService.request).toHaveBeenLastCalledWith(
      "",
      "/api/v1/clusters/cluster-id/nodes/labels/values",
      "GET",
      { responseType: "json" },
      {},
      { ...query, key: "nvidia.com/gpu.product" },
    );
  });
});
