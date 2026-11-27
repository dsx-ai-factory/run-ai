/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";

// models
import { ResourceType, type Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { UnifiedPolicyInfo } from "@/swagger-models/workloads-client";
import type { THybridFormState } from "@/models/workload-hybrid.model";
// services

import { permissionService } from "@/services/authorization/permission.service/permission.service";

export const useTrainingStore = defineStore("Training", {
  state: () => ({
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,

    flatFormState: null as THybridFormState | null,
    distributedWorkerFlatFormState: null as THybridFormState | null,
    distributedMasterFlatFormState: null as THybridFormState | null,

    distPolicyFlat: null as UnifiedPolicyInfo | null | undefined,

    distFlatFromContext: {
      fromCopyId: null as string | null,
      templateId: null as string | null,
    },
  }),
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Trainings);
    },
    setFlatFormState(flatData: THybridFormState | null): void {
      this.flatFormState = flatData;
    },
    getFlatFormState(): THybridFormState | null {
      return this.flatFormState;
    },
    clearFlatFormState(): void {
      this.flatFormState = null;
    },
    setDistributedWorkerFlatFormState(flatData: THybridFormState | null): void {
      this.distributedWorkerFlatFormState = flatData;
    },
    getDistributedWorkerFlatFormState(): THybridFormState | null {
      return this.distributedWorkerFlatFormState;
    },
    clearDistributedWorkerFlatFormState(): void {
      this.distributedWorkerFlatFormState = null;
    },
    setDistributedMasterFlatFormState(flatData: THybridFormState | null): void {
      this.distributedMasterFlatFormState = flatData;
    },
    getDistributedMasterFlatFormState(): THybridFormState | null {
      return this.distributedMasterFlatFormState;
    },
    clearDistributedMasterFlatFormState(): void {
      this.distributedMasterFlatFormState = null;
    },
    setDistFlatFromContext(context: { fromCopyId: string | null; templateId: string | null }): void {
      this.distFlatFromContext = context;
    },
    clearDistFlatFromContext(): void {
      this.distFlatFromContext = { fromCopyId: null, templateId: null };
    },
    setDistPolicyFlat(policy: UnifiedPolicyInfo | null | undefined): void {
      this.distPolicyFlat = policy;
    },
    getDistPolicyFlat(): UnifiedPolicyInfo | null | undefined {
      return this.distPolicyFlat;
    },
    clearDistPolicyFlat(): void {
      this.distPolicyFlat = null;
    },
  },
});
