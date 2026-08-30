import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { WorkloadByTypeWidget } from "./index";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { WorkloadByTypeWidgetMock, WorkloadByTypeWidgetMockFilterByNodepool } from "./workload-by-type-widget.mock";
import { Phase, WorkloadTelemetryType } from "@/swagger-models/workloads-client";
import { EWorkloadTelemetryGroupBy, type EWorkloadTelemetryPointMeta } from "@/models/workload.model";
import type { EChartsOption, BarSeriesOption } from "echarts";
interface SeriesOptionsType {
  name?: string;
  data?: unknown[];
  color?: string;
}
import { chartColorPaletteArray, UNKNOWN_COLOR } from "@/models/chart.model";

interface WorkloadByTypeWidgetInstance {
  assignColorsToSeries: (series: SeriesOptionsType[]) => SeriesOptionsType[];
  getColor: (type: string) => string;
  convertSeriesToECharts: (
    series: SeriesOptionsType[],
    isDrilldown?: boolean,
  ) => { categories: string[]; series: BarSeriesOption[] };
  buildChartOption: (categories: string[], series: BarSeriesOption[], isDrilldown?: boolean) => EChartsOption;
  drillDownTo: (meta: EWorkloadTelemetryPointMeta) => Promise<void>;
  getDrillDownSeries: (meta: EWorkloadTelemetryPointMeta) => Promise<SeriesOptionsType[]>;
  chartOption: EChartsOption | null;
  drillBreadcrumbs: string[];
  drillStack: unknown[];
  drilldownAxisCategories: string[];
  currentDrillMeta: EWorkloadTelemetryPointMeta | null;
}

vi.mock("@/services/cluster/workload.service/workload.service");

const mockRouter = {
  push: vi.fn(),
};

let wrapper: VueWrapper<InstanceType<typeof WorkloadByTypeWidget>>;
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterId = "34234-322-sdfs323";

describe("WorkloadByTypeWidget component", () => {
  const createComponent = (props?: { nodePoolName?: string }) => {
    wrapper = mount(WorkloadByTypeWidget, {
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
        hasDepartments: true,
      },
    });

    wrapper.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  };

  beforeEach(async () => {
    vi.clearAllMocks();
  });

  const clickOnGridLink = async () => {
    await wrapper.find(".widget-link-btn").trigger("click");
  };

  it("should render WorkloadByTypeWidget component", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadByTypeWidgetMock));
    createComponent();
    await flushPromises();
    expect(wrapper.find('[aid="workload-group-by-select"]').exists()).toBe(true);
  });

  it("uses the 24px overflow-menu icon in its custom chart header", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadByTypeWidgetMock));
    createComponent();
    await flushPromises();

    expect(wrapper.getComponent({ name: "RunaiActionsMenu" }).props("largeIcon")).toBe(true);
  });

  it("exposes a stable guided-tour target for drilldown labels", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadByTypeWidgetMock));
    createComponent();
    await flushPromises();

    const vm = wrapper.vm as unknown as WorkloadByTypeWidgetInstance;
    vm.drilldownAxisCategories = ["test-cluster"];
    await wrapper.vm.$nextTick();

    const tourTarget = wrapper.find('[aid="workload-by-type-drilldown-label"]');
    expect(tourTarget.element.tagName).toBe("SPAN");
    expect(tourTarget.text()).toBe("test-cluster");
    expect(tourTarget.element.parentElement?.classList.contains("workload-by-type-axis-label")).toBe(true);
  });

  it("should render WorkloadByTypeWidget component with filter", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
      Promise.resolve(WorkloadByTypeWidgetMockFilterByNodepool),
    );
    createComponent({ nodePoolName: "node-pool-2" });
    await flushPromises();
    expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledWith(
      WorkloadTelemetryType.WorkloadsCount,
      clusterId,
      "node-pool-2",
      undefined,
      undefined,
      [EWorkloadTelemetryGroupBy.ClusterId, EWorkloadTelemetryGroupBy.Category],
      [Phase.Running],
    );
    expect(wrapper.find(".widget-link-btn").text()).toBe("All workloads");
  });

  it("should link to projects table when clicked", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadByTypeWidgetMock));
    createComponent();
    await flushPromises();
    await clickOnGridLink();
    expect(mockRouter.push).toHaveBeenCalledWith({ name: "workload-index" });
  });

  it("should trigger get telemetry data on filter change", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadByTypeWidgetMock));
    createComponent();
    await flushPromises();
    await wrapper.setProps({ nodePoolName: "node-pool-2" });
    expect(workloadService.getWorkloadsTelemetry).toHaveBeenCalledTimes(2);
  });

  it("should display loading skeleton when loading", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadByTypeWidgetMock));
    createComponent();
    expect(wrapper.findComponent(".chart-widget-skeleton").exists()).toBe(true);
    await flushPromises();
    expect(wrapper.findComponent(".chart-widget-skeleton").exists()).toBe(false);
  });

  it("should display error message when telemetry fetch fails", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(workloadService.getWorkloadsTelemetry)
      .mockRejectedValueOnce({ response: { status: 500 } })
      .mockResolvedValueOnce(WorkloadByTypeWidgetMock);

    createComponent();
    await flushPromises();
    expect(wrapper.find(".error-section").exists()).toBe(true);
    expect(wrapper.find(".error-section").text()).toBe("Data is not available right now");

    // validate that error message is not displayed after successful fetch
    await wrapper.setProps({ nodePoolName: "node-pool-2" });
    await flushPromises();
    expect(wrapper.find(".error-section").exists()).toBe(false);
    errorSpy.mockRestore();
  });

  it("propagates drilldown failures so the caller preserves the error state", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(workloadService.getWorkloadsTelemetry)
      .mockResolvedValueOnce(WorkloadByTypeWidgetMock)
      .mockRejectedValueOnce(new Error("drilldown failed"));
    createComponent();
    await flushPromises();

    const vm = wrapper.vm as unknown as WorkloadByTypeWidgetInstance;
    await vm.drillDownTo({ entityId: clusterId, groupByKey: EWorkloadTelemetryGroupBy.ClusterId });
    await wrapper.vm.$nextTick();

    expect(wrapper.find(".error-section").exists()).toBe(true);
    expect(vm.drillBreadcrumbs).toEqual([]);
    errorSpy.mockRestore();
  });

  it("should keep breadcrumbs hidden until drilldown data is ready", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadByTypeWidgetMock));
    createComponent();
    await flushPromises();

    const vm = wrapper.vm as unknown as WorkloadByTypeWidgetInstance;
    vm.chartOption = { series: [] };

    let resolveDrilldown: (series: SeriesOptionsType[]) => void = () => undefined;
    vi.spyOn(vm, "getDrillDownSeries").mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveDrilldown = resolve;
        }),
    );

    const meta = { entityId: clusterId, groupByKey: EWorkloadTelemetryGroupBy.ClusterId };
    const drilldownPromise = vm.drillDownTo(meta);

    expect(vm.drillBreadcrumbs).toEqual([]);
    expect(vm.drillStack).toHaveLength(0);

    resolveDrilldown([{ name: "Training", data: [{ name: "project-a", y: 2 }] }]);
    await drilldownPromise;

    expect(vm.drillBreadcrumbs).toHaveLength(2);
    expect(vm.drillBreadcrumbs[1]).toBe("Departments");
    expect(vm.drillStack).toHaveLength(1);
  });

  it("should use full-size drilldown geometry and show final project labels", async () => {
    vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() => Promise.resolve(WorkloadByTypeWidgetMock));
    createComponent();
    await flushPromises();

    const vm = wrapper.vm as unknown as WorkloadByTypeWidgetInstance;

    const intermediate = vm.convertSeriesToECharts(
      [
        {
          name: "Training",
          data: [
            {
              name: "department-a",
              y: 2,
              drilldown: true,
              meta: { entityId: "department-a", groupByKey: EWorkloadTelemetryGroupBy.DepartmentId },
            },
          ],
        },
      ],
      true,
    );
    const intermediateOption = vm.buildChartOption(intermediate.categories, intermediate.series, true);

    expect(intermediateOption.grid).toMatchObject({ top: 52, bottom: 86, left: 56, right: 12 });
    expect(intermediateOption.tooltip).toMatchObject({
      trigger: "axis",
      axisPointer: { type: "line", lineStyle: { opacity: 0, width: 0 } },
    });
    expect(intermediateOption.xAxis).toMatchObject({
      axisPointer: { type: "line", lineStyle: { opacity: 0, width: 0 } },
    });
    expect(intermediateOption.yAxis).toMatchObject({
      axisPointer: { type: "line", lineStyle: { opacity: 0, width: 0 } },
    });
    expect((intermediateOption.series as BarSeriesOption[])[0]).toMatchObject({
      barMaxWidth: 30,
      barGap: "25%",
      barCategoryGap: "40%",
    });
    expect(intermediateOption.xAxis).toMatchObject({ axisLabel: { show: false } });
    expect(intermediateOption.yAxis).toMatchObject({ max: 3, interval: 1 });

    const finalLevel = vm.convertSeriesToECharts([{ name: "Training", data: [{ name: "project-a", y: 2 }] }], true);
    const finalOption = vm.buildChartOption(finalLevel.categories, finalLevel.series, true);

    expect(finalOption.xAxis).toMatchObject({
      data: ["project-a"],
      axisLabel: { show: true },
    });

    const longProjectFinalLevel = vm.convertSeriesToECharts(
      [
        {
          name: "Training",
          data: [
            { name: "project-a", y: 2 },
            { name: "chartseed-20260512-vision", y: 1 },
            { name: "chartseed-20260512-burst", y: 1 },
            { name: "chartseed-20260512-platform", y: 1 },
          ],
        },
      ],
      true,
    );
    const longProjectFinalOption = vm.buildChartOption(
      longProjectFinalLevel.categories,
      longProjectFinalLevel.series,
      true,
    );

    expect(longProjectFinalOption.grid).toMatchObject({ top: 52, bottom: 86, left: 56, right: 12 });
    expect(longProjectFinalOption.xAxis).toMatchObject({
      axisLabel: {
        rotate: 0,
        overflow: "truncate",
        ellipsis: "…",
      },
    });
  });

  describe("Dynamic Color Assignment", () => {
    beforeEach(() => {
      vi.mocked(workloadService.getWorkloadsTelemetry).mockImplementation(() =>
        Promise.resolve(WorkloadByTypeWidgetMock),
      );
    });

    it("should assign colors dynamically based on first order of appearance", () => {
      createComponent();
      const vm = wrapper.vm as unknown as WorkloadByTypeWidgetInstance;

      // Create test series with different types
      const series = [
        { name: "TypeA", data: [] },
        { name: "TypeB", data: [] },
        { name: "TypeC", data: [] },
      ] as unknown as SeriesOptionsType[];

      const coloredSeries = vm.assignColorsToSeries(series);

      // Verify colors are assigned
      expect(coloredSeries[0].color).toBeDefined();
      expect(coloredSeries[1].color).toBeDefined();
      expect(coloredSeries[2].color).toBeDefined();

      // Verify colors are from the palette
      expect(chartColorPaletteArray).toContain(coloredSeries[0].color);
      expect(chartColorPaletteArray).toContain(coloredSeries[1].color);
      expect(chartColorPaletteArray).toContain(coloredSeries[2].color);
    });

    it("should lock colors once assigned - same type gets same color", () => {
      createComponent();
      const vm = wrapper.vm as unknown as WorkloadByTypeWidgetInstance;

      // First assignment
      const series1 = [
        { name: "Training", data: [] },
        { name: "Inference", data: [] },
      ] as unknown as SeriesOptionsType[];
      const coloredSeries1 = vm.assignColorsToSeries(series1);
      const trainingColor1 = coloredSeries1.find((s: SeriesOptionsType) => s.name === "Training")?.color;
      const inferenceColor1 = coloredSeries1.find((s: SeriesOptionsType) => s.name === "Inference")?.color;

      // Second assignment with same types
      const series2 = [
        { name: "Inference", data: [] },
        { name: "Training", data: [] },
      ] as unknown as SeriesOptionsType[];
      const coloredSeries2 = vm.assignColorsToSeries(series2);
      const trainingColor2 = coloredSeries2.find((s: SeriesOptionsType) => s.name === "Training")?.color;
      const inferenceColor2 = coloredSeries2.find((s: SeriesOptionsType) => s.name === "Inference")?.color;

      // Verify colors are locked - same type gets same color
      expect(trainingColor1).toBe(trainingColor2);
      expect(inferenceColor1).toBe(inferenceColor2);
    });

    it("should sort series alphabetically before color assignment", () => {
      createComponent();
      const vm = wrapper.vm as unknown as WorkloadByTypeWidgetInstance;

      // Create series in non-alphabetical order
      const series = [
        { name: "Zebra", data: [] },
        { name: "Apple", data: [] },
        { name: "Banana", data: [] },
      ] as unknown as SeriesOptionsType[];

      const coloredSeries = vm.assignColorsToSeries(series);

      // Verify series are sorted alphabetically
      expect(coloredSeries[0].name).toBe("Apple");
      expect(coloredSeries[1].name).toBe("Banana");
      expect(coloredSeries[2].name).toBe("Zebra");
    });

    it("should assign colors in order without repeating until all colors are used", () => {
      createComponent();
      const vm = wrapper.vm as unknown as WorkloadByTypeWidgetInstance;

      // Create more series than colors in palette
      // Use names that sort alphabetically correctly: TypeA, TypeB, TypeC, etc.
      const series: SeriesOptionsType[] = [];
      const totalSeries = chartColorPaletteArray.length + 3;
      for (let i = 0; i < totalSeries; i++) {
        series.push({ name: `Type${String.fromCharCode(65 + i)}`, data: [] } as unknown as SeriesOptionsType);
      }

      const coloredSeries = vm.assignColorsToSeries(series);

      // Verify first N series get unique colors (where N = palette length)
      // After sorting alphabetically, they should be TypeA, TypeB, TypeC, etc.
      const colorsAssigned = coloredSeries
        .slice(0, chartColorPaletteArray.length)
        .map((s: SeriesOptionsType) => s.color);
      const uniqueColors = new Set(colorsAssigned);
      expect(uniqueColors.size).toBe(chartColorPaletteArray.length);

      // Verify that colors are assigned in order from the palette
      // First color should be the first palette color, second should be second, etc.
      for (let i = 0; i < chartColorPaletteArray.length; i++) {
        expect(coloredSeries[i].color).toBe(chartColorPaletteArray[i]);
      }

      // Verify colors cycle after palette is exhausted
      // The (N+1)th series should get color[0] again (cycles back)
      const paletteLength = chartColorPaletteArray.length;
      expect(coloredSeries[paletteLength]?.color).toBe(chartColorPaletteArray[0]);

      // Verify the next series also cycles correctly
      if (totalSeries > paletteLength + 1) {
        expect(coloredSeries[paletteLength + 1]?.color).toBe(chartColorPaletteArray[1]);
      }
    });

    it("should maintain color consistency across multiple calls", () => {
      createComponent();
      const vm = wrapper.vm as unknown as WorkloadByTypeWidgetInstance;

      // First call
      const series1 = [
        { name: "Training", data: [] },
        { name: "Inference", data: [] },
        { name: "Workspace", data: [] },
      ] as unknown as SeriesOptionsType[];
      const coloredSeries1 = vm.assignColorsToSeries(series1);
      const trainingColor1 = coloredSeries1.find((s: SeriesOptionsType) => s.name === "Training")?.color;
      const inferenceColor1 = coloredSeries1.find((s: SeriesOptionsType) => s.name === "Inference")?.color;

      // Second call with different order and additional types
      const series2 = [
        { name: "Workspace", data: [] },
        { name: "Training", data: [] },
        { name: "Inference", data: [] },
        { name: "NewType", data: [] },
      ] as unknown as SeriesOptionsType[];
      const coloredSeries2 = vm.assignColorsToSeries(series2);
      const trainingColor2 = coloredSeries2.find((s: SeriesOptionsType) => s.name === "Training")?.color;
      const inferenceColor2 = coloredSeries2.find((s: SeriesOptionsType) => s.name === "Inference")?.color;

      // Verify consistency - same types get same colors
      expect(trainingColor1).toBe(trainingColor2);
      expect(inferenceColor1).toBe(inferenceColor2);
    });

    it("should return UNKNOWN_COLOR for empty or undefined type", () => {
      createComponent();
      const vm = wrapper.vm as unknown as WorkloadByTypeWidgetInstance;

      expect(vm.getColor("")).toBe(UNKNOWN_COLOR);
      expect(vm.getColor(undefined as unknown as string)).toBe(UNKNOWN_COLOR);
    });
  });
});
