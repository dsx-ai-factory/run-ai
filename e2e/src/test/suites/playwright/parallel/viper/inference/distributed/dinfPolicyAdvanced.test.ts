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
import { validateGetDistributedInference } from './dinfUtils';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/policyUtils';
import { ScopeType, DistributedInferencePolicyV2 } from '../../../../../../../modules/api/apis/policyApi';
import { StatusCodes } from 'http-status-codes';
import { ComputeAssetSpec } from '../../../../../../../modules/api/apis/assetsApi';
import { DistributedInferenceSpecSpec } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import { DistributedInferenceLeaderWorkerSpecCompute } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceLeaderWorkerSpecCompute';

supportedFrom(SupportedVersions.V2_23)(function () {
    test.describe(
        `distributed inference workload policy advanced tests: ${tags.resource.distInferenceWorkload} ${tags.tool.api} ${tags.teams.viper}`,
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

            test('should create distributed inference policy with rules and validate enforcement', async function () {
                const defaults = {
                    leader: {
                        image: config.quickStartImage.inferenceHelloWorld
                    },
                    worker: {
                        image: config.quickStartImage.inferenceHelloWorld
                    }
                };

                const rules = {
                    workers: {
                        min: 1,
                        max: 5
                    },
                    replicas: {
                        min: 1,
                        max: 3
                    },
                    leader: {
                        compute: {
                            gpuDevicesRequest: {
                                max: 2
                            },
                            cpuCoreRequest: {
                                min: 0.1,
                                max: 2.0
                            }
                        }
                    },
                    worker: {
                        compute: {
                            cpuCoreRequest: {
                                min: 0.1,
                                max: 1.0
                            }
                        }
                    }
                };

                logger.info(`Creating distributed inference policy with rules: ${JSON.stringify(rules)}`);

                // Create the policy with rules
                const policyResponse = await policyUtils.createPolicyV2(
                    policyName,
                    project.id.toString(),
                    WorkloadType.DistributedInference,
                    defaults,
                    rules,
                    ScopeType.Project
                );

                expect(policyResponse.response.statusCode).to.equal(StatusCodes.OK);
                const effectiveRules = (policyResponse.body as DistributedInferencePolicyV2).effective.rules;
                expect(effectiveRules.workers.min).to.equal(rules.workers.min);
                expect(effectiveRules.workers.max).to.equal(rules.workers.max);
                expect(effectiveRules.replicas.min).to.equal(rules.replicas.min);
                expect(effectiveRules.replicas.max).to.equal(rules.replicas.max);

                logger.info('Successfully created distributed inference policy with rules');

                // Test workload that should be allowed by rules
                const validWorkloadSpec: DistributedInferenceSpecSpec = {
                    workers: 2, // Within 1-5 range
                    replicas: 2, // Within 1-3 range
                    servingPort: {
                        port: 8080
                    },
                    leader: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        compute: new DistributedInferenceLeaderWorkerSpecCompute()
                    },
                    worker: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        compute: new DistributedInferenceLeaderWorkerSpecCompute()
                    }
                };

                validWorkloadSpec.leader.compute.gpuDevicesRequest = 1; // Within max 2
                validWorkloadSpec.leader.compute.cpuCoreRequest = 0.5; // Within 0.1-2.0
                validWorkloadSpec.worker.compute.cpuCoreRequest = 0.5; // Within 0.1-1.0

                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    validWorkloadSpec,
                    project
                );

                const workload = await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Running,
                    120,
                    6
                );
                validateGetDistributedInference(workload, validWorkloadSpec);

                logger.info('Successfully validated that valid workload was accepted by policy rules');
            });

            test('should apply policy defaults for compute resources', async function () {
                const defaults = {
                    workers: 1,
                    replicas: 1,
                    leader: {
                        compute: {
                            gpuDevicesRequest: 1,
                            cpuCoreRequest: 1.0,
                            cpuMemoryRequest: '1Gi'
                        }
                    },
                    worker: {
                        compute: {
                            cpuCoreRequest: 0.5,
                            cpuMemoryRequest: '512Mi'
                        }
                    }
                };

                logger.info(`Creating policy with compute resource defaults: ${JSON.stringify(defaults)}`);

                // Create the policy
                await policyUtils.createPolicyV2(
                    policyName,
                    project.id.toString(),
                    WorkloadType.DistributedInference,
                    defaults,
                    undefined,
                    ScopeType.Project
                );

                // Create workload with minimal compute specs - should inherit policy compute defaults
                const workloadSpec = {
                    servingPort: {
                        port: 8080
                    },
                    leader: {
                        image: config.quickStartImage.inferenceHelloWorld
                        // compute should come from policy
                    },
                    worker: {
                        image: config.quickStartImage.inferenceAutoScale
                        // compute should come from policy
                    }
                };

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

                // Validate compute resources were applied from policy
                const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);

                // Check GPU allocation matches policy defaults (only leader has GPU)
                const expectedGpuRequest = defaults.leader.compute.gpuDevicesRequest * defaults.replicas;

                expect(workload.body.workloadRequestedResources?.gpu?.request).to.be.closeTo(
                    expectedGpuRequest,
                    0.01
                );

                logger.info('Successfully validated that compute resource defaults were applied from policy');
            });

            test('should validate policy with invalid configurations', async function () {
                const invalidDefaults = {
                    leader: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        compute: {
                            gpuDevicesRequest: -1 // Invalid negative value
                        }
                    },
                    workers: -1, // Invalid negative workers
                    replicas: 0 // Invalid zero replicas
                };

                logger.info('Testing policy validation with invalid configurations');

                try {
                    // This should fail validation
                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        WorkloadType.DistributedInference,
                        invalidDefaults,
                        undefined,
                        ScopeType.Project,
                        true // validation only
                    );

                    // If we reach here, validation didn't catch the error
                    throw new Error('Expected policy validation to fail with invalid configurations');
                } catch (error) {
                    // This is expected - policy validation should reject invalid configs
                    expect(error.statusCode).to.equal(400);
                    logger.info('Successfully validated that invalid policy configurations are rejected');
                }
            });

            test('should create workload with policy overrides', async function () {
                const defaults = {
                    leader: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        compute: {
                            cpuCoreRequest: 0.5
                        }
                    },
                    worker: {
                        image: config.quickStartImage.inferenceHelloWorld,
                        compute: {
                            cpuCoreRequest: 0.3
                        }
                    },
                    workers: 1,
                    replicas: 1
                };

                logger.info('Creating policy with defaults that can be overridden');

                // Create the policy
                await policyUtils.createPolicyV2(
                    policyName,
                    project.id.toString(),
                    WorkloadType.DistributedInference,
                    defaults,
                    undefined,
                    ScopeType.Project
                );

                // Create workload that overrides some policy defaults
                const overrideWorkloadSpec = {
                    workers: 2, // Override policy default of 1
                    replicas: 2, // Override policy default of 1
                    servingPort: {
                        port: 8080
                    },
                    leader: {
                        image: config.quickStartImage.inferenceAutoScale, // Override policy default image
                        compute: new ComputeAssetSpec()
                    },
                    worker: {
                        image: config.quickStartImage.inferenceAutoScale, // Override policy default image
                        compute: new ComputeAssetSpec()
                    }
                };

                overrideWorkloadSpec.leader.compute.cpuCoreRequest = 1.0; // Override policy default
                overrideWorkloadSpec.worker.compute.cpuCoreRequest = 0.5; // Override policy default

                [workloadId, workloadName] = await createDistributedInference(
                    workloadName,
                    overrideWorkloadSpec,
                    project
                );

                await waitUtils.waitForBackendWorkloadPhaseInHistory(
                    workloadId,
                    workloadName,
                    Phase.Pending,
                    120
                );

                // Validate that workload used override values, not policy defaults
                const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);

                const distributedInference = await workloadsCrudUtils.getDistributedInference(workloadId);
                expect(distributedInference.body.spec?.workers).to.equal(overrideWorkloadSpec.workers);
                expect(distributedInference.body.spec?.replicas).to.equal(overrideWorkloadSpec.replicas);
                expect(workload.body.images).to.include(config.quickStartImage.inferenceAutoScale);
                expect(workload.body.images).to.not.include(config.quickStartImage.inferenceHelloWorld);

                logger.info(
                    'Successfully validated that workload overrides take precedence over policy defaults'
                );
            });
        }
    );
});
