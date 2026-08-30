import { describe, expect, test } from "vitest";

import { MetricsType } from "@/swagger-models/cluster-service-client";
import {
  GPU_ALLOCATION_RENDER_ORDER,
  getGpuAllocationSeriesLayout,
} from "./gpu-allocation-time-range-widget/gpu-allocation-time-range-widget.vue";
import {
  ESeriesType,
  GPU_DISTRIBUTION_RENDER_ORDER,
} from "./gpu-utilization-distribution-time-range-widget/gpu-utilization-distribution-time-range-widget.vue";

describe("node-pool time-series parity", () => {
  test("renders total capacity independently above the allocated/unallocated stack", () => {
    expect(GPU_ALLOCATION_RENDER_ORDER).toEqual([
      MetricsType.TotalGpu,
      MetricsType.AllocatedGpu,
      MetricsType.UnallocatedGpu,
    ]);
    expect(getGpuAllocationSeriesLayout(MetricsType.TotalGpu)).toEqual({ z: 1 });
    expect(getGpuAllocationSeriesLayout(MetricsType.AllocatedGpu)).toEqual({ stack: "allocation", z: 2 });
    expect(getGpuAllocationSeriesLayout(MetricsType.UnallocatedGpu)).toEqual({ stack: "allocation", z: 2 });
  });

  test("reverses Highcharts normal-stack insertion order for ECharts", () => {
    expect(GPU_DISTRIBUTION_RENDER_ORDER).toEqual([
      ESeriesType.ZeroToTwentyFive,
      ESeriesType.TwentyFiveToFifty,
      ESeriesType.FiftyToSeventyFive,
      ESeriesType.SeventyFiveToOneHundred,
    ]);
  });
});
