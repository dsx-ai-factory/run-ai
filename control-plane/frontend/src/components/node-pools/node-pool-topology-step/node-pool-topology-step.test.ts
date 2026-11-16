/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { createMemoryHistory, createRouter } from "vue-router";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Step from "./node-pool-topology-step.vue";
import { saveTopologyPool } from "./node-pool-topology.service";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import {
  NodepoolGPUNetworkAccelerationDetection as Detection,
  type Nodepool,
} from "@/swagger-models/cluster-service-client";
import Radio from "@/components/common/runai-radio-options/runai-radio-options.vue";
vi.mock("./node-pool-topology.service", () => ({ saveTopologyPool: vi.fn() }));
const pool = {
  ...nodePoolUtil.createNodePoolDraft("cluster"),
  id: "pool",
  name: "Training",
  clusterName: "Tokyo",
  gpuNetworkAccelerationLabelKey: "example.com/rack",
} as Nodepool;
const supports = { mnnvl: true, numa: true, gpuDevice: true, networkTopology: true };
async function render(overrides = {}) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: "/", component: { template: "<div />" } }],
  });
  await router.push("/");
  const wrapper = mount(Step, {
    props: { pool, supports, canSave: true, canReadTopologies: true, canCreateTopology: true, ...overrides },
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn }), router],
      stubs: {
        NetworkTopologySelection: {
          name: "NetworkTopologySelection",
          props: ["networkTopologyId"],
          emits: ["loaded", "select", "creating"],
          template: "<div aid='topology-selector' />",
        },
      },
    },
  });
  await flushPromises();
  return wrapper;
}
beforeEach(() => vi.clearAllMocks());
describe("node pool topology step", () => {
  it("preserves a saved Auto custom label and its draft when MNNVL is hidden", async () => {
    const wrapper = await render();
    expect(wrapper.get<HTMLInputElement>('[aid="topology-mnnvl-label"]').element.value).toBe("example.com/rack");
    const options = wrapper.findAllComponents(Radio)[0];
    options.vm.$emit("update:model-value", Detection.DontUse);
    await flushPromises();
    expect(wrapper.find('[aid="topology-mnnvl-label"]').exists()).toBe(false);
    options.vm.$emit("update:model-value", Detection.Use);
    await flushPromises();
    expect(wrapper.get<HTMLInputElement>('[aid="topology-mnnvl-label"]').element.value).toBe("example.com/rack");
    wrapper.unmount();
  });
  it("rejects an invalid label before issuing a save and explains allowed characters", async () => {
    const wrapper = await render({ pool: { ...pool, gpuNetworkAccelerationLabelKey: "bad key" } });
    expect(wrapper.text()).toContain("optional DNS prefix");
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    expect(saveTopologyPool).not.toHaveBeenCalled();
    wrapper.unmount();
  });
  it("reflects draft placement and topology names in the shared summary", async () => {
    const wrapper = await render();
    const selector = wrapper.findComponent({ name: "NetworkTopologySelection" });
    selector.vm.$emit("loaded", [{ id: "fabric", name: "Training fabric" }]);
    selector.vm.$emit("select", "fabric");
    wrapper.findAllComponents(Radio)[1].vm.$emit("update:model-value", "spread");
    await flushPromises();
    expect(wrapper.get(".summary-topology").text()).toContain("Network topology: Training fabric");
    expect(wrapper.get(".summary-topology").text()).toContain("GPU node placement strategy: Spread");
    wrapper.unmount();
  });
  it("keeps the draft after a failed save and prevents duplicate submissions", async () => {
    let reject!: (reason: Error) => void;
    vi.mocked(saveTopologyPool).mockImplementationOnce(
      () =>
        new Promise((_resolve, fail) => {
          reject = fail;
        }),
    );
    const wrapper = await render();
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    await flushPromises();
    expect(saveTopologyPool).toHaveBeenCalledTimes(1);
    expect(wrapper.emitted("saved")).toBeUndefined();
    reject(new Error("Service unavailable"));
    await flushPromises();
    expect(wrapper.get('[role="alert"]').text()).toContain("Service unavailable");
    expect(wrapper.get<HTMLInputElement>('[aid="topology-mnnvl-label"]').element.value).toBe("example.com/rack");
    vi.mocked(saveTopologyPool).mockResolvedValue(pool);
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    await flushPromises();
    expect(wrapper.emitted("saved")).toEqual([[pool]]);
    wrapper.unmount();
  });
  it("hides unsupported NUMA and GPU device controls and disables old-cluster MNNVL", async () => {
    const wrapper = await render({ supports: { ...supports, numa: false, gpuDevice: false, mnnvl: false } });
    expect(wrapper.find('[aid="topology-numa"]').exists()).toBe(false);
    expect(wrapper.find('[aid="topology-gpu-device-placement"]').exists()).toBe(false);
    expect(wrapper.findAllComponents(Radio)[0].props("disable")).toBe(true);
    expect(wrapper.get(".summary-topology").text()).not.toContain("GPU device placement");
    wrapper.unmount();
  });
  it("saves other fields when an unsupported cluster has an inactive invalid MNNVL label", async () => {
    vi.mocked(saveTopologyPool).mockResolvedValue(pool);
    const wrapper = await render({
      pool: { ...pool, gpuNetworkAccelerationDetection: Detection.Use, gpuNetworkAccelerationLabelKey: "invalid key" },
      supports: { ...supports, mnnvl: false },
    });
    expect(wrapper.findAllComponents(Radio)[0].props("modelValue")).toBe(Detection.Auto);
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    await flushPromises();
    expect(saveTopologyPool).toHaveBeenCalledTimes(1);
    expect(vi.mocked(saveTopologyPool).mock.calls[0][1]).not.toHaveProperty("gpuNetworkAccelerationDetection");
    wrapper.unmount();
  });
  it("blocks navigation and pool saving while a topology creation request is pending", async () => {
    const wrapper = await render();
    const selector = wrapper.findComponent({ name: "NetworkTopologySelection" });
    selector.vm.$emit("creating", true);
    await flushPromises();
    expect(wrapper.emitted("saving")?.at(-1)).toEqual([true]);
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    expect(saveTopologyPool).not.toHaveBeenCalled();
    selector.vm.$emit("creating", false);
    await flushPromises();
    expect(wrapper.emitted("saving")?.at(-1)).toEqual([false]);
    wrapper.unmount();
  });
  it("blocks writes with no pool update permission", async () => {
    const wrapper = await render({ canSave: false });
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    expect(saveTopologyPool).not.toHaveBeenCalled();
    expect(wrapper.text()).toContain("View only");
    wrapper.unmount();
  });
});
