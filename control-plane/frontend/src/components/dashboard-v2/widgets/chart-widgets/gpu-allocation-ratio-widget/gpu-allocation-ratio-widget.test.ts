/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia, type TestingPinia } from "@pinia/testing";
import { GpuAllocationRatioWidget } from "@/components/dashboard-v2/widgets/chart-widgets/gpu-allocation-ratio-widget/index";
import { departmentService } from "@/services/control-plane/department.service/department.service";
import {
  mockGpuAllocationRatio,
  mockGpuAllocationRatioFilterByNodepool,
  mockGpuAllocationRatioUnlimitedQuota,
  mockGpuAllocationRatioMixedQuota,
} from "@/components/dashboard-v2/widgets/chart-widgets/gpu-allocation-ratio-widget/gpu-allocation-ratio-widget.mock";
import type { DepartmentsMetricData, IChartArrayItemConfig } from "@/models/metrics.model";

vi.mock("@/services/control-plane/department.service/department.service");

const mockRouter = {
  push: vi.fn(),
};

vi.mock("vue-router", () => ({
  useRouter: () => mockRouter,
  useRoute: () => ({}),
}));

let wrapper: VueWrapper<InstanceType<typeof GpuAllocationRatioWidget>>;
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });
const clusterId = "34234-322-sdfs323";

interface GpuAllocationRatioWidgetTestVm {
  allocationClusterByNodePoolArray: IChartArrayItemConfig[];
  breadcrumbPath: Array<{ label: string }>;
  chartOption: unknown;
  handleAxisCategoryClick: (category: string) => void;
  isChartEmpty: boolean;
  showDrilldownData: (series: IChartArrayItemConfig[], parentGroupId?: string) => void;
}

function getTestVm(): GpuAllocationRatioWidgetTestVm {
  return wrapper.vm as unknown as GpuAllocationRatioWidgetTestVm;
}

describe("GpuAllocationRatioWidget component", () => {
  const createComponent = (props?: { nodePoolName?: string }, hasDepartments = false) => {
    wrapper = mount(GpuAllocationRatioWidget, {
      global: {
        plugins: [Quasar, pinia],
      },
      props: {
        ...props,
        clusterId,
        hasDepartments: hasDepartments,
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

  it("should render AllocatedGpuWidget component", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockImplementation(() =>
      Promise.resolve(mockGpuAllocationRatio as DepartmentsMetricData),
    );
    createComponent();
    await flushPromises();
    expect(wrapper.find('[aid="allocation-ratio-by-node-pool"]').text()).eql("Allocation ratio by node pool");
  });

  it("should render AllocatedGpuWidget component with filter", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockImplementation(() =>
      Promise.resolve(mockGpuAllocationRatioFilterByNodepool as DepartmentsMetricData),
    );
    createComponent({ nodePoolName: "node-pool-2" });
    await flushPromises();
    expect(departmentService.getDepartmentsMetric).toHaveBeenCalledWith(clusterId, "node-pool-2");
    expect(wrapper.find(".widget-link-btn").text()).eql("All projects");
  });

  it("should link to projects table when clicked", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockImplementation(() =>
      Promise.resolve(mockGpuAllocationRatio as DepartmentsMetricData),
    );
    createComponent();
    await flushPromises();
    await clickOnGridLink();
    expect(mockRouter.push).toHaveBeenCalledWith({ name: "project-index" });
  });

  it("should link to departments table when clicked", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockImplementation(() =>
      Promise.resolve(mockGpuAllocationRatio as DepartmentsMetricData),
    );
    createComponent({}, true);
    await flushPromises();
    await clickOnGridLink();
    expect(mockRouter.push).toHaveBeenCalledWith({ name: "department-index" });
  });

  it("drills a cluster down to all node-pool department series", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockResolvedValue(mockGpuAllocationRatio as DepartmentsMetricData);
    createComponent({}, true);
    await flushPromises();

    const vm = getTestVm();
    vm.handleAxisCategoryClick("barak-1-1711864663");
    await wrapper.vm.$nextTick();

    expect(vm.breadcrumbPath.map((crumb) => crumb.label)).toEqual(["barak-1-1711864663", "Departments"]);
    const chartOption = vm.chartOption as {
      grid: { bottom: number };
      xAxis: { data: string[] };
      yAxis: { max: number };
      series: Array<{ id: string; name: string; universalTransition: { enabled: boolean; divideShape: string } }>;
      animationDuration: number;
      animationDurationUpdate: number;
      animationEasing: string;
      animationEasingUpdate: string;
    };
    expect(chartOption.xAxis.data).toEqual(["default"]);
    expect(chartOption.series.map((series) => series.name).sort()).toEqual(["default", "node-pool-2"]);
    expect(chartOption.series.every((series) => series.id.startsWith("gpu-allocation-ratio-"))).toEqual(true);
    expect(chartOption.series.every((series) => series.universalTransition.enabled)).toEqual(true);
    expect(chartOption.series.every((series) => series.universalTransition.divideShape === "clone")).toEqual(true);
    expect(chartOption.grid.bottom).toEqual(76);
    expect(chartOption.yAxis.max).toEqual(100);
    expect(chartOption.animationDuration).toEqual(500);
    expect(chartOption.animationDurationUpdate).toEqual(500);
    expect(chartOption.animationEasing).toEqual("sinusoidalInOut");
    expect(chartOption.animationEasingUpdate).toEqual("sinusoidalInOut");
    expect(wrapper.find(".breadcrumb-current").text()).toEqual("Departments");
  });

  it("aligns sparse node-pool series with their department categories", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockResolvedValue(mockGpuAllocationRatio as DepartmentsMetricData);
    createComponent({}, true);
    await flushPromises();

    const sparseDepartmentSeries: IChartArrayItemConfig[] = [
      {
        id: "departments-empty",
        name: "chart-qa-nodepool-empty",
        data: [{ name: "chart-qa-department-beta", y: 0, meta: { quota: 2, allocated: 0 } }],
      },
      {
        id: "departments-live",
        name: "chart-qa-nodepool-live",
        data: [{ name: "chart-qa-department-alpha", y: 75, meta: { quota: 4, allocated: 3 } }],
      },
      {
        id: "departments-default",
        name: "default",
        data: [{ name: "default", y: 0, meta: { quota: 9, allocated: 0 } }],
      },
    ];

    const vm = getTestVm();
    vm.showDrilldownData(sparseDepartmentSeries);

    const chartOption = vm.chartOption as {
      xAxis: { data: string[] };
      series: Array<{ data: Array<{ value: number } | null> }>;
    };
    expect(chartOption.xAxis.data).toEqual(["chart-qa-department-beta", "chart-qa-department-alpha", "default"]);
    expect(chartOption.series.map((series) => series.data.map((point) => point?.value ?? null))).toEqual([
      [0, null, null],
      [null, 75, null],
      [null, null, 0],
    ]);
  });

  it("should trigger get telemetry data on filter change", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockImplementation(() =>
      Promise.resolve(mockGpuAllocationRatio as DepartmentsMetricData),
    );
    createComponent();
    await flushPromises();
    await wrapper.setProps({ nodePoolName: "node-pool-2" });
    expect(departmentService.getDepartmentsMetric).toHaveBeenCalledTimes(2);
  });

  it("should display loading skeleton when loading", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockImplementation(() =>
      Promise.resolve(mockGpuAllocationRatio as DepartmentsMetricData),
    );
    createComponent();
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(true);
    await flushPromises();
    expect(wrapper.findComponent(".q-skeleton").exists()).toEqual(false);
  });

  it("should display error message when telemetry fetch fails", async () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    vi.mocked(departmentService.getDepartmentsMetric).mockRejectedValueOnce({ response: { status: 500 } });
    createComponent();
    await flushPromises();
    expect(wrapper.find(".error-section").exists()).toEqual(true);
    expect(wrapper.find(".error-section").text()).eql("Data is not available right now");
    errorSpy.mockRestore();
  });

  it("should keep unlimited quota entries visible with y=0 and meta.quota=-1", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockImplementation(() =>
      Promise.resolve(mockGpuAllocationRatioUnlimitedQuota as DepartmentsMetricData),
    );
    createComponent();
    await flushPromises();

    const vm = getTestVm();
    // All unlimited quotas should result in quotaY = -1 and allocation ratio (y) = 0
    const clusterData = vm.allocationClusterByNodePoolArray;
    const defaultPool = clusterData.find((item) => item.name === "default");
    expect(defaultPool).toBeDefined();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const dataPoint = defaultPool?.data[0] as any;
    expect(dataPoint?.meta?.quota).toEqual(-1);
    expect(dataPoint?.y).toEqual(0);
    // Unlimited entries should NOT be filtered out — chart still shows data
    expect(vm.isChartEmpty).toEqual(false);
  });

  it("should mark aggregate as unlimited when any quota is unlimited", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockImplementation(() =>
      Promise.resolve(mockGpuAllocationRatioMixedQuota as DepartmentsMetricData),
    );
    createComponent();
    await flushPromises();

    const vm = getTestVm();
    const clusterData = vm.allocationClusterByNodePoolArray;
    const defaultPool = clusterData.find((item) => item.name === "default");
    expect(defaultPool).toBeDefined();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const dataPoint = defaultPool?.data[0] as any;
    // Any unlimited quota in the mix makes the aggregate unlimited
    // allocated = 30 + 20 = 50, quota = -1 (unlimited), ratio = 0
    expect(dataPoint?.meta?.quota).toEqual(-1);
    expect(dataPoint?.meta?.allocated).toEqual(50);
    expect(dataPoint?.y).toEqual(0);
  });

  it("should display Unlimited in tooltip for unlimited quota entries", async () => {
    vi.mocked(departmentService.getDepartmentsMetric).mockImplementation(() =>
      Promise.resolve(mockGpuAllocationRatioUnlimitedQuota as DepartmentsMetricData),
    );
    createComponent();
    await flushPromises();

    const vm = getTestVm();
    const chartOpt = vm.chartOption as Record<string, unknown>;
    const tooltip = chartOpt?.tooltip as Record<string, unknown>;
    const formatter = tooltip?.formatter as (params: Record<string, unknown>) => string;
    expect(formatter).toBeDefined();

    const mockParams = {
      data: { name: "default", value: 0, meta: { allocated: 10, quota: -1 } },
      color: "#0654FE",
      seriesName: "default",
    };
    const tooltipHtml = formatter(mockParams);
    expect(tooltipHtml).toContain("Unlimited");
    expect(tooltipHtml).toContain("GPU allocation ratio: <b>-</b>");
  });
});
