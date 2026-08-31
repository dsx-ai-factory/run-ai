/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, vi } from "vitest";
import { flushPromises, mount, type VueWrapper } from "@vue/test-utils";
import { IdleGpuWorkloadWidget } from "@/components/dashboard-v2/widgets/grid-widget/idle-gpu-workload-widget/index";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import type { Workload } from "@/swagger-models/workloads-client";
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
vi.mock("vue-router", () => ({
  useRoute: vi.fn(),
  useRouter: vi.fn(() => ({
    push: vi.fn(),
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
      "idleAllocatedGpus>=1",
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
});
