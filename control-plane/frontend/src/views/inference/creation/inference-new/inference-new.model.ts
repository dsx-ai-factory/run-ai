import type { EInferenceArchitecture, EServerType } from "@/models/inference.model";

export interface IInferenceNewModel {
  name: string;
  clusterId: string;
  projectId: number;
  serverType: EServerType;
  architecture: EInferenceArchitecture;
  templateId: string | null;
}
