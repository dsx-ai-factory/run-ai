import type { ITableColumn } from "@/models/table.model";
import type { EWorkloadStatus } from "@/common/status.constant";
import type { AssetCreationRequest, Scope } from "@/swagger-models/assets-service-client";
import { ScopeType } from "@/swagger-models/authorization-client";
import type { SelectType } from "@/components/common/runai-select/runai-select.vue";

/**
 * Maximum string length before truncation is applied in table columns
 */
export const MAX_STRING_LENGTH_BEFORE_TRUNCATION = 25;

export enum EMemoryUnitLable {
  MB = "MB",
  GB = "GB",
  TB = "TB",
  MiB = "MiB",
  GiB = "GiB",
  TiB = "TiB",
}

export enum EMemoryUnitValue {
  MB = "M",
  GB = "G",
  TB = "T",
  GiB = "Gi",
  MiB = "Mi",
  TiB = "Ti",
}

export enum ESidebarState {
  Mini = "mini",
  Full = "full",
  Hidden = "hidden",
}

export type ObjectType = Record<string, unknown>;
// eslint-disable-next-line
export type CallbackFunction = (data?: any) => void;

export type SelectUpdateFilterFunction = (func: { (): void }) => void;

export const basicMemoryUnitOptions: ISelectOption[] = [
  { label: EMemoryUnitLable.MB, value: EMemoryUnitValue.MB },
  { label: EMemoryUnitLable.GB, value: EMemoryUnitValue.GB },
];

export const allMemoryUnitOptions: ISelectOption[] = [
  { label: EMemoryUnitLable.MB, value: EMemoryUnitValue.MB },
  { label: EMemoryUnitLable.MiB, value: EMemoryUnitValue.MiB },
  { label: EMemoryUnitLable.GB, value: EMemoryUnitValue.GB },
  { label: EMemoryUnitLable.GiB, value: EMemoryUnitValue.GiB },
  { label: EMemoryUnitLable.TB, value: EMemoryUnitValue.TB },
  { label: EMemoryUnitLable.TiB, value: EMemoryUnitValue.TiB },
];

export interface ISelectOption {
  label: string;
  value: string | boolean | number | null | undefined;
  icon?: string;
  disable?: boolean;
  toolTip?: string; // TODO: we should rename this to info (this shows the question mark with tooltip)
  toolTipWidth?: string;
  disabledTooltip?: string; // when the line is disabled, you can add a tooltip without a question mark icon
  disabledTooltipWidth?: string;
  svg?: string;
  aid?: string;
  warningTooltip?: string;
  isHtmlTooltipText?: boolean;
  type?: SelectType;
  subLabel?: string;
  subLabelColor?: string;
  roleStageBadge?: string; // Role stage badge (e.g., "New", "Legacy") for role options
}

export interface IStorageTableColumns {
  columns: Array<ITableColumn>;
}

export interface ILocalStatus {
  oldStatus: string | undefined;
  loadingStatus: EWorkloadStatus;
  creationTimestamp: number;
}

export interface IDrawerSettings {
  isOpen: boolean;
  widthPercentage: number;
}

export interface IWebSocketSubscription {
  onMessage: (messageEvent: MessageEvent) => void;
  onError: (errorEvent: Event) => void;
  onOpen: (openEvent: Event) => void;
  onClose: (closeEvent: CloseEvent) => void;
  unsubscribe: () => void;
  websocket?: WebSocket;
}

export interface IUIAssetCreationMeta extends Omit<AssetCreationRequest, "scope"> {
  scope: Scope | null;
}

export interface IScopeModel {
  scope: Scope | null;
  projectId?: number | null;
  departmentId?: string | null;
  clusterId?: string | null;
  parentClusterUuid?: string | null;
}

export const ScopeTypeSystemExcluded = {
  Tenant: ScopeType.Tenant,
  Department: ScopeType.Department,
  Cluster: ScopeType.Cluster,
  Project: ScopeType.Project,
} as const;

export interface HttpError extends Error {
  response?: {
    status: number; // HTTP response status code
    data?: object; // Optional: data returned from the response
  };
}
export type ScopeTypeSystemExcluded = (typeof ScopeTypeSystemExcluded)[keyof typeof ScopeTypeSystemExcluded];

export type IValidationRule = (val: string) => boolean | string | Promise<boolean | string>;

export type TTreeDisabledScopes = {
  [key in ScopeType]?: boolean;
};

export type TTreeSupportedVersions = {
  [key in ScopeType]?: string;
};

export type TQuasarTooltipAnchor =
  | "top left"
  | "top middle"
  | "top right"
  | "top start"
  | "top end"
  | "center left"
  | "center middle"
  | "center right"
  | "center start"
  | "center end"
  | "bottom left"
  | "bottom middle"
  | "bottom right"
  | "bottom start"
  | "bottom end"
  | undefined;
