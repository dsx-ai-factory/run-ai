import type { EnvironmentVariableOfAsset, SpecificRunConnectionInfo } from "@/swagger-models/assets-service-client";
import type { IServingEndpointAccessData } from "@/components/section/inference/serving-endpoint-section/serving-endpoint-access.model";

export interface IUIWorkloadEnvSectionModel {
  args: string;
  command?: string;
  environmentVariables: Array<EnvironmentVariableOfAsset>;
  connections: Array<SpecificRunConnectionInfo>;
  runAsUid?: number | null;
  runAsGid?: number | null;
  supplementalGroups?: string | null;
  servingPortAccess?: IServingEndpointAccessData | null;
}

export interface IWorkloadEnvSectionOptions {
  isToolBoxRequired?: boolean;
  cardsTitle?: string;
  canAddEnvVariable?: boolean;
}
