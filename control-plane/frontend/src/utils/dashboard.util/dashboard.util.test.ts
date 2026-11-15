/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { dashboardUtil } from "@/utils/dashboard.util";

describe("dashboardUtil.getCsvFileName", () => {
  const date = new Date(2026, 1, 10, 9, 5);

  it("names the file after the widget, cluster and node pool", () => {
    expect(dashboardUtil.getCsvFileName("Resources allocation", "Prod Cluster", "gpu-a100", date)).toBe(
      "resources-allocation_prod-cluster_gpu-a100_2026-02-10_09-05.csv",
    );
  });

  it("replaces characters that are not filename-safe", () => {
    expect(dashboardUtil.getCsvFileName("Resources utilization", "Prod (EU)/a", "pool:1", date)).toBe(
      "resources-utilization_prod-eu-a_pool-1_2026-02-10_09-05.csv",
    );
  });

  it("omits parts that have no filename-safe characters", () => {
    expect(dashboardUtil.getCsvFileName("Resources allocation", "集群", "pool-a", date)).toBe(
      "resources-allocation_pool-a_2026-02-10_09-05.csv",
    );
  });

  it("omits the cluster and node pool when they are not set", () => {
    expect(dashboardUtil.getCsvFileName("Workloads by status", undefined, undefined, date)).toBe(
      "workloads-by-status_2026-02-10_09-05.csv",
    );
  });
});
