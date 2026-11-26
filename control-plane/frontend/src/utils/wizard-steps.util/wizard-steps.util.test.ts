/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { filterDepartmentFormV3Steps, type IStepFilterContext } from "./wizard-steps.util";
import { EStepNames } from "@/models/org-unit-forms.model";
import { ResourceType } from "@/swagger-models/authorization-client";

describe("filterDepartmentFormV3Steps", () => {
  it.each<{ name: string; context: IStepFilterContext; want: EStepNames[] }>([
    {
      name: "shows every step when a node pool has CPU quota on",
      context: { nodepoolCapabilities: { cpuQuota: true } },
      want: [
        EStepNames.ScopeName,
        EStepNames.GpuQuota,
        EStepNames.CPUQuota,
        EStepNames.Projects,
        EStepNames.SchedulingRules,
        EStepNames.AccessRules,
      ],
    },
    {
      name: "keeps GPU quota and drops CPU quota when no node pool has quota or CPU quota on",
      context: { nodepoolCapabilities: { cpuQuota: false, quota: false } },
      want: [
        EStepNames.ScopeName,
        EStepNames.GpuQuota,
        EStepNames.Projects,
        EStepNames.SchedulingRules,
        EStepNames.AccessRules,
      ],
    },
    {
      name: "drops access rules without permission to read them",
      context: {
        nodepoolCapabilities: { cpuQuota: true },
        hasPermission: (resource) => resource !== ResourceType.AccessRules,
      },
      want: [
        EStepNames.ScopeName,
        EStepNames.GpuQuota,
        EStepNames.CPUQuota,
        EStepNames.Projects,
        EStepNames.SchedulingRules,
      ],
    },
  ])("$name", ({ context, want }) => {
    expect(filterDepartmentFormV3Steps(context).map((step) => step.name)).toEqual(want);
  });
});
