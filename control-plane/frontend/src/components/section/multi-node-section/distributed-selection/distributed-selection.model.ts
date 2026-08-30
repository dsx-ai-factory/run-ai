import type { DistributedFramework } from "@/swagger-models/assets-service-client";

export interface IDistributedSelectionModel {
  isDistributed: boolean;
  distFramework: DistributedFramework | null;
}
