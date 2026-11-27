/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { assetsApi } from '../../../../../../modules/api';
import {
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { EnvironmentAsset, Scope } from '../../../../../../modules/api/apis/assetsApi';

test.describe(`Count Assets API ${tags.teams.viper}`, { tag: [tags.services.assets, tags.runOn.parallelCI] }, () => {
    let projectName: string;
    let namespace: string;
    let project: Project;
    let environment: EnvironmentAsset;

    test.beforeAll(async function () {
        assetsApi.workspaceApi.accessToken = flowRootState.users.system_administrator.accessToken;
        projectName = objectGenerator.generateNameByPrefix('p');
        namespace = Utils.namespaceFromProjectName(projectName);
        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 0,
            maxAllowedGpus: 0
        });
    });

    test.afterAll('deleting project and namespace', async () => {
        try {
            await operations.deleteProjectAndNamespace(project.id);
            logger.info(`deleted project ${projectName} an namespace ${namespace}`);
        } catch (error) {
            logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
            throw error;
        }
    });
    test('Sanity check of GET /api/v1/asset/count filters', async ({}) => {
        await test.step('create assets', async (): Promise<void> => {
            environment = await environmentUtils.createEnvironmentWithArbitraryFields(project.id);
        });

        await test.step('should return asset count without filters', async () => {
            const { response, body } = await assetsApi.workspaceApi.countAssets(
                'environment',
                environment.meta.name,
                Scope.Project.toString(),
                project.id
            );
            expect(response.statusCode).toBe(200);
            expect(body.result).toBe(1);
        });

        await test.step('cleanup', async () => {
            await environmentUtils.delete(environment.meta.id);
        });
    });
});
