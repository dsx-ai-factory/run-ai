/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    dataSourcesUtils,
    departmentUtils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils
} from '../../../../../../modules/helpers';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { kubeCli } from '../../../../../../modules/cli';
import { test } from '@playwright/test';

test.describe(`Cluster-wide-data-source API, tags: ${tags.resource.datasource} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit}`, { tag: [tags.services.assets, tags.runOn.parallelCI] }, () => {
    let projectA: Project;
    let projectB: Project;
    let dataSourceName: string;

    async function createProjectInDifferentDepartment(): Promise<Project> {
        const secondDepartmentName = objectGenerator.generateShortNameByPrefix('department-for-e2e');
        const departmentIdForProjB = (
            await departmentUtils.createBackendDepartmentWithName(secondDepartmentName)
        ).body.id;
        projectB = await projectResourcesUtils.createProjectAndNamespace('p-b', Number(departmentIdForProjB));
        return projectB;
    }

    async function printAssetSyncLogs(): Promise<void> {
        logger.info(`test ${this.currentTest.title} failed, asset-sync logs:`);
        await kubeCli.printLogs('assets-sync');
    }

    test.beforeEach(async ({}, testInfo) => {
        projectA = await projectResourcesUtils.createProjectAndNamespace('p-a');
        const scope = testInfo.title.split(' ')[0];
        if (scope === Scope.Department.toString()) {
            projectB = await createProjectInDifferentDepartment();
        } else {
            projectB = await projectResourcesUtils.createProjectAndNamespace('p-b');
        }
    });

    test.afterEach(async ({}, testInfo) => {
        if (testInfo.status === 'failed') {
            await printAssetSyncLogs();
        }

        for (const p of [projectA, projectB]) {
            try {
                await operations.deleteProjectAndNamespace(p.id);
            } catch (error) {
                logger.error(`failed to delete project ${p.name} and namespace ${p.namespace}`);
                throw error;
            }
        }
    });

    async function waitForPvc(pvcName: string, scope: Scope): Promise<void> {
        if ([Scope.Project, Scope.Department].includes(scope)) {
            logger.info(`check that there is no claim name ${pvcName} in namespace ${projectA.namespace}`);
            await waitUtils.waitForPvc(pvcName, projectB.namespace, 10, false);
        }

        if (scope !== Scope.Project) {
            await waitUtils.waitForPvc(pvcName, 'runai');
        }
    }

    test.describe('PVC', function () {
        const scopes = [Scope.Department, Scope.Project];
        test.describe('department-wide PVC', function () {
            scopes.forEach(function (scope) {
                test(`${scope} scope`, async function () {
                    dataSourceName = objectGenerator.generateShortNameByPrefix(`pvc-api-${scope}-scope`);
                    const createdPvc = await dataSourcesUtils.createPVC(dataSourceName, scope, projectA);

                    await waitForPvc(dataSourceName, scope);
                    await dataSourcesUtils.deletePvcAsset(createdPvc.meta.id);
                });
            });
        });
    });
});
