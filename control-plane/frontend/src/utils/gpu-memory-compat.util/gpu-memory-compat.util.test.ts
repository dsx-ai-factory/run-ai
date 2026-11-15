/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { legacyComputeForForm, legacyPolicyForForm } from "./gpu-memory-compat.util";

describe("GPU memory aliases in existing forms", () => {
  it("preserves whole GPU requests and policy locks without inventing device requests", () => {
    const compute = { gpuPortionRequest: 4, gpuComputeMode: "time-slicing" };
    expect(legacyComputeForForm(compute)).toEqual(compute);
    const policy = {
      defaults: { compute },
      rules: { compute: { gpuPortionRequest: { canEdit: false }, gpuMemoryPortionRequest: { canEdit: false } } },
    };
    expect(legacyPolicyForForm(policy)).toEqual({
      defaults: { compute },
      rules: { compute: { gpuPortionRequest: { canEdit: false } } },
    });
  });

  it("allows editing and clearing a dual-alias asset without retaining stale values", () => {
    const response = {
      gpuPortionRequest: 0.25,
      gpuMemoryPortionRequest: 0.25,
      gpuPortionLimit: 0.5,
      gpuMemoryPortionLimit: 0.5,
      cpuCoreRequest: 2,
    };
    const edited = legacyComputeForForm(response);
    edited.gpuPortionRequest = 0.75;
    delete (edited as Partial<typeof edited>).gpuPortionLimit;
    expect(JSON.parse(JSON.stringify(edited))).toEqual({ gpuPortionRequest: 0.75, cpuCoreRequest: 2 });
    expect(response.gpuPortionRequest).toBe(0.25);
    expect(response.gpuMemoryPortionLimit).toBe(0.5);
  });

  it.each([0, null, 0.25])("preserves a new-only response value %s", (value) => {
    expect(legacyComputeForForm({ gpuMemoryPortionRequest: value })).toEqual({ gpuPortionRequest: value });
  });

  it("keeps only the edited policy rule and preserves metadata with similarly named keys", () => {
    const response = {
      meta: { gpuMemoryPortionRequest: "user metadata" },
      policy: {
        rules: {
          worker: {
            compute: {
              gpuPortionRequest: { canEdit: false },
              gpuMemoryPortionRequest: { canEdit: false },
            },
          },
        },
      },
      effective: { defaults: { leader: { compute: { gpuPortionRequest: 0.25, gpuMemoryPortionRequest: 0.25 } } } },
    };
    const edited = legacyPolicyForForm(response);
    edited.policy.rules.worker.compute.gpuPortionRequest = { canEdit: true };
    expect(edited.policy.rules.worker.compute).toEqual({ gpuPortionRequest: { canEdit: true } });
    expect(edited.effective.defaults.leader.compute).toEqual({ gpuPortionRequest: 0.25 });
    expect(edited.meta).toEqual(response.meta);
    expect(response.policy.rules.worker.compute.gpuPortionRequest.canEdit).toBe(false);
  });
});
