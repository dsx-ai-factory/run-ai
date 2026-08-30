import type { Scope, SpecificRunCreationFields } from "@/swagger-models/assets-service-client";
import type { IUIDistributed, IUIWorkloadAssets, IUIWorkloadSpecificEnv } from "./workload.model";
import type { IUIAssetCreationMeta } from "./global.model";

export interface IUIWorkloadTemplateCreation {
  name: string;
  description?: string | null;
  scope: Scope | null;
  projectId?: number | null;
  departmentId?: string | null;
  namespace: string;
  clusterId?: string | null;
  assets: IUIWorkloadAssets;
  specificEnv: IUIWorkloadSpecificEnv;
  distributed?: IUIDistributed | null;
  enableEditingMaster?: boolean;
}

export interface IUIWorkloadTemplateCreationRequestModel {
  meta: IUIAssetCreationMeta;
  spec: SpecificRunCreationFields;
}
