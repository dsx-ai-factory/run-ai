import type { GVK, MappingFieldType } from "@/swagger-models/policy-service-client";

export type { GVK, PolicyRule } from "@/swagger-models/policy-service-client";

export enum EEnforcementType {
  Set = "set",
  Default = "default",
  Required = "required",
  Blocked = "blocked",
  Match = "match",
  Options = "options",
  Length = "length",
  Range = "range",
  QuantityRange = "quantityRange",
  Elements = "elements",
  Items = "items",
}

export interface IGvkWorkloadTypeOption {
  key: string;
  label: string;
  icon?: string;
  gvk: GVK;
}

export interface IMappingParam {
  name: string;
  type: MappingFieldType;
  enumValues?: string[];
  required?: boolean;
}

export interface IFieldInfo {
  type: MappingFieldType;
  builtin: string | null;
  /** Formatter argument that identifies an independent Set/Default target; null means the field is flat. */
  setDefaultGroupingArg: string | null;
  itemSelectors: string[];
  allowedEnforcements: EEnforcementType[];
  setParams: IMappingParam[];
  specSelectorParams: IMappingParam[];
  itemSelectorParams: IMappingParam[];
}

export interface IEnforcementDraft {
  id: string;
  type: EEnforcementType | null;
  value: string;
  setParams: Record<string, string>;
  options: string[];
  min: string;
  max: string;
  step: string;
}

export interface IRuleDraft {
  id: string;
  field: string | null;
  specSelector: string | null;
  itemSelector: string | null;
  selectorParams: Record<string, string>;
  enforcements: IEnforcementDraft[];
}
