import { EWorkloadTelemetryGroupBy } from "@/models/workload.model";
import { Phase } from "@/swagger-models/workloads-client";
import type { TelemetryResponseValuesInner } from "@/swagger-models/workloads-client";

export const departmentPhaseTelemetryMockResponse = [
  {
    description: "should aggregate grouped data for charts",
    telemetryValues: [
      {
        value: "5",
        groups: [
          { key: EWorkloadTelemetryGroupBy.DepartmentId, value: "123", name: "default" },
          { key: EWorkloadTelemetryGroupBy.Phase, value: Phase.Running },
        ],
      },
      {
        value: "5",
        groups: [
          { key: EWorkloadTelemetryGroupBy.DepartmentId, value: "321", name: "dep-1" },
          { key: EWorkloadTelemetryGroupBy.Phase, value: Phase.Running },
        ],
      },
      {
        value: "3",
        groups: [
          { key: EWorkloadTelemetryGroupBy.DepartmentId, value: "123", name: "default" },
          { key: EWorkloadTelemetryGroupBy.Phase, value: Phase.Pending },
        ],
      },
    ],
    expected: [
      {
        name: Phase.Running,
        meta: {},
        data: [
          { name: "default", y: 5, drilldown: true, meta: { entityId: "123", groupByKey: "DepartmentId" } },
          { name: "dep-1", y: 5, drilldown: true, meta: { entityId: "321", groupByKey: "DepartmentId" } },
        ],
      },
      {
        name: Phase.Pending,
        meta: {},
        data: [{ name: "default", y: 3, drilldown: true, meta: { entityId: "123", groupByKey: "DepartmentId" } }],
      },
    ],
  },
  {
    description: "should filter out zero values",
    telemetryValues: [
      {
        value: "5",
        groups: [
          { key: EWorkloadTelemetryGroupBy.DepartmentId, value: "123", name: "default" },
          { key: EWorkloadTelemetryGroupBy.Phase, value: Phase.Running },
        ],
      },
      {
        value: "0",
        groups: [
          { key: EWorkloadTelemetryGroupBy.DepartmentId, value: "321", name: "dep-1" },
          { key: EWorkloadTelemetryGroupBy.Phase, value: Phase.Running },
        ],
      },
      {
        value: "0.5",
        groups: [
          { key: EWorkloadTelemetryGroupBy.DepartmentId, value: "123", name: "default" },
          { key: EWorkloadTelemetryGroupBy.Phase, value: Phase.Pending },
        ],
      },
    ],
    expected: [
      {
        meta: {},
        name: Phase.Running,
        data: [{ name: "default", y: 5, drilldown: true, meta: { entityId: "123", groupByKey: "DepartmentId" } }],
      },
      {
        meta: {},
        name: Phase.Pending,
        data: [{ name: "default", y: 0.5, drilldown: true, meta: { entityId: "123", groupByKey: "DepartmentId" } }],
      },
    ],
  },
];

export const telemetryResponseToEnrich: TelemetryResponseValuesInner[] = [
  {
    value: "2.00",
    groups: [
      {
        key: "ClusterId",
        value: "eafdce8e-9f20-4c84-bc2f-6db5a174e861",
      },
      {
        key: "Type",
        value: "Training",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "ClusterId",
        value: "eafdce8e-9f20-4c84-bc2f-6db5a174e861",
      },
      {
        key: "Type",
        value: "Workspace",
      },
    ],
  },
];

export const sumAllTelemetryValuesMockResponse = [
  {
    value: "32",
    groups: [
      {
        key: "ClusterId",
        value: "eafdce8e-9f20-4c84-bc2f-6db5a174e861",
      },
      {
        key: "Nodepool",
        value: "default",
      },
    ],
  },
  {
    value: "15",
    groups: [
      {
        key: "ClusterId",
        value: "eafdce8e-9f20-4c84-bc2f-6db5a174e861",
      },
      {
        key: "Nodepool",
        value: "np-2",
      },
    ],
  },
];

export const mockMetricData = [
  {
    metadata: { departmentId: 1, departmentName: "Department 1" },
    current: {
      projectResources: [
        { projectName: "Project 1", gpu: { quota: 10, allocated: 5 } },
        { projectName: "Project 2", gpu: { quota: 20, allocated: 15 } },
        { projectName: "Project 3", gpu: { quota: 30, allocated: 10 } },
      ],
    },
  },
  {
    metadata: { departmentId: 2, departmentName: "Department 2" },
    current: {
      projectResources: [
        { projectName: "Project 4", gpu: { quota: 20, allocated: 10 } },
        { projectName: "Project 5", gpu: { quota: 15, allocated: 5 } },
        { projectName: "Project 6", gpu: { quota: 25, allocated: 20 } },
      ],
    },
  },
  {
    metadata: { departmentId: 3, departmentName: "Department 3" },
    current: {
      projectResources: [
        { projectName: "Project 7", gpu: { quota: 30, allocated: 20 } },
        { projectName: "Project 8", gpu: { quota: 0, allocated: 5 } },
        { projectName: "Project 9", gpu: { quota: 25, allocated: 20 } },
      ],
    },
  },
];

export const lowestAllocationRatioProjectNames: string[] = ["Project 3", "Project 5", "Project 8"];

export const appInitialDataMock = [
  {
    metadata: {
      departmentId: 1,
      departmentName: "default",
    },
    current: {
      resources: [
        {
          numberOfPendingWorkloads: 0,
          gpu: {
            quota: -1,
            allocated: 0,
          },
          cpu: {
            quota: -1,
            allocated: 0,
          },
          memory: {
            quota: -1,
            allocated: 0,
          },
          nodepoolName: "default",
        },
      ],
      projectResources: [],
    },
  },
];

export const projectsPhaseTelemetryMockResponse = [
  {
    value: "1",
    groups: [
      {
        key: "ProjectId",
        value: "3",
        name: "default-project-1",
      },
      {
        key: "Phase",
        value: "Completed",
      },
    ],
  },
  {
    value: "13",
    groups: [
      {
        key: "ProjectId",
        value: "3",
        name: "default-project-1",
      },
      {
        key: "Phase",
        value: "Failed",
      },
    ],
  },
  {
    value: "9",
    groups: [
      {
        key: "ProjectId",
        value: "3",
        name: "default-project-1",
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
        name: "default-project-1",
      },
      {
        key: "Phase",
        value: "Stopped",
      },
    ],
  },
  {
    value: "7",
    groups: [
      {
        key: "ProjectId",
        value: "4",
        name: "default-project-2",
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
        value: "4",
        name: "default-project-2",
      },
      {
        key: "Phase",
        value: "Running",
      },
    ],
  },
];

export const workloadsTelemetryResponseGroupByProjectId = [
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(0)-(5h)",
        name: "(0)-(5h)",
      },
      {
        key: "ProjectId",
        value: "4500002",
        name: "test-project-2",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(5h)-(10h)",
        name: "(5h)-(10h)",
      },
      {
        key: "ProjectId",
        value: "4500002",
        name: "test-project-2",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(10h)-(15h)",
        name: "(10h)-(15h)",
      },
      {
        key: "ProjectId",
        value: "4500002",
        name: "test-project-2",
      },
    ],
  },
  {
    value: "2",
    groups: [
      {
        key: "Bucket",
        value: "(15h)+",
        name: "(15h)+",
      },
      {
        key: "ProjectId",
        value: "4500002",
        name: "test-project-2",
      },
    ],
  },
  // test-project-7: total = 3 (should be selected - top priority)
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(0)-(5h)",
        name: "(0)-(5h)",
      },
      {
        key: "ProjectId",
        value: "4500007",
        name: "test-project-7",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(5h)-(10h)",
        name: "(5h)-(10h)",
      },
      {
        key: "ProjectId",
        value: "4500007",
        name: "test-project-7",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(10h)-(15h)",
        name: "(10h)-(15h)",
      },
      {
        key: "ProjectId",
        value: "4500007",
        name: "test-project-7",
      },
    ],
  },
  {
    value: "3",
    groups: [
      {
        key: "Bucket",
        value: "(15h)+",
        name: "(15h)+",
      },
      {
        key: "ProjectId",
        value: "4500007",
        name: "test-project-7",
      },
    ],
  },
  // test-project-11: total = 3 (should be selected)
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(0)-(5h)",
        name: "(0)-(5h)",
      },
      {
        key: "ProjectId",
        value: "4500011",
        name: "test-project-11",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(5h)-(10h)",
        name: "(5h)-(10h)",
      },
      {
        key: "ProjectId",
        value: "4500011",
        name: "test-project-11",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(10h)-(15h)",
        name: "(10h)-(15h)",
      },
      {
        key: "ProjectId",
        value: "4500011",
        name: "test-project-11",
      },
    ],
  },
  {
    value: "3",
    groups: [
      {
        key: "Bucket",
        value: "(15h)+",
        name: "(15h)+",
      },
      {
        key: "ProjectId",
        value: "4500011",
        name: "test-project-11",
      },
    ],
  },
  // test-project-12: total = 3 (should be selected)
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(0)-(5h)",
        name: "(0)-(5h)",
      },
      {
        key: "ProjectId",
        value: "4500012",
        name: "test-project-12",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(5h)-(10h)",
        name: "(5h)-(10h)",
      },
      {
        key: "ProjectId",
        value: "4500012",
        name: "test-project-12",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(10h)-(15h)",
        name: "(10h)-(15h)",
      },
      {
        key: "ProjectId",
        value: "4500012",
        name: "test-project-12",
      },
    ],
  },
  {
    value: "3",
    groups: [
      {
        key: "Bucket",
        value: "(15h)+",
        name: "(15h)+",
      },
      {
        key: "ProjectId",
        value: "4500012",
        name: "test-project-12",
      },
    ],
  },
  // test-project-15: total = 3 (should be selected)
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(0)-(5h)",
        name: "(0)-(5h)",
      },
      {
        key: "ProjectId",
        value: "4500015",
        name: "test-project-15",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(5h)-(10h)",
        name: "(5h)-(10h)",
      },
      {
        key: "ProjectId",
        value: "4500015",
        name: "test-project-15",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(10h)-(15h)",
        name: "(10h)-(15h)",
      },
      {
        key: "ProjectId",
        value: "4500015",
        name: "test-project-15",
      },
    ],
  },
  {
    value: "3",
    groups: [
      {
        key: "Bucket",
        value: "(15h)+",
        name: "(15h)+",
      },
      {
        key: "ProjectId",
        value: "4500015",
        name: "test-project-15",
      },
    ],
  },
  // test-project-16: total = 3 (should be selected)
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(0)-(5h)",
        name: "(0)-(5h)",
      },
      {
        key: "ProjectId",
        value: "4500016",
        name: "test-project-16",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(5h)-(10h)",
        name: "(5h)-(10h)",
      },
      {
        key: "ProjectId",
        value: "4500016",
        name: "test-project-16",
      },
    ],
  },
  {
    value: "0",
    groups: [
      {
        key: "Bucket",
        value: "(10h)-(15h)",
        name: "(10h)-(15h)",
      },
      {
        key: "ProjectId",
        value: "4500016",
        name: "test-project-16",
      },
    ],
  },
  {
    value: "3",
    groups: [
      {
        key: "Bucket",
        value: "(15h)+",
        name: "(15h)+",
      },
      {
        key: "ProjectId",
        value: "4500016",
        name: "test-project-16",
      },
    ],
  },
];
