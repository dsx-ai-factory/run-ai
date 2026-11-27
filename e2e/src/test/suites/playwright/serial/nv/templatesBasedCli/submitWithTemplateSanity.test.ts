/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    config,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState,
    templateUtils,
    waitUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType, WorkloadType } from '../../../../../../modules/api/apis/workloadsApi/generated/api';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliV2WorkloadType,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

// Test parameters for different workload types
const WorkloadTypeParams = [
    {
        testName: 'Training',
        workloadType: WorkloadType.Training,
        templateName: 'test-pytorch-training',
        cliWorkloadType: cliV2WorkloadType.training,
        gpuDevicesRequest: 0
    },
    {
        testName: 'PyTorch Distributed',
        workloadType: WorkloadType.Distributed,
        templateName: 'test-pytorch-distributed',
        cliWorkloadType: cliV2WorkloadType.pyTorch,
        gpuDevicesRequest: 0,
        distributedFramework: 'PyTorch',
        numWorkers: 2
    },
    {
        testName: 'MPI Distributed',
        workloadType: WorkloadType.Distributed,
        templateName: 'test-mpi-distributed',
        cliWorkloadType: cliV2WorkloadType.mpi,
        gpuDevicesRequest: 0,
        distributedFramework: 'MPI',
        numWorkers: 2
    },
    {
        testName: 'TensorFlow Distributed',
        workloadType: WorkloadType.Distributed,
        templateName: 'test-tf-distributed',
        cliWorkloadType: cliV2WorkloadType.tensorFlow,
        gpuDevicesRequest: 0,
        distributedFramework: 'TF',
        numWorkers: 2
    },
    {
        testName: 'XGBoost Distributed',
        workloadType: WorkloadType.Distributed,
        templateName: 'test-xgboost-distributed',
        cliWorkloadType: cliV2WorkloadType.xgBoost,
        gpuDevicesRequest: 0,
        distributedFramework: 'XGBoost',
        numWorkers: 2
    },
    {
        testName: 'Workspace',
        workloadType: WorkloadType.Workspace,
        templateName: 'test-jupyter-workspace',
        cliWorkloadType: cliV2WorkloadType.workspace,
        gpuDevicesRequest: 0
    },
    {
        testName: 'Inference',
        workloadType: WorkloadType.Inference,
        templateName: 'test-inference-single',
        cliWorkloadType: cliV2WorkloadType.inference,
        gpuDevicesRequest: 0,
        isInference: true
    },
    {
        testName: 'Distributed Inference',
        workloadType: WorkloadType.DistributedInference,
        templateName: 'test-inference-distributed',
        cliWorkloadType: cliV2WorkloadType.inference,
        gpuDevicesRequest: 0,
        isInference: true,
        isDistributedInference: true,
        numWorkers: 1
    }
];

test.describe(`CLI Submit With Template - Sanity Tests, tags: ${tags.resource.templateV2} ${tags.teams.viper}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-template-sanity', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            logger.info(`Created project: ${projectName} (ID: ${project.id})`);

            // Set project context once for all tests
            runaiCliV2.projectSet(projectName);
            logger.info(`Set project context to: ${projectName}`);
        });

        test.afterAll('Cleanup resources', async () => {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`Deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`Failed to delete project ${projectName}:`, error);
            }
        });

        // T5539372: Parametrized test for all workload types
        WorkloadTypeParams.forEach(function (params) {
            test(`T5539372: Submit ${params.testName} workload with valid template`, async function () {
                const workloadName = objectGenerator.generateShortNameByPrefix(
                    `test-${params.testName.toLowerCase().replace(/\s+/g, '-')}`,
                    5
                );
                let templateId: string;
                let templateName: string;

                await test.step(`Create ${params.testName} template`, async function () {
                    templateName = objectGenerator.generateShortNameByPrefix(params.templateName, 5);

                    const cpuOnly = {
                        gpuDevicesRequest: 0,
                        cpuCoreRequest: 0.1,
                        cpuMemoryRequest: '100M'
                    };
                    const oneGpu = {
                        gpuDevicesRequest: 1,
                        cpuCoreRequest: 0.1,
                        cpuMemoryRequest: '100M'
                    };

                    const workloadImage = params.isInference
                        ? config.quickStartImage.inferenceHelloWorld
                        : 'ubuntu';
                    const compute = params.gpuDevicesRequest > 0 ? oneGpu : cpuOnly;

                    let spec: any;
                    if (params.isDistributedInference) {
                        // Distributed inference uses leader/worker replicas, not the flat spec.
                        spec = {
                            workers: params.numWorkers || 1,
                            replicas: 1,
                            servingPort: { port: 8080 },
                            leader: { image: workloadImage, imagePullPolicy: 'IfNotPresent', compute },
                            worker: { image: workloadImage, imagePullPolicy: 'IfNotPresent', compute }
                        };
                    } else {
                        spec = {
                            image: workloadImage,
                            imagePullPolicy: 'IfNotPresent',
                            compute
                        };

                        if (params.isInference) {
                            spec.servingPort = { port: 8080, protocol: 'http' };
                        } else {
                            spec.command = 'sleep 100';
                        }

                        if (params.distributedFramework) {
                            spec.distributedFramework = params.distributedFramework;
                            spec.numWorkers = params.numWorkers || 2;
                        }
                    }

                    const response = await templateUtils.createWorkloadTemplateV2(
                        {
                            workloadType: params.workloadType,
                            scopeType: ScopeType.Cluster,
                            scopeId: flowRootState.cluster.uuid,
                            templateName: templateName,
                            description: `Template for ${params.testName} testing`
                        },
                        spec,
                        params.distributedFramework ? { masterSpecSameAsWorker: true } : undefined,
                        flowRootState.users.system_administrator.accessToken
                    );

                    expect(response.response.statusCode).toEqual(200);
                    templateId = response.body.meta.id;
                    logger.info(`Created template: ${templateName} (ID: ${templateId})`);
                });

                await test.step(`Submit ${params.testName} workload using template via CLI`, async function () {
                    const submitFlag = {
                        [SubmitFlagsMerged.template]: templateName
                    } as Partial<typeof SubmitFlagsMerged>;

                    if (params.isInference) {
                        submitFlag[SubmitFlagsMerged.image] = 'ubuntu:latest';
                        submitFlag[SubmitFlagsMerged.servingPort] = '8080';
                    }

                    logger.info(
                        `Submitting ${params.testName} workload: ${workloadName} with template: ${templateName}`
                    );

                    const stdout = runaiCliV2.submit(
                        submitFlag,
                        workloadName,
                        params.cliWorkloadType,
                        undefined,
                        params.isDistributedInference
                    );
                    logger.info(`CLI output: ${stdout}`);

                    expect(stdout).not.toContain('Error');
                    expect(stdout).not.toContain('Failed');

                    logger.info(`Waiting for ${params.testName} workload pods to be running...`);

                    await waitUtils.waitForPods(workloadName, namespace, 1, 120, ['Running'], true);
                    logger.info(`${params.testName} workload pods are running`);
                });

                await test.step(`Verify ${params.testName} workload exists`, function () {
                    logger.info(`Describing ${params.testName} workload: ${workloadName}`);

                    const stdout = runaiCliV2.describe(
                        {},
                        workloadName,
                        params.cliWorkloadType,
                        params.isDistributedInference
                    );
                    expect(stdout).toContain(workloadName);
                });

                await test.step(`Cleanup: Delete ${params.testName} template and workload`, async function () {
                    try {
                        runaiCliV2.delete(
                            {},
                            workloadName,
                            params.cliWorkloadType,
                            true,
                            params.isDistributedInference
                        );
                        logger.info(`Successfully deleted ${params.testName} workload: ${workloadName}`);

                        await templateUtils.deleteWorkloadTemplateV2(
                            templateId,
                            params.workloadType,
                            flowRootState.users.system_administrator.accessToken
                        );
                        logger.info(`Deleted template: ${templateId}`);
                    } catch (error) {
                        logger.error(`Failed to cleanup ${params.testName} resources:`, error);
                    }
                });
            });
        });
    });
});
