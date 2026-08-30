export enum Status {
  SUCCEEDED = "Succeeded",
  FAILED = "Failed",
  RUNNING = "Running",
  COMPLETE = "Complete",
  PROGRESSING = "Progressing",
  PENDING = "Pending",
  DELETED = "Deleted",
  SUSPENDED = "Suspended",
  SWAPPED = "Swapped",
}

interface ValueType {
  value: string;
  type: string;
}

export interface Pod {
  clusterId: string;
  created: number;
  started: number;
  completed: number;
  name: string;
  podType: string;
  nodeId: string;
  nodePool: string;
  podId: string;
  project: string;
  status: Status;
  imageName: string;
  user: string;
  dynamicData: Record<string, string | number | ValueType>;
  training_lib: string;
  totalRuntime: number;
  totalWaitTime: number;
}
