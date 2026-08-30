import type { EWorkloadName } from "@/models/workload.model";
import type { DistributedFramework } from "@/swagger-models/assets-service-client";

export interface IWorkloadTypesSectionModel {
  workloadTypes?: Record<EWorkloadName, boolean>;
  distributed: {
    isDistributed: boolean;
    distFramework: DistributedFramework | null;
  };
}
