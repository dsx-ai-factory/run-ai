/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { Notify, Quasar } from "quasar";
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
function mountComponent(props = {}) {
  wrapper = mount(NetworkTopologySelection, {
    props: { clusterId: "cluster-1", ...props },
    global: {
      plugins: [[Quasar, { plugins: { Notify } }]],
      stubs: { NetworkTopologyForm: true },
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
  it("reports successfully loaded topology records to the summary owner", async () => {
    mountComponent();
    await flushPromises();
    expect(wrapper.emitted("loaded")).toEqual([[[topology]]]);
  });

  it("distinguishes an empty topology list from a failed request", async () => {
    vi.mocked(clusterService.getNetworkTopologies).mockResolvedValue([]);
    mountComponent();
    await flushPromises();
    expect(wrapper.get('[aid="network-topologies-empty"]').text()).toContain("No network topologies found");
    expect(wrapper.find('[role="alert"]').exists()).toBe(false);
    expect(wrapper.emitted("loaded")).toEqual([[[]]]);
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
    expect(wrapper.find('[aid="network-topologies-empty"]').exists()).toBe(false);
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
});
