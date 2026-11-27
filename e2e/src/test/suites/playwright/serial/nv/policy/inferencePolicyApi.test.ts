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
    policyUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { ScopeType } from '../../../../../../modules/api/apis/policyApi';
import { WorkloadType } from '../../../../../../modules/helpers/utils/policyUtils';
import { StatusCodes } from 'http-status-codes';
import { PolicyDefaultsType } from '../../../../../../modules/helpers/config/types';

test.describe(`Inference Workload Policy API Tests, tags: ${tags.tool.api} ${tags.resource.policy} ${tags.teams.loong} ${tags.teams.viper}`, () => {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let projectName: string;
        let namespace: string;
        let project: Project;
        let projectScopedCrdPolicyName: string;
        let policyCreated = false;
        let policyDeleted = false;

        const defaults = {
            command: 'Ubuntu',
            args: '-x my-script.py',
            image: 'python:3.8',
            imagePullPolicy: 'Always',
            workingDir: '/home/myfolder',
            createHomeDir: true,
            compute: {
                gpuDevicesRequest: 1,
                cpuCoreRequest: 0.5,
                cpuCoreLimit: 2,
                cpuMemoryRequest: '20M',
                cpuMemoryLimit: '30M',
                largeShmRequest: false
            }
        };

        const rules = {
            command: {
                required: true,
                canEdit: true
            },
            args: {
                required: true,
                canEdit: true
            },
            image: {
                required: true,
                canEdit: true
            },
            compute: {
                cpuCoreRequest: {
                    required: true,
                    canEdit: true,
                    min: 0,
                    max: 100,
                    step: 0
                },
                cpuCoreLimit: {
                    required: true,
                    canEdit: true,
                    min: 0,
                    max: 100,
                    step: 0
                },
                largeShmRequest: {
                    required: true,
                    canEdit: true
                },
                gpuDevicesRequest: {
                    required: true,
                    canEdit: true,
                    min: 1,
                    max: 100,
                    step: 1
                }
            }
        };

        const patchedDefaults = {
            image: 'python:3.9'
        };

        const patchedRules = {
            compute: {
                gpuDevicesRequest: {
                    required: true,
                    canEdit: true,
                    min: 1,
                    max: 200,
                    step: 2
                }
            }
        };

        const updatedDefaults = {
            ...defaults,
            image: patchedDefaults.image
        };

        const updatedRules = {
            ...rules,
            compute: {
                ...rules.compute,
                gpuDevicesRequest: patchedRules.compute.gpuDevicesRequest
            }
        };

        test.beforeAll(async function () {
            projectName = objectGenerator.generateShortNameByPrefix('proj-exclude-nodes', 5);
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 0,
                maxAllowedGpus: 0
            });

            projectScopedCrdPolicyName = policyUtils.getPolicyName(ScopeType.Project, project.id.toString());
        });

        test.afterAll('deleting project and namespace', async () => {
            if (policyCreated && !policyDeleted) {
                try {
                    await policyUtils.deleteWorkloadPolicyV2(WorkloadType.Inference, {
                        scope: ScopeType.Project.toString(),
                        scopeId: project.id.toString()
                    });
                } catch (error) {
                    logger.error(`failed to delete policy:\n ${error.message}`);
                }
            }

            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName} and namespace ${namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${projectName} and namespace ${namespace}`, error);
            }
        });

        test('T5169495: Inference CRUD tests', async function () {
            await test.step('Verify there is no inference policy in project before testing', async function () {
                await policyUtils.validateNoWorkloadPolicy(WorkloadType.Inference, {
                    scope: ScopeType.Project.toString(),
                    scopeId: project.id.toString()
                });
            });

            await test.step('Create a project scoped inference policy', async function () {
                const response = await policyUtils.createPolicyV2(
                    project.name,
                    project.id.toString(),
                    WorkloadType.Inference,
                    defaults,
                    rules,
                    ScopeType.Project
                );
                expect(response.response.statusCode).toEqual(StatusCodes.OK);
                policyCreated = true;

                await policyUtils.waitForPolicyToBeSynced(
                    WorkloadType.Inference,
                    namespace,
                    projectScopedCrdPolicyName
                );
            });

            await test.step('Verify the created inference policy', async function () {
                await policyUtils.validatePolicyBody(
                    WorkloadType.Inference,
                    {
                        scope: ScopeType.Project.toString(),
                        scopeId: project.id.toString()
                    },
                    defaults as PolicyDefaultsType,
                    rules
                );
            });

            await test.step('Update the inference policy', async function () {
                const response = await policyUtils.updatePolicyV2(
                    project.name,
                    project.id.toString(),
                    WorkloadType.Inference,
                    updatedDefaults,
                    updatedRules,
                    ScopeType.Project
                );
                expect(response.response.statusCode).toEqual(StatusCodes.OK);

                await policyUtils.waitForPolicyToBeSynced(
                    WorkloadType.Inference,
                    namespace,
                    projectScopedCrdPolicyName
                );
            });

            await test.step('Verify the updated inference policy', async function () {
                await policyUtils.validatePolicyBody(
                    WorkloadType.Inference,
                    {
                        scope: ScopeType.Project.toString(),
                        scopeId: project.id.toString()
                    },
                    updatedDefaults as PolicyDefaultsType,
                    updatedRules
                );
            });

            await test.step('Delete the inference policy', async function () {
                const statusCode = await policyUtils.deleteWorkloadPolicyV2(WorkloadType.Inference, {
                    scope: ScopeType.Project.toString(),
                    scopeId: project.id.toString()
                });
                expect(statusCode).toEqual(StatusCodes.NO_CONTENT);
                policyDeleted = true;
            });

            await test.step('Verify there is no inference policy in project after testing', async function () {
                await policyUtils.validateNoWorkloadPolicy(WorkloadType.Inference, {
                    scope: ScopeType.Project.toString(),
                    scopeId: project.id.toString()
                });
            });
        });
    });
});
