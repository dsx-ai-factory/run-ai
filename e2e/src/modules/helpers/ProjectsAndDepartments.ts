/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    logger,
    objectGenerator,
    flowRootState,
    crds,
    waitUtils,
    Utils,
    projectResourcesUtils,
    nodePoolUtils,
    orgUnitUtils
} from './';
import { orgUnitService, projectsDepartmentsApi } from '../../modules/api';
import { Project } from '../../modules/api/apis/backendUndocumentedApi';
import { StatusCodes } from 'http-status-codes';
import { expect } from 'chai';

export class ProjectsAndDepartments {
    async createDepartment(departmentNamePrefix: string, departmentOptions: object): Promise<string> {
        const departmentName = objectGenerator.generateNameByPrefix(departmentNamePrefix);
        const defaultNodePool = await nodePoolUtils.getDefaultNodePoolFromBackend();

        if (!departmentOptions['nodePoolsResources'] || !departmentOptions['nodePoolsResources'].length) {
            if (departmentOptions['resources']) {
                departmentOptions['nodePoolsResources'] = [
                    {
                        cpu: departmentOptions['resources']?.cpu,
                        gpu: departmentOptions['resources']?.gpu,
                        memory: departmentOptions['resources']?.memory,
                        nodePool: { id: String(defaultNodePool.id), name: defaultNodePool.name }
                    }
                ];
            } else if (departmentOptions['deservedGpus'] && departmentOptions['maxAllowedGpus']) {
                departmentOptions['nodePoolsResources'] = [
                    {
                        gpu: {
                            deserved: departmentOptions['deservedGpus'],
                            maxAllowed: departmentOptions['maxAllowedGpus']
                        },
                        memory: {
                            deserved: null,
                            maxAllowed: null
                        },
                        cpu: {
                            deserved: null,
                            maxAllowed: null
                        },
                        nodePool: { id: String(defaultNodePool.id), name: defaultNodePool.name }
                    }
                ];
            }
        }

        const createDepartmentReq = orgUnitUtils.getDepartmentCreateRequestFromV1(
            {
                name: departmentName,
                ...departmentOptions
            },
            flowRootState.cluster.uuid || '',
            defaultNodePool
        );
        const postRes = await orgUnitService.departmentApi.createDepartment(createDepartmentReq);
        expect(postRes.response).property('statusCode');
        expect(postRes.response.statusCode.valueOf()).equal(StatusCodes.CREATED);
        expect(postRes.body).property('id');
        const departmentId = postRes.body.id || '0';
        const getRes = await orgUnitUtils.getDepartment(departmentId);
        expect(getRes.id).equal(postRes.body.id);
        await waitUtils.waitForCrd(crds.departmentCrd, departmentName);
        return departmentId || '-1';
    }

    async createProject(
        projectNameSuffix: string,
        projectOptions: object,
        departmentId = 1
    ): Promise<Project> {
        const projectName = objectGenerator.generateNameByPrefix(`p-${projectNameSuffix}`);
        const project = await projectResourcesUtils.createProject(
            Utils.namespaceFromProjectName(projectName),
            {
                name: projectName,
                departmentId: departmentId,
                ...projectOptions
            }
        );
        logger.info(`created project with id ${project.id} and name ${project.name}`);
        return project;
    }

    async createProjectInDepartment(
        departmentId: number,
        projectOptions: object,
        suffix = 'team'
    ): Promise<Project> {
        const project = await this.createProject(suffix, projectOptions, departmentId);
        return project;
    }
}
