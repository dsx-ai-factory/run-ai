/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { createMemoryHistory, createRouter } from "vue-router";
import { createTestingPinia } from "@pinia/testing";
import { Quasar, Notify } from "quasar";
import { beforeEach, describe, expect, it, vi } from "vitest";
import Step from "./node-pool-topology-step.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import {
  NodepoolGPUNetworkAccelerationDetection as Detection,
  type Nodepool,
} from "@/swagger-models/cluster-service-client";
import Radio from "@/components/common/runai-radio-options/runai-radio-options.vue";
const pool = {
  ...nodePoolUtil.createNodePoolDraft("cluster"),
  id: "pool",
  name: "Training",
  clusterName: "Tokyo",
  gpuNetworkAccelerationLabelKey: "example.com/rack",
} as Nodepool;
const supports = { mnnvl: true, numa: true, gpuDevice: true };
const patchNodepool = vi.spyOn(clusterService, "patchNodepool");
async function render(overrides = {}) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: "/", component: { template: "<div />" } }],
  });
  await router.push("/");
  const wrapper = mount(Step, {
    props: { pool, supports, canSave: true, canReadTopologies: true, canCreateTopology: true, ...overrides },
    global: {
      plugins: [[Quasar, { plugins: { Notify } }], createTestingPinia({ createSpy: vi.fn }), router],
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
beforeEach(() => patchNodepool.mockReset().mockResolvedValue(pool));
describe("node pool topology step", () => {
  it("opens Cluster & name from a saved pool without saving topology", async () => {
    const wrapper = await render();
    try {
      await wrapper.get('[aid="wizard-step-item-identity"]').trigger("click");
      expect(wrapper.emitted("open-identity")).toHaveLength(1);
      expect(patchNodepool).not.toHaveBeenCalled();
    } finally {
      wrapper.unmount();
    }
  });
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
  it("shows a single field error for an invalid label and saves the corrected input", async () => {
    const wrapper = await render();
    try {
      const input = wrapper.get<HTMLInputElement>('[aid="topology-mnnvl-label"]');
      const save = wrapper.get('[aid="topology-save-close"]');
      expect(input.attributes("maxlength")).toBe("317");
      await input.setValue("Upper.example/key");
      expect(input.attributes("aria-invalid")).toBe("true");
      const error = wrapper.get(`#${input.attributes("aria-describedby")}`);
      expect(error.text()).toBe(
        "Enter a label / Key must fit the following regular expression (([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?/?([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]",
      );
      expect(wrapper.findAll('[role="alert"]')).toHaveLength(1);
      await save.trigger("click");
      expect(patchNodepool).not.toHaveBeenCalled();
      await input.setValue("example.com/GPU_Rack");
      expect(wrapper.find('[role="alert"]').exists()).toBe(false);
      expect(input.attributes("aria-invalid")).toBeUndefined();
      await save.trigger("click");
      await flushPromises();
      expect(patchNodepool).toHaveBeenCalledWith(pool.id, { gpuNetworkAccelerationLabelKey: "example.com/GPU_Rack" });
    } finally {
      wrapper.unmount();
    }
  });
  it("closes an unedited pool without making a patch request", async () => {
    const wrapper = await render();
    try {
      expect(wrapper.find('[aid="topology-cancel"]').exists()).toBe(false);
      expect(wrapper.find("#node-pool-topology-label-help").exists()).toBe(false);
      expect(wrapper.find('.q-list[role="list"]').exists()).toBe(false);
      await wrapper.get('[aid="topology-save-close"]').trigger("click");
      await flushPromises();
      expect(patchNodepool).not.toHaveBeenCalled();
      expect(wrapper.emitted("saved")).toEqual([[pool]]);
    } finally {
      wrapper.unmount();
    }
  });
  it("reflects draft placement and topology names in the shared summary", async () => {
    const wrapper = await render();
    const selector = wrapper.findComponent({ name: "NetworkTopologySelection" });
    selector.vm.$emit("loaded", [{ id: "fabric", name: "Training fabric" }]);
    selector.vm.$emit("select", "fabric");
    wrapper.findAllComponents(Radio)[1].vm.$emit("update:model-value", "spread");
    await flushPromises();
    const topologyValues = Object.fromEntries(
      wrapper.findAll(".summary-topology dl > div").map((row) => [row.get("dt").text(), row.get("dd").text()]),
    );
    expect(topologyValues).toMatchObject({
      "Network topology": "Training fabric",
      "GPU node placement strategy": "Spread",
    });
    wrapper.unmount();
  });
  it("keeps the draft after a failed save and prevents duplicate submissions", async () => {
    let reject!: (reason: Error) => void;
    patchNodepool.mockImplementationOnce(
      () =>
        new Promise((_resolve, fail) => {
          reject = fail;
        }),
    );
    const wrapper = await render();
    const notify = vi.spyOn(wrapper.vm.$q, "notify").mockImplementation(() => () => {});
    await wrapper.get('[aid="topology-mnnvl-label"]').setValue("example.com/Edited");
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    await flushPromises();
    expect(patchNodepool).toHaveBeenCalledTimes(1);
    expect(patchNodepool).toHaveBeenCalledWith(pool.id, { gpuNetworkAccelerationLabelKey: "example.com/Edited" });
    expect(wrapper.emitted("saved")).toBeUndefined();
    reject(new Error("Service unavailable"));
    await flushPromises();
    expect(wrapper.find('[role="alert"]').exists()).toBe(false);
    expect(notify).toHaveBeenCalledTimes(1);
    expect(notify).toHaveBeenCalledWith(
      expect.objectContaining({ type: "negative", message: "Node pool Training couldn't be saved" }),
    );
    expect(wrapper.get<HTMLInputElement>('[aid="topology-mnnvl-label"]').element.value).toBe("example.com/Edited");
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
    const wrapper = await render({
      pool: { ...pool, gpuNetworkAccelerationDetection: Detection.Use, gpuNetworkAccelerationLabelKey: "invalid key" },
      supports: { ...supports, mnnvl: false },
    });
    expect(wrapper.findAllComponents(Radio)[0].props("modelValue")).toBe(Detection.Auto);
    wrapper.findAllComponents(Radio)[1].vm.$emit("update:model-value", "spread");
    await flushPromises();
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    await flushPromises();
    expect(patchNodepool).toHaveBeenCalledWith(pool.id, {
      schedulingConfiguration: { gpuPodsPlacementStrategy: { node: "spread" } },
    });
    wrapper.unmount();
  });
  it("preserves an untouched legacy invalid label when saving unrelated edits", async () => {
    const wrapper = await render({ pool: { ...pool, gpuNetworkAccelerationLabelKey: "legacy key" } });
    try {
      wrapper.findAllComponents(Radio)[1].vm.$emit("update:model-value", "spread");
      await wrapper.get('[aid="topology-save-close"]').trigger("click");
      await flushPromises();
      expect(patchNodepool).toHaveBeenCalledWith(pool.id, {
        schedulingConfiguration: { gpuPodsPlacementStrategy: { node: "spread" } },
      });
    } finally {
      wrapper.unmount();
    }
  });
  it("validates a saved label when detection is enabled", async () => {
    const wrapper = await render({
      pool: {
        ...pool,
        gpuNetworkAccelerationDetection: Detection.DontUse,
        gpuNetworkAccelerationLabelKey: "legacy key",
      },
    });
    try {
      wrapper.findAllComponents(Radio)[0].vm.$emit("update:model-value", Detection.Use);
      await flushPromises();
      expect(wrapper.get('[aid="topology-mnnvl-label"]').attributes("aria-invalid")).toBe("true");
      expect(wrapper.get('[aid="topology-save-close"]').attributes("disabled")).toBeDefined();
      await wrapper.get('[aid="topology-mnnvl-label"]').setValue("");
      await wrapper.get('[aid="topology-save-close"]').trigger("click");
      await flushPromises();
      expect(patchNodepool).toHaveBeenCalledWith(pool.id, {
        gpuNetworkAccelerationDetection: Detection.Use,
        gpuNetworkAccelerationLabelKey: "nvidia.com/gpu.clique",
      });
    } finally {
      wrapper.unmount();
    }
  });
  it("skips an edited topology when read permission is revoked before saving", async () => {
    const wrapper = await render({ pool: { ...pool, networkTopologyId: "original" } });
    wrapper.getComponent({ name: "NetworkTopologySelection" }).vm.$emit("select", null);
    await wrapper.setProps({ canReadTopologies: false });
    expect(wrapper.findComponent({ name: "NetworkTopologySelection" }).exists()).toBe(false);
    await wrapper.get('[aid="topology-mnnvl-label"]').setValue("");
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    await flushPromises();
    expect(patchNodepool).toHaveBeenCalledWith(pool.id, { gpuNetworkAccelerationLabelKey: "nvidia.com/gpu.clique" });
    wrapper.unmount();
  });
  it("blocks navigation and pool saving while a topology creation request is pending", async () => {
    const wrapper = await render();
    const selector = wrapper.findComponent({ name: "NetworkTopologySelection" });
    selector.vm.$emit("creating", true);
    await flushPromises();
    expect(wrapper.emitted("saving")?.at(-1)).toEqual([true]);
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    expect(patchNodepool).not.toHaveBeenCalled();
    selector.vm.$emit("creating", false);
    await flushPromises();
    expect(wrapper.emitted("saving")?.at(-1)).toEqual([false]);
    wrapper.unmount();
  });
  it("blocks writes with no pool update permission", async () => {
    const wrapper = await render({ canSave: false });
    expect(wrapper.get("fieldset").attributes("disabled")).toBeDefined();
    expect(wrapper.get('[aid="topology-mnnvl-label"]').attributes("readonly")).toBeDefined();
    expect(wrapper.get('[aid="topology-save-close"]').attributes("disabled")).toBeDefined();
    await wrapper.get('[aid="topology-save-close"]').trigger("click");
    expect(patchNodepool).not.toHaveBeenCalled();
    expect(wrapper.text()).not.toContain("View only");
    wrapper.unmount();
  });
});
