/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import WorkloadByCategoryTimeRangeWidget from "./workload-by-category-time-range-widget.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { mockWorkloadByCategoryGroupedMetrics } from "./workload-by-category-time-range-widget.mock";

// models
import type { GroupedMetricsResponse } from "@/swagger-models/cluster-service-client";
import type { IRangeDates } from "@/utils/date.util";
import type { EChartsOption } from "echarts";
import { chartUtil } from "@/utils/chart.util";

vi.mock("@/services/control-plane/cluster.service/cluster.service");

function mountComponent(
  props: { clusterId: string; filterByDates: IRangeDates; nodePoolName?: string },
  mockResolvedValue?: GroupedMetricsResponse,
  mockRejectedValue?: string,
) {
  if (mockResolvedValue) {
    vi.mocked(clusterService.getClusterGroupedMetrics).mockResolvedValue(mockResolvedValue);
  } else if (mockRejectedValue) {
    vi.mocked(clusterService.getClusterGroupedMetrics).mockRejectedValue(mockRejectedValue);
  }

  return mount(WorkloadByCategoryTimeRangeWidget, {
    global: {
      plugins: [Quasar],
      directives: {
        clickOutside: vi.fn(),
        permission: vi.fn(),
      },
    },
    props: {
      clusterId: props.clusterId,
      filterByDates: props.filterByDates,
      nodePoolName: props.nodePoolName,
    },
  });
}

describe("workload-by-category-time-range-widget.vue", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders with loading state initially", async () => {
    const wrapper = mountComponent(
      {
        clusterId: "test-cluster-id",
        filterByDates: {
          dateStart: new Date("2024-03-01T00:00:00.000Z"),
          dateEnd: new Date("2024-03-01T01:00:00.000Z"),
          label: "Last 1 hour",
        },
      },
      mockWorkloadByCategoryGroupedMetrics,
    );

    expect(wrapper.exists()).toBe(true);
    expect(wrapper.findComponent({ name: "generic-metric-widget" }).exists()).toBe(true);
  });

  it("passes correct props to GenericMetricWidget", () => {
    const props = {
      clusterId: "test-cluster-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
      nodePoolName: "test-node-pool",
    };

    const wrapper = mountComponent(props, mockWorkloadByCategoryGroupedMetrics);
    const genericMetricWidget = wrapper.findComponent({ name: "generic-metric-widget" });

    expect(genericMetricWidget.props("filterByDates")).toEqual(props.filterByDates);
    expect(genericMetricWidget.props("getDataArgs")).toEqual([props.clusterId, props.nodePoolName]);
    expect(genericMetricWidget.props("config")).toBeDefined();
    expect(genericMetricWidget.props("config").useCanonicalTimeBuckets).toBe(true);
    expect(genericMetricWidget.props("config").useConnectedTooltipAnchor).toBe(true);
  });

  it("passes correct data args when nodePoolName is not provided", () => {
    const props = {
      clusterId: "test-cluster-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    };

    const wrapper = mountComponent(props, mockWorkloadByCategoryGroupedMetrics);
    const genericMetricWidget = wrapper.findComponent({ name: "generic-metric-widget" });

    expect(genericMetricWidget.props("getDataArgs")).toEqual([props.clusterId, ""]);
  });

  it("keeps the connected tooltip anchor out of the visible legend", () => {
    const wrapper = mountComponent({
      clusterId: "test-cluster-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    });
    const genericMetricWidget = wrapper.findComponent({ name: "generic-metric-widget" });
    const chartOptions: EChartsOption = {
      series: [
        chartUtil.getConnectedTooltipAnchorSeries(["2024-03-01T00:00:00.000Z"]),
        { name: "Deploy", type: "line", data: [[0, 2]], itemStyle: { color: "#2CAFFE" } },
        { name: "Train", type: "line", data: [[0, 4]], itemStyle: { color: "#FE6A35" } },
      ],
      legend: {},
      yAxis: {},
      grid: {},
      xAxis: {},
    };
    (genericMetricWidget.vm as unknown as { chartOptions: EChartsOption }).chartOptions = chartOptions;

    wrapper.vm.applyHighchartsParityOptions();

    expect((chartOptions.legend as { data?: string[] }).data).toEqual(["Deploy", "Train"]);
  });

  it("emits register event", () => {
    const wrapper = mountComponent({
      clusterId: "test-cluster-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    });

    const genericMetricWidget = wrapper.findComponent({ name: "generic-metric-widget" });
    genericMetricWidget.vm.$emit("register");

    expect(wrapper.emitted("register")).toBeTruthy();
  });

  it("emits unregister event", () => {
    const wrapper = mountComponent({
      clusterId: "test-cluster-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    });

    const genericMetricWidget = wrapper.findComponent({ name: "generic-metric-widget" });
    genericMetricWidget.vm.$emit("unregister");

    expect(wrapper.emitted("unregister")).toBeTruthy();
  });

  it("emits zoom events", () => {
    const wrapper = mountComponent({
      clusterId: "test-cluster-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    });

    const genericMetricWidget = wrapper.findComponent({ name: "generic-metric-widget" });
    const zoomData = { min: 123456789, max: 123456790 };

    genericMetricWidget.vm.$emit("on-zoom", zoomData);
    expect(wrapper.emitted("on-zoom")).toBeTruthy();
    expect(wrapper.emitted("on-zoom")?.[0]).toEqual([zoomData]);

    genericMetricWidget.vm.$emit("on-reset-zoom", zoomData);
    expect(wrapper.emitted("on-reset-zoom")).toBeTruthy();
    expect(wrapper.emitted("on-reset-zoom")?.[0]).toEqual([zoomData]);
  });
});
