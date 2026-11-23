/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    computeResourceUtils,
    crds,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { assetsApi } from '../../../../../../../modules/api';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import {
    InferenceCreationRequestV2,
    ModelInferenceCreationRequestV2,
    InferenceV2,
    ModelInferenceV2
} from '../../../../../../../modules/api/apis/assetsApi';
import { expect } from 'chai';
import { test } from '@playwright/test';
import { StatusCodes } from 'http-status-codes';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { UserRole } from '../../../../../../../modules/helpers/config/roleMapping';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';


test.describe(
    `inference workload api tests:  ${tags.resource.inferenceWorkload} ${tags.tool.api} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {

        const CONCURRENCY_METRIC = 'concurrency';
        let project: Project;
        let workloadName: string;

        test.beforeAll(async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                this.skip();
            }

            project = await orgUnitUtils.createProjectAndValidate({
                name: objectGenerator.generateShortNameByPrefix('project'),
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 2, limit: 2 } }]
            });
        });

        test.beforeEach(function () {
            workloadName = objectGenerator.generateNameByPrefix('workload');
        });

        test.afterAll(async function () {
            try {
                if (project) {
                    await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id, 120);
                }
            } catch (error) {
                logger.warn('deleteProjectNamespaceAndWorkloads failed', error);
            }
            try {
                if (project) {
                    await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
                }
            } catch (error) {
                logger.warn('deleteProjectNamespaceAndWorkloads failed', error);
            }
        });

        async function createInferenceV2(
            role: UserRole,
            workloadCreationRequest: InferenceCreationRequestV2
        ): Promise<InferenceV2 | undefined> {
            try {
                assetsApi.inferenceApi.accessToken = flowRootState.users[role].accessToken;
                const inferenceWorkload =
                    await assetsApi.inferenceApi.createInferenceV2(workloadCreationRequest);
                expect(inferenceWorkload).to.not.be.undefined;
                expect(inferenceWorkload.response.statusCode).to.equal(StatusCodes.OK);
                return inferenceWorkload.body;
            } catch (error: unknown) {
                logger.error(`Could not create inference v2 with role ${role}. error ${error}`);
                return undefined;
            }
        }

        async function createModelInferenceV2(
            role: UserRole,
            workloadCreationRequest: ModelInferenceCreationRequestV2
        ): Promise<ModelInferenceV2 | undefined> {
            try {
                assetsApi.modelInferenceApi.accessToken = flowRootState.users[role].accessToken;
                const modelInferenceWorkload =
                    await assetsApi.modelInferenceApi.createModelInferenceV2(workloadCreationRequest);
                expect(modelInferenceWorkload.response.statusCode).to.equal(StatusCodes.OK);
                expect(modelInferenceWorkload.body).to.not.be.undefined;
                return modelInferenceWorkload.body;
            } catch (error: unknown) {
                logger.error(`Could not create model inference v2 with role ${role}. error ${error}`);
                return undefined;
            }
        }

        supportedFrom(SupportedVersions.V2_18)(function () {
            test('creates a basic inference workload (from assets v2)', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                logger.debug(`creating inference workload (from assets v2) ${workloadName}`);

                // get required assets ids
                const environmentId = await environmentUtils.getId('gpt2');
                expect(environmentId).to.not.be.undefined;

                const computeId = await computeResourceUtils.getId('cpu-only');
                expect(computeId).to.not.be.undefined;

                // init the workload creation request
                const request: InferenceCreationRequestV2 = {
                    meta: {
                        name: workloadName,
                        projectId: project.id.toString(),
                        clusterId: flowRootState.cluster.uuid
                    },
                    assets: {
                        environment: environmentId,
                        compute: computeId
                    },
                    specificRunParams: {
                        autoScaling: {
                            minReplicas: 1,
                            maxReplicas: 2,
                            thresholdMetric: CONCURRENCY_METRIC,
                            thresholdValue: 100
                        }
                    }
                };

                // create the workload in the control-plane
                const inferenceWorkload = await createInferenceV2('ml_engineer', request);
                expect(inferenceWorkload).to.not.be.undefined;
                expect(inferenceWorkload.meta).to.not.be.null;
                expect(inferenceWorkload.assets).to.not.be.null;
                expect(inferenceWorkload.assets.environment.id).to.equal(environmentId);
                expect(inferenceWorkload.assets.compute.id).to.equal(computeId);
                expect(inferenceWorkload.specificRunParams).to.not.be.null;
                expect(inferenceWorkload.specificRunParams.autoScaling.minReplicas).to.equal(1);
                expect(inferenceWorkload.specificRunParams.autoScaling.maxReplicas).to.equal(2);
                expect(inferenceWorkload.specificRunParams.autoScaling.thresholdMetric).to.equal(
                    'concurrency'
                );
                expect(inferenceWorkload.specificRunParams.autoScaling.thresholdValue).to.equal(100);

                logger.debug(`inference workload ${workloadName} created`);

                // check the workload in the cluster
                await waitUtils.waitForCrd(
                    crds.runaiInferenceWorkloadCrd,
                    workloadName,
                    project.status.namespace,
                    60
                );

                logger.debug(`inference workload ${workloadName} found in cluster`);
            });
        });
    }
);
