/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ECustomCell, USER_DISPLAY_NAME_COL_DEFAULTS, type ITableColumn } from "@/models/table.model";
import { dateUtil } from "@/utils/date.util";
import { EColumnFilterType } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import { AuditLogResultEnum, AuditLogSubjectTypeEnum } from "@/swagger-models/audit-service-client";
import {
  AUDIT_APP_SUBJECT_TYPE,
  AUDIT_LEGACY_SERVICE_ACCOUNT_SUBJECT_TYPE,
  AUDIT_SERVICE_ACCOUNT_SUBJECT_TYPE,
  AUDIT_USER_SUBJECT_TYPE,
} from "@/models/audit.model";

export const allAuditColumnsMap: Record<string, ITableColumn> = {
  subject: {
    name: "subject",
    label: "Subject",
    field: (row) => row.subject,
    sortable: true,
    align: "left",
    format: (subject: string) => subject || "-",
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    customCellConfig: {
      [ECustomCell.USER_DISPLAY_NAME_COL]: {
        skipWhen: (row) => (row as { subject_type?: string }).subject_type !== AUDIT_USER_SUBJECT_TYPE,
      },
    },
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.audits("subject", (log) => log.subject),
    },
  },
  subjectType: {
    name: "subject_type",
    label: "Subject Type",
    field: (row) => row.subject_type,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: Object.keys(AuditLogSubjectTypeEnum).map((key) => {
        const label = key === AUDIT_APP_SUBJECT_TYPE ? "Service account" : key;
        return { label, value: key };
      }),
    },
    format: (subject_type: string) => {
      if (!subject_type) return "-";
      if (
        subject_type === AUDIT_SERVICE_ACCOUNT_SUBJECT_TYPE ||
        subject_type === AUDIT_LEGACY_SERVICE_ACCOUNT_SUBJECT_TYPE
      ) {
        return "Service account";
      }
      return subject_type;
    },
  },
  sourceIp: {
    name: "source_ip",
    label: "Source IP",
    field: (row) => row.source_ip,
    format: (source_ip: string) => source_ip || "-",
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.audits("source_ip", (log) => log.source_ip),
    },
  },
  dateAndTime: {
    name: "timestamp",
    label: "Date & time",
    field: (row) => row.timestamp,
    sortable: true,
    align: "left",
    hideFilter: true,
    format: (createdAt: string) => dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
  event: {
    name: "action",
    label: "Event",
    field: (row) => row.action,
    sortable: true,
    align: "left",
    format: (action: string) => action || "-",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.audits("action", (log) => log.action),
    },
  },
  eventId: {
    name: "id",
    label: "Event ID",
    field: (row) => {
      return row.id;
    },
    format: (id: string) => id || "-",
    sortable: true,
    align: "left",
  },
  status: {
    name: "result",
    label: "Status",
    field: (row) => row.result,
    sortable: true,
    align: "left",
    format: (result: string) => result || "-",
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: Object.keys(AuditLogResultEnum).map((key) => ({
        label: key,
        value: key,
      })),
    },
    customCell: ECustomCell.AUDIT_STATUS_COL,
  },
  entityType: {
    name: "entity_type",
    label: "Entity type",
    field: (row) => row.entity_type,
    sortable: true,
    align: "left",
    format: (entity_type: string) => entity_type || "-",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.audits("entity_type", (log) => log.entity_type),
    },
  },
  entityName: {
    name: "entity_name",
    label: "Entity name",
    field: (row) => row.entity_name,
    sortable: true,
    align: "left",
    format: (entity_name: string) => entity_name || "-",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.audits("entity_name", (log) => log.entity_name),
    },
  },
  entityId: {
    name: "entity_id",
    label: "Entity ID",
    field: (row) => row.entity_id,
    sortable: true,
    align: "left",
    hideFilter: true,
    format: (entity_id: string) => entity_id || "-",
  },
  url: {
    name: "url",
    label: "URL",
    field: (row) => row.url,
    sortable: true,
    align: "left",
    format: (url: string) => url || "-",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.audits("url", (log) => log.url),
    },
  },
  httpMethod: {
    name: "http_method",
    label: "HTTP method",
    field: (row) => row.http_method,
    sortable: true,
    align: "left",
    hideFilter: true,
    format: (http_method: string) => http_method || "-",
  },
};

export const auditIndexColumns: Array<ITableColumn> = [
  { ...allAuditColumnsMap.subject, display: true, mandatory: true },
  { ...allAuditColumnsMap.subjectType, display: true },
  { ...allAuditColumnsMap.sourceIp, display: true },
  { ...allAuditColumnsMap.dateAndTime, display: true },
  { ...allAuditColumnsMap.event, display: true },
  { ...allAuditColumnsMap.eventId, display: false },
  { ...allAuditColumnsMap.status, display: true },
  { ...allAuditColumnsMap.entityType, display: true },
  { ...allAuditColumnsMap.entityName, display: true, hideFilter: true },
  { ...allAuditColumnsMap.entityId, display: false },
  { ...allAuditColumnsMap.url, display: false },
  { ...allAuditColumnsMap.httpMethod, display: false },
];
