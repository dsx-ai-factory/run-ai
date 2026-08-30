import type { TrainingSpecificRunParams, WorkspaceSpecificRunParams } from "@/swagger-models/assets-service-client";
import type { IUIWorkloadSpecificEnv } from "@/models/workload.model";

export interface IWorkloadSpecificRunParams extends WorkspaceSpecificRunParams, TrainingSpecificRunParams {}
export type SpecificRunParamsKeys = (keyof IUIWorkloadSpecificEnv)[];
