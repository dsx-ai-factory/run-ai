/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from "vitest";
import { policyUtil, policyIdSeperator } from "./policy.util";
import { PolicyType, ScopeType } from "@/swagger-models/policy-service-client";

describe("policyUtil.isGvkPolicyId", () => {
  it("treats a backend UUID (no separator) as a GVK policy id", () => {
    expect(policyUtil.isGvkPolicyId("550e8400-e29b-41d4-a716-446655440000")).toBe(true);
  });

  it("treats a native composite id (with separator) as NOT a GVK id", () => {
    const nativeId = policyUtil.createPolicyId(PolicyType.Training, ScopeType.Project, 123);
    expect(nativeId.includes(policyIdSeperator)).toBe(true);
    expect(policyUtil.isGvkPolicyId(nativeId)).toBe(false);
  });

  it("treats a tenant-scoped native id (type + scope) as NOT a GVK id", () => {
    const nativeId = policyUtil.createPolicyId(PolicyType.Workspace, ScopeType.Tenant);
    expect(policyUtil.isGvkPolicyId(nativeId)).toBe(false);
  });

  it("returns false for an empty id", () => {
    expect(policyUtil.isGvkPolicyId("")).toBe(false);
  });
});

describe("policyUtil.resolveSelectedPolicyId", () => {
  it.each([PolicyType.Workspace, PolicyType.Training, PolicyType.Distributed, PolicyType.Inference])(
    "composes the project policy id for native type %s",
    (policyType) => {
      expect(policyUtil.resolveSelectedPolicyId(policyType, ScopeType.Project, 4500001)).toBe(
        `${policyType}_project_4500001`,
      );
    },
  );

  it("keeps a GVK policy id unchanged", () => {
    const policyId = "550e8400-e29b-41d4-a716-446655440000";

    expect(policyUtil.resolveSelectedPolicyId(policyId, ScopeType.Project, 4500001)).toBe(policyId);
  });
});
