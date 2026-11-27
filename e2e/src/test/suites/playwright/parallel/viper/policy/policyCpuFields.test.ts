/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { policyApi as policyApiHandler } from '../../../../../../modules/api';

import {
    config,
    crds,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    policyUtils,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { assert, expect } from 'chai';
import { WorkloadType } from '../../../../../../modules/helpers/utils/policyUtils';
import { test, TestInfo } from '@playwright/test';
import {
    TrainingSpecSpec,
    WorkspaceSpecSpec
} from '../../../../../../modules/api/apis/workloadsApi/generated/model/models';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

const WorkloadTypeParams = [
    {
        policyType: 'training',
        image: config.quickStartImage.training,
        jobKind: crds.runaiTrainingWorkloadCrd.kind,
        crd: crds.runaiTrainingWorkloadCrd,
        kind: WorkloadType.Training,
        isInteractive: false
    },
    {
        policyType: 'interactive',
        image: config.quickStartImage.interactive,
        jobKind: crds.runaiInteractiveWorkloadCrd.kind,
        crd: crds.runaiInteractiveWorkloadCrd,
        kind: WorkloadType.Interactive,
        isInteractive: true
    }
];
WorkloadTypeParams.forEach(function (params) {
    test.describe(
        `${params.policyType} policy - sanity checks of cpu & cpu-limit, tags: ${tags.resource.policy} ${tags.tool.api} ${tags.teams.viper}`,
        { tag: [tags.services.policy, tags.runOn.parallelCI] },
        function () {
            let projectName: string;
            let namespace: string;
            let project: Project;
            let jobName: string;
            let policyName: string;
            let workloadId: string;
            let anyWorkloadSubmissionEnabled: boolean;

            async function _createPolicy(defaults: object, rules?: object): Promise<void> {
                await policyUtils.createPolicy(
                    params.isInteractive,
                    policyName,
                    namespace,
                    project.id,
                    params.kind,
                    defaults,
                    rules
                );
            }

            async function _updatePolicy(defaults: object, rules?: object): Promise<void> {
                await policyUtils.updateProjectScopedPolicy(
                    params.isInteractive,
                    policyName,
                    namespace,
                    project.id,
                    params.kind,
                    defaults,
                    rules
                );
            }

            async function _submitWorkload(cpu?: number, cpuLimit?: number): Promise<string> {
                const spec: any = {
                    image: params.image,
                    command: 'sleep infinity'
                };

                // CPU values go in the compute object, not at the root level
                if (cpu !== undefined || cpuLimit !== undefined) {
                    spec.compute = {};
                    if (cpu !== undefined) {
                        spec.compute.cpuCoreRequest = cpu;
                    }
                    if (cpuLimit !== undefined) {
                        spec.compute.cpuCoreLimit = cpuLimit;
                    }
                }

                let createResult;
                if (params.isInteractive) {
                    createResult = await workloadsCrudUtils.createWorkspace(
                        jobName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        spec as WorkspaceSpecSpec
                    );
                } else {
                    createResult = await workloadsCrudUtils.createTraining(
                        jobName,
                        false,
                        flowRootState.cluster.uuid,
                        project.id.toString(),
                        spec as TrainingSpecSpec
                    );
                }

                expect(createResult.response.statusCode).to.equal(202);

                // Give the API a moment to process the workload
                await waitUtils.sleep(1000);

                return createResult.body.workloadId;
            }

            async function _getBatchJobCpuValue(
                resourceType: 'requests' | 'limits'
            ): Promise<number | undefined> {
                await waitUtils.waitForBatchJobToExist(jobName, namespace);
                const job = await kubeCli.getBatchJob(jobName, namespace);
                const quantity = job.spec?.template.spec?.containers?.[0].resources?.[resourceType]?.cpu;
                if (quantity === undefined) {
                    return undefined;
                }

                const stringQuantity = String(quantity);
                return stringQuantity.endsWith('m')
                    ? Number.parseFloat(stringQuantity.slice(0, -1)) / 1000
                    : Number.parseFloat(stringQuantity);
            }

            async function _getCpuValue(): Promise<number | undefined> {
                if (anyWorkloadSubmissionEnabled) {
                    return _getBatchJobCpuValue('requests');
                }

                let getResult;
                if (params.isInteractive) {
                    getResult = await workloadsCrudUtils.getWorkspace(workloadId);
                } else {
                    getResult = await workloadsCrudUtils.getTraining(workloadId);
                }

                return getResult.body.spec?.compute?.cpuCoreRequest;
            }

            async function _getCpuLimitValue(): Promise<number | undefined> {
                if (anyWorkloadSubmissionEnabled) {
                    return _getBatchJobCpuValue('limits');
                }

                let getResult;
                if (params.isInteractive) {
                    getResult = await workloadsCrudUtils.getWorkspace(workloadId);
                } else {
                    getResult = await workloadsCrudUtils.getTraining(workloadId);
                }

                return getResult.body.spec?.compute?.cpuCoreLimit;
            }

            test.beforeAll(async function () {
                policyApiHandler.policyApi.accessToken = flowRootState.users.system_administrator.accessToken;
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
                anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
                    params.isInteractive
                        ? AnyWorkloadSubmissionType.Workspace
                        : AnyWorkloadSubmissionType.Training
                );
            });

            test.beforeEach(async function ({}, testInfo: TestInfo) {
                jobName = objectGenerator.generateShortNameByPrefix('job-for-policy');
                projectName = objectGenerator.generateShortNameByPrefix('project-for-policy');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 1,
                    maxAllowedGpus: 1
                });
                policyName = `project-${project.id}`;
            });

            test.afterEach(async ({}, testInfo) => {
                // Delete workload if it was created
                if (workloadId) {
                    try {
                        if (params.isInteractive) {
                            await workloadsCrudUtils.deleteWorkspace(workloadId);
                        } else {
                            await workloadsCrudUtils.deleteTraining(workloadId);
                        }
                    } catch {
                        logger.error(`failed to delete workload ${workloadId}`);
                    }
                    workloadId = undefined;
                }

                // Delete policy
                if (testInfo.status !== 'skipped') {
                    try {
                        if (params.isInteractive) {
                            await policyApiHandler.policyApi.deleteWorkspacePolicy(
                                'project',
                                undefined,
                                project.id.toString()
                            );
                        } else {
                            await policyApiHandler.policyApi.deleteTrainingPolicy(
                                'project',
                                undefined,
                                project.id.toString()
                            );
                        }
                    } catch {
                        logger.error(`failed to delete policy for project ${project.id}`);
                    }
                }

                // Delete project
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch (error) {
                    logger.error(`failed to delete project ${projectName} an namespace ${namespace}`);
                    throw error;
                }
            });

            test.describe('default', function () {
                test.describe('cpu', function () {
                    test('validate cpu default value', async function () {
                        const cpuCoreRequest = 0.1;
                        await _createPolicy({ compute: { cpuCoreRequest: cpuCoreRequest } });
                        workloadId = await _submitWorkload();
                        const cpusCount = await _getCpuValue();
                        expect(cpusCount).to.equal(cpuCoreRequest);
                    });

                    test('override cpu default value', async function () {
                        const numOfCpus = 1;
                        await _createPolicy({ compute: { cpuCoreRequest: 0 } });
                        workloadId = await _submitWorkload(numOfCpus);
                        const cpusCount = await _getCpuValue();
                        expect(cpusCount).to.equal(numOfCpus);
                    });
                });
                test('update cpu default value', async function () {
                    const cpuCoreRequest = 0.1;
                    await _createPolicy({ compute: { cpuCoreRequest: 0.2 } });
                    await _updatePolicy({ compute: { cpuCoreRequest: cpuCoreRequest } });
                    // With API submission, policy is applied immediately from database
                    workloadId = await _submitWorkload();
                    const cpusCount = await _getCpuValue();
                    expect(cpusCount).to.equal(cpuCoreRequest);
                });

                test.describe('cpu-limit', function () {
                    test('validate cpu-limit default value', async function () {
                        const cpuCoreLimit = 0.1;
                        await _createPolicy({ compute: { cpuCoreLimit: cpuCoreLimit } });
                        workloadId = await _submitWorkload();
                        const cpusLimitCount = await _getCpuLimitValue();
                        expect(cpusLimitCount).to.equal(cpuCoreLimit);
                    });
                    test('override cpu-limit default value', async function () {
                        const cpuCoreLimit = 0.1;
                        await _createPolicy({ compute: { cpuCoreLimit: 1 } });
                        workloadId = await _submitWorkload(undefined, cpuCoreLimit);
                        const cpusLimitCount = await _getCpuLimitValue();
                        expect(cpusLimitCount).to.equal(cpuCoreLimit);
                    });
                });
            });
            test.describe('rules', function () {
                let rules: object;
                let MinCpuRequest: number;
                let MaxCpuRequest: number;

                test.describe('cpu', function () {
                    test.beforeEach(function () {
                        MinCpuRequest = 0.1;
                        MaxCpuRequest = 3;
                        rules = {
                            compute: {
                                cpuCoreRequest: {
                                    min: MinCpuRequest,
                                    max: MaxCpuRequest
                                }
                            }
                        };
                    });
                    test('can submit a job without cpu, when its not required', async function () {
                        await _createPolicy(undefined, rules);
                        workloadId = await _submitWorkload();
                        // Should succeed without specifying CPU
                    });

                    test('cannot submit a job without cpu, when its required', async function () {
                        rules['compute']['cpuCoreRequest']['required'] = true;
                        await _createPolicy(undefined, rules);
                        try {
                            workloadId = await _submitWorkload();
                            assert.fail('Expected submission to fail due to required CPU');
                        } catch (error) {
                            // API should return error for missing required field
                            expect(error.message).to.match(/cpu|required/i);
                        }
                    });
                    test('update cpu min value', async function () {
                        const newMinOfCpus = 0.3;
                        const cpuForWorkload = 0.2;
                        await _createPolicy(undefined, rules);

                        // Delete the old policy first (update might fail due to effective policy validation)
                        if (params.isInteractive) {
                            await policyApiHandler.policyApi.deleteWorkspacePolicy(
                                'project',
                                undefined,
                                project.id.toString()
                            );
                        } else {
                            await policyApiHandler.policyApi.deleteTrainingPolicy(
                                'project',
                                undefined,
                                project.id.toString()
                            );
                        }

                        // Create new policy with updated rules and defaults that satisfy the rules
                        // (effective policy validation requires defaults to satisfy rules)
                        rules['compute'].cpuCoreRequest.min = newMinOfCpus;
                        await _createPolicy({ compute: { cpuCoreRequest: newMinOfCpus } }, rules);

                        // With API submission, policy is applied immediately from database
                        try {
                            workloadId = await _submitWorkload(cpuForWorkload);
                            assert.fail('Expected submission to fail due to CPU below minimum');
                        } catch (error) {
                            // API should return error for CPU value below minimum
                            expect(error.message).to.match(/cpu|less|min/i);
                        }
                    });

                    test('can submit a job with min-cpu < #cpu < max-cpu', async function () {
                        const numOfCpus = MinCpuRequest + 0.1;
                        await _createPolicy(undefined, rules);
                        workloadId = await _submitWorkload(numOfCpus);
                        const cpusCount = await _getCpuValue();
                        expect(cpusCount).to.equal(numOfCpus);
                    });

                    test('cannot submit a job with #cpu < min-cpu', async function () {
                        await _createPolicy(undefined, rules);
                        try {
                            workloadId = await _submitWorkload(0);
                            assert.fail('Expected submission to fail due to CPU below minimum');
                        } catch (error) {
                            expect(error.message).to.match(/cpu|less|min/i);
                        }
                    });

                    test('cannot submit a job with max-cpu < #cpu', async function () {
                        await _createPolicy(undefined, rules);
                        try {
                            workloadId = await _submitWorkload(9);
                            assert.fail('Expected submission to fail due to CPU above maximum');
                        } catch (error) {
                            expect(error.message).to.match(/cpu|greater|max/i);
                        }
                    });
                });

                test.describe('cpu-limit', function () {
                    const MinCpuLimit = 0.2;
                    const MaxCpuLimit = 4;
                    test.beforeEach(function () {
                        rules = {
                            compute: {
                                cpuCoreLimit: {
                                    min: MinCpuLimit,
                                    max: MaxCpuLimit
                                }
                            }
                        };
                    });

                    test('can submit a job with min-cpu-limit < #cpu-limit < max-cpu-limit', async function () {
                        const cpuLimit = MinCpuLimit + 0.05;
                        await _createPolicy(undefined, rules);
                        workloadId = await _submitWorkload(undefined, cpuLimit);
                        const cpusLimitCount = await _getCpuLimitValue();
                        expect(cpusLimitCount).to.equal(cpuLimit);
                    });

                    test('cannot submit a job with #cpu-limit < min-cpu-limit', async function () {
                        const cpuLimit = MinCpuLimit - 0.1;
                        await _createPolicy(undefined, rules);
                        try {
                            workloadId = await _submitWorkload(undefined, cpuLimit);
                            assert.fail('Expected submission to fail due to CPU limit below minimum');
                        } catch (error) {
                            expect(error.message).to.match(/cpuLimit|less|min/i);
                        }
                    });

                    test('cannot submit a job with max-cpu-limit < #cpu-limit', async function () {
                        const cpuLimit = MaxCpuLimit + 1;
                        await _createPolicy(undefined, rules);
                        try {
                            workloadId = await _submitWorkload(undefined, cpuLimit);
                            assert.fail('Expected submission to fail due to CPU limit above maximum');
                        } catch (error) {
                            expect(error.message).to.match(/cpuLimit|greater|max/i);
                        }
                    });
                });
            });
        }
    );
});
