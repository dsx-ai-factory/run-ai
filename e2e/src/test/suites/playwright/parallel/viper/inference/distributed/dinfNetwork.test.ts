/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { randomBytes } from 'crypto';
import { expect } from 'chai';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { isLwsInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { test } from '@playwright/test';
import { DistributedInferenceSpecSpec } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import { DistributedInferenceServingPortAccessAuthorizationTypeEnum } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceServingPortAccessAuthorizationTypeEnum';
import { distributedInferenceUtils, inferenceUtils } from '../../../../../../../modules/helpers';
import {
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

supportedFrom(SupportedVersions.V2_23)(function () {
    test.describe(
        `distributed inference workload network tests: ${tags.resource.distInferenceWorkload} ${tags.tool.api} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit}`,
        { tag: [tags.resource.distInferenceWorkload] },
        function () {
            let project: Project;
            let workloadId: string;
            let workloadName: string;
            let workloadSpec: DistributedInferenceSpecSpec;

            test.beforeAll(async function () {
                if (!(await isLwsInstalled())) {
                    logger.warn('LeaderWorkerSet (LWS) is not installed on the cluster, skipping test');
                    test.skip();
                }

                project = await orgUnitUtils.createProjectAndValidate({
                    name: objectGenerator.generateShortNameByPrefix('project'),
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 2, limit: 2 } }]
                });
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateNameByPrefix('workload');
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;

                workloadSpec = distributedInferenceUtils.getBaseWorkloadSpec();
            });

            test.afterAll(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and project ${project.status.namespace}; error: ${error}`
                    );
                }
            });

            test.afterEach(async function () {
                if (project === undefined) {
                    return;
                }

                try {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads of project ${project.status.namespace}; error: ${error}`
                    );
                }
            });

            shouldRunNativeWorkloadCrdTests(() =>
            test('creates a workload without external url', async function () {
                workloadSpec.servingPort.exposeExternally = false;

                // create the workload and wait for it to be ready
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                // verify the url is an internal one
                const url = await inferenceUtils.getServingUrl(workloadId);
                expect(url).to.contain('cluster.local');

                // test the serving endpoint by sending a request
                await inferenceUtils.sendInferenceRequest(project.status.namespace, url);
                })
            );

            test('creates a workload with external url', async function () {
                workloadSpec.servingPort.exposeExternally = true;

                // create the workload and wait for it to be ready
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                // verify the url is an external one
                const url = await inferenceUtils.getServingUrl(workloadId);
                expect(url).to.contain('https://');
                expect(url).to.not.contain('cluster.local');

                // test the serving endpoint by sending a request
                await inferenceUtils.sendInferenceRequestExternalUrl(url);
            });

            test('creates a workload with custom external url', async function () {
                // using a unique url to avoid conflicts from previous tests
                const customUrl = process.env.CONTROL_PLANE_URL + '/test-' + randomBytes(2).toString('hex');
                workloadSpec.servingPort.exposeExternally = true;
                workloadSpec.servingPort.exposedUrl = customUrl;

                // create the workload and wait for it to be ready
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                // verify the url is the custom one
                const url = await inferenceUtils.getServingUrl(workloadId);
                expect(url).to.equal(customUrl);

                // test the serving endpoint by sending a request
                await inferenceUtils.sendInferenceRequestExternalUrl(url);
            });

            test('creates a workload with authorizationType of authenticatedUsers', async function () {
                workloadSpec.servingPort.authorizationType =
                    DistributedInferenceServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers;

                // create the workload and wait for it to be ready
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                // test the serving endpoint by sending a request
                // should fail without a token and succeed with it
                const url = await inferenceUtils.getServingUrl(workloadId);
                await inferenceUtils.sendInferenceRequestExternalUrl(url, '', false);
                await inferenceUtils.sendInferenceRequestExternalUrl(
                    url,
                    flowRootState.users.ml_engineer.accessToken
                );
            });

            test('creates a workload with authorizationType of authorizedUsers', async function () {
                workloadSpec.servingPort.authorizationType =
                    DistributedInferenceServingPortAccessAuthorizationTypeEnum.AuthorizedUsers;
                workloadSpec.servingPort.authorizedUsers = [flowRootState.users.ml_engineer.email];

                // create the workload and wait for it to be ready
                [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                    workloadName,
                    workloadSpec,
                    String(project.id)
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(workloadId, Phase.Running, 120);

                // test the serving endpoint by sending a request
                // should fail without a token and succeed with it
                const url = await inferenceUtils.getServingUrl(workloadId);
                await inferenceUtils.sendInferenceRequestExternalUrl(url, '', false);
                await inferenceUtils.sendInferenceRequestExternalUrl(
                    url,
                    flowRootState.users.ml_engineer.accessToken
                );
            });
        }
    );
});
