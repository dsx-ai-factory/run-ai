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

  it("omits the cluster and node pool when they are not set", () => {
    expect(dashboardUtil.getCsvFileName("Workloads by status", undefined, undefined, date)).toBe(
      "workloads-by-status_2026-02-10_09-05.csv",
    );
  });
});
