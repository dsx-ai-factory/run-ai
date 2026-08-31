/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount, flushPromises } from "@vue/test-utils";
import { beforeEach, afterEach, vi, describe, it, expect } from "vitest";
import { createTestingPinia } from "@pinia/testing";
import { setActivePinia } from "pinia";
import { Quasar } from "quasar";
import QuotaAllocationTimeRangeWidget from "./quota-allocation-time-range-widget.vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { mockQuotaAllocationMetrics } from "./quota-allocation-time-range-widget.mock";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";
import { PREDEFINED_HOURS, PREDEFINED_MINUTES } from "@/models/date.model";
import { OrgUnitMetricType } from "@/swagger-models/org-unit-service-client";

import type { MetricsResponse } from "@/swagger-models/cluster-service-client";

vi.mock("@/services/control-plane/cluster.service/cluster.service");
vi.mock("@/services/control-plane/org-unit.service/org-unit.service");

const pinia = createTestingPinia({ createSpy: vi.fn });
setActivePinia(pinia);

function getSeries(chartOptions: unknown): Record<string, unknown>[] {
  const series = (chartOptions as { series?: unknown } | null)?.series;
  return (Array.isArray(series) ? series : series ? [series] : []) as Record<string, unknown>[];
}

function getXAxis(chartOptions: unknown): Record<string, unknown> {
  const xAxis = (chartOptions as { xAxis?: unknown } | null)?.xAxis;
  return (Array.isArray(xAxis) ? xAxis[0] : xAxis) as Record<string, unknown>;
}

describe("quota-allocation-time-range-widget.vue", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function mountComponent(
    props: {
      clusterId: string;
      nodePoolName?: string;
      orgUnitMetricSource?: IOrgUnitMetricSource | null;
    },
    mockResolvedValue?: MetricsResponse,
    mockRejectedValue?: string,
  ): VueWrapper<InstanceType<typeof QuotaAllocationTimeRangeWidget>> {
    if (mockResolvedValue) {
      vi.mocked(clusterService.getClusterMetrics).mockResolvedValue(mockResolvedValue);
      vi.mocked(clusterService.getNodepoolMetrics).mockResolvedValue(mockResolvedValue);
      vi.mocked(orgUnitService.getDepartmentMetrics).mockResolvedValue(mockResolvedValue as never);
      vi.mocked(orgUnitService.getProjectMetrics).mockResolvedValue(mockResolvedValue as never);
    } else if (mockRejectedValue) {
      vi.mocked(clusterService.getClusterMetrics).mockRejectedValue(mockRejectedValue);
      vi.mocked(clusterService.getNodepoolMetrics).mockRejectedValue(mockRejectedValue);
    }

    return mount(QuotaAllocationTimeRangeWidget, {
      global: {
        plugins: [Quasar, pinia],
      },
      props,
    });
  }

  it("renders chart wrapper and loads data on mount", async () => {
    const wrapper = mountComponent({ clusterId: "cluster-1" }, mockQuotaAllocationMetrics);

    expect(wrapper.find(".quota-allocation-time-range-widget").exists()).toBe(true);
    expect(wrapper.find(".chart-widget-wrapper").exists()).toBe(true);

    await flushPromises();

    expect(clusterService.getClusterMetrics).toHaveBeenCalled();
    expect(getSeries(wrapper.vm.chartOptions)).toHaveLength(2);
  });

  it("does not consume page scrolling while area zoom is inactive", () => {
    const wrapper = mountComponent({ clusterId: "cluster-1" }, mockQuotaAllocationMetrics);
    const dataZoom = Array.isArray(wrapper.vm.chartOptions?.dataZoom)
      ? wrapper.vm.chartOptions.dataZoom[0]
      : wrapper.vm.chartOptions?.dataZoom;

    expect(dataZoom).toMatchObject({
      type: "inside",
      disabled: true,
      zoomOnMouseWheel: false,
      moveOnMouseWheel: false,
      moveOnMouseMove: false,
      preventDefaultMouseMove: false,
    });
  });

  it("renders series with correct types", async () => {
    const wrapper = mountComponent({ clusterId: "cluster-1" }, mockQuotaAllocationMetrics);

    await flushPromises();

    const series = getSeries(wrapper.vm.chartOptions);
    const allocatedSeries = series[0];
    expect(allocatedSeries?.name).toBe("Allocation");
    expect(allocatedSeries?.type).toBe("line");

    const quotaSeries = series[1];
    expect(quotaSeries?.name).toBe("Quota");
    expect(quotaSeries?.type).toBe("line");
  });

  it("shows error state on fetch failure", async () => {
    const wrapper = mountComponent({ clusterId: "cluster-1" }, undefined, "Network error");

    await flushPromises();

    expect(clusterService.getClusterMetrics).toHaveBeenCalled();
    expect(wrapper.find(".error-section").exists()).toBe(true);
  });

  it("uses nodepool metrics when nodePoolName is provided", async () => {
    mountComponent({ clusterId: "cluster-1", nodePoolName: "default" }, mockQuotaAllocationMetrics);

    await flushPromises();

    expect(clusterService.getNodepoolMetrics).toHaveBeenCalledWith(
      "cluster-1",
      "default",
      expect.any(String),
      expect.any(String),
      expect.any(Array),
      expect.any(Number),
    );
    expect(clusterService.getClusterMetrics).not.toHaveBeenCalled();
  });

  it("reloads data when date range changes", async () => {
    const wrapper = mountComponent({ clusterId: "cluster-1" }, mockQuotaAllocationMetrics);

    await flushPromises();
    vi.mocked(clusterService.getClusterMetrics).mockClear();

    const dateEnd = new Date();
    const dateStart = new Date(dateEnd.getTime() - 14 * 24 * 60 * 60 * 1000);
    wrapper.vm.filterByDates = { dateStart, dateEnd, label: "Last 14 days" };

    await flushPromises();

    expect(clusterService.getClusterMetrics).toHaveBeenCalled();
  });

  it("uses native hourly time-axis ticks for the Last 6 hours preset", async () => {
    const wrapper = mountComponent({ clusterId: "cluster-1" }, mockQuotaAllocationMetrics);
    const dateEnd = new Date("2026-07-13T11:00:00.000Z");
    const dateStart = new Date(dateEnd.getTime() - 6 * 60 * 60 * 1000);

    wrapper.vm.filterByDates = {
      dateStart,
      dateEnd,
      label: PREDEFINED_HOURS.LAST_6_HOURS,
    };
    await flushPromises();

    expect(getXAxis(wrapper.vm.chartOptions)).toMatchObject({
      type: "time",
      splitNumber: 14,
      interval: 60 * 60 * 1000,
      minInterval: 60 * 60 * 1000,
      maxInterval: 60 * 60 * 1000,
      min: dateStart.getTime(),
      max: dateEnd.getTime(),
    });
  });

  it("keeps one-minute native ticks for the Last 15 minutes preset", async () => {
    const wrapper = mountComponent({ clusterId: "cluster-1" }, mockQuotaAllocationMetrics);
    const dateEnd = new Date("2026-07-13T11:00:00.000Z");
    const dateStart = new Date(dateEnd.getTime() - 15 * 60 * 1000);

    wrapper.vm.filterByDates = {
      dateStart,
      dateEnd,
      label: PREDEFINED_MINUTES.LAST_15_MINUTES,
    };
    await flushPromises();

    expect(getXAxis(wrapper.vm.chartOptions)).toMatchObject({
      type: "time",
      interval: 60 * 1000,
      minInterval: 60 * 1000,
      maxInterval: 60 * 1000,
      min: dateStart.getTime(),
      max: dateEnd.getTime(),
    });
  });

  it("uses department metrics when orgUnitMetricSource is department", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    mountComponent({ clusterId: "cluster-1", orgUnitMetricSource: source }, mockQuotaAllocationMetrics);

    await flushPromises();

    expect(orgUnitService.getDepartmentMetrics).toHaveBeenCalledWith(
      "dept-1",
      [OrgUnitMetricType.GpuAllocation, OrgUnitMetricType.GpuQuota],
      expect.any(String),
      expect.any(String),
      expect.any(Number),
      undefined,
    );
    expect(clusterService.getClusterMetrics).not.toHaveBeenCalled();
  });

  it("uses project metrics when orgUnitMetricSource is project", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Project, id: "proj-1" };
    mountComponent({ clusterId: "cluster-1", orgUnitMetricSource: source }, mockQuotaAllocationMetrics);

    await flushPromises();

    expect(orgUnitService.getProjectMetrics).toHaveBeenCalledWith(
      "proj-1",
      [OrgUnitMetricType.GpuAllocation, OrgUnitMetricType.GpuQuota],
      expect.any(String),
      expect.any(String),
      expect.any(Number),
      undefined,
    );
    expect(clusterService.getClusterMetrics).not.toHaveBeenCalled();
  });

  it("passes nodePoolName to org-unit metrics when both are set", async () => {
    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    mountComponent(
      { clusterId: "cluster-1", nodePoolName: "gpu-pool", orgUnitMetricSource: source },
      mockQuotaAllocationMetrics,
    );

    await flushPromises();

    expect(orgUnitService.getDepartmentMetrics).toHaveBeenCalledWith(
      "dept-1",
      expect.any(Array),
      expect.any(String),
      expect.any(String),
      expect.any(Number),
      "gpu-pool",
    );
  });

  it("reloads data when orgUnitMetricSource changes", async () => {
    const wrapper = mountComponent({ clusterId: "cluster-1" }, mockQuotaAllocationMetrics);

    await flushPromises();
    vi.mocked(clusterService.getClusterMetrics).mockClear();

    const source: IOrgUnitMetricSource = { type: EOrgUnitMetricSourceType.Department, id: "dept-1" };
    await wrapper.setProps({ orgUnitMetricSource: source });
    await flushPromises();

    expect(orgUnitService.getDepartmentMetrics).toHaveBeenCalled();
    expect(clusterService.getClusterMetrics).not.toHaveBeenCalled();
  });
});
