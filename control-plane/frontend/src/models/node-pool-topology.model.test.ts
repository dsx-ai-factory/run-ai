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
import { buildTopologyPatch, topologyDraftFromPool, type ITopologyDraft } from "./node-pool-topology.model";
import { DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY } from "./node-pool.model";

const pool = { id: "p", clusterId: "c", name: "Pool" } as Nodepool;
const supported = { mnnvl: true, numa: true, gpuDevice: true };
const original = topologyDraftFromPool(pool);

describe("topology draft and patch", () => {
  it("uses server defaults for missing settings and keeps an unconfigured label blank", () => {
    expect(original).toEqual({
      networkTopologyId: null,
      gpuNetworkAccelerationDetection: Detection.Auto,
      gpuNetworkAccelerationLabelKey: "",
      numaAwareScheduling: false,
      gpuNodePlacement: PlacementStrategy.Binpack,
      gpuDevicePlacement: PlacementStrategy.Binpack,
      cpuNodePlacement: PlacementStrategy.Binpack,
    });
    expect(buildTopologyPatch(original, { ...original }, supported, true)).toEqual({});
  });

  it("reads existing settings and omits untouched fields from a single placement edit", () => {
    const initial = topologyDraftFromPool({
      ...pool,
      networkTopologyId: "topology",
      gpuNetworkAccelerationLabelKey: "custom.example/clique",
      schedulingConfiguration: {
        numaAwareScheduling: true,
        gpuPodsPlacementStrategy: { node: PlacementStrategy.Spread, device: PlacementStrategy.Binpack },
        cpuPodsPlacementStrategy: { node: PlacementStrategy.Spread },
        minGuaranteedRuntime: "10m",
        timeBasedFairShare: { enabled: true, historicalUsageWeight: 2 },
      },
    });
    expect(initial).toEqual({
      networkTopologyId: "topology",
      gpuNetworkAccelerationDetection: Detection.Auto,
      gpuNetworkAccelerationLabelKey: "custom.example/clique",
      numaAwareScheduling: true,
      gpuNodePlacement: PlacementStrategy.Spread,
      gpuDevicePlacement: PlacementStrategy.Binpack,
      cpuNodePlacement: PlacementStrategy.Spread,
    });
    expect(
      buildTopologyPatch(initial, { ...initial, gpuDevicePlacement: PlacementStrategy.Spread }, supported, true),
    ).toEqual({
      schedulingConfiguration: { gpuPodsPlacementStrategy: { device: PlacementStrategy.Spread } },
    });
  });

  it("reads legacy node placement strategies when the new fields are missing", () => {
    const initial = topologyDraftFromPool({
      ...pool,
      schedulingConfiguration: {
        placementStrategy: { gpu: PlacementStrategy.Spread, cpu: PlacementStrategy.Spread },
        gpuPodsPlacementStrategy: { node: PlacementStrategy.Binpack },
      },
    });
    expect(initial.gpuNodePlacement).toBe(PlacementStrategy.Binpack);
    expect(initial.cpuNodePlacement).toBe(PlacementStrategy.Spread);
    expect(buildTopologyPatch(initial, { ...initial }, supported, true)).toEqual({});
  });

  it("sends only the edited CPU, GPU and NUMA fields, including false", () => {
    const initial = { ...original, numaAwareScheduling: true };
    const draft = {
      ...initial,
      numaAwareScheduling: false,
      gpuNodePlacement: PlacementStrategy.Spread,
      cpuNodePlacement: PlacementStrategy.Spread,
    };
    expect(buildTopologyPatch(initial, draft, supported, true)).toEqual({
      schedulingConfiguration: {
        numaAwareScheduling: false,
        gpuPodsPlacementStrategy: { node: PlacementStrategy.Spread },
        cpuPodsPlacementStrategy: { node: PlacementStrategy.Spread },
      },
    });
  });

  it("distinguishes changing, clearing and preserving a saved topology", () => {
    const initial = { ...original, networkTopologyId: "old" };
    const changed = { ...initial, networkTopologyId: "new" };
    expect(buildTopologyPatch(initial, changed, supported, true)).toEqual({ networkTopologyId: "new" });
    expect(buildTopologyPatch(initial, original, supported, true)).toEqual({ networkTopologyId: null });
    expect(buildTopologyPatch(initial, changed, supported, false)).toEqual({});
    expect(buildTopologyPatch(initial, initial, supported, true)).toEqual({});
  });

  it.each([Detection.Auto, Detection.Use])("restores the default when a custom %s label is cleared", (mode) => {
    const initial = { ...original, gpuNetworkAccelerationDetection: mode, gpuNetworkAccelerationLabelKey: "custom" };
    const draft = { ...initial, gpuNetworkAccelerationLabelKey: "  " };
    expect(buildTopologyPatch(initial, draft, supported, true)).toEqual({
      gpuNetworkAccelerationLabelKey: DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY,
    });
    expect(draft.gpuNetworkAccelerationLabelKey).toBe("  ");
  });

  it("omits blank and explicit defaults that represent the same label", () => {
    const explicitDefault = { ...original, gpuNetworkAccelerationLabelKey: DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY };
    expect(buildTopologyPatch(explicitDefault, original, supported, true)).toEqual({});
    expect(buildTopologyPatch(original, explicitDefault, supported, true)).toEqual({});
    expect(
      buildTopologyPatch(original, { ...original, gpuNetworkAccelerationDetection: Detection.Use }, supported, true),
    ).toEqual({
      gpuNetworkAccelerationDetection: Detection.Use,
    });
  });

  it.each([Detection.Auto, Detection.Use])(
    "uses the default key when legacy detection changes from DontUse to %s",
    (mode) => {
      const initial = topologyDraftFromPool({ ...pool, gpuNetworkAccelerationDetection: Detection.DontUse });
      expect(
        buildTopologyPatch(initial, { ...initial, gpuNetworkAccelerationDetection: mode }, supported, true),
      ).toEqual({
        gpuNetworkAccelerationDetection: mode,
      });
    },
  );

  it("persists corrections to a legacy label with surrounding whitespace", () => {
    const initial = { ...original, gpuNetworkAccelerationLabelKey: "example.com/rack " };
    const draft = { ...initial, gpuNetworkAccelerationLabelKey: "example.com/rack" };
    expect(buildTopologyPatch(initial, draft, supported, true)).toEqual({
      gpuNetworkAccelerationLabelKey: "example.com/rack",
    });
  });

  it("omits an inactive label and keeps the draft for reactivation", () => {
    const draft: ITopologyDraft = {
      ...original,
      gpuNetworkAccelerationDetection: Detection.DontUse,
      gpuNetworkAccelerationLabelKey: "custom.example/retained",
    };
    expect(buildTopologyPatch(original, draft, supported, true)).toEqual({
      gpuNetworkAccelerationDetection: Detection.DontUse,
    });
    draft.gpuNetworkAccelerationDetection = Detection.Use;
    expect(buildTopologyPatch(original, draft, supported, true)).toEqual({
      gpuNetworkAccelerationDetection: Detection.Use,
      gpuNetworkAccelerationLabelKey: "custom.example/retained",
    });
  });

  it("skips fields that became unsupported while the user was editing", () => {
    const draft = {
      ...original,
      gpuNetworkAccelerationDetection: Detection.Use,
      gpuNetworkAccelerationLabelKey: "changed",
      gpuDevicePlacement: PlacementStrategy.Spread,
      gpuNodePlacement: PlacementStrategy.Spread,
      numaAwareScheduling: true,
    };
    expect(buildTopologyPatch(original, draft, { mnnvl: false, numa: false, gpuDevice: false }, true)).toEqual({
      schedulingConfiguration: { gpuPodsPlacementStrategy: { node: PlacementStrategy.Spread } },
    });
  });
});
