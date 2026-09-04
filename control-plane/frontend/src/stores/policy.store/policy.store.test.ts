/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";

// The store module pulls in API-backed services at import time; stub them so importing the pure
// normalizer under test never constructs a real HTTP client.
vi.mock("@/services/control-plane/policy.service/policy.service", () => ({
  policyService: { listPolicies: vi.fn() },
}));
vi.mock("@/services/control-plane/gvk-policy.service/gvk-policy.service", () => ({
  gvkPolicyService: { deleteGvkPolicy: vi.fn() },
}));
vi.mock("@/services/authorization/permission.service/permission.service", () => ({ permissionService: {} }));
// The policy table-model imports autocompleteUtil, which transitively constructs many service clients
// at import time. The sort/filter tests below don't use autocomplete, so stub it out.
vi.mock("@/utils/autocomplete.util/autocomplete.util", () => ({
  autocompleteUtil: { policies: () => () => Promise.resolve([]) },
}));

import { createPinia, setActivePinia } from "pinia";
import { toPolicy, usePolicyStore, type IPolicy } from "./policy.store";
import { policyService } from "@/services/control-plane/policy.service/policy.service";
import { policyUtil } from "@/utils/policy.util/policy.util";
import {
  PolicySyncStatusOfClusterStatusEnum,
  PolicyType,
  ScopeType,
  type PolicyListEntry,
} from "@/swagger-models/policy-service-client";
// Real grid plumbing — the sort/filter tests run the normalized rows through the *actual* column
// model + filter service the grid uses, so they prove behavior rather than re-implement it.
import { filterService } from "@/services/filter.service/filter.service";
import { allPolicyColumnsMap, policyIndexColumns } from "@/table-models/policy.table-model";
import { getField } from "@/utils/table-format.util";
import type { ITableColumn } from "@/models/table.model";
import type { IFilterModel } from "@/models/filter.model";

const READY = PolicySyncStatusOfClusterStatusEnum.Ready;

// A native unified-list row (the shape `listPolicies({ includeGvkPolicies: true })` returns).
function nativeEntry(
  over: {
    name?: string;
    type?: PolicyType;
    projectId?: number;
    clusterId?: string;
    date?: string;
    status?: PolicyListEntry["status"];
  } = {},
): PolicyListEntry {
  const date = over.date ?? "2026-02-01T00:00:00Z";
  return {
    type: over.type ?? PolicyType.Workspace,
    meta: {
      id: `native-${over.projectId ?? 50}`,
      name: over.name ?? "native-policy",
      scope: ScopeType.Project,
      projectId: over.projectId ?? 50,
      clusterId: over.clusterId ?? "cluster-a",
      createdBy: "zoe@nvidia.com",
      createdAt: date,
      updatedBy: "zoe@nvidia.com",
      updatedAt: date,
    },
    status: over.status,
  } as PolicyListEntry;
}

// An any-workload (GVK) unified-list row: a `group/Kind/Version` type + a UUID `meta.id`.
function gvkEntry(
  over: {
    id?: string;
    group?: string;
    kind?: string;
    name?: string;
    clusterId?: string;
    date?: string;
    status?: PolicyListEntry["status"];
  } = {},
): PolicyListEntry {
  const group = over.group ?? "apps";
  const kind = over.kind ?? "Deployment";
  const date = over.date ?? "2026-01-15T00:00:00Z";
  return {
    type: `${group}/${kind}/v1` as PolicyType,
    meta: {
      id: over.id ?? "550e8400-e29b-41d4-a716-446655440000",
      name: over.name ?? `${kind.toLowerCase()}-policy`,
      scope: ScopeType.Project,
      projectId: 60,
      clusterId: over.clusterId ?? "cluster-a",
      createdBy: "alice@nvidia.com",
      createdAt: date,
      updatedBy: "alice@nvidia.com",
      updatedAt: date,
    },
    status: over.status,
  } as PolicyListEntry;
}

describe("toPolicy normalizer", () => {
  it("detects a GVK row by its group/Kind/Version type, routes it by the UUID id, and parses the Kind", () => {
    const row = toPolicy(gvkEntry({ id: "uuid-1", group: "apps", kind: "Deployment" }), "");
    expect(row.isGvk).toBe(true);
    expect(row.policyId).toBe("uuid-1");
    expect(row.gvkKind).toBe("Deployment");
    expect(policyUtil.isGvkPolicyId(row.policyId)).toBe(true); // routes view/edit/delete to GVK endpoints
  });

  it("parses the Kind for a core-group (empty group) type", () => {
    expect(toPolicy(gvkEntry({ group: "", kind: "Pod" }), "").gvkKind).toBe("Pod");
  });

  it("treats a native row as non-GVK and builds the synthetic composite policyId", () => {
    const row = toPolicy(nativeEntry({ type: PolicyType.Workspace, projectId: 50 }), "");
    expect(row.isGvk).toBeUndefined();
    expect(row.gvkKind).toBeUndefined();
    expect(policyUtil.isGvkPolicyId(row.policyId)).toBe(false); // composite id, not a UUID
    expect(row.policyId).toContain("50");
  });

  it("narrows the per-cluster status to the cluster selected in the grid filter", () => {
    const row = toPolicy(
      gvkEntry({
        clusterId: "cluster-a",
        status: {
          clusters: [
            { clusterId: "cluster-a", status: READY },
            { clusterId: "cluster-b", status: READY },
          ],
        },
      }),
      "cluster-a",
    );
    expect(row.status?.clusters).toEqual([{ clusterId: "cluster-a", status: READY }]);
  });

  it("narrows status to none when no cluster is selected (matches native rows)", () => {
    const entry = gvkEntry({ status: { clusters: [{ clusterId: "cluster-a", status: READY }] } });
    expect(toPolicy(entry, "").status?.clusters).toEqual([]);
  });
});

describe("policies grid — sort & filter through the real column model (unified list)", () => {
  const nativeWorkspace = toPolicy(
    nativeEntry({ name: "alpha-workspace", type: PolicyType.Workspace, projectId: 50, clusterId: "cluster-a" }),
    "",
  );
  const nativeTraining = toPolicy(
    nativeEntry({
      name: "beta-training",
      type: PolicyType.Distributed,
      projectId: 51,
      clusterId: "cluster-b",
      date: "2026-03-01T00:00:00Z",
    }),
    "",
  );
  const gvkDeployment = toPolicy(
    gvkEntry({ id: "gvk-deploy", kind: "Deployment", name: "gamma-deploy", clusterId: "cluster-a" }),
    "",
  );
  const gvkCronJob = toPolicy(
    gvkEntry({
      id: "gvk-cron",
      group: "batch",
      kind: "CronJob",
      name: "delta-cron",
      clusterId: "cluster-b",
      date: "2026-04-01T00:00:00Z",
    }),
    "",
  );
  const rows: IPolicy[] = [nativeWorkspace, nativeTraining, gvkDeployment, gvkCronJob];

  // Build a column filter exactly as the grid does: column name + the column's own field accessor.
  function colFilter(col: ITableColumn, term: string): IFilterModel[] {
    return [{ name: col.name, term, field: col.field }];
  }
  function names(filters: IFilterModel[]): Array<string | undefined> {
    return filterService.filterByColumns(rows, filters, policyIndexColumns).map((r) => r.meta?.name);
  }

  // Mirrors Quasar QTable's default client-side sort: the policy columns define no custom `sort` fn,
  // so the grid sorts on each column's raw `field` value — numbers numerically, anything else via
  // localeCompare. Lets us assert the unified rows order correctly across both policy kinds.
  function sortNames(col: ITableColumn, descending = false): Array<string | undefined> {
    const dir = descending ? -1 : 1;
    return [...rows]
      .sort((a, b) => {
        const av = getField(col, a as unknown as Record<string, unknown>);
        const bv = getField(col, b as unknown as Record<string, unknown>);
        if (av == null) return bv == null ? 0 : -dir;
        if (bv == null) return dir;
        if (typeof av === "number" && typeof bv === "number") return (av - bv) * dir;
        return String(av).localeCompare(String(bv)) * dir;
      })
      .map((r) => r.meta?.name);
  }

  it("filters Type by a GVK Kind, matching only the GVK row", () => {
    expect(names(colFilter(allPolicyColumnsMap.type, "deployment"))).toEqual(["gamma-deploy"]);
  });

  it("filters Type by 'training', matching the native distributed row (mapped to Training)", () => {
    expect(names(colFilter(allPolicyColumnsMap.type, "training"))).toEqual(["beta-training"]);
  });

  it("filters by the Policy (name) column across both kinds", () => {
    expect(names(colFilter(allPolicyColumnsMap.policy, "cron"))).toEqual(["delta-cron"]);
    expect(names(colFilter(allPolicyColumnsMap.policy, "alpha"))).toEqual(["alpha-workspace"]);
  });

  it("scopes GVK rows by clusterId exactly like native rows (dropped the GVK exemption)", () => {
    const onA = names(colFilter(allPolicyColumnsMap.cluster, "cluster-a"));
    expect(onA).toEqual(expect.arrayContaining(["alpha-workspace", "gamma-deploy"]));
    expect(onA).not.toContain("beta-training"); // native row on cluster-b is filtered out
    expect(onA).not.toContain("delta-cron"); // GVK row on cluster-b is now filtered out too
  });

  it("sorts the unified list by Creation time across both kinds", () => {
    expect(sortNames(allPolicyColumnsMap.createdAt)).toEqual([
      "gamma-deploy",
      "alpha-workspace",
      "beta-training",
      "delta-cron",
    ]);
  });
});

describe("policies grid — Status column renders the unified per-cluster status", () => {
  function statusOf(row: IPolicy) {
    return allPolicyColumnsMap.status.format?.(row.status, row);
  }

  it("renders 'Ready' (green) for a GVK row whose selected cluster reports Ready", () => {
    const row = toPolicy(
      gvkEntry({ clusterId: "cluster-a", status: { clusters: [{ clusterId: "cluster-a", status: READY }] } }),
      "cluster-a",
    );
    const opts = statusOf(row);
    expect(opts?.status).toBe("Ready");
    expect(opts?.color).toBe("success");
  });

  it("falls back to '-' when no cluster is selected", () => {
    const row = toPolicy(gvkEntry({ status: { clusters: [{ clusterId: "cluster-a", status: READY }] } }), "");
    expect(statusOf(row)?.status).toBe("-");
  });
});

describe("loadPolicies action — unified list", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.mocked(policyService.listPolicies).mockReset();
  });

  it("requests the unified list with includeGvkPolicies and normalizes native + GVK rows", async () => {
    vi.mocked(policyService.listPolicies).mockResolvedValue([
      nativeEntry({ name: "native-ws", type: PolicyType.Workspace, projectId: 50 }),
      gvkEntry({ id: "gvk-1", kind: "Deployment", name: "deploy" }),
    ]);

    const rows = await usePolicyStore().loadPolicies();

    expect(policyService.listPolicies).toHaveBeenCalledWith({ includeGvkPolicies: true });
    expect(rows.find((r) => r.policyId === "gvk-1")).toMatchObject({ isGvk: true, gvkKind: "Deployment" });
    const native = rows.find((r) => !r.isGvk);
    expect(native?.isGvk).toBeUndefined();
    expect(native?.policyId).toContain("_"); // synthetic composite id, not a UUID
  });

  it("a cluster-scoped grid on a version without any-workload support still loads (only native rows, no break)", async () => {
    // An under-2.26 cluster supports no any-workload (GVK) policies, so the unified list returns only
    // native rows for it. Selecting that cluster in the grid must not break or hide its native rows.
    vi.mocked(policyService.listPolicies).mockResolvedValue([
      nativeEntry({
        name: "old-native",
        projectId: 7,
        clusterId: "old-cluster",
        status: { clusters: [{ clusterId: "old-cluster", status: READY }] },
      }),
    ]);

    const rows = await usePolicyStore().loadPolicies({
      columnFilters: [{ name: "cluster", term: "old-cluster", field: allPolicyColumnsMap.cluster.field }],
    });

    expect(rows).toHaveLength(1);
    expect(rows[0].isGvk).toBeUndefined();
    expect(rows[0].status?.clusters).toEqual([{ clusterId: "old-cluster", status: READY }]);
  });

  it("scopes a GVK row's status to the selected (supported) cluster", async () => {
    vi.mocked(policyService.listPolicies).mockResolvedValue([
      gvkEntry({
        id: "gvk-new",
        kind: "Deployment",
        clusterId: "new-cluster",
        status: {
          clusters: [
            { clusterId: "new-cluster", status: READY },
            { clusterId: "other", status: READY },
          ],
        },
      }),
    ]);

    const rows = await usePolicyStore().loadPolicies({
      columnFilters: [{ name: "cluster", term: "new-cluster", field: allPolicyColumnsMap.cluster.field }],
    });

    expect(rows[0].isGvk).toBe(true);
    expect(rows[0].gvkKind).toBe("Deployment");
    expect(rows[0].status?.clusters).toEqual([{ clusterId: "new-cluster", status: READY }]);
  });
});
