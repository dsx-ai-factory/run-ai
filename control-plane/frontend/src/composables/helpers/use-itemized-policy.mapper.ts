import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import type { EnvironmentVariableOfAsset, Toleration } from "@/swagger-models/assets-service-client";
import type { ExtendedResource } from "@/swagger-models/policy-service-client";

interface IIndexTableInfo {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getEntityIdentifierValue: (e: any) => string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getDefinedExcludeValue: (e: any) => boolean;
}

export enum ItemizedMapperEntities {
  EnvironmentVariables = "environmentVariables",
  Tolerations = "tolerations",
  ExtendedResources = "extendedResources",
  DefaultNameValue = "defaultNameValue",
}

export const useItemizedPolicyMapper: Record<ItemizedMapperEntities, IIndexTableInfo> = {
  [ItemizedMapperEntities.EnvironmentVariables]: {
    getEntityIdentifierValue: (environmentVariable: EnvironmentVariableOfAsset) => environmentVariable.name || "",
    getDefinedExcludeValue: (environmentVariable: EnvironmentVariableOfAsset) => environmentVariable.exclude || false,
  },
  [ItemizedMapperEntities.Tolerations]: {
    getEntityIdentifierValue: (toleration: Toleration) => toleration.name || "",
    getDefinedExcludeValue: (toleration: Toleration) => toleration.exclude || false,
  },
  [ItemizedMapperEntities.ExtendedResources]: {
    getEntityIdentifierValue: (listItem: ExtendedResource) => listItem.resource || "",
    getDefinedExcludeValue: (listItem: ExtendedResource) => listItem.exclude || false,
  },
  [ItemizedMapperEntities.DefaultNameValue]: {
    // for both annotations and labels
    getEntityIdentifierValue: (listItem: IItemizedListItem) => listItem.name || "",
    getDefinedExcludeValue: (listItem: IItemizedListItem) => listItem.exclude || false,
  },
};
