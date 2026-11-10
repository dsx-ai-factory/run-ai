/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { BadRequestError, ConflictError, InternalServerError } from "@/models/http-response.model";
import { flushPromises } from "@vue/test-utils";
import { effectScope, nextTick } from "vue";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { useNodePoolForm } from "./use-node-pool-form.composable";
import { PROPORTIONAL_CAPABILITIES } from "@/models/node-pool.model";
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNodepools: vi.fn(),
    createNodepool: vi.fn(),
  },
}));
beforeEach(() => {
  vi.spyOn(console, "error").mockImplementation(() => {});
});
const scopes: ReturnType<typeof effectScope>[] = [];
function setup() {
  vi.mocked(clusterService.getNodepools).mockResolvedValue([]);
  const scope = effectScope();
  scopes.push(scope);
  const state = scope.run(() => useNodePoolForm(""));
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
  it("refreshes names after the backend rejects a duplicate as a server error and clears the conflict for another cluster", async () => {
    const state = setup();
    state.draft.clusterId = "cluster-a";
    state.draft.name = "taken";
    await flushPromises();
    vi.mocked(clusterService.createNodepool).mockRejectedValueOnce(new InternalServerError("Failed to create nodepool"));
    vi.mocked(clusterService.getNodepools).mockResolvedValueOnce([{ name: "taken" }] as never);
    await state.create();
    expect(state.nameConflict.value).toBe(true);
    state.draft.clusterId = "cluster-b";
    await flushPromises();
    expect(state.nameConflict.value).toBe(false);
    expect(state.draft.name).toBe("taken");
  });
  it("marks a POST name conflict and clears it when the name changes", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    state.draft.name = "taken";
    await flushPromises();
    vi.mocked(clusterService.createNodepool).mockRejectedValueOnce(new ConflictError("Node pool name already exists"));
    await state.create();
    expect(state.nameConflict.value).toBe(true);
    state.draft.name = "new-name";
    await flushPromises();
    expect(state.nameConflict.value).toBe(false);
  });
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
    await flushPromises();
    resolveOld([]);
    await flushPromises();
    expect(state.draft.labelKey).toBe("");
    expect(state.loading.value).toBe(false);
  });
  it("blocks creation while cluster data failed to load", async () => {
    const state = setup();
    vi.mocked(clusterService.getNodepools).mockRejectedValueOnce(new Error("offline"));
    state.draft.clusterId = "cluster";
    await flushPromises();
    await state.create();
    expect(state.loadError.value).toBeTruthy();
    expect(clusterService.createNodepool).not.toHaveBeenCalled();
  });
  it("sends description and explicit Proportional capabilities when creating a pool", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await flushPromises();
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
  it("uses one POST without name preflight, quota resets or a follow-up PATCH", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    await flushPromises();
    vi.mocked(clusterService.createNodepool).mockResolvedValue({
      id: "123",
      schedulingCapabilities: PROPORTIONAL_CAPABILITIES,
    } as never);
    expect((await state.create())?.id).toBe("123");
    expect(clusterService.createNodepool).toHaveBeenCalledTimes(1);
  });
  it("blocks concurrent submissions and retains input after a failed POST", async () => {
    const state = setup();
    state.draft.clusterId = "cluster";
    state.draft.name = "pool";
    await flushPromises();
    let rejectCreate!: (reason: Error) => void;
    vi.mocked(clusterService.createNodepool).mockImplementationOnce(
      () =>
        new Promise((_, reject) => {
          rejectCreate = reject;
        }),
    );
    const pending = state.create();
    expect(await state.create()).toBeUndefined();
    rejectCreate(new Error("offline"));
    await pending;
    expect(state.created.value).toBeUndefined();
    expect(state.draft.name).toBe("pool");
    expect(state.submitError.value).toBeTruthy();
    expect(clusterService.createNodepool).toHaveBeenCalledTimes(1);
  });
});
