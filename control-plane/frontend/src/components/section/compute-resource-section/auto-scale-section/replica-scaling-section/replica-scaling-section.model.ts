export const MAX_REPLICAS_ALLOWED = 10000;

export interface IReplicaScalingSection {
  minReplicas: number | undefined;
  maxReplicas: number | undefined;
}

export enum EReplicaScalingSectionEvents {
  UpdateModelValue = "update:model-value",
}
