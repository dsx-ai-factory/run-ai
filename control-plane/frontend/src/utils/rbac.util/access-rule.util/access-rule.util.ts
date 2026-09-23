/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  accessRuleNoStatusMessage,
  accessRulePhaseLabel,
  AccessRuleStatusMap,
  type IAccessRuleDisplayConfig,
  type IAccessRuleManagementModalOptions,
  ResourceTypeGroupIdOrder,
} from "@/models/access-rule.model";
import { EAccessRuleModalPage } from "@/models/access-rule.model";
import type { IAccessRuleRecords } from "@/models/access-rule.model";
import type { IDepartmentTable } from "@/models/department.model";
import type { IProjectTable } from "@/models/project.model";
import { deepCopy } from "@/utils/common.util";
import {
  type AccessRule,
  AccessRulePhase,
  type Permission,
  type ResourceTypeGroupId,
  ScopeType,
} from "@/swagger-models/authorization-client";
import type { TRolesTypes } from "@/models/roles.model";
import type { IUser } from "@/models/user.model";
import type { ServiceAccountWithRoles } from "@/models/service-accounts.model";
import type { IStatusColOptions } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import type { IAccessRuleTable } from "@/table-models/access-rule.table-model";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import commonUtil from "@/utils/common.util/common.util";
import { docsUtil } from "@/utils/docs.util/docs.util";
import type { QNotifyCreateOptions } from "quasar";
import { alertUtil } from "@/utils/alert.util";
import { ErrorAlert } from "@/utils/error-alert.util";

export const accessRuleUtil = {
  getAccessRuleModalDisplayConfig,
  getAccessRuleModalHeader,
  getAccessRuleModalSubHeader,
  isScopePage,
  getUserSubjectIds,
  sortAndGroupPermissionsByGroupId,
  getStatusColOptions,
  addStatusExplanationToAccessRules,
  getCreatedAccessRulesNotification,
};

const CLOUD_OPERATOR_ROLE_NAME = "Cloud operator";

function getAccessRuleModalDisplayConfig(page: EAccessRuleModalPage): IAccessRuleDisplayConfig {
  switch (page) {
    case EAccessRuleModalPage.Project:
      return { subject: true, users: true, roles: true, scope: false };
    case EAccessRuleModalPage.Department:
      return { subject: true, users: true, roles: true, scope: false };
    case EAccessRuleModalPage.Application:
      return { subject: false, users: false, roles: true, scope: true };
    case EAccessRuleModalPage.User:
      return { subject: false, users: false, roles: true, scope: true };
    default:
      return { subject: false, users: false, roles: false, scope: false };
  }
}

function getAccessRuleModalHeader(options: IAccessRuleManagementModalOptions): string {
  switch (options.page) {
    case EAccessRuleModalPage.Project:
      return `Access Rules for ${options.scopeName}`;
    case EAccessRuleModalPage.Department:
      return `Access Rules for ${options.scopeName}`;
    case EAccessRuleModalPage.Application:
      return `Access Rules for ${options.subjectId}`;
    case EAccessRuleModalPage.User:
      return `Access Rules for ${options.subjectId}`;
    default:
      return "";
  }
}
function getAccessRuleModalSubHeader(page: EAccessRuleModalPage): string {
  switch (page) {
    case EAccessRuleModalPage.Project:
      return "Add rules to authorize access to this project";
    case EAccessRuleModalPage.Department:
      return "Add rules to authorize access to this department";
    case EAccessRuleModalPage.Application:
      return "Add rules and select roles and scopes to authorize this service account's access ";
    case EAccessRuleModalPage.User:
      return "Add rules and select roles and scopes to authorize this user's access ";
    default:
      return "";
  }
}

function isScopePage(page: EAccessRuleModalPage): boolean {
  return [EAccessRuleModalPage.Project, EAccessRuleModalPage.Department].some((p: EAccessRuleModalPage) => p === page);
}

export function enrichServiceAccountsWithRoles(
  serviceAccounts: ServiceAccountWithRoles[],
  serviceAccountAccessRules: IAccessRuleRecords,
): ServiceAccountWithRoles[] {
  const serviceAccountNameToRolesMap = _getSubjectIdToRolesMap(serviceAccountAccessRules);
  if (serviceAccountNameToRolesMap.size > 0) {
    _addRolesToServiceAccounts(serviceAccounts, serviceAccountNameToRolesMap);
  }
  return serviceAccounts;
}

export function enrichScopeEntityWithAccessRules<T extends IDepartmentTable | IProjectTable>(
  items: T[],
  accessRules: AccessRule[],
): IDepartmentTable[] | IProjectTable[] {
  const itemsWithAccessRules = deepCopy(items) as IDepartmentTable[] | IProjectTable[];
  const accessRulesByScope = accessRules.reduce(
    (assignmentsMap: Record<string, string[]>, item: AccessRule) => {
      assignmentsMap[item.scopeId] ??= [];
      assignmentsMap[item.scopeId].push(item.subjectType);
      return assignmentsMap;
    },
    {} as Record<string, string[]>,
  );
  for (const item of itemsWithAccessRules) {
    item.rolesNames = accessRulesByScope[String(item.id)] || [];
  }
  return itemsWithAccessRules;
}

function _getSubjectIdToRolesMap(subjectAccessRules: IAccessRuleRecords): Map<string, Array<string>> {
  const subjectIdToRolesMap = new Map<string, Array<string>>();
  subjectAccessRules.accessRules.forEach((accessRule) => {
    const subjectId = accessRule.subjectId;
    const roleName = accessRule.roleName;
    if (subjectIdToRolesMap.has(subjectId)) {
      const userRoles = subjectIdToRolesMap.get(subjectId);
      if (userRoles != undefined && !userRoles?.includes(roleName)) {
        userRoles?.push(roleName);
        subjectIdToRolesMap.set(subjectId, userRoles);
      }
    } else {
      subjectIdToRolesMap.set(subjectId, [roleName]);
    }
  });
  return subjectIdToRolesMap;
}

function _addRolesToServiceAccounts(
  serviceAccounts: ServiceAccountWithRoles[],
  serviceAccountNameToRolesMap: Map<string, Array<string>>,
): void {
  serviceAccounts.forEach((serviceAccount) => {
    if (serviceAccount.name && serviceAccountNameToRolesMap.has(serviceAccount.name)) {
      serviceAccount.roles = [];
      serviceAccount.roles.push(...(serviceAccountNameToRolesMap.get(serviceAccount.name) as Array<TRolesTypes>));
    }
  });
}

function getUserSubjectIds(user: IUser): Array<string> {
  return user.groups ? [user.username].concat(user.groups || []) : [user.username];
}

/**
 * Sorts and groups permissions by their group ID.
 *
 * @param {Permission[]} permissions - An array of permissions to be sorted and grouped.
 *
 * @returns {Permission[]} - The sorted and grouped permissions.
 *
 * The function works in the following steps:
 * 1. It first sorts the permissions based on their display name.
 * 2. It then groups the sorted permissions by their group ID.
 * 3. Finally, it orders the groups according to the ResourceTypeGroupIdOrder and returns the result.
 */
function sortAndGroupPermissionsByGroupId(permissions: Permission[]): Permission[] {
  const sortedPermissions = [...permissions].sort((a, b) => a.displayName.localeCompare(b.displayName));

  const groupedPermissions: Record<string, Permission[]> = sortedPermissions.reduce(
    (groups, permission) => {
      groups[permission.groupId] = [...(groups[permission.groupId] || []), permission];
      return groups;
    },
    {} as Record<string, Permission[]>,
  );

  return ResourceTypeGroupIdOrder.flatMap((key: ResourceTypeGroupId) => groupedPermissions[key] || []);
}

function getStatusColOptions(accessRule: IAccessRuleTable): IStatusColOptions {
  if (accessRule.scopeType === ScopeType.System) {
    // System scope access rules will not have a status and should always display “No issues found”
    return AccessRuleStatusMap[AccessRulePhase.Ready];
  }
  if (accessRule.roleName === CLOUD_OPERATOR_ROLE_NAME) {
    // Cloud operator role will not have a status and should always display “No issues found”
    return AccessRuleStatusMap[AccessRulePhase.Ready];
  }

  const defaultStatus: IStatusColOptions = {
    status: accessRulePhaseLabel.NoStatus,
    displayAnimation: false,
    tooltipText: accessRuleNoStatusMessage,
  };
  const status = accessRule.status;
  if (!status || !status.phase) return defaultStatus;

  const statusOptions: IStatusColOptions = AccessRuleStatusMap[status.phase];
  const tooltipText = status.explanation ? status.explanation : status.message;
  return tableUtil.getStatusColOptions(statusOptions, tooltipText);
}

const getAccessRuleExplanation = (clusterNames: string[]): string => {
  const linkHTML = docsUtil.getSupportHTMLLink();
  if (clusterNames.length === 1) {
    return `Issues were found while enforcing the access rule <br> in the cluster.${linkHTML ? ` ${linkHTML}` : ""}`;
  }
  if (clusterNames.length > 1) {
    return `Issues were found while enforcing the access rule in cluster(s) <br> ${commonUtil.previewStrings(
      clusterNames,
      3,
    )}.${linkHTML ? ` <br> ${linkHTML}` : ""}`;
  }
  return "";
};

const _extractClusterIds = (input: string): string[] => input.match(/\b[0-9a-fA-F-]+\b/g) || [];
const _getClusterNames = (clusters: ClustersMinimalData[], clusterIds: string[]): string[] =>
  clusters.filter((cluster) => clusterIds.includes(cluster.uuid)).map((cluster) => cluster.name);

function addStatusExplanationToAccessRules(
  accessRules: IAccessRuleTable[],
  clusters: ClustersMinimalData[],
): IAccessRuleTable[] {
  return accessRules.map((accessRule) => {
    const status = accessRule.status || {};
    if (status.phase === AccessRulePhase.NotReady && status.details) {
      const clusterIds = _extractClusterIds(status.details);
      const clusterNames = _getClusterNames(clusters, clusterIds);

      return {
        ...accessRule,
        status: {
          ...status,
          explanation: getAccessRuleExplanation(clusterNames),
        },
      };
    }
    return accessRule;
  });
}

/**
 * Builds the notification for a batch access-rule creation result. Surfaces the failed count
 * instead of a fractional success count (e.g. "0/1 created") when any rule was rejected, since
 * the API can return HTTP 200 with a non-empty `failed` array.
 *
 * @param succeededCount - number of rules created successfully
 * @param failedCount - number of rules that failed to create
 * @param singleSuccessMessage - message to show when exactly one rule is created (callers vary in
 *   whether they name the subject, e.g. "Access rule for alice@run.ai created" vs "Access rule created")
 */
function getCreatedAccessRulesNotification(
  succeededCount: number,
  failedCount: number,
  singleSuccessMessage: string,
): QNotifyCreateOptions {
  if (failedCount > 0) {
    const message =
      failedCount === 1
        ? ErrorAlert.failedCreateMessage("access rule")
        : `${failedCount} access rule(s) couldn't be created`;
    return alertUtil.getError(message);
  }

  const message = succeededCount === 1 ? singleSuccessMessage : `${succeededCount} access rule(s) created`;
  return alertUtil.getSuccess(message);
}
