/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    flowRootState,
    dataSourcesUtils,
    waitUtils
} from '../../../../../../modules/helpers';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Scope } from '../../../../../../modules/api/apis/assetsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    cliV2WorkloadType,
    cliV2TrainingType,
    cliLoginAsSystemAdmin
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';

test.describe(`CLI Submit with Datasource - Workload Types, tags: ${tags.resource.datasource} ${tags.teams.loong}`, () => {
    supportedFrom(SupportedVersions.V2_24)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-ds-submit', 5);
            namespace = Utils.namespaceFromProjectName(projectName);

            await test.step('Create project via API', async function () {
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                });
                logger.info(`Created project: ${projectName} (ID: ${project.id})`);
            });

            await test.step('Ensure CLI is logged in', function () {
                cliLoginAsSystemAdmin();
                logger.info('Ensured CLI is logged in as system admin');
            });

            await test.step('Set project context for CLI', function () {
                runaiCliV2.projectSet(projectName);
                logger.info(`Set CLI project context to: ${projectName}`);
            });
        });

        test.afterAll(async function () {
            await test.step('Cleanup project', async function () {
                if (project?.id) {
                    await operations.deleteProjectAndNamespace(project.id);
                    logger.info(`Deleted project: ${projectName}`);
                }
            });
        });

        // Helper function: Create PVC datasource
        async function createPvcDatasource(dsName: string): Promise<string> {
            const pvcAsset = await dataSourcesUtils.createPVC(dsName, Scope.Project, project);
            logger.info(`Created PVC datasource: ${dsName} (ID: ${pvcAsset.meta.id})`);
            return pvcAsset.meta.id;
        }

        // Helper function: Cleanup workload and datasource
        async function cleanupWorkloadAndDatasource(
            jobName: string,
            workloadType: any,
            dataSourceId: string
        ): Promise<void> {
            try {
                runaiCliV2.delete({}, jobName, workloadType);
                await dataSourcesUtils.deletePvcAsset(
                    dataSourceId,
                    flowRootState.users.system_administrator.idToken
                );
                logger.info('Cleaned up workload and datasource');
            } catch (error) {
                logger.error('Cleanup error:', error);
            }
        }

        // Test configuration array
        const workloadTestCases = [
            {
                id: 'T5585456',
                name: 'pytorch training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.pyTorch,
                prefix: 'pytorch',
                image: 'pytorch/pytorch:latest',
                extraFlags: { '-g': '1' }
            },
            {
                id: 'T5585457',
                name: 'mpi training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.mpi,
                prefix: 'mpi',
                image: 'mpioperator/mpi-pi:latest',
                extraFlags: { '-g': '1', '--workers': '2' }
            },
            {
                id: 'T5585458',
                name: 'tensorflow training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.tensorFlow,
                prefix: 'tf',
                image: 'tensorflow/tensorflow:latest',
                extraFlags: { '-g': '1' }
            },
            {
                id: 'T5585459',
                name: 'xgboost training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.xgBoost,
                prefix: 'xgb-train',
                image: 'xgboost/xgboost:latest',
                extraFlags: { '--workers': '2' }
            },
            {
                id: 'T5585460',
                name: 'jax training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.jax,
                prefix: 'jax-train',
                image: 'google/jax:latest',
                extraFlags: { '-g': '1' }
            },
            {
                id: 'T5585461',
                name: 'standard training',
                type: cliV2WorkloadType.training,
                trainingType: cliV2TrainingType.standard,
                prefix: 'std-train',
                image: 'ubuntu:latest',
                extraFlags: { '-g': '1' }
            },
            {
                id: 'T5585462',
                name: 'standard inference',
                type: cliV2WorkloadType.inference,
                trainingType: null,
                prefix: 'inference',
                image: 'nginx:latest',
                extraFlags: { '-g': '1', '--serving-port': '8080' }
            },
            {
                id: 'T5585463',
                name: 'workspace',
                type: cliV2WorkloadType.workspace,
                trainingType: null,
                prefix: 'workspace',
                image: 'jupyter/base-notebook:latest',
                extraFlags: { '-g': '1' }
            },
            {
                id: 'T5585464',
                name: 'xgboost workload',
                type: cliV2WorkloadType.xgBoost,
                trainingType: null,
                prefix: 'xgboost',
                image: 'xgboost/xgboost:latest',
                extraFlags: { '--workers': '2' }
            },
            {
                id: 'T5585465',
                name: 'jax workload',
                type: cliV2WorkloadType.jax,
                trainingType: null,
                prefix: 'jax',
                image: 'google/jax:latest',
                extraFlags: { '-g': '1' }
            }
        ];

        // Generate test cases from configuration
        workloadTestCases.forEach((testCase) => {
            test(`${testCase.id}: Submit ${testCase.name} with datasource`, async function () {
                const dsName = objectGenerator.generateShortNameByPrefix(`ds-${testCase.prefix}`, 5);
                const jobName = objectGenerator.generateShortNameByPrefix(`job-${testCase.prefix}`, 5);
                let dataSourceId: string;

                await test.step('Create PVC datasource via API', async function () {
                    dataSourceId = await createPvcDatasource(dsName);
                });

                await test.step(`Submit ${testCase.name} with datasource`, function () {
                    logger.info(`Submitting ${testCase.name}: ${jobName}`);
                    const submitFlags = {
                        '--image': testCase.image,
                        '--datasource': [`type=pvc,name=${dsName}`],
                        ...testCase.extraFlags
                    } as any;

                    const stdout = testCase.trainingType
                        ? runaiCliV2.submit(submitFlags, jobName, testCase.type, testCase.trainingType)
                        : runaiCliV2.submit(submitFlags, jobName, testCase.type);

                    logger.info(`Submit output: ${stdout}`);
                    expect(stdout).toBeTruthy();
                    logger.info(`${testCase.name} submitted successfully`);
                });

                await test.step('Verify workload is created', async function () {
                    await waitUtils.waitForBackendWorkload(jobName, String(project.id), 60);
                    logger.info('Verified: Workload created successfully');
                });

                await test.step('Cleanup', async function () {
                    await cleanupWorkloadAndDatasource(jobName, testCase.type, dataSourceId);
                });
            });
        });
    });
});
