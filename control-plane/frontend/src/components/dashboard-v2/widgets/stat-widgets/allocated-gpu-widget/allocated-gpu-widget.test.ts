/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { AllocatedGpuWidget } from "@/components/dashboard-v2/widgets/stat-widgets/allocated-gpu-widget/index";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { usePermissionStore } from "@/stores/permissions.store";
import {
  AllocatedGpuWidgetTelemetryMock,
  mockAllocatedGpuTelemetryFilterByNodepool,
} from "@/components/dashboard-v2/widgets/stat-widgets/allocated-gpu-widget/allocated-gpu-widget.mock";

vi.mock("@/services/cluster/workload.service/workload.service");
vi.mock("@/services/control-plane/cluster.service/cluster.service");

const mockRouter = {
  push: vi.fn(),
};

let wrapper: VueWrapper<InstanceType<typeof AllocatedGpuWidget>>;
let pinia: TestingPinia;
const clusterId = "34234-322-sdfs323";

describe("ReadGpuWidget component", () => {
  const createComponent = (props?: { nodePoolName?: string }, hasReadNodesPermission = false) => {
    pinia = createTestingPinia({ createSpy: vi.fn });
    const permissionStore = usePermissionStore(pinia);
    vi.mocked(permissionStore.hasReadNodesPermission).mockReturnValue(hasReadNodesPermission);

    wrapper = mount(AllocatedGpuWidget, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {},
          $router: mockRouter,
        },
      },
      props: {
        ...props,
        clusterId,
      },
    });

    wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  };

  beforeEach(async () => {
    vi.clearAllMocks();
  });

  const getStatData = () => {
    return wrapper.find(".stat-data").text();
  };

  const clickOnGridLink = async () => {
    await wrapper.find(".widget-link-btn").trigger("click");
  };

  it("should render AllocatedGpuWidget component", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
      Promise.resolve(AllocatedGpuWidgetTelemetryMock),
    );
    createComponent();
    await flushPromises();
    expect(getStatData()).toBe("89");
  });

  it("should render AllocatedGpuWidget component with filter", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
      Promise.resolve(mockAllocatedGpuTelemetryFilterByNodepool),
    );
    createComponent({ nodePoolName: "node-pool-2" });
    await flushPromises();
    expect(getStatData()).toBe("2");
  });

  describe("permission-based redirect", () => {
    it("should link to nodes table when user has READ Nodes permission", async () => {
      vi.mocked(clusterService.getNodeTelemetry).mockImplementation(() =>
        Promise.resolve(AllocatedGpuWidgetTelemetryMock),
      );
      createComponent(undefined, true); // hasReadNodesPermission = true
      await flushPromises();
      await clickOnGridLink();
      expect(mockRouter.push).toHaveBeenCalledWith({ name: "node-index" });
    });

    it("should link to workloads table when user lacks READ Nodes permission", async () => {
      vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
        Promise.resolve(AllocatedGpuWidgetTelemetryMock),
      );
      createComponent(undefined, false); // hasReadNodesPermission = false
      await flushPromises();
      await clickOnGridLink();
      expect(mockRouter.push).toHaveBeenCalledWith({ name: "workload-index" });
    });

    it("should show 'All nodes' link text when user has READ Nodes permission", async () => {
      vi.mocked(clusterService.getNodeTelemetry).mockImplementation(() =>
        Promise.resolve(AllocatedGpuWidgetTelemetryMock),
      );
      createComponent(undefined, true);
      await flushPromises();
      expect(wrapper.find(".widget-link-btn").text()).toBe("All nodes");
    });

    it("should show 'All workloads' link text when user lacks READ Nodes permission", async () => {
      vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
        Promise.resolve(AllocatedGpuWidgetTelemetryMock),
      );
      createComponent(undefined, false);
      await flushPromises();
      expect(wrapper.find(".widget-link-btn").text()).toBe("All workloads");
    });
  });

  describe("permission-based data source", () => {
    it("should use node telemetry when user has READ Nodes permission", async () => {
      vi.mocked(clusterService.getNodeTelemetry).mockImplementation(() =>
        Promise.resolve(AllocatedGpuWidgetTelemetryMock),
      );
      createComponent(undefined, true);
      await flushPromises();
      expect(clusterService.getNodeTelemetry).toHaveBeenCalled();
      expect(workloadService.getWorkloadsTelemetry).not.toHaveBeenCalled();
    });

    it("should use workloads telemetry when user lacks READ Nodes permission", async () => {
      vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
        Promise.resolve(AllocatedGpuWidgetTelemetryMock),
      );
      createComponent(undefined, false);
      await flushPromises();
      expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalled();
      expect(clusterService.getNodeTelemetry).not.toHaveBeenCalled();
    });
  });

  it("should trigger get telemetry data on filter change", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
      Promise.resolve(AllocatedGpuWidgetTelemetryMock),
    );
    createComponent();
    await flushPromises();
    await wrapper.setProps({ nodePoolName: "node-pool-2" });
    expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledTimes(2);
  });

  it("should display loading skeleton when loading", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
      Promise.resolve(AllocatedGpuWidgetTelemetryMock),
    );
    createComponent();
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(true);
    await flushPromises();
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(false);
  });

  it("should display error message when telemetry fetch fails", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(workloadService.getWorkloadsTelemetry).mockRejectedValueOnce({ response: { status: 500 } });
    createComponent();
    await flushPromises();
    expect(wrapper.find(".error-section").exists()).toEqual(true);
    expect(wrapper.find(".error-section").text()).eql("Data is not available right now");
    errorSpy.mockRestore();
  });
});
