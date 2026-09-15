/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, type Mock } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import { Quasar } from "quasar";
import ElementDetailsTab from "./element-details-tab.vue";
import DetailsPanelView from "./details-panel-view.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { WorkloadElementResponse } from "@/swagger-models/workloads-client";

vi.mock("@/services/control-plane/cluster.service/cluster.service", () => ({
  clusterService: {
    getNetworkTopologyLevelsByName: vi.fn().mockResolvedValue([]),
  },
}));

vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ push: vi.fn() }),
}));

describe("ElementDetailsTab", () => {
  const mockElement: WorkloadElementResponse = {
    id: "element-1",
    name: "prefill-worker",
    hierarchyLevel: 1,
    createdAt: "2026-01-13T10:32:00Z",
    topology: { name: "my-gb200-topology" },
    kubernetesRef: { name: "prefill-worker", gvk: { group: "apps", version: "v1", kind: "Deployment" } },
    podsRequestedResources: null,
    allocatedResources: null,
  };

  // shallowMount auto-stubs children by object identity, which works regardless of whether the
  // child declares its own `name` — unlike string-keyed `stubs`, which only intercepts <script
  // setup> children that declare a matching defineOptions({ name }).
  const createWrapper = (props: Record<string, unknown> = {}) =>
    shallowMount(ElementDetailsTab, {
      props: { element: mockElement, clusterId: "cluster-123", ...props },
      global: {
        plugins: [[Quasar, { plugins: {} }]],
      },
    });

  it("feeds an empty sections model to details-panel-view when there is no element", async () => {
    const wrapper = createWrapper({ element: null });
    await flushPromises();

    // The empty-state UI itself lives in details-panel-view (tested there) — here we only need to
    // confirm this tab feeds it an empty sections model when there's nothing to show.
    expect(wrapper.vm.accordionSections).toEqual([]);
    expect(wrapper.findComponent(DetailsPanelView).props("sections")).toEqual([]);
  });

  it("renders details-panel-view with the built sections when an element is given", async () => {
    const wrapper = createWrapper({ elementPodsCount: 4 });
    await flushPromises();

    const view = wrapper.findComponent(DetailsPanelView);
    expect(view.exists()).toBe(true);
    expect(view.props("sections")).toEqual(wrapper.vm.accordionSections);
    expect(wrapper.vm.accordionSections.map((s) => s.id)).toEqual(["general", "network", "compute"]);
  });

  it("fetches topology levels scoped to the element's topology name and feeds them into the network section", async () => {
    (clusterService.getNetworkTopologyLevelsByName as Mock).mockResolvedValue(["region", "zone"]);

    const wrapper = createWrapper();
    await flushPromises();

    expect(clusterService.getNetworkTopologyLevelsByName).toHaveBeenCalledWith("my-gb200-topology", "cluster-123");
    const network = wrapper.vm.accordionSections.find((s) => s.id === "network");
    expect((network?.data as { topologyLevels: string[] }).topologyLevels).toEqual(["region", "zone"]);
  });

  it("re-fetches when the selected element changes", async () => {
    const wrapper = createWrapper();
    await flushPromises();
    (clusterService.getNetworkTopologyLevelsByName as Mock).mockClear();

    await wrapper.setProps({ element: { ...mockElement, id: "element-2", topology: { name: "other-topology" } } });
    await flushPromises();

    expect(clusterService.getNetworkTopologyLevelsByName).toHaveBeenCalledWith("other-topology", "cluster-123");
  });

  it("re-fetches when only the clusterId changes (same element)", async () => {
    const wrapper = createWrapper();
    await flushPromises();
    (clusterService.getNetworkTopologyLevelsByName as Mock).mockClear();

    await wrapper.setProps({ clusterId: "cluster-456" });
    await flushPromises();

    expect(clusterService.getNetworkTopologyLevelsByName).toHaveBeenCalledWith("my-gb200-topology", "cluster-456");
  });

  it("discards a superseded topology response so it cannot overwrite the current element's levels", async () => {
    let resolveFirst: (levels: string[]) => void = () => {};
    (clusterService.getNetworkTopologyLevelsByName as Mock)
      .mockImplementationOnce(() => new Promise((resolve) => (resolveFirst = resolve)))
      .mockResolvedValueOnce(["rack", "host"]);

    const wrapper = createWrapper();
    await flushPromises();

    await wrapper.setProps({ element: { ...mockElement, id: "element-2", topology: { name: "other-topology" } } });
    await flushPromises();

    // The stale first request resolves after the second one has already completed.
    resolveFirst(["region", "zone"]);
    await flushPromises();

    const network = wrapper.vm.accordionSections.find((s) => s.id === "network");
    expect((network?.data as { topologyLevels: string[] }).topologyLevels).toEqual(["rack", "host"]);
  });

  it("feeds elementPodsCount into the Compute section", async () => {
    const wrapper = createWrapper({ elementPodsCount: 4 });
    await flushPromises();

    const compute = wrapper.vm.accordionSections.find((s) => s.id === "compute");
    expect((compute?.data as { items: { label: string; value: string }[] }).items).toEqual(
      expect.arrayContaining([{ label: "Pods", value: "4" }]),
    );
  });
});
