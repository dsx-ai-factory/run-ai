/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import http from 'http';
import { test, expect } from '@playwright/test';
import {
    tags,
    flowRootState,
    Utils,
    config,
    operations,
    TrainingWorkload,
    objectGenerator,
    waitUtils,
    logger,
    metricUtils,
    authUtils,
    orgUnitUtils,
    nodePoolUtils,
    usersManager,
    workloadsUtilsV2,
    crds,
    settings
} from '../../../../../../modules/helpers';
import { authorizationUtils } from '../../../../../../modules/helpers';
import { UserPostResponse } from '../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';
import { Project } from '../../../../../../modules/api/apis/orgUnitApi';
import { AccessRule } from '../../../../../../modules/api/apis/authorizationApi/generated/api';
import {
    shouldRunNativeWorkloadCrdTests,
    shouldRunTests
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

shouldRunTests(settings.isSelfHosted)(function () {
    shouldRunNativeWorkloadCrdTests(() => {
    test.describe(
        `Export workload metrics tests, tags: ${tags.resource.metrics} ${tags.teams.phoenix}`,
        { tag: [tags.services.metricsService, tags.services.workloads, tags.runOn.parallelCI] },
        () => {
            type TestContext = {
                project: Project;
                namespace: string;
                user: UserPostResponse;
                userToken: string;
                accessRule: AccessRule;
                workload: TrainingWorkload;
                workloadId: string;
            };

            const metricsCollectorRole = 'Integration metrics collector';
            const ctx = {} as TestContext;

            async function createProject(): Promise<{ project: Project; namespace: string }> {
                const projectName = objectGenerator.generateNameByPrefix('project');
                const namespace = Utils.namespaceFromProjectName(projectName);

                    const defaultNodePoolId = (
                        await nodePoolUtils.getDefaultNodePoolFromBackend()
                    ).id.toString();

                const project = await orgUnitUtils.createProjectAndValidate({
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ nodePool: { id: defaultNodePoolId, name: 'default' } }]
                });

                await waitUtils.waitForProjectRbac(namespace);

                return { project, namespace };
            }

            async function createMetricsCollectorUser(): Promise<{
                user: UserPostResponse;
                userToken: string;
                accessRule: AccessRule;
            }> {
                const adminToken = flowRootState.users.system_administrator.accessToken;

                const { body: user } = await usersManager.createLocalUser(
                    adminToken,
                    objectGenerator.generateEmailByPrefix('metrics_export_test'),
                    false,
                    false
                );

                const userToken = await authUtils.authenticateUser(user.username, user.tempPassword);

                const roles = await authorizationUtils.getRoles(adminToken, undefined, true);
                const role = roles.find((r) => r.name === metricsCollectorRole);
                expect(role).toBeDefined();

                const accessRule = await authorizationUtils.createAccessRuleForUser(
                    adminToken,
                    user.username,
                    role.id
                );

                return { user, userToken: userToken.accessToken, accessRule };
            }

            async function submitWorkload(): Promise<{ workload: TrainingWorkload; workloadId: string }> {
                const workloadName = objectGenerator.generateNameByPrefix('failed-workload');
                const workload = new TrainingWorkload(
                    config.quickStartImage.failedImage,
                    ctx.namespace,
                    workloadName
                );

                await operations.submitWorkloads([workload]);
                const workloadId = await workloadsUtilsV2.getWorkloadIdByWorkloadNameAndNamespace(
                    crds.runaiTrainingWorkloadCrd,
                    workload.name,
                    ctx.namespace
                );
                return { workload, workloadId };
            }

            test.beforeAll(async function () {
                const [{ project, namespace }, { user, userToken, accessRule }] = await Promise.all([
                    createProject(),
                    createMetricsCollectorUser()
                ]);
                ctx.project = project;
                ctx.namespace = namespace;
                ctx.user = user;
                ctx.userToken = userToken;
                ctx.accessRule = accessRule;
            });

            test.beforeEach(async function () {
                const { workload, workloadId } = await submitWorkload();
                ctx.workload = workload;
                ctx.workloadId = workloadId;
            });

            test.afterEach(async function () {
                try {
                    await operations.deleteWorkloads([ctx.workload]);
                } catch (error) {
                    logger.warn('Failed to clean up workload', error);
                }
            });

            test.afterAll(async function () {
                try {
                    await Promise.all([
                        orgUnitUtils.deleteProjectAndNamespace(ctx.project.id, ctx.project.name),
                        usersManager.deleteLocalUser(
                            flowRootState.users.system_administrator.accessToken,
                            ctx.user.id
                        ),
                        authorizationUtils.deleteAccessRule(
                            ctx.accessRule,
                            flowRootState.users.system_administrator.accessToken
                        )
                    ]);
                } catch (error) {
                    logger.warn('Failed to clean up project and user', error);
                }
            });

            test('should export metrics for failed workload', async () => {
                let result: { response: http.IncomingMessage; body: string };
                await waitUtils.waitForCondition(
                    async () => {
                        result = await metricUtils.getWorkloadsScrapeableMetrics(ctx.userToken);
                        return result.body.includes(ctx.workloadId);
                    },
                    `Waiting for scrapable metrics to contain workload id ${ctx.workloadId}`,
                    true,
                    300,
                    10
                );

                const { body } = result;
                logger.debug(`Scrapable metrics result: ${JSON.stringify(body)}`);
                const requiredMetricParts = [
                    'runai_external_pod_info',
                    `workload_id="${ctx.workloadId}"`,
                    `tenant_id="${flowRootState.cluster.tenantId}"`,
                    `cluster_id="${flowRootState.cluster.uuid}"`,
                    'status="Error"'
                ];
                for (const part of requiredMetricParts) {
                    expect(body).toContain(part);
                }
            });
        }
    );
    });
});
