import type { IFilterBy, IPaginationFilter } from "@/models/filter.model";
import { controlPlaneService } from "@/services/control-plane/control-plane.service/control-plane.service";
import type { IAccessRuleRecords } from "@/models/access-rule.model";
import type { IUser } from "@/models/user.model";
import { API } from "@/common/api.constant";
import type {
  AccessRulesBatchCreateFields,
  AuthenticatedUsersAccessRuleCreationFields,
  AccessRulesBatchFields,
  AccessRulesMigrateDeprecatedResponse,
  MeAccessRule,
} from "@/swagger-models/authorization-client";
import {
  AccessRulesSortFilterFields,
  type AccessRule,
  type AccessRuleCreationFields,
  type ScopeType,
  SubjectType,
} from "@/swagger-models/authorization-client";
import { filterService } from "@/services/filter.service/filter.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { authorizationApi } from "@/services/infra/client-apis/authorization-api/authorization-api";
import { httpService } from "@/services/infra/https.service/http.service";
import type { AxiosResponse } from "axios";
import type { ServiceAccount } from "@/swagger-models/identity-manager-client";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";
import { API_LIMIT_MAX_RECORDS } from "@/composables/helpers/use-export-file.mapper";
export interface IAccessRulesFilter {
  subjectType?: SubjectType;
  subjectIdFilter?: string;
  subjectIds?: Array<string>;
  limit?: number;
  offset?: number;
  lastUpdated?: string;
  includeDeleted?: boolean;
  clusterId?: string;
  scopeType?: ScopeType;
  scopeId?: string;
  roleId?: number;
}

export const accessRuleService = {
  deleteAccessRule,
  createAccessRule,
  getSubjectIds,
  getAccessRules,
  getAccessRulesCount,
  list,
  getAccessRule,
  getAccessRulesV2,
  getMeAccessRules,
  countMeAccessRules,
  accessRulesBatch,
  accessRulesBatchCreate,
  createAuthenticatedUsersAccessRule,
  getAllAccessRules,
  getSsoGroupSubjectIds,
  migrateDeprecatedAccessRules,
};

async function createAccessRule(accessRuleCreationFields: AccessRuleCreationFields): Promise<AccessRule> {
  return await controlPlaneService.post(`${API.accessRules}`, accessRuleCreationFields);
}

async function deleteAccessRule(accessRuleId: number): Promise<void> {
  await controlPlaneService.delete(`${API.accessRules}/${accessRuleId}`);
}

async function getAccessRules(filter: IAccessRulesFilter): Promise<IAccessRuleRecords> {
  return await controlPlaneService.get(`${API.accessRules}`, {
    ...filter,
    subjectIds: filter.subjectIds?.join(","),
  } as IAccessRulesFilter);
}

async function getSubjectIds(isSSO: boolean, subjectType: SubjectType, searchQuery: string): Promise<Array<string>> {
  let subjectIds: Array<string> = new Array<string>();
  if (subjectType === SubjectType.User) {
    const subjects: Array<IUser> = await getUsers();
    subjectIds = subjects.map((user: IUser) => user.username);
  } else if (subjectType === SubjectType.ServiceAccount) {
    const subjects: Array<ServiceAccount> = await getServiceAccounts();
    subjectIds = subjects.map((serviceAccount: ServiceAccount) => serviceAccount.name);
  }
  subjectIds = subjectIds.filter((name) => {
    return name.toLowerCase().includes(searchQuery.toLowerCase());
  });
  if (isSSO) {
    const accessRules: IAccessRuleRecords = await controlPlaneService.get(
      `${API.accessRules}?subjectType=${subjectType}&subjectIdFilter=${searchQuery}`,
    );
    const subjectIdsSet = new Set<string>();
    accessRules.accessRules.forEach((accessRule: AccessRule) => {
      subjectIdsSet.add(accessRule.subjectId);
    });
    subjectIds.forEach((subjectId: string) => {
      subjectIdsSet.add(subjectId);
    });

    // Convert Set to array
    subjectIds = Array.from(subjectIdsSet);
  }
  return subjectIds;
}

async function getUsers(): Promise<Array<IUser>> {
  return controlPlaneService.get(`${API.v1}/users`);
}

async function getServiceAccounts(): Promise<Array<ServiceAccount>> {
  return serviceAccountsService.list();
}

//pagination
async function list(filterBy: IFilterBy = {}): Promise<AccessRule[]> {
  const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy, undefined, true);
  const response = await authorizationApi.accessRulesApi.getAccessRules(
    undefined,
    undefined,
    filters.limit,
    filters.offset,
    undefined,
    undefined,
    undefined,
    undefined,
    filters.sortOrder,
    filters.sortBy as AccessRulesSortFilterFields | undefined,
    filters.filterBy,
    filters.search,
  );

  return response.data.accessRules;
}

async function getAccessRulesV2(
  filterBy?: Array<string>,
  sortBy?: string,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
  search?: string,
): Promise<AccessRule[]> {
  const response = await authorizationApi.accessRulesApi.getAccessRules(
    undefined,
    undefined,
    limit,
    offset,
    undefined,
    undefined,
    undefined,
    undefined,
    sortOrder,
    sortBy as AccessRulesSortFilterFields | undefined,
    filterBy,
    search,
  );
  return response.data.accessRules;
}

async function getAccessRulesCount(filterBy?: Array<string>, search?: string): Promise<number> {
  const response = await authorizationApi.accessRulesApi.countAccessRules(undefined, filterBy, search);
  return response.data.count;
}

async function getAccessRule(id: number): Promise<AccessRule> {
  const response = await authorizationApi.accessRulesApi.getAccessRule(id);
  return response.data;
}

async function getMeAccessRules(limit?: number, offset?: number, search?: string): Promise<MeAccessRule[]> {
  const response = await authorizationApi.meApi.getMeAccessRules(limit, offset, search);
  return response.data.accessRules?.accessRules || [];
}

async function countMeAccessRules(search?: string): Promise<number> {
  const response = await authorizationApi.meApi.countMeAccessRules(search);
  return response.data.count;
}

async function accessRulesBatch(accessRulesBatchFields?: AccessRulesBatchFields) {
  const response = await authorizationApi.accessRulesApi.accessRulesBatch(accessRulesBatchFields);
  return response.data;
}

async function accessRulesBatchCreate(accessRulesBatchCreateFields?: AccessRulesBatchCreateFields) {
  const response = await authorizationApi.accessRulesApi.accessRulesBatchCreate(accessRulesBatchCreateFields);
  return response.data;
}

async function createAuthenticatedUsersAccessRule(
  fields: AuthenticatedUsersAccessRuleCreationFields,
): Promise<AccessRule> {
  try {
    const response: AxiosResponse = await authorizationApi.accessRulesApi.createAuthenticatedUsersAccessRule(fields);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getAllAccessRules(
  filterBy?: Array<string>,
  sortBy?: string,
  sortOrder?: "asc" | "desc",
  search?: string,
): Promise<AccessRule[]> {
  const totalCount = await getAccessRulesCount(filterBy, search);
  const chunkSize = API_LIMIT_MAX_RECORDS;
  const recordsPromises = [];
  const chunks = Math.ceil(totalCount / chunkSize);

  for (let i = 0; i < chunks; i++) {
    const offset = i * chunkSize;
    recordsPromises.push(getAccessRulesV2(filterBy, sortBy, sortOrder, offset, chunkSize, search));
  }

  const recordsChunks = await Promise.all(recordsPromises);
  const accessRules: AccessRule[] = [];
  for (const response of recordsChunks) {
    accessRules.push(...response);
  }
  return accessRules;
}

async function migrateDeprecatedAccessRules(
  oldRoleId: number,
  newRoleId: number,
): Promise<AccessRulesMigrateDeprecatedResponse> {
  try {
    const response: AxiosResponse = await authorizationApi.accessRulesApi.accessRulesMigrateDeprecated({
      oldRoleId,
      newRoleId,
    });
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getSsoGroupSubjectIds(searchQuery: string): Promise<Array<string>> {
  const response = await authorizationApi.accessRulesApi.getAccessRules(
    SubjectType.Group,
    undefined,
    undefined,
    undefined,
    undefined,
    undefined,
    undefined,
    undefined,
    undefined,
    undefined,
    [filterUtil.getContainsFilterString(AccessRulesSortFilterFields.SubjectId, searchQuery)],
  );
  const subjectIdsSet = new Set<string>();
  response.data.accessRules.forEach((accessRule: AccessRule) => {
    subjectIdsSet.add(accessRule.subjectId);
  });
  return Array.from(subjectIdsSet);
}
