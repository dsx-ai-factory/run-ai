import { vi, describe, it, expect } from "vitest";
import { NodeStatusEnum, type Node, type MetricsResponse } from "@/swagger-models/cluster-service-client";
import { NrtHealthyConditionType } from "@/models/node.model";
import { nodeUtil } from "@/utils/node.util/node.util";

vi.mock("@/utils/docs.util/docs.util", () => ({
  docsUtil: {
    getNumaAwareSchedulingHTMLLink: () =>
      'For more information see the <a href="https://docs.example.test/numa">Configuring NUMA-Aware Scheduling guide.</a>',
  },
}));

describe("getStatusColOptions", () => {
  it("renders the backend-provided message plus the NUMA guide link for a missing NRT prerequisite", () => {
    const node = {
      status: NodeStatusEnum.MissingPrerequisites,
      conditions: [
        {
          type: NrtHealthyConditionType,
          reason: "PolicyNotEnforcing",
          message:
            "The following prerequisites are not met on the node: Kubelet Topology Manager Policy misconfigured. As a result, NUMA-aware scheduling may be impacted.",
        },
      ],
    } as Node;

    const result = nodeUtil.getStatusColOptions(node);

    expect(result.status).toBe("Missing Prerequisites");
    expect(result.tooltipText).toBe(
      "The following prerequisites are not met on the node: Kubelet Topology Manager Policy misconfigured. As a result, NUMA-aware scheduling may be impacted. " +
        'For more information see the <a href="https://docs.example.test/numa">Configuring NUMA-Aware Scheduling guide.</a><br>',
    );
  });

  it("omits the NRT condition from the tooltip when it has no message", () => {
    const node = {
      status: NodeStatusEnum.MissingPrerequisites,
      conditions: [{ type: NrtHealthyConditionType, reason: "PolicyNotEnforcing" }],
    } as Node;

    const result = nodeUtil.getStatusColOptions(node);

    expect(result.tooltipText).toBe("");
  });

  it("still renders static messages for existing condition types", () => {
    const node = {
      status: NodeStatusEnum.MissingPrerequisites,
      conditions: [{ type: "DiskPressure", reason: "DiskPressure" }],
    } as Node;

    const result = nodeUtil.getStatusColOptions(node);

    expect(result.tooltipText).toBe("Disk pressure: Disk capacity is low.<br>");
  });
});

describe("parseNodeGpuMeasurements", () => {
  it("should correctly map resources to the relevant GPU using the label", () => {
    const metricsResponse: MetricsResponse = {
      measurements: [
        {
          type: "GPU_ALLOCATION_PER_GPU",
          labels: { gpu: "0" },
          values: [{ value: "1", timestamp: "2025-05-13T10:32:17Z" }],
        },
        {
          type: "GPU_MEMORY_ALLOCATION_BYTES_PER_GPU",
          labels: { gpu: "0" },
          values: [{ value: "12582912000", timestamp: "2025-05-13T10:32:17Z" }],
        },
        {
          type: "GPU_MEMORY_USAGE_BYTES_PER_GPU",
          labels: { gpu: "0" },
          values: [{ value: "12582912000", timestamp: "2025-05-13T10:32:17Z" }],
        },
        {
          type: "GPU_IDLE_TIME_SECONDS_PER_GPU",
          labels: { gpu: "0" },
          values: [{ value: "0", timestamp: "2025-05-13T10:32:17Z" }],
        },
      ],
    };

    const expectedOutput = [
      {
        allocatedCompute: 1,
        allocatedMemory: 12582912000,
        computeUtilization: 0,
        idleTime: "0",
        index: 0,
        memoryUtilization: 0,
        usedMemory: 12582912000,
      },
    ];

    const result = nodeUtil.parseNodeGpuMeasurements(metricsResponse);
    expect(result).toEqual(expectedOutput);
  });

  it("should handle multiple GPUs correctly", () => {
    const metricsResponse: MetricsResponse = {
      measurements: [
        {
          type: "GPU_ALLOCATION_PER_GPU",
          labels: { gpu: "0" },
          values: [{ value: "1", timestamp: "2025-05-13T10:32:17Z" }],
        },
        {
          type: "GPU_ALLOCATION_PER_GPU",
          labels: { gpu: "1" },
          values: [{ value: "2", timestamp: "2025-05-13T10:32:17Z" }],
        },
      ],
    };

    const expectedOutput = [
      {
        allocatedCompute: 1,
        allocatedMemory: 0,
        computeUtilization: 0,
        idleTime: "",
        index: 0,
        memoryUtilization: 0,
        usedMemory: 0,
      },
      {
        allocatedCompute: 2,
        allocatedMemory: 0,
        computeUtilization: 0,
        idleTime: "",
        index: 1,
        memoryUtilization: 0,
        usedMemory: 0,
      },
    ];

    const result = nodeUtil.parseNodeGpuMeasurements(metricsResponse);
    expect(result).toEqual(expectedOutput);
  });

  it("should handle missing labels gracefully", () => {
    const metricsResponse: MetricsResponse = {
      measurements: [
        {
          type: "GPU_ALLOCATION_PER_GPU",
          labels: {},
          values: [{ value: "1", timestamp: "2025-05-13T10:32:17Z" }],
        },
      ],
    };

    const result = nodeUtil.parseNodeGpuMeasurements(metricsResponse);
    expect(result).toEqual([]);
  });

  it("should handle missing values gracefully", () => {
    const metricsResponse: MetricsResponse = {
      measurements: [
        {
          type: "GPU_ALLOCATION_PER_GPU",
          labels: { gpu: "0" },
          values: [],
        },
      ],
    };

    const result = nodeUtil.parseNodeGpuMeasurements(metricsResponse);
    expect(result).toEqual([
      {
        allocatedCompute: 0,
        allocatedMemory: 0,
        computeUtilization: 0,
        idleTime: "",
        index: 0,
        memoryUtilization: 0,
        usedMemory: 0,
      },
    ]);
  });

  it("should handle additional measurement types", () => {
    const metricsResponse: MetricsResponse = {
      measurements: [
        {
          type: "GPU_UTILIZATION_PER_GPU",
          labels: { gpu: "0" },
          values: [{ value: "75", timestamp: "2025-05-13T10:32:17Z" }],
        },
      ],
    };

    const expectedOutput = [
      {
        allocatedCompute: 0,
        allocatedMemory: 0,
        computeUtilization: 75,
        idleTime: "",
        index: 0,
        memoryUtilization: 0,
        usedMemory: 0,
      },
    ];

    const result = nodeUtil.parseNodeGpuMeasurements(metricsResponse);
    expect(result).toEqual(expectedOutput);
  });
});
