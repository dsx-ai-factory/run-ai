/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, vi } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { IdleGpuWorkloadWidget } from "@/components/dashboard-v2/widgets/grid-widget/idle-gpu-workload-widget/index";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import type { Workload } from "@/swagger-models/workloads-client";
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import { RunaiTable } from "@/components/common";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
vi.mock("@/services/cluster/workload.service/workload.service");

const MOCK_WORKLOADS = [
  {
    createdAt: "2025-01-21T10:06:02.647837Z",
    name: "runai-idle-gpu-1-1737453880",
    submittedBy: "test@run.ai",
    projectName: "alon",
    allocatedResources: { gpu: 8, gpuMemory: "96G", cpu: 0.8, cpuMemory: "838860800" },
    phase: "Running",
    idleAllocatedGpus: 8,
  },
];
const mockRouter = {
  push: vi.fn(),
};
const { routerPushMock } = vi.hoisted(() => ({ routerPushMock: vi.fn() }));
vi.mock("vue-router", () => ({
  useRoute: vi.fn(),
  useRouter: vi.fn(() => ({
    push: routerPushMock,
  })),
}));
let wrapper: VueWrapper<InstanceType<typeof IdleGpuWorkloadWidget>>;
describe("idle-gpu-workload-widget", () => {
  const mountComponent = (nodePoolName?: string) => {
    wrapper = mount(IdleGpuWorkloadWidget, {
      global: {
        plugins: [Quasar, pinia],
        directives: {
          permission: vi.fn(),
        },
        mocks: {
          $route: {},
          $router: mockRouter,
        },
      },
      props: {
        clusterId: "test-cluster",
        nodePoolName,
      },
    });
  };

  beforeEach(() => {
    routerPushMock.mockClear();
  });

  it("renders idle GPU workloads correctly", async () => {
    mountComponent();
    await flushPromises();
    expect(wrapper.find('[aid="idle-workloads"]').exists()).toBe(true);
  });

  it("displays no data message when there are no idle workloads", async () => {
    vi.mocked(workloadService.getWorkloads).mockResolvedValue([]);
    mountComponent();
    await flushPromises();
    expect(wrapper.findComponent(RunaiTableNoData).exists()).toBe(true);
  });

  it("handles error state correctly", async () => {
    vi.mocked(workloadService.getWorkloads).mockRejectedValue(new Error("Failed to load data"));
    mountComponent();
    await flushPromises();
    expect(wrapper.find(".error-section").exists()).toBe(true);
  });

  it("applies background color and sort to idle gpus column", async () => {
    vi.mocked(workloadService.getWorkloads).mockResolvedValue(MOCK_WORKLOADS as Workload[]);
    mountComponent();
    await flushPromises();
    expect(wrapper.find(".light-blue-primary-opacity-background").exists()).toBe(true);
    expect(wrapper.find(".sort-desc").text()).toBe("Idle GPU devicesarrow_upward");
    expect(wrapper.find("[aid='idleAllocatedGpus']").text()).toBe("8.00");
    expect(wrapper.find("[aid='idleAllocatedGpus']").classes()).includes("light-blue-primary-opacity-background");
  });

  it("should not change sort when clicking on the sort icon", async () => {
    vi.mocked(workloadService.getWorkloads).mockResolvedValue(MOCK_WORKLOADS as Workload[]);
    mountComponent();
    await flushPromises();
    const sortIcon = wrapper.find(".sort-desc");
    sortIcon.trigger("click");
    expect(wrapper.find(".sort-desc").text()).toBe("Idle GPU devicesarrow_upward");
  });

  it("should get idle workloads for a specific node pool", async () => {
    vi.mocked(workloadService.getWorkloads).mockResolvedValue(MOCK_WORKLOADS as Workload[]);
    mountComponent("test-node-pool");
    await flushPromises();
    expect(workloadService.getWorkloads).toHaveBeenCalledWith(false, undefined, 5, "desc", "idleAllocatedGpus", [
      "idleAllocatedGpus>=0.01",
      "clusterId==test-cluster",
      "nodepool==test-node-pool",
    ]);
  });

  it("should clear workloads and show no data when node pool filter returns empty response", async () => {
    vi.mocked(workloadService.getWorkloads).mockResolvedValue(MOCK_WORKLOADS as Workload[]);
    mountComponent();
    await flushPromises();

    // Verify workloads are displayed initially
    expect(wrapper.find("[aid='idleAllocatedGpus']").exists()).toBe(true);
    expect(wrapper.findComponent(RunaiTableNoData).exists()).toBe(false);

    // Update filter to return empty response
    vi.mocked(workloadService.getWorkloads).mockResolvedValue([]);
    await wrapper.setProps({ nodePoolName: "empty-node-pool" });
    await flushPromises();

    // Verify widget shows no data message instead of old workloads
    expect(wrapper.findComponent(RunaiTableNoData).exists()).toBe(true);
  });

  it("links All workloads to idle workloads in the selected node pool", async () => {
    vi.mocked(workloadService.getWorkloads).mockResolvedValue(MOCK_WORKLOADS as Workload[]);
    mountComponent("pool-a");
    await flushPromises();

    wrapper.findComponent(ChartWidgetWrapper).vm.$emit("link-clicked");

    const { name, query } = routerPushMock.mock.calls[0][0];
    expect(name).toBe(WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX);
    expect(JSON.parse(query.advancedFilters)).toEqual([
      expect.objectContaining({ name: "idleAllocatedGpus", operator: ">=", minValue: "0.01" }),
      expect.objectContaining({ name: "nodepool", term: "pool-a", value: "==" }),
    ]);
  });

  it("links a clicked row to that workload only", async () => {
    vi.mocked(workloadService.getWorkloads).mockResolvedValue(MOCK_WORKLOADS as Workload[]);
    mountComponent("pool-a");
    await flushPromises();

    wrapper.findComponent(RunaiTable).vm.$emit("workload-clicked", MOCK_WORKLOADS[0]);

    const { query } = routerPushMock.mock.calls[0][0];
    expect(JSON.parse(query.advancedFilters)).toEqual([
      expect.objectContaining({ name: "name", term: MOCK_WORKLOADS[0].name, value: "==" }),
    ]);
  });
});
