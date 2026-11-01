/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { allPolicyColumnsMap } from "./policy.table-model";
import { policyService } from "@/services/control-plane/policy.service/policy.service";
import {
  PolicyPhase,
  PolicySyncStatusOfClusterStatusEnum,
  PolicyType,
  ScopeType,
  type PolicyListEntry,
} from "@/swagger-models/policy-service-client";

vi.mock("@/services/control-plane/policy.service/policy.service", () => ({
  policyService: { listPolicies: vi.fn() },
}));

describe("policy status autocomplete", () => {
  beforeEach(() => {
    const metadata = {
      scope: ScopeType.Cluster,
      name: "policy",
      createdBy: "test@run.ai",
      createdAt: "2026-11-01T00:00:00Z",
      updatedBy: "test@run.ai",
      updatedAt: "2026-11-01T00:00:00Z",
    };
    const gvkRows = [PolicyPhase.Degraded, PolicyPhase.PendingUpdate, PolicyPhase.Ready].map(
      (phase): PolicyListEntry => ({
        type: "apps/Deployment/v1" as PolicyType,
        meta: { ...metadata, id: phase, clusterId: "cluster-a" },
        status: { clusters: [{ clusterId: "cluster-a", status: PolicySyncStatusOfClusterStatusEnum.Ready }] },
        availability: { phase },
      }),
    );
    vi.mocked(policyService.listPolicies).mockResolvedValue([
      ...gvkRows,
      {
        type: PolicyType.Workspace,
        meta: { ...metadata, id: "native", clusterId: "cluster-b" },
        status: { clusters: [{ clusterId: "cluster-b", status: PolicySyncStatusOfClusterStatusEnum.Failed }] },
      },
    ]);
  });

  it("suggests availability phases together with native sync statuses", async () => {
    const options = await allPolicyColumnsMap.status.filter?.autocompleteOptionsFn?.("", "");
    expect(options).toEqual(["Degraded", "Pending update", "Ready", "Failed"]);
  });

  it("finds degraded status by search text", async () => {
    const options = await allPolicyColumnsMap.status.filter?.autocompleteOptionsFn?.("degr", "cluster-a");
    expect(options).toEqual(["Degraded"]);
  });

  it("keeps suggestions scoped to the selected cluster", async () => {
    const options = await allPolicyColumnsMap.status.filter?.autocompleteOptionsFn?.("", "cluster-b");
    expect(options).toEqual(["Failed"]);
  });
});
