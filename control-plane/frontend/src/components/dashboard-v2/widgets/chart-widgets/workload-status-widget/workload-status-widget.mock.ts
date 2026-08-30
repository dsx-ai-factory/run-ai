import { EWorkloadTelemetryGroupBy } from "@/models/workload.model";

export const WorkloadStatusWidgetMock = {
  type: "WORKLOADS_COUNT",
  timestamp: "2024-04-01T14:41:39.079386916Z",
  values: [
    {
      value: "1",
      groups: [
        {
          key: "ProjectId",
          value: "1",
          name: "suricata",
        },
        {
          key: "Phase",
          value: "Failed",
        },
      ],
    },
    {
      value: "6",
      groups: [
        {
          key: "ProjectId",
          value: "1",
          name: "suricata",
        },
        {
          key: "Phase",
          value: "Running",
        },
      ],
    },
    {
      value: "7",
      groups: [
        {
          key: "ProjectId",
          value: "2",
          name: "griffin",
        },
        {
          key: "Phase",
          value: "Running",
        },
      ],
    },
    {
      value: "1",
      groups: [
        {
          key: "ProjectId",
          value: "3",
          name: "lemur",
        },
        {
          key: "Phase",
          value: "Pending",
        },
      ],
    },
    {
      value: "21",
      groups: [
        {
          key: "ProjectId",
          value: "3",
          name: "lemur",
        },
        {
          key: "Phase",
          value: "Running",
        },
      ],
    },
  ],
};

export const WorkloadStatusWidgetMockFilterByNodepool = {
  type: "WORKLOADS_COUNT",
  timestamp: "2024-04-01T14:42:28.470771436Z",
  values: [
    {
      value: "1",
      groups: [
        {
          key: "ProjectId",
          value: "2",
          name: "griffin",
        },
        {
          key: "Phase",
          value: "Running",
        },
      ],
    },
  ],
};

export const mockDataWithMissingPhases = {
  values: [
    {
      value: "5",
      groups: [
        { key: EWorkloadTelemetryGroupBy.ClusterId, value: "cluster-1", name: "Test Cluster" },
        { key: EWorkloadTelemetryGroupBy.Phase, value: "Running", name: "Running" },
      ],
    },
  ],
};
