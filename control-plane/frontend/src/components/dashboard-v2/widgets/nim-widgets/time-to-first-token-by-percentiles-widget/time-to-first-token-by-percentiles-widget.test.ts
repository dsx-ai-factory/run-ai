/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";

// Component under test
import TimeToFirstTokenByPercentilesWidget from "./time-to-first-token-by-percentiles-widget.vue";

// Models and types
import type { IRangeDates } from "@/utils/date.util";
import { WorkloadMetricType, PodMetricType } from "@/swagger-models/workloads-client";

// Config imports
import {
  workloadTimeToFirstTokenByPercentilesWidgetConfig,
  podTimeToFirstTokenByPercentilesWidgetConfig,
  convertHistogramValueToArray,
  TIME_TO_FIRST_TOKEN_BY_PERCENTILES_WIDGET_CHART_ID,
} from "./time-to-first-token-by-percentiles-widget-config";

// Mock the services
vi.mock("@/services/cluster/workload.service/workload.service", () => ({
  workloadService: {
    getWorkloadMetrics: vi.fn(),
  },
}));

vi.mock("@/services/control-plane/pod.service/pod.service", () => ({
  podService: {
    getPodMetrics: vi.fn(),
  },
}));

// Mock the child component
vi.mock("@/components/dashboard-v2/widgets/chart-widgets/histogram-metric-widget", () => ({
  HistogramMetricWidget: {
    name: "HistogramMetricWidget",
    template: "<div data-testid='histogram-metric-widget'></div>",
    props: ["config", "filterByDates", "getDataArgs"],
    emits: ["register", "unregister", "on-zoom", "on-reset-zoom"],
  },
}));

describe("time-to-first-token-by-percentiles-widget.vue", () => {
  let wrapper: VueWrapper<InstanceType<typeof TimeToFirstTokenByPercentilesWidget>>;

  const mockWorkloadId = "test-workload-123";
  const mockPodId = "test-pod-456";
  const mockFilterByDates: IRangeDates = {
    dateStart: new Date("2024-03-19T00:00:00.346Z"),
    dateEnd: new Date("2024-03-19T23:59:59.658Z"),
    label: "Yesterday",
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount();
    }
  });

  function mountComponent(props: {
    workloadId: string;
    podId?: string;
    filterByDates: IRangeDates;
  }): VueWrapper<InstanceType<typeof TimeToFirstTokenByPercentilesWidget>> {
    return mount(TimeToFirstTokenByPercentilesWidget, {
      global: {
        plugins: [Quasar],
      },
      props,
    });
  }

  describe("Component Initialization", () => {
    it("should render without crashing", () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      expect(wrapper.exists()).toBe(true);
    });

    it("should render the HistogramMetricWidget child component", () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      const childComponent = wrapper.findComponent({ name: "HistogramMetricWidget" });
      expect(childComponent.exists()).toBe(true);
    });
  });

  describe("Props", () => {
    it("should accept required props", () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      expect(wrapper.props("workloadId")).toBe(mockWorkloadId);
      expect(wrapper.props("filterByDates")).toEqual(mockFilterByDates);
    });

    it("should accept optional podId prop", () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        podId: mockPodId,
        filterByDates: mockFilterByDates,
      });

      expect(wrapper.props("podId")).toBe(mockPodId);
    });

    it("should handle missing podId prop", () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      expect(wrapper.props("podId")).toBeUndefined();
    });
  });

  describe("Computed Properties", () => {
    describe("widgetConfig", () => {
      it("should return workload config when podId is not provided", () => {
        wrapper = mountComponent({
          workloadId: mockWorkloadId,
          filterByDates: mockFilterByDates,
        });

        const widgetConfig = wrapper.vm.widgetConfig;
        expect(widgetConfig).toBe(workloadTimeToFirstTokenByPercentilesWidgetConfig);
        expect(widgetConfig.metricTypes).toContain(WorkloadMetricType.NimTimeToFirstTokenSecondsPercentiles);
      });

      it("should return pod config when podId is provided", () => {
        wrapper = mountComponent({
          workloadId: mockWorkloadId,
          podId: mockPodId,
          filterByDates: mockFilterByDates,
        });

        const widgetConfig = wrapper.vm.widgetConfig;
        expect(widgetConfig).toBe(podTimeToFirstTokenByPercentilesWidgetConfig);
        expect(widgetConfig.metricTypes).toContain(PodMetricType.NimTimeToFirstTokenSecondsPercentiles);
      });

      it("should have correct chart configuration", () => {
        wrapper = mountComponent({
          workloadId: mockWorkloadId,
          filterByDates: mockFilterByDates,
        });

        const widgetConfig = wrapper.vm.widgetConfig;
        expect(widgetConfig.chartTitle).toBe("Time to first token (TTFT) by percentiles");
        expect(widgetConfig.chartId).toBe(TIME_TO_FIRST_TOKEN_BY_PERCENTILES_WIDGET_CHART_ID);
        expect(widgetConfig.chartOptions?.type).toBe("line");
        expect(widgetConfig.chartOptions?.height).toBe(250);
        expect(widgetConfig.chartOptions?.yAxisTitle).toBe("Milliseconds");
        expect(widgetConfig.chartOptions?.sharedTooltip).toBe(true);
        expect(widgetConfig.chartOptions?.showSharedCrosshair).toBe(true);
      });
    });

    describe("getDataArgs", () => {
      it("should return array with workloadId when podId is not provided", () => {
        wrapper = mountComponent({
          workloadId: mockWorkloadId,
          filterByDates: mockFilterByDates,
        });

        const getDataArgs = wrapper.vm.getDataArgs;
        expect(getDataArgs).toEqual([mockWorkloadId]);
      });

      it("should return array with workloadId and podId when podId is provided", () => {
        wrapper = mountComponent({
          workloadId: mockWorkloadId,
          podId: mockPodId,
          filterByDates: mockFilterByDates,
        });

        const getDataArgs = wrapper.vm.getDataArgs;
        expect(getDataArgs).toEqual([mockWorkloadId, mockPodId]);
      });
    });
  });

  describe("Child Component Integration", () => {
    it("should pass correct props to HistogramMetricWidget", () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        podId: mockPodId,
        filterByDates: mockFilterByDates,
      });

      const childComponent = wrapper.findComponent({ name: "HistogramMetricWidget" });

      expect(childComponent.props("config")).toBe(podTimeToFirstTokenByPercentilesWidgetConfig);
      expect(childComponent.props("filterByDates")).toEqual(mockFilterByDates);
      expect(childComponent.props("getDataArgs")).toEqual([mockWorkloadId, mockPodId]);
    });

    it("should pass workload config when podId is not provided", () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      const childComponent = wrapper.findComponent({ name: "HistogramMetricWidget" });

      expect(childComponent.props("config")).toBe(workloadTimeToFirstTokenByPercentilesWidgetConfig);
      expect(childComponent.props("getDataArgs")).toEqual([mockWorkloadId]);
    });
  });

  describe("Event Handling", () => {
    it("should emit register event from child component", async () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      const childComponent = wrapper.findComponent({ name: "HistogramMetricWidget" });
      childComponent.vm.$emit("register");

      expect(wrapper.emitted("register")).toBeTruthy();
      expect(wrapper.emitted("register")?.[0]).toEqual([]);
    });

    it("should emit unregister event from child component", async () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      const childComponent = wrapper.findComponent({ name: "HistogramMetricWidget" });
      childComponent.vm.$emit("unregister");

      expect(wrapper.emitted("unregister")).toBeTruthy();
      expect(wrapper.emitted("unregister")?.[0]).toEqual([]);
    });

    it("should emit on-zoom event from child component with event data", async () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      const zoomEvent = { chartId: "test-chart", min: 1000, max: 2000 };
      const childComponent = wrapper.findComponent({ name: "HistogramMetricWidget" });
      childComponent.vm.$emit("on-zoom", zoomEvent);

      expect(wrapper.emitted("on-zoom")).toBeTruthy();
      expect(wrapper.emitted("on-zoom")?.[0]).toEqual([zoomEvent]);
    });

    it("should emit on-reset-zoom event from child component with event data", async () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      const resetZoomEvent = "test-chart-id";
      const childComponent = wrapper.findComponent({ name: "HistogramMetricWidget" });
      childComponent.vm.$emit("on-reset-zoom", resetZoomEvent);

      expect(wrapper.emitted("on-reset-zoom")).toBeTruthy();
      expect(wrapper.emitted("on-reset-zoom")?.[0]).toEqual([resetZoomEvent]);
    });
  });

  describe("Component Name", () => {
    it("should have correct component name", () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      expect(wrapper.vm.$options.name).toBe("time-to-first-token-by-percentiles-widget");
    });
  });

  describe("Emits Declaration", () => {
    it("should declare all required emits", () => {
      wrapper = mountComponent({
        workloadId: mockWorkloadId,
        filterByDates: mockFilterByDates,
      });

      const emits = wrapper.vm.$options.emits as string[];
      expect(emits).toContain("register");
      expect(emits).toContain("unregister");
      expect(emits).toContain("on-zoom");
      expect(emits).toContain("on-reset-zoom");
    });
  });

  describe("Props Validation", () => {
    it("should require workloadId prop", () => {
      const componentOptions = TimeToFirstTokenByPercentilesWidget;
      const workloadIdProp = componentOptions.props?.workloadId;

      expect(workloadIdProp).toBeDefined();
      expect(workloadIdProp?.required).toBe(true);
    });

    it("should require filterByDates prop", () => {
      const componentOptions = TimeToFirstTokenByPercentilesWidget;
      const filterByDatesProp = componentOptions.props?.filterByDates;

      expect(filterByDatesProp).toBeDefined();
      expect(filterByDatesProp?.required).toBe(true);
    });

    it("should make podId prop optional", () => {
      const componentOptions = TimeToFirstTokenByPercentilesWidget;
      const podIdProp = componentOptions.props?.podId;

      expect(podIdProp).toBeDefined();
      expect(podIdProp?.required).toBe(false);
    });
  });
});

describe("convertHistogramValueToArray utility function", () => {
  it("should convert histogram values to array format correctly", () => {
    const mockHistogramValues = [
      {
        timestamp: "2024-03-19T09:39:00.346Z",
        data: {
          p50: "10",
          p90: "20",
          p99: "30",
        },
      },
      {
        timestamp: "2024-03-19T10:39:00.346Z",
        data: {
          p50: "11",
          p90: "21",
          p99: "31",
        },
      },
    ];

    const result = convertHistogramValueToArray(mockHistogramValues);

    expect(result).toEqual({
      "P50 requests": [
        { timestamp: "2024-03-19T09:39:00.346Z", value: "10" },
        { timestamp: "2024-03-19T10:39:00.346Z", value: "11" },
      ],
      "P90 requests": [
        { timestamp: "2024-03-19T09:39:00.346Z", value: "20" },
        { timestamp: "2024-03-19T10:39:00.346Z", value: "21" },
      ],
      "P99 requests": [
        { timestamp: "2024-03-19T09:39:00.346Z", value: "30" },
        { timestamp: "2024-03-19T10:39:00.346Z", value: "31" },
      ],
    });
  });

  it("should handle empty histogram values array", () => {
    const result = convertHistogramValueToArray([]);
    expect(result).toEqual({});
  });

  it("should skip histogram values with missing timestamp or data", () => {
    const mockHistogramValues = [
      {
        timestamp: "2024-03-19T09:39:00.346Z",
        data: {
          p50: "10",
        },
      },
      {
        timestamp: "2024-03-19T10:39:00.346Z",
        data: {
          p50: "20",
        },
      },
      {
        timestamp: "2024-03-19T11:39:00.346Z",
        data: {
          p50: "30",
        },
      },
    ];

    const result = convertHistogramValueToArray(mockHistogramValues);

    expect(result).toEqual({
      "P50 requests": [
        { timestamp: "2024-03-19T09:39:00.346Z", value: "10" },
        { timestamp: "2024-03-19T10:39:00.346Z", value: "20" },
        { timestamp: "2024-03-19T11:39:00.346Z", value: "30" },
      ],
    });
  });

  it("should handle histogram values with empty data object", () => {
    const mockHistogramValues = [
      {
        timestamp: "2024-03-19T09:39:00.346Z",
        data: {},
      },
    ];

    const result = convertHistogramValueToArray(mockHistogramValues);
    expect(result).toEqual({});
  });
});
