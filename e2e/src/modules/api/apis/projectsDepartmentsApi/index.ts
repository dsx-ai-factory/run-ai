/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ProjectsV1Api, DepartmentsV1Api } from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';
import type { ProjectStatus as OrgUnitProjectStatus } from '../orgUnitApi/generated/model/projectStatus';

export * from './generated/model/models';
export {
    GetDepartmentsV1200ResponseInner as Department,
    AssignedResourcesV1 as AssignedResources,
    AssignedResourcesV1Cpu as AssignedResourcesCpu,
    AssignedResourcesV1Gpu as AssignedResourcesGpu,
    AssignedResourcesV1Memory as AssignedResourcesMemory,
    NodePoolAssignedResourcesV1Response as NodePoolAssignedResources,
    NodePoolAssignedResourcesV1ResponseNodePool as NodePoolAssignedResourcesCreateNodePool,
    DepartmentV1AccessControl as DepartmentAccessControl,
    NodePoolAssignedResourcesV1Cpu as NodePoolAssignedResourcesCreateCpu,
    NodePoolAssignedResourcesV1Gpu as NodePoolAssignedResourcesCreateGpu,
    NodePoolAssignedResourcesV1Memory as NodePoolAssignedResourcesCreateMemory,
    ProjectV1NodeAffinityTrain as NodeAffinity,
    ProjectV1NodeAffinityTrain as JobsNodeAffinityInteractive,
    ProjectV1NodeAffinityTrain as JobsNodeAffinityTrain
} from './generated/model/models';

import type {
    ProjectV1,
    ProjectV1NodeAffinity,
    AssignedResourcesV1,
    ProjectV1AllOfStatus,
    ProjectV1CreationRequest,
    NodePoolAssignedResourcesV1
} from './generated/model/models';

// Legacy helpers also handle creation inputs and partial responses, while the public
// spec marks fields required for full responses and node-pool updates.
export type Project = Partial<Omit<ProjectV1, 'status' | 'permissions' | 'resources' | 'nodeAffinity'>> & {
    status?: Omit<Partial<ProjectV1AllOfStatus>, 'phase'> & Omit<OrgUnitProjectStatus, 'phase'> & {
        phase?: ProjectV1AllOfStatus['phase'] | OrgUnitProjectStatus['phase'];
    };
    permissions?: Partial<ProjectV1['permissions']>;
    resources?: AssignedResourcesV1;
    nodeAffinity?: ProjectV1NodeAffinity;
};
export type NodePoolAssignedResourcesCreate = Omit<NodePoolAssignedResourcesV1, 'id' | 'gpu' | 'cpu' | 'memory'> & {
    id?: number;
    gpu?: Partial<NodePoolAssignedResourcesV1['gpu']>;
    cpu?: Partial<NodePoolAssignedResourcesV1['cpu']>;
    memory?: Partial<NodePoolAssignedResourcesV1['memory']>;
};
export type ProjectCreateRequest = Omit<ProjectV1CreationRequest, 'nodePoolsResources'> & {
    nodePoolsResources?: NodePoolAssignedResourcesCreate[];
};

// Preserve the helper-facing names while using the same legacy routes from the org-unit spec.
class ProjectsApi extends ProjectsV1Api {
    getProjects(...args: Parameters<ProjectsV1Api['getProjectsV1']>): ReturnType<ProjectsV1Api['getProjectsV1']> {
        return this.getProjectsV1(...args);
    }

    updateProjectStatus(...args: Parameters<ProjectsV1Api['updateProjectStatusV1']>): ReturnType<ProjectsV1Api['updateProjectStatusV1']> {
        return this.updateProjectStatusV1(...args);
    }

}

class DepartmentsApi extends DepartmentsV1Api {
    getDepartments(...args: Parameters<DepartmentsV1Api['getDepartmentsV1']>): ReturnType<DepartmentsV1Api['getDepartmentsV1']> {
        return this.getDepartmentsV1(...args);
    }

    updateDepartmentAdmins(...args: Parameters<DepartmentsV1Api['setDepartmentAdminsV1']>): ReturnType<DepartmentsV1Api['setDepartmentAdminsV1']> {
        return this.setDepartmentAdminsV1(...args);
    }

}

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
