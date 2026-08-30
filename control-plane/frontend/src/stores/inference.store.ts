// Stores
import { defineStore } from "pinia";
import { useWorkloadStore } from "./workload.store";
import { useClusterStore } from "./cluster.store";

// Models
import {
  type AssetsDefaults,
  type Inference,
  type InferenceCreationRequestV2,
  type InferenceV2,
  type PolicyInfo,
  type InferenceSpecificRunParams,
  Scope,
} from "@/swagger-models/assets-service-client";
import { EWorkloadType, type IUISuperWorkloadCreationRequest, type IUIInferenceCreation } from "@/models/workload.model";
import { ResourceType, type Action, type PermittedScopes } from "@/swagger-models/authorization-client";
import { EServerType } from "@/models/inference.model";

// utils
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { deepCopy } from "@/utils/common.util";

// services
import { inferenceService } from "@/services/control-plane/inference.service/inference.service";
import { dataSourceService } from "@/services/control-plane/data-source.service/data-source.service";
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
    inference: inferenceWorkloadUtil.getEmptyUIInferenceCreation() as IUIInferenceCreation,
    loadedInferenceModelWithDefaultsAndPolicy: null as InferenceV2 | null,
    selectedInference: null as Inference | null,

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
        this.inferenceFormUIState.modelBasedOn === EServerType.TGI ||
        this.inferenceFormUIState.modelBasedOn === EServerType.VLLM ||
        this.inferenceFormUIState.modelBasedOn === EServerType.Nim
      );
    },
    inferenceName(): string {
      return this.inference.name || "";
    },
    loadedPolicyInfo(): PolicyInfo | null | undefined {
      return this.loadedInferenceModelWithDefaultsAndPolicy?.policyInfo || null;
    },
    // these details are treated as read only after setting (saving the origin model with policies set)
    loadedSpecificRunParamsWithDefaults(): InferenceSpecificRunParams | null {
      return deepCopy(this.loadedInferenceModelWithDefaultsAndPolicy?.specificRunParams || null);
    },
    loadedAssetsAttributesDefaults(): AssetsDefaults | null {
      return this.loadedInferenceModelWithDefaultsAndPolicy?.policyInfo?.attributeDefaults?.assets || null;
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
    async createInference(inference: IUIInferenceCreation): Promise<Inference | InferenceV2> {
      const workloadStore = useWorkloadStore();

      let workloadVolumes: Array<string> | undefined;
      if (inference.assets.uiVolumes?.length) {
        workloadVolumes = await dataSourceService.createWorkloadVolumes(inference.name, inference.assets.uiVolumes, {
          scope: Scope.Project,
          projectId: inference.projectId,
        });
      }

      const workloadCreationRequest: InferenceCreationRequestV2 = inferenceWorkloadUtil.getInferenceCreationRequestV2(
        inference,
        this.supportingServingPortAccess(inference.clusterId),
        workloadVolumes,
      );
      const workloadCreated = await inferenceService.createInferenceV2(workloadCreationRequest);

      workloadStore.setWorkloadAdded(workloadCreated, EWorkloadType.Inference);
      return workloadCreated;
    },
    setInference(inference: IUIInferenceCreation): void {
      this.inference = inference;
    },
    resetCreateInferenceData(): void {
      this.inference = inferenceWorkloadUtil.getEmptyUIInferenceCreation();
      this.loadedInferenceModelWithDefaultsAndPolicy = null;
    },
    setLoadedInferenceModelWithDefaults(inference: InferenceV2): void {
      this.loadedInferenceModelWithDefaultsAndPolicy = inference;
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
