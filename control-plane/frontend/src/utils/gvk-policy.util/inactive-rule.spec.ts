/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, expect, it } from "vitest";
import { gvkPolicyUtil } from "./gvk-policy.util";
import {
  ScopeType,
  type EffectiveRulesEffectiveRulesInner,
  type GetGvkPolicyResp,
} from "@/swagger-models/policy-service-client";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";

const rule = { field: "image", specSelector: "allContainers", enforcements: [{ required: true }] };
const parent: EffectiveRulesEffectiveRulesInner = {
  ...rule,
  allowOverride: false,
  sourceOfRule: { scope: ScopeType.Cluster, clusterId: "cluster-a" },
};
const policy = {
  scope: ScopeType.Project,
  projectId: "42",
  rules: [{ ...rule, enforcements: [{ blocked: true }] }],
} as GetGvkPolicyResp;

describe("inactive policy rules", () => {
  it("retains the winning rule and adds the inactive local rule with its own source and enforcements", () => {
    const rows = gvkPolicyUtil.policyViewRules(policy, [parent]);
    expect(rows).toHaveLength(2);
    expect(rows[0]).toEqual(parent);
    expect(rows[0].blockedBy).toBeUndefined();
    expect(rows[1]).toMatchObject({
      enforcements: [{ blocked: true }],
      sourceOfRule: { scope: ScopeType.Project, projectId: 42 },
      blockedBy: parent.sourceOfRule,
    });
  });

  it.each([
    { allowOverride: true },
    { allowOverride: undefined },
    { sourceOfRule: { scope: ScopeType.Cluster, isFallback: true } },
    { sourceOfRule: { scope: ScopeType.Project, projectId: 42 } },
    { sourceOfRule: null },
    { field: "command" },
    { specSelector: "mainContainer" },
    { itemSelector: "namedItem" },
    { specSelector: "allContainers(name=other)" },
  ])("does not mark unblocked or different targets: %j", (change) => {
    const effective = { ...parent, ...change };
    expect(gvkPolicyUtil.policyViewRules(policy, [effective])).toEqual([effective]);
  });

  it("recognizes department locks for project policies, but not for cluster policies", () => {
    const effective = { ...parent, sourceOfRule: { scope: ScopeType.Department, departmentId: "department-a" } };
    expect(gvkPolicyUtil.policyViewRules(policy, [effective])).toHaveLength(2);
    expect(gvkPolicyUtil.policyViewRules({ ...policy, scope: ScopeType.Cluster }, [effective])).toEqual([effective]);
    expect(gvkPolicyUtil.policyViewRules(null, [effective])).toEqual([effective]);
  });

  it.each([ScopeType.Cluster, ScopeType.Department])("uses the requested tooltip wording for %s", (scope) => {
    const source = { scope, clusterId: "c", departmentId: "d" };
    const orgUnits = [{ id: scope === ScopeType.Cluster ? "c" : "d", type: scope, name: "demo" }] as OrgUnit[];
    expect(gvkPolicyUtil.inactiveRuleTooltip(source, orgUnits, "runai")).toBe(
      `This rule is inactive because there is a policy rule for the same target in place in organization runai / ${scope} demo that prevents lower scopes from overriding it.`,
    );
  });

  it("handles unavailable source names without displaying undefined", () => {
    expect(gvkPolicyUtil.inactiveRuleTooltip(parent.sourceOfRule, [])).toContain("in cluster cluster-a that");
    expect(gvkPolicyUtil.inactiveRuleTooltip(undefined, [], "runai")).toBe("");
  });
});
