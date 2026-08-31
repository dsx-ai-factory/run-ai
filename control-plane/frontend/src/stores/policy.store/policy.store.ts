/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Stores
import { defineStore } from "pinia";
// Services
import { policyService } from "@/services/control-plane/policy.service/policy.service";
import { gvkPolicyService } from "@/services/control-plane/gvk-policy.service/gvk-policy.service";

// Models
import type { IFilterBy } from "@/models/filter.model";
import {
  PolicyType,
  type PolicyListEntry,
  type PolicySyncStatusOfCluster,
  PolicySyncStatusOfClusterStatusEnum,
  type PolicySyncStatus,
} from "@/swagger-models/policy-service-client";
import type { Action } from "@/swagger-models/authorization-client";
import { type PermittedScopes, ResourceType } from "@/swagger-models/authorization-client";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { policyUtil, type IListPolicyQuery } from "@/utils/policy.util/policy.util";

export interface IPolicy extends PolicyListEntry {
  policyId: string;
  isGvk?: boolean;
  gvkKind?: string;
}

export interface ISyncedPolicyStatus {
  isSycned: boolean;
  errorMessage?: string;
}

// GVK rows are keyed by backend UUID `meta.id`; native rows by composite `type_scope_scopeId`.
export function toPolicy(policy: PolicyListEntry, clusterIdFilter: string): IPolicy {
  const isGvk: boolean = policyUtil.isGvkPolicyListEntry(policy);
  const policyId: string = isGvk ? (policy.meta?.id ?? "") : policyUtil.createPolicyIdFromPolicyListEntry(policy);

  const normalized: IPolicy = isGvk
    ? { ...policy, policyId, isGvk: true, gvkKind: policyUtil.gvkKindFromPolicyType(policy.type) }
    : { ...policy, policyId };

  if (policy.status) {
    const currentClusterStatus: PolicySyncStatus = {
      ...policy.status,
      clusters: policy.status.clusters?.filter((cluster) => cluster.clusterId === clusterIdFilter),
    };
    normalized.status = currentClusterStatus;
  }
  return normalized;
}

export const usePolicyStore = defineStore("Policy", {
  state: () => ({
    loading: false as boolean,
    policies: [] as IPolicy[],
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,
  }),
  getters: {
    policyList(): Array<IPolicy> {
      return this.policies;
    },
  },
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Policies);
    },
    getPolicyByPolicyId(policyId: string): IPolicy | null {
      return this.policies.find((policy: IPolicy) => policy.policyId === policyId) || null;
    },
    async loadPolicies(filterBy: IFilterBy = {}): Promise<IPolicy[]> {
      const policies: PolicyListEntry[] = await policyService.listPolicies({ includeGvkPolicies: true });

      const clusterIdFilter: string = filterBy.columnFilters?.find((f) => f.name === "cluster")?.term || "";
      this.policies = policies.map((policy: PolicyListEntry) => toPolicy(policy, clusterIdFilter));
      return this.policies;
    },
    async remove(policy: IPolicy): Promise<void> {
      if (policy.isGvk) {
        return await gvkPolicyService.deleteGvkPolicy(policy.policyId);
      }
      if (!policy.meta) return;
      switch (policy.type) {
        case PolicyType.Training:
          return await policyService.deleteTrainingPolicy(policy.meta);
        case PolicyType.Workspace:
          return await policyService.deleteWorkspacePolicy(policy.meta);
        case PolicyType.Distributed:
          return await policyService.deleteDistributedPolicy(policy.meta);
        case PolicyType.Inference:
          return await policyService.deleteInferencePolicy(policy.meta);
        case PolicyType.DistributedInference:
          return await policyService.deleteDistributedInferencePolicy(policy.meta);
        default:
          return;
      }
    },
    async verifyPolicySync(policyId: string, clusterId: string): Promise<ISyncedPolicyStatus> {
      const syncStatus: ISyncedPolicyStatus = { isSycned: false };

      const queryObject: IListPolicyQuery = policyUtil.getQueryObjectFromPolicyId(policyId);
      const modifiedPolicy: PolicyListEntry = (await policyService.listPolicies(queryObject))[0];

      if (!modifiedPolicy?.status?.clusters?.length) return syncStatus;

      const relevantCluster = modifiedPolicy.status.clusters.find(
        (c: PolicySyncStatusOfCluster) => c.clusterId === clusterId,
      );

      if (
        relevantCluster?.status === PolicySyncStatusOfClusterStatusEnum.Ready ||
        relevantCluster?.status === PolicySyncStatusOfClusterStatusEnum.Failed
      ) {
        syncStatus.isSycned = true;
        syncStatus.errorMessage = relevantCluster?.errorMessage || "";
        return syncStatus;
      }
      return syncStatus;
    },
  },
});
