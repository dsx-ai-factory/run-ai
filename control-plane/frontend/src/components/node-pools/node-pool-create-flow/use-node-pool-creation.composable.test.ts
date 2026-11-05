/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { BadRequestError, ForbiddenError } from "@/models/http-response.model";
import { flushPromises } from "@vue/test-utils";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { effectScope, nextTick } from "vue";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { useNodePoolCreation } from "./use-node-pool-creation.composable";
import { PROPORTIONAL_CAPABILITIES } from "./node-pool-create-flow.model";
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNodepools: vi.fn(),
    createNodepool: vi.fn(),
    patchNodepool: vi.fn(),
    isNodePoolNameUnique: vi.fn(),
  },
}));
vi.mock("@/services/control-plane/org-unit.service/org-unit.service", () => ({
  orgUnitService: { resetNodePoolQuotas: vi.fn() },
}));
beforeEach(() => {
  vi.spyOn(console, "error").mockImplementation(() => {});
});
const scopes: ReturnType<typeof effectScope>[] = [];
function setup() {
  vi.mocked(orgUnitService.resetNodePoolQuotas).mockResolvedValue(undefined);
  vi.mocked(clusterService.getNodepools).mockResolvedValue([]);
  vi.mocked(clusterService.isNodePoolNameUnique).mockResolvedValue(true);
  const scope = effectScope();
  scopes.push(scope);
  const state = scope.run(() => useNodePoolCreation(""));
  if (!state) throw new Error("Failed to create test scope");
  return state;
}
afterEach(() => {
  scopes.forEach((s) => s.stop());
  scopes.length = 0;
  vi.resetAllMocks();
  vi.restoreAllMocks();
});
describe("node pool creation", () => {
  it("preserves the backend explanation when creation is rejected", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await flushPromises();
    vi.mocked(clusterService.createNodepool).mockRejectedValueOnce(
      new BadRequestError("This node pool name is reserved"),
    );
    await state.create();
    expect(state.submitError.value).toContain("This node pool name is reserved");
    expect(state.created.value).toBeUndefined();
  });
  it("preserves partial creation context and the reason setup failed", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await flushPromises();
    vi.mocked(clusterService.createNodepool).mockResolvedValue({
      id: "123",
      schedulingCapabilities: { quota: true },
    } as never);
    vi.mocked(clusterService.patchNodepool).mockRejectedValueOnce(
      new ForbiddenError("Editing scheduling capabilities is disabled"),
    );
    await state.create();
    expect(state.submitError.value).toContain("was created");
    expect(state.submitError.value).toContain("Retry setup");
    expect(state.submitError.value).toContain("Editing scheduling capabilities is disabled");
    expect(state.created.value?.id).toBe("123");
  });
  it("ignores stale cluster responses and resets the selected label", async () => {
    const state = setup();
    let resolveOld!: (value: never[]) => void;
    vi.mocked(clusterService.getNodepools).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveOld = resolve;
        }),
    );
    state.draft.clusterId = "old";
    state.draft.labelKey = "old-key";
    await nextTick();
    state.draft.clusterId = "new";
    await nextTick();
    await Promise.resolve();
    resolveOld([]);
    await Promise.resolve();
    expect(state.draft.labelKey).toBe("");
    expect(state.loading.value).toBe(false);
  });
  it("blocks creation while cluster data failed to load", async () => {
    const state = setup();
    vi.mocked(clusterService.getNodepools).mockRejectedValueOnce(new Error("offline"));
    state.draft.clusterId = "cluster";
    await nextTick();
    await Promise.resolve();
    await state.create();
    expect(state.loadError.value).toBeTruthy();
    expect(clusterService.createNodepool).not.toHaveBeenCalled();
  });
  it("sends description and explicit Proportional capabilities when creating a pool", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await nextTick();
    await Promise.resolve();
    Object.assign(state.draft, { name: "pool", description: "My pool", labelKey: "type", labelValue: "gpu" });
    vi.mocked(clusterService.createNodepool).mockResolvedValue({
      id: "123",
      schedulingCapabilities: PROPORTIONAL_CAPABILITIES,
    } as never);
    await state.create();
    expect(clusterService.createNodepool).toHaveBeenCalledWith(
      expect.objectContaining({ description: "My pool", schedulingCapabilities: PROPORTIONAL_CAPABILITIES }),
    );
    expect(state.draft.description).toBe("My pool");
    expect(state.created.value?.id).toBe("123");
    await state.create();
    expect(clusterService.createNodepool).toHaveBeenCalledTimes(1);
  });
  it("does not create after disposal while name validation is pending", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await nextTick();
    await Promise.resolve();
    let resolveName: (value: boolean) => void = () => {};
    vi.mocked(clusterService.isNodePoolNameUnique).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveName = resolve;
        }),
    );
    const creation = state.create();
    scopes[0].stop();
    resolveName(true);
    await creation;
    expect(clusterService.createNodepool).not.toHaveBeenCalled();
  });
  it("retains input on failure and distinguishes a backend mode mismatch from a failed creation", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await nextTick();
    await Promise.resolve();
    state.draft.name = "pool";
    vi.mocked(clusterService.createNodepool).mockRejectedValueOnce(new Error("offline"));
    await state.create();
    expect(state.submitError.value).toBeTruthy();
    expect(state.draft.name).toBe("pool");
    vi.mocked(clusterService.createNodepool).mockResolvedValueOnce({
      id: "123",
      schedulingCapabilities: { quota: true },
    } as never);
    vi.mocked(clusterService.patchNodepool).mockRejectedValueOnce(new Error("offline"));
    await state.create();
    expect(state.created.value?.id).toBe("123");
    expect(state.modeMismatch.value).toBe(true);
    expect(state.submitError.value).toContain("Proportional");
  });
  it("waits for Proportional setup before reporting creation complete", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await nextTick();
    await Promise.resolve();
    const pool = {
      id: "123",
      schedulingCapabilities: { quota: true, ranks: true, adjustableWeights: false, cpuQuota: false },
    };
    vi.mocked(clusterService.createNodepool).mockResolvedValue(pool as never);
    let resolvePatch!: (value: never) => void;
    vi.mocked(clusterService.patchNodepool).mockImplementation(
      () =>
        new Promise((resolve) => {
          resolvePatch = resolve;
        }),
    );
    const creation = state.create();
    await flushPromises();
    expect(state.created.value?.id).toBe("123");
    expect(state.submitting.value).toBe(true);
    expect(clusterService.patchNodepool).toHaveBeenCalledWith("123", {
      schedulingCapabilities: PROPORTIONAL_CAPABILITIES,
    });
    expect(await state.create()).toBeUndefined();
    resolvePatch({ ...pool, schedulingCapabilities: PROPORTIONAL_CAPABILITIES } as never);
    expect((await creation)?.schedulingCapabilities).toEqual(PROPORTIONAL_CAPABILITIES);
    expect(state.modeMismatch.value).toBe(false);
  });
  it("retries only PATCH after setup failure and keeps the pool identity and draft", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await nextTick();
    await Promise.resolve();
    state.draft.description = "Preserve me";
    vi.mocked(clusterService.createNodepool).mockResolvedValue({
      id: "123",
      schedulingCapabilities: { quota: true },
    } as never);
    vi.mocked(clusterService.patchNodepool).mockRejectedValueOnce(new Error("offline"));
    expect(await state.create()).toBeUndefined();
    expect(state.created.value?.id).toBe("123");
    expect(state.submitError.value).toContain("Proportional");
    vi.mocked(clusterService.patchNodepool).mockResolvedValue({
      id: "123",
      schedulingCapabilities: PROPORTIONAL_CAPABILITIES,
    } as never);
    expect((await state.create())?.id).toBe("123");
    expect(state.draft.description).toBe("Preserve me");
    expect(clusterService.createNodepool).toHaveBeenCalledTimes(1);
    expect(clusterService.isNodePoolNameUnique).toHaveBeenCalledTimes(1);
    expect(clusterService.patchNodepool).toHaveBeenCalledTimes(2);
    expect(orgUnitService.resetNodePoolQuotas).toHaveBeenCalledExactlyOnceWith("cluster", "123");
    expect(state.submitError.value).toBe("");
  });
  it("does not report success when PATCH returns unchanged capabilities", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await nextTick();
    await Promise.resolve();
    const pool = { id: "123", schedulingCapabilities: { quota: true } };
    vi.mocked(clusterService.createNodepool).mockResolvedValue(pool as never);
    vi.mocked(clusterService.patchNodepool).mockResolvedValue(pool as never);
    expect(await state.create()).toBeUndefined();
    expect(state.modeMismatch.value).toBe(true);
    expect(state.submitError.value).toContain("Proportional");
  });
  it("retains the created pool and stops before PATCH if its quota preparation fails", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await flushPromises();
    vi.mocked(clusterService.createNodepool).mockResolvedValue({
      id: "123",
      schedulingCapabilities: { quota: true },
    } as never);
    vi.mocked(orgUnitService.resetNodePoolQuotas).mockRejectedValueOnce(new Error("offline"));
    expect(await state.create()).toBeUndefined();
    expect(state.created.value?.id).toBe("123");
    expect(clusterService.patchNodepool).not.toHaveBeenCalled();
    expect(state.submitError.value).toContain("Proportional");
    vi.mocked(clusterService.patchNodepool).mockResolvedValue({
      id: "123",
      schedulingCapabilities: PROPORTIONAL_CAPABILITIES,
    } as never);
    expect((await state.create())?.id).toBe("123");
    expect(clusterService.createNodepool).toHaveBeenCalledTimes(1);
  });
});
