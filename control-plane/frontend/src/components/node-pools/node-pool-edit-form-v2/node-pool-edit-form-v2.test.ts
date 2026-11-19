/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, shallowMount, flushPromises } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import { Quasar, QSelect, QInput, Notify } from "quasar";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { nodeLabelsService } from "@/services/control-plane/node-labels.service/node-labels.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import RunaiNameValidation from "@/components/common/runai-name-validation/runai-name-validation.vue";
import RunaiDescriptionInput from "@/components/common/runai-description-input/runai-description-input.vue";
import NodePoolEditFormV2 from "./node-pool-edit-form-v2.vue";
import { BadRequestError } from "@/models/http-response.model";
import { ENABLE_NODE_POOL_CREATE_FLOW } from "@/constants/node-pool.constants";
import { requestToLeave } from "@/services/infra/router.service/router.service";
import { eventEmitter } from "@/services/infra/event-emitter.service/event-emitter.service";
import { NAVIGATION_EVENTS } from "@/common/event.constant";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: { getNodepools: vi.fn().mockResolvedValue([]), createNodepool: vi.fn() },
}));
vi.mock("@/services/control-plane/node-labels.service/node-labels.service", () => ({
  nodeLabelsService: { keys: vi.fn().mockResolvedValue([]), values: vi.fn().mockResolvedValue([]) },
}));
describe("Step 1 form", () => {
  beforeEach(() => vi.clearAllMocks());
  it("keeps the new flow disabled by default", () => expect(ENABLE_NODE_POOL_CREATE_FLOW).toBe(false));
  it("shows description and reveals the value field only after a key is selected", async () => {
    const wrapper = shallowMount(NodePoolEditFormV2, {
      global: {
        plugins: [[Quasar, { plugins: { Notify } }], createTestingPinia({ createSpy: vi.fn })],
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
    await flushPromises();
    expect(wrapper.find('[aid="node-pool-label-value"]').exists()).toBe(true);
    wrapper.unmount();
  });
  function renderForm(initialClusterId = "") {
    return mount(NodePoolEditFormV2, {
      props: { initialClusterId },
      global: {
        plugins: [[Quasar, { plugins: { Notify } }], createTestingPinia({ createSpy: vi.fn })],
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
    expect(nodeLabelsService.keys).toHaveBeenLastCalledWith("cluster", { search: "gpu", offset: 0, limit: 50 });
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
  it("shows the backend error detail, preserves input and clears the error on successful retry", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(clusterService.createNodepool)
      .mockRejectedValueOnce(new BadRequestError("This label is reserved by the cluster administrator"))
      .mockResolvedValueOnce({ id: "created-pool", name: "valid-pool", clusterId: "cluster" } as never);
    const wrapper = renderForm("cluster");
    const notify = vi.spyOn(wrapper.vm.$q, "notify").mockImplementation(() => () => {});
    await flushPromises();
    wrapper.findComponent(RunaiNameValidation).vm.$emit("update:modelValue", "valid-pool");
    wrapper.findComponent(RunaiDescriptionInput).vm.$emit("update:modelValue", "Keep this description");
    wrapper
      .findAllComponents(QSelect)
      .find((c) => c.props("label") === "Key")
      ?.vm.$emit("update:modelValue", "node-type");
    await flushPromises();
    wrapper
      .findAllComponents(QSelect)
      .find((c) => c.props("label") === "Value")
      ?.vm.$emit("update:modelValue", "training");
    await flushPromises();
    await wrapper.find('[aid="create-node-pool-continue"]').trigger("click");
    await flushPromises();
    expect(wrapper.get('[role="alert"]').text()).toContain("This label is reserved by the cluster administrator");
    expect(notify).toHaveBeenCalledWith(
      expect.objectContaining({ message: "Node pool valid-pool couldn't be created" }),
    );
    expect(wrapper.findComponent(RunaiNameValidation).props("modelValue")).toBe("valid-pool");
    expect(wrapper.findComponent(RunaiDescriptionInput).props("modelValue")).toBe("Keep this description");
    expect(wrapper.findComponent(RunaiNameValidation).findComponent(QInput).props("disable")).toBe(false);
    await wrapper.find('[aid="create-node-pool-continue"]').trigger("click");
    await flushPromises();
    expect(wrapper.text()).not.toContain("This label is reserved by the cluster administrator");
    expect(wrapper.emitted("created")).toHaveLength(1);
    expect(clusterService.createNodepool).toHaveBeenCalledTimes(2);
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
  it("refuses header Close while creation is pending without queuing a leave dialog", async () => {
    let finish!: (value: Nodepool) => void;
    vi.mocked(clusterService.getNodepools).mockResolvedValue([]);
    vi.mocked(clusterService.createNodepool)
      .mockReset()
      .mockImplementation(
        () =>
          new Promise((resolve) => {
            finish = resolve;
          }),
      );
    const wrapper = renderForm("cluster");
    vi.spyOn(wrapper.vm.$q, "notify").mockImplementation(() => () => {});
    try {
      await flushPromises();
      wrapper.findComponent(RunaiNameValidation).vm.$emit("update:modelValue", "pending-pool");
      wrapper
        .findAllComponents(QSelect)
        .find((input) => input.props("label") === "Key")
        ?.vm.$emit("update:modelValue", "night.test/pool");
      await flushPromises();
      wrapper
        .findAllComponents(QSelect)
        .find((input) => input.props("label") === "Value")
        ?.vm.$emit("update:modelValue", "pending");
      await flushPromises();
      await wrapper.get('[aid="create-node-pool-continue"]').trigger("click");
      await flushPromises();
      expect(clusterService.createNodepool).toHaveBeenCalledTimes(1);
      let allowed: boolean | undefined;
      void requestToLeave().then((value) => {
        allowed = value;
      });
      await flushPromises();
      expect(allowed).toBe(false);
      expect(document.querySelector('[aid="leave-page-modal-cancel-btn"]')).toBeNull();
      expect(wrapper.emitted("leave-approved")).toBeUndefined();
    } finally {
      eventEmitter.emit(NAVIGATION_EVENTS.LEAVE_PAGE_STATUS, { allowToLeave: false });
      finish?.({ id: "created", name: "pending-pool", clusterId: "cluster" } as Nodepool);
      await flushPromises();
      wrapper.unmount();
    }
  });
});
