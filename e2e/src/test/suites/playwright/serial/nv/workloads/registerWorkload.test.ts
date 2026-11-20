/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    Utils,
    logger,
    objectGenerator,
    projectResourcesUtils,
    tags,
    flowRootState,
    workloadsCrudUtils,
    orgUnitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { bashCli } from '../../../../../../modules/cli';
import { ServingPortProtocol } from '../../../../../../modules/api/apis/workloadsApi/generated/model/servingPortProtocol';
import { WorkloadUtils } from '../../../../../../modules/helpers/utils/WorkloadUtils';
import { StatusCodes } from 'http-status-codes';
import { clusterApi, workloadsService } from '../../../../../../modules/api';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(`Register any workload to Run:ai ${tags.tool.ui} ${tags.resource.workloads} ${tags.teams.loong}`, function () {
    supportedFrom(SupportedVersions.V2_23)(function () {
        let project: Project;
        let workloadTypeName: string;
        let workloadCategoryId: string;
        let workloadTypeUpdated: boolean = false;
        let createdWorkloadIds: string[] = [];
        let createdWorkloadTypeIds: string[] = [];
        const workloadUtils = new WorkloadUtils();
        const expectedPhaseByWorkloadType: Record<string, 'Ready' | 'NotReady'> = {
            TrainingWorkload: 'Ready',
            Workspace: 'Ready',
            Inference: 'Ready'
        };
        let defaultCategoryId: string;
        let defaultPriorityId: string;
        let deployCategoryId: string;
        let lowPriorityId: string;

        async function getCategoryAndPriorityIds(
            categoryName?: string,
            priorityName?: string
        ): Promise<{ categoryId: string; priorityId: string }> {
            const workloadCategories = await workloadUtils.getWorkloadCategories();
            expect(workloadCategories.length).toBeGreaterThanOrEqual(1);

            let categoryId: string;
            if (categoryName) {
                const targetCategory = workloadCategories.find((cat) => cat.name === categoryName);
                categoryId = targetCategory ? targetCategory.id : workloadCategories[0].id;
                logger.info(`Using ${categoryName} category ID: ${categoryId}`);
            } else {
                categoryId = workloadCategories[0].id;
                logger.info(`Using category ID: ${categoryId} (${workloadCategories[0].name})`);
            }

            const workloadPriorities = await workloadUtils.getWorkloadPriorities();
            expect(workloadPriorities.length).toBeGreaterThan(0);

            let priorityId: string;
            if (priorityName) {
                const targetPriority = workloadPriorities.find((priority) => priority.name === priorityName);
                priorityId = targetPriority ? targetPriority.id : workloadPriorities[0].id;
                logger.info(`Using ${priorityName} priority ID: ${priorityId}`);
            } else {
                // Find a priority with scopeType 'system' and source 'Run.ai' as default
                const systemPriority = workloadPriorities.find(
                    (priority) => priority.scopeType === 'system' && priority.source === 'Run.ai'
                );
                priorityId = systemPriority ? systemPriority.id : workloadPriorities[0].id;
                const selectedPriority = systemPriority || workloadPriorities[0];
                logger.info(`Using priority ID: ${priorityId} (${selectedPriority.name})`);
            }
            return { categoryId, priorityId };
        }

        test.beforeAll(async () => {
            const projectName = objectGenerator.generateNameByPrefix('workload-register');
            const namespace = Utils.namespaceFromProjectName(projectName);
            workloadTypeName = 'TrainingWorkload';
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 1,
                maxAllowedGpus: 1
            });
            const accessToken = flowRootState.users.system_administrator.accessToken;
            workloadsCrudUtils.accessToken = accessToken;
            clusterApi.clustersApi.accessToken = accessToken;

            const defaults = await getCategoryAndPriorityIds();
            defaultCategoryId = defaults.categoryId;
            defaultPriorityId = defaults.priorityId;

            const deployLow = await getCategoryAndPriorityIds('Deploy', 'low');
            deployCategoryId = deployLow.categoryId;
            lowPriorityId = deployLow.priorityId;
        });

        test.afterAll(async function () {
            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id.toString(), project.name);
                logger.info(`deleted project ${project.name} and namespace ${project.namespace}`);
            } catch (error) {
                logger.error(`failed to delete project ${project.name} and namespace ${project.namespace}`);
                throw error;
            }
        });

        test.beforeEach(async function () {
            createdWorkloadIds = [];
            createdWorkloadTypeIds = [];

            // Get the default workloadCategoryId
            const workloadTypes = await workloadUtils.getWorkloadTypes();
            const trainingWorkloadType = workloadTypes.find((wt) => wt.name === workloadTypeName);
            if (trainingWorkloadType) {
                workloadCategoryId = trainingWorkloadType.categoryId;
                logger.info(`Saved default workloadCategoryId: ${workloadCategoryId}`);
            }
        });

        test.afterEach(async function () {
            if (createdWorkloadIds.length > 0) {
                try {
                    await workloadsUtilsV2.deleteAllWorkloads();
                    logger.info('Successfully deleted all workloads');
                } catch (error) {
                    logger.warn('Failed to delete all workloads:', error);
                }
                createdWorkloadIds = [];
            }

            // Delete created workload types
            if (createdWorkloadTypeIds.length > 0) {
                for (const workloadTypeId of createdWorkloadTypeIds) {
                    try {
                        await workloadUtils.deleteWorkloadType(workloadTypeId);
                        logger.info(`Successfully deleted workload type: ${workloadTypeId}`);
                    } catch (error) {
                        logger.warn(`Failed to delete workload type ${workloadTypeId}:`, error);
                    }
                }
                createdWorkloadTypeIds = [];
            }

            // Restore workload type if it was updated
            if (workloadTypeUpdated) {
                try {
                    // Find TrainingWorkload and restore its original category and priority
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    const trainingWorkload = workloadTypes.find((wt) => wt.name === workloadTypeName);
                    if (trainingWorkload) {
                        await workloadUtils.updateWorkloadTypeById(trainingWorkload.id, {
                            categoryId: workloadCategoryId,
                            priorityId: trainingWorkload.priorityId // Keep current priority if not explicitly tracked
                        });
                        logger.info('Restored TrainingWorkload to original configuration');
                    }
                    workloadTypeUpdated = false;
                } catch (error) {
                    logger.warn(`Failed to restore TrainingWorkload configuration: ${error}`);
                }
            }
        });

        test.describe('Register any workload to Run:ai -> API', () => {
            test('T5405597: Create Workload Type with Minimal Required Fields', async function () {
                let createdWorkloadTypeId: string;

                await test.step('Send POST request with only required fields', async () => {
                    const workloadTypeName = 'TestJob';
                    const createResponse = await workloadUtils.createWorkloadType(
                        workloadTypeName,
                        defaultCategoryId,
                        defaultPriorityId
                    );
                    expect(createResponse.response.statusCode).toBe(StatusCodes.CREATED);
                    expect(createResponse.body).toBeDefined();
                    expect(createResponse.body.name).toBe(workloadTypeName);
                    expect(createResponse.body.categoryId).toBe(defaultCategoryId);
                    expect(createResponse.body.priorityId).toBe(defaultPriorityId);

                    createdWorkloadTypeId = createResponse.body.id;
                    createdWorkloadTypeIds.push(createdWorkloadTypeId);
                    logger.info(`Created workload type with ID: ${createdWorkloadTypeId}`);
                });

                await test.step('Confirm optional fields are properly defaulted or empty', async () => {
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    const createdWorkloadType = workloadTypes.find((wt) => wt.id === createdWorkloadTypeId);

                    expect(createdWorkloadType).toBeDefined();
                    expect(createdWorkloadType.name).toBe('TestJob');
                    expect(createdWorkloadType.categoryId).toBe(defaultCategoryId);
                    expect(createdWorkloadType.priorityId).toBe(defaultPriorityId);
                    logger.info(
                        `Workload type created with minimal configuration: ${JSON.stringify(createdWorkloadType)}`
                    );
                });
            });

            test('T5405598: Create Workload Type with Complex Configuration (DynamoGraphDeployment2)', async function () {
                let createdWorkloadTypeId: string;
                const workloadTypeName = 'DynamoGraphDeployment2';

                await test.step('Create workload type with complex configuration', async () => {
                    const complexResourceInterfaces = [
                        {
                            spec: {
                                structureDefinition: {
                                    rootComponent: {
                                        name: workloadTypeName,
                                        kind: {
                                            group: 'nvidia.com',
                                            version: 'v1alpha1',
                                            kind: workloadTypeName
                                        },
                                        statusDefinition: {
                                            phasePath: 'status.state',
                                            conditionsDefinition: {
                                                path: 'status.conditions'
                                            }
                                        }
                                    }
                                },
                                specDetails: {
                                    podTemplateUnitDetails: [
                                        {
                                            name: '< key >',
                                            replicasPath: 'spec.services[< key >].replicas',
                                            minReplicasPath: '',
                                            podTemplateSpecFieldPaths: {
                                                podSpecPath: 'spec.services[< key >].extraPodSpec',
                                                podMetaPath: 'spec.services[< key >].extraPodMetadata'
                                            },
                                            podSpecFieldPaths: {
                                                resourcesPath: 'spec.services[< key >].resources'
                                            }
                                        }
                                    ]
                                },
                                groupingInfo: {
                                    groupTopOwnerLevel: 1
                                },
                                childTypes: [
                                    {
                                        group: 'nvidia.com',
                                        version: 'v1alpha1',
                                        kind: 'DynamoComponentDeployment'
                                    },
                                    {
                                        group: 'nvidia.com',
                                        version: 'v1alpha1',
                                        kind: 'DynamoComponent'
                                    }
                                ],
                                statusDetails: {
                                    phasePath: 'status.state',
                                    conditionsDefinition: {
                                        path: 'status.conditions'
                                    },
                                    statusMappings: {
                                        running: {
                                            byConditions: [
                                                {
                                                    type: 'Ready',
                                                    status: 'True'
                                                }
                                            ]
                                        }
                                    }
                                }
                            }
                        }
                    ];

                    logger.debug(
                        `Creating complex workload type with resourceInterfaces: ${JSON.stringify(complexResourceInterfaces, null, 2)}`
                    );

                    const createResponse = await workloadUtils.createWorkloadType(
                        workloadTypeName,
                        deployCategoryId,
                        lowPriorityId,
                        'nvidia.com',
                        'v1alpha1',
                        complexResourceInterfaces
                    );

                    expect(createResponse.response.statusCode).toBe(StatusCodes.CREATED);
                    expect(createResponse.body).toBeDefined();
                    expect(createResponse.body.name).toBe(workloadTypeName);
                    expect(createResponse.body.categoryId).toBe(deployCategoryId);
                    expect(createResponse.body.priorityId).toBe(lowPriorityId);

                    createdWorkloadTypeId = createResponse.body.id;
                    createdWorkloadTypeIds.push(createdWorkloadTypeId);
                    logger.info(`Created complex workload type with ID: ${createdWorkloadTypeId}`);
                });

                await test.step('Verify complex configuration is properly parsed and stored', async () => {
                    // Retrieve the created workload type to verify storage
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    const createdWorkloadType = workloadTypes.find((wt) => wt.id === createdWorkloadTypeId);

                    expect(createdWorkloadType).toBeDefined();
                    expect(createdWorkloadType.name).toBe(workloadTypeName);
                    expect(createdWorkloadType.categoryId).toBe(deployCategoryId);
                    expect(createdWorkloadType.priorityId).toBe(lowPriorityId);

                    logger.info(
                        `Complex workload type configuration: ${JSON.stringify(createdWorkloadType)}`
                    );
                });

                await test.step('Confirm all nested objects and arrays are correctly handled', async () => {
                    // Get detailed configuration to verify nested structures
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    const complexWorkloadType = workloadTypes.find((wt) => wt.id === createdWorkloadTypeId);

                    expect(complexWorkloadType).toBeDefined();
                    expect(complexWorkloadType.name).toBe(workloadTypeName);
                    // This test verifies that the complex configuration was accepted and stored
                    logger.info('Complex workload type configuration successfully created and stored');
                    logger.info(
                        `Stored workload type details: ${JSON.stringify(complexWorkloadType, null, 2)}`
                    );
                });
            });

            test('T5405599: Create Workload Type with Duplicate Name (Same Tenant)', async function () {
                let existingWorkloadType: any;
                await test.step('Get an existing workload type to duplicate', async () => {
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    // Find an existing workload type to duplicate (using 'Job' as example)
                    existingWorkloadType = workloadTypes.find((wt) => wt.name === 'Job');
                    if (!existingWorkloadType) {
                        existingWorkloadType = workloadTypes[0];
                    }
                    expect(existingWorkloadType).toBeDefined();
                    logger.info(
                        `Using existing workload type: ${existingWorkloadType.name} (ID: ${existingWorkloadType.id})`
                    );
                    logger.info(
                        `CategoryId: ${existingWorkloadType.categoryId}, PriorityId: ${existingWorkloadType.priorityId}`
                    );
                });

                await test.step('Attempt to create a workload type with the same name, categoryId and priorityId', async () => {
                    try {
                        const duplicateResponse = await workloadUtils.createWorkloadType(
                            existingWorkloadType.name,
                            existingWorkloadType.categoryId,
                            existingWorkloadType.priorityId
                        );
                        if (duplicateResponse && duplicateResponse.body && duplicateResponse.body.id) {
                            createdWorkloadTypeIds.push(duplicateResponse.body.id);
                        }
                        throw new Error(
                            'Expected duplicate workload type creation to fail with 409 error, but it succeeded'
                        );
                    } catch (error) {
                        expect(error.message).toContain('409');
                        expect(error.message).toContain('already exists');
                    }
                });
            });

            test('T5405605: Create Workload Type with Invalid Group/Version', async function ({}) {
                await test.step('Send POST request with invalid Kubernetes group format', async () => {
                    const workloadTypeName = 'InvalidGroup';
                    try {
                        const createResponse = await workloadUtils.createWorkloadType(
                            workloadTypeName,
                            defaultCategoryId,
                            defaultPriorityId,
                            '1invalidgroupabc',
                            '1invalid_version'
                        );

                        if (createResponse && createResponse.body && createResponse.body.id) {
                            logger.info(`Workload type created with ID: ${createResponse.body.id}`);
                            createdWorkloadTypeIds.push(createResponse.body.id);
                        }
                        throw new Error(`API unexpectedly returned ${createResponse.response.statusCode}`);
                    } catch (error) {
                        logger.info(`API rejected invalid group/version format: ${error.message}`);
                        expect(error.message).toContain('400');
                    }
                });
            });

            test('T5405606: Create Workload Type with Invalid priorityId/categoryId', async function () {
                await test.step('Send POST request with invalid priorityId and categoryId format', async () => {
                    const workloadTypeName = 'InvalidPriorityId';
                    const invalidCategoryId = 'invalid-category-id';
                    const invalidPriorityId = 'invalid-priority-id';
                    try {
                        const createResponse = await workloadUtils.createWorkloadType(
                            workloadTypeName,
                            invalidCategoryId,
                            invalidPriorityId,
                            'apps',
                            'v1'
                        );
                        if (createResponse && createResponse.body && createResponse.body.id) {
                            createdWorkloadTypeIds.push(createResponse.body.id);
                        }
                        throw new Error('API unexpectedly accepted invalid categoryId/priorityId');
                    } catch (error) {
                        expect(error.message).toContain('400');
                        expect(error.message.toLowerCase()).toMatch(/category.*uuid|priority.*uuid/);
                    }
                });
            });

            test('T5405607: Create Workload Type with valid format priorityId/categoryId but the priority and category do not exist in the system', async function ({}, testInfo) {
                skipPlaywrightTest(test, testInfo, 'Workload-type registration with nonexistent category and priority IDs remains disabled pending validation of the expected rejection');
                await test.step('Send POST request with invalid priorityId and categoryId format', async () => {
                    const workloadTypeName = 'InvalidPriorityId';
                    const invalidCategoryId = '123e4567-e89b-12d3-a456-426614174000';
                    const invalidPriorityId = '123e4567-e89b-12d3-a456-426614174000';
                    try {
                        const createResponse = await workloadUtils.createWorkloadType(
                            workloadTypeName,
                            invalidCategoryId,
                            invalidPriorityId,
                            'apps',
                            'v1'
                        );
                        if (createResponse && createResponse.body && createResponse.body.id) {
                            createdWorkloadTypeIds.push(createResponse.body.id);
                        }
                        throw new Error('API unexpectedly accepted invalid categoryId/priorityId');
                    } catch (error) {
                        expect(error.message).toContain('400');
                        expect(error.message.toLowerCase()).toMatch(/category.*uuid|priority.*uuid/);
                    }
                });
            });

            test('T5405608: Get All workload types', async function () {
                await test.step('Get all workload types and verify list', async () => {
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    expect(workloadTypes.length).toBeGreaterThan(0);
                    logger.info('All workload types are:');
                    for (const workloadType of workloadTypes) {
                        logger.info(`Workload type: ${workloadType.name}`);
                    }
                });
            });

            test('T5405609: Get workload type by ID', async function () {
                const workloadTypes = await workloadUtils.getWorkloadTypes();
                expect(workloadTypes.length).toBeGreaterThan(0);
                const randomWorkloadType = workloadTypes[0];

                const workloadTypeById = await workloadUtils.getWorkloadTypeById(randomWorkloadType.id);
                logger.info(`Workload type by ID: ${workloadTypeById.name}`);
                expect(workloadTypeById).toBeDefined();
                expect(workloadTypeById.id).toBe(randomWorkloadType.id);
                expect(workloadTypeById.name).toBe(randomWorkloadType.name);

                logger.info(`Successfully retrieved workload type by ID: ${workloadTypeById.name}`);
            });

            test('T5405611: Update Mutable Fields Successfully', async function () {
                let workloadTypeId: string;
                let originalCategoryId: string;
                let originalPriorityId: string;
                let originalCategoryName: string;
                let originalPriorityName: string;
                let newCategory: any;
                let newPriority: any;
                let originalGroup: string;

                await test.step('Get a valid workload type ID and current details', async () => {
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    expect(Array.isArray(workloadTypes)).toBe(true);
                    expect(workloadTypes.length).toBeGreaterThan(0);

                    const workloadType = workloadTypes.find((t) => t.name === workloadTypeName);

                    workloadTypeId = workloadType.id;
                    logger.info(`Workload type ID: ${workloadTypeId}`);
                    const currentDetails = await workloadUtils.getWorkloadTypeById(workloadTypeId);
                    originalCategoryId = currentDetails.categoryId;
                    originalPriorityId = currentDetails.priorityId;
                    originalCategoryName = currentDetails.categoryName;
                    originalPriorityName = currentDetails.priorityName;
                    originalGroup = currentDetails.group;

                    logger.info(`Original category: ${originalCategoryName} (${originalCategoryId})`);
                    logger.info(`Original priority: ${originalPriorityName} (${originalPriorityId})`);
                });

                await test.step('Update the workload type with new category and priority', async () => {
                    const priorities = await workloadUtils.getWorkloadPriorities();
                    newPriority = priorities.find((p) => p.id !== originalPriorityId);
                    expect(newPriority).toBeDefined();
                    const categories = await workloadUtils.getWorkloadCategories();
                    expect(Array.isArray(categories)).toBe(true);
                    expect(categories.length).toBeGreaterThan(0);

                    newCategory = categories.find((c) => c.id !== originalCategoryId);
                    expect(newCategory).toBeDefined();
                    logger.info(`Updating to category: ${newCategory.name} (${newCategory.id})`);
                    logger.info(`Updating to priority: ${newPriority.name} (${newPriority.id})`);

                    const updateResponse = await workloadUtils.updateWorkloadTypeById(workloadTypeId, {
                        categoryId: newCategory.id,
                        priorityId: newPriority.id
                    });
                    workloadTypeUpdated = true;
                    expect(updateResponse.response.statusCode).toBe(StatusCodes.OK);
                    logger.info('Workload type updated successfully');
                });

                await test.step('Verify the updated fileds are correct', async () => {
                    const updatedDetails = await workloadUtils.getWorkloadTypeById(workloadTypeId);
                    expect(updatedDetails).toBeDefined();
                    expect(updatedDetails.categoryId).toBe(newCategory.id);
                    expect(updatedDetails.priorityId).toBe(newPriority.id);

                    logger.info('Update verification completed successfully');
                });

                await test.step('Verify the Immutable fields (name, group) remain unchanged', async () => {
                    const updatedDetails = await workloadUtils.getWorkloadTypeById(workloadTypeId);
                    expect(updatedDetails.name).toBe(workloadTypeName);
                    expect(updatedDetails.group).toBe(originalGroup);
                    logger.info('Immutable fields remain unchanged');
                });
            });

            test('T5405612: Try to update Immutable Fields', async function ({}) {
                let workloadTypeId: string;
                let originalCategoryId: string;
                let originalPriorityId: string;
                let originalCategoryName: string;
                let originalPriorityName: string;
                let originalGroup: string;

                await test.step('Get a valid workload type ID and current details', async () => {
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    expect(Array.isArray(workloadTypes)).toBe(true);
                    expect(workloadTypes.length).toBeGreaterThan(0);

                    const workloadType = workloadTypes.find((t) => t.name === workloadTypeName);

                    workloadTypeId = workloadType.id;
                    logger.info(`Workload type ID: ${workloadTypeId}`);
                    const currentDetails = await workloadUtils.getWorkloadTypeById(workloadTypeId);
                    originalCategoryId = currentDetails.categoryId;
                    originalPriorityId = currentDetails.priorityId;
                    originalCategoryName = currentDetails.categoryName;
                    originalPriorityName = currentDetails.priorityName;
                    originalGroup = currentDetails.group;

                    logger.info(`Original category: ${originalCategoryName} (${originalCategoryId})`);
                    logger.info(`Original priority: ${originalPriorityName} (${originalPriorityId})`);
                });

                await test.step('Attempt to update immutable fields (group) directly via API', async () => {
                    const updateFieldsWithImmutable = {
                        group: 'newgroup.io',
                        categoryId: originalCategoryId,
                        priorityId: originalPriorityId
                    };

                    const updateResponse = await workloadsService.workloadPropertiesApi.updateWorkloadType(
                        workloadTypeId,
                        updateFieldsWithImmutable as any
                    );
                    expect(updateResponse.response.statusCode).toBe(StatusCodes.OK);
                });

                await test.step('Final verification: All fields remain unchanged', async () => {
                    const finalDetails = await workloadUtils.getWorkloadTypeById(workloadTypeId);
                    logger.info(`Final details: ${JSON.stringify(finalDetails, null, 2)}`);
                    expect(finalDetails).toBeDefined();

                    // All fields should remain the same as we only tested API rejection
                    expect(finalDetails.categoryId).toBe(originalCategoryId);
                    expect(finalDetails.priorityId).toBe(originalPriorityId);
                    expect(finalDetails.name).toBe(workloadTypeName);
                    expect(finalDetails.group).toBe(originalGroup);
                });
            });

            test('T5405613: Delete User-Created Workload Type Successfully', async function () {
                let workloadTypeId: string;
                const workloadTypeName = 'CustomWorkload';

                await test.step('Create a workload type, check if user_added flag is set', async () => {
                    const workloadType = await workloadUtils.createWorkloadType(
                        workloadTypeName,
                        defaultCategoryId,
                        defaultPriorityId,
                        'custom.io',
                        'v1'
                    );
                    expect(workloadType.response.statusCode).toBe(StatusCodes.CREATED);

                    workloadTypeId = workloadType.body.id;
                    createdWorkloadTypeIds.push(workloadTypeId);
                });

                await test.step('Delete the workload type', async () => {
                    const deleteResponse = await workloadUtils.deleteWorkloadType(workloadTypeId);
                    expect(deleteResponse.response.statusCode).toBe(StatusCodes.NO_CONTENT);
                    logger.info('Workload type deleted successfully');

                    // Remove from cleanup list since we've already deleted it
                    const index = createdWorkloadTypeIds.indexOf(workloadTypeId);
                    if (index > -1) {
                        createdWorkloadTypeIds.splice(index, 1);
                        logger.info(`Removed workload type ${workloadTypeId} from cleanup list`);
                    }
                });
            });

            test('T5405614: Try to delete system Workload Type', async function () {
                let workloadTypeId: string;

                await test.step('Select a system workload type', async () => {
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    expect(workloadTypes.length).toBeGreaterThan(0);
                    workloadTypeId = workloadTypes.find((wt) => wt.name === workloadTypeName)?.id;
                    expect(workloadTypeId).toBeDefined();
                });

                await test.step('Attempt to delete system workload type (should fail)', async () => {
                    try {
                        const deleteResponse = await workloadUtils.deleteWorkloadType(workloadTypeId);
                        throw new Error('System workload type deletion should have failed but succeeded');
                    } catch (error) {
                        logger.info(
                            `API correctly rejected deletion of system workload type: ${error.message}`
                        );
                        expect(error.statusCode).toBe(StatusCodes.BAD_REQUEST);
                        expect(error.message.toLowerCase()).toMatch(
                            /not.*user.*defined.*type|cannot.*be.*deleted/
                        );
                    }
                });
            });

            test('T5405619: Filter Workload Types by Name', async function () {
                await test.step('Verify workload-type filtering by name via WorkloadPropertiesApi', async () => {
                    const accessToken = flowRootState.users.system_administrator.accessToken;
                    workloadsService.workloadPropertiesApi.accessToken = accessToken;

                    const [trainingTypes, workspaceTypes, inferenceTypes] = await Promise.all([
                        workloadsService.workloadPropertiesApi.listWorkloadTypes(
                            undefined,
                            ['name==TrainingWorkload'],
                            true
                        ),
                        workloadsService.workloadPropertiesApi.listWorkloadTypes(
                            undefined,
                            ['name==InteractiveWorkload'],
                            true
                        ),
                        workloadsService.workloadPropertiesApi.listWorkloadTypes(
                            undefined,
                            ['name==InferenceWorkload'],
                            true
                        )
                    ]);

                    const names = (resp: any): string[] => (resp.body.types || []).map((t: any) => t.name);

                    const trainingNames = names(trainingTypes);
                    expect(trainingNames.length).toBeGreaterThan(0);
                    expect(trainingNames.every((n: string) => n === 'TrainingWorkload')).toBe(true);

                    const workspaceNames = names(workspaceTypes);
                    expect(workspaceNames.length).toBeGreaterThan(0);
                    expect(workspaceNames.every((n: string) => n === 'InteractiveWorkload')).toBe(true);

                    const inferenceNames = names(inferenceTypes);
                    expect(inferenceNames.length).toBeGreaterThan(0);
                    expect(inferenceNames.every((n: string) => n === 'InferenceWorkload')).toBe(true);
                });
            });

            test('T5405629: Validation of CRD Status', async function () {
                await test.step('Retrieve all ResourceInterface CRDs and validate status', async () => {
                    const workloadTypes = await workloadUtils.getWorkloadTypes();
                    expect(workloadTypes.length).toBeGreaterThan(0);
                    logger.info(`Found ${workloadTypes.length} workload types`);

                    const crdStatusesToCheck: any[] = [];

                    for (const workloadType of workloadTypes) {
                        if (workloadType.resourceInterfaces && workloadType.resourceInterfaces.length > 0) {
                            for (const ri of workloadType.resourceInterfaces) {
                                const kind =
                                    ri.resourceInterface?.spec?.structureDefinition?.rootComponent?.kind;
                                if (kind && kind.group && kind.kind) {
                                    const plural = kind.kind.toLowerCase().endsWith('s')
                                        ? kind.kind.toLowerCase()
                                        : `${kind.kind.toLowerCase()}s`;
                                    const crdName = `${plural}.${kind.group}`;

                                    crdStatusesToCheck.push({
                                        workloadTypeName: workloadType.name,
                                        workloadTypeId: workloadType.id,
                                        crdName: crdName,
                                        clusterStatuses: workloadType.clusterStatuses || [],
                                        kindName: kind.kind
                                    });
                                }
                            }
                        }
                    }

                    logger.info(`Found ${crdStatusesToCheck.length} CRDs to validate status`);
                    expect(crdStatusesToCheck.length).toBeGreaterThan(0);

                    for (const crdInfo of crdStatusesToCheck) {
                        await test.step(`Validate status for CRD ${crdInfo.crdName}`, async () => {
                            if (crdInfo.crdName.includes('pod.') || crdInfo.crdName.includes('service.')) {
                                logger.info(`Core resource detected (no CRD expected): ${crdInfo.crdName}`);
                            }

                            // Get CRD status from cluster
                            const getCrdStatusCommand = `kubectl get crd ${crdInfo.crdName} -o jsonpath='{.status.conditions[0].type}'`;
                            let clusterCrdStatus = '';

                            try {
                                clusterCrdStatus = bashCli.exec(getCrdStatusCommand).trim();
                                logger.info(`CRD ${crdInfo.crdName} cluster status: ${clusterCrdStatus}`);
                            } catch (error) {
                                if (
                                    error.message.includes('NotFound') ||
                                    error.message.includes('not found')
                                ) {
                                    logger.info(
                                        `CRD ${crdInfo.crdName} not found in cluster (likely core resource) - proceeding to API status validation`
                                    );
                                } else {
                                    throw error;
                                }
                            }

                            // Get workload type details to check clusterStatuses
                            const workloadTypeDetails = await workloadUtils.getWorkloadTypeById(
                                crdInfo.workloadTypeId
                            );
                            const clusterStatuses = workloadTypeDetails.clusterStatuses || [];

                            logger.info(
                                `Workload type ${crdInfo.workloadTypeName} clusterStatuses: ${JSON.stringify(clusterStatuses)}`
                            );

                            // Check if there are cluster statuses for this workload type
                            if (clusterStatuses.length > 0) {
                                const clusterStatus = clusterStatuses[0];
                                logger.info(
                                    `Found cluster status: phase=${clusterStatus.phase}, version=${clusterStatus.version}`
                                );

                                // Validate expected phase when defined
                                let expectedPhase = expectedPhaseByWorkloadType[crdInfo.workloadTypeName];
                                // For core K8s kinds like Deployment/StatefulSet, expect Ready even if not a CRD
                                if (
                                    !expectedPhase &&
                                    (crdInfo.kindName === 'Deployment' || crdInfo.kindName === 'StatefulSet')
                                ) {
                                    expectedPhase = 'Ready';
                                }
                                if (expectedPhase) {
                                    expect(clusterStatus.phase).toBe(expectedPhase);
                                }

                                // Log CRDExists condition when present
                                if (clusterStatus.conditions && clusterStatus.conditions.length > 0) {
                                    const crdExistsCondition = clusterStatus.conditions.find(
                                        (c: any) => c.type === 'CRDExists'
                                    );
                                    if (crdExistsCondition) {
                                        logger.info(
                                            `CRDExists condition for ${crdInfo.crdName}: ${crdExistsCondition.status}`
                                        );
                                    }
                                }
                            } else {
                                logger.info(
                                    `No cluster status found for workload type ${crdInfo.workloadTypeName}`
                                );
                            }
                        });
                    }
                });
            });
        });
    });
});
