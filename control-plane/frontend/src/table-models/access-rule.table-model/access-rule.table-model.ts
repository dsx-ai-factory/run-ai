/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  subjectTypeOptionsWithSSO,
  type AccessRuleWithGroup,
  accessRuleSubjectTypeFilterOptions,
  accessRuleStatusFilterOptions,
} from "@/models/access-rule.model";
import type { ISelectOption } from "@/models/global.model";
import { ECustomCell, USER_DISPLAY_NAME_COL_DEFAULTS, type ITableColumn } from "@/models/table.model";
import { ERoleStage } from "@/models/roles.model";
import type { AccessRuleStatusResponse, MeAccessRule } from "@/swagger-models/authorization-client";
import {
  AccessRulesSortFilterFields,
  SubjectTypeExtended,
  type AccessRule,
} from "@/swagger-models/authorization-client";
import { dateUtil } from "@/utils/date.util";
import { EColumnFilterType } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util/autocomplete.util";
import { accessRuleUtil } from "@/utils/rbac.util/access-rule.util";

const subjectTypeOptionsMap: Map<string, string> = new Map(
  subjectTypeOptionsWithSSO.map((option: ISelectOption) => [option.value as string, option.label as string]),
);
// Map deprecated "app" to "Service account" for backward compatibility
subjectTypeOptionsMap.set("app", "Service account");

export interface IAccessRuleStatusResponseTable extends AccessRuleStatusResponse {
  explanation?: string;
}

export interface IAccessRuleTable extends AccessRule {
  status?: IAccessRuleStatusResponseTable;
}

export const allAccessRulesColumnsMap: Record<string, ITableColumn> = {
  type: {
    name: AccessRulesSortFilterFields.SubjectType,
    label: "Type",
    field: (row: IAccessRuleTable) => row.subjectType,
    sortable: true,
    align: "left",
    display: true,
    format: (subjectType: SubjectTypeExtended) => subjectTypeOptionsMap.get(subjectType) || "-",
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: accessRuleSubjectTypeFilterOptions,
    },
  },
  group: {
    name: "group",
    label: "Group",
    field: (row: AccessRuleWithGroup) => row.group,
    sortable: false,
    align: "left",
    display: false,
    format: (group: string) => group || "-",
    hideFilter: true,
  },
  subject: {
    name: AccessRulesSortFilterFields.SubjectId,
    label: "Subject",
    field: (row: AccessRule) => (row.subjectType === SubjectTypeExtended.AllAuthenticatedUsers ? "-" : row.subjectId),
    sortable: true,
    align: "left",
    display: true,
    // 230px from Figma.
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    customCellConfig: {
      [ECustomCell.USER_DISPLAY_NAME_COL]: {
        skipWhen: (row) => (row as AccessRule).subjectType !== SubjectTypeExtended.User,
      },
    },
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.accessRules(AccessRulesSortFilterFields.SubjectId, (r) => r.subjectId),
    },
  },
  role: {
    name: AccessRulesSortFilterFields.RoleName,
    label: "Role",
    field: (row: IAccessRuleTable) => row.roleName,
    sortable: true,
    align: "left",
    display: true,
    customCell: ECustomCell.ROLE_WITH_BADGE_COL,
    customCellConfig: {
      [ECustomCell.ROLE_WITH_BADGE_COL]: { isTextOnly: true, hiddenStages: [ERoleStage.New] },
    },
    customCellEvent: { emitName: "role-clicked" },
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.accessRules(AccessRulesSortFilterFields.RoleName, (r) => r.roleName),
    },
  },
  scope: {
    name: AccessRulesSortFilterFields.ScopeName,
    label: "Scope",
    field: (row: IAccessRuleTable) => row.scopeName,
    sortable: true,
    align: "left",
    display: true,
    customCell: ECustomCell.ORG_TREE_COL,
    customCellEvent: { emitName: "scope-clicked" },
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.accessRules(AccessRulesSortFilterFields.ScopeName, (r) => r.scopeName),
    },
  },
  status: {
    name: AccessRulesSortFilterFields.Phase,
    label: "Status",
    field: (row: IAccessRuleTable) => row,
    sortable: true,
    align: "left",
    format: (accessRule: IAccessRuleTable) => accessRuleUtil.getStatusColOptions(accessRule),
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: accessRuleStatusFilterOptions,
    },
    customCell: ECustomCell.STATUS_COL,
  },
  createdBy: {
    name: AccessRulesSortFilterFields.CreatedBy,
    label: "Authorized by",
    field: (row: IAccessRuleTable) => row.createdBy,
    sortable: true,
    align: "left",
    display: true,
    format: (createdBy: string) => createdBy || "-",
    // 230px from Figma.
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.accessRules(AccessRulesSortFilterFields.CreatedBy, (r) => r.createdBy),
    },
  },
  creationTime: {
    name: AccessRulesSortFilterFields.CreatedAt,
    label: "Creation time",
    field: (row: IAccessRuleTable) => row.createdAt,
    sortable: true,
    align: "left",
    display: true,
    format: (createdAt: string) => createdAt && dateUtil.dateAndTimeFormat(new Date(createdAt)),
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  lastUpdated: {
    name: "lastUpdated",
    label: "Last updated",
    field: (row: IAccessRuleTable) => row.updatedAt,
    hideFilter: true,
    sortable: false,
    align: "left",
    display: true,
    format: (createdAt: string) => createdAt && dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
};

export const accessRulesIndexColumns: Array<ITableColumn> = [
  { ...allAccessRulesColumnsMap.type, display: true, mandatory: true },
  { ...allAccessRulesColumnsMap.subject, display: true },
  { ...allAccessRulesColumnsMap.role, display: true },
  { ...allAccessRulesColumnsMap.scope, display: true },
  { ...allAccessRulesColumnsMap.status, display: true },
  { ...allAccessRulesColumnsMap.createdBy, display: true },
  { ...allAccessRulesColumnsMap.creationTime, display: true },
  { ...allAccessRulesColumnsMap.lastUpdated, display: true },
];

export const accessRulesMiniTableColumns: Array<ITableColumn> = [
  { ...allAccessRulesColumnsMap.subject, display: true },
  { ...allAccessRulesColumnsMap.type, display: true },
  { ...allAccessRulesColumnsMap.scope, display: true },
  { ...allAccessRulesColumnsMap.role, display: true },
  { ...allAccessRulesColumnsMap.createdBy, display: true },
  { ...allAccessRulesColumnsMap.lastUpdated, display: true },
];

export const roleAccessRulesMiniTableColumns: Array<ITableColumn> = [
  { ...allAccessRulesColumnsMap.subject, display: true },
  { ...allAccessRulesColumnsMap.type, display: true },
  { ...allAccessRulesColumnsMap.scope, display: true },
  { ...allAccessRulesColumnsMap.createdBy, display: true },
  { ...allAccessRulesColumnsMap.lastUpdated, display: true },
];

export const subjectRolesMiniTableColumns: Array<ITableColumn> = [
  { ...allAccessRulesColumnsMap.role, display: true },
  { ...allAccessRulesColumnsMap.scope, display: true },
  { ...allAccessRulesColumnsMap.createdBy, display: true },
  { ...allAccessRulesColumnsMap.creationTime, display: true },
  { ...allAccessRulesColumnsMap.lastUpdated, display: true },
];

export const subjectRolesMiniTableColumnsWithSSO: Array<ITableColumn> = [
  { ...allAccessRulesColumnsMap.role, display: true },
  { ...allAccessRulesColumnsMap.scope, display: true },
  { ...allAccessRulesColumnsMap.type, display: true },
  { ...allAccessRulesColumnsMap.group, display: true },
  { ...allAccessRulesColumnsMap.createdBy, display: true },
  { ...allAccessRulesColumnsMap.creationTime, display: true },
  { ...allAccessRulesColumnsMap.lastUpdated, display: true },
];

/*** Me Access Rules ***/
export const allMeAccessRulesColumnsMap: ITableColumn[] = [
  {
    name: "roleName",
    label: "Role",
    align: "left",
    sortable: false,
    field: (row: MeAccessRule) => row.roleName,
  },
  {
    name: AccessRulesSortFilterFields.ScopeName,
    label: "Scope",
    align: "left",
    field: (row: MeAccessRule) => row.scopeName,
    customCell: ECustomCell.ORG_TREE_COL,
    sortable: false,
    customCellEvent: { emitName: "scope-clicked" },
  },
];
