import type { AssetIdAndKind } from "@/swagger-models/assets-service-client";

export interface INimModelStoreSection {
  selectedDataSources: Array<AssetIdAndKind>;
}
