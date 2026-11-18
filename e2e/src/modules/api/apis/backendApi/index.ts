/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    AuthenticationApi,
    AuthorizationApi,
    TenantApi,
    ApplicationApi,
    SsoApi,
    ClustersApi,
    DepartmentsApi,
    OrganizationalApi,
    ProjectsApi
} from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class BackendApi {
    authenticationApi: AuthenticationApi;

    authorizationApi: AuthorizationApi;

    tenantApi: TenantApi;

    applicationApi: ApplicationApi;

    ssoApi: SsoApi;

    clusterApi: ClustersApi;

    departmentApi: DepartmentsApi;

    organizationalApi: OrganizationalApi;

    projectApi: ProjectsApi;

    constructor(url: string) {
        this.authenticationApi = new AuthenticationApi(url);
        this.authenticationApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.authorizationApi = new AuthorizationApi(url);
        this.authorizationApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.tenantApi = new TenantApi(url);
        this.tenantApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.applicationApi = new ApplicationApi(url);
        this.applicationApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.ssoApi = new SsoApi(url);
        this.ssoApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.clusterApi = new ClustersApi(url);
        this.clusterApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.departmentApi = new DepartmentsApi(url);
        this.departmentApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.organizationalApi = new OrganizationalApi(url);
        this.organizationalApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.projectApi = new ProjectsApi(url);
        this.projectApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
