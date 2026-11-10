/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, shallowMount, flushPromises } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { Quasar, QSelect, QInput } from "quasar";
import { describe, it, expect, vi } from "vitest";
import { nodeLabelsService } from "@/services/control-plane/node-labels.service/node-labels.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import RunaiNameValidation from "@/components/common/runai-name-validation/runai-name-validation.vue";
import RunaiDescriptionInput from "@/components/common/runai-description-input/runai-description-input.vue";
import NodePoolEditFormV2 from "./node-pool-edit-form-v2.vue";
import { ENABLE_NODE_POOL_CREATE_FLOW } from "@/constants/node-pool.constants";
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: { getNodepools: vi.fn().mockResolvedValue([]), createNodepool: vi.fn() },
}));
vi.mock("@/services/control-plane/node-labels.service/node-labels.service", () => ({
  nodeLabelsService: { keys: vi.fn().mockResolvedValue([]), values: vi.fn().mockResolvedValue([]) },
}));
describe("Step 1 form", () => {
  it("keeps the new flow disabled by default", () => expect(ENABLE_NODE_POOL_CREATE_FLOW).toBe(false));
  it("shows description and reveals the value field only after a key is selected", async () => {
    const wrapper = shallowMount(NodePoolEditFormV2, {
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        renderStubDefaultSlot: true,
        stubs: {
          QInput: { props: ["prefix"], template: "<div />" },
          QSelect: { props: ["prefix"], emits: ["update:model-value"], template: "<div />" },
        },
      },
    });
    expect(wrapper.findComponent(RunaiDescriptionInput).props("customAid")).toBe("node-pool-description");
    expect(wrapper.find('[aid="node-pool-label-value"]').exists()).toBe(false);
    wrapper.findComponent(QSelect).vm.$emit("update:model-value", "node-type");
    await wrapper.vm.$nextTick();
    expect(wrapper.find('[aid="node-pool-label-value"]').exists()).toBe(true);
    wrapper.unmount();
  });
  function renderForm(initialClusterId = "") {
    return mount(NodePoolEditFormV2, {
      props: { initialClusterId },
      global: {
        plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
        mocks: { $route: { query: {} } },
        stubs: {
          RunaiWizardLayout: { template: '<div><slot /><slot name="content-footer" /></div>' },
          RunaiExpansionWrapper: { template: "<div><slot /></div>" },
          RunaiExpansionItem: { template: "<section><slot /></section>" },
          RunaiClusterDropdown: {
            components: { QSelect },
            props: ["clusterId", "disabled"],
            template: `<q-select label="Cluster" :model-value="clusterId" :disable="disabled" :rules="[v => !!v || 'Select a cluster']" />`,
          },
        },
      },
    });
  }
  it("populates the key selector from the cluster API and keeps manual entry available after failure", async () => {
    const wrapper = renderForm("cluster");
    await flushPromises();
    vi.mocked(nodeLabelsService.keys).mockResolvedValueOnce(["gpu-type"]);
    const key = wrapper.findAllComponents(QSelect).find((input) => input.props("label") === "Key");
    if (!key) throw new Error("Key selector missing");
    const update = vi.fn((fn) => fn());
    key.vm.$emit("filter", "gpu", update);
    await flushPromises();
    expect(nodeLabelsService.keys).toHaveBeenLastCalledWith("cluster", { contains: "gpu", offset: 0, limit: 50 });
    expect(key.props("options")).toContainEqual({ label: "gpu-type", value: "gpu-type", disable: false });
    expect(update).toHaveBeenCalled();
    vi.mocked(nodeLabelsService.keys).mockRejectedValueOnce(new Error("404"));
    key.vm.$emit("filter", "custom-key", update);
    await flushPromises();
    expect(key.props("disable")).toBe(false);
    expect(key.props("hint")).toContain("manually");
    expect(key.props("options")).toContainEqual({ label: "custom-key (new key)", value: "custom-key", disable: false });
    wrapper.unmount();
  });
  it("validates the shared description input before submitting without a debounce delay", async () => {
    const wrapper = renderForm("cluster");
    await flushPromises();
    const description = wrapper.findAllComponents(QInput).find((input) => input.props("label") === "Description");
    expect(description?.props("debounce")).toBe(0);
    await description?.find("input, textarea").setValue("א".repeat(126));
    await wrapper.find('[aid="create-node-pool-continue"]').trigger("click");
    await flushPromises();
    expect(wrapper.text()).toContain("Description exceeds the 250-byte limit");
    expect(clusterService.createNodepool).not.toHaveBeenCalled();
    wrapper.unmount();
  });
  it("keeps the missing cluster error visible after submitting", async () => {
    const wrapper = renderForm();
    await flushPromises();
    await wrapper.find('[aid="create-node-pool-continue"]').trigger("click");
    await flushPromises();
    expect(wrapper.text()).toContain("Select a cluster");
    expect(wrapper.text()).toContain("Enter a name");
    wrapper.unmount();
  });
  it("keeps missing key and value errors visible after submitting", async () => {
    const wrapper = renderForm("cluster");
    await flushPromises();
    wrapper.findComponent(RunaiNameValidation).vm.$emit("update:modelValue", "valid-pool");
    await flushPromises();
    await wrapper.find('[aid="create-node-pool-continue"]').trigger("click");
    await flushPromises();
    expect(wrapper.text()).toContain("Enter a key");
    wrapper
      .findAllComponents(QSelect)
      .find((c) => c.props("label") === "Key")
      ?.vm.$emit("update:modelValue", "node-type");
    await flushPromises();
    await wrapper.find('[aid="create-node-pool-continue"]').trigger("click");
    await flushPromises();
    expect(wrapper.text()).toContain("Select a value");
    wrapper.unmount();
  });
  it("shows name-format and cluster-specific duplicate-name errors", async () => {
    vi.mocked(clusterService.getNodepools).mockResolvedValueOnce([
      { id: "taken", name: "occupied", clusterId: "cluster" },
    ] as never);
    const wrapper = renderForm("cluster");
    await flushPromises();
    for (const [name, error] of [
      ["1pool", "Enter a name starting with a lowercase letter"],
      ["occupied", "This name already exists"],
    ]) {
      wrapper.findComponent(RunaiNameValidation).vm.$emit("update:modelValue", name);
      await flushPromises();
      await wrapper.find('[aid="create-node-pool-continue"]').trigger("click");
      await flushPromises();
      expect(wrapper.text()).toContain(error);
    }
    expect(clusterService.createNodepool).not.toHaveBeenCalled();
    wrapper.unmount();
  });
});
