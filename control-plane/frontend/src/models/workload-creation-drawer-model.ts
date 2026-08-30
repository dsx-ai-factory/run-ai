import type { AssetKind, ComplianceInfo } from "@/swagger-models/assets-service-client";

export interface IDrawerListItem {
  id: string;
  name: string;
  description?: string;
  kind?: AssetKind;
  policy?: ComplianceInfo | null;
  icons?: IDrawerListItemIcon[];
  createdAt?: string;
  lastUsed?: string;
  disabled?: boolean;
  disabledInfo?: string;
  details: TItemDetail[];
}

export interface IDrawerListItemIcon {
  name: string;
  icon: string;
}

export const enum WorkloadCreationDrawerEntities {
  DataSource = "data-source",
  Environment = "environment",
  Compute = "compute-resource",
  ModelStorage = "model-storage",
}

export const enum ExtendedSections {
  POLICY = "policy",
  DETAILS = "details",
}

export type TItemDetail = {
  title?: string;
  content: string[];
  icon?: string;
  sortPriority?: number;
  spaceContent?: boolean;
  noSeperator?: boolean;
  subtitle?: string;
};
