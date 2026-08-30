import { describe, it, expect, beforeEach, vi, afterEach, test } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { ReadyGpuWidget } from "@/components/dashboard-v2/widgets/stat-widgets/ready-gpu-widget/index";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import type { IFilterModel } from "@/models/filter.model";
import { useClusterStore } from "@/stores/cluster.store";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import {
  mockReadyGpuTelemetry,
  mockReadyGpuTelemetryFilterByNodepool,
  mockTotalGpuTelemetry,
  mockTotalGpuTelemetryFilterByNodepool,
} from "@/components/dashboard-v2/widgets/stat-widgets/ready-gpu-widget/ready-gpu-widget.mock";
import { NodeTelemetryType, type TelemetryResponse } from "@/swagger-models/cluster-service-client";
import { filterService } from "@/services/filter.service/filter.service";
import { NODE_ROUTE_NAMES } from "@/router/node.routes/node.routes.names";
import { ETableFilters } from "@/models/table.model";
import { nodeColumnName } from "@/table-models/node.table-model";

vi.mock("@/services/control-plane/cluster.service/cluster.service");

const mockRouter = {
  push: vi.fn(),
};

let wrapper: VueWrapper<InstanceType<typeof ReadyGpuWidget>>;
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterStore = useClusterStore(pinia);
const clusterId = "34234-322-sdfs323";
clusterStore.$patch({
  currentCluster: { uuid: clusterId, name: "demo-cluster" },
});

describe("ReadGpuWidget component", () => {
  const createComponent = (props?: { nodePoolName?: string }) => {
    wrapper = mount(ReadyGpuWidget, {
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

  async function testRedirectToNodes(columnFilters: IFilterModel[]) {
    const setColumnFilterSpy = vi.spyOn(filterService, "setColumnFilter");
    await wrapper.find(".widget-link-btn").trigger("click");
    columnFilters.forEach((columnFilter) => {
      expect(setColumnFilterSpy).toHaveBeenCalledWith(
        expect.anything(),
        columnFilter.term,
        columnFilter.name,
        ETableFilters.NODE,
      );
    });
    expect(setColumnFilterSpy).toHaveBeenCalledTimes(columnFilters.length);
    expect(mockRouter.push).toHaveBeenCalledWith({
      name: NODE_ROUTE_NAMES.NODE_INDEX,
    });
  }
  beforeEach(() => {
    vi.spyOn(clusterService, "getNodeTelemetry").mockImplementation(
      (telemetryType, _clusterId, nodepool): Promise<TelemetryResponse> => {
        if (telemetryType === NodeTelemetryType.ReadyGpus) {
          if (nodepool) {
            return Promise.resolve({ ...mockReadyGpuTelemetryFilterByNodepool, timestamp: new Date().toISOString() });
          }
          return Promise.resolve({ ...mockReadyGpuTelemetry, timestamp: new Date().toISOString() });
        } else {
          if (nodepool) {
            return Promise.resolve({ ...mockTotalGpuTelemetryFilterByNodepool, timestamp: new Date().toISOString() });
          }
          return Promise.resolve({ ...mockTotalGpuTelemetry, timestamp: new Date().toISOString() });
        }
      },
    );
  });

  afterEach(async () => {
    wrapper.unmount();
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  const getStatData = () => {
    return wrapper.find(".stat-data").text();
  };

  const clickOnGridLink = async () => {
    await wrapper.find(".widget-link-btn").trigger("click");
  };

  describe("ready / total", () => {
    beforeEach(async () => {
      vi.mocked(clusterStore.isClusterVersionSufficient).mockImplementation(() => true);
    });

    it("should render ReadyGpuWidget component", async () => {
      createComponent();
      await flushPromises();
      expect(getStatData()).toBe("144/160");
    });

    it("should render ReadyGpuWidget component with filter", async () => {
      createComponent({ nodePoolName: "node-pool-2" });
      await flushPromises();
      expect(getStatData()).toBe("16/150");
    });

    it("should link to nodes table when clicked", async () => {
      createComponent();
      await flushPromises();
      await clickOnGridLink();
      expect(mockRouter.push).toHaveBeenCalledWith({ name: "node-index" });
    });

    it("should trigger get telemetry data on filter change", async () => {
      createComponent();
      await flushPromises();
      await wrapper.setProps({ nodePoolName: "node-pool-2" });
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledTimes(4);
    });

    it("should display loading skeleton when loading", async () => {
      createComponent();
      expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(true);
      await flushPromises();
      expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(false);
    });

    it("should display error message when telemetry fetch fails", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(clusterService.getNodeTelemetry).mockRejectedValue({ response: { status: 500 } });
      createComponent();
      await flushPromises();
      expect(wrapper.find(".error-section").exists()).toEqual(true);
      expect(wrapper.find(".error-section").text()).eql("Data is not available right now");
      errorSpy.mockRestore();
    });

    test("Redirect to workloads with project filter", async () => {
      createComponent({ nodePoolName: "node-pool-2" });
      await flushPromises();
      await testRedirectToNodes([
        {
          name: nodeColumnName.nodePool,
          term: "node-pool-2",
        },
        {
          name: nodeColumnName.status,
          term: "Ready",
        },
      ] as IFilterModel[]);
    });
  });

  describe("backwards compatibility", () => {
    beforeEach(async () => {
      vi.mocked(clusterStore.isClusterVersionSufficient).mockImplementation(() => false);
    });

    it("should get node telemetry data with correct parameters with filter", async () => {
      createComponent({ nodePoolName: "node-pool-2" });
      await flushPromises();
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledTimes(2);
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledWith("TOTAL_GPUS", clusterId, "node-pool-2", []);
      expect(getStatData()).toBe("16/150");
    });

    it("should render correct title without filter", async () => {
      createComponent();
      await flushPromises();
      expect(wrapper.find(".title").text()).toEqual("Ready / Total GPU devices");
      expect(getStatData()).toBe("144/160");
    });

    it("should redirect without filter", async () => {
      createComponent();
      await flushPromises();
      const removeColumnFilterSpy = vi.spyOn(wrapper.vm, "removeNodePoolFilter");
      await wrapper.find(".widget-link-btn").trigger("click");
      expect(mockRouter.push).toHaveBeenCalledWith({ name: "node-index" });
      expect(removeColumnFilterSpy).toHaveBeenCalledTimes(1);
      expect(getStatData()).toBe("144/160");
    });

    it("should redirect with node pool filter ", async () => {
      createComponent({ nodePoolName: "node-pool-2" });
      await flushPromises();
      const setNodePoolColumnFilterSpy = vi.spyOn(wrapper.vm, "setNodePoolColumnFilter");
      await wrapper.find(".widget-link-btn").trigger("click");
      expect(mockRouter.push).toHaveBeenCalledWith({ name: "node-index" });
      expect(setNodePoolColumnFilterSpy).toHaveBeenCalledTimes(1);
    });
  });
});
