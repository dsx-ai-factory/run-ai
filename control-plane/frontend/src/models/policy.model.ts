import type {
  TrainingPolicyChangeRequestV2,
  WorkspacePolicyChangeRequestV2,
  DistributedPolicyChangeRequestV2,
  WorkspacePolicyV2,
  TrainingPolicyV2,
  DistributedPolicyV2,
  DistributedInferencePolicyV2,
  DistributedPolicyDefaultsAndRulesV2,
  TrainingPolicyDefaultsAndRulesV2,
  WorkspacePolicyDefaultsAndRulesV2,
  DistributedInferencePolicyDefaultsAndRulesV2,
  InferencePolicyV2,
  SourceOfRule,
  EnvironmentVariablesDefaults,
  AnnotationsDefaults,
  LabelsDefaults,
  TolerationsDefaults,
} from "@/swagger-models/policy-service-client";

export type TPolicyScopes = "tenant" | "cluster" | "department" | "project";

export type TPolicyChangeRequest =
  | TrainingPolicyChangeRequestV2
  | WorkspacePolicyChangeRequestV2
  | DistributedPolicyChangeRequestV2;

export type TWorkloadPolicy = WorkspacePolicyV2 | TrainingPolicyV2 | DistributedPolicyV2 | DistributedInferencePolicyV2;

export type TPolicyDefaultAndRules =
  | WorkspacePolicyDefaultsAndRulesV2
  | TrainingPolicyDefaultsAndRulesV2
  | DistributedPolicyDefaultsAndRulesV2
  | DistributedInferencePolicyDefaultsAndRulesV2;

export interface IPolicyMap {
  workspace: WorkspacePolicyV2;
  training: TrainingPolicyV2;
  distributed: DistributedPolicyV2;
  inference: InferencePolicyV2;
  distributedInference: DistributedInferencePolicyV2;
}

export interface IGenericSelectPolicyRules {
  sourceOfRule?: SourceOfRule | null;
  required?: boolean | null;
  canEdit?: boolean | null;
  options?: Array<{ displayed?: string | null | undefined; value: string | null }> | null;
  isClosedList?: boolean;
}

export interface IGenericRadioOptionsPolicyRules {
  sourceOfRule?: SourceOfRule | null;
  required?: boolean | null;
  canEdit?: boolean | null;
  options?: Array<{ displayed?: string | null | undefined; value: string | null }> | null;
}

export interface ILoadedItemizeDefaults {
  environmentVariables?: EnvironmentVariablesDefaults | null;
  annotations?: AnnotationsDefaults | null;
  labels?: LabelsDefaults | null;
  tolerations?: TolerationsDefaults | null;
}
