/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    config,
    externalWorkloadUtils,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { isKubeflowInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { PyTorchJob, TFJob, XGBoostJob } from '../../../../../../../modules/helpers/models/externalWorkload';
import { kubeCli } from '../../../../../../../modules/cli';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { test } from '@playwright/test';
import { BaseWorkload } from '../../../../../../../modules/helpers';

test.describe(
    `External workloads API tests, tags: , ${tags.teams.viper}, ${tags.resource.workloads}`,
    { tag: [tags.services.externalWorkloadIntegrator, tags.runOn.serialCI] },
    () => {
        let project: Project, namespace: string;

        test.beforeAll(async () => {
            const projectName = objectGenerator.generateNameByPrefix('external-workloads-project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 2,
                maxAllowedGpus: 2
            });
            if (!project.id) {
                throw new Error('Project ID is required for the test');
            }
        });

        test.afterAll(async () => {
            if (project.id) {
                await operations.deleteProjectAndNamespace(project.id);
            }
        });

        test.describe('External workloads that require Kubeflow', () => {
            const workloads: BaseWorkload[] = [];

            test.beforeAll(async () => {
                if (!(await isKubeflowInstalled())) {
                    test.skip();
                }
            });

            test.afterEach(async () => {
                // Delete all workloads by type, and remove them from the list
                while (workloads.length > 0) {
                    const workload = workloads.pop();
                    if (workload) {
                        if (workload instanceof TFJob) {
                            await kubeCli.deleteTfJob(workload.name, namespace);
                        } else if (workload instanceof XGBoostJob) {
                            await kubeCli.deleteXGBoostJob(workload.name, namespace);
                        } else if (workload instanceof PyTorchJob) {
                            await kubeCli.deletePyTorchJob(workload.name, namespace);
                        } else {
                            await kubeCli.deleteWorkload(workload);
                        }
                    }
                }
            });

            test('Should validate TFJob workload through kubectl apply', async () => {
                const tfJobName = objectGenerator.generateNameByPrefix('tf-workload');
                const resources = {
                    requests: { cpu: '100m', memory: '100Mi', 'nvidia.com/gpu': '1' },
                    limits: { cpu: '150m', memory: '150Mi', 'nvidia.com/gpu': '1' }
                };
                const tfExternalWorkload = new TFJob(
                    tfJobName,
                    namespace,
                    config.quickStartImage.ubuntu,
                    'sleep infinity',
                    1,
                    'OnFailure',
                    resources
                );

                workloads.push(tfExternalWorkload);
                const workloadId = await externalWorkloadUtils.submitWorkloadAndGetWorkloadId(
                    tfExternalWorkload,
                    project.id
                );

                const expectedType = 'TFJob';
                const expectedSource = 'Other';
                const expectedResources = {
                    gpu: { request: 1, limit: 1 },
                    cpu: { request: 0.1, limit: 0.15 },
                    cpuMemory: { request: '100Mi', limit: '150Mi' }
                };
                const expectedRequestedPods = 1;

                await workloadsUtilsV2.getWorkloadAndValidate(workloadId, {
                    expectedType: expectedType,
                    expectedSource: expectedSource,
                    expectedWorkloadRequestedResources: expectedResources,
                    expectedRequestedPods: expectedRequestedPods
                });
            });

            test('Should validate XGBoostJob workload through kubectl apply', async () => {
                const xgboostJobName = objectGenerator.generateNameByPrefix('xgboost-workload');
                const resourcesMaster = {
                    limits: { 'nvidia.com/gpu': '1', cpu: '200m' },
                    requests: { 'nvidia.com/gpu': '1', cpu: '100m' }
                };
                const xgboostExternalWorkload = new XGBoostJob(
                    xgboostJobName,
                    namespace,
                    config.quickStartImage.ubuntu,
                    'sleep infinity',
                    1,
                    2,
                    'OnFailure',
                    resourcesMaster
                );

                workloads.push(xgboostExternalWorkload);
                const workloadId = await externalWorkloadUtils.submitWorkloadAndGetWorkloadId(
                    xgboostExternalWorkload,
                    project.id
                );

                const expectedType = 'XGBoostJob';
                const expectedSource = 'Other';
                const expectedResources = {
                    gpu: { request: 1, limit: 1 },
                    cpu: { request: 0.1, limit: 0.2 }
                };
                const expectedRequestedPods = 3;

                await workloadsUtilsV2.getWorkloadAndValidate(workloadId, {
                    expectedType: expectedType,
                    expectedSource: expectedSource,
                    expectedWorkloadRequestedResources: expectedResources,
                    expectedRequestedPods: expectedRequestedPods
                });
            });

            test('Should validate PyTorchJob workload through kubectl apply', async () => {
                const pyTorchJobName = objectGenerator.generateNameByPrefix('pytorch-workload');
                const masterReplicas = 1;
                const masterGpuNumber = 1;
                const masterResources = {
                    limits: { 'nvidia.com/gpu': masterGpuNumber.toString(), cpu: '150m' },
                    requests: { 'nvidia.com/gpu': masterGpuNumber.toString(), cpu: '100m' }
                };
                const workerReplicas = 2;
                const workerGpuNumber = 1;
                const workerResources = {
                    limits: { 'nvidia.com/gpu': workerGpuNumber.toString() },
                    requests: { 'nvidia.com/gpu': workerGpuNumber.toString() }
                };
                const pyTorchExternalWorkload = new PyTorchJob(
                    pyTorchJobName,
                    namespace,
                    config.quickStartImage.ubuntu,
                    'sleep infinity',
                    masterReplicas,
                    workerReplicas,
                    'OnFailure',
                    masterResources,
                    workerResources
                );

                workloads.push(pyTorchExternalWorkload);
                const workloadId = await externalWorkloadUtils.submitWorkloadAndGetWorkloadId(
                    pyTorchExternalWorkload,
                    project.id
                );

                const expectedType = 'PyTorchJob';
                const expectedSource = 'Other';
                const expectedRequestedGpuResources =
                    masterReplicas * masterGpuNumber + workerReplicas * workerGpuNumber;
                const expectedRequestedCpuResources = masterReplicas * 0.1;
                const expectedLimitCpuResources = masterReplicas * 0.15;
                const expectedResources = {
                    gpu: { request: expectedRequestedGpuResources, limit: expectedRequestedGpuResources },
                    cpu: { request: expectedRequestedCpuResources, limit: expectedLimitCpuResources }
                };
                const expectedRequestedPods = masterReplicas + workerReplicas;

                await workloadsUtilsV2.getWorkloadAndValidate(workloadId, {
                    expectedType: expectedType,
                    expectedSource: expectedSource,
                    expectedWorkloadRequestedResources: expectedResources,
                    expectedRequestedPods: expectedRequestedPods
                });
            });
        });
    }
);
