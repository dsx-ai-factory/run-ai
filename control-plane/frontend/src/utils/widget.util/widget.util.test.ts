import { describe, expect, test } from "vitest";
//util
import { widgetUtil } from "@/utils/widget.util";
//model
import { EWorkloadTelemetryGroupBy } from "@/models/workload.model";
//mock
import {
  appInitialDataMock,
  departmentPhaseTelemetryMockResponse,
  lowestAllocationRatioProjectNames,
  mockMetricData,
  projectsPhaseTelemetryMockResponse,
  sumAllTelemetryValuesMockResponse,
  telemetryResponseToEnrich,
  workloadsTelemetryResponseGroupByProjectId,
} from "@/utils/widget.util/widget.util.mock";

import type { IEChartsSeriesOption } from "@/utils/widget.util/widget.util";
import type { TelemetryResponseValuesInner } from "@/swagger-models/workloads-client";
import type { MetricsData } from "@/models/metrics.model";

describe("widgetUtil", () => {
  describe("aggregateTelemetryGroupedData", () => {
    departmentPhaseTelemetryMockResponse.forEach(({ description, telemetryValues, expected }) => {
      test(description, () => {
        const result: IEChartsSeriesOption[] = widgetUtil.aggregateTelemetryGroupedData(
          telemetryValues,
          [EWorkloadTelemetryGroupBy.DepartmentId, EWorkloadTelemetryGroupBy.Phase],
          true,
        );
        expect(result).toEqual(expected);
      });
    });

    test("should include prevGroupByKey and prevEntityId in the series name", () => {
      const telemetryValues: TelemetryResponseValuesInner[] = [
        {
          value: "10",
          groups: [
            { key: "DepartmentId", value: "1" },
            { key: "Phase", value: "Pending" },
          ],
        },
        {
          value: "5",
          groups: [
            { key: "DepartmentId", value: "1" },
            { key: "Phase", value: "Running" },
          ],
        },
      ];

      const result: IEChartsSeriesOption[] = widgetUtil.aggregateTelemetryGroupedData(
        telemetryValues,
        [EWorkloadTelemetryGroupBy.DepartmentId, EWorkloadTelemetryGroupBy.Phase],
        true,
        "DepartmentId",
        "45000",
      );

      result.forEach((series) => {
        const meta = (series as IEChartsSeriesOption & { meta: { groupByKey: string; entityId: string } }).meta;
        expect(meta.groupByKey).toBe("DepartmentId");
        expect(meta.entityId).toBe("45000");
      });
    });
  });

  describe("enrichTelemetryResponseByKeyValue", () => {
    test("should enrich TelemetryResponseValuesInner with a new key-value pair in groups", () => {
      const keyToEnrich = "ClusterId";
      const value = "barak-chart-1709102860";
      const key = "name";

      const result: TelemetryResponseValuesInner[] = widgetUtil.enrichTelemetryResponseByKeyValue(
        telemetryResponseToEnrich,
        keyToEnrich,
        value,
        key,
      );

      expect(result).toEqual([
        {
          value: "2.00",
          groups: [
            {
              key: "ClusterId",
              value: "eafdce8e-9f20-4c84-bc2f-6db5a174e861",
              name: "barak-chart-1709102860",
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
              name: "barak-chart-1709102860",
            },
            {
              key: "Type",
              value: "Workspace",
            },
          ],
        },
      ]);
    });

    test("should not enrich", () => {
      const keyToEnrich = "ClusterId_";
      const value = "barak-chart-1709102860";
      const key = "name";

      const result: TelemetryResponseValuesInner[] = widgetUtil.enrichTelemetryResponseByKeyValue(
        telemetryResponseToEnrich,
        keyToEnrich,
        value,
        key,
      );

      expect(result).toEqual(telemetryResponseToEnrich);
    });
  });

  describe("sumAllTelemetryValues", () => {
    test("should return sum of all telemetry values", () => {
      const result = widgetUtil.sumAllTelemetryValues(sumAllTelemetryValuesMockResponse);
      expect(result).toBe(47);
    });

    test("should return 0 if telemetry values are empty", () => {
      const result = widgetUtil.sumAllTelemetryValues([]);
      expect(result).toBe(0);
    });
  });

  describe("filterTopAllocationRatioProjects", () => {
    // Test case for filtering top allocation ratio projects
    test("should filter top allocation ratio projects correctly", () => {
      const filteredData: MetricsData[] = widgetUtil.filterTopAllocationRatioProjects(
        mockMetricData as MetricsData[],
        2,
      );

      //should return 3 departments
      expect(filteredData).toHaveLength(3);

      //should return 2 projects for each department
      filteredData.forEach((department: MetricsData) => expect(department.current.projectResources).toHaveLength(2));

      //should not return the lowest allocation ratio projects for each department
      filteredData.forEach((department: MetricsData) => {
        department.current.projectResources.forEach((project, index) => {
          expect(project.projectName).not.toEqual(lowestAllocationRatioProjectNames[index]);
        });
      });
    });

    test("should handle cases where the number of top projects to filter is greater than the total number of projects available", () => {
      const filteredData: MetricsData[] = widgetUtil.filterTopAllocationRatioProjects(
        mockMetricData as MetricsData[],
        10,
      );

      expect(filteredData).toHaveLength(3);
      filteredData.forEach((department: MetricsData) => {
        expect(department.current.projectResources).toHaveLength(mockMetricData[0].current.projectResources.length);
      });
    });

    test("should handle the case when no projects at all", () => {
      const filteredData: MetricsData[] = widgetUtil.filterTopAllocationRatioProjects(
        appInitialDataMock as unknown as MetricsData[],
        2,
      );
      expect(filteredData).toStrictEqual(appInitialDataMock);
    });
  });

  describe("limitProjectsTelemetryResponse", () => {
    test("should limit projects to maximum displayed projects", () => {
      const projects = [...projectsPhaseTelemetryMockResponse];

      const result: TelemetryResponseValuesInner[] = widgetUtil.limitProjectsTelemetryResponse(
        projects as TelemetryResponseValuesInner[],
        1,
      );

      result.forEach((project) => {
        const groups = (project as TelemetryResponseValuesInner).groups;
        expect(groups?.[0]?.name).toEqual("default-project-1");
      });
    });
    test("should handle empty projects", () => {
      const result: TelemetryResponseValuesInner[] = widgetUtil.limitProjectsTelemetryResponse([], 3);
      expect(result).toEqual([]);
    });

    test("should limit workloadsTelemetryResponseGroupByProjectId to maximum displayed projects", () => {
      const response = [...workloadsTelemetryResponseGroupByProjectId];

      const result: TelemetryResponseValuesInner[] = widgetUtil.limitProjectsTelemetryResponse(
        response as TelemetryResponseValuesInner[],
        5,
      );
      expect(result).toHaveLength(20); // 5 projects * 4 buckets each
    });

    test("should limit workloadsTelemetryResponseGroupByProjectId to top 5 projects by total workload count", () => {
      const response = [...workloadsTelemetryResponseGroupByProjectId];

      const result: TelemetryResponseValuesInner[] = widgetUtil.limitProjectsTelemetryResponse(
        response as TelemetryResponseValuesInner[],
        5,
      );

      // Extract project names from the result
      const projectNames = new Set<string>();
      result.forEach((item) => {
        const projectGroup = item.groups?.find((group) => group.key === "ProjectId");
        if (projectGroup?.name) {
          projectNames.add(projectGroup.name);
        }
      });

      // Should have exactly 5 projects
      expect(projectNames.size).toBe(5);

      // Each project should have all 4 buckets
      projectNames.forEach((projectName) => {
        const projectBuckets = result.filter((item) => {
          const projectGroup = item.groups?.find((group) => group.key === "ProjectId");
          return projectGroup?.name === projectName;
        });
        expect(projectBuckets).toHaveLength(4); // Each project should have 4 buckets
      });
    });

    test("should prioritize projects with maximum pending time across all buckets", () => {
      const response = [...workloadsTelemetryResponseGroupByProjectId];

      const result: TelemetryResponseValuesInner[] = widgetUtil.limitProjectsTelemetryResponse(
        response as TelemetryResponseValuesInner[],
        5,
      );

      // Calculate total pending time for each project in the result
      const projectTotals: { [projectName: string]: number } = {};
      result.forEach((item) => {
        const projectGroup = item.groups?.find((group) => group.key === "ProjectId");
        if (projectGroup?.name) {
          const value = parseFloat(item.value || "0");
          projectTotals[projectGroup.name] = (projectTotals[projectGroup.name] || 0) + value;
        }
      });

      // The top 5 projects should be those with the highest total pending time
      const projectNames = Object.keys(projectTotals);
      const totals = Object.values(projectTotals);

      const maxTotal = Math.max(...totals);
      expect(maxTotal).toBe(3);

      totals.forEach((total) => {
        expect(total).toBe(maxTotal);
      });

      expect(projectNames).toHaveLength(5);
      expect(totals).toHaveLength(5);
      expect(totals.every((total) => total === maxTotal)).toBe(true);

      // Verify these are the expected top projects (the 5 projects with total = 3 in our reduced mock data)
      const expectedTopProjects = [
        "test-project-7",
        "test-project-11",
        "test-project-12",
        "test-project-15",
        "test-project-16",
      ];

      projectNames.forEach((projectName) => {
        expect(expectedTopProjects).toContain(projectName);
      });
    });
  });

  describe("combineDistributedIntoTraining", () => {
    test("should combine distributed into training when both are present", () => {
      const data: TelemetryResponseValuesInner[] = [
        {
          value: "10",
          groups: [{ key: "Type", value: "Distributed" }],
        },
        {
          value: "5",
          groups: [{ key: "Type", value: "Training" }],
        },
      ];

      const result = widgetUtil.combineDistributedTelemetryIntoTraining(data);

      expect(result).toEqual([
        {
          value: "15",
          groups: [{ key: "Type", value: "Training" }],
        },
      ]);
    });

    test("should convert distributed to training when training is not present", () => {
      const data: TelemetryResponseValuesInner[] = [
        {
          value: "10",
          groups: [{ key: "Type", value: "Distributed" }],
        },
      ];

      const result = widgetUtil.combineDistributedTelemetryIntoTraining(data);

      expect(result).toEqual([
        {
          value: "10",
          groups: [{ key: "Type", value: "Training" }],
        },
      ]);
    });

    test("should not change data when neither distributed nor training is present", () => {
      const data: TelemetryResponseValuesInner[] = [
        {
          value: "10",
          groups: [{ key: "Type", value: "Other" }],
        },
      ];

      const result = widgetUtil.combineDistributedTelemetryIntoTraining(data);

      expect(result).toEqual(data);
    });
  });

  describe("gpuUtilizationPerNodeParser", () => {
    test("should parse GPU utilization data correctly", () => {
      const metricsResponse = {
        measurements: [
          {
            type: "GpuUtilization",
            values: [{ timestamp: "1", value: "50" }],
          },
          {
            type: "GpuUtilizationPerGpu",
            labels: { gpu: "0" },
            values: [{ timestamp: "1", value: "60" }],
          },
          {
            type: "GpuUtilizationPerGpu",
            labels: { gpu: "1" },
            values: [{ timestamp: "1", value: "55" }],
          },
        ],
      };

      const expected = {
        Average: [{ timestamp: "1", value: "50" }],
        "GPU-0": [{ timestamp: "1", value: "60" }],
        "GPU-1": [{ timestamp: "1", value: "55" }],
      };

      const result = widgetUtil.gpuUtilizationPerNodeParser(metricsResponse, "GpuUtilization");
      expect(result).toEqual(expected);
    });

    test("should handle empty measurements", () => {
      const metricsResponse = { measurements: [] };
      const expected = {};
      const result = widgetUtil.gpuUtilizationPerNodeParser(metricsResponse, "GpuUtilization");
      expect(result).toEqual(expected);
    });

    test("should merge values from duplicate GPU labels", () => {
      const metricsResponse = {
        measurements: [
          {
            type: "GpuUtilization",
            values: [{ timestamp: "1", value: "50" }],
          },
          {
            type: "GpuUtilization",
            values: [{ timestamp: "2", value: "55" }],
          },
          {
            type: "GpuUtilizationPerGpu",
            labels: { gpu: "0" },
            values: [{ timestamp: "1", value: "60" }],
          },
          {
            type: "GpuUtilizationPerGpu",
            labels: { gpu: "0" },
            values: [{ timestamp: "2", value: "65" }],
          },
        ],
      };

      const result = widgetUtil.gpuUtilizationPerNodeParser(metricsResponse, "GpuUtilization");
      expect(result.Average).toEqual(
        expect.arrayContaining([
          { timestamp: "1", value: "50" },
          { timestamp: "2", value: "55" },
        ]),
      );
      expect(result.Average).toHaveLength(2);
      expect(result["GPU-0"]).toEqual(
        expect.arrayContaining([
          { timestamp: "1", value: "60" },
          { timestamp: "2", value: "65" },
        ]),
      );
      expect(result["GPU-0"]).toHaveLength(2);
    });
  });
});
