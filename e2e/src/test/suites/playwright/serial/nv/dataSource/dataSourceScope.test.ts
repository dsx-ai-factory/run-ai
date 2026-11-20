/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    objectGenerator,
    flowRootState,
    projectResourcesUtils,
    authorizationUtils,
    usersManager,
    operations,
    dataSourcesUtils
} from '../../../../../../modules/helpers';
import { assetsApi } from '../../../../../../modules/api';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { DataSourcePage } from '../../../../../../modules/ui/runaiV2/pages/DataSourcePage';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { tags } from '../../../../../../modules/helpers/config/tags';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { SubjectType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { ScopeType } from '../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { LoginUi } from '../../../../../../modules/ui/runaiV2/pages/LoginUiUtils';
import { authorizationApi } from '../../../../../../modules/api';
import { OrgUnitUtils } from '../../../../../../modules/helpers/utils/OrgUnitUtils';
import { logger } from '../../../../../../modules/helpers/logger/logger';

test.describe(`Create the data source by using the PVC pre-defined in project scope, tags: ${tags.tool.ui} ${tags.resource.datasource} ${tags.teams.loong} ${tags.runOn.dgxc}`, () => {
    supportedFrom(SupportedVersions.V2_18)(function () {
        const DEFAULT_STORAGE_CLASS = 'standard';

        async function loginAsUser(auth: AuthPage, loginUi: LoginUi, user: any): Promise<void> {
            await loginUi.logoutToLoginPage();
            const password = user.tempPassword ?? user.password;
            await auth.signIn(user.username, password);
        }

        async function deletePvcByName(accessToken: string, pvcName: string): Promise<void> {
            logger.info(`Deleting PVC: ${pvcName}`);
            try {
                assetsApi.datasourcesApi.accessToken = accessToken;
                const dataSources = (await assetsApi.datasourcesApi.listDatasourceAssets()).body.entries;
                logger.info(`Data sources names: ${dataSources.map((ds) => ds.meta.name).join(', ')}`);

                const pvcAsset = dataSources.find((ds) => ds.meta.name === pvcName);
                if (!pvcAsset) {
                    logger.warn(
                        `PVC ${pvcName} not found in data sources list, may have been already deleted`
                    );
                    return;
                }

                const pvcAssetId = pvcAsset.meta.id;
                await dataSourcesUtils.deletePvcAsset(pvcAssetId);
                logger.info(`Successfully deleted PVC: ${pvcName}`);
            } catch (e) {
                logger.error(`Error deleting PVC: ${pvcName}`, e);
            }
        }

        async function createPVC(
            dataSource: DataSourcePage,
            name: string,
            scope: Scope,
            scopeEntityName: string,
            storageClass: string = DEFAULT_STORAGE_CLASS
        ): Promise<void> {
            await dataSource.createPVC({
                name,
                scope,
                scopeEntityName,
                storageClass
            });
            await dataSourcesUtils.waitForDataSourceStatusInTable(name, 'No issues found', dataSource);
        }

        async function verifyUserCanOnlySeePVC(
            dataSource: DataSourcePage,
            shouldSee: string,
            shouldNotSee: string
        ): Promise<void> {
            await dataSource.selectDataSourceInTableByName(shouldSee);
            await expect(dataSource.page.locator(`td:has-text("${shouldNotSee}")`)).toHaveCount(0);
        }

        async function loginAndVerifyPVCAccess(
            auth: AuthPage,
            loginUi: LoginUi,
            dataSourcePage: DataSourcePage,
            user: any,
            pvcToSee: string,
            pvcNotToSee: string
        ): Promise<void> {
            await loginAsUser(auth, loginUi, user);
            await dataSourcePage.goToPage();
            await verifyUserCanOnlySeePVC(dataSourcePage, pvcToSee, pvcNotToSee);
        }

        let videoName: string;
        let createdProjects: any[] = [];
        let createdDepartments: any[] = [];
        let createdUsers: any[] = [];
        let createdPVCs: string[] = [];

        function generateUniqueEmail(prefix: string = 'user'): string {
            return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}@example.com`;
        }

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let dataSource: DataSourcePage;
        let loginUi: LoginUi;
        let adminUsername: string;
        let adminPassword: string;
        let adminUser: any;
        let adminTokens: any;
        let orgUnitUtils: OrgUnitUtils;

        async function assignUserToScope(
            user: any,
            role: any,
            projectOrDepartment: any,
            scopeType: ScopeType
        ): Promise<void> {
            authorizationApi.accessRulesApi.accessToken = adminTokens.idToken;
            const userMap = new Map([[user.username, role.id]]);
            await authorizationUtils.createAccessRulesForUsers(
                userMap,
                String(projectOrDepartment.id),
                scopeType,
                SubjectType.User
            );
        }

        test.beforeEach(async () => {
            createdProjects = [];
            createdDepartments = [];
            createdUsers = [];
            createdPVCs = [];
            browserEngine = new BrowserEngine();
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            dataSource = new DataSourcePage(browserEngine.currentPage);
            loginUi = new LoginUi(browserEngine.currentPage);
            adminUsername = flowRootState.users.system_administrator.email;
            adminPassword = flowRootState.users.system_administrator.password;
            adminUser = { username: adminUsername, password: adminPassword };
            adminTokens = flowRootState.users.system_administrator;
            orgUnitUtils = new OrgUnitUtils();
        });

        test.afterEach(async ({}, testInfo) => {
            try {
                for (const pvc of createdPVCs) {
                    await deletePvcByName(adminTokens.idToken, pvc);
                }
                for (const project of createdProjects) {
                    try {
                        await operations.deleteProjectAndNamespace(project.id);
                    } catch (e) {
                        logger.error('Error deleting project:', project.id, e);
                    }
                }
                for (const dept of createdDepartments) {
                    try {
                        await orgUnitUtils.deleteDepartment(dept.id);
                    } catch (e) {
                        logger.error('Error deleting department:', dept.id, e);
                    }
                }
                for (const user of createdUsers) {
                    try {
                        await usersManager.deleteLocalUser(adminTokens.idToken, user.id);
                    } catch (e) {
                        logger.error('Error deleting user:', user.username, e);
                    }
                }
            } catch (e) {
                logger.error('Error during cleanup:', e);
            }
            await browserEngine.stopAndArchive(testInfo, videoName);
        });

        test('T3990117: create PVC data sources in different projects and verify user access', async () => {
            videoName = 'dataSourceScope-T3990117';
            test.setTimeout(300 * 1000);
            let user1, user2, project1, project2, pvcName1, pvcName2;
            const user1Email = generateUniqueEmail('user1');
            const user2Email = generateUniqueEmail('user2');

            await test.step('Create projects (project1, project2)', async () => {
                project1 = await projectResourcesUtils.createProjectAndNamespace(
                    objectGenerator.generateNameByPrefix('project1')
                );
                project2 = await projectResourcesUtils.createProjectAndNamespace(
                    objectGenerator.generateNameByPrefix('project2')
                );
                createdProjects.push(project1, project2);
            });

            await test.step('Create users and assign researcher role in project1 and project2 scopes', async () => {
                const researcherRole = flowRootState.users.l1_researcher.role;
                user1 = (await usersManager.createLocalUser(adminTokens.idToken, user1Email, false)).body;
                user2 = (await usersManager.createLocalUser(adminTokens.idToken, user2Email, false)).body;
                await assignUserToScope(user1, researcherRole, project1, ScopeType.Project);
                await assignUserToScope(user2, researcherRole, project2, ScopeType.Project);
                createdUsers.push(user1, user2);
            });

            await test.step('Create PVC1 in project1 scope and PVC2 in project2 scope', async () => {
                await auth.signIn(adminUsername, adminPassword);
                await dataSource.goToPage();
                pvcName1 = objectGenerator.generateNameByPrefix('e2e-pvc1');
                await createPVC(dataSource, pvcName1, Scope.Project, project1.name, DEFAULT_STORAGE_CLASS);
                pvcName2 = objectGenerator.generateNameByPrefix('e2e-pvc2');
                await createPVC(dataSource, pvcName2, Scope.Project, project2.name, DEFAULT_STORAGE_CLASS);
                createdPVCs.push(pvcName1, pvcName2);
            });

            await test.step('Verify user1 (project1 scope) can only see PVC1 (project1 scope)', async () => {
                await loginAndVerifyPVCAccess(auth, loginUi, dataSource, user1, pvcName1, pvcName2);
            });

            await test.step('Verify user2 (project2 scope) can only see PVC2 (project2 scope)', async () => {
                await loginAndVerifyPVCAccess(auth, loginUi, dataSource, user2, pvcName2, pvcName1);
            });
        });

        test('T3990118: create cluster scope PVC and verify project user can see it', async () => {
            videoName = 'dataSourceScope-T3990118';
            let user1, project1, clusterPVCName;
            const user1Email = generateUniqueEmail('user1');

            await test.step('Create project (project1)', async () => {
                project1 = await projectResourcesUtils.createProjectAndNamespace(
                    objectGenerator.generateNameByPrefix('project1')
                );
                createdProjects.push(project1);
            });

            await test.step('Create user and assign researcher role in project1 scope', async () => {
                const researcherRole = flowRootState.users.l1_researcher.role;
                user1 = (await usersManager.createLocalUser(adminTokens.idToken, user1Email, false)).body;
                await assignUserToScope(user1, researcherRole, project1, ScopeType.Project);
                createdUsers.push(user1);
            });

            await test.step('Create PVC in cluster scope', async () => {
                await auth.signIn(adminUsername, adminPassword);
                await dataSource.goToPage();
                clusterPVCName = objectGenerator.generateNameByPrefix('e2e-cluster-pvc');
                await createPVC(
                    dataSource,
                    clusterPVCName,
                    Scope.Cluster,
                    flowRootState.cluster.name,
                    DEFAULT_STORAGE_CLASS
                );
                await dataSourcesUtils.waitForDataSourceStatusInTable(
                    clusterPVCName,
                    'No issues found',
                    dataSource
                );
                createdPVCs.push(clusterPVCName);
            });

            await test.step('Verify user1 (project1 scope) can see cluster scope PVC', async () => {
                await loginAsUser(auth, loginUi, user1);
                await dataSource.goToPage();
                const pvcRow = dataSource.page.locator(`tr:has(td:has-text("${clusterPVCName}"))`);
                await expect(pvcRow).toHaveCount(1);
            });
        });

        test('T3990119: department/project scope PVC access control', async () => {
            videoName = 'dataSourceScope-T3990119';
            test.setTimeout(300 * 1000);
            let user1Dept,
                user2Dept,
                department1,
                department2,
                projectDepartment1,
                projectDepartment2,
                pvcDept1,
                pvcProjDept2;
            const user1DeptEmail = generateUniqueEmail('user1-dept');
            const user2DeptEmail = generateUniqueEmail('user2-dept');
            const departmentName1 = objectGenerator.generateNameByPrefix('department1');
            const departmentName2 = objectGenerator.generateNameByPrefix('department2');

            await test.step('Create departments and projects in department scopes', async () => {
                department1 = await orgUnitUtils.createDepartment({
                    name: departmentName1,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
                            cpu: null,
                            memory: null,
                            nodePool: null
                        }
                    ]
                });
                department2 = await orgUnitUtils.createDepartment({
                    name: departmentName2,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [
                        {
                            gpu: { deserved: 0, limit: -1, overQuotaWeight: 0 },
                            cpu: null,
                            memory: null,
                            nodePool: null
                        }
                    ]
                });
                projectDepartment1 = await projectResourcesUtils.createProjectAndNamespace(
                    objectGenerator.generateNameByPrefix('project-department1'),
                    Number(department1.id),
                    0,
                    0
                );
                projectDepartment2 = await projectResourcesUtils.createProjectAndNamespace(
                    objectGenerator.generateNameByPrefix('project-department2'),
                    Number(department2.id),
                    0,
                    0
                );
                createdProjects.push(projectDepartment1, projectDepartment2);
                createdDepartments.push(department1, department2);
            });

            await test.step('Create users and assign researcher role in department1 and department2 scopes', async () => {
                const researcherRole = flowRootState.users.l1_researcher.role;
                user1Dept = (await usersManager.createLocalUser(adminTokens.idToken, user1DeptEmail, false)).body;
                user2Dept = (await usersManager.createLocalUser(adminTokens.idToken, user2DeptEmail, false)).body;
                await assignUserToScope(user1Dept, researcherRole, department1, ScopeType.Department);
                await assignUserToScope(user2Dept, researcherRole, department2, ScopeType.Department);
                createdUsers.push(user1Dept, user2Dept);
            });

            await test.step('Create PVC in department1 scope and PVC in projectDepartment2 scope', async () => {
                await auth.signIn(adminUsername, adminPassword);
                await dataSource.goToPage();
                pvcDept1 = objectGenerator.generateNameByPrefix('e2e-pvc-dept1');
                await createPVC(
                    dataSource,
                    pvcDept1,
                    Scope.Department,
                    department1.name,
                    DEFAULT_STORAGE_CLASS
                );
                pvcProjDept2 = objectGenerator.generateNameByPrefix('e2e-pvc-proj-dept2');
                await createPVC(
                    dataSource,
                    pvcProjDept2,
                    Scope.Project,
                    projectDepartment2.name,
                    DEFAULT_STORAGE_CLASS
                );
                createdPVCs.push(pvcDept1, pvcProjDept2);
            });

            await test.step('Verify user1Dept (department1 scope) can only see PVC in department1 scope', async () => {
                await loginAsUser(auth, loginUi, user1Dept);
                await dataSource.goToPage();
                await verifyUserCanOnlySeePVC(dataSource, pvcDept1, pvcProjDept2);
            });

            await test.step('Verify user2Dept (department2 scope) can only see PVC in projectDepartment2 scope', async () => {
                await loginAsUser(auth, loginUi, user2Dept);
                await dataSource.goToPage();
                await verifyUserCanOnlySeePVC(dataSource, pvcProjDept2, pvcDept1);
            });
        });
    });
});
