/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { effectScope, ref } from "vue";
import { afterEach, describe, expect, it, vi } from "vitest";
import { nodeLabelsService } from "@/services/control-plane/node-labels.service/node-labels.service";
import { useNodeLabelOptions } from "./use-node-label-options.composable";
vi.mock("@/services/control-plane/node-labels.service/node-labels.service", () => ({
  nodeLabelsService: { keys: vi.fn(), values: vi.fn() },
}));
const scopes: ReturnType<typeof effectScope>[] = [];
afterEach(() => {
  scopes.splice(0).forEach((s) => s.stop());
  vi.resetAllMocks();
});
function setup() {
  const cluster = ref("cluster-a"),
    key = ref("nvidia.com/gpu.product");
  const scope = effectScope();
  scopes.push(scope);
  const state = scope.run(() => useNodeLabelOptions(cluster, key));
  if (!state) throw new Error("Failed to initialize label options");
  return { state, cluster, key, scope };
}
describe("node label autocomplete", () => {
  it("passes literal search and exact key to the API and paginates using raw result counts", async () => {
    const { state } = setup();
    vi.mocked(nodeLabelsService.values)
      .mockResolvedValueOnce(Array.from({ length: 50 }, (_, i) => `H${i}`))
      .mockResolvedValueOnce(["H100"]);
    await state.searchValues("H");
    await state.moreValues();
    expect(nodeLabelsService.values).toHaveBeenNthCalledWith(1, "cluster-a", "nvidia.com/gpu.product", {
      contains: "H",
      offset: 0,
      limit: 50,
    });
    expect(nodeLabelsService.values).toHaveBeenNthCalledWith(2, "cluster-a", "nvidia.com/gpu.product", {
      contains: "H",
      offset: 50,
      limit: 50,
    });
    expect(state.values.options).toHaveLength(51);
    expect(state.values.more).toBe(false);
  });
  it("ignores responses from previous searches and clusters", async () => {
    const { state, cluster } = setup();
    let resolve!: (v: string[]) => void;
    vi.mocked(nodeLabelsService.keys)
      .mockImplementationOnce(() => new Promise((r) => (resolve = r)))
      .mockResolvedValueOnce(["gpu"]);
    const old = state.searchKeys("old");
    await state.searchKeys("g");
    resolve(["old"]);
    await old;
    expect(state.keys.options).toEqual(["gpu"]);
    cluster.value = "cluster-b";
    expect(state.keys.options).toEqual([]);
  });
  it("clears values on key change and discards the old key's response", async () => {
    const { state, key } = setup();
    let resolve!: (v: string[]) => void;
    vi.mocked(nodeLabelsService.values).mockImplementationOnce(() => new Promise((r) => (resolve = r)));
    const old = state.searchValues("");
    key.value = "other";
    resolve(["H100"]);
    await old;
    expect(state.values.options).toEqual([]);
    expect(state.values.loading).toBe(false);
  });
  it("reports unavailable suggestions without affecting form entry and can retry", async () => {
    const { state } = setup();
    vi.mocked(nodeLabelsService.keys).mockRejectedValueOnce(new Error("404")).mockResolvedValueOnce([]);
    await state.searchKeys("");
    expect(state.keys.error).toContain("manually");
    expect(state.keys.loading).toBe(false);
    await state.searchKeys("");
    expect(state.keys.error).toBe("");
    expect(state.keys.options).toEqual([]);
  });
  it("does not request values without a key and ignores responses after disposal", async () => {
    const { state, key, scope } = setup();
    key.value = "";
    await state.searchValues("");
    expect(nodeLabelsService.values).not.toHaveBeenCalled();
    let resolve!: (v: string[]) => void;
    vi.mocked(nodeLabelsService.keys).mockImplementationOnce(() => new Promise((r) => (resolve = r)));
    const request = state.searchKeys("");
    scope.stop();
    resolve(["gpu"]);
    await request;
    expect(state.keys.options).toEqual([]);
  });
});
