/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { MockedFunction } from "vitest";
import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { flushPromises, mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import type { IRangeDates } from "@/utils/date.util";
import { AdvancedWorkloadMetricWidget } from "@/components/dashboard-v2/widgets/chart-widgets/advanced-metric/advanced-workload-metric-widget/index";
import type { TestingPinia } from "@pinia/testing";
import { createTestingPinia } from "@pinia/testing";
import { dateUtil, TimeUnit } from "@/utils/date.util";
import { stringUtil } from "@/utils/string.util";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { workloadPodMetricsMockResponse } from "@/components/dashboard-v2/widgets/chart-widgets/workload-status-widget/advanced-workload-metric-widget.mock";
import { chartUtil } from "@/utils/chart.util";
import {
  advancedWidgetConfigMapper,
  podMetricAdvancedWidgets,
} from "@/components/dashboard-v2/widgets/chart-widgets/advanced-metric/advanced-workload-metric-widget/widget-config-mapper";
import type { IGenericMetricWidgetConfig } from "@/models/chart.model";

vi.mock("@/services/cluster/workload.service/workload.service");
let wrapper: VueWrapper<InstanceType<typeof AdvancedWorkloadMetricWidget>>;
const pinia: TestingPinia = createTestingPinia({ createSpy: vi.fn });

const mockGetWorkloadPodAdvancedMetrics = workloadService.getWorkloadPodMetrics as MockedFunction<
  typeof workloadService.getWorkloadPodMetrics
>;
mockGetWorkloadPodAdvancedMetrics.mockImplementation(() => {
  return Promise.resolve(workloadPodMetricsMockResponse);
});
describe("NodePoolFreeResourcesWidget component", () => {
  const createComponent = (props: {
    workloadId: string;
    podId: string;
    filterByDates: IRangeDates;
    loading: boolean;
  }) => {
    wrapper = mount(AdvancedWorkloadMetricWidget, {
      global: {
        plugins: [Quasar, pinia],
      },
      props,
    });
  };

  beforeEach(async () => {
    createComponent({
      workloadId: "12dsadasd1231",
      podId: "12dsadasd1231sdsd",
      filterByDates: {
        dateStart: dateUtil.adjustDateBy(TimeUnit.hour, new Date(), -1) as Date,
        dateEnd: new Date(),
        label: "",
      },
      loading: false,
    });
    await flushPromises();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("should render AdvancedWorkloadMetricWidget component", async () => {
    expect(wrapper.find('[aid="advanced-workload-metric-container"]').exists()).eql(true);
  });

  it("should render all charts correctly", async () => {
    Object.values(advancedWidgetConfigMapper).forEach((config: IGenericMetricWidgetConfig) => {
      const aid = `[aid="${stringUtil.slug(config.chartTitle)}"]`;
      expect(wrapper.find(aid).exists()).eql(true);
    });
  });

  it("uses the connected-tooltip data contract for every advanced chart", () => {
    podMetricAdvancedWidgets.forEach(({ config }) => {
      expect(config.useCanonicalTimeBuckets).toBe(true);
      expect(config.useConnectedTooltipAnchor).toBe(true);
    });
  });

  it.each([
    "NVLink transmitted bandwidth",
    "NVLink received bandwidth",
    "PCIe transmitted bandwidth",
    "PCIe received bandwidth",
    "GPU swap memory",
  ])("uses the Highcharts-like zero scale for %s", (chartTitle) => {
    const config = Object.values(advancedWidgetConfigMapper).find((item) => item.chartTitle === chartTitle);

    expect(config?.chartOptions.highchartsLikeYAxis).toBe(true);
  });

  it("should call getWorkloadPodAdvancedMetrics upon props changes", async () => {
    const dateStart = dateUtil.adjustDateBy(TimeUnit.hour, new Date(), -2) as Date;
    const dateEnd = new Date();
    const timeframes = chartUtil.getTimeFrames(dateStart, dateEnd);

    await wrapper.setProps({
      filterByDates: {
        dateStart,
        dateEnd,
        label: "",
      },
    });
    await flushPromises();
    Object.values(advancedWidgetConfigMapper).forEach((config: IGenericMetricWidgetConfig) => {
      expect(mockGetWorkloadPodAdvancedMetrics).toHaveBeenCalledWith(
        "12dsadasd1231",
        "12dsadasd1231sdsd",
        config.metricTypes,
        dateStart.toISOString(),
        dateEnd.toISOString(),
        timeframes.length,
      );
    });
  });
});
