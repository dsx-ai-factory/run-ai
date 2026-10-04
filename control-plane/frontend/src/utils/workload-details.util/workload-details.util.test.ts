/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { workloadDetailsUtil } from "@/utils/workload-details.util/workload-details.util";
import type { WorkloadDetailed } from "@/swagger-models/workloads-client";
import { formatK8sResourceMemory } from "@/utils/format.util/format.util";

describe("workloadDetailsUtil.parseComputeDetails", () => {
  it("includes allocation rows when allocatedResources is present", () => {
    const workloadDetailed = {
      allocatedResources: {
        gpu: 4,
        gpuMemory: "536870912",
        cpu: 1,
        cpuMemory: "104857600",
      },
    } as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseComputeDetails(workloadDetailed)[0]?.content ?? [];

    expect(content).toEqual(
      expect.arrayContaining([
        "GPU compute allocation: 4",
        `GPU memory allocation: ${formatK8sResourceMemory("536870912")}`,
        `CPU memory allocation: ${formatK8sResourceMemory("104857600")}`,
        expect.stringMatching(/^CPU compute allocation: /),
      ]),
    );
  });

  it("omits allocation rows when allocatedResources is absent", () => {
    const workloadDetailed = {} as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseComputeDetails(workloadDetailed)[0]?.content ?? [];

    expect(content.some((line) => line.includes("allocation"))).toBe(false);
  });

  it("omits individual allocation rows whose value is missing", () => {
    const workloadDetailed = {
      allocatedResources: { gpu: 0, cpu: 2 },
    } as unknown as WorkloadDetailed;

    const content = workloadDetailsUtil.parseComputeDetails(workloadDetailed)[0]?.content ?? [];

    expect(content).toContain("GPU compute allocation: 0");
    expect(content.some((line) => line.startsWith("CPU compute allocation"))).toBe(true);
    expect(content.some((line) => line.startsWith("GPU memory allocation"))).toBe(false);
    expect(content.some((line) => line.startsWith("CPU memory allocation"))).toBe(false);
  });
});
