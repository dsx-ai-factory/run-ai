/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import { getSchedulingMode } from "@/models/node-pool.model";
import NodePoolWizardSummary from "./node-pool-wizard-summary.vue";
import { createNodePoolDraft } from "@/utils/node-pool.util/node-pool.util";
import { NodepoolGPUNetworkAccelerationDetection, PlacementStrategy } from "@/swagger-models/cluster-service-client";

describe("node pool wizard summary", () => {
  it("distinguishes guaranteed quota and custom settings from proportional", () => {
    expect(getSchedulingMode({ quota: true, ranks: true, adjustableWeights: false, cpuQuota: false })).toBe(
      "Guaranteed quota",
    );
    expect(getSchedulingMode({ quota: false, ranks: false, adjustableWeights: false, cpuQuota: false })).toBe("Custom");
  });
  function render() {
    return mount(NodePoolWizardSummary, {
      props: { pool: createNodePoolDraft("cluster"), clusterName: "Tokyo" },
      global: { plugins: [Quasar] },
    });
  }
  it("shows initial configuration and unknown resources without claiming zero capacity", () => {
    const wrapper = render();
    expect(wrapper.text()).toContain("Tokyo");
    expect(wrapper.text()).toContain("Proportional");
    expect(wrapper.text()).toContain("Nodes: None");
    expect(wrapper.text()).toContain("MNNVL: Auto detect");
    expect(wrapper.text()).toContain("NUMA-aware scheduling: Disabled");
    expect(wrapper.text()).toContain("GPU device placement strategy: Bin-pack");
    wrapper.unmount();
  });
  it("updates from changed draft or saved configuration including false values", async () => {
    const wrapper = render();
    await wrapper.setProps({
      pool: {
        ...createNodePoolDraft("cluster"),
        networkTopologyId: "topology-id",
        networkTopologyName: "Fabric",
        schedulingCapabilities: { adjustableWeights: true, quota: false, ranks: true, cpuQuota: false },
        gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.Use,
        schedulingConfiguration: {
          numaAwareScheduling: true,
          gpuPodsPlacementStrategy: { node: PlacementStrategy.Spread, device: PlacementStrategy.Spread },
          cpuPodsPlacementStrategy: { node: PlacementStrategy.Spread },
        },
      },
    });
    expect(wrapper.text()).toContain("Ranked");
    expect(wrapper.text()).not.toContain("Proportional");
    expect(wrapper.text()).toContain("Network topology: Fabric");
    expect(wrapper.text()).toContain("MNNVL: Is present");
    expect(wrapper.text()).toContain("NUMA-aware scheduling: Enabled");
    expect(wrapper.text()).toContain("GPU device placement strategy: Spread");
    await wrapper.setProps({
      pool: {
        ...createNodePoolDraft(),
        schedulingCapabilities: undefined,
        gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.DontUse,
      },
    });
    expect(wrapper.text()).toContain("Scheduling modeUnknown");
    expect(wrapper.text()).toContain("MNNVL: Is not present");
    expect(wrapper.text()).toContain("NUMA-aware scheduling: Disabled");
    wrapper.unmount();
  });
  it("renders measured zero and supplied resource totals distinctly from unknown", async () => {
    const wrapper = render();
    await wrapper.setProps({ resources: { nodes: 0, gpus: 0, gpuMemoryBytes: 0, gpuTypes: [] } });
    expect(wrapper.text()).toContain("Nodes: 0");
    expect(wrapper.text()).toContain("GPUs: 0");
    expect(wrapper.text()).toContain("GPU memory: 0 Bytes");
    await wrapper.setProps({ resources: { nodes: 2, gpus: 8, gpuTypes: ["A100", "H100"] } });
    expect(wrapper.text()).toContain("Nodes: 2");
    expect(wrapper.text()).toContain("GPUs: 8");
    expect(wrapper.text()).toContain("GPU types: A100, H100");
    expect(wrapper.text()).toContain("GPU memory: None");
    wrapper.unmount();
  });
});
