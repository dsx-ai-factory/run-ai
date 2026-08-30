import type { StringOption } from "@/swagger-models/assets-service-client";
import type { UnifiedPolicyInfoPerReplicaType, UnifiedProjectInfo } from "@/swagger-models/workloads-client";
import type { IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import type { IUIEnvironmentFields, IUISuperWorkloadCreationRequest, IUISuperWorkloadSpec } from "./workload.model";
import type { WorkloadCreationDrawerEntities } from "./workload-creation-drawer-model";

export type THybridFormState = {
  projectInfo: UnifiedProjectInfo;
  projectNodePoolsResources: StringOption[];
  selectedTemplates: THybridSelectedTemplates;
  policy: UnifiedPolicyInfoPerReplicaType | undefined | null;
  policyViewDetails: IPolicyViewDetails;
  selectedProjectNodeTypes: string[];
  workloadModel: IUISuperWorkloadCreationRequest;
  completedSteps?: string[];
};

export type THybridSelectedTemplates = {
  [WorkloadCreationDrawerEntities.Environment]: {
    name: string;
    id: string;
    defaultData: IUIEnvironmentFields | null;
    clearedState: IUIEnvironmentFields | null;
    renderKey: number;
    sectionKey: number;
    isLinked?: boolean;
  };
  [WorkloadCreationDrawerEntities.Compute]: {
    name: string;
    id: string;
    defaultData: IUISuperWorkloadSpec["compute"] | null;
    clearedState: IUISuperWorkloadSpec["compute"] | null;
    renderKey: number;
    sectionKey: number;
    isLinked?: boolean;
  };
  [WorkloadCreationDrawerEntities.DataSource]: {
    name: string;
    id: string;
    defaultData: IUISuperWorkloadSpec["storage"] | null;
    clearedState: IUISuperWorkloadSpec["storage"] | null;
    sectionKey: number;
  };
  [WorkloadCreationDrawerEntities.ModelStorage]: {
    name: string;
    id: string;
    defaultData: IUISuperWorkloadSpec["storage"] | null;
    clearedState: IUISuperWorkloadSpec["storage"] | null;
    sectionKey: number;
  };
};

export const enum ESubmissionOptions {
  Hybrid = "hybrid",
  Assets = "assets",
}

export const hybridSubmissionOptions = [
  { label: "Flexible (recommended)", value: ESubmissionOptions.Hybrid },
  { label: "Original", value: ESubmissionOptions.Assets },
];

export interface IHideSpecificSections {
  environment?: IHideSpecificSectionsEnvironment;
  computeResource?: IHideSpecificSectionsComputeResource;
}

export interface IHideSpecificSectionsComputeResource {
  nodes?: boolean;
  tolerations?: boolean;
  affinity?: boolean;
  topology?: boolean;
}
export interface IHideSpecificSectionsEnvironment {
  image?: boolean;
  imagePullSecrets?: boolean;
  configMapAsEnvVar?: boolean;
  userCredentials?: boolean;
  tools?: boolean;
}
