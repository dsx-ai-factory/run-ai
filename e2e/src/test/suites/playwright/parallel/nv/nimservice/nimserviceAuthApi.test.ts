/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    inferenceUtils,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    workloadAuthUtils,
    workloadsManagerUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import {
    createNgcAuthSecret,
    buildNimServiceAuthSpec,
    createRunningNimService,
    createNimServiceExpectingError
} from '../../../../../../modules/helpers/utils/nimServiceUtils';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { isNimOperatorInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, TestInfo } from '@playwright/test';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(
    'NIM service auth API',
    { tag: [tags.resource.mimService, tags.teams.viper] },
    function () {
        let project: Project;
        let workloadName: string;
        let adminEmail: string;
        let adminToken: string;
        let researcherEmail: string;
        let researcherToken: string;

        supportedFrom(SupportedVersions.V2_25)(function () {
            test.beforeAll(async function () {
                if (!(await isNimOperatorInstalled())) {
                    logger.warn('k8s-nim-operator is not installed, skipping tests');
                    test.skip();
                }

                adminEmail = flowRootState.users.system_administrator.email;
                adminToken = flowRootState.users.system_administrator.accessToken;
                researcherEmail = flowRootState.users.l1_researcher.email;
                researcherToken = flowRootState.users.l1_researcher.accessToken;

                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('project'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 2, limit: 2 } }]
                });
                await createNgcAuthSecret(project.status.namespace);
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateNameByPrefix('workload');
                workloadsManagerUtils.accessToken = adminToken;
            });

            test.afterAll(async function () {
                if (!project) return;
                try {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
                } catch (error) {
                    logger.error(`Failed to delete workloads for project ${project.name}: ${error}`);
                }

                try {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(`Failed to delete project ${project.name}: ${error}`);
                }
            });

            test.afterEach(async function () {
                if (!project) return;
                try {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
                    await workloadAuthUtils.verifyAuthProxyDeploymentCleanedUp(project.status.namespace);
                } catch (error) {
                    logger.error(`Failed to clean up workloads for project ${project.name}: ${error}`);
                }
            });

            test.describe('NIM service auth types', function () {
                test.setTimeout(240 * 1000);
                test('Create NIMService with public endpoint (default) (T5839294)', async function () {
                    const { externalUrl } = await createRunningNimService({
                        name: workloadName,
                        project,
                        authorizationType: 'public'
                    });

                    await workloadAuthUtils.verifyNoAuthProxyDeployment(project.status.namespace);
                    await inferenceUtils.sendInferenceRequestExternalUrl(
                        externalUrl,
                        '',
                        true,
                        false,
                        workloadName,
                        project.status.namespace
                    );
                });

                test('Create NIMService with authenticatedUsers authorization (T5839295)', async function () {
                    const { externalUrl } = await createRunningNimService({
                        name: workloadName,
                        project,
                        authorizationType: 'authenticatedUsers'
                    });

                    await workloadAuthUtils.verifyAuthProxyDeploymentExists(project.status.namespace);

                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, adminToken, 200);
                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, researcherToken, 200);

                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, '', [401, 302]);
                });

                test('Create NIMService with authorizedUsers authorization (T5839296)', async function () {
                    const { externalUrl } = await createRunningNimService({
                        name: workloadName,
                        project,
                        authorizationType: 'authorizedUsers',
                        authorizedUsers: [adminEmail]
                    });

                    await workloadAuthUtils.verifyAuthProxyDeploymentExists(project.status.namespace);

                    const networking = await workloadAuthUtils.getWorkloadRunnerNetworking(
                        project.status.namespace,
                        workloadName
                    );
                    logger.info(
                        `WorkloadRunner auth config: ${JSON.stringify(networking?.services?.[0]?.auth)}`
                    );

                    await workloadAuthUtils.verifyUsernameLoaderInitContainer(project.status.namespace);

                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, adminToken, 200);
                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, researcherToken, 403);

                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, '', [401, 302]);
                });

                test('Create NIMService with authorizedGroups authorization (T5839297)', async function ({}, testInfo: TestInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'oauth2-proxy --allowed-group enforcement requires a groups OIDC claim mapper on the Keycloak client; CI env not yet configured'
                    );
                    const groupName = 'system-administrators';
                    const { externalUrl } = await createRunningNimService({
                        name: workloadName,
                        project,
                        authorizationType: 'authorizedGroups',
                        authorizedGroups: [groupName]
                    });

                    await workloadAuthUtils.verifyAuthProxyDeploymentExists(project.status.namespace);
                    await workloadAuthUtils.verifyAuthProxyGroupArgs(project.status.namespace, [groupName]);

                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, adminToken, 200);
                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, researcherToken, 403);
                });

                test('Multiple authorized users can all access endpoint (T5839307)', async function ({}, testInfo: TestInfo) {
                    const { externalUrl } = await createRunningNimService({
                        name: workloadName,
                        project,
                        authorizationType: 'authorizedUsers',
                        authorizedUsers: [adminEmail, researcherEmail]
                    });

                    await workloadAuthUtils.verifyAuthProxyDeploymentExists(project.status.namespace);

                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, adminToken, 200);
                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, researcherToken, 200);

                    if (flowRootState.users.l2_researcher?.accessToken) {
                        await workloadAuthUtils.verifyExternalUrlHttpStatus(
                            externalUrl,
                            flowRootState.users.l2_researcher.accessToken,
                            403
                        );
                    }
                });

                test('Multiple authorized groups can all access endpoint (T5839308)', async function ({}, testInfo: TestInfo) {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'oauth2-proxy --allowed-group enforcement requires a groups OIDC claim mapper on the Keycloak client; CI env not yet configured'
                    );
                    const groups = ['ml-engineers', 'data-scientists'];
                    const { externalUrl } = await createRunningNimService({
                        name: workloadName,
                        project,
                        authorizationType: 'authorizedGroups',
                        authorizedGroups: groups
                    });

                    await workloadAuthUtils.verifyAuthProxyDeploymentExists(project.status.namespace);
                    await workloadAuthUtils.verifyAuthProxyGroupArgs(project.status.namespace, groups);

                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, adminToken, 200);
                    await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, researcherToken, 403);
                });
            });

            test.describe('NIM service auth validation', function () {
                test('Auth fields require exposeExternally=true (T5839298)', async function () {
                    const spec1 = buildNimServiceAuthSpec({
                        authorizationType: 'authenticatedUsers',
                        exposeExternally: false
                    });
                    await createNimServiceExpectingError(
                        workloadName,
                        project,
                        spec1,
                        400,
                        ['exposeexternally', 'expose'],
                        'auth requires exposeExternally=true'
                    );

                    const spec2 = buildNimServiceAuthSpec({
                        authorizationType: 'authorizedUsers',
                        authorizedUsers: ['alice@example.com'],
                        exposeExternally: false
                    });
                    await createNimServiceExpectingError(
                        objectGenerator.generateNameByPrefix('workload'),
                        project,
                        spec2,
                        400,
                        ['exposeexternally', 'expose'],
                        'auth requires exposeExternally=true'
                    );
                });

                test('authorizedUsers and authorizedGroups are mutually exclusive (T5839299)', async function () {
                    const spec1 = buildNimServiceAuthSpec({
                        authorizationType: 'authorizedUsers',
                        authorizedUsers: ['alice@example.com'],
                        authorizedGroups: ['ml-engineers']
                    });
                    await createNimServiceExpectingError(
                        workloadName,
                        project,
                        spec1,
                        400,
                        ['mutual', 'exclusive', 'together', 'both'],
                        'authorizedUsers and authorizedGroups mutually exclusive'
                    );

                    const spec2 = buildNimServiceAuthSpec({
                        authorizationType: 'authorizedGroups',
                        authorizedUsers: ['alice@example.com'],
                        authorizedGroups: ['ml-engineers']
                    });
                    await createNimServiceExpectingError(
                        objectGenerator.generateNameByPrefix('workload'),
                        project,
                        spec2,
                        400,
                        ['mutual', 'exclusive', 'together', 'both'],
                        'authorizedUsers and authorizedGroups mutually exclusive'
                    );
                });

                test('authorizedUsers required when authorizationType=authorizedUsers (T5839300)', async function () {
                    const spec1 = buildNimServiceAuthSpec({
                        authorizationType: 'authorizedUsers',
                        authorizedUsers: []
                    });
                    await createNimServiceExpectingError(
                        workloadName,
                        project,
                        spec1,
                        400,
                        ['authorizedusers', 'empty', 'required'],
                        'authorizedUsers must be non-empty'
                    );

                    const spec2 = buildNimServiceAuthSpec({
                        authorizationType: 'authorizedUsers'
                    });
                    await createNimServiceExpectingError(
                        objectGenerator.generateNameByPrefix('workload'),
                        project,
                        spec2,
                        400,
                        ['authorizedusers', 'required'],
                        'authorizedUsers is required'
                    );
                });

                test('authorizedGroups required when authorizationType=authorizedGroups (T5839301)', async function () {
                    const spec1 = buildNimServiceAuthSpec({
                        authorizationType: 'authorizedGroups',
                        authorizedGroups: []
                    });
                    await createNimServiceExpectingError(
                        workloadName,
                        project,
                        spec1,
                        400,
                        ['authorizedgroups', 'empty', 'required'],
                        'authorizedGroups must be non-empty'
                    );

                    const spec2 = buildNimServiceAuthSpec({
                        authorizationType: 'authorizedGroups'
                    });
                    await createNimServiceExpectingError(
                        objectGenerator.generateNameByPrefix('workload'),
                        project,
                        spec2,
                        400,
                        ['authorizedgroups', 'required'],
                        'authorizedGroups is required'
                    );
                });

                test('authorizedUsers/Groups must NOT be set for public type (T5839302)', async function () {
                    const spec1 = buildNimServiceAuthSpec({
                        authorizationType: 'public',
                        authorizedUsers: ['alice@example.com']
                    });
                    await createNimServiceExpectingError(
                        workloadName,
                        project,
                        spec1,
                        400,
                        ['public', 'authorizedusers', 'must not'],
                        'authorizedUsers must not be set for public type'
                    );

                    const spec2 = buildNimServiceAuthSpec({
                        authorizationType: 'authenticatedUsers',
                        authorizedGroups: ['ml-engineers']
                    });
                    await createNimServiceExpectingError(
                        objectGenerator.generateNameByPrefix('workload'),
                        project,
                        spec2,
                        400,
                        ['authenticatedusers', 'authorizedgroups', 'must not'],
                        'authorizedGroups must not be set for authenticatedUsers type'
                    );
                });

                test('Invalid authorizationType value (T5839303)', async function () {
                    const spec = buildNimServiceAuthSpec({
                        authorizationType: 'invalidType'
                    });
                    await createNimServiceExpectingError(
                        workloadName,
                        project,
                        spec,
                        400,
                        ['invalid', 'authorizationtype', 'public', 'authenticatedusers'],
                        'invalid authorizationType value'
                    );
                });
            });
        });
    }
);
