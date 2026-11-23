/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authorizationUtils,
    config,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    settings,
    tags,
    unleashClient,
    Utils,
    apiUtils,
    idpUtils,
    waitUtils
} from '../../../../../../../modules/helpers';
import {
    idpNameOidc,
    keycloakClientName,
    realmName
} from '../../../../../../../modules/helpers/config/const';
import { Keycloak } from '../../../../../../../modules/helpers/utils/Keycloak';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { expect, test } from '@playwright/test';
import { WorkspacePage } from '../../../../../../../modules/ui/runaiV2/pages/WorkspacePage';
import { EWorkloadStatus, WorkloadPage } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { EnvironmentAsset, Scope, UidGidSource } from '../../../../../../../modules/api/apis/assetsApi';
import { EnvironmentCreationRequest } from '../../../../../../../modules/api/apis/assetsApi';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { kubeCli } from '../../../../../../../modules/cli';
import { EFeatureToggleKeys } from '../../../../../../../modules/helpers/utils/Unleash/featureToggle';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

shouldRunTests(settings.isSelfHosted && !isOCP() && process.env.INIT_SSO_USERS_AND_IDP !== 'true')(
    function () {
        test.describe(`Workspace submit with when uidGidSource set to fromIdpToken, tags: ${tags.teams.viper} ${tags.tool.ui}`, function () {
            let browserEngine: BrowserEngine;
            let auth: AuthPage;
            let projectName: string;
            let namespace: string;
            let project: Project;
            let environmentAsset: EnvironmentAsset;
            let workspace: WorkspacePage;
            let workload: WorkloadPage;

            async function createEnvironmentWithUidGidSource(
                environmentName: string,
                project: Project,
                uidGidSource: UidGidSource
            ): Promise<EnvironmentAsset> {
                const environmentCreationRequest: EnvironmentCreationRequest = {
                    meta: {
                        name: environmentName,
                        scope: Scope.Project,
                        projectId: project.id,
                        clusterId: project.clusterUuid,
                        workloadSupportedTypes: {
                            workspace: true,
                            training: true
                        }
                    },
                    spec: {
                        image: 'alpine',
                        command: 'sleep infinity',
                        uidGidSource: uidGidSource,
                        overrideUidGidInWorkspace: true
                    }
                };
                return await environmentUtils.create(environmentCreationRequest);
            }

            const ssoUser = 'sso-user';

            const testParams = {
                uid: 4444,
                gid: 5555
            };

            test.beforeAll(async function () {
                // Cleanup existing resources from failed previous runs
                try {
                    logger.debug('Cleaning up existing IDP and realm before test suite');
                    await idpUtils.deleteIdp(
                        flowRootState.users.system_administrator.accessToken,
                        idpNameOidc
                    );
                } catch (error) {
                    // 404 is expected if IDP doesn't exist
                    if (!error.message?.includes('404')) {
                        logger.debug(`IDP cleanup: ${error.message}`);
                    }
                }

                try {
                    await Keycloak.deleteRealm(realmName);
                    // Wait for Keycloak to finish deletion
                    await new Promise((resolve) => setTimeout(resolve, 1000));
                } catch (error) {
                    // Expected if realm doesn't exist
                    logger.debug(`Realm cleanup: ${error.message}`);
                }

                // Now proceed with fresh setup
                await Keycloak.setupSSORealmOIDCForUidGidClaims(
                    realmName,
                    keycloakClientName,
                    ssoUser,
                    config.controlPlaneAdminPassword,
                    testParams.uid.toString(),
                    testParams.gid.toString(),
                    []
                );

                await authorizationUtils.createAccessRuleForUser(
                    flowRootState.users.system_administrator.accessToken,
                    `${ssoUser}@example.com`,
                    flowRootState.users.system_administrator.role.id
                );

                projectName = objectGenerator.generateNameByPrefix('p-e2e-workspace');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 1,
                    maxAllowedGpus: 1
                });

                if (!settings.isSelfHosted) {
                    await unleashClient.disableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
                }
                await apiUtils.disableHybridSubmission();

                logger.info(`created project ${projectName}`);

                environmentAsset = await createEnvironmentWithUidGidSource(
                    objectGenerator.generateNameByPrefix('environment'),
                    project,
                    UidGidSource.FromIdpToken
                );
            });

            test.afterAll(async ({}, testInfo) => {
                if (!settings.isSelfHosted) {
                    await unleashClient.enableFeature(EFeatureToggleKeys.SHOW_HYBRID_SUBMISSION);
                }
                await apiUtils.enableHybridSubmission();

                try {
                    !!environmentAsset?.meta?.id && (await environmentUtils.delete(environmentAsset.meta.id));
                    !!project?.id && (await operations.deleteProjectAndNamespace(project.id));
                } catch (error) {
                    logger.error('Failed to cleanup environment/project:', error);
                } finally {
                    try {
                        await idpUtils.deleteIdp(
                            flowRootState.users.system_administrator.accessToken,
                            idpNameOidc
                        );
                        await Keycloak.deleteRealm(realmName);
                        await operations.forceCleanup();
                    } catch (error) {
                        logger.error('Failed to cleanup:', error);
                    }
                }
            });

            test.beforeEach(async () => {
                browserEngine = new BrowserEngine();
                await browserEngine.start();
                auth = new AuthPage(browserEngine.currentPage);
                workload = new WorkloadPage(browserEngine.currentPage);
                workspace = new WorkspacePage(browserEngine.currentPage);
            });

            test.afterEach('stop browser', async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, 'uidGidSourceFromIdpToken');
            });

            test('workspace submit', async function () {
                const workspaceName = objectGenerator.generateShortNameByPrefix('w-from-idp-token');
                await auth.loginSSO(`${ssoUser}@example.com`, config.controlPlaneAdminPassword);
                await workload.navigateToWorkloads();
                await workload.openWorkspaceForm();
                await workload.selectClusterFromDropDown(flowRootState.cluster.name);
                await workload.workloadFormSelectProject(projectName);
                await workspace.fillWorkspaceName(workspaceName);
                await workspace.waitForWorkspaceNameCount(workspaceName);
                await workspace.selectAssetsSubmissionOption();
                await workspace.clickOnContinue();
                await workspace.selectEnvironment(environmentAsset.meta.name);
                await workspace.selectComputeResource('cpu-only');
                await workspace.clickOnCreateWorkspaceBtn();
                await workspace.waitForWorkloadUIStatus(workspaceName, EWorkloadStatus.Running);
                if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Workspace)) {
                    await waitUtils.waitForBatchJobToExist(workspaceName, namespace);
                    const job = await kubeCli.getBatchJob(workspaceName, namespace);
                    const securityContext = job.spec?.template.spec?.containers[0].securityContext;
                    expect(securityContext?.runAsGroup).toBe(testParams.gid);
                    expect(securityContext?.runAsUser).toBe(testParams.uid);
                } else {
                    const interactiveWorkload: any = await kubeCli.getInteractiveWorkload(
                        workspaceName,
                        namespace
                    );
                    expect(interactiveWorkload.spec.runAsGid.value).toBe(testParams.gid);
                    expect(interactiveWorkload.spec.runAsUid.value).toBe(testParams.uid);
                }
                await workload.deleteWorkload(workspaceName);
            });
        });
    }
);
