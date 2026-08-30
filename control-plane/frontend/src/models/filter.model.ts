import type { DistributedFramework, Scope } from "@/swagger-models/assets-service-client";
import type { EWorkloadTypeSupportedInAssetsService } from "@/models/workload.model";
import type { ISelectOption } from "@/models/global.model";

export enum ESortOrder {
  Asc = "asc",
  Desc = "desc",
}

export const DEFAULT_PAGE_SIZE = 20;

export const CLUSTER_COLUMN_FILTER_NAME = "cluster";
export const ALL_CLUSTER_FILTER_VALUE = "";
export const ALL_CLUSTER_FILTER_LABEL = "All";
export type TFilterField = string | ((row: any) => any); // eslint-disable-line @typescript-eslint/no-explicit-any
export const GLOBAL_FILTER_ICON = "ph-fill ph-globe";

export interface IFilterModel {
  field: TFilterField;
  name: string;
  term: string;
}

export interface IFilterOption {
  label: string;
  field: TFilterField;
  name: string;
  autocompleteOptionsFn?: (searchQuery: string, clusterUuid?: string) => Promise<string[]>;
}

export interface ILabeledFilter {
  field: TFilterField;
  term: string;
  label: string;
  name: string;
}

export interface IFilterBy {
  sortBy?: string | null;
  descending?: boolean;
  query?: string;
  page?: number;
  rowsPerPage?: number;
  columnFilters?: Array<IFilterModel>;
  searchTerm?: string;
  displayedColumns?: Array<string>;
  forceLoad?: boolean;
  clusterUuid?: string;
  advancedFilters?: IAdvancedFilterModel[];
  isAllClustersFilter?: boolean;
  /**
   * For server-side fetching only
   */
  rowsNumber?: number; // How many total database rows are there to be added to the table. If set, causes the QTable to emit @request when data is required.
}

export interface IPaginationFilter {
  offset?: number;
  limit?: number;
  sortOrder?: ESortOrder;
  sortBy?: string | null;
  filterBy?: Array<string>;
  search?: string;
  deleted?: boolean;
}

export interface IJobFilter extends IFilterBy {
  type?: string;
}
export interface IAssetsFilter extends IFilterBy {
  projectId?: number | null;
  departmentId?: string | null;
  scope?: Scope;
  workspaceId?: string;
  workloadTemplateId?: string;
  usageInfo?: boolean;
  complyToProject?: number;
  complyToWorkloadType?: EWorkloadTypeSupportedInAssetsService;
  isWorkspace?: boolean;
  isTraining?: boolean;
  isInference?: boolean;
  isDistributed?: boolean;
  isDistributedInference?: boolean;
  distributedFramework?: DistributedFramework | null;
  clusterUuid?: string;
  statusInfo?: boolean;
  complyToReplicaType?: "Worker" | "Master" | "Leader";
  assetIds?: string;
  dataVolumes?: boolean;
  includeDescendants?: boolean;
  privateOnly?: boolean;
  includePrivate?: boolean;
}

// ------ Advanced Filter -------

export enum EDateFilterOperator {
  Before = "before",
  On = "on",
  After = "after",
}

export enum EFilterOperator {
  Equals = "==",
  NotEquals = "!=",
  LessThanOrEqual = "<=",
  GreaterThanOrEqual = ">=",
  Contains = "=@",
  DoesNotContain = "!@",
  StartsWith = "=^",
  EndsWith = "=$",
  Range = "range",
}

export const numericFilterSupportedOperators: EFilterOperator[] = [
  EFilterOperator.Range,
  EFilterOperator.Equals,
  EFilterOperator.NotEquals,
  EFilterOperator.LessThanOrEqual,
  EFilterOperator.GreaterThanOrEqual,
];

export enum EFilterOperatorLabel {
  Equals = "equals",
  NotEquals = "does not equal",
  LessThanOrEqual = "less than or equal",
  GreaterThanOrEqual = "greater than or equal",
  Contains = "contains",
  DoesNotContain = "does not contain",
  StartsWith = "starts with",
  EndsWith = "ends with",
  Range = "range",
}

export interface IAdvancedFilterOption extends IFilterOption {
  type: EColumnFilterType;
  selectOptions?: ISelectOption[];
  excludeOperators?: EFilterOperator[];
  inputType?: "checkbox" | "radio";
  autocompleteOptionsFn?: (searchQuery: string, clusterUuid?: string) => Promise<string[]>;
}

export enum EColumnFilterType {
  FreeText = "FreeText",
  EnumBased = "EnumBased",
  Numeric = "Numeric",
  Date = "Date",
}

interface IFilterColumnModel {
  name: string;
  label: string;
  type: EColumnFilterType;
}

export interface IEnumBasedFilterSelectOption extends ISelectOption {
  hidden?: boolean;
  value: string | boolean;
}

export interface IEnumBasedFilterModel extends IFilterColumnModel {
  selectOptions: IEnumBasedFilterSelectOption[];
  selectedValues: Array<string | boolean>;
  inputType?: "checkbox" | "radio";
  maxWidth?: string; //e.g "200px"
}

export interface IFreeTextFilterModel extends IFilterColumnModel {
  term: string;
  value: EFilterOperator;
  excludeOperators?: EFilterOperator[];
  autocompleteOptionsFn?: (searchQuery: string, clusterUuid?: string) => Promise<string[]>;
}

export interface IDateFilterModel extends IFilterColumnModel {
  value: EDateFilterOperator;
  date: string;
}

export interface INumericFilterModel extends IFilterColumnModel {
  minValue: number;
  maxValue: number;
  operator: EFilterOperator;
}

export type IAdvancedFilterModel = IEnumBasedFilterModel | IFreeTextFilterModel | IDateFilterModel | INumericFilterModel;

export interface IFreeTextFilterSelectOption extends Omit<ISelectOption, "value" | "label"> {
  label: EFilterOperatorLabel;
  value: EFilterOperator;
}

export const freeTextFilterSelectOptions: IFreeTextFilterSelectOption[] = [
  {
    label: EFilterOperatorLabel.Contains,
    value: EFilterOperator.Contains,
  },
  {
    label: EFilterOperatorLabel.DoesNotContain,
    value: EFilterOperator.DoesNotContain,
  },
  {
    label: EFilterOperatorLabel.Equals,
    value: EFilterOperator.Equals,
  },
  {
    label: EFilterOperatorLabel.NotEquals,
    value: EFilterOperator.NotEquals,
  },
];

export const numericFilterSelectOptions: ISelectOption[] = [
  {
    label: ">=",
    value: EFilterOperator.GreaterThanOrEqual,
  },
  {
    label: "=",
    value: EFilterOperator.Equals,
  },
  {
    label: "!=",
    value: EFilterOperator.NotEquals,
  },
  {
    label: "<=",
    value: EFilterOperator.LessThanOrEqual,
  },
  {
    label: EFilterOperatorLabel.Range,
    value: EFilterOperator.Range,
  },
];

export const filterOperatorMap: Record<EFilterOperator, EFilterOperatorLabel> = {
  [EFilterOperator.Contains]: EFilterOperatorLabel.Contains,
  [EFilterOperator.Equals]: EFilterOperatorLabel.Equals,
  [EFilterOperator.NotEquals]: EFilterOperatorLabel.NotEquals,
  [EFilterOperator.GreaterThanOrEqual]: EFilterOperatorLabel.GreaterThanOrEqual,
  [EFilterOperator.DoesNotContain]: EFilterOperatorLabel.DoesNotContain,
  [EFilterOperator.LessThanOrEqual]: EFilterOperatorLabel.LessThanOrEqual,
  [EFilterOperator.StartsWith]: EFilterOperatorLabel.StartsWith,
  [EFilterOperator.EndsWith]: EFilterOperatorLabel.EndsWith,
  [EFilterOperator.Range]: EFilterOperatorLabel.Range,
};

// ------ Advanced Filter -------
