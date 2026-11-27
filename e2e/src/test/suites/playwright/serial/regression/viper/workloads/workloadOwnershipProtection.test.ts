/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    authUtils,
    config,
    crds,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    usersManager,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2,
    versionUtils
} from '../../../../../../../modules/helpers';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsCrudUtils';
import { workloadsCrudService } from '../../../../../../../modules/api';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import {
    shouldRunTests,
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { PROPERTY_TO_ROLE_MAPPING } from '../../../../../../../modules/helpers/config/roleMapping';
import { ScopeType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/scopeType';
import { UserPostResponse } from '../../../../../../../modules/api/apis/identityManagerApi/generated/model/userPostResponse';
import { SubjectType } from '../../../../../../../modules/api/apis/authorizationApi/generated/model/subjectType';
import { Training } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/training';
import { Inference } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/inference';
import { DistributedInference } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInference';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { expect as playwrightExpect, test } from '@playwright/test';
import { expect } from 'chai';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { authorizationUtils } from '../../../../../../../modules/helpers';
import {
    isKnativeInstalled,
    isLwsInstalled
} from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { StatusCodes } from 'http-status-codes';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { kubeCli } from '../../../../../../../modules/cli';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(`Workload Ownership Protection tests, tags: ${tags.tool.api} ${tags.operations.submit} ${tags.resource.users} ${tags.teams.viper}`, function () {
    // Type for requirement check results
    type RequirementCheckResult = {
        supported: boolean;
        reason: string;
    };
    // Workload type configuration table
    const WorkloadTypeConfigs = [
        {
            type: WorkloadType.TRAINING,
            displayName: 'Training',
            teamTag: tags.teams.viper,
            crd: crds.runaiTrainingWorkloadCrd,
            apiService: (): typeof workloadsCrudService.trainingsApi => workloadsCrudService.trainingsApi,
            getFunction: workloadsCrudUtils.getTraining,
            deleteFunction: workloadsCrudUtils.deleteTraining,
            suspendFunction: workloadsCrudUtils.suspendTraining,
            resumeFunction: workloadsCrudUtils.resumeTraining,
            supportsSuspendResume: true,
            supportsUpdate: false,
            spec: { image: 'ubuntu', command: 'sleep infinity' },
            namePrefix: 'train-ownership',
            operationType: 'suspend/resume',
            permittedRole: PROPERTY_TO_ROLE_MAPPING.l1_researcher,
            notPermittedRole: PROPERTY_TO_ROLE_MAPPING.ml_engineer,
            minVersion: SupportedVersions.V2_16, // Oldest customers supported version
            additionalRequirementCheck: null, // No additional requirements
            additionalRequirementName: null
        },
        {
            type: WorkloadType.INFERENCE,
            displayName: 'Inference',
            teamTag: tags.teams.viper,
            crd: crds.runaiInferenceWorkloadCrd,
            apiService: (): typeof workloadsCrudService.inferencesApi => workloadsCrudService.inferencesApi,
            getFunction: workloadsCrudUtils.getInference,
            deleteFunction: workloadsCrudUtils.deleteInference,
            updateFunction: workloadsCrudUtils.updateInferenceSpec,
            supportsSuspendResume: false,
            supportsUpdate: true,
            spec: {
                image: config.quickStartImage.inferenceHelloWorld,
                servingPort: { container: 8080 }
            },
            namePrefix: 'inf-ownership',
            operationType: 'update replicas',
            permittedRole: PROPERTY_TO_ROLE_MAPPING.ml_engineer,
            notPermittedRole: PROPERTY_TO_ROLE_MAPPING.l1_researcher,
            minVersion: SupportedVersions.V2_19,
            additionalRequirementCheck: isKnativeInstalled,
            additionalRequirementName: 'Knative'
        },
        {
            type: WorkloadType.DISTRIBUTED_INFERENCE,
            displayName: 'Distributed Inference',
            teamTag: tags.teams.viper,
            crd: crds.runaiDistributedInferenceWorkloadCrd,
            apiService: (): typeof workloadsCrudService.distributedInferencesApi =>
                workloadsCrudService.distributedInferencesApi,
            getFunction: workloadsCrudUtils.getDistributedInference,
            deleteFunction: workloadsCrudUtils.deleteDistributedInference,
            updateFunction: workloadsCrudUtils.updateDistributedInferenceSpec,
            supportsSuspendResume: false,
            supportsUpdate: true,
            spec: {
                replicas: 1,
                servingPort: { port: 8080 },
                leader: { image: config.quickStartImage.inferenceHelloWorld }
            },
            namePrefix: 'dinf-ownership',
            operationType: 'update replicas',
            permittedRole: PROPERTY_TO_ROLE_MAPPING.ml_engineer,
            notPermittedRole: PROPERTY_TO_ROLE_MAPPING.l1_researcher,
            minVersion: SupportedVersions.V2_23,
            additionalRequirementCheck: isLwsInstalled,
            additionalRequirementName: 'LWS'
        }
    ];
    // Configure retries for all tests in this suite
    test.describe.configure({ retries: 3 });

    // Log retry attempts for better visibility
    test.beforeEach(({}, testInfo) => {
        if (testInfo.retry > 0) {
            logger.warn(`🔄 RETRY #${testInfo.retry} for test: ${testInfo.title}`);
        } else {
            logger.info(`▶️  Starting test: ${testInfo.title}`);
        }
    });

    shouldRunTests(!isOCP())(function () {
        let project: Project;
        const testUsers: UserPostResponse[] = [];
        const testUserTokenIds: string[] = [];

        const requirementResults = new Map<string, RequirementCheckResult>();

        let adminUser: number;
        let permittedUser: number; // User with permissions for the current workload type
        let notPermittedUser: number; // User without permissions for the current workload type
        let anyWorkloadSubmissionEnabled: boolean;
        let kubernetesServerVersion: string | undefined;
        let isAffectedByJobResumeBug = false;

        async function createUserWithRule(
            emailPrefix: string,
            rule: string,
            scope: ScopeType,
            scopeId: string
        ): Promise<UserPostResponse> {
            const email = objectGenerator.generateEmailByPrefixWithRandom(emailPrefix, 6);

            const user = (
                await usersManager.createLocalUser(
                    flowRootState.users.system_administrator.accessToken,
                    email
                )
            ).body;
            const token = await authUtils.authenticateUser(user.username, user.tempPassword);

            testUsers.push(user);
            testUserTokenIds.push(token.idToken);

            const roles = await authorizationUtils.getRoles(
                flowRootState.users.system_administrator.accessToken
            );
            const roleObj = roles.find((r) => r.name === rule);
            playwrightExpect(roleObj).not.toBeUndefined();
            await authorizationUtils.createAccessRuleWithDefaults(
                user.username,
                SubjectType.User,
                roleObj!.id,
                scopeId,
                scope
            );
            logger.info(`Created user with email: ${email}, and access-rule: ${rule}`);
            return user;
        }

        async function createUsersForWorkloadType(config: (typeof WorkloadTypeConfigs)[0]): Promise<void> {
            // Create permitted user for this workload type
            await createUserWithRule(
                'permitted-user',
                config.permittedRole,
                ScopeType.Project,
                project.id.toString()
            );
            permittedUser = testUsers.length - 1;

            // Create not permitted user for this workload type
            await createUserWithRule(
                'not-permitted-user',
                config.notPermittedRole,
                ScopeType.Project,
                project.id.toString()
            );
            notPermittedUser = testUsers.length - 1;
        }

        async function createRequiredUsers(): Promise<void> {
            // Only create department admin globally
            await createUserWithRule(
                'admin',
                PROPERTY_TO_ROLE_MAPPING.department_admin,
                ScopeType.Cluster,
                flowRootState.cluster.uuid
            );
            adminUser = testUsers.length - 1;
        }

        function setAccessToken(config: (typeof WorkloadTypeConfigs)[0], userIndex: number): void {
            // Get the token and ensure it's valid
            const token = testUserTokenIds[userIndex];
            if (!token) {
                logger.error(
                    `No token available for user index ${userIndex}, available tokens: ${testUserTokenIds.length}`
                );
                throw new Error(`No token available for user index ${userIndex}`);
            }

            logger.debug(`Setting access token for user index ${userIndex}, token length: ${token.length}`);

            try {
                // Set token on the config's API service (for consistency)
                config.apiService().accessToken = token;

                // Use the proper WorkloadsCrudUtils accessToken setter that handles all API services
                workloadsCrudUtils.accessToken = token;

                logger.debug(`Access token successfully set for user index ${userIndex}`);
            } catch (error) {
                logger.error(`Failed to set access token for user index ${userIndex}: ${error.message}`);
                throw error;
            }
        }

        // Training utility function similar to inference utils
        async function createTraining(
            workloadName: string,
            workloadSpec: any,
            project: Project
        ): Promise<string[]> {
            const result = await workloadsCrudUtils.createTraining(
                workloadName,
                true /*add suffix*/,
                flowRootState.cluster.uuid,
                String(project.id),
                workloadSpec,
                testUserTokenIds[permittedUser]
            );

            expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);

            const workloadIdReturn = result.body.workloadId;
            const workloadNameReturn = result.body.name;

            return [workloadIdReturn, workloadNameReturn];
        }

        async function createInference(
            workloadName: string,
            workloadSpec: any,
            project: Project
        ): Promise<string[]> {
            const result = await workloadsCrudUtils.createInference(
                workloadName,
                true /*add suffix*/,
                flowRootState.cluster.uuid,
                String(project.id),
                workloadSpec,
                testUserTokenIds[permittedUser]
            );

            expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);

            const workloadIdReturn = result.body.workloadId;
            const workloadNameReturn = result.body.name;

            return [workloadIdReturn, workloadNameReturn];
        }

        async function createDistributedInference(
            workloadName: string,
            workloadSpec: any,
            project: Project
        ): Promise<string[]> {
            const result = await workloadsCrudUtils.createDistributedInference(
                workloadName,
                true /*add suffix*/,
                flowRootState.cluster.uuid,
                String(project.id),
                workloadSpec,
                testUserTokenIds[permittedUser]
            );

            expect(result.response.statusCode).to.equal(StatusCodes.ACCEPTED);

            const workloadIdReturn = result.body.workloadId;
            const workloadNameReturn = result.body.name;

            return [workloadIdReturn, workloadNameReturn];
        }

        async function createAndWaitForRunningWorkload(
            config: (typeof WorkloadTypeConfigs)[0]
        ): Promise<Training | Inference | DistributedInference> {
            const workloadName = objectGenerator.generateShortNameByPrefix(config.namePrefix);

            logger.info(
                `Creating ${config.type} workload: ${workloadName} in project ${project.id} (${project.namespace})`
            );

            setAccessToken(config, permittedUser);

            let workloadId: string;
            let actualWorkloadName: string;

            // Use dedicated utility functions for each workload type
            if (config.type === WorkloadType.TRAINING) {
                [workloadId, actualWorkloadName] = await createTraining(workloadName, config.spec, project);
            } else if (config.type === WorkloadType.INFERENCE) {
                [workloadId, actualWorkloadName] = await createInference(workloadName, config.spec, project);
            } else if (config.type === WorkloadType.DISTRIBUTED_INFERENCE) {
                [workloadId, actualWorkloadName] = await createDistributedInference(
                    workloadName,
                    config.spec,
                    project
                );
            } else {
                throw new Error(`Unsupported workload type: ${config.type}`);
            }

            logger.info(
                `Created ${config.type} workload ${actualWorkloadName} with ID ${workloadId}, waiting for Running phase...`
            );

            await waitUtils.waitForBackendWorkloadPhase(actualWorkloadName, project.id, Phase.Running, 90);

            logger.info(`${config.type} workload ${actualWorkloadName} is now Running`);

            if (anyWorkloadSubmissionEnabled && config.type === WorkloadType.TRAINING) {
                const workload = await workloadsUtilsV2.waitForBackendWorkloadByNameAndProjectName(
                    actualWorkloadName,
                    project.name
                );
                expect(workload.sourceApi).to.equal('WorkloadsV2');
                return {
                    name: workload.name,
                    workloadId: workload.id
                } as Training;
            }

            const { body: workload } = await config.getFunction.call(workloadsCrudUtils, workloadId);
            return workload;
        }

        async function waitForStopped(
            config: (typeof WorkloadTypeConfigs)[0],
            workloadName: string
        ): Promise<void> {
            if (anyWorkloadSubmissionEnabled && config.type === WorkloadType.TRAINING) {
                await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Stopped);
            } else {
                await waitUtils.waitForJobToBeSuspend(workloadName, project.namespace, config.crd.kind);
            }
        }

        async function waitForRunning(
            config: (typeof WorkloadTypeConfigs)[0],
            workloadName: string
        ): Promise<void> {
            if (anyWorkloadSubmissionEnabled && config.type === WorkloadType.TRAINING) {
                await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Running);
            } else {
                await waitUtils.waitForJobToBeRunning(workloadName, project.namespace, config.crd.kind);
            }
        }

        async function performWorkloadOperation(
            config: (typeof WorkloadTypeConfigs)[0],
            operation: 'suspend' | 'resume' | 'update' | 'delete',
            userIndex: number,
            payload?: any
        ): Promise<void> {
            // Set token before each operation
            setAccessToken(config, userIndex);

            logger.debug(`Performing ${operation} operation with user index ${userIndex}`);

            switch (operation) {
                case 'suspend':
                    return await (config as any).suspendFunction.call(workloadsCrudUtils, payload);
                case 'resume':
                    return await (config as any).resumeFunction.call(workloadsCrudUtils, payload);
                case 'update':
                    return await config.updateFunction!.call(
                        workloadsCrudUtils,
                        payload.workloadId,
                        payload.updateData
                    );
                case 'delete':
                    return await config.deleteFunction.call(workloadsCrudUtils, payload);
            }
        }

        // For operations that should fail (don't use retry)
        async function performWorkloadOperationExpectingFailure(
            config: (typeof WorkloadTypeConfigs)[0],
            operation: 'suspend' | 'resume' | 'update' | 'delete',
            userIndex: number,
            payload?: any
        ): Promise<void> {
            // Set token before each operation
            setAccessToken(config, userIndex);

            logger.debug(
                `Performing ${operation} operation (expecting failure) with user index ${userIndex}`
            );

            switch (operation) {
                case 'suspend':
                    return await (config as any).suspendFunction.call(workloadsCrudUtils, payload);
                case 'resume':
                    return await (config as any).resumeFunction.call(workloadsCrudUtils, payload);
                case 'update':
                    return await config.updateFunction!.call(
                        workloadsCrudUtils,
                        payload.workloadId,
                        payload.updateData
                    );
                case 'delete':
                    return await config.deleteFunction.call(workloadsCrudUtils, payload);
            }
        }

        async function expectOperationToFail(
            config: (typeof WorkloadTypeConfigs)[0],
            operation: 'suspend' | 'resume' | 'update' | 'delete',
            userIndex: number,
            payload?: any
        ): Promise<void> {
            await playwrightExpect(
                performWorkloadOperationExpectingFailure(config, operation, userIndex, payload)
            ).rejects.toThrow(/insufficient permissions/i);
        }

        let updateCounter = 1; // Start from 1 for replica updates

        function getUpdatePayload(workloadType: string): { spec: { replicas: number } } {
            return {
                spec: {
                    replicas: ++updateCounter // Increment each time
                }
            };
        }

        async function checkAllRequirements(): Promise<void> {
            for (const config of WorkloadTypeConfigs) {
                try {
                    // Check additional requirements based on workload type
                    let additionalResult: RequirementCheckResult = { supported: true, reason: '' };

                    if (config.additionalRequirementCheck) {
                        const isSupported = await config.additionalRequirementCheck();
                        additionalResult = {
                            supported: isSupported,
                            reason: isSupported ? '' : `${config.additionalRequirementName} not available`
                        };
                    }

                    requirementResults.set(config.type, additionalResult);

                    if (additionalResult.supported) {
                        logger.info(
                            `${config.displayName} workload requirements is supported and will be tested`
                        );
                    } else {
                        logger.warn(
                            `${config.displayName} workload requirements not met: ${additionalResult.reason}, tests will be skipped`
                        );
                    }
                } catch (error) {
                    requirementResults.set(config.type, {
                        supported: false,
                        reason: `Failed to check requirements: ${error.message}`
                    });
                    logger.warn(
                        `Failed to check requirements for ${config.displayName}: ${error.message}, tests will be skipped`
                    );
                }
            }
        }

        test.beforeAll(async function () {
            // Check all workload type requirements once
            await checkAllRequirements();
            anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
                AnyWorkloadSubmissionType.Training
            );
            if (anyWorkloadSubmissionEnabled) {
                kubernetesServerVersion = await kubeCli.getServerVersion();
                isAffectedByJobResumeBug =
                    versionUtils.isKubernetesVersionAffectedByJobResumeBug(kubernetesServerVersion);
            }

            await operations.enableWorkloadOwnershipProtection();

            // we already wait for the workload ownership protection to be enabled in enableWorkloadOwnershipProtection
            // await waitUtils.waitForWorkloadOwnershipProtectionReported(true);
            project = await projectResourcesUtils.createProjectAndNamespace('p');
            await createRequiredUsers(); // Only creates admin user now
        });

        test.afterAll(async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch (error) {
                logger.error('error deleting project: ' + error);
            }

            // Cleanup users
            await Promise.all(
                testUsers.map((user) =>
                    authUtils.deleteUser(user.id, flowRootState.users.system_administrator.accessToken)
                )
            );

            await operations.disableWorkloadOwnershipProtection();

            // we already wait for the workload ownership protection to be enabled in disableWorkloadOwnershipProtection
            // await waitUtils.waitForWorkloadOwnershipProtectionReported(false);
        });

        // Generate tests for all workload types
        WorkloadTypeConfigs.forEach((config) => {
            supportedFrom(config.minVersion)(() => {
                test.describe(`${config.displayName} workload ownership protection, tags: ${config.teamTag}`, function () {
                    // Local variables for this workload type
                    let localPermittedUser: number;
                    let localNotPermittedUser: number;

                    test.beforeAll(async function () {
                        const result = requirementResults.get(config.type);
                        if (!result?.supported) {
                            logger.warn(
                                `${config.displayName} requirements not met: ${result?.reason || 'Unknown error'}, skipping all tests for this workload type`
                            );
                            test.skip();
                        }
                        await createUsersForWorkloadType(config);

                        // Set local variables after users are created
                        localPermittedUser = permittedUser;
                        localNotPermittedUser = notPermittedUser;

                        logger.debug(
                            `Users created for ${config.displayName}: permitted=${localPermittedUser}, notPermitted=${localNotPermittedUser}, admin=${adminUser}`
                        );
                    });

                    test.beforeEach(function ({}, testInfo) {
                        logger.debug(`Setting up test: ${testInfo.title} for ${config.displayName}`);

                        // Ensure all users have valid tokens before each test
                        const userIndicesToCheck = [adminUser, localPermittedUser, localNotPermittedUser];
                        for (const userIndex of userIndicesToCheck) {
                            if (userIndex !== undefined && testUserTokenIds[userIndex]) {
                                const token = testUserTokenIds[userIndex];
                                logger.debug(`User ${userIndex} has token of length: ${token.length}`);
                            } else {
                                logger.warn(`User index ${userIndex} is undefined or has no token`);
                            }
                        }

                        logger.debug(
                            `Total users created: ${testUsers.length}, Total tokens: ${testUserTokenIds.length}`
                        );
                        logger.debug(
                            `Admin user index: ${adminUser}, Permitted user index: ${localPermittedUser}, Not-permitted user index: ${localNotPermittedUser}`
                        );

                        // Reset all API tokens to a known state before each test
                        workloadsCrudUtils.accessToken =
                            flowRootState.users.system_administrator.accessToken;
                    });

                    // Helper function to generate tests for each user scenario
                    function generateUserScenarioTests(
                        operation: 'suspend' | 'resume' | 'update' | 'delete',
                        getPayload: (workload: any) => any
                    ): void {
                        const userScenarios = [
                            {
                                actor: 'Submitter',
                                userIndexGetter: (): number => localPermittedUser,
                                target: 'his',
                                shouldFail: false
                            },
                            {
                                actor: 'Department admin',
                                userIndexGetter: (): number => adminUser,
                                target: 'someone elses',
                                shouldFail: false
                            },
                            {
                                actor: 'Different user',
                                userIndexGetter: (): number => localNotPermittedUser,
                                target: 'someone elses',
                                shouldFail: true
                            }
                        ];

                        userScenarios.forEach(({ actor, userIndexGetter, target, shouldFail }) => {
                            const testName = `${actor} ${shouldFail ? 'cannot' : 'can'} ${operation} ${target} ${config.type} workload`;

                            test(testName, async function ({}, testInfo) {
                                if (config.crd === crds.runaiInferenceWorkloadCrd) {
                                    skipPlaywrightTest(
                                        test,
                                        testInfo,
                                        'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                                    );
                                }
                                const userIndex = userIndexGetter();
                                logger.debug(`Using user index ${userIndex} for test: ${testName}`);

                                const workload = await createAndWaitForRunningWorkload(config);
                                const payload = getPayload(workload);

                                switch (operation) {
                                    case 'suspend':
                                        if (shouldFail) {
                                            await expectOperationToFail(config, operation, userIndex, payload);
                                        } else {
                                            logger.info(
                                                `Performing suspend operation for ${config.type} workload ${workload.name} by user ${userIndex}`
                                            );
                                            await performWorkloadOperation(
                                                config,
                                                operation,
                                                userIndex,
                                                payload
                                            );
                                            logger.info(
                                                'Suspend operation completed, waiting for workload to be suspended...'
                                            );
                                            await waitForStopped(config, workload.name);
                                            logger.info(`Workload ${workload.name} is now suspended`);
                                        }
                                        break;

                                    case 'resume':
                                        // First suspend the workload
                                        await performWorkloadOperation(
                                            config,
                                            'suspend',
                                            localPermittedUser,
                                            workload.workloadId
                                        );
                                        await waitForStopped(config, workload.name);

                                        if (shouldFail) {
                                            await expectOperationToFail(config, operation, userIndex, payload);
                                        } else {
                                            await performWorkloadOperation(
                                                config,
                                                operation,
                                                userIndex,
                                                payload
                                            );
                                            await test.step('verify the resumed workload reaches Running', async function (step) {
                                                step.skip(
                                                    anyWorkloadSubmissionEnabled &&
                                                        isAffectedByJobResumeBug &&
                                                        config.type === WorkloadType.TRAINING,
                                                    `Kubernetes ${kubernetesServerVersion} is affected by the Job resume bug`
                                                );
                                                await waitForRunning(config, workload.name);
                                            });
                                        }
                                        break;

                                    case 'update':
                                    case 'delete':
                                        if (shouldFail) {
                                            await expectOperationToFail(config, operation, userIndex, payload);
                                        } else {
                                            await performWorkloadOperation(
                                                config,
                                                operation,
                                                userIndex,
                                                payload
                                            );
                                        }
                                        break;

                                    default:
                                        throw new Error(`Unsupported operation: ${operation}`);
                                }
                            });
                        });
                    }

                    // Suspend/Resume operations
                    if (config.supportsSuspendResume) {
                        test.describe('suspend/resume operations', function () {
                            generateUserScenarioTests('suspend', (workload) => workload.workloadId);

                            generateUserScenarioTests('resume', (workload) => workload.workloadId);
                        });
                    }

                    // Update operations
                    if (config.supportsUpdate) {
                        test.describe(`${config.operationType} operations`, function () {
                            generateUserScenarioTests('update', (workload) => ({
                                workloadId: workload.workloadId,
                                updateData: getUpdatePayload(config.type)
                            }));
                        });
                    }

                    // Delete operations
                    test.describe('delete operations', function () {
                        generateUserScenarioTests('delete', (workload) => workload.workloadId);
                    });
                });
            });
        });
    });
});
