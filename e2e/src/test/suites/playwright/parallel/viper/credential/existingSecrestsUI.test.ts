/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    credentialUtils,
    departmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    tags
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { BrowserEngine } from '../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { CredentialPage } from '../../../../../../modules/ui/runaiV2/pages/CredentialPage';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { SecretType } from '../../../../../../modules/api/apis/k8sObjectsTrackerApi/generated/model/secretType';
import { V1ObjectMeta } from '@kubernetes/client-node/';
import { test } from '@playwright/test';
import { expect } from 'chai';

const resourceTypes = [
    SecretType.AccessKey,
    SecretType.DockerRegistry,
    SecretType.Password,
    SecretType.Generic
];

test.describe(`Existing Secrets UI, tags: ${tags.tool.ui} ${tags.resource.credential} ${tags.teams.viper}`, function () {
    let project: Project;
    let projectB: Project;
    let secondDepartmentId: number;
    let browserEngine: BrowserEngine;
    let videoName: string;
    let auth: AuthPage;
    let credential: CredentialPage;
    let browserStarted = false;
    const createdSecrets: Array<Pick<V1ObjectMeta, 'name' | 'namespace'>> = [];

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

    async function cleanupCreatedSecrets(): Promise<void> {
        const secretsToDelete = createdSecrets.splice(0);
        for (const secret of secretsToDelete) {
            if (!secret.name || !secret.namespace) continue;
            try {
                await kubeCli.deleteSecret(secret.name, secret.namespace);
            } catch (e) {
                logger.warn(`failed to delete secret ${secret.namespace}/${secret.name}: ${e}`);
            }
        }
    }

    async function initBrowserAndLogin(): Promise<void> {
        await browserEngine.start();
        browserStarted = true;
        auth = new AuthPage(browserEngine.currentPage);
        credential = new CredentialPage(browserEngine.currentPage);
        await auth.signIn(
            flowRootState.users.system_administrator.email,
            flowRootState.users.system_administrator.password
        );
    }

    async function createCredentialUsingExistingSecret(options: {
        resourceType: SecretType;
        videoNamePrefix: string;
        hasClusterWideLabel: boolean;
        secretScope: Scope;
        secretScopeId: string;
        credentialScopeEntityName: string;
        secretScopeIdLabel?: string;
        secretNamespace?: string;
        isolationScopeId?: string;
    }): Promise<void> {
        const {
            resourceType,
            videoNamePrefix,
            hasClusterWideLabel,
            secretScope,
            secretScopeId,
            credentialScopeEntityName,
            secretScopeIdLabel,
            secretNamespace,
            isolationScopeId
        } = options;
        videoName = `${videoNamePrefix}-${resourceType}`;

        const nameOfNewCredential = objectGenerator.generateNameByPrefix(`new-${resourceType}`);
        const existingSecretMeta = await credentialUtils.createExistingSecret(
            resourceType,
            hasClusterWideLabel,
            secretScopeIdLabel,
            secretNamespace
        );
        createdSecrets.push(existingSecretMeta);

        await credentialUtils.waitForExistingSecret(
            existingSecretMeta.name,
            secretScope.toString(),
            secretScopeId,
            resourceType
        );

        if (isolationScopeId) {
            const secretsInOtherScope = await credentialUtils.getSecretInfo(
                existingSecretMeta.name,
                secretScope.toString(),
                isolationScopeId,
                resourceType
            );
            expect(secretsInOtherScope).to.be.empty;
        }

        await credential.goToPage();
        await credential.createCredentialFromExistingSecret(
            resourceType,
            nameOfNewCredential,
            existingSecretMeta.name,
            { scope: secretScope, scopeEntityName: credentialScopeEntityName }
        );
        await credential.deleteCredential(nameOfNewCredential);
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
        await cleanupCreatedSecrets();
        try {
            await Promise.all([project, projectB].map(deleteProjectAndNamespace));
        } finally {
            await departmentUtils.deleteDepartmentCleanup(secondDepartmentId);
        }
    });

    test.afterEach('stop browser and cleanup secret', async ({}, testInfo) => {
        try {
            if (browserStarted) {
                await browserEngine.stopAndArchive(testInfo, 'existing-' + videoName);
                browserStarted = false;
            }
        } catch (e) {
            logger.error(e);
        }

        await cleanupCreatedSecrets();
    });

    test.describe('create credential using existing secret', function () {
        test.beforeEach(async function () {
            await initBrowserAndLogin();
        });

        test.describe('secret with cluster-wide label', function () {
            test.describe('create cluster scoped secret using existing secret', function () {
                resourceTypes.forEach(function (resourceType) {
                    test(`${resourceType}`, async function () {
                        await createCredentialUsingExistingSecret({
                            resourceType,
                            videoNamePrefix: 'cluster-scoped',
                            hasClusterWideLabel: true,
                            secretScope: Scope.Cluster,
                            secretScopeId: flowRootState.cluster.uuid,
                            credentialScopeEntityName: flowRootState.cluster.name
                        });
                    });
                });
            });
        });

        test.describe('secret with department label', function () {
            test.describe('create department scoped secret using existing secret', function () {
                resourceTypes.forEach(function (resourceType) {
                    test(resourceType.toString(), async function () {
                        const scopeId = project.departmentId.toString();
                        await createCredentialUsingExistingSecret({
                            resourceType,
                            videoNamePrefix: 'department-scoped',
                            hasClusterWideLabel: false,
                            secretScope: Scope.Department,
                            secretScopeId: scopeId,
                            credentialScopeEntityName: 'default',
                            secretScopeIdLabel: scopeId,
                            isolationScopeId: projectB.departmentId.toString()
                        });
                    });
                });
            });
        });

        test.describe('secret with project namespace', function () {
            test.describe('create project scoped secret using existing secret', function () {
                resourceTypes.forEach(function (resourceType) {
                    test(resourceType.toString(), async function () {
                        await createCredentialUsingExistingSecret({
                            resourceType,
                            videoNamePrefix: 'project-scoped',
                            hasClusterWideLabel: false,
                            secretScope: Scope.Project,
                            secretScopeId: project.id.toString(),
                            credentialScopeEntityName: project.name,
                            secretNamespace: project.namespace,
                            isolationScopeId: projectB.id.toString()
                        });
                    });
                });
            });
        });
    });
});
