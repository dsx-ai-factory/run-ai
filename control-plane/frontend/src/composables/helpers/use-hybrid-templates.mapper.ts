import {
  EWorkloadType,
  type IUIEnvironmentFields,
  type IUISuperWorkloadCreationRequest,
  type IUISuperWorkloadSpec,
  type IUIWorkloadStorageFields,
  type IWorkloadSupportedRoutes,
} from "@/models/workload.model";
import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";
import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import {
  type SupersetSpec,
  type UnifiedWorkload,
  type UnifiedWorkloadRequest,
  WorkloadType,
} from "@/swagger-models/workloads-client";
import { inferenceWorkloadUtil } from "@/utils/inference-workload.util/inference-workload.util";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { workloadUtil } from "@/utils/workload.util";

interface IUseHybridTemplatesMapper {
  getUnified: (req: UnifiedWorkloadRequest) => Promise<UnifiedWorkload>;
  getRelevantEnvironmentFields: (model: IUISuperWorkloadCreationRequest) => IUIEnvironmentFields;
  getUIModel: (unifiedWorkload: UnifiedWorkload) => IUISuperWorkloadCreationRequest;
  creationRoutes?: IWorkloadSupportedRoutes;
  templateCreationRoutes?: IWorkloadSupportedRoutes;
  combineWorkloadModelWithEnvironmentDetails?: (
    initialWorkloadModel: IUISuperWorkloadCreationRequest,
    envFields: IUIEnvironmentFields,
  ) => IUISuperWorkloadCreationRequest;
  getRelevantComputeFields?: (model: IUISuperWorkloadCreationRequest) => SupersetSpec["compute"];
  combineWorkloadModelWithComputeDetails?: (
    initialWorkloadModel: IUISuperWorkloadCreationRequest,
    envFields: IUISuperWorkloadSpec["compute"],
  ) => IUISuperWorkloadCreationRequest;
  getRelevantStorageFields?: (model: IUISuperWorkloadCreationRequest) => IUIWorkloadStorageFields;
  combineWorkloadModelWithStorageDetails?: (
    initialWorkloadModel: IUISuperWorkloadCreationRequest,
    storageFields: IUIWorkloadStorageFields,
  ) => IUISuperWorkloadCreationRequest;
}

export enum EHybridFormTypes {
  Workspace = "Workspace",
  Training = "Training",
  DistributedMaster = "DistributedMaster",
  DistributedWorker = "DistributedWorker",
  Inference = "Inference",
  DistributedInferenceLeader = "DistributedInferenceLeader",
  DistributedInferenceWorker = "DistributedInferenceWorker",
}

export const useHybridTemplatesMapper: Record<EHybridFormTypes, IUseHybridTemplatesMapper> = {
  [EHybridFormTypes.Workspace]: {
    creationRoutes: workloadUtil.getSupportedRoutes(
      WORKSPACE_ROUTE_NAMES.WORKSPACE_HYBRID,
      WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW,
    ),
    templateCreationRoutes: workloadUtil.getSupportedRoutes(
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_HYBRID,
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
    ),
    getRelevantEnvironmentFields: workloadHybridUtil.getRelevantEnvironmentFields,
    getUIModel: workloadHybridUtil.getHybridWorkspaceModel,
    getUnified: async (options?: UnifiedWorkloadRequest) => {
      const request: UnifiedWorkloadRequest = {
        templateId: options?.templateId,
        clusterId: options?.clusterId,
        projectId: options?.projectId,
        workloadType: EWorkloadType.Workspace,
        workloadId: options?.workloadId,
        assetId: options?.assetId,
        assetIds: options?.assetIds,
        dataVolumeIds: options?.dataVolumeIds,
      };
      return await workloadService.getUnifiedWorkload(request);
    },
  },
  [EHybridFormTypes.Training]: {
    creationRoutes: workloadUtil.getSupportedRoutes(
      TRAINING_ROUTE_NAMES.TRAINING_HYBRID,
      TRAINING_ROUTE_NAMES.TRAINING_NEW,
    ),
    templateCreationRoutes: workloadUtil.getSupportedRoutes(
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID,
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
    ),
    getRelevantEnvironmentFields: workloadHybridUtil.getRelevantEnvironmentFields,
    getUIModel: workloadHybridUtil.getHybridTrainingModel,
    getUnified: async (options?: UnifiedWorkloadRequest) => {
      const request: UnifiedWorkloadRequest = {
        templateId: options?.templateId,
        clusterId: options?.clusterId,
        projectId: options?.projectId,
        workloadType: EWorkloadType.Training,
        workloadId: options?.workloadId,
        assetId: options?.assetId,
        assetIds: options?.assetIds,
        dataVolumeIds: options?.dataVolumeIds,
      };
      return await workloadService.getUnifiedWorkload(request);
    },
  },
  [EHybridFormTypes.DistributedWorker]: {
    creationRoutes: workloadUtil.getSupportedRoutes(
      TRAINING_ROUTE_NAMES.DISTRIBUTED_HYBRID,
      TRAINING_ROUTE_NAMES.TRAINING_NEW,
    ),
    templateCreationRoutes: workloadUtil.getSupportedRoutes(
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID_DISTRIBUTED,
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
    ),
    getRelevantEnvironmentFields: workloadHybridUtil.getRelevantEnvironmentFields,
    getUIModel: workloadHybridUtil.getHybridDistributedModel,
    getUnified: async (options?: UnifiedWorkloadRequest) => {
      const request: UnifiedWorkloadRequest = {
        templateId: options?.templateId,
        clusterId: options?.clusterId,
        projectId: options?.projectId,
        workloadType: EWorkloadType.Distributed,
        workloadId: options?.workloadId,
        assetId: options?.assetId,
        assetIds: options?.assetIds,
        dataVolumeIds: options?.dataVolumeIds,
      };
      return await workloadService.getUnifiedWorkload(request);
    },
  },
  [EHybridFormTypes.DistributedMaster]: {
    creationRoutes: workloadUtil.getSupportedRoutes(
      TRAINING_ROUTE_NAMES.DISTRIBUTED_HYBRID,
      TRAINING_ROUTE_NAMES.TRAINING_NEW,
    ),
    templateCreationRoutes: workloadUtil.getSupportedRoutes(
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID_DISTRIBUTED,
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_NEW,
    ),
    getRelevantEnvironmentFields: workloadHybridUtil.getRelevantEnvironmentFields,
    getUIModel: workloadHybridUtil.getHybridDistributedMasterModel,
    getUnified: async (options?: UnifiedWorkloadRequest) => {
      const request: UnifiedWorkloadRequest = {
        clusterId: options?.clusterId,
        projectId: options?.projectId,
        workloadType: EWorkloadType.Distributed,
        workloadId: options?.workloadId,
        assetId: options?.assetId,
        assetIds: options?.assetIds,
        dataVolumeIds: options?.dataVolumeIds,
      };
      return await workloadService.getUnifiedWorkload(request);
    },
  },
  [WorkloadType.Inference]: {
    creationRoutes: workloadUtil.getSupportedRoutes(
      INFERENCE_ROUTE_NAMES.INFERENCE_FLAT_HYBRID,
      INFERENCE_ROUTE_NAMES.INFERENCE_NEW,
    ),
    getRelevantEnvironmentFields: workloadHybridUtil.getInferenceRelevantEnvironmentFields,
    getUIModel: inferenceWorkloadUtil.getHybridInferenceModel,
    getUnified: async (options?: UnifiedWorkloadRequest) => {
      const request: UnifiedWorkloadRequest = {
        clusterId: options?.clusterId,
        projectId: options?.projectId,
        workloadType: EWorkloadType.Inference,
        workloadId: options?.workloadId,
        assetId: options?.assetId,
        assetIds: options?.assetIds,
        dataVolumeIds: options?.dataVolumeIds,
        templateId: options?.templateId,
      };
      return await workloadService.getUnifiedWorkload(request);
    },
    templateCreationRoutes: workloadUtil.getSupportedRoutes(
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_CUSTOM_NEW,
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NEW,
    ),
  },
  [EHybridFormTypes.DistributedInferenceLeader]: {
    creationRoutes: workloadUtil.getSupportedRoutes(
      INFERENCE_ROUTE_NAMES.INFERENCE_DISTRIBUTED_EDIT,
      INFERENCE_ROUTE_NAMES.INFERENCE_NEW,
    ),
    getRelevantEnvironmentFields: workloadHybridUtil.getDistributedInferenceLeaderRelevantEnvironmentFields,
    getUIModel: inferenceWorkloadUtil.getHybridDistributedModel,
    getUnified: async (options?: UnifiedWorkloadRequest) => {
      const request: UnifiedWorkloadRequest = {
        clusterId: options?.clusterId,
        projectId: options?.projectId,
        workloadType: EWorkloadType.DistributedInference,
        workloadId: options?.workloadId,
        assetId: options?.assetId,
        assetIds: options?.assetIds,
        dataVolumeIds: options?.dataVolumeIds,
        templateId: options?.templateId,
      };
      return await workloadService.getUnifiedWorkload(request);
    },
    templateCreationRoutes: workloadUtil.getSupportedRoutes(
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_CUSTOM_NEW,
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NEW,
    ),
    combineWorkloadModelWithEnvironmentDetails:
      workloadHybridUtil.combineDistributedInferenceLeaderWorkloadModelWithEnvironmentDetails,
    getRelevantComputeFields: (model: IUISuperWorkloadCreationRequest) => {
      return model.spec.leader?.compute || null;
    },
    combineWorkloadModelWithComputeDetails:
      workloadHybridUtil.combineDistributedInferenceLeaderWorkloadModelWithComputeDetails,
    getRelevantStorageFields: (model: IUISuperWorkloadCreationRequest) => model.spec.leader?.storage || {},
    combineWorkloadModelWithStorageDetails:
      workloadHybridUtil.combineDistributedInferenceLeaderWorkloadModelWithStorageDetails,
  },
  [EHybridFormTypes.DistributedInferenceWorker]: {
    creationRoutes: workloadUtil.getSupportedRoutes(
      INFERENCE_ROUTE_NAMES.INFERENCE_DISTRIBUTED_EDIT,
      INFERENCE_ROUTE_NAMES.INFERENCE_NEW,
    ),
    getRelevantEnvironmentFields: (model: IUISuperWorkloadCreationRequest) =>
      workloadHybridUtil.getBaseEnvironmentFields(model.spec.worker || {}),
    getUIModel: inferenceWorkloadUtil.getHybridDistributedModel,
    getUnified: async (options?: UnifiedWorkloadRequest) => {
      const request: UnifiedWorkloadRequest = {
        clusterId: options?.clusterId,
        projectId: options?.projectId,
        workloadType: EWorkloadType.DistributedInference,
        workloadId: options?.workloadId,
        assetId: options?.assetId,
        assetIds: options?.assetIds,
        dataVolumeIds: options?.dataVolumeIds,
        templateId: options?.templateId,
      };
      return await workloadService.getUnifiedWorkload(request);
    },
    templateCreationRoutes: workloadUtil.getSupportedRoutes(
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_CUSTOM_NEW,
      WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NEW,
    ),
    combineWorkloadModelWithEnvironmentDetails:
      workloadHybridUtil.combineDistributedInferenceWorkerWorkloadModelWithEnvironmentDetails,
    getRelevantComputeFields: (model: IUISuperWorkloadCreationRequest) => {
      return model.spec.worker?.compute || null;
    },
    combineWorkloadModelWithComputeDetails:
      workloadHybridUtil.combineDistributedInferenceWorkerWorkloadModelWithComputeDetails,
    getRelevantStorageFields: (model: IUISuperWorkloadCreationRequest) => model.spec.worker?.storage || {},
    combineWorkloadModelWithStorageDetails:
      workloadHybridUtil.combineDistributedInferenceWorkerWorkloadModelWithStorageDetails,
  },
};
