/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { beforeEach, describe, expect, it, vi } from "vitest";
import { nodeLabelsService } from "./node-labels.service";
const captured = vi.hoisted(() => ({ urls: [] as string[] }));
vi.mock("@/services/infra/client-apis/cluster-service-api/cluster-service-api", async () => {
  const { NodesApi, Configuration } = await import("@/swagger-models/cluster-service-client");
  const { default: axios } = await import("axios");
  const transport = axios.create({
    adapter: async (config) => {
      captured.urls.push(config.url || "");
      return { data: ["gpu-type"], status: 200, statusText: "OK", headers: {}, config };
    },
  });
  return { clusterServiceApi: { nodesApi: new NodesApi(new Configuration(), "https://example.test", transport) } };
});
beforeEach(() => {
  captured.urls.length = 0;
});
describe("merged node label API contract", () => {
  it("uses the generated client's search parameter for keys and exact-key values", async () => {
    const query = { search: "H%_", offset: 50, limit: 50 };
    expect(await nodeLabelsService.keys("cluster-id", query)).toEqual(["gpu-type"]);
    expect(await nodeLabelsService.values("cluster-id", "nvidia.com/gpu.product", query)).toEqual(["gpu-type"]);
    expect(captured.urls).toHaveLength(2);
    const [keys, values] = captured.urls.map((url) => new URL(url));
    expect(keys.pathname).toBe("/api/v1/clusters/cluster-id/nodes/labels/keys");
    expect(values.pathname).toBe("/api/v1/clusters/cluster-id/nodes/labels/values");
    for (const url of [keys, values]) {
      expect(url.searchParams.get("search")).toBe("H%_");
      expect(url.searchParams.has("contains")).toBe(false);
      expect(url.searchParams.get("offset")).toBe("50");
      expect(url.searchParams.get("limit")).toBe("50");
    }
    expect(values.searchParams.get("key")).toBe("nvidia.com/gpu.product");
  });
});
