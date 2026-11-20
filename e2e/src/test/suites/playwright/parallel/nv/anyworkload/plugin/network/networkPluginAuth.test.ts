/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadAuthUtils,
    workloadsManagerUtils,
    workloadUtils,
    workloadsUtilsV2
} from '../../../../../../../../modules/helpers';
import {
    createNgcAuthSecret,
    createRunningNimService
} from '../../../../../../../../modules/helpers/utils/nimServiceUtils';
import { Project } from '../../../../../../../../modules/api/apis/orgUnitApi';
import { isNimOperatorInstalled } from '../../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { kubeCli } from '../../../../../../../../modules/cli';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, TestInfo } from '@playwright/test';

test.describe(
    'Anyworkload Controller auth',
    { tag: [tags.resource.mimService, tags.teams.viper] },
    function () {
        let project: Project;
        let workloadName: string;
        let adminEmail: string;
        let adminToken: string;
        let researcherToken: string;

        supportedFrom(SupportedVersions.V2_25)(function () {
            test.beforeAll(async function () {
                if (!(await isNimOperatorInstalled())) {
                    logger.warn('k8s-nim-operator is not installed, skipping tests');
                    test.skip();
                }

                adminEmail = flowRootState.users.system_administrator.email;
                adminToken = flowRootState.users.system_administrator.accessToken;
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

            test('Auth proxy deployment created with correct configuration (T5839304)', async function () {
                test.setTimeout(240 * 1000);
                await createRunningNimService({
                    name: workloadName,
                    project,
                    authorizationType: 'authorizedUsers',
                    authorizedUsers: [adminEmail],
                    skipExternalUrl: true
                });

                await workloadAuthUtils.verifyAuthProxyConfig(project.status.namespace, workloadName);
                await workloadAuthUtils.verifyUsernameLoaderInitContainer(project.status.namespace);
                await workloadAuthUtils.verifyAuthProxyServicePort(project.status.namespace);
                await workloadAuthUtils.verifyIngressBackendPointsTo(project.status.namespace, workloadName, true);
            });

            test('Full lifecycle: Create NIMService with auth, verify access, delete (T5839306)', async function () {
                test.setTimeout(240 * 1000);
                const { workloadId, externalUrl } = await createRunningNimService({
                    name: workloadName,
                    project,
                    authorizationType: 'authorizedUsers',
                    authorizedUsers: [adminEmail]
                });

                await workloadAuthUtils.verifyAuthProxyDeploymentExists(project.status.namespace);
                await workloadAuthUtils.verifyOwnerReferences(project.status.namespace);
                await workloadAuthUtils.verifyAuthProxyConfig(project.status.namespace, workloadName);

                await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, adminToken, 200);
                await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, researcherToken, 403);

                await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, '', [401, 302]);

                await workloadUtils.deleteMultipleWorkloads([workloadId]);
                await workloadAuthUtils.verifyAuthProxyDeploymentCleanedUp(project.status.namespace);

                await waitUtils.waitForCondition(
                    async () => {
                        try {
                            await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, adminToken, 404);
                            return true;
                        } catch {
                            return false;
                        }
                    },
                    'Wait for external URL to return 404 after deletion',
                    true,
                    30,
                    2
                );

                logger.info('Verified full lifecycle: create → access → delete → cleanup → inaccessible');
            });

            test('Network plugin E2E: service create mode with auth (T5839309)', async function () {
                test.setTimeout(240 * 1000);
                const { workloadId, externalUrl } = await createRunningNimService({
                    name: workloadName,
                    project,
                    authorizationType: 'authorizedUsers',
                    authorizedUsers: [adminEmail]
                });

                await workloadAuthUtils.verifyAuthProxyDeploymentExists(project.status.namespace);
                await workloadAuthUtils.verifyAuthProxyServicePort(project.status.namespace);
                await workloadAuthUtils.verifyIngressBackendPointsTo(project.status.namespace, workloadName, true);
                await workloadAuthUtils.verifyOwnerReferences(project.status.namespace);

                await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, adminToken, 200);
                await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, researcherToken, 403);

                await workloadUtils.deleteMultipleWorkloads([workloadId]);
                await workloadAuthUtils.verifyAuthProxyDeploymentCleanedUp(project.status.namespace);
            });

            test('Network plugin: service ref mode with auth (T5839310)', async function ({}, testInfo: TestInfo) {
                test.setTimeout(240 * 1000);
                const { externalUrl } = await createRunningNimService({
                    name: workloadName,
                    project,
                    authorizationType: 'authenticatedUsers'
                });

                await workloadAuthUtils.verifyAuthProxyDeploymentExists(project.status.namespace);
                await workloadAuthUtils.verifyIngressBackendPointsTo(project.status.namespace, workloadName, true);

                const networking = await workloadAuthUtils.getWorkloadRunnerNetworking(project.status.namespace, workloadName);
                expect(networking?.services, 'WorkloadRunner networking.services should exist').to.not.be
                    .undefined;
                expect(
                    networking.services.length,
                    'WorkloadRunner should have at least one service'
                ).to.be.greaterThan(0);
                const hasRef = networking.services.some((s: any) => s.ref);
                expect(hasRef, 'WorkloadRunner should use service ref mode').to.be.true;
                logger.info('Verified WorkloadRunner uses service ref mode');

                await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, adminToken, 200);
            });

            test('Network plugin: no auth — direct routing to workload service (T5839311)', async function () {
                test.setTimeout(240 * 1000);
                const { externalUrl } = await createRunningNimService({
                    name: workloadName,
                    project,
                    authorizationType: 'public'
                });

                await workloadAuthUtils.verifyNoAuthProxyDeployment(project.status.namespace);
                await workloadAuthUtils.verifyIngressBackendPointsTo(project.status.namespace, workloadName, false);
                await workloadAuthUtils.verifyExternalUrlHttpStatus(externalUrl, '', 200);
            });

            test('Network plugin validation: service ref requires auth (T5839312)', async function () {
                test.setTimeout(240 * 1000);
                await createRunningNimService({
                    name: workloadName,
                    project,
                    authorizationType: 'authenticatedUsers',
                    skipExternalUrl: true
                });

                const networking = await workloadAuthUtils.getWorkloadRunnerNetworking(project.status.namespace, workloadName);

                const refService = networking?.services?.find((s: any) => s.ref);
                expect(refService, 'NIMService should use service ref mode').to.not.be.undefined;

                const auth = refService?.auth || networking?.auth;
                expect(auth, 'When using ref mode, auth section must be present').to.not.be.undefined;
                expect(auth.authorizationType).to.equal('authenticatedUsers');

                logger.info(
                    'Verified WorkloadRunner has ref + auth (ref requires auth constraint satisfied)'
                );
            });

            test('Network plugin validation: ingress backend must match service (T5839314)', async function () {
                test.setTimeout(240 * 1000);
                await createRunningNimService({
                    name: workloadName,
                    project,
                    authorizationType: 'authenticatedUsers',
                    skipExternalUrl: true
                });

                const ingress = await workloadAuthUtils.waitForWorkloadIngress(project.status.namespace, workloadName);
                const backend = ingress.spec?.rules?.[0]?.http?.paths?.[0]?.backend;
                const backendServiceName = backend?.service?.name || '';
                const backendPort = backend?.service?.port?.number || backend?.service?.port?.name;

                expect(backendServiceName).to.not.be.empty;
                expect(backendPort).to.not.be.undefined;

                const services = await kubeCli.CoreV1Api.listNamespacedService({
                    namespace: project.status.namespace
                });
                const backendSvc = services.items.find((s) => s.metadata?.name === backendServiceName);
                expect(backendSvc, `Backend service ${backendServiceName} should exist in namespace`).to.not
                    .be.undefined;

                const hasMatchingPort = backendSvc!.spec?.ports?.some(
                    (p) => p.port === backendPort || p.name === backendPort
                );
                expect(hasMatchingPort, `Backend service should expose port ${backendPort}`).to.be.true;

                logger.info(
                    `Verified ingress backend ${backendServiceName}:${backendPort} matches existing service`
                );
            });

            test('Network plugin: cluster-wide defaults applied to ingress (T5839316)', async function () {
                test.setTimeout(240 * 1000);
                await createRunningNimService({
                    name: workloadName,
                    project,
                    authorizationType: 'public'
                });

                const ingress = await workloadAuthUtils.waitForWorkloadIngress(project.status.namespace, workloadName);

                const ingressClassName = ingress.spec?.ingressClassName;
                expect(ingressClassName, 'IngressClassName should be set from cluster config').to.not.be
                    .undefined;
                logger.info(`Verified ingressClassName: ${ingressClassName}`);

                const annotations = ingress.metadata?.annotations || {};
                const hasBodySize = Object.keys(annotations).some(
                    (k) =>
                        k.toLowerCase().includes('body-size') || k.toLowerCase().includes('proxy-body-size')
                );
                expect(hasBodySize, 'Body size annotation should be applied by default').to.be.true;
                logger.info('Verified body size annotation is present');

                expect(
                    Object.keys(annotations).length,
                    'Ingress should have cluster-wide annotations merged'
                ).to.be.greaterThan(0);
                logger.info(`Verified cluster-wide annotations: ${JSON.stringify(annotations)}`);
            });
        });
    }
);
