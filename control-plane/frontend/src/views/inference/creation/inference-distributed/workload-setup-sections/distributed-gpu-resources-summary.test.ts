import { describe, it, expect } from "vitest";
import type { WorkspaceSpecSpecAllOfCompute } from "@/swagger-models/workloads-client";

/**
 * Unit tests for distributed-gpu-resources-summary.vue
 *
 * These tests verify the GPU resource calculation and display logic
 * for distributed inference workloads.
 */

// Helper functions extracted from the component for testing
function getGpuInfo(compute: WorkspaceSpecSpecAllOfCompute | undefined): {
  value: number;
  display: string;
  unit: string;
} {
  if (!compute) {
    return { value: 0, display: "0 GPUs", unit: "devices" };
  }

  const gpuDevices = compute.gpuDevicesRequest ?? 1;
  const gpuRequestType = compute.gpuRequestType;

  // Handle GPU memory request
  if (gpuRequestType === "memory" && compute.gpuMemoryRequest) {
    const memoryRequest = compute.gpuMemoryRequest;
    const isGB = memoryRequest.toUpperCase().endsWith("G");
    const memoryUnit = isGB ? "GB" : "MB";
    const memoryValue = parseFloat(memoryRequest) * gpuDevices;
    return {
      value: memoryValue,
      display: `${memoryValue} ${memoryUnit}`,
      unit: memoryUnit,
    };
  }

  // Handle GPU portion (fractioning with % of device)
  if (gpuRequestType === "portion" && compute.gpuPortionRequest != null) {
    const portionValue = gpuDevices * compute.gpuPortionRequest;
    return {
      value: portionValue,
      display: `${portionValue.toFixed(2)} GPUs`,
      unit: "portion",
    };
  }

  // Fallback: check if gpuMemoryRequest exists even without gpuRequestType
  if (compute.gpuMemoryRequest) {
    const memoryRequest = compute.gpuMemoryRequest;
    const isGB = memoryRequest.toUpperCase().endsWith("G");
    const memoryUnit = isGB ? "GB" : "MB";
    const memoryValue = parseFloat(memoryRequest) * gpuDevices;
    return {
      value: memoryValue,
      display: `${memoryValue} ${memoryUnit}`,
      unit: memoryUnit,
    };
  }

  // Fallback: check if gpuPortionRequest exists even without gpuRequestType
  if (compute.gpuPortionRequest != null && compute.gpuPortionRequest !== undefined) {
    const portionValue = gpuDevices * compute.gpuPortionRequest;
    return {
      value: portionValue,
      display: `${portionValue.toFixed(2)} GPUs`,
      unit: "portion",
    };
  }

  // Default: GPU devices (no fractioning)
  const devices = compute.gpuDevicesRequest ?? 0;
  return {
    value: devices,
    display: `${devices} GPUs`,
    unit: "devices",
  };
}

function formatGpuValue(value: number, unit: string): string {
  if (unit === "portion") {
    return `${value.toFixed(2)} GPUs`;
  }
  if (unit === "GB" || unit === "MB") {
    return `${value} ${unit}`;
  }
  // devices
  return `${value} GPUs`;
}

function calculateTotalGpuDisplay(
  leaderCompute: WorkspaceSpecSpecAllOfCompute | undefined,
  workerCompute: WorkspaceSpecSpecAllOfCompute | undefined,
  replicas: number,
  numWorkers: number,
): string {
  const leaderInfo = getGpuInfo(leaderCompute);
  const workerInfo = getGpuInfo(workerCompute);
  const isWholeGpuAllocation = leaderInfo.unit === "devices" && workerInfo.unit === "devices";

  const workerTotalPerReplica = workerInfo.value * numWorkers;

  if (isWholeGpuAllocation) {
    const total = (leaderInfo.value + workerTotalPerReplica) * replicas;
    return formatGpuValue(total, "devices");
  }

  const leaderTotal = leaderInfo.value * replicas;
  const workerTotal = workerTotalPerReplica * replicas;

  if (leaderInfo.unit === workerInfo.unit) {
    return formatGpuValue(leaderTotal + workerTotal, leaderInfo.unit);
  }

  return `${formatGpuValue(leaderTotal, leaderInfo.unit)} + ${formatGpuValue(workerTotal, workerInfo.unit)}`;
}

function calculateTooltipText(
  leaderCompute: WorkspaceSpecSpecAllOfCompute | undefined,
  workerCompute: WorkspaceSpecSpecAllOfCompute | undefined,
  replicas: number,
  numWorkers: number,
): string {
  const leaderInfo = getGpuInfo(leaderCompute);
  const workerInfo = getGpuInfo(workerCompute);
  const isWholeGpuAllocation = leaderInfo.unit === "devices" && workerInfo.unit === "devices";

  const leaderDisplay = formatGpuValue(leaderInfo.value, leaderInfo.unit);
  const workerTotalValue = workerInfo.value * numWorkers;
  const workerDisplay = formatGpuValue(workerTotalValue, workerInfo.unit);

  if (isWholeGpuAllocation) {
    return `((leader: ${leaderDisplay}) + (workers: ${workerDisplay})) x (${replicas} replicas)`;
  }

  return `(leader: ${leaderDisplay} x ${replicas} replicas) + (workers: ${workerDisplay} x ${replicas} replicas)`;
}

describe("distributed-gpu-resources-summary", () => {
  describe("getGpuInfo", () => {
    it("should return 0 GPUs for undefined compute", () => {
      const result = getGpuInfo(undefined);

      expect(result).toEqual({ value: 0, display: "0 GPUs", unit: "devices" });
    });

    it("should return GPU devices when no fractioning is used", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuDevicesRequest: 4,
      };

      const result = getGpuInfo(compute);

      expect(result).toEqual({ value: 4, display: "4 GPUs", unit: "devices" });
    });

    it("should handle memory request in GB", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "memory",
        gpuMemoryRequest: "2G",
        gpuDevicesRequest: 1,
      };

      const result = getGpuInfo(compute);

      expect(result).toEqual({ value: 2, display: "2 GB", unit: "GB" });
    });

    it("should handle memory request in MB", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "memory",
        gpuMemoryRequest: "500M",
        gpuDevicesRequest: 1,
      };

      const result = getGpuInfo(compute);

      expect(result).toEqual({ value: 500, display: "500 MB", unit: "MB" });
    });

    it("should multiply memory by gpuDevicesRequest", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "memory",
        gpuMemoryRequest: "1G",
        gpuDevicesRequest: 3,
      };

      const result = getGpuInfo(compute);

      expect(result).toEqual({ value: 3, display: "3 GB", unit: "GB" });
    });

    it("should handle portion request", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "portion",
        gpuPortionRequest: 0.5,
        gpuDevicesRequest: 1,
      };

      const result = getGpuInfo(compute);

      expect(result).toEqual({ value: 0.5, display: "0.50 GPUs", unit: "portion" });
    });

    it("should multiply portion by gpuDevicesRequest", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "portion",
        gpuPortionRequest: 0.25,
        gpuDevicesRequest: 2,
      };

      const result = getGpuInfo(compute);

      expect(result).toEqual({ value: 0.5, display: "0.50 GPUs", unit: "portion" });
    });

    it("should fallback to memory request even without gpuRequestType", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuMemoryRequest: "1G",
        gpuDevicesRequest: 1,
      };

      const result = getGpuInfo(compute);

      expect(result).toEqual({ value: 1, display: "1 GB", unit: "GB" });
    });

    it("should fallback to portion request even without gpuRequestType", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {
        gpuPortionRequest: 0.75,
        gpuDevicesRequest: 1,
      };

      const result = getGpuInfo(compute);

      expect(result).toEqual({ value: 0.75, display: "0.75 GPUs", unit: "portion" });
    });

    it("should return 0 GPUs when gpuDevicesRequest is not set", () => {
      const compute: WorkspaceSpecSpecAllOfCompute = {};

      const result = getGpuInfo(compute);

      expect(result).toEqual({ value: 0, display: "0 GPUs", unit: "devices" });
    });
  });

  describe("formatGpuValue", () => {
    it("should format devices as GPUs", () => {
      expect(formatGpuValue(4, "devices")).toBe("4 GPUs");
    });

    it("should format portion with 2 decimal places", () => {
      expect(formatGpuValue(0.5, "portion")).toBe("0.50 GPUs");
    });

    it("should format GB memory", () => {
      expect(formatGpuValue(2, "GB")).toBe("2 GB");
    });

    it("should format MB memory", () => {
      expect(formatGpuValue(500, "MB")).toBe("500 MB");
    });
  });

  describe("totalGpuDisplay - Whole GPU allocation", () => {
    it("should calculate total: (leader: 1 GPU) + (workers: 4 GPUs) x 3 replicas = 15 GPUs", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 1 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 4 };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 3, 1);

      expect(result).toBe("15 GPUs");
    });

    it("should calculate total with multiple workers: (leader: 1) + (2 workers x 2 GPUs each) x 2 replicas = 10 GPUs", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 1 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 2 };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 2, 2);

      expect(result).toBe("10 GPUs");
    });

    it("should handle single replica", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 2 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 4 };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 1, 1);

      expect(result).toBe("6 GPUs");
    });
  });

  describe("totalGpuDisplay - Mixed allocation (whole GPU + memory fraction)", () => {
    it("should show separately: 3 GPUs + 1.5 GB when leader uses GPUs and workers use memory", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 1 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "memory",
        gpuMemoryRequest: "500M",
        gpuDevicesRequest: 1,
      };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 3, 1);

      expect(result).toBe("3 GPUs + 1500 MB");
    });

    it("should show separately when leader uses memory and workers use GPUs", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "memory",
        gpuMemoryRequest: "1G",
        gpuDevicesRequest: 1,
      };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 2 };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 2, 1);

      expect(result).toBe("2 GB + 4 GPUs");
    });
  });

  describe("totalGpuDisplay - Same fractional allocation", () => {
    it("should combine when both use memory (same unit)", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "memory",
        gpuMemoryRequest: "1G",
        gpuDevicesRequest: 1,
      };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "memory",
        gpuMemoryRequest: "2G",
        gpuDevicesRequest: 1,
      };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 2, 1);

      // (1 GB * 2) + (2 GB * 1 * 2) = 2 + 4 = 6 GB
      expect(result).toBe("6 GB");
    });

    it("should combine when both use portion (same unit)", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "portion",
        gpuPortionRequest: 0.5,
        gpuDevicesRequest: 1,
      };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "portion",
        gpuPortionRequest: 0.25,
        gpuDevicesRequest: 1,
      };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 2, 2);

      // leader: 0.5 * 2 = 1, workers: 0.25 * 2 * 2 = 1, total = 2
      expect(result).toBe("2.00 GPUs");
    });
  });

  describe("tooltipText - Whole GPU allocation", () => {
    it("should format as ((leader: X) + (workers: Y)) x (Z replicas)", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 1 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 4 };

      const result = calculateTooltipText(leaderCompute, workerCompute, 3, 1);

      expect(result).toBe("((leader: 1 GPUs) + (workers: 4 GPUs)) x (3 replicas)");
    });

    it("should include numWorkers in worker total", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 1 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 2 };

      const result = calculateTooltipText(leaderCompute, workerCompute, 3, 2);

      // workers: 2 GPUs * 2 numWorkers = 4 GPUs
      expect(result).toBe("((leader: 1 GPUs) + (workers: 4 GPUs)) x (3 replicas)");
    });
  });

  describe("tooltipText - Mixed/fractional allocation", () => {
    it("should format as (leader: X x Z replicas) + (workers: Y x Z replicas)", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 1 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "memory",
        gpuMemoryRequest: "500M",
        gpuDevicesRequest: 1,
      };

      const result = calculateTooltipText(leaderCompute, workerCompute, 3, 1);

      expect(result).toBe("(leader: 1 GPUs x 3 replicas) + (workers: 500 MB x 3 replicas)");
    });

    it("should show correct memory unit in tooltip", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 2 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "memory",
        gpuMemoryRequest: "1G",
        gpuDevicesRequest: 1,
      };

      const result = calculateTooltipText(leaderCompute, workerCompute, 2, 1);

      expect(result).toBe("(leader: 2 GPUs x 2 replicas) + (workers: 1 GB x 2 replicas)");
    });

    it("should format portion values with decimals", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "portion",
        gpuPortionRequest: 0.5,
        gpuDevicesRequest: 1,
      };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = {
        gpuRequestType: "portion",
        gpuPortionRequest: 0.25,
        gpuDevicesRequest: 1,
      };

      const result = calculateTooltipText(leaderCompute, workerCompute, 2, 2);

      // workers: 0.25 * 2 = 0.5
      expect(result).toBe("(leader: 0.50 GPUs x 2 replicas) + (workers: 0.50 GPUs x 2 replicas)");
    });
  });

  describe("edge cases", () => {
    it("should handle zero GPUs for leader", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 0 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 4 };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 2, 1);

      expect(result).toBe("8 GPUs");
    });

    it("should handle zero GPUs for workers", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 2 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 0 };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 3, 1);

      expect(result).toBe("6 GPUs");
    });

    it("should handle undefined compute objects", () => {
      const result = calculateTotalGpuDisplay(undefined, undefined, 3, 1);

      expect(result).toBe("0 GPUs");
    });

    it("should handle large number of replicas", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 1 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 2 };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 100, 1);

      expect(result).toBe("300 GPUs");
    });

    it("should handle large number of workers", () => {
      const leaderCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 1 };
      const workerCompute: WorkspaceSpecSpecAllOfCompute = { gpuDevicesRequest: 1 };

      const result = calculateTotalGpuDisplay(leaderCompute, workerCompute, 2, 10);

      // (1 + 10) * 2 = 22
      expect(result).toBe("22 GPUs");
    });
  });
});
