/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IStatusColOptions, ITableColumn } from "@/models/table.model";
import type { ISelectOption } from "@/models/global.model";
import type { IOrgTreeNodeId } from "@/models/org-tree.model";
import type { ScopeType } from "@/swagger-models/authorization-client";
import {
  SubjectTypeExtended,
  type AccessRule,
  ResourceTypeGroupId,
  AccessRulePhase,
} from "@/swagger-models/authorization-client";
import type { IUser } from "@/models/user.model";
import { ResourceType } from "@/swagger-models/authorization-client";

export enum ERoleName {
  SystemAdmin = "System administrator",
}

export enum EAccessRuleModalPage {
  Project = "project",
  Department = "department",
  User = "user",
  Application = "application",
}

export const ALL_AUTHENTICATED_USERS_LABEL = "All authenticated users and service accounts";

const allAuthenticatedUsersOption: ISubjectTypeOption = {
  label: ALL_AUTHENTICATED_USERS_LABEL,
  value: SubjectTypeExtended.AllAuthenticatedUsers,
};

export const subjectTypeOptions: ISubjectTypeOption[] = [
  {
    label: "User",
    value: SubjectTypeExtended.User,
  },
  {
    label: "Service account",
    value: SubjectTypeExtended.ServiceAccount,
  },
  allAuthenticatedUsersOption,
];

export const subjectTypeOptionsWithSSO: ISubjectTypeOption[] = [
  {
    label: "User",
    value: SubjectTypeExtended.User,
  },
  {
    label: "SSO Group",
    value: SubjectTypeExtended.Group,
  },
  {
    label: "Service account",
    value: SubjectTypeExtended.ServiceAccount,
  },
  allAuthenticatedUsersOption,
];

export const excludedResourcesTypes: Set<ResourceType> = new Set([ResourceType.ClusterConfig, ResourceType.AccessKeys]);

export interface IAccessRuleRecords {
  accessRules: AccessRule[];
  displayRecords: number;
  totalRecords: number;
}
export interface IAccessRuleTableModalOptions {
  accessRules?: AccessRule[];
  scopeType?: ScopeType;
  scopeId?: string;
  subjectType?: SubjectTypeExtended;
  subjectId?: string;
  columns: Array<ITableColumn>;
  header: string;
  loading?: boolean;
}

export interface IAccessRuleManagementModalOptions {
  subjectId?: string;
  subjectType?: SubjectTypeExtended;
  scopeType?: ScopeType;
  scopeName?: string;
  scopeId?: string;
  page: EAccessRuleModalPage;
  user?: IUser;
  userGroups?: string[];
}

export interface IAccessRuleManagementFormOptions extends IAccessRuleManagementModalOptions {}

export interface IAccessRuleDisplayConfig {
  subject: boolean;
  users: boolean;
  roles: boolean;
  scope: boolean;
}

export interface ISubjectTypeOption extends ISelectOption {
  value: SubjectTypeExtended;
}

export interface IRoleIdOption extends ISelectOption {
  value: number;
}
export interface IAccessRuleItem {
  subjectTypeOption: null | ISubjectTypeOption;
  roleOption: null | IRoleIdOption;
  scopeInput: null | IOrgTreeNodeId;
  scopeName?: string;
  scopeType?: ScopeType;
  subjectType?: SubjectTypeExtended;
  roleName?: string;
  subjectId: string;
  accessRuleId?: number;
  saved: boolean;
  isInherited?: boolean;
}

export interface AccessRuleWithGroup extends AccessRule {
  group: string;
}

const subjectTypeFilterLabels = new Map<SubjectTypeExtended, string>([
  [SubjectTypeExtended.Group, "SSO Group"],
  [SubjectTypeExtended.ServiceAccount, "Service account"],
  [SubjectTypeExtended.AllAuthenticatedUsers, ALL_AUTHENTICATED_USERS_LABEL],
]);

export const accessRuleSubjectTypeFilterOptions: ISelectOption[] = Object.entries(SubjectTypeExtended)
  .filter(([, value]) => value !== SubjectTypeExtended.App) // Hide deprecated App from filter
  .map(([key, value]) => ({
    label: subjectTypeFilterLabels.get(value) ?? key,
    value: value,
  }));

/**
 * Order of resource types.
 * This array defines the order in which resource types should be displayed in the role permissions table.
 *
 * @type {ResourceTypeGroupId[]}
 */
export const ResourceTypeGroupIdOrder = [
  ResourceTypeGroupId.Organization,
  ResourceTypeGroupId.PhysicalResource,
  ResourceTypeGroupId.Iam,
  ResourceTypeGroupId.Workload,
  ResourceTypeGroupId.WorkloadAsset,
];

export const accessRuleNoStatusMessage =
  "Status cannot be displayed because the current version of the cluster is not up to date or unknown";

export const accessRulePhaseLabel = {
  [AccessRulePhase.Creating]: "Creating",
  [AccessRulePhase.Ready]: "No issues found",
  [AccessRulePhase.Deleting]: "Deleting",
  [AccessRulePhase.NotReady]: "Issues found",
  NoStatus: "No status",
};

export const AccessRuleStatusMap: Record<AccessRulePhase, IStatusColOptions> = {
  [AccessRulePhase.Creating]: {
    status: `${accessRulePhaseLabel.Creating}...`,
    displayAnimation: true,
    tooltipText: "The access rule is being created in the cluster",
  },
  [AccessRulePhase.Ready]: {
    status: accessRulePhaseLabel.Ready,
    color: "success",
    displayAnimation: false,
  },
  [AccessRulePhase.Deleting]: {
    status: `${accessRulePhaseLabel.Deleting}...`,
    displayAnimation: true,
  },
  [AccessRulePhase.NotReady]: {
    status: accessRulePhaseLabel.NotReady,
    color: "warning",
    displayAnimation: false,
  },
};

export const accessRuleStatusFilterOptions: ISelectOption[] = [
  {
    label: accessRulePhaseLabel.Ready,
    value: AccessRulePhase.Ready,
  },
  {
    label: accessRulePhaseLabel.Creating,
    value: AccessRulePhase.Creating,
  },
  {
    label: accessRulePhaseLabel.Deleting,
    value: AccessRulePhase.Deleting,
  },
  {
    label: accessRulePhaseLabel.NotReady,
    value: AccessRulePhase.NotReady,
  },
];

/** Bulk actions **/
export const SCOPE_MAX_AMOUNT = 10;
