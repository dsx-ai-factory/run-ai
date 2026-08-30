import type { IFilterBy, IPaginationFilter } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { httpService } from "@/services/infra/https.service/http.service";
import {
  type User,
  type UserCreationRequest,
  type UserPostResponse,
  type UserPatchRequest,
  type BulkUserCreationRequest,
  type BulkUserCreationResponse,
  type UserMinimal,
  UsersFilterSortFields,
} from "@/swagger-models/identity-manager-client";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";
import { userUtils } from "@/utils/user.util";

export const userService = {
  remove,
  create,
  resetPassword,
  listUsers,
  countUsers,
  isUserExist,
  getUserById,
  updateUser,
  getUsers,
  getUsersMinimal,
  resolveUsersByUsernames,
  countUsersClean,
  createBatch,
};

async function listUsers(filterBy: IFilterBy = {}, isLocalUsers: boolean): Promise<User[]> {
  const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(
    filterBy,
    [...Object.values(UsersFilterSortFields)],
    true,
  );
  const filterByArray = userUtils.getFilterByArrays(filters, isLocalUsers);
  return await getUsers(
    filterByArray,
    filters.sortBy as UsersFilterSortFields | undefined,
    filters.sortOrder,
    filters?.offset,
    filters?.limit,
    filters?.search,
  );
}

async function getUsers(
  filterBy?: Array<string>,
  sortBy?: string,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
  search?: string,
): Promise<User[]> {
  try {
    const response = await identityManagerApi.usersApi.getUsers(
      filterBy,
      sortBy as UsersFilterSortFields | undefined,
      sortOrder,
      offset,
      limit,
      search,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function countUsers(filterBy: IFilterBy = {}, isLocalUsers: boolean, externalFilter?: string[]): Promise<number> {
  const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(
    filterBy,
    [...Object.values(UsersFilterSortFields)],
    true,
  );
  try {
    const filterByArray = externalFilter ? externalFilter : userUtils.getFilterByArrays(filters, isLocalUsers);
    const response = await identityManagerApi.usersApi.countUsers(filterByArray, filters.search);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getUsersMinimal(
  search?: string,
  filterBy?: Array<string>,
  id?: Array<string>,
  limit?: number,
  options?: { signal?: AbortSignal },
): Promise<UserMinimal[]> {
  try {
    const response = await identityManagerApi.usersApi.getUsersMinimal(filterBy, search, id, limit, options);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function resolveUsersByUsernames(usernames: Array<string>, timeoutMs?: number): Promise<UserMinimal[]> {
  if (usernames.length === 0) return [];
  const filterBy = usernames.map((u) => filterUtil.getEqualsFilterString(UsersFilterSortFields.Username, u));
  const options = timeoutMs !== undefined ? { signal: AbortSignal.timeout(timeoutMs) } : undefined;
  return getUsersMinimal(undefined, filterBy, undefined, usernames.length, options);
}

async function countUsersClean(filterBy?: Array<string>, search?: string): Promise<number> {
  try {
    const response = await identityManagerApi.usersApi.countUsers(filterBy, search);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function isUserExist(username: string): Promise<boolean> {
  const filterByArray = [`username=@${username}`];
  const users = await userService.getUsers(filterByArray);
  return users.some((user) => user.username === username);
}

async function getUserById(userId: string): Promise<User> {
  try {
    const response = await identityManagerApi.usersApi.getUserById(userId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateUser(userId: string, patch: UserPatchRequest): Promise<User> {
  try {
    const response = await identityManagerApi.usersApi.patchUserById(userId, patch);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function remove(userId: string): Promise<void> {
  try {
    const response = await identityManagerApi.usersApi.deleteUserById(userId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function create(userCreationRequest: UserCreationRequest): Promise<UserPostResponse> {
  try {
    const response = await identityManagerApi.usersApi.createUser(userCreationRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createBatch(bulkUserCreationRequest: BulkUserCreationRequest): Promise<BulkUserCreationResponse> {
  try {
    const response = await identityManagerApi.usersApi.batchCreateUsers(bulkUserCreationRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function resetPassword(userId: string): Promise<string> {
  try {
    const response = await identityManagerApi.usersApi.resetUserPassword(userId);
    return response.data as string;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

export default userService;
