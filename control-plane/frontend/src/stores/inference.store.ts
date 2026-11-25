/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Stores
import { defineStore } from "pinia";
import { useClusterStore } from "./cluster.store";

// Models
import type { IUISuperWorkloadCreationRequest } from "@/models/workload.model";
import {
  getEmptyInferenceNewModel,
  type IInferenceNewModel,
} from "@/views/inference/creation/inference-new/inference-new.model";
import { ResourceType, type Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import { EServerType } from "@/models/inference.model";

// services
import { permissionService } from "@/services/authorization/permission.service/permission.service";

// constants
import { MIN_CLUSTER_VERSION_FOR_SERVING_ENDPOINT_ACCESS } from "@/common/version.constant";
import type { THybridFormState } from "@/models/workload-hybrid.model";
import {
  ERunaiCredentialSource,
  type ERunaiCredentialType,
} from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";
import type { ETokenProvisionMethods } from "@/components/workload-creation/hybrid-hugging-face-model-section/hugging-face-access-section/hugging-face-access-section.model";
import type { ENimImagePullCredentialType } from "@/components/workload-creation/hybrid-nim-model-section/nim-access-model/nim-image-pull-credential-selector/nim-image-pull-credential-selector.model";

export interface IInferenceFormUIState {
  accessFromStorage?: boolean;
  credentialsSource?: ERunaiCredentialSource;
  credentialType?: ERunaiCredentialType;
  modelName?: string;
  modelBasedOn?: EServerType;
  selectedModelSource?: string;
  hfSelectedTokenProvisionMethod?: ETokenProvisionMethods;
  nimHfImagePullCredentialType?: ENimImagePullCredentialType;
}

function getEmptyInferenceFormUIState(): IInferenceFormUIState {
  return {
    accessFromStorage: false,
    credentialsSource: ERunaiCredentialSource.SharedSecret,
    modelName: "",
    modelBasedOn: undefined,
  };
}

export interface IInferenceDistributedState {
  leader: THybridFormState;
  worker: THybridFormState | null;
  isCopyFlow?: boolean;
}

export const useInferenceStore = defineStore("Inference", {
  state: () => ({
    inference: getEmptyInferenceNewModel(),

    // flat submission data ONLY
    flatFormState: null as THybridFormState | null,
    presetFormModel: null as IUISuperWorkloadCreationRequest | null,
    distributedState: null as IInferenceDistributedState | null,
    // we need to keep some UI states that is not part of the flat workload model
    // because it is part of the UI and not part of the workload model
    inferenceFormUIState: getEmptyInferenceFormUIState() as IInferenceFormUIState,
    actionPermissionsByScopes: {} as Record<Action, PermittedScopes>,

    // template model
    templateModel: null as IUISuperWorkloadCreationRequest | null,
  }),
  getters: {
    isHuggingFaceOrNimModel(): boolean {
      return (
        this.inferenceFormUIState.modelBasedOn === EServerType.VLLM ||
        this.inferenceFormUIState.modelBasedOn === EServerType.Nim
      );
    },
    supportingServingPortAccess:
      () =>
      (clusterId: string): boolean => {
        const clusterStore = useClusterStore();
        return clusterStore.isClusterVersionSufficient(clusterId, MIN_CLUSTER_VERSION_FOR_SERVING_ENDPOINT_ACCESS);
      },
  },
  actions: {
    async loadActionPermissions(): Promise<void> {
      if (Object.keys(this.actionPermissionsByScopes).length > 0) return;
      this.actionPermissionsByScopes = await permissionService.getPermittedScopesForAction(ResourceType.Inferences);
    },
    setInference(inference: IInferenceNewModel): void {
      this.inference = inference;
    },
    resetCreateInferenceData(): void {
      this.inference = getEmptyInferenceNewModel();
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
    setInferenceFormUIState(inferenceFormUIState: IInferenceFormUIState): void {
      this.inferenceFormUIState = inferenceFormUIState;
    },
    getInferenceFormUIState(): IInferenceFormUIState {
      return this.inferenceFormUIState;
    },
    clearInferenceFormUIState(): void {
      this.inferenceFormUIState = getEmptyInferenceFormUIState();
    },
    setPresetFormModel(workloadData: IUISuperWorkloadCreationRequest | null): void {
      this.presetFormModel = workloadData;
    },
    getPresetFormModel(): IUISuperWorkloadCreationRequest | null {
      return this.presetFormModel;
    },
    clearPresetFormModel(): void {
      this.presetFormModel = null;
    },
    setTemplateModel(templateModel: IUISuperWorkloadCreationRequest): void {
      this.templateModel = templateModel;
    },
    getTemplateModel(): IUISuperWorkloadCreationRequest | null {
      return this.templateModel;
    },
    clearTemplateModel(): void {
      this.templateModel = null;
    },
    clearAllStates(): void {
      this.clearTemplateModel();
      this.clearPresetFormModel();
      this.clearInferenceFormUIState();
      this.clearFlatFormState();
    },
    saveDistributedFormState(distributedState: IInferenceDistributedState): void {
      this.distributedState = distributedState;
    },
    getDistributedState(): IInferenceDistributedState | null {
      return this.distributedState;
    },
    clearDistributedFormState(): void {
      this.distributedState = null;
    },
  },
});
