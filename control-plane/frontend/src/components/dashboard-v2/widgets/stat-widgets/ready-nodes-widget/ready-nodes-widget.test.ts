import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { useClusterStore } from "@/stores/cluster.store";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { ReadyNodesWidget } from "@/components/dashboard-v2/widgets/stat-widgets/ready-nodes-widget/index";
import {
  mockReadyNodesTelemetry,
  mockReadyNodesTelemetryFilterByNodepool,
  mockTotalNodesTelemetry,
  mockTotalNodesTelemetryFilterByNodepool,
} from "@/components/dashboard-v2/widgets/stat-widgets/ready-nodes-widget/ready-nodes-widget.mock";
import { NodeTelemetryType } from "@/swagger-models/cluster-service-client";

vi.mock("@/services/control-plane/cluster.service/cluster.service");

const mockRouter = {
  push: vi.fn(),
};

let wrapper: VueWrapper<InstanceType<typeof ReadyNodesWidget>>;
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterStore = useClusterStore(pinia);
const clusterId = "34234-322-sdfs323";

describe("ReadNodesWidget component", () => {
  const createComponent = (props?: { nodePoolName?: string }) => {
    wrapper = mount(ReadyNodesWidget, {
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
    vi.spyOn(clusterService, "getNodeTelemetry").mockImplementation((telemetryType, _clusterId, nodepool) => {
      if (telemetryType === NodeTelemetryType.ReadyGpuNodes) {
        if (nodepool) {
          return Promise.resolve(mockReadyNodesTelemetryFilterByNodepool);
        }
        return Promise.resolve(mockReadyNodesTelemetry);
      } else {
        if (nodepool) {
          return Promise.resolve(mockTotalNodesTelemetryFilterByNodepool);
        }
        return Promise.resolve(mockTotalNodesTelemetry);
      }
    });
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
  describe("ready / total", () => {
    beforeEach(async () => {
      vi.mocked(clusterStore.isClusterVersionSufficient).mockImplementation(() => true);
    });
    afterEach(async () => {
      vi.clearAllMocks();
    });
    it("should render ReadNodesWidget component", async () => {
      createComponent();
      await flushPromises();
      expect(getStatData()).toBe("144/160");
    });

    it("should render ReadNodesWidget component with filter", async () => {
      createComponent({ nodePoolName: "Node pool" });
      await flushPromises();
      expect(getStatData()).toBe("16/150");
    });

    it("should link to nodes table when clicked", async () => {
      vi.mocked(clusterService.getNodeTelemetry).mockImplementation(() => Promise.resolve(mockReadyNodesTelemetry));
      createComponent();
      await flushPromises();
      await clickOnGridLink();
      expect(mockRouter.push).toHaveBeenCalledWith({ name: "node-index" });
    });

    it("should trigger get telemetry data on filter change", async () => {
      vi.mocked(clusterService.getNodeTelemetry).mockImplementation(() => Promise.resolve(mockReadyNodesTelemetry));
      createComponent();
      await flushPromises();
      await wrapper.setProps({ nodePoolName: "Node pool" });
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledTimes(4);
    });

    it("should display loading skeleton when loading", async () => {
      vi.mocked(clusterService.getNodeTelemetry).mockImplementation(() => Promise.resolve(mockReadyNodesTelemetry));
      createComponent();
      expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(true);
      await flushPromises();
      expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(false);
    });

    it("should display error message when telemetry fetch fails", async () => {
      const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
      vi.mocked(clusterService.getNodeTelemetry).mockRejectedValueOnce({ response: { status: 500 } });
      createComponent();
      await flushPromises();
      expect(wrapper.find(".error-section").exists()).toEqual(true);
      expect(wrapper.find(".error-section").text()).eql("Data is not available right now");
      errorSpy.mockRestore();
    });
  });

  describe("backwards compatibility", () => {
    beforeEach(async () => {
      vi.mocked(clusterStore.isClusterVersionSufficient).mockImplementation(() => false);
    });

    it("should get node telemetry data with correct parameters with filter", async () => {
      createComponent({ nodePoolName: "node-pool-2" });
      await flushPromises();
      expect(clusterService.getNodeTelemetry).toHaveBeenCalledWith("TOTAL_GPU_NODES", clusterId, "node-pool-2");
      expect(getStatData()).toBe("16/150");
    });

    it("should render correct title without filter", async () => {
      createComponent();
      await flushPromises();

      expect(wrapper.find(".title").text()).toEqual("Ready / Total GPU nodes");
      expect(getStatData()).toBe("144/160");
    });

    it("should redirect without filter", async () => {
      createComponent();
      await flushPromises();
      const setColumnFilterSpy = vi.spyOn(wrapper.vm, "setColumnFilter");
      await wrapper.find(".widget-link-btn").trigger("click");
      expect(mockRouter.push).toHaveBeenCalledWith({ name: "node-index" });
      expect(setColumnFilterSpy).toHaveBeenCalledTimes(1);
      expect(getStatData()).toBe("144/160");
    });
  });
});
