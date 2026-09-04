/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";

// models
import type {
  AssetsDefaults,
  PolicyInfo,
  TrainingCreationRequestV2,
  TrainingSpecificRunParams,
  TrainingV2,
} from "@/swagger-models/assets-service-client";
import { ResourceType, type Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import type { IUIWorkloadCreation } from "@/models/workload.model";
import type { UnifiedPolicyInfo } from "@/swagger-models/workloads-client";
import type { THybridFormState } from "@/models/workload-hybrid.model";
// services
import { trainingService } from "@/services/control-plane/training.service/training.service";

// utils
import { workloadUtil } from "@/utils/workload.util/workload.util";
import { permissionService } from "@/services/authorization/permission.service/permission.service";
import { deepCopy } from "@/utils/common.util/common.util";

export const useTrainingStore = defineStore("Training", {
  state: () => ({
    training: workloadUtil.getEmptyUIWorkloadCreation() as IUIWorkloadCreation,
    templateId: null as string | null,
    selectedTraining: null as TrainingV2 | null,
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,

    // these details are treat as read only after setting (saving the origin model with policies set)
    loadedTrainingModelWithDefaultsAndPolicy: null as TrainingV2 | null,

    // flat submission data ONLY
    flatFormState: null as THybridFormState | null,
    distributedWorkerFlatFormState: null as THybridFormState | null,
    distributedMasterFlatFormState: null as THybridFormState | null,

    distPolicyFlat: null as UnifiedPolicyInfo | null | undefined,

    distFlatFromContext: {
      fromCopyId: null as string | null,
      templateId: null as string | null,
    },
  }),
  getters: {
    trainingName(): string {
      return this.training.name || "";
    },
    // these details are treated as read only after setting (saving the origin model with policies set)
    loadedPolicyInfo(): PolicyInfo | null {
      return this.loadedTrainingModelWithDefaultsAndPolicy?.policyInfo || null;
    },
    // these details are treated as read only after setting (saving the origin model with policies set)
    loadedPolicyInfoDistributedMaster(): PolicyInfo | null {
      return this.loadedTrainingModelWithDefaultsAndPolicy?.distributed?.master?.policyInfo || null;
    },
    // these details are treated as read only after setting (saving the origin model with policies set)
    loadedSpecificRunParamsWithDefaults(): TrainingSpecificRunParams | null {
      return deepCopy(this.loadedTrainingModelWithDefaultsAndPolicy?.specificRunParams || null);
    },
    // these details are treated as read only after setting (saving the origin model with policies set)
    loadedSpecificRunParamsWithDefaultsDistributedMaster(): TrainingSpecificRunParams | null {
      return deepCopy(this.loadedTrainingModelWithDefaultsAndPolicy?.distributed?.master?.specificRunParams || null);
    },
    loadedAssetsAttributesDefaults(): AssetsDefaults | null {
      return this.loadedTrainingModelWithDefaultsAndPolicy?.policyInfo?.attributeDefaults?.assets || null;
    },
    loadedAssetsAttributesDefaultsMaster(): AssetsDefaults | null {
      return (
        this.loadedTrainingModelWithDefaultsAndPolicy?.distributed?.master?.policyInfo?.attributeDefaults?.assets || null
      );
    },
  },
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Trainings);
    },
    setLoadedTrainingModelWithDefaults(model: TrainingV2): void {
      this.loadedTrainingModelWithDefaultsAndPolicy = model;
    },
    setTraining(training: IUIWorkloadCreation): void {
      this.training = training;
    },
    setTemplate(templateId: string | null): void {
      this.templateId = templateId;
    },
    resetCreateTrainingData(): void {
      this.templateId = null;
      this.loadedTrainingModelWithDefaultsAndPolicy = null;
      this.training = workloadUtil.getEmptyUIWorkloadCreation();
    },
    async createTrainingV2(training: TrainingCreationRequestV2): Promise<TrainingV2> {
      return trainingService.createTrainingV2(training);
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
