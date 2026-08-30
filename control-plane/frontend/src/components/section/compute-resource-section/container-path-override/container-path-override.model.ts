import { type AssetIdAndKind } from "@/swagger-models/assets-service-client";

export interface ISelectedDatasourcesContainerPathData extends AssetIdAndKind {
  name: string;
  origin: string;
}

export interface IContainerPathOverride {
  dataSourceId: string;
  containerPathOverride: string;
}

export interface ITabProps {
  key: string;
  id: string;
  label: string;
  isInvalid?: boolean;
  icon?: string;
  removeable?: boolean;
}
