/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ProjectsApi, DepartmentsApi } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class ProjectsDepartmentsApi {
    projectsApi: ProjectsApi;

    departmentsApi: DepartmentsApi;

    constructor(url: string) {
        this.projectsApi = new ProjectsApi(url);
        this.projectsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.departmentsApi = new DepartmentsApi(url);
        this.departmentsApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
