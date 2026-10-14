/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { shallowMount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { Quasar } from "quasar";
import GvkPolicyView from "./gvk-policy-view.vue";
import { RunaiTable } from "@/components/common/runai-table";
import { ScopeType, type EffectiveRulesEffectiveRulesInner } from "@/swagger-models/policy-service-client";
import type { ITableColumn } from "@/models/table.model";

describe("gvk-policy-view.vue", () => {
  it("places the inactive warning at the right end and leaves the effective rule unmarked", () => {
    const rule = {
      field: "image",
      specSelector: "allContainers",
      enforcements: [{ required: true }],
      sourceOfRule: { scope: ScopeType.Project, projectId: 42 },
    };
    const wrapper = shallowMount(GvkPolicyView, {
      props: {
        rules: [rule, { ...rule, blockedBy: { scope: ScopeType.Cluster, clusterId: "cluster-a" } }],
        tenantName: "runai",
      },
      global: { plugins: [Quasar] },
    });
    const table = wrapper.findComponent(RunaiTable);
    const columns = table.props("columns") as ITableColumn[];
    const warning = columns.at(-1);
    const rows = table.props("rows");
    expect(warning?.name).toBe("inactiveWarning");
    expect(warning?.cellIcon?.(rows[0])).toBeUndefined();
    expect(warning?.cellIcon?.(rows[1])).toEqual({
      name: "ph ph-warning-circle",
      color: "warning",
      size: "20px",
      tooltip:
        "This rule is inactive because there is a policy rule for the same target in place in organization runai / cluster cluster-a that prevents lower scopes from overriding it.",
    });
  });
  it("renders the allowOverride value for every effective rule", () => {
    const rules: EffectiveRulesEffectiveRulesInner[] = [
      {
        field: "image",
        specSelector: "mainContainer",
        allowOverride: true,
        enforcements: [{ required: true }],
        sourceOfRule: { scope: ScopeType.Cluster },
      },
      {
        field: "command",
        specSelector: "mainContainer",
        enforcements: [{ blocked: true }],
        sourceOfRule: { scope: ScopeType.Cluster },
      },
    ];
    const wrapper = shallowMount(GvkPolicyView, {
      props: { rules },
      global: { plugins: [Quasar] },
    });
    const table = wrapper.findComponent(RunaiTable);
    const rows = table.props("rows") as Array<{ allowOverride: boolean }>;
    const columns = table.props("columns") as Array<{
      name: string;
      label: string;
      format?: (value: boolean) => string;
    }>;
    const allowOverrideColumn = columns.find((column) => column.name === "allowOverride");

    expect(rows.map((row) => row.allowOverride)).toEqual([true, false]);
    expect(columns.map((column) => column.label).slice(-2)).toEqual(["Source", "Override by descendant scope"]);
    expect(allowOverrideColumn).toBeDefined();
    expect(allowOverrideColumn?.format?.(true)).toBe("Allowed");
    expect(allowOverrideColumn?.format?.(false)).toBe("Not allowed");
  });
});
