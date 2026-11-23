/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { identityManagerApi } from '../../api';
import {
    CountUsers200Response,
    User,
    UserCreationRequest,
    UserMinimal,
    UsersApi,
    UsersFilterSortFields
} from '../../api/apis/identityManagerApi/generated/api';

export class UsersUtils {
    async getUsersCount(accessToken: string, filterBy?: Array<string>): Promise<CountUsers200Response> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const response = await identityManagerApi.usersApi.countUsers(filterBy);
        return response.body;
    }

    async getUsersFilteredBy(
        accessToken: string,
        filterBy?: Array<string>,
        sortBy?: UsersFilterSortFields,
        sortOrder?: 'asc' | 'desc',
        offset?: number,
        limit?: number,
        search?: string
    ): Promise<User[]> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const response = await identityManagerApi.usersApi.getUsers(
            filterBy,
            sortBy,
            sortOrder,
            offset,
            limit,
            search
        );
        return response.body;
    }

    async getUsersMinimal(
        accessToken: string,
        filterBy?: Array<string>,
        search?: string,
        id?: Array<string>,
        limit?: number
    ): Promise<UserMinimal[]> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const response = await identityManagerApi.usersApi.getUsersMinimal(filterBy, search, id, limit);
        return response.body;
    }

    async validateNonEmailSSoUser(
        accessToken: string,
        username: string,
        expectedFirstName?: string,
        expectedLastName?: string
    ): Promise<User> {
        const allUsers = await this.getUsersFilteredBy(
            accessToken,
            undefined,
            UsersFilterSortFields.Username,
            undefined,
            undefined,
            undefined,
            username
        );
        const users = allUsers.filter((u) => u.username === username);
        expect(users.length, `expected user '${username}' in /users`).to.equal(1);
        expect(users[0].username).to.equal(
            username,
            'non-email SSO user should land in /users keyed by the upstream username'
        );

        const minimalAll = await this.getUsersMinimal(accessToken, [`username==${username}`]);
        const minimal = minimalAll.filter((u) => u.username === username);
        expect(minimal.length, `expected user '${username}' in /users-minimal by username filter`).to.equal(1);
        expect(minimal[0].username).to.equal(username);
        if (expectedFirstName !== undefined) {
            expect(minimal[0].firstName).to.equal(expectedFirstName);
        }
        if (expectedLastName !== undefined) {
            expect(minimal[0].lastName).to.equal(expectedLastName);
        }

        return users[0];
    }

    async changeUserPassword(
        accessToken: string,
        currentPassword: string,
        newPassword: string
    ): Promise<any> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const passwordResponse = await identityManagerApi.usersApi.changeUserPassword({
            currentPassword,
            newPassword
        });
        return passwordResponse;
    }

    async resetUserPassword(accessToken: string, userId: string): Promise<any> {
        identityManagerApi.usersApi.accessToken = accessToken;
        const passwordResponse = await identityManagerApi.usersApi.resetUserPassword(userId);
        return passwordResponse.response;
    }

    async createUserInSpecificTenant(
        accessToken: string,
        tenantUrl: string,
        user: UserCreationRequest
    ): Promise<any> {
        const apiInstance = new UsersApi();
        apiInstance.basePath = tenantUrl;
        apiInstance.accessToken = accessToken;
        const userResponse = await apiInstance.createUser(user);
        return userResponse.response;
    }

    async searchUsersMinimal(accessToken: string, search: string): Promise<UserMinimal[]> {
        return this.getUsersMinimal(accessToken, null, search);
    }

}
