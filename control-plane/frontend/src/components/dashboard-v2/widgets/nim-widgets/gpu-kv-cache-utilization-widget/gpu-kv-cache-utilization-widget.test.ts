/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import { Quasar } from "quasar";
import { createTestingPinia } from "@pinia/testing";
import GpuKvCacheUtilizationWidget from "./gpu-kv-cache-utilization-widget.vue";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { podService } from "@/services/control-plane/pod.service/pod.service";

// models
import type { IRangeDates } from "@/utils/date.util";

// Mock the workload service
vi.mock("@/services/cluster/workload.service/workload.service");
// Mock the pod service
vi.mock("@/services/control-plane/pod.service/pod.service");

function mountComponent(props: { workloadId: string; filterByDates: IRangeDates; podId?: string }) {
  const pinia = createTestingPinia({ createSpy: vi.fn });

  return mount(GpuKvCacheUtilizationWidget, {
    global: {
      plugins: [Quasar, pinia],
      directives: {
        clickOutside: vi.fn(),
        permission: vi.fn(),
      },
    },
    props,
  });
}

describe("gpu-kv-cache-utilization-widget.vue", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(workloadService.getWorkloadMetrics).mockResolvedValue({ measurements: [] });
    vi.mocked(podService.getPodMetrics).mockResolvedValue({ measurements: [] });
  });

  it("renders with loading state initially", async () => {
    const wrapper = mountComponent({
      workloadId: "test-workload-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    });

    expect(wrapper.exists()).toBe(true);
    expect(wrapper.findComponent({ name: "generic-metric-widget" }).exists()).toBe(true);
  });

  it("passes correct props to GenericMetricWidget", () => {
    const props = {
      workloadId: "test-workload-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    };

    const wrapper = mountComponent(props);
    const genericMetricWidget = wrapper.findComponent({ name: "generic-metric-widget" });

    expect(genericMetricWidget.props("filterByDates")).toEqual(props.filterByDates);
    expect(genericMetricWidget.props("getDataArgs")).toEqual([props.workloadId]);
    expect(genericMetricWidget.props("config")).toBeDefined();
  });

  it("passes correct data args when workloadId is provided", () => {
    const props = {
      workloadId: "test-workload-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    };

    const wrapper = mountComponent(props);
    const genericMetricWidget = wrapper.findComponent({ name: "generic-metric-widget" });

    expect(genericMetricWidget.props("getDataArgs")).toEqual([props.workloadId]);
  });

  it("emits register event", () => {
    const wrapper = mountComponent({
      workloadId: "test-workload-id",
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
      workloadId: "test-workload-id",
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
      workloadId: "test-workload-id",
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

  it("computes widgetConfig correctly", () => {
    const wrapper = mountComponent({
      workloadId: "test-workload-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    });

    const vm = wrapper.vm as InstanceType<typeof GpuKvCacheUtilizationWidget>;
    expect(vm.widgetConfig).toBeDefined();
    expect(vm.widgetConfig.chartTitle).toBe("GPU KV cache utilization");
    expect(vm.widgetConfig.chartId).toBe("gpu-kv-cache-utilization-widget-id");
  });

  it("computes getDataArgs correctly", () => {
    const wrapper = mountComponent({
      workloadId: "test-workload-id",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    });

    const vm = wrapper.vm as InstanceType<typeof GpuKvCacheUtilizationWidget>;
    expect(vm.getDataArgs).toEqual(["test-workload-id"]);
  });

  it("handles empty workloadId correctly", () => {
    const wrapper = mountComponent({
      workloadId: "",
      filterByDates: {
        dateStart: new Date("2024-03-01T00:00:00.000Z"),
        dateEnd: new Date("2024-03-01T01:00:00.000Z"),
        label: "Last 1 hour",
      },
    });

    const vm = wrapper.vm as InstanceType<typeof GpuKvCacheUtilizationWidget>;
    expect(vm.getDataArgs).toEqual([""]);
  });

  describe("pod functionality", () => {
    it("uses pod widget config when podId is provided", () => {
      const wrapper = mountComponent({
        workloadId: "test-workload-id",
        podId: "test-pod-id",
        filterByDates: {
          dateStart: new Date("2024-03-01T00:00:00.000Z"),
          dateEnd: new Date("2024-03-01T01:00:00.000Z"),
          label: "Last 1 hour",
        },
      });

      const vm = wrapper.vm as InstanceType<typeof GpuKvCacheUtilizationWidget>;
      expect(vm.widgetConfig).toBeDefined();
      expect(vm.widgetConfig.chartTitle).toBe("GPU KV cache utilization");
      expect(vm.widgetConfig.chartId).toBe("gpu-kv-cache-utilization-widget-id");
    });

    it("passes correct data args when both workloadId and podId are provided", () => {
      const wrapper = mountComponent({
        workloadId: "test-workload-id",
        podId: "test-pod-id",
        filterByDates: {
          dateStart: new Date("2024-03-01T00:00:00.000Z"),
          dateEnd: new Date("2024-03-01T01:00:00.000Z"),
          label: "Last 1 hour",
        },
      });

      const vm = wrapper.vm as InstanceType<typeof GpuKvCacheUtilizationWidget>;
      expect(vm.getDataArgs).toEqual(["test-workload-id", "test-pod-id"]);
    });

    it("passes correct props to GenericMetricWidget when podId is provided", () => {
      const props = {
        workloadId: "test-workload-id",
        podId: "test-pod-id",
        filterByDates: {
          dateStart: new Date("2024-03-01T00:00:00.000Z"),
          dateEnd: new Date("2024-03-01T01:00:00.000Z"),
          label: "Last 1 hour",
        },
      };

      const wrapper = mountComponent(props);
      const genericMetricWidget = wrapper.findComponent({ name: "generic-metric-widget" });

      expect(genericMetricWidget.props("filterByDates")).toEqual(props.filterByDates);
      expect(genericMetricWidget.props("getDataArgs")).toEqual([props.workloadId, props.podId]);
      expect(genericMetricWidget.props("config")).toBeDefined();
    });

    it("handles empty podId correctly", () => {
      const wrapper = mountComponent({
        workloadId: "test-workload-id",
        podId: "",
        filterByDates: {
          dateStart: new Date("2024-03-01T00:00:00.000Z"),
          dateEnd: new Date("2024-03-01T01:00:00.000Z"),
          label: "Last 1 hour",
        },
      });

      const vm = wrapper.vm as InstanceType<typeof GpuKvCacheUtilizationWidget>;
      expect(vm.getDataArgs).toEqual(["test-workload-id"]);
    });
  });
});
