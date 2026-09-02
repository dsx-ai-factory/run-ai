/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { IdleAllocatedGpuWidget } from "@/components/dashboard-v2/widgets/stat-widgets/idle-allocated-gpu-widget/index";
import { idleAllocatedGpuWidgetMock } from "@/components/dashboard-v2/widgets/stat-widgets/idle-allocated-gpu-widget/idle-allocated-gpu-widget.mock";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { NodeTelemetryType } from "@/swagger-models/cluster-service-client";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { usePermissionStore } from "@/stores/permissions.store";
import { EWorkloadTelemetryGroupBy } from "@/models/workload.model";

vi.mock("@/services/control-plane/cluster.service/cluster.service");
vi.mock("@/services/cluster/workload.service/workload.service");

const mockRouter = {
  push: vi.fn(),
};

let wrapper: VueWrapper<InstanceType<typeof IdleAllocatedGpuWidget>>;
let pinia: TestingPinia;
const clusterId = "34234-322-sdfs323";

describe("IdleAllocatedGpuWidget component", () => {
  const factory = (props: { nodePoolName?: string; clusterId: string }, hasReadNodesPermission = true) => {
    pinia = createTestingPinia({ createSpy: vi.fn });
    const permissionStore = usePermissionStore(pinia);
    vi.mocked(permissionStore.hasReadNodesPermission).mockReturnValue(hasReadNodesPermission);

    wrapper = mount(IdleAllocatedGpuWidget, {
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

  beforeEach(() => {
    vi.mocked(clusterService.getNodeTelemetry).mockImplementation(() => Promise.resolve(idleAllocatedGpuWidgetMock));
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
      Promise.resolve(idleAllocatedGpuWidgetMock),
    );
  });

  afterEach(async () => {
    vi.clearAllMocks();
  });

  const getStatData = () => {
    return wrapper.find(".stat-data").text();
  };

  const clickOnGridLink = async () => {
    await wrapper.find(".widget-link-btn").trigger("click");
  };

  describe("with READ Nodes permission", () => {
    it("should render widget using node telemetry", async () => {
      factory({ clusterId }, true);
      await flushPromises();
      expect(getStatData()).toBe("2");
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledWith(
        NodeTelemetryType.IdleAllocatedGpus,
        clusterId,
        undefined,
        [],
      );
      expect(workloadService.getWorkloadsTelemetry).not.toHaveBeenCalled();
    });

    it("should trigger get telemetry data on node pool filter change", async () => {
      factory({ clusterId }, true);
      await flushPromises();
      const nodePoolName = "node-pool-2";
      await wrapper.setProps({ nodePoolName });
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledTimes(2);
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledWith(
        NodeTelemetryType.IdleAllocatedGpus,
        clusterId,
        nodePoolName,
        [],
      );
    });

    it("should trigger get telemetry data on cluster filter change", async () => {
      factory({ clusterId }, true);
      await flushPromises();
      const newClusterId = "sda1235sd2asd";
      await wrapper.setProps({ clusterId: newClusterId });
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledTimes(2);
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledWith(
        NodeTelemetryType.IdleAllocatedGpus,
        newClusterId,
        undefined,
        [],
      );
    });
  });

  describe("without READ Nodes permission", () => {
    it("should render widget using workload telemetry as fallback", async () => {
      factory({ clusterId }, false);
      await flushPromises();
      expect(getStatData()).toBe("2");
      expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledWith(
        "IDLE_ALLOCATED_GPUS",
        clusterId,
        undefined,
        undefined,
        undefined,
        [EWorkloadTelemetryGroupBy.ClusterId],
      );
      expect(clusterService.getNodeTelemetry).not.toHaveBeenCalled();
    });

    it("should trigger get workload telemetry data on node pool filter change", async () => {
      factory({ clusterId }, false);
      await flushPromises();
      const nodePoolName = "node-pool-2";
      await wrapper.setProps({ nodePoolName });
      expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledTimes(2);
      expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledWith(
        "IDLE_ALLOCATED_GPUS",
        clusterId,
        nodePoolName,
        undefined,
        undefined,
        [EWorkloadTelemetryGroupBy.ClusterId],
      );
    });
  });

  describe("html", () => {
    it("should link to workloads table when clicked", async () => {
      factory({ clusterId });
      await flushPromises();
      await clickOnGridLink();
      expect(mockRouter.push).toHaveBeenCalledWith({
        name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
        query: { sortBy: "idleAllocatedGpus", descending: "true" },
      });
    });

    it("should display loading skeleton when loading", async () => {
      factory({ clusterId });
      expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(true);
      await flushPromises();
      expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(false);
    });

    it("should display error message when telemetry fetch fails", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(clusterService.getNodeTelemetry).mockRejectedValueOnce({ response: { status: 500 } });
      factory({ clusterId });
      await flushPromises();
      expect(wrapper.find(".error-section").exists()).toEqual(true);
      expect(wrapper.find(".error-section").text()).eql("Data is not available right now");
      errorSpy.mockRestore();
    });
  });
});
