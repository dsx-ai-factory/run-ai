/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { effectScope, ref, type EffectScope } from "vue";
import { flushPromises } from "@vue/test-utils";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import type { INodePoolSummaryResources } from "@/models/node-pool.model";
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: { getNodepool: vi.fn() },
}));
vi.mock("@/composables/node-pool/node-pool-scheduling-data.util", () => ({ loadSchedulingResources: vi.fn() }));
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { loadSchedulingResources } from "@/composables/node-pool/node-pool-scheduling-data.util";
import { useNodePoolWizardData } from "./use-node-pool-wizard-data.composable";

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: unknown) => void;
  const promise = new Promise<T>((done, fail) => {
    resolve = done;
    reject = fail;
  });
  return { promise, resolve, reject };
}
const pool = (id: string) => ({ id, clusterId: `cluster-${id}`, name: `Pool ${id}` }) as Nodepool;
let scopes: EffectScope[] = [];
function setup(id = "a") {
  const poolId = ref(id);
  const scope = effectScope();
  scopes.push(scope);
  const data = scope.run(() => useNodePoolWizardData(poolId));
  if (!data) throw new Error("Failed to create topology data scope");
  return { poolId, scope, data };
}
beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(loadSchedulingResources).mockResolvedValue({});
});
afterEach(() => {
  scopes.forEach((scope) => scope.stop());
  scopes = [];
});

describe("wizard pool loading", () => {
  it("loads a direct pool ID with its own capacity without department data", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue(pool("a"));
    vi.mocked(loadSchedulingResources).mockResolvedValue({ gpus: 8, nodes: 2 });
    const { data } = setup();
    expect(data.loading.value).toBe(true);
    await flushPromises();
    expect(data.pool.value?.id).toBe("a");
    expect(data.resources.value).toEqual({ gpus: 8, nodes: 2 });
    expect(data.loading.value).toBe(false);
    expect(data.loadError.value).toBe("");
    expect(clusterService.getNodepool).toHaveBeenCalledWith("a");
    expect(loadSchedulingResources).toHaveBeenCalledWith(pool("a"));
  });

  it("keeps known capacity during a refresh and clears it if the new reading fails", async () => {
    const pending = deferred<INodePoolSummaryResources>();
    vi.mocked(clusterService.getNodepool).mockResolvedValue(pool("a"));
    vi.mocked(loadSchedulingResources).mockResolvedValueOnce({ gpus: 8 }).mockReturnValueOnce(pending.promise);
    const { data } = setup();
    await flushPromises();
    const refresh = data.refresh();
    await flushPromises();
    expect(data.resources.value?.gpus).toBe(8);
    pending.reject(new Error("Metrics unavailable"));
    await refresh;
    expect(data.resources.value?.gpus).toBeUndefined();
    expect(data.telemetryError.value).toContain("Metrics unavailable");
  });

  it("clears the prior pool and capacity synchronously when the ID changes", async () => {
    const next = deferred<Nodepool>();
    vi.mocked(clusterService.getNodepool).mockResolvedValueOnce(pool("a")).mockReturnValueOnce(next.promise);
    vi.mocked(loadSchedulingResources).mockResolvedValue({ gpus: 8 });
    const { poolId, data } = setup();
    await flushPromises();
    poolId.value = "b";
    expect(data.pool.value).toBeUndefined();
    expect(data.resources.value).toBeUndefined();
    expect(data.loading.value).toBe(true);
    next.resolve(pool("b"));
    await flushPromises();
    expect(data.pool.value?.id).toBe("b");
  });

  it("ignores late pool responses after a different route finishes loading", async () => {
    const first = deferred<Nodepool>();
    vi.mocked(clusterService.getNodepool).mockReturnValueOnce(first.promise).mockResolvedValueOnce(pool("b"));
    const { poolId, data } = setup();
    poolId.value = "b";
    await flushPromises();
    first.resolve(pool("a"));
    await flushPromises();
    expect(data.pool.value?.id).toBe("b");
    expect(loadSchedulingResources).toHaveBeenCalledTimes(1);
    expect(loadSchedulingResources).toHaveBeenCalledWith(pool("b"));
  });

  it.each(["success", "failure"])("ignores late capacity %s from the previous route", async (result) => {
    const oldCapacity = deferred<INodePoolSummaryResources>();
    vi.mocked(clusterService.getNodepool).mockResolvedValueOnce(pool("a")).mockResolvedValueOnce(pool("b"));
    vi.mocked(loadSchedulingResources).mockReturnValueOnce(oldCapacity.promise).mockResolvedValueOnce({ gpus: 4 });
    const { poolId, data } = setup();
    await flushPromises();
    poolId.value = "b";
    await flushPromises();
    if (result === "success") oldCapacity.resolve({ gpus: 99 });
    else oldCapacity.reject(new Error("Stale capacity failure"));
    await flushPromises();
    expect(data.pool.value?.id).toBe("b");
    expect(data.resources.value).toEqual({ gpus: 4 });
    expect(data.telemetryError.value).toBe("");
    expect(data.loading.value).toBe(false);
  });

  it("separates capacity failure from pool failure and keeps unknown capacity unknown", async () => {
    vi.mocked(clusterService.getNodepool).mockResolvedValue(pool("a"));
    vi.mocked(loadSchedulingResources).mockRejectedValue(new Error("Metrics unavailable"));
    const { data } = setup();
    await flushPromises();
    expect(data.pool.value?.id).toBe("a");
    expect(data.resources.value).toBeUndefined();
    expect(data.loadError.value).toBe("");
    expect(data.telemetryError.value).toMatch(/Metrics unavailable/);
    expect(data.loading.value).toBe(false);
  });

  it("shows pool errors and retries without retaining the failure", async () => {
    vi.mocked(clusterService.getNodepool)
      .mockRejectedValueOnce(new Error("Pool unavailable"))
      .mockResolvedValueOnce(pool("a"));
    const { data } = setup();
    await flushPromises();
    expect(data.pool.value).toBeUndefined();
    expect(data.loadError.value).toMatch(/Pool unavailable/);
    expect(data.telemetryError.value).toBe("");
    expect(data.loading.value).toBe(false);
    await data.refresh();
    expect(data.pool.value?.id).toBe("a");
    expect(data.loadError.value).toBe("");
  });

  it("does not request an empty pool ID or update state after disposal", async () => {
    const pending = deferred<Nodepool>();
    vi.mocked(clusterService.getNodepool).mockReturnValue(pending.promise);
    const { poolId, scope, data } = setup("");
    expect(clusterService.getNodepool).not.toHaveBeenCalled();
    expect(data.loading.value).toBe(false);
    poolId.value = "a";
    scope.stop();
    pending.resolve(pool("a"));
    await flushPromises();
    expect(data.pool.value).toBeUndefined();
    expect(data.resources.value).toBeUndefined();
    expect(loadSchedulingResources).not.toHaveBeenCalled();
    await data.refresh();
    expect(clusterService.getNodepool).toHaveBeenCalledTimes(1);
  });
});
