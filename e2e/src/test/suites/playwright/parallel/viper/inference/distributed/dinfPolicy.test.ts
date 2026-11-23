/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    policyUtils,
    tags,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { test } from '@playwright/test';
import { expect } from 'chai';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { isLwsInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { createDistributedInference } from '../inferenceUtils.js';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/policyUtils';
import { ScopeType, DistributedInferencePolicyV2 } from '../../../../../../../modules/api/apis/policyApi';
import { StatusCodes } from 'http-status-codes';
import { ComputeAssetSpec } from '../../../../../../../modules/api/apis/assetsApi';

supportedFrom(SupportedVersions.V2_23)(function () {
    test.describe(
        `distributed inference workload policy tests: ${tags.resource.distInferenceWorkload} ${tags.tool.api} ${tags.teams.viper}`,
        { tag: [tags.resource.distInferenceWorkload] },
        function () {
            let project: Project;
            let projectName: string;
            let workloadId: string;
            let workloadName: string;
            let policyName: string;

            test.beforeAll(async function () {
                if (!(await isLwsInstalled())) {
                    logger.warn('LeaderWorkerSet (LWS) is not installed on the cluster, skipping test');
                    test.skip();
                }

                projectName = objectGenerator.generateShortNameByPrefix('p');
                project = await orgUnitUtils.createProjectAndValidate({
                    name: projectName,
                    clusterId: flowRootState.cluster.uuid,
                    resources: [{ gpu: { deserved: 8, limit: 8 } }]
                });
                policyName = `${project.name}-dinf-policy`;
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateNameByPrefix('w');
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
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

                // Cleanup policy if it exists
                try {
                    await policyUtils.deleteWorkloadPolicyV2(WorkloadType.DistributedInference, {
                        scope: 'project',
                        scopeId: project.id.toString()
                    });
                } catch (error) {
                    logger.debug(`Policy cleanup error: ${error.message}`);
                }
            });

            test('should create simple distributed inference policy with basic defaults', async function () {
                const defaults = {
                    workers: 1,
                    replicas: 1
                };

                logger.info(
                    `Creating simple distributed inference policy with defaults: ${JSON.stringify(defaults)}`
                );

                // Create the policy
                const policyResponse = await policyUtils.createPolicyV2(
                    policyName,
                    project.id.toString(),
                    WorkloadType.DistributedInference,
                    defaults,
                    undefined, // no rules
                    ScopeType.Project
                );

                expect(policyResponse.response.statusCode).to.equal(StatusCodes.OK);
                logger.info(`Policy response: ${JSON.stringify(policyResponse.body, null, 2)}`);

                const effectiveDefaults = (policyResponse.body as DistributedInferencePolicyV2).effective
                    .defaults;
                expect(effectiveDefaults.workers).to.equal(defaults.workers);
                expect(effectiveDefaults.replicas).to.equal(defaults.replicas);

                logger.info('Successfully created simple distributed inference policy');
            });

            test('should create distributed inference policy and validate workload creation', async function () {
                const defaults = {
                    workers: 1,
                    replicas: 1
                };

                logger.info(
                    `Creating distributed inference policy with basic defaults: ${JSON.stringify(defaults)}`
                );

                // Create the policy
                const policyResponse = await policyUtils.createPolicyV2(
                    policyName,
                    project.id.toString(),
                    WorkloadType.DistributedInference,
                    defaults,
                    undefined, // no rules
                    ScopeType.Project
                );

                expect(policyResponse.response.statusCode).to.equal(StatusCodes.OK);
                const effectiveDefaults = (policyResponse.body as DistributedInferencePolicyV2).effective
                    .defaults;
                expect(effectiveDefaults.workers).to.equal(defaults.workers);
                expect(effectiveDefaults.replicas).to.equal(defaults.replicas);

                logger.info('Successfully created distributed inference policy with defaults');

                const workloadSpec = {
                    servingPort: {
                        port: 8080
                    },
                    leader: {
                        image: config.quickStartImage.inferenceHelloWorld
                    },
                    worker: {
                        image: config.quickStartImage.inferenceHelloWorld
                    }
                };

                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    workloadSpec,
                    project
                );

                // Wait for the workload to be running
                await waitUtils.waitForBackendWorkloadPhaseInHistory(
                    workloadId,
                    workloadName,
                    Phase.Pending,
                    120
                );

                // Validate that the workload was created successfully
                const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);

                // Validate the basic workload structure
                expect(workload.body.additionalFields.workers).to.equal(defaults.workers);
                expect(workload.body.additionalFields.specReplicas).to.equal(defaults.replicas);

                // Validate the images were set correctly
                expect(workload.body.images).to.include(config.quickStartImage.inferenceHelloWorld);

                logger.info('Successfully validated that workload was created with policy in place');
            });

            test('should validate workload overrides policy defaults when explicitly specified', async function () {
                const defaults = {
                    leader: {
                        compute: {
                            cpuCoreRequest: 0.5,
                            cpuMemoryRequest: '512Mi'
                        }
                    },
                    worker: {
                        compute: {
                            cpuCoreRequest: 0.3,
                            cpuMemoryRequest: '256Mi'
                        }
                    }
                };

                logger.info('Creating policy with defaults for override test');

                // Create the policy
                await policyUtils.createPolicyV2(
                    policyName,
                    project.id.toString(),
                    WorkloadType.DistributedInference,
                    defaults,
                    undefined,
                    ScopeType.Project
                );

                // Create a workload that explicitly overrides the policy defaults
                const workloadSpec = {
                    workers: 2, // Override default
                    replicas: 2, // Override default
                    servingPort: {
                        port: 8080
                    },
                    leader: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        compute: new ComputeAssetSpec()
                    },
                    worker: {
                        image: config.quickStartImage.inferenceAutoScale,
                        compute: new ComputeAssetSpec()
                    }
                };

                // Set compute explicitly to override policy defaults
                workloadSpec.leader.compute.cpuCoreRequest = 1.0; // Override policy default of 0.5
                workloadSpec.worker.compute.cpuCoreRequest = 0.8; // Override policy default of 0.3

                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    workloadSpec,
                    project
                );

                await waitUtils.waitForBackendWorkloadPhaseInHistory(
                    workloadId,
                    workloadName,
                    Phase.Pending,
                    120
                );

                // Validate that workload uses override values, not policy defaults
                const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);

                expect(workload.body.workloadRequestedResources.cpu.request).to.be.closeTo(
                    2 * (2 * 0.8 + 1),
                    0.01
                );
                expect(workload.body.workloadRequestedResources.cpuMemory.request).to.not.be.null;

                logger.info(
                    'Successfully validated that workload overrides take precedence over policy defaults'
                );
            });
        }
    );
});
