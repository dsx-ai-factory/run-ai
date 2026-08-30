import type { MetricsResponse } from "@/swagger-models/workloads-client";

export const workloadPodMetricsMockResponse: MetricsResponse = {
  measurements: [
    {
      type: "GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU",
      values: [
        {
          value: "0",
          timestamp: "2024-05-15T10:57:55.443Z",
        },
        {
          value: "0",
          timestamp: "2024-05-15T10:58:10.62Z",
        },
        {
          value: "100",
          timestamp: "2024-05-15T10:58:25.797Z",
        },
        {
          value: "100",
          timestamp: "2024-05-15T10:58:40.974Z",
        },
      ],
    },
    {
      type: "GPU_SM_ACTIVITY_PER_GPU",
      values: [
        {
          value: "0.010799999999999999",
          timestamp: "2024-05-15T10:57:55.443Z",
        },
        {
          value: "0.010799999999999999",
          timestamp: "2024-05-15T10:58:10.62Z",
        },
        {
          value: "93.38159999999999",
          timestamp: "2024-05-15T10:58:25.797Z",
        },
        {
          value: "93.38159999999999",
          timestamp: "2024-05-15T10:58:40.974Z",
        },
      ],
    },
    {
      type: "GPU_SM_OCCUPANCY_PER_GPU",
      values: [
        {
          value: "0.0101",
          timestamp: "2024-05-15T10:57:55.443Z",
        },
        {
          value: "0.0101",
          timestamp: "2024-05-15T10:58:10.62Z",
        },
        {
          value: "86.79390000000001",
          timestamp: "2024-05-15T10:58:25.797Z",
        },
        {
          value: "86.79390000000001",
          timestamp: "2024-05-15T10:58:40.974Z",
        },
      ],
    },
    {
      type: "GPU_TENSOR_ACTIVITY_PER_GPU",
      values: [
        {
          value: "0",
          timestamp: "2024-05-15T10:57:55.443Z",
        },
        {
          value: "0",
          timestamp: "2024-05-15T10:58:10.62Z",
        },
        {
          value: "0",
          timestamp: "2024-05-15T10:58:25.797Z",
        },
        {
          value: "0",
          timestamp: "2024-05-15T10:58:40.974Z",
        },
      ],
    },
    {
      type: "GPU_FP64_ENGINE_ACTIVITY_PER_GPU",
      values: [
        {
          value: "0",
          timestamp: "2024-05-15T10:57:55.443Z",
        },
        {
          value: "0",
          timestamp: "2024-05-15T10:58:10.62Z",
        },
        {
          value: "0",
          timestamp: "2024-05-15T10:58:25.797Z",
        },
        {
          value: "0",
          timestamp: "2024-05-15T10:58:40.974Z",
        },
      ],
    },
    {
      type: "GPU_FP32_ENGINE_ACTIVITY_PER_GPU",
      values: [
        {
          value: "0.00019999999999999998",
          timestamp: "2024-05-15T10:57:55.443Z",
        },
        {
          value: "0.00019999999999999998",
          timestamp: "2024-05-15T10:58:10.62Z",
        },
        {
          value: "2.0946",
          timestamp: "2024-05-15T10:58:25.797Z",
        },
        {
          value: "2.0946",
          timestamp: "2024-05-15T10:58:40.974Z",
        },
      ],
    },
    {
      type: "GPU_FP16_ENGINE_ACTIVITY_PER_GPU",
      values: [
        {
          value: "0",
          timestamp: "2024-05-15T10:57:55.443Z",
        },
        {
          value: "0",
          timestamp: "2024-05-15T10:58:10.62Z",
        },
        {
          value: "0",
          timestamp: "2024-05-15T10:58:25.797Z",
        },
        {
          value: "0",
          timestamp: "2024-05-15T10:58:40.974Z",
        },
      ],
    },
  ],
};
