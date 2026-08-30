export enum EEventStatus {
  Succeeded = "Succeeded",
  Failed = "Failed",
}

// Mirrors control-plane/identity-manager/pkg/utils/consts.go
export const AUDIT_USER_SUBJECT_TYPE = "User";
export const AUDIT_APP_SUBJECT_TYPE = "App";
export const AUDIT_SERVICE_ACCOUNT_SUBJECT_TYPE = "ServiceAccount";
export const AUDIT_LEGACY_SERVICE_ACCOUNT_SUBJECT_TYPE = "service-account";
export const AUDIT_SERVICE_ACCOUNT_SUBJECT_TYPE_VALUES = [
  AUDIT_APP_SUBJECT_TYPE,
  AUDIT_SERVICE_ACCOUNT_SUBJECT_TYPE,
  AUDIT_LEGACY_SERVICE_ACCOUNT_SUBJECT_TYPE,
];

export type AuditEventSortBy =
  | "timestamp"
  | "subject"
  | "subject_type"
  | "source_ip"
  | "action"
  | "http_method"
  | "result"
  | "url"
  | "entity_type"
  | "entity_name"
  | "entity_id";
