/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { runaiHeaderUtil } from "./runai-header.util";
import { describe, it, expect } from "vitest";
import { menuSubjects } from "@/components/infra/runai-layout/routes-model/routes-model";

describe("runaiHeaderUtil.findLabelPath", () => {
  it("should return the correct path for a label in the first level", () => {
    const result = runaiHeaderUtil.findLabelPath([menuSubjects.analytics, menuSubjects.workloadManager], "Overview");
    expect(result).toEqual(["Analytics", "Dashboard", "Overview"]);
  });

  it("should return the correct path for a nested label", () => {
    const result = runaiHeaderUtil.findLabelPath([menuSubjects.analytics, menuSubjects.workloadManager], "Environments");
    expect(result).toEqual(["Workload manager", "Assets", "Environments"]);
  });

  it("should return the correct path for a label with spaces", () => {
    const result = runaiHeaderUtil.findLabelPath(
      [menuSubjects.analytics, menuSubjects.workloadManager],
      "Compute resources",
    );
    expect(result).toEqual(["Workload manager", "Assets", "Compute resources"]);
  });

  it("should return null if the label is not found", () => {
    const result = runaiHeaderUtil.findLabelPath(
      [menuSubjects.analytics, menuSubjects.workloadManager],
      "NonExistingLabel",
    );
    expect(result).toBeNull();
  });
});
