/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { describe, expect, it } from "vitest";
import {
  NodepoolGPUNetworkAccelerationDetection as Detection,
  PlacementStrategy,
  type Nodepool,
} from "@/swagger-models/cluster-service-client";
import {
  buildTopologyPatch,
  topologyDraftFromPool,
  validateTopologyLabelKey,
  type TopologyDraft,
} from "./node-pool-topology.model";

const pool = { id: "p", clusterId: "c", name: "Pool" } as Nodepool;
const supported = { mnnvl: true, numa: true, gpuDevice: true, networkTopology: true };

describe("topology draft and patch", () => {
  it("hydrates omitted configuration with server defaults without assigning a custom label", () => {
    expect(topologyDraftFromPool(pool)).toEqual({
      networkTopologyId: null,
      gpuNetworkAccelerationDetection: Detection.Auto,
      gpuNetworkAccelerationLabelKey: "",
      numaAwareScheduling: false,
      gpuNodePlacement: PlacementStrategy.Binpack,
      gpuDevicePlacement: PlacementStrategy.Binpack,
      cpuNodePlacement: PlacementStrategy.Binpack,
    });
  });

  it("preserves saved custom Auto labels and independent placement choices", () => {
    const actual = topologyDraftFromPool({
      ...pool,
      networkTopologyId: "topology",
      gpuNetworkAccelerationDetection: Detection.Auto,
      gpuNetworkAccelerationLabelKey: "custom.example/clique",
      schedulingConfiguration: {
        numaAwareScheduling: true,
        gpuPodsPlacementStrategy: { node: PlacementStrategy.Spread, device: PlacementStrategy.Binpack },
        cpuPodsPlacementStrategy: { node: PlacementStrategy.Spread },
        minGuaranteedRuntime: "10m",
        timeBasedFairShare: { enabled: true, historicalUsageWeight: 2 },
      },
    });
    expect(actual).toEqual({
      networkTopologyId: "topology",
      gpuNetworkAccelerationDetection: Detection.Auto,
      gpuNetworkAccelerationLabelKey: "custom.example/clique",
      numaAwareScheduling: true,
      gpuNodePlacement: PlacementStrategy.Spread,
      gpuDevicePlacement: PlacementStrategy.Binpack,
      cpuNodePlacement: PlacementStrategy.Spread,
    });
    expect(buildTopologyPatch(actual, supported)).toEqual({
      networkTopologyId: "topology",
      gpuNetworkAccelerationDetection: Detection.Auto,
      gpuNetworkAccelerationLabelKey: "custom.example/clique",
      schedulingConfiguration: {
        numaAwareScheduling: true,
        gpuPodsPlacementStrategy: { node: PlacementStrategy.Spread, device: PlacementStrategy.Binpack },
        cpuPodsPlacementStrategy: { node: PlacementStrategy.Spread },
      },
    });
  });

  it.each([Detection.Auto, Detection.Use])("assigns the default label only when saving blank %s settings", (mode) => {
    const draft = {
      ...topologyDraftFromPool(pool),
      gpuNetworkAccelerationDetection: mode,
      gpuNetworkAccelerationLabelKey: "  ",
    };
    expect(buildTopologyPatch(draft, supported).gpuNetworkAccelerationLabelKey).toBe("nvidia.com/gpu.clique");
    expect(draft.gpuNetworkAccelerationLabelKey).toBe("  ");
  });

  it("omits the inactive DontUse label without discarding it from the draft", () => {
    const draft: TopologyDraft = {
      ...topologyDraftFromPool(pool),
      gpuNetworkAccelerationDetection: Detection.DontUse,
      gpuNetworkAccelerationLabelKey: "custom.example/retained",
    };
    expect(buildTopologyPatch(draft, supported)).not.toHaveProperty("gpuNetworkAccelerationLabelKey");
    draft.gpuNetworkAccelerationDetection = Detection.Use;
    expect(buildTopologyPatch(draft, supported).gpuNetworkAccelerationLabelKey).toBe("custom.example/retained");
  });

  it("omits unsupported MNNVL, device and NUMA fields while preserving a saved topology association", () => {
    const draft = { ...topologyDraftFromPool(pool), networkTopologyId: "existing", numaAwareScheduling: true };
    expect(buildTopologyPatch(draft, { mnnvl: false, numa: false, gpuDevice: false, networkTopology: false })).toEqual({
      networkTopologyId: "existing",
      schedulingConfiguration: {
        gpuPodsPlacementStrategy: { node: PlacementStrategy.Binpack },
        cpuPodsPlacementStrategy: { node: PlacementStrategy.Binpack },
      },
    });
  });

  it("sends explicit null when clearing topology and rejects an active invalid custom label", () => {
    expect(buildTopologyPatch(topologyDraftFromPool(pool), supported)).toHaveProperty("networkTopologyId", null);
    const draft = { ...topologyDraftFromPool(pool), gpuNetworkAccelerationLabelKey: "Invalid Prefix/key" };
    expect(() => buildTopologyPatch(draft, supported)).toThrow(/label/i);
    expect(() => buildTopologyPatch(draft, { ...supported, mnnvl: false })).not.toThrow();
    expect(() =>
      buildTopologyPatch({ ...draft, gpuNetworkAccelerationDetection: Detection.DontUse }, supported),
    ).not.toThrow();
  });
});

describe("Kubernetes topology label keys", () => {
  it.each([
    "",
    "  ",
    "clique",
    "GPU_Cl.ique-1",
    "example.com/gpu.clique",
    "a".repeat(63),
    `${"a".repeat(63)}.${"b".repeat(63)}.${"c".repeat(63)}.${"d".repeat(61)}/key`,
  ])("accepts a valid qualified key or blank default: %s", (value) => {
    expect(validateTopologyLabelKey(value)).toBe(true);
  });

  it.each([
    "/key",
    "example.com/",
    "example.com/a/b",
    "Upper.example/key",
    "example..com/key",
    "-example.com/key",
    "example-.com/key",
    "example.com/-key",
    "example.com/key_",
    "example.com/key.",
    "a key",
    " key",
    "key ",
    "key$",
    "a".repeat(64),
    `${"a".repeat(64)}.example/key`,
    `${"a".repeat(63)}.${"b".repeat(63)}.${"c".repeat(63)}.${"d".repeat(62)}/key`,
  ])("rejects an invalid qualified key: %s", (value) => {
    expect(validateTopologyLabelKey(value)).toEqual(expect.any(String));
  });
});
