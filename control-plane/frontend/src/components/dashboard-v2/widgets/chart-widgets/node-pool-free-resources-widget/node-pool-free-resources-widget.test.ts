import NodePoolFreeGpuChartComponent from "@/components/dashboard-v2/widgets/chart-widgets/node-pool-free-resources-widget/node-pool-free-gpu-chart/node-pool-free-gpu-chart.vue";
import { describe, it, expect, beforeEach, test, vi, type MockedFunction, afterEach } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
//service
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { NodeTelemetryType } from "@/swagger-models/cluster-service-client";
//model
import { ENodeTelemetryGroupBy } from "@/models/cluster.model";
//mock
import {
  mockNodePoolFreeGpusResponse,
  mockNodeTotalGpusResponse,
} from "@/components/dashboard-v2/widgets/chart-widgets/node-pool-free-resources-widget/node-pool-free-resources.mock";
//cmps
import { NodePoolFreeResourcesWidget } from "@/components/dashboard-v2/widgets/chart-widgets/node-pool-free-resources-widget/index";
import { useClusterStore } from "@/stores/cluster.store";
vi.mock("@/services/control-plane/cluster.service/cluster.service");

const mockRouter = {
  push: vi.fn(),
};
const clusterId = "34234-322-sdfs323";
const mockClusters = [{ name: "demo-cluster", uuid: clusterId }];

const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
let wrapper: VueWrapper<InstanceType<typeof NodePoolFreeResourcesWidget>>;
const clusterStore = useClusterStore(pinia);

clusterStore.$patch({
  currentCluster: { uuid: clusterId, name: "demo-cluster" },
  clusters: mockClusters,
});
describe("NodePoolFreeResourcesWidget component", () => {
  const createComponent = (props: { nodePoolName?: string; clusterId: string } = { clusterId }) => {
    wrapper = mount(NodePoolFreeResourcesWidget, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {},
          $router: mockRouter,
        },
      },
      props,
    });

    wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  };

  const mockGetTotalGpusDevices = clusterService.getNodeTelemetry as MockedFunction<
    typeof clusterService.getNodeTelemetry
  >;

  const mockGetFreeGpusByNodePool = clusterService.getNodeTelemetry as MockedFunction<
    typeof clusterService.getNodeTelemetry
  >;
  const getNodeTelemetrySpy = vi.spyOn(clusterService, "getNodeTelemetry");
  beforeEach(async () => {
    mockGetFreeGpusByNodePool.mockResolvedValue(mockNodePoolFreeGpusResponse);
    mockGetTotalGpusDevices.mockResolvedValue(mockNodeTotalGpusResponse);
    createComponent();
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
  });

  it("should render correct title", async () => {
    expect(wrapper.find("[aid='free-nodes-by-node-pool']").text()).toEqual("Free nodes by node pool");
  });

  it("renders without errors and render loading on component created", () => {
    expect(wrapper.exists()).toEqual(true);
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(true);
  });

  it("initializes with correct components", async () => {
    expect(wrapper.vm.isScatterChart).toEqual(true);
    expect(wrapper.vm.breadcrumbs).toEqual(["demo-cluster"]);
    await flushPromises();

    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(false);
    expect(wrapper.findComponent(".node-pool-free-gpu-chart").exists()).toEqual(true);
  });

  it("keeps the scatter plot close to its legend", async () => {
    await flushPromises();

    const scatterChart = wrapper.findComponent(NodePoolFreeGpuChartComponent);
    expect(scatterChart.exists()).toEqual(true);
    const grid = Array.isArray(scatterChart.vm.scatterChartOptions.grid)
      ? scatterChart.vm.scatterChartOptions.grid[0]
      : scatterChart.vm.scatterChartOptions.grid;
    expect(grid?.bottom).toEqual(55);
  });

  it("anchors a single zero-only node pool at the left baseline without drawing the baseline over its marker", () => {
    const singlePoolChart = mount(NodePoolFreeGpuChartComponent, {
      global: { plugins: [Quasar] },
      props: {
        nodePoolFreeGpusTelemetry: [
          {
            value: "0",
            groups: [
              { key: ENodeTelemetryGroupBy.Nodepool, value: "default" },
              { key: ENodeTelemetryGroupBy.Node, value: "runai-cluster-worker" },
            ],
          },
        ],
      },
    });

    const options = singlePoolChart.vm.scatterChartOptions;
    const xAxis = Array.isArray(options.xAxis) ? options.xAxis[0] : options.xAxis;
    const series = Array.isArray(options.series) ? options.series[0] : options.series;

    expect(xAxis).toMatchObject({ min: 0, max: 1 });
    expect(series).toMatchObject({
      z: 3,
      data: [[0, 0]],
      markLine: { z: 1, data: [{ yAxis: 0 }] },
    });
    singlePoolChart.unmount();
  });

  it("preserves half-step axis padding when multiple node pools are shown", () => {
    const multiPoolChart = mount(NodePoolFreeGpuChartComponent, {
      global: { plugins: [Quasar] },
      props: {
        nodePoolFreeGpusTelemetry: [
          {
            value: "2",
            groups: [
              { key: ENodeTelemetryGroupBy.Nodepool, value: "alpha" },
              { key: ENodeTelemetryGroupBy.Node, value: "worker-alpha" },
            ],
          },
          {
            value: "4",
            groups: [
              { key: ENodeTelemetryGroupBy.Nodepool, value: "beta" },
              { key: ENodeTelemetryGroupBy.Node, value: "worker-beta" },
            ],
          },
        ],
      },
    });

    const options = multiPoolChart.vm.scatterChartOptions;
    const xAxis = Array.isArray(options.xAxis) ? options.xAxis[0] : options.xAxis;

    expect(xAxis).toMatchObject({ min: -0.5, max: 1.5 });
    multiPoolChart.unmount();
  });

  test("fetches total GPU devices data", async () => {
    await flushPromises();
    expect(getNodeTelemetrySpy).toHaveBeenCalledWith(NodeTelemetryType.TotalGpus, wrapper.vm.clusterId, undefined, [
      ENodeTelemetryGroupBy.Node,
    ]);
    expect(wrapper.vm.nodeTotalGpuDevicesTelemetry).toEqual(mockNodeTotalGpusResponse.values);
    expect(wrapper.vm.totalGpusError).toEqual(false);
    expect(wrapper.vm.displayLoading).toEqual(false);
  });

  test("fetches FreeGpus devices data", async () => {
    await flushPromises();
    expect(getNodeTelemetrySpy).toHaveBeenCalledWith(NodeTelemetryType.FreeGpus, wrapper.vm.clusterId, undefined, [
      ENodeTelemetryGroupBy.Nodepool,
      ENodeTelemetryGroupBy.Node,
    ]);
    expect(wrapper.vm.freeGpusError).toEqual(false);
    expect(wrapper.vm.displayLoading).toEqual(false);
  });

  it("drills up correctly", async () => {
    wrapper.vm.breadcrumbs = ["Node pools", "TestNodePool"];
    wrapper.vm.drillUp(0);
    expect(wrapper.vm.breadcrumbs).toEqual(["Node pools"]);
    expect(wrapper.vm.isScatterChart).toEqual(true);
  });

  it("should redirect", async () => {
    await wrapper.find(".widget-link-btn").trigger("click");
    expect(mockRouter.push).toHaveBeenCalledWith({ name: "node-index" });
  });
});
