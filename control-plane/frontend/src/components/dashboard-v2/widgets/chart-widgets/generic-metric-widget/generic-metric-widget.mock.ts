import type { MetricsResponse } from "@/swagger-models/workloads-client";
import { NodeMetricType } from "@/swagger-models/cluster-service-client";

export const mockMetricsResponse: MetricsResponse = {
  measurements: [
    {
      type: NodeMetricType.GpuUtilizationPerGpu,
      values: [
        { timestamp: "2024-03-19T09:39:00.346Z", value: "10" },
        { timestamp: "2024-03-19T10:39:00.346Z", value: "20" },
        { timestamp: "2024-03-19T11:39:00.346Z", value: "30" },
      ],
    },
    {
      type: NodeMetricType.GpuUtilizationPerGpu,
      values: [
        { timestamp: "2024-03-19T09:39:00.346Z", value: "15" },
        { timestamp: "2024-03-19T10:39:00.346Z", value: "25" },
        { timestamp: "2024-03-19T11:39:00.346Z", value: "35" },
      ],
    },
  ],
};
