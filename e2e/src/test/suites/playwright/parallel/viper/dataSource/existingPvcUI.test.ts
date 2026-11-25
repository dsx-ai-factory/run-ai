/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    dataSourcesUtils,
    departmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    tags
} from '../../../../../../modules/helpers';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { V1ObjectMeta } from '@kubernetes/client-node/';
import { kubeCli } from '../../../../../../modules/cli';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { expect, test } from '@playwright/test';

test.describe(`Existing PVC UI, tags: ${tags.tool.ui} ${tags.resource.credential} ${tags.teams.viper}`, function () {
    let project: Project;
    let projectB: Project;
    let secondDepartmentId: number;
    let browserEngine: BrowserEngine;
    let videoName: string;
    let auth: AuthPage;
    let dataSource: DataSourcePage;
    let existingPVCMeta: V1ObjectMeta;
    let nameOfNewPvc: string;

    async function deleteProjectAndNamespace(projectToDelete: Project): Promise<void> {
        try {
            await operations.deleteProjectAndNamespace(projectToDelete.id);
            logger.info(`deleted project ${projectToDelete.name} and namespace ${projectToDelete.namespace}`);
        } catch (error) {
            logger.error(
                `failed to delete project ${projectToDelete.name} and namespace ${projectToDelete.namespace}`
            );
            throw error;
        }
    }

    test.beforeAll(async function () {
        browserEngine = new BrowserEngine();
        const secondDepartmentName = objectGenerator.generateShortNameByPrefix('department-for-e2e');
        const [createdProject, createdDepartment] = await Promise.all([
            projectResourcesUtils.createProjectAndNamespace('p-e2e-for-cred'),
            departmentUtils.createBackendDepartmentWithName(secondDepartmentName)
        ]);
        project = createdProject;
        secondDepartmentId = Number(createdDepartment.body.id);
        projectB = await projectResourcesUtils.createProjectAndNamespace('p-b', secondDepartmentId);

        await projectUtils.waitForRoleBindingForUserInProject(
            project.namespace,
            flowRootState.users.system_administrator.email,
            projectUtils.newRbacRunaiSystemAdministratorRoleBindingName
        );
    });
    test.afterAll(async function () {
        try {
            await Promise.all([project, projectB].map(deleteProjectAndNamespace));
        } finally {
            await departmentUtils.deleteDepartmentCleanup(secondDepartmentId);
        }
    });
    test.beforeEach(async function () {
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        dataSource = new DataSourcePage(browserEngine.currentPage);
        await auth.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
        nameOfNewPvc = objectGenerator.generateNameByPrefix('new-pvc');
    });

    test.afterEach('stop browser', async ({}, testInfo) => {
        await browserEngine.stopAndArchive(testInfo, 'existing-' + videoName);
        await kubeCli.deletePvc(existingPVCMeta.name, existingPVCMeta.namespace);
    });

    test.describe('pvc with cluster-wide label', function () {
        test('create cluster scoped pvc using existing pvc', async function () {
            videoName = 'cluster-scoped-pvc';
            existingPVCMeta = await dataSourcesUtils.createExistingPVC(true);
            await dataSource.createPvcFromExistingPvc(nameOfNewPvc, existingPVCMeta.name, {
                scope: Scope.Cluster,
                scopeEntityName: flowRootState.cluster.name
            });
        });
    });

    test.describe('pvc with department label', function () {
        test('create and delete department scoped pvc using existing pvc', async function () {
            videoName = 'department-scoped-pvc';
            const scopeId = project.departmentId.toString();
            existingPVCMeta = await dataSourcesUtils.createExistingPVC(false, scopeId);
            const pvcsInOtherDepartment = await dataSourcesUtils.getPvcInfo(
                existingPVCMeta.name,
                Scope.Department.toString(),
                projectB.departmentId.toString()
            );
            expect(
                pvcsInOtherDepartment,
                `PVC ${existingPVCMeta.name} should not be visible in department ${projectB.departmentId}`
            ).toEqual([]);
            await dataSource.createPvcFromExistingPvc(nameOfNewPvc, existingPVCMeta.name, {
                scope: Scope.Department,
                scopeEntityName: 'default'
            });
            await dataSource.deleteDataSource(nameOfNewPvc);
        });
    });

    test.describe('pvc with  project namespace', function () {
        test('create and delete project scoped pvc using existing pvc', async function () {
            videoName = 'project-scoped-pvc';
            existingPVCMeta = await dataSourcesUtils.createExistingPVC(false, undefined, project.namespace);
            const pvcsInOtherProject = await dataSourcesUtils.getPvcInfo(
                existingPVCMeta.name,
                Scope.Project.toString(),
                projectB.id.toString()
            );
            expect(
                pvcsInOtherProject,
                `PVC ${existingPVCMeta.name} should not be visible in project ${projectB.id}`
            ).toEqual([]);
            await dataSource.createPvcFromExistingPvc(nameOfNewPvc, existingPVCMeta.name, {
                scope: Scope.Project,
                scopeEntityName: project.name
            });
            await dataSource.deleteDataSource(nameOfNewPvc);
        });
    });
});
