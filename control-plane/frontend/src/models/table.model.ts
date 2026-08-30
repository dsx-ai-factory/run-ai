import type { EFilterOperator, EColumnFilterType } from "@/models/filter.model";
import type { ISelectOption } from "@/models/global.model";
import type {
  ICustomCellGpuQuotaColConfig,
  ICustomCellQuotaDistributionColConfig,
  ICustomCellEditableWeightColConfig,
  ICustomCellWeightDistributionColConfig,
} from "@/table-models/org-unit.table-model";

export enum ECustomHeader {
  RadioButtonsTableHeader = "radio-buttons-table-header",
}

export interface ICustomHeaderRadioButtonsSelectOption extends ISelectOption {
  shortLabel?: string;
}

export interface ICustomHeaderRadioButtonsConfig {
  options: ICustomHeaderRadioButtonsSelectOption[];
  label: string;
  iconName: string;
  selectedValueFilterByKey: string;
  chipPrefix: string;
  headerStyle?: string;
}

export type ICustomHeaderConfig = ICustomHeaderRadioButtonsConfig;
export enum ECustomCell {
  NAME_COL = "name-col",
  LIST_COL = "list-col",
  LINK_COL = "link-col",
  ROLE_ASSIGMENT_COL = "role-assigment-col",
  APP_STATUS_COL = "applications-status-col",
  AUDIT_STATUS_COL = "audit-status-col",
  ORG_TREE_COL = "org-tree-col",
  SCOPE_COL = "scope-col",
  STRINGS_LIST_COL = "strings-list-col",
  STATUS_COL = "status-col",
  WORKLOAD_STATUS_COL = "workload-status-col",
  HTML_COL = "html-col",
  SOURCE_NAME_COL = "source-name-col",
  CLUSTER_ID_TO_NAME_COL = "cluster-id-to-name-col",
  WORKLOAD_CONNECTION_COL = "workload-connection-col",
  DESCRIPTION_COL = "description-col",
  TOOL_ACCESS_COL = "tool-access-col",
  DEPARTMENT_QUOTA_COL = "department-quota-col",
  PROGRESS_COL = "progress-col",
  ORG_UNIT_GPU_EDITABLE_QUOTA_COL = "org-unit-gpu-editable-quota-col",
  QUOTA_DISTRIBUTION_COL = "org-unit-quota-distribution-col",
  ORG_UNIT_EDITABLE_WEIGHT_COL = "org-unit-editable-weight-col",
  WEIGHT_DISTRIBUTION_COL = "org-unit-weight-distribution-col",
  ICON_BUTTON_COL = "icon-button-col",
  ROLE_WITH_BADGE_COL = "role-with-badge-col",
  APPLICATION_DEPARTMENT_COL = "application-department-col",
  LEGACY_ROLE_COL = "legacy-role-col",
  NEW_ROLE_COL = "new-role-col",
  USER_DISPLAY_NAME_COL = "user-display-name-col",
}

export interface ICustomCellUserDisplayNameColConfig {
  skipWhen?: (row: unknown) => boolean;
}

// Shared defaults for any column that renders a user with display name.
export const USER_DISPLAY_NAME_COL_DEFAULTS = {
  style: "width: 230px; max-width: 230px",
  customCell: ECustomCell.USER_DISPLAY_NAME_COL,
} as const;

export interface ICustomCellLinkColConfig {
  isTextOnly?: boolean; //make link col display text instead of a link button
  tooltipText?: string; //tooltip text to display on hover
  tooltipAnchor?: string; //tooltip anchor position (e.g., "center right", "top middle")
  tooltipOffset?: [number, number]; //tooltip offset [x, y] in pixels
}

/* eslint-disable */
export interface ICustomCellIconButtonColConfig {
  icon: string; // Icon name (e.g., "ph-bold ph-list", "ph-fill ph-terminal")
  color?: string; // Icon color defaults to "b-sub" (black-54)
  size?: string; // Icon size (e.g., "18px", "sm", "md")
  isDisabled?: (row: any) => boolean; // Function to determine if button is disabled
  getTooltip?: (row: any) => string; // Function to get tooltip text based on row data
}

export interface ICustomCellRoleWithBadgeColConfig {
  isTextOnly?: boolean;
  hiddenStages?: string[];
}
/* eslint-enable */

/* eslint-disable @typescript-eslint/no-explicit-any */
export interface ICustomCellNewRoleColConfig {
  roleOptions: string[];
  getReviewRoleId: (roleName: string) => number | undefined;
}
/* eslint-enable @typescript-eslint/no-explicit-any */

export interface ICustomCellConfig {
  [ECustomCell.NEW_ROLE_COL]?: ICustomCellNewRoleColConfig;
  [ECustomCell.LINK_COL]?: ICustomCellLinkColConfig;
  [ECustomCell.ORG_UNIT_GPU_EDITABLE_QUOTA_COL]?: ICustomCellGpuQuotaColConfig;
  [ECustomCell.QUOTA_DISTRIBUTION_COL]?: ICustomCellQuotaDistributionColConfig;
  [ECustomCell.ORG_UNIT_EDITABLE_WEIGHT_COL]?: ICustomCellEditableWeightColConfig;
  [ECustomCell.WEIGHT_DISTRIBUTION_COL]?: ICustomCellWeightDistributionColConfig;
  [ECustomCell.ICON_BUTTON_COL]?: ICustomCellIconButtonColConfig;
  [ECustomCell.ROLE_WITH_BADGE_COL]?: ICustomCellRoleWithBadgeColConfig;
  [ECustomCell.USER_DISPLAY_NAME_COL]?: ICustomCellUserDisplayNameColConfig;
}
/* eslint-disable */
// based on QTableProps["columns"] props
export interface ITableColumn {
  name: string;
  label: string;
  field: string | ((row: any) => any);
  required?: boolean;
  align?: "left" | "right" | "center";
  sortable?: boolean;
  sort?: (a: any, b: any, rowA: any, rowB: any) => number;
  sortOrder?: "ad" | "da";
  format?: (val: any, row: any) => any;
  style?: string | ((row: any) => string);
  classes?: string | ((row: any) => string);
  headerStyle?: string;
  headerClasses?: string;
  display?: boolean;
  mandatory?: boolean;
  hideFilter?: boolean;
  filterKey?: string;
  exportFormat?: (row: any) => any;
  hideExport?: boolean;
  sanitizeCell?: boolean; // Whether to force quotes around this field's value in CSV export
  filter?: {
    type: EColumnFilterType;
    selectOptions?: ISelectOption[];
    excludeOperators?: EFilterOperator[];
    inputType?: "checkbox" | "radio";
    maxWidth?: string; //e.g "200px"
    autocompleteOptionsFn?: (searchQuery: string, clusterUuid?: string) => Promise<string[]>;
  };
  /**** custom cell ****/
  customCell?: ECustomCell;
  customCellEvent?: { emitName: string }; // the event name that will be emitted
  customCellConfig?: ICustomCellConfig;
  /**** cell icon - optional icon to display next to cell content ****/
  cellIcon?: (row: any) => ICellIconConfig | undefined;
  /**** custom header ****/
  customHeader?: ECustomHeader;
  customHeaderEventName?: string; // the event name that will be emitted
  customHeaderConfig?: ICustomHeaderConfig;
}

export interface ICellIconConfig {
  name: string; // Icon name (e.g., "ph-fill ph-star")
  color?: string; // Icon color (e.g., "primary")
  size?: string; // Icon size (e.g., "14px")
  tooltip?: string; // Optional tooltip text
}
/* eslint-enable */

export enum ETableFilters {
  NODE = "node-table-filters",
  NODE_V2 = "node-v2-table-filters",
  NODE_POOL = "node-pool-table-filters",
  NODE_POOL_V2 = "node-pool-v2-table-filters",
  DEPARTMENT = "department-table-filters-advanced",
  COMPUTE_RESOURCE = "compute-resource-table-filters",
  DATA_SOURCE = "data-source-table-filters",
  DATA_VOLUME = "data-volume-table-filters",
  TEMPLATE = "template-table-filters",
  TEMPLATE_V2 = "template-table-filters-v2",
  ENVIRONMENT = "environment-table-filters",
  AUDIT = "audit-table-filters-advanced",
  CREDENTIAL = "credential-table-filters",
  PROJECT = "project-table-filters-advanced",
  ROLE = "role-table-filters",
  APPLICATION = "application-table-filters",
  CLUSTER = "cluster-table-filters",
  USERS = "users-table-filters-v2",
  ACCESS_RULE = "access-rule-table-filters",
  WORKLOAD = "workload-table-filter",
  WORKLOAD_DELETED = "workload-deleted-table-filters",
  POLICY = "policy-table-filters",
  REPORT = "report-table-filters",
  AI_APPLICATION = "ai-application-table-filter",
}

export const MULTI_CLUSTER_SUPPORTED_TABLE_FILTERS: ETableFilters[] = [
  ETableFilters.NODE_POOL_V2,
  ETableFilters.DEPARTMENT,
  ETableFilters.PROJECT,
  ETableFilters.COMPUTE_RESOURCE,
  ETableFilters.DATA_SOURCE,
  ETableFilters.TEMPLATE,
  ETableFilters.ENVIRONMENT,
  ETableFilters.POLICY,
];

export enum ETableExportCsvFilesNames {
  ServiceAccount = "service-accounts",
  Audit = "events-history",
  Cluster = "clusters",
  Credential = "credentials",
  DataSource = "data-sources",
  Department = "departments",
  Environment = "environments",
  Node = "nodes",
  Project = "projects",
  AccessRule = "access-rules",
  Role = "roles",
  Training = "trainings",
  User = "users",
  Template = "templates",
  Workload = "workloads",
  Workspace = "workspaces",
  NodePool = "node-pools",
  AIApplication = "ai-applications",
}

/* eslint-disable */
// based on QTableProps["columns"] props
export interface ITableColumn {
  name: string;
  label: string;
  field: string | ((row: any) => any);
  required?: boolean;
  align?: "left" | "right" | "center";
  sortable?: boolean;
  sort?: (a: any, b: any, rowA: any, rowB: any) => number;
  sortOrder?: "ad" | "da";
  format?: (val: any, row: any) => any;
  style?: string | ((row: any) => string);
  classes?: string | ((row: any) => string);
  headerStyle?: string;
  headerClasses?: string;
  display?: boolean;
  customCell?: ECustomCell;
  customCellEvent?: { emitName: string }; // the event name that will be emitted
  mandatory?: boolean;
  hideFilter?: boolean;
  filterKey?: string;
  exportFormat?: (row: any) => any;
  quoteInCsvExport?: boolean; // Whether to force quotes around this field's value in CSV export
  ignoreFilter?: (row: any) => boolean; // ignoring filter under specific circumstances
}

export type StatusColor = "success" | "warning" | "negative" | "info" | "neutral";

export interface IStatusColOptions {
  status: string;
  tooltipText?: string;
  displayAnimation: boolean;
  color?: StatusColor | string;
  filterKey?: string; //only for client side filtering
  statusUpdatedTimeInMs?: number | string | Date;
  colStyle?: Object; // example usage - {width: '5px'}
  tooltipMaxWidth?: string;
}
/* eslint-enable */

export interface IProgressColOptions {
  value: number | null;
  color?: string;
}

export const HIGHLIGHT_COLUMN_CONFIG = {
  classes: "light-blue-primary-opacity-background",
  headerClasses: "light-blue-primary-opacity-background",
};
