import type { IDetailsPanelItem, TDetailsPanelItemGroup } from "@/models/details-panel.model";

export interface IPvcGeneralSectionData {
  groups: TDetailsPanelItemGroup[];
}

export interface IPvcRequestSectionData {
  items: IDetailsPanelItem[];
}

export interface IPvcStorageClassSectionData {
  items: IDetailsPanelItem[];
}
