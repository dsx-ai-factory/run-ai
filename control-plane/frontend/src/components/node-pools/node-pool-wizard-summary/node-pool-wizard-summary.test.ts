/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import { getSchedulingMode } from "@/models/node-pool.model";
import NodePoolWizardSummary from "./node-pool-wizard-summary.vue";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
const { createNodePoolDraft } = nodePoolUtil;
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
  function summaryValues(wrapper: ReturnType<typeof render>) {
    return Object.fromEntries(wrapper.findAll("dl > div").map((row) => [row.get("dt").text(), row.get("dd").text()]));
  }
  it("shows initial configuration and unknown resources without claiming zero capacity", () => {
    const wrapper = render();
    expect(wrapper.text()).toContain("Tokyo");
    expect(wrapper.text()).toContain("Proportional");
    expect(summaryValues(wrapper)).toMatchObject({
      Nodes: "None",
      MNNVL: "Auto detect",
      "NUMA-aware scheduling": "Disabled",
      "GPU device placement strategy": "Bin-pack",
    });
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
    expect(summaryValues(wrapper)).toMatchObject({
      "Network topology": "Fabric",
      MNNVL: "Is present",
      "NUMA-aware scheduling": "Enabled",
      "GPU device placement strategy": "Spread",
    });
    await wrapper.setProps({
      pool: {
        ...createNodePoolDraft(),
        schedulingCapabilities: undefined,
        gpuNetworkAccelerationDetection: NodepoolGPUNetworkAccelerationDetection.DontUse,
      },
    });
    expect(wrapper.text()).toContain("Scheduling modeUnknown");
    expect(summaryValues(wrapper)).toMatchObject({
      MNNVL: "Is not present",
      "NUMA-aware scheduling": "Disabled",
    });
    wrapper.unmount();
  });
  it("omits configuration unsupported by the selected cluster when opted in", async () => {
    const wrapper = render();
    await wrapper.setProps({ showNuma: false, showGpuDevice: false });
    const values = summaryValues(wrapper);
    expect(values).not.toHaveProperty("NUMA-aware scheduling");
    expect(values).not.toHaveProperty("GPU device placement strategy");
    expect(values["GPU node placement strategy"]).toBe("Bin-pack");
    wrapper.unmount();
  });
  it("renders measured zero and supplied resource totals distinctly from unknown", async () => {
    const wrapper = render();
    await wrapper.setProps({ resources: { nodes: 0, gpus: 0, gpuMemoryBytes: 0, gpuTypes: [] } });
    expect(summaryValues(wrapper)).toMatchObject({ Nodes: "0", GPUs: "0", "GPU memory": "0 Bytes" });
    await wrapper.setProps({ resources: { nodes: 2, gpus: 8, gpuTypes: ["A100", "H100"] } });
    expect(summaryValues(wrapper)).toMatchObject({
      Nodes: "2",
      GPUs: "8",
      "GPU types": "A100, H100",
      "GPU memory": "None",
    });
    wrapper.unmount();
  });
});
