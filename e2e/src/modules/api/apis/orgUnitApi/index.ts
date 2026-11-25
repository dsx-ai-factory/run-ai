/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { DepartmentsApi, NodeTypesApi, OrgUnitApi, ProjectsApi, ReportsApi, SettingsApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class OrgUnitService {
    orgUnitApi: OrgUnitApi;

    projectApi: ProjectsApi;

    departmentApi: DepartmentsApi;

    nodeTypeApi: NodeTypesApi;

    reportsApi: ReportsApi;

    settingsApi: SettingsApi;

    constructor(url: string) {
        this.orgUnitApi = new OrgUnitApi(url);
        this.orgUnitApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.projectApi = new ProjectsApi(url);
        this.projectApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.departmentApi = new DepartmentsApi(url);
        this.departmentApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.nodeTypeApi = new NodeTypesApi(url);
        this.nodeTypeApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.reportsApi = new ReportsApi(url);
        this.reportsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.settingsApi = new SettingsApi(url);
        this.settingsApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
