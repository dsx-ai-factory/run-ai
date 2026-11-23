/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { DOMWrapper, flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { Notify, QSelect, Quasar } from "quasar";
import type { Component } from "vue";
import NetworkTopologySelection from "./network-topology-selection.vue";
import { RunaiSelect } from "@/components/common/runai-select";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { NetworkTopology, NetworkTopologyCreateRequest } from "@/swagger-models/cluster-service-client";

const topology: NetworkTopology = {
  id: "topology-1",
  name: "Fabric A",
  clusterId: "cluster-1",
  levels: ["rack", "node"],
  tenantId: 1,
  createdAt: "2026-11-16T00:00:00Z",
};
const request: NetworkTopologyCreateRequest = { name: "Fabric A", clusterId: "cluster-1", levels: ["rack", "node"] };
function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: Error) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}
let wrapper: VueWrapper<InstanceType<typeof NetworkTopologySelection>>;
function mountComponent(props = {}, attachTo?: HTMLElement, formStub: Component | boolean = true) {
  wrapper = mount(NetworkTopologySelection, {
    attachTo,
    props: { clusterId: "cluster-1", ...props },
    global: {
      plugins: [[Quasar, { plugins: { Notify } }]],
      stubs: { NetworkTopologyForm: formStub },
      mocks: { $route: { query: {} } },
    },
  });
  return wrapper;
}

beforeEach(() => {
  vi.spyOn(clusterService, "getNetworkTopologies").mockResolvedValue([topology]);
  vi.spyOn(clusterService, "createNetworkTopology").mockResolvedValue(topology);
  vi.spyOn(console, "error").mockImplementation(() => {});
});
afterEach(() => {
  wrapper?.unmount();
  vi.restoreAllMocks();
});

describe("network topology selection", () => {
  it("shows the Far to Close guide alongside the numbered topology levels in the creation dialog", async () => {
    mountComponent({}, document.body, false);
    await flushPromises();
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    await flushPromises();

    const modal = document.body.querySelector('[aid="network-topologies-modal-body"]');
    const guide = modal?.querySelector(".topology-distance-guide");
    expect(Array.from(guide?.querySelectorAll(".topology-distance-label") ?? [], (label) => label.textContent)).toEqual([
      "Far",
      "Close",
    ]);
    expect(guide?.parentElement?.querySelector(".list-numbers-col")?.textContent?.trim()).toBe("1");
  });

  it("moves levels with the row buttons and saves their new order", async () => {
    vi.spyOn(clusterService, "getNetworkTopologyCount").mockResolvedValue(0);
    mountComponent({}, document.body, false);
    await flushPromises();
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    await flushPromises();

    const modal = new DOMWrapper(document.body);
    const moveUp = '[aid="network-topology-move-up"]';
    const moveDown = '[aid="network-topology-move-down"]';
    const add = modal.get('[aid="network-topology-add-label-btn"]');
    const inputs = () => modal.findAll<HTMLInputElement>('input[aid="network-topology-label-key-input"]');
    const values = () => inputs().map((input) => input.element.value);

    expect(modal.find(moveUp).attributes("aria-disabled")).toBe("true");
    expect(modal.get(moveDown).attributes("aria-disabled")).toBe("true");
    await add.trigger("click");
    await add.trigger("click");
    await modal.get<HTMLInputElement>('input[aid="network-topology-name-input"]').setValue("fabric-a");
    const [far, middle, close] = inputs();
    await far.setValue("zone");
    await middle.setValue("rack");
    await close.setValue("zone/Node");

    await modal.findAll(moveUp)[0].trigger("click");
    expect(values()).toEqual(["zone", "rack", "zone/Node"]);
    const raiseRack = modal.findAll<HTMLButtonElement>(moveUp)[1];
    raiseRack.element.focus();
    await raiseRack.trigger("click");
    expect(values()).toEqual(["rack", "zone", "zone/Node"]);
    expect(document.activeElement).toBe(raiseRack.element);
    expect(raiseRack.attributes("aria-disabled")).toBe("true");
    expect(modal.get('[role="status"]').text()).toContain("position 1 of 3");

    await modal.findAll(moveDown)[1].trigger("click");
    expect(values()).toEqual(["rack", "zone/Node", "zone"]);
    expect(modal.findAll(moveDown)[2].attributes("aria-disabled")).toBe("true");
    await modal.get('[aid="network-topologies-modal-save-button"]').trigger("click");
    await flushPromises();
    expect(clusterService.createNetworkTopology).toHaveBeenCalledOnce();
    expect(clusterService.createNetworkTopology).toHaveBeenCalledWith({
      ...request,
      name: "fabric-a",
      levels: ["rack", "zone/Node", "zone"],
    });
  });

  it("locks all creation fields while saving and retains the draft when the request fails", async () => {
    const pending = deferred<NetworkTopology>();
    vi.mocked(clusterService.createNetworkTopology).mockReturnValueOnce(pending.promise);
    mountComponent({}, document.body, {
      data: () => ({ name: "Draft fabric" }),
      template: '<div><input aid="draft-name" v-model="name"><button>Add label</button></div>',
    });
    await flushPromises();
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    await flushPromises();
    const input = document.body.querySelector('[aid="draft-name"]') as HTMLInputElement;
    const button = input.parentElement?.querySelector("button") as HTMLButtonElement;
    expect(input.value).toBe("Draft fabric");
    const creating = wrapper.vm.createNetworkTopology({ ...request, name: input.value });
    await flushPromises();
    expect(input.matches(":disabled")).toBe(true);
    expect(button.matches(":disabled")).toBe(true);
    expect(input.closest("fieldset")?.hasAttribute("inert")).toBe(true);
    pending.reject(new Error("unavailable"));
    await creating;
    await flushPromises();
    expect(input.matches(":disabled")).toBe(false);
    expect(input.closest("fieldset")?.hasAttribute("inert")).toBe(false);
    expect(input.value).toBe("Draft fabric");
    expect(wrapper.getComponent(RunaiBaseDialog).props("modelValue")).toBe(true);
    await wrapper.vm.createNetworkTopology({ ...request, name: input.value });
    expect(clusterService.createNetworkTopology).toHaveBeenLastCalledWith({ ...request, name: "Draft fabric" });
    expect(clusterService.createNetworkTopology).toHaveBeenCalledTimes(2);
  });

  it("returns focus to the selector when the creation menu trigger has been removed", async () => {
    mountComponent({}, document.body);
    await flushPromises();
    const menuAction = document.createElement("button");
    document.body.append(menuAction);
    menuAction.focus();
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    await flushPromises();
    menuAction.remove();
    wrapper.getComponent(RunaiBaseDialog).vm.$emit("close");
    await flushPromises();
    expect(document.activeElement).toBe(wrapper.get('[role="combobox"]').element);
  });

  it("does not restore focus to the selector after it becomes disabled", async () => {
    mountComponent({}, document.body);
    await flushPromises();
    const focus = vi.spyOn(wrapper.get<HTMLElement>('[role="combobox"]').element, "focus");
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    await flushPromises();
    await wrapper.setProps({ disable: true });
    wrapper.getComponent(RunaiBaseDialog).vm.$emit("close");
    await flushPromises();
    expect(focus).not.toHaveBeenCalled();
  });

  it("does not restore focus while a different cluster's topologies are loading", async () => {
    mountComponent({}, document.body);
    await flushPromises();
    const focus = vi.spyOn(wrapper.get<HTMLElement>('[role="combobox"]').element, "focus");
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    await flushPromises();
    const pending = deferred<NetworkTopology[]>();
    vi.mocked(clusterService.getNetworkTopologies).mockReturnValueOnce(pending.promise);
    await wrapper.setProps({ clusterId: "cluster-2" });
    await flushPromises();
    expect(focus).not.toHaveBeenCalled();
    pending.resolve([]);
    await flushPromises();
  });

  it("returns focus to the selector after successful topology creation", async () => {
    mountComponent({}, document.body);
    await flushPromises();
    const menuAction = document.createElement("button");
    document.body.append(menuAction);
    menuAction.focus();
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    await flushPromises();
    menuAction.remove();
    await wrapper.vm.createNetworkTopology(request);
    await flushPromises();
    expect(document.activeElement).toBe(wrapper.get('[role="combobox"]').element);
  });

  it("reports successfully loaded topology records to the summary owner", async () => {
    mountComponent();
    await flushPromises();
    expect(wrapper.emitted("loaded")).toEqual([[[topology]]]);
  });

  it("shows the empty message and create action inside the dropdown", async () => {
    vi.mocked(clusterService.getNetworkTopologies).mockResolvedValue([]);
    mountComponent({}, document.body);
    await flushPromises();
    expect(wrapper.text()).not.toContain("No network topologies found");
    expect(wrapper.find('[role="alert"]').exists()).toBe(false);
    expect(wrapper.emitted("loaded")).toEqual([[[]]]);

    wrapper.getComponent(QSelect).vm.showPopup();
    await flushPromises();
    const menu = new DOMWrapper(document.body).get('[role="listbox"]');
    expect(menu.text()).toContain("No network topologies found");
    await menu.get(".custom-action-item").trigger("click");
    expect(wrapper.getComponent(RunaiBaseDialog).props("modelValue")).toBe(true);
  });

  it("keeps the saved topology visible after a load failure and retries without changing the selection", async () => {
    vi.mocked(clusterService.getNetworkTopologies).mockRejectedValueOnce(new Error("offline"));
    mountComponent({ networkTopologyId: "topology-1", networkTopologyName: "Saved Fabric" });
    await flushPromises();
    expect(wrapper.get('[role="alert"]').text()).toContain("Failed to load network topologies");
    expect(wrapper.getComponent(RunaiSelect).props("options")).toContainEqual({
      value: "topology-1",
      label: "Saved Fabric",
    });
    await wrapper.get('[aid="network-topologies-retry"]').trigger("click");
    await flushPromises();
    expect(wrapper.find('[role="alert"]').exists()).toBe(false);
    expect(wrapper.emitted("loaded")).toEqual([[[topology]]]);
    expect(wrapper.emitted("select")).toBeUndefined();
  });

  it("ignores an older cluster response after the selected cluster changes", async () => {
    const first = deferred<NetworkTopology[]>();
    vi.mocked(clusterService.getNetworkTopologies).mockReturnValueOnce(first.promise);
    mountComponent();
    const current = { ...topology, id: "topology-2", clusterId: "cluster-2", name: "Fabric B" };
    vi.mocked(clusterService.getNetworkTopologies).mockResolvedValue([current]);
    await wrapper.setProps({ clusterId: "cluster-2" });
    await flushPromises();
    first.resolve([topology]);
    await flushPromises();
    expect(wrapper.getComponent(RunaiSelect).props("options")).toEqual([{ value: "topology-2", label: "Fabric B" }]);
    expect(wrapper.emitted("loaded")).toEqual([[[current]]]);
  });

  it("does not emit a late load result after the component is removed", async () => {
    const pending = deferred<NetworkTopology[]>();
    vi.mocked(clusterService.getNetworkTopologies).mockReturnValueOnce(pending.promise);
    mountComponent();
    const instance = wrapper.vm;
    wrapper.unmount();
    pending.resolve([topology]);
    await flushPromises();
    expect(instance.networkTopologies).toEqual([]);
    expect(wrapper.emitted("loaded")).toBeUndefined();
  });

  it.each([{ disable: true }, { canCreate: false }])(
    "prevents topology creation when access is restricted: %j",
    async (props) => {
      mountComponent(props);
      await flushPromises();
      const select = wrapper.getComponent(RunaiSelect);
      expect(select.props("customAction")).toBeUndefined();
      select.vm.$emit("custom-action");
      await wrapper.vm.createNetworkTopology(request);
      expect(wrapper.getComponent(RunaiBaseDialog).props("modelValue")).toBe(false);
      expect(clusterService.createNetworkTopology).not.toHaveBeenCalled();
    },
  );

  it("blocks selection changes while read-only", async () => {
    mountComponent({ disable: true, networkTopologyId: "saved" });
    await flushPromises();
    expect(wrapper.getComponent(RunaiSelect).props("disable")).toBe(true);
    wrapper.getComponent(RunaiSelect).vm.$emit("update:model-value", null);
    expect(wrapper.emitted("select")).toBeUndefined();
  });

  it("shows a clear control only for a selection and clears through the actual control", async () => {
    mountComponent();
    await flushPromises();
    expect(wrapper.find('[aria-label="Clear"]').exists()).toBe(false);
    await wrapper.setProps({ networkTopologyId: "topology-1" });
    expect(wrapper.text()).toContain("Fabric A");
    await wrapper.get('[aria-label="Clear"]').trigger("click");
    expect(wrapper.emitted("select")).toEqual([[null]]);
    await wrapper.setProps({ networkTopologyId: null });
    expect(wrapper.find('[aria-label="Clear"]').exists()).toBe(false);
  });

  it("allows one creation request and keeps the dialog open while it is pending", async () => {
    const pending = deferred<NetworkTopology>();
    vi.mocked(clusterService.createNetworkTopology).mockReturnValue(pending.promise);
    mountComponent();
    await flushPromises();
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    const first = wrapper.vm.createNetworkTopology(request);
    await wrapper.vm.createNetworkTopology(request);
    await wrapper.vm.$nextTick();
    wrapper.getComponent(RunaiBaseDialog).vm.$emit("close");
    await wrapper.vm.$nextTick();
    expect(clusterService.createNetworkTopology).toHaveBeenCalledTimes(1);
    expect(wrapper.getComponent(RunaiBaseDialog).props("modelValue")).toBe(true);
    expect(wrapper.getComponent(RunaiBaseDialog).props("isModalPersistent")).toBe(true);
    pending.resolve({ ...topology, id: "created" });
    await first;
    await flushPromises();
    expect(wrapper.emitted("select")).toEqual([["created"]]);
    expect(wrapper.getComponent(RunaiBaseDialog).props("modelValue")).toBe(false);
  });

  it("explains a stale cluster submission and lets the user retry for the current cluster", async () => {
    mountComponent();
    await flushPromises();
    const notify = vi.spyOn(wrapper.vm.$q, "notify").mockImplementation(() => () => {});
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    await wrapper.vm.createNetworkTopology({ ...request, clusterId: "old-cluster" });

    expect(clusterService.createNetworkTopology).not.toHaveBeenCalled();
    expect(notify).toHaveBeenCalledWith(expect.objectContaining({ message: expect.stringMatching(/cluster.*save/i) }));
    expect(wrapper.getComponent(RunaiBaseDialog).props("modelValue")).toBe(true);
    await wrapper.vm.createNetworkTopology(request);
    expect(wrapper.emitted("select")).toEqual([["topology-1"]]);
  });

  it.each(["resolve", "reject"])("reports pending creation immediately and releases it after %s", async (outcome) => {
    const pending = deferred<NetworkTopology>();
    vi.mocked(clusterService.createNetworkTopology).mockReturnValue(pending.promise);
    mountComponent();
    await flushPromises();
    expect(wrapper.emitted("creating")).toBeUndefined();
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    const creating = wrapper.vm.createNetworkTopology(request);
    expect(wrapper.emitted("creating")).toEqual([[true]]);
    if (outcome === "resolve") pending.resolve({ ...topology, id: "created" });
    else pending.reject(new Error("unavailable"));
    await creating;
    expect(wrapper.emitted("creating")).toEqual([[true], [false]]);
  });

  it("keeps the creation dialog and saved selection after a failed creation", async () => {
    vi.mocked(clusterService.createNetworkTopology).mockRejectedValue(new Error("unavailable"));
    mountComponent({ networkTopologyId: "topology-1" });
    await flushPromises();
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    await wrapper.vm.createNetworkTopology(request);
    await flushPromises();
    expect(wrapper.getComponent(RunaiBaseDialog).props("modelValue")).toBe(true);
    expect(wrapper.emitted("select")).toBeUndefined();
  });

  it("does not emit a completed creation after the component is removed", async () => {
    const pending = deferred<NetworkTopology>();
    vi.mocked(clusterService.createNetworkTopology).mockReturnValue(pending.promise);
    const onCreating = vi.fn();
    mountComponent({ onCreating });
    await flushPromises();
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    const instance = wrapper.vm;
    const creating = instance.createNetworkTopology(request);
    expect(onCreating.mock.calls).toEqual([[true]]);
    wrapper.unmount();
    expect(onCreating.mock.calls).toEqual([[true], [false]]);
    pending.resolve({ ...topology, id: "created" });
    await creating;
    expect(instance.networkTopologies).toEqual([topology]);
    expect(wrapper.emitted("select")).toBeUndefined();
  });

  it("ignores a creation response for the cluster the user has left", async () => {
    const pending = deferred<NetworkTopology>();
    vi.mocked(clusterService.createNetworkTopology).mockReturnValue(pending.promise);
    mountComponent();
    await flushPromises();
    wrapper.getComponent(RunaiSelect).vm.$emit("custom-action");
    const creating = wrapper.vm.createNetworkTopology(request);
    expect(wrapper.emitted("creating")).toEqual([[true]]);
    await wrapper.setProps({ clusterId: "cluster-2" });
    expect(wrapper.emitted("creating")).toEqual([[true], [false]]);
    pending.resolve({ ...topology, id: "created" });
    await creating;
    await flushPromises();
    expect(wrapper.emitted("select")).toBeUndefined();
    expect(wrapper.getComponent(RunaiBaseDialog).props("modelValue")).toBe(false);
  });

  it("keeps the new creation pending if an earlier visit to the same cluster returns an error", async () => {
    const previous = deferred<NetworkTopology>();
    const current = deferred<NetworkTopology>();
    vi.mocked(clusterService.createNetworkTopology)
      .mockReturnValueOnce(previous.promise)
      .mockReturnValueOnce(current.promise);
    mountComponent();
    await flushPromises();
    const notify = vi.spyOn(wrapper.vm.$q, "notify").mockImplementation(() => () => {});
    const select = wrapper.getComponent(RunaiSelect);
    select.vm.$emit("custom-action");
    const previousSave = wrapper.vm.createNetworkTopology(request);
    await wrapper.setProps({ clusterId: "cluster-2" });
    await wrapper.setProps({ clusterId: "cluster-1" });
    await flushPromises();
    select.vm.$emit("custom-action");
    const currentSave = wrapper.vm.createNetworkTopology(request);

    previous.reject(new Error("earlier creation failed"));
    await previousSave;
    expect(notify).not.toHaveBeenCalled();
    expect(wrapper.emitted("creating")).toEqual([[true], [false], [true]]);
    expect(wrapper.getComponent(RunaiBaseDialog).props("isModalPersistent")).toBe(true);

    current.resolve({ ...topology, id: "current-topology" });
    await currentSave;
    expect(wrapper.emitted("select")).toEqual([["current-topology"]]);
    expect(wrapper.emitted("creating")?.at(-1)).toEqual([false]);
  });
});
