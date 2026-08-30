import type { TelemetryResponse } from "@/swagger-models/workloads-client";

export const mockPendingByNodepoolTelemetry: TelemetryResponse = {
  type: "WORKLOADS_COUNT",
  timestamp: "2024-03-01T00:00:00.000Z",
  values: [
    {
      value: "3",
      groups: [
        { key: "ClusterId", value: "cluster-123", name: "cluster-123" },
        { key: "RequestedNodepools", value: "default", name: "default" },
      ],
    },
    {
      value: "5",
      groups: [
        { key: "ClusterId", value: "cluster-123", name: "cluster-123" },
        { key: "RequestedNodepools", value: "gpu-pool", name: "gpu-pool" },
      ],
    },
  ],
};

export const mockEmptyTelemetry: TelemetryResponse = {
  type: "WORKLOADS_COUNT",
  timestamp: "2024-03-01T00:00:00.000Z",
  values: [],
};
