/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { beforeEach, describe, expect, it, vi } from "vitest";
vi.mock("@/services/infra/client-apis/cluster-service-api/cluster-service-api", () => ({
  clusterServiceApi: { nodePoolsApi: { patchNodepool: vi.fn(), updateNodepool: vi.fn() } },
}));
vi.mock("@/services/infra/https.service/http.service", () => ({ httpService: { handleHttpError: vi.fn() } }));
import { clusterServiceApi } from "@/services/infra/client-apis/cluster-service-api/cluster-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import { saveTopologyPool } from "./node-pool-topology.service";

describe("topology persistence", () => {
  beforeEach(() => vi.clearAllMocks());
  it("uses a narrow PATCH instead of a full update that resets other steps", async () => {
    const payload = { networkTopologyId: "topology", schedulingConfiguration: { numaAwareScheduling: true } };
    const saved = { id: "p", ...payload };
    vi.mocked(clusterServiceApi.nodePoolsApi.patchNodepool).mockResolvedValue({ data: saved } as never);
    expect(await saveTopologyPool("p", payload)).toEqual(saved);
    expect(clusterServiceApi.nodePoolsApi.patchNodepool).toHaveBeenCalledWith("p", payload);
    expect(clusterServiceApi.nodePoolsApi.updateNodepool).not.toHaveBeenCalled();
  });

  it("normalizes backend errors for the form's save failure handling", async () => {
    const failure = { response: { status: 403 } };
    const normalized = new Error("Permission denied");
    vi.mocked(clusterServiceApi.nodePoolsApi.patchNodepool).mockRejectedValue(failure);
    vi.mocked(httpService.handleHttpError).mockReturnValue(normalized as never);
    await expect(saveTopologyPool("p", { networkTopologyId: null })).rejects.toBe(normalized);
    expect(httpService.handleHttpError).toHaveBeenCalledWith(failure);
  });
});
