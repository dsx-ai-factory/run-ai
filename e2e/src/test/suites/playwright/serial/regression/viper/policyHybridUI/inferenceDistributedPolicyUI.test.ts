/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    apiUtils,
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    orgUnitUtils,
    policyUtils,
    runaiHealthChecker,
    tags,
    User,
    workloadsCrudUtils,
    workloadsUtilsV2,
    workloadUtils
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { WorkloadPage } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { expect, test } from '@playwright/test';
import {
    DistributedInference,
    Inference
} from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { InferenceFlatUIUtil } from '../../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import {
    Annotation,
    Capability,
    DistributedInferenceStartupPolicy,
    Label,
    PvcVolumeMode,
    ScopeType,
    UidGidSource
} from '../../../../../../../modules/api/apis/policyApi';
import { HybridForm } from '../../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { IInferenceMeta } from '../../../../../../../modules/helpers/models/inferenceFlatUI.models';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { WorkloadType as PolicyWorkloadType } from '../../../../../../../modules/helpers/utils/policyUtils';
import { ElementValidationUtils } from '../../../../../../../modules/ui/utils/ElementValidationUtils';
import { backendClusterApi } from '../../../../../../../modules/api';

test.describe(`Inferences Policy based on distributed UI, tags: ${tags.tool.ui} ${tags.resource.policy} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`, function () {
    test.setTimeout(60000);

    const suiteName = 'inference-policy-distributed';
    let nodePoolName: string;
    let nodePoolId: number;
    let project: Project;
    let traceName: string;

    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let workload: WorkloadPage;
    let inferenceFlatUIUtil: InferenceFlatUIUtil;
    let hybridForm: HybridForm;
    let elementValidationUtils: ElementValidationUtils;
    let inferenceMeta: IInferenceMeta;
    let inferenceName: string;
    let inferenceWorkload: Inference;
    let distributedInferenceWorkload: DistributedInference;

    function findItem(items: Array<Annotation | Label> | null, itemName: string, itemValue: string): boolean {
        const item = items?.find((item) => item.name === itemName && item.value === itemValue);
        return !!item;
    }

    async function initBrowser(): Promise<BrowserEngine> {
        // Initialize browser once for all tests with video recording AND tracing support
        await browserEngine.start();
        auth = new AuthPage(browserEngine.currentPage);
        workload = new WorkloadPage(browserEngine.currentPage);
        hybridForm = new HybridForm(browserEngine.currentPage);
        inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);
        elementValidationUtils = new ElementValidationUtils(browserEngine.currentPage);

        const currentUser: User = flowRootState.users.system_administrator;
        await auth.signIn(currentUser.email, currentUser.password);

        return browserEngine;
    }

    test.beforeAll('before all inference policy test based on distributed UI', async function () {
        if (!(await isKnativeInstalled())) {
            logger.warn('knative is not installed on the cluster, skipping test');
            test.skip();
        }

        browserEngine = new BrowserEngine();

        backendClusterApi.nodePoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;

        await apiUtils.enableHybridSubmission();
        // Create node pool for tests that reference it in policies
        nodePoolName = objectGenerator.generateShortNameByPrefix('np');
        const nodePoolLabelKeyValue = nodePoolUtils.generateKeyValueLabels();
        const createResponse = await nodePoolUtils.createBackendNodePool(
            nodePoolName,
            nodePoolLabelKeyValue.key,
            nodePoolLabelKeyValue.value
        );
        nodePoolId = createResponse.body.id;
        logger.info(`Created node pool ${nodePoolName} with id ${nodePoolId}`);

        const projectName = objectGenerator.generateNameByPrefix('project-e2e-for-inference-policy');
        project = await orgUnitUtils.createProjectAndValidate({
            name: projectName,
            clusterId: flowRootState.cluster.uuid,
            resources: [{ gpu: { deserved: 1, limit: 1 } }]
        });

        logger.info(`created project ${projectName}`);

        browserEngine = await initBrowser();

        logger.info('Browser initialized and user signed in');
    });

    test.afterAll('after all inference policy test based on distributed UI', async ({}, testInfo) => {
        // Close browser once after all tests and save video
        if (browserEngine && browserEngine.browser) {
            try {
                await browserEngine.stopAndArchive(testInfo, suiteName);
                logger.info('Browser closed and video saved');
            } catch (error) {
                logger.error(`Error closing browser: ${error}`);
            }
        }

        if (!(await runaiHealthChecker.isSelfHostedCluster())) {
            // This condition should be removed when fixing problem with saas envs with cluster version.
            return;
        }

        try {
            await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id, 120);
            await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            logger.info(
                `deleted workloads and project ${project.name} and namespace ${project.status.namespace}`
            );
        } catch (error) {
            logger.error(
                `failed to delete workloads and project ${project.name} and namespace ${project.status.namespace}. Error: ${JSON.stringify(error)}`
            );
        }

        // Delete node pool
        if (nodePoolId) {
            try {
                await nodePoolUtils.cleanupDeleteNodePool({ nodePoolName, nodePoolId });
                logger.info(`Deleted node pool ${nodePoolName}`);
            } catch (error) {
                logger.error(`Failed to delete node pool ${nodePoolName}: ${error}`);
            }
        }
    });

    test.beforeEach('before each inference policy test based on distributed UI', async function () {
        if (!browserEngine || !browserEngine.browser) {
            browserEngine = await initBrowser();
        }

        inferenceName = objectGenerator.generateNameByPrefix('inference');

        // Generate trace name from test title (use test.title in Mocha)
        traceName = objectGenerator.generateNameByPrefix(suiteName);

        // Start tracing for this test (context and page stay alive)
        await browserEngine.startTracing(traceName);

        inferenceMeta = {
            clusterName: flowRootState.cluster.name,
            projectName: project.name,
            workloadName: inferenceName
        };

        // Navigate to workloads page for each test
        await workload.navigateToWorkloads();
    });

    test.afterEach('after each inference policy test based on distributed UI', async ({}, testInfo) => {
        logger.debug(`waiting for workload ${inferenceName} to be deleted`);

        // cleanup inference policy
        try {
            await workloadUtils.deletePolicy(project.id);
        } catch (error) {
            logger.debug(`Error deleting inference policy: ${error}`);
        }

        // cleanup distributed inference policy
        try {
            await policyUtils.deleteWorkloadPolicyV2(PolicyWorkloadType.DistributedInference, {
                scope: 'project',
                scopeId: project.id.toString()
            });
        } catch (error) {
            logger.debug(`Error deleting distributed inference policy: ${error}`);
        }

        // delete inference workload
        if (inferenceWorkload) {
            await workloadsUtilsV2.deleteWorkload(
                inferenceWorkload.workloadId,
                WorkloadType.INFERENCE,
                flowRootState.users.system_administrator.accessToken
            );
        }
        inferenceWorkload = undefined;

        // delete distributed inference workload
        if (distributedInferenceWorkload) {
            await workloadsUtilsV2.deleteWorkload(
                distributedInferenceWorkload.workloadId,
                WorkloadType.DISTRIBUTED_INFERENCE,
                flowRootState.users.system_administrator.accessToken
            );
        }
        distributedInferenceWorkload = undefined;

        logger.debug(`inference ${inferenceName} deleted!`);

        // Stop and archive tracing for this test (browser, context, and page remain alive)
        await browserEngine.stopAndArchiveTracing(testInfo, traceName);
    });

    supportedFrom(SupportedVersions.V2_23)(function () {
        test.describe('Test inference policy defaults', function () {
            test.describe('Test environment section', function () {
                test('distributed inference environment policy defaults - leader and worker similar', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('env-policy');
                    const commonDefaults = {
                        args: 'x=1',
                        command: 'test-command',
                        image: 'ubuntu',
                        environmentVariables: {
                            instances: [
                                {
                                    name: 'test-env-var',
                                    value: 'test-value'
                                }
                            ]
                        },
                        security: {
                            uidGidSource: UidGidSource.Custom,
                            runAsUid: 1000,
                            runAsGid: 1000,
                            supplementalGroups: '1000,1001',
                            capabilities: [Capability.NetAdmin]
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: commonDefaults,
                        worker: commonDefaults
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        undefined,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();

                    // Validate leader defaults
                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;
                    const leaderDefaults = policyDefaults.leader;

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridForm.hybridEnvironmentSection.hybridImageInput,
                        leaderDefaults.image
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridForm.hybridEnvironmentSection.servingPortPort,
                        policyDefaults.servingPort.port.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.cmdInput,
                        leaderDefaults.command
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.argsInput,
                        leaderDefaults.args
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.envVarNameInput,
                        leaderDefaults.environmentVariables.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.envVarValue,
                        leaderDefaults.environmentVariables.instances[0].value
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.uidInput,
                        leaderDefaults.security.runAsUid.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.gidInput,
                        leaderDefaults.security.runAsGid.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.supplementalGroupsInput,
                        leaderDefaults.security.supplementalGroups
                    );
                    await elementValidationUtils.expectElementValueToContain(
                        '.capabilities-select',
                        leaderDefaults.security.capabilities[0].toString()
                    );

                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn(); // next step - worker
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker defaults (same as leader)
                    const workerDefaults = policyDefaults.worker;
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridForm.hybridEnvironmentSection.hybridImageInput,
                        workerDefaults.image
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.cmdInput,
                        workerDefaults.command
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.argsInput,
                        workerDefaults.args
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.envVarNameInput,
                        workerDefaults.environmentVariables.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.envVarValue,
                        workerDefaults.environmentVariables.instances[0].value
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.uidInput,
                        workerDefaults.security.runAsUid.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.gidInput,
                        workerDefaults.security.runAsGid.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.supplementalGroupsInput,
                        workerDefaults.security.supplementalGroups
                    );
                    await elementValidationUtils.expectElementValueToContain(
                        '.capabilities-select',
                        workerDefaults.security.capabilities[0].toString()
                    );

                    // Go to workload setup step and submit
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn(); // next step - workload setup
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Check leader spec
                    expect(distributedInferenceWorkload.spec.leader.image).toBe(leaderDefaults.image);
                    expect(distributedInferenceWorkload.spec.leader.command).toBe(leaderDefaults.command);
                    expect(distributedInferenceWorkload.spec.leader.args).toBe(leaderDefaults.args);
                    expect(distributedInferenceWorkload.spec.leader.environmentVariables).toEqual([
                        {
                            name: leaderDefaults.environmentVariables.instances[0].name,
                            value: leaderDefaults.environmentVariables.instances[0].value
                        }
                    ]);
                    expect(distributedInferenceWorkload.spec.leader.security.capabilities).toEqual(
                        leaderDefaults.security.capabilities
                    );
                    expect(distributedInferenceWorkload.spec.leader.security.runAsUid).toBe(
                        leaderDefaults.security.runAsUid
                    );
                    expect(distributedInferenceWorkload.spec.leader.security.runAsGid).toBe(
                        leaderDefaults.security.runAsGid
                    );
                    expect(distributedInferenceWorkload.spec.leader.security.supplementalGroups).toBe(
                        leaderDefaults.security.supplementalGroups
                    );

                    // Check worker spec (same as leader)
                    expect(distributedInferenceWorkload.spec.worker.image).toBe(workerDefaults.image);
                    expect(distributedInferenceWorkload.spec.worker.command).toBe(workerDefaults.command);
                    expect(distributedInferenceWorkload.spec.worker.args).toBe(workerDefaults.args);
                    expect(distributedInferenceWorkload.spec.worker.environmentVariables).toEqual([
                        {
                            name: workerDefaults.environmentVariables.instances[0].name,
                            value: workerDefaults.environmentVariables.instances[0].value
                        }
                    ]);
                    expect(distributedInferenceWorkload.spec.worker.security.capabilities).toEqual(
                        workerDefaults.security.capabilities
                    );
                    expect(distributedInferenceWorkload.spec.worker.security.runAsUid).toBe(
                        workerDefaults.security.runAsUid
                    );
                    expect(distributedInferenceWorkload.spec.worker.security.runAsGid).toBe(
                        workerDefaults.security.runAsGid
                    );
                    expect(distributedInferenceWorkload.spec.worker.security.supplementalGroups).toBe(
                        workerDefaults.security.supplementalGroups
                    );

                    logger.info(
                        'test distributed inference policy environment section with similar leader and worker passed!'
                    );
                });

                test('distributed inference environment policy defaults - leader and worker different', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('env-policy-diff');
                    const leaderDefaults = {
                        args: 'leader-args',
                        command: 'leader-command',
                        image: 'ubuntu:leader',
                        environmentVariables: {
                            instances: [
                                {
                                    name: 'leader-env-var',
                                    value: 'leader-value'
                                }
                            ]
                        },
                        security: {
                            uidGidSource: UidGidSource.Custom,
                            runAsUid: 1000,
                            runAsGid: 1000,
                            supplementalGroups: '1000,1001',
                            capabilities: [Capability.NetAdmin]
                        }
                    };

                    const workerDefaults = {
                        args: 'worker-args',
                        command: 'worker-command',
                        image: 'ubuntu:worker',
                        environmentVariables: {
                            instances: [
                                {
                                    name: 'worker-env-var',
                                    value: 'worker-value'
                                }
                            ]
                        },
                        security: {
                            uidGidSource: UidGidSource.Custom,
                            runAsUid: 2000,
                            runAsGid: 2000,
                            supplementalGroups: '2000,2001',
                            capabilities: [Capability.SysAdmin]
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: leaderDefaults,
                        worker: workerDefaults
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        undefined,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();

                    // Validate leader defaults
                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridForm.hybridEnvironmentSection.hybridImageInput,
                        leaderDefaults.image
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridForm.hybridEnvironmentSection.servingPortPort,
                        policyDefaults.servingPort.port.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.cmdInput,
                        leaderDefaults.command
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.argsInput,
                        leaderDefaults.args
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.envVarNameInput,
                        leaderDefaults.environmentVariables.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.envVarValue,
                        leaderDefaults.environmentVariables.instances[0].value
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.uidInput,
                        leaderDefaults.security.runAsUid.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.gidInput,
                        leaderDefaults.security.runAsGid.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.supplementalGroupsInput,
                        leaderDefaults.security.supplementalGroups
                    );
                    await elementValidationUtils.expectElementValueToContain(
                        '.capabilities-select',
                        leaderDefaults.security.capabilities[0].toString()
                    );

                    // Fill serving port and go to worker step
                    await inferenceFlatUIUtil.fillServingPortOptions({
                        port: 3131
                    });
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn(); // next step - worker
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker defaults (different from leader)
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridForm.hybridEnvironmentSection.hybridImageInput,
                        workerDefaults.image
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.cmdInput,
                        workerDefaults.command
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.argsInput,
                        workerDefaults.args
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.envVarNameInput,
                        workerDefaults.environmentVariables.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.envVarValue,
                        workerDefaults.environmentVariables.instances[0].value
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.uidInput,
                        workerDefaults.security.runAsUid.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.gidInput,
                        workerDefaults.security.runAsGid.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.supplementalGroupsInput,
                        workerDefaults.security.supplementalGroups
                    );
                    await elementValidationUtils.expectElementValueToContain(
                        '.capabilities-select',
                        workerDefaults.security.capabilities[0].toString()
                    );

                    // Go to workload setup step and submit
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn(); // next step - workload setup
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Check leader spec
                    expect(distributedInferenceWorkload.spec.leader.image).toBe(leaderDefaults.image);
                    expect(distributedInferenceWorkload.spec.leader.command).toBe(leaderDefaults.command);
                    expect(distributedInferenceWorkload.spec.leader.args).toBe(leaderDefaults.args);
                    expect(distributedInferenceWorkload.spec.leader.environmentVariables).toEqual([
                        {
                            name: leaderDefaults.environmentVariables.instances[0].name,
                            value: leaderDefaults.environmentVariables.instances[0].value
                        }
                    ]);
                    expect(distributedInferenceWorkload.spec.leader.security.capabilities).toEqual(
                        leaderDefaults.security.capabilities
                    );
                    expect(distributedInferenceWorkload.spec.leader.security.runAsUid).toBe(
                        leaderDefaults.security.runAsUid
                    );
                    expect(distributedInferenceWorkload.spec.leader.security.runAsGid).toBe(
                        leaderDefaults.security.runAsGid
                    );
                    expect(distributedInferenceWorkload.spec.leader.security.supplementalGroups).toBe(
                        leaderDefaults.security.supplementalGroups
                    );

                    // Check worker spec (different from leader)
                    expect(distributedInferenceWorkload.spec.worker.image).toBe(workerDefaults.image);
                    expect(distributedInferenceWorkload.spec.worker.command).toBe(workerDefaults.command);
                    expect(distributedInferenceWorkload.spec.worker.args).toBe(workerDefaults.args);
                    expect(distributedInferenceWorkload.spec.worker.environmentVariables).toEqual([
                        {
                            name: workerDefaults.environmentVariables.instances[0].name,
                            value: workerDefaults.environmentVariables.instances[0].value
                        }
                    ]);
                    expect(distributedInferenceWorkload.spec.worker.security.capabilities).toEqual(
                        workerDefaults.security.capabilities
                    );
                    expect(distributedInferenceWorkload.spec.worker.security.runAsUid).toBe(
                        workerDefaults.security.runAsUid
                    );
                    expect(distributedInferenceWorkload.spec.worker.security.runAsGid).toBe(
                        workerDefaults.security.runAsGid
                    );
                    expect(distributedInferenceWorkload.spec.worker.security.supplementalGroups).toBe(
                        workerDefaults.security.supplementalGroups
                    );

                    logger.info(
                        'test distributed inference policy environment section with different leader and worker passed!'
                    );
                });
            });

            test.describe('Test compute section', function () {
                // Helper function to convert cores to millicores for UI validation
                function coresToMillicores(cores: number): string {
                    return (cores * 1000).toString();
                }

                // Helper function to extract numeric value from memory format (e.g., '100M' -> '100')
                function memoryToUIValue(memory: string): string {
                    const match = memory.match(/^(\d+)/);
                    return match ? match[1] : memory;
                }

                test('distributed inference compute policy defaults - leader and worker similar', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('compute-policy');
                    const commonComputeDefaults = {
                        gpuDevicesRequest: 1,
                        cpuCoreRequest: 0.5,
                        cpuMemoryRequest: '100M'
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: {
                            image: 'ubuntu',
                            compute: commonComputeDefaults
                        },
                        worker: {
                            image: 'ubuntu',
                            compute: commonComputeDefaults
                        }
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        undefined,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);

                    // Validate leader compute defaults
                    const hybridComputeSection = hybridForm.hybridComputeSection;
                    const leaderComputeDefaults = policyDefaults.leader.compute;

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.gpuDevicesInput,
                        leaderComputeDefaults.gpuDevicesRequest.toString()
                    );
                    await hybridForm.toggleFoldingSubSection('.cpu-resources-section', true);
                    // UI displays CPU cores in millicores (0.5 cores = 500 millicores)
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.cpuCoresRequestInput,
                        coresToMillicores(leaderComputeDefaults.cpuCoreRequest)
                    );
                    // UI displays CPU memory as numeric value (100M -> 100)
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.cpuMemoryRequestInput,
                        memoryToUIValue(leaderComputeDefaults.cpuMemoryRequest)
                    );

                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn(); // next step - worker
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker compute defaults (same as leader)
                    const workerComputeDefaults = policyDefaults.worker.compute;
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.gpuDevicesInput,
                        workerComputeDefaults.gpuDevicesRequest.toString()
                    );
                    await hybridForm.toggleFoldingSubSection('.cpu-resources-section', true);
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.cpuCoresRequestInput,
                        coresToMillicores(workerComputeDefaults.cpuCoreRequest)
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.cpuMemoryRequestInput,
                        memoryToUIValue(workerComputeDefaults.cpuMemoryRequest)
                    );

                    // Go to workload setup step and submit
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn(); // next step - workload setup
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Check leader compute spec
                    expect(distributedInferenceWorkload.spec.leader.compute.gpuDevicesRequest).toBe(
                        leaderComputeDefaults.gpuDevicesRequest
                    );
                    expect(distributedInferenceWorkload.spec.leader.compute.cpuCoreRequest).toBe(
                        leaderComputeDefaults.cpuCoreRequest
                    );
                    expect(distributedInferenceWorkload.spec.leader.compute.cpuMemoryRequest).toBe(
                        leaderComputeDefaults.cpuMemoryRequest
                    );

                    // Check worker compute spec (same as leader)
                    expect(distributedInferenceWorkload.spec.worker.compute.gpuDevicesRequest).toBe(
                        workerComputeDefaults.gpuDevicesRequest
                    );
                    expect(distributedInferenceWorkload.spec.worker.compute.cpuCoreRequest).toBe(
                        workerComputeDefaults.cpuCoreRequest
                    );
                    expect(distributedInferenceWorkload.spec.worker.compute.cpuMemoryRequest).toBe(
                        workerComputeDefaults.cpuMemoryRequest
                    );

                    logger.info(
                        'test distributed inference policy compute section with similar leader and worker passed!'
                    );
                });

                test('distributed inference compute policy defaults - leader and worker different', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('compute-policy-diff');
                    const leaderComputeDefaults = {
                        gpuDevicesRequest: 2,
                        cpuCoreRequest: 1.0,
                        cpuMemoryRequest: '200M'
                    };

                    const workerComputeDefaults = {
                        gpuDevicesRequest: 1,
                        cpuCoreRequest: 0.5,
                        cpuMemoryRequest: '100M'
                    };

                    const policyDefaults = {
                        workers: 3,
                        servingPort: { port: 3131 },
                        leader: {
                            image: 'ubuntu',
                            compute: leaderComputeDefaults
                        },
                        worker: {
                            image: 'ubuntu',
                            compute: workerComputeDefaults
                        }
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        undefined,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Validate leader compute defaults
                    const hybridComputeSection = hybridForm.hybridComputeSection;

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.gpuDevicesInput,
                        leaderComputeDefaults.gpuDevicesRequest.toString()
                    );
                    await hybridForm.toggleFoldingSubSection('.cpu-resources-section', true);
                    // UI displays CPU cores in millicores
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.cpuCoresRequestInput,
                        coresToMillicores(leaderComputeDefaults.cpuCoreRequest)
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.cpuMemoryRequestInput,
                        memoryToUIValue(leaderComputeDefaults.cpuMemoryRequest)
                    );

                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn(); // next step - worker
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker compute defaults (different from leader)
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.workersInput,
                        policyDefaults.workers.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.gpuDevicesInput,
                        workerComputeDefaults.gpuDevicesRequest.toString()
                    );
                    await hybridForm.toggleFoldingSubSection('.cpu-resources-section', true);
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.cpuCoresRequestInput,
                        coresToMillicores(workerComputeDefaults.cpuCoreRequest)
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.cpuMemoryRequestInput,
                        memoryToUIValue(workerComputeDefaults.cpuMemoryRequest)
                    );

                    // Go to workload setup step and submit
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn(); // next step - workload setup
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Check leader compute spec
                    expect(distributedInferenceWorkload.spec.leader.compute.gpuDevicesRequest).toBe(
                        leaderComputeDefaults.gpuDevicesRequest
                    );
                    expect(distributedInferenceWorkload.spec.leader.compute.cpuCoreRequest).toBe(
                        leaderComputeDefaults.cpuCoreRequest
                    );
                    expect(distributedInferenceWorkload.spec.leader.compute.cpuMemoryRequest).toBe(
                        leaderComputeDefaults.cpuMemoryRequest
                    );

                    // Check worker compute spec (different from leader)
                    expect(distributedInferenceWorkload.spec.worker.compute.gpuDevicesRequest).toBe(
                        workerComputeDefaults.gpuDevicesRequest
                    );
                    expect(distributedInferenceWorkload.spec.worker.compute.cpuCoreRequest).toBe(
                        workerComputeDefaults.cpuCoreRequest
                    );
                    expect(distributedInferenceWorkload.spec.worker.compute.cpuMemoryRequest).toBe(
                        workerComputeDefaults.cpuMemoryRequest
                    );

                    expect(distributedInferenceWorkload.spec.workers).toBe(policyDefaults.workers);

                    logger.info(
                        'test distributed inference policy compute section with different leader and worker passed!'
                    );
                });
            });

            test.describe('Test storage section', function () {
                test('distributed inference storage policy defaults - leader and worker similar', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('storage-policy');

                    const commonStorageDefaults = {
                        image: 'ubuntu',
                        storage: {
                            pvc: {
                                instances: [
                                    {
                                        claimName: 'common-claim-storage',
                                        path: '/home/storage',
                                        claimInfo: {
                                            storageClass: 'standard',
                                            size: '1G',
                                            accessModes: {
                                                readWriteOnce: true,
                                                readOnlyMany: false,
                                                readWriteMany: false
                                            },
                                            volumeMode: PvcVolumeMode.Filesystem
                                        }
                                    }
                                ]
                            }
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: commonStorageDefaults,
                        worker: commonStorageDefaults
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        undefined,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Validate leader storage defaults
                    const hybridDataSourceSection = hybridForm.hybridDataSourceSection;
                    const pvcDefaults = policyDefaults.leader.storage.pvc.instances[0];

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.storageClass,
                        pvcDefaults.claimInfo.storageClass
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.accessMode,
                        'Read-write by one node'
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.claimSize,
                        '1'
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.unit,
                        'GB'
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.volumeMode,
                        pvcDefaults.claimInfo.volumeMode.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.path,
                        pvcDefaults.path
                    );

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Submit the workload
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify leader storage
                    const leaderPvc = distributedInferenceWorkload.spec.leader.storage.pvc[0];
                    expect(leaderPvc.path).toBe(pvcDefaults.path);
                    expect(leaderPvc.claimInfo.storageClass).toBe(pvcDefaults.claimInfo.storageClass);
                    expect(leaderPvc.claimInfo.size).toBe(pvcDefaults.claimInfo.size);
                    expect(leaderPvc.claimInfo.volumeMode).toBe(pvcDefaults.claimInfo.volumeMode);
                    expect(leaderPvc.claimInfo.accessModes).toEqual({
                        readWriteOnce: true,
                        readOnlyMany: false,
                        readWriteMany: false
                    });

                    // Verify worker storage (should be the same)
                    const workerPvc = distributedInferenceWorkload.spec.worker.storage.pvc[0];
                    expect(workerPvc.path).toBe(pvcDefaults.path);
                    expect(workerPvc.claimInfo.storageClass).toBe(pvcDefaults.claimInfo.storageClass);
                    expect(workerPvc.claimInfo.size).toBe(pvcDefaults.claimInfo.size);
                    expect(workerPvc.claimInfo.volumeMode).toBe(pvcDefaults.claimInfo.volumeMode);
                    expect(workerPvc.claimInfo.accessModes).toEqual({
                        readWriteOnce: true,
                        readOnlyMany: false,
                        readWriteMany: false
                    });

                    logger.info(
                        'test distributed inference policy storage section with similar leader and worker passed!'
                    );
                });

                test('distributed inference storage policy defaults - leader and worker different', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('storage-policy-diff');

                    const leaderStorageDefaults = {
                        image: 'ubuntu',
                        storage: {
                            pvc: {
                                instances: [
                                    {
                                        claimName: 'leader-claim-storage',
                                        path: '/home/leader-storage',
                                        claimInfo: {
                                            storageClass: 'standard',
                                            size: '2G',
                                            accessModes: {
                                                readWriteOnce: true,
                                                readOnlyMany: false,
                                                readWriteMany: false
                                            },
                                            volumeMode: PvcVolumeMode.Filesystem
                                        }
                                    }
                                ]
                            }
                        }
                    };

                    const workerStorageDefaults = {
                        image: 'ubuntu',
                        storage: {
                            pvc: {
                                instances: [
                                    {
                                        claimName: 'worker-claim-storage',
                                        path: '/home/worker-storage',
                                        claimInfo: {
                                            storageClass: 'standard',
                                            size: '1G',
                                            accessModes: {
                                                readWriteOnce: false,
                                                readOnlyMany: true,
                                                readWriteMany: false
                                            },
                                            volumeMode: PvcVolumeMode.Filesystem
                                        }
                                    }
                                ]
                            }
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: leaderStorageDefaults,
                        worker: workerStorageDefaults
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        undefined,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Validate leader storage defaults
                    const hybridDataSourceSection = hybridForm.hybridDataSourceSection;
                    const leaderPvcDefaults = policyDefaults.leader.storage.pvc.instances[0];

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.storageClass,
                        leaderPvcDefaults.claimInfo.storageClass
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.accessMode,
                        'Read-write by one node'
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.claimSize,
                        '2'
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.unit,
                        'GB'
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.volumeMode,
                        leaderPvcDefaults.claimInfo.volumeMode.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.path,
                        leaderPvcDefaults.path
                    );

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker storage defaults (should be different)
                    const workerPvcDefaults = policyDefaults.worker.storage.pvc.instances[0];

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.storageClass,
                        workerPvcDefaults.claimInfo.storageClass
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.accessMode,
                        'Read-only by many nodes'
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.claimSize,
                        '1'
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.unit,
                        'GB'
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.volumeMode,
                        workerPvcDefaults.claimInfo.volumeMode.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.path,
                        workerPvcDefaults.path
                    );

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Submit the workload
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify leader storage
                    const leaderPvc = distributedInferenceWorkload.spec.leader.storage.pvc[0];
                    expect(leaderPvc.path).toBe(leaderPvcDefaults.path);
                    expect(leaderPvc.claimInfo.storageClass).toBe(leaderPvcDefaults.claimInfo.storageClass);
                    expect(leaderPvc.claimInfo.size).toBe(leaderPvcDefaults.claimInfo.size);
                    expect(leaderPvc.claimInfo.volumeMode).toBe(leaderPvcDefaults.claimInfo.volumeMode);
                    expect(leaderPvc.claimInfo.accessModes).toEqual({
                        readWriteOnce: true,
                        readOnlyMany: false,
                        readWriteMany: false
                    });

                    // Verify worker storage (should be different)
                    const workerPvc = distributedInferenceWorkload.spec.worker.storage.pvc[0];
                    expect(workerPvc.path).toBe(workerPvcDefaults.path);
                    expect(workerPvc.claimInfo.storageClass).toBe(workerPvcDefaults.claimInfo.storageClass);
                    expect(workerPvc.claimInfo.size).toBe(workerPvcDefaults.claimInfo.size);
                    expect(workerPvc.claimInfo.volumeMode).toBe(workerPvcDefaults.claimInfo.volumeMode);
                    expect(workerPvc.claimInfo.accessModes).toEqual({
                        readWriteOnce: false,
                        readOnlyMany: true,
                        readWriteMany: false
                    });

                    logger.info(
                        'test distributed inference policy storage section with different leader and worker passed!'
                    );
                });
            });

            test.describe('Test general section', function () {
                test('distributed inference general policy defaults - leader and worker similar', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('general-policy');

                    const commonGeneralDefaults = {
                        image: 'ubuntu',
                        annotations: {
                            instances: [
                                {
                                    name: 'common-annotation',
                                    value: 'common-annotation-value'
                                }
                            ]
                        },
                        labels: {
                            instances: [
                                {
                                    name: 'common-label',
                                    value: 'common-label-value'
                                }
                            ]
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: commonGeneralDefaults,
                        worker: commonGeneralDefaults
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        undefined,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Validate leader general defaults - open general section first
                    const hybridGeneralSection = hybridForm.hybridGeneralSection;
                    await hybridForm.locator(hybridGeneralSection.sectionSelector).click();

                    const leaderDefaults = policyDefaults.leader;

                    await elementValidationUtils.expectInputValueToEqual(
                        '.annotations-section [aid="itemized-name-input"]',
                        leaderDefaults.annotations.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.annotations-section [aid="itemized-value-input"]',
                        leaderDefaults.annotations.instances[0].value
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.labels-section [aid="itemized-name-input"]',
                        leaderDefaults.labels.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.labels-section [aid="itemized-value-input"]',
                        leaderDefaults.labels.instances[0].value
                    );

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker general defaults (same as leader) - open general section
                    await hybridForm.locator(hybridGeneralSection.sectionSelector).click();

                    const workerDefaults = policyDefaults.worker;

                    await elementValidationUtils.expectInputValueToEqual(
                        '.annotations-section [aid="itemized-name-input"]',
                        workerDefaults.annotations.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.annotations-section [aid="itemized-value-input"]',
                        workerDefaults.annotations.instances[0].value
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.labels-section [aid="itemized-name-input"]',
                        workerDefaults.labels.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.labels-section [aid="itemized-value-input"]',
                        workerDefaults.labels.instances[0].value
                    );

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Submit the workload
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify leader annotations and labels
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.leader.annotations,
                            leaderDefaults.annotations.instances[0].name,
                            leaderDefaults.annotations.instances[0].value
                        )
                    ).toBe(true);
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.leader.labels,
                            leaderDefaults.labels.instances[0].name,
                            leaderDefaults.labels.instances[0].value
                        )
                    ).toBe(true);

                    // Verify worker annotations and labels (same as leader)
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.worker.annotations,
                            workerDefaults.annotations.instances[0].name,
                            workerDefaults.annotations.instances[0].value
                        )
                    ).toBe(true);
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.worker.labels,
                            workerDefaults.labels.instances[0].name,
                            workerDefaults.labels.instances[0].value
                        )
                    ).toBe(true);

                    logger.info(
                        'test distributed inference policy general section with similar leader and worker passed!'
                    );
                });

                test('distributed inference general policy defaults - leader and worker different', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('general-policy-diff');

                    const leaderGeneralDefaults = {
                        image: 'ubuntu',
                        annotations: {
                            instances: [
                                {
                                    name: 'leader-annotation',
                                    value: 'leader-annotation-value'
                                }
                            ]
                        },
                        labels: {
                            instances: [
                                {
                                    name: 'leader-label',
                                    value: 'leader-label-value'
                                }
                            ]
                        }
                    };

                    const workerGeneralDefaults = {
                        image: 'ubuntu',
                        annotations: {
                            instances: [
                                {
                                    name: 'worker-annotation',
                                    value: 'worker-annotation-value'
                                }
                            ]
                        },
                        labels: {
                            instances: [
                                {
                                    name: 'worker-label',
                                    value: 'worker-label-value'
                                }
                            ]
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: leaderGeneralDefaults,
                        worker: workerGeneralDefaults
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        undefined,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();

                    // Validate leader general defaults - open general section first
                    const hybridGeneralSection = hybridForm.hybridGeneralSection;
                    await hybridForm.locator(hybridGeneralSection.sectionSelector).click();

                    const leaderDefaults = policyDefaults.leader;

                    await elementValidationUtils.expectInputValueToEqual(
                        '.annotations-section [aid="itemized-name-input"]',
                        leaderDefaults.annotations.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.annotations-section [aid="itemized-value-input"]',
                        leaderDefaults.annotations.instances[0].value
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.labels-section [aid="itemized-name-input"]',
                        leaderDefaults.labels.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.labels-section [aid="itemized-value-input"]',
                        leaderDefaults.labels.instances[0].value
                    );

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker general defaults (different from leader) - open general section
                    await hybridForm.locator(hybridGeneralSection.sectionSelector).click();

                    const workerDefaults = policyDefaults.worker;

                    await elementValidationUtils.expectInputValueToEqual(
                        '.annotations-section [aid="itemized-name-input"]',
                        workerDefaults.annotations.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.annotations-section [aid="itemized-value-input"]',
                        workerDefaults.annotations.instances[0].value
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.labels-section [aid="itemized-name-input"]',
                        workerDefaults.labels.instances[0].name
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        '.labels-section [aid="itemized-value-input"]',
                        workerDefaults.labels.instances[0].value
                    );

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Submit the workload
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify leader annotations and labels
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.leader.annotations,
                            leaderDefaults.annotations.instances[0].name,
                            leaderDefaults.annotations.instances[0].value
                        )
                    ).toBe(true);
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.leader.labels,
                            leaderDefaults.labels.instances[0].name,
                            leaderDefaults.labels.instances[0].value
                        )
                    ).toBe(true);

                    // Verify worker annotations and labels (different from leader)
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.worker.annotations,
                            workerDefaults.annotations.instances[0].name,
                            workerDefaults.annotations.instances[0].value
                        )
                    ).toBe(true);
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.worker.labels,
                            workerDefaults.labels.instances[0].name,
                            workerDefaults.labels.instances[0].value
                        )
                    ).toBe(true);

                    logger.info(
                        'test distributed inference policy general section with different leader and worker passed!'
                    );
                });
            });
        });

        test.describe('Test inference policy rules', function () {
            test.describe('Test environment section', function () {
                test('distributed inference environment policy rules - leader and worker similar', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('env-rules-policy');

                    const commonDefaults = {
                        image: 'ubuntu',
                        command: 'test-command',
                        args: '--arg1 --arg2',
                        environmentVariables: {
                            instances: [
                                {
                                    name: 'test-env-var',
                                    value: 'test-env-value'
                                }
                            ]
                        },
                        security: {
                            uidGidSource: UidGidSource.Custom,
                            runAsUid: 1000,
                            runAsGid: 1000,
                            supplementalGroups: '1000,1001',
                            capabilities: [Capability.NetAdmin]
                        }
                    };

                    const commonRules = {
                        image: {
                            canEdit: false
                        },
                        command: {
                            canEdit: false
                        },
                        args: {
                            canEdit: false
                        },
                        environmentVariables: {
                            instances: {
                                canAdd: false,
                                locked: ['test-env-var']
                            }
                        },
                        security: {
                            runAsUid: {
                                canEdit: false
                            },
                            runAsGid: {
                                canEdit: false
                            },
                            supplementalGroups: {
                                canEdit: false
                            },
                            capabilities: {
                                canEdit: false
                            }
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: commonDefaults,
                        worker: commonDefaults
                    };

                    const policyRules = {
                        leader: commonRules,
                        worker: commonRules
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        policyRules,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Validate leader environment fields are disabled
                    const hybridFormPage = hybridForm.page;
                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;

                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.hybridImageInput)
                    ).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.cmdInput)).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.argsInput)).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.envVarNameInput)
                    ).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.envVarValue)).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.envVarRemoveBtn)
                    ).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.uidInput)).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.gidInput)).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.supplementalGroupsInput)
                    ).toBeDisabled();

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker environment fields are disabled (same as leader)
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.hybridImageInput)
                    ).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.cmdInput)).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.argsInput)).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.envVarNameInput)
                    ).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.envVarValue)).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.envVarRemoveBtn)
                    ).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.uidInput)).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.gidInput)).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.supplementalGroupsInput)
                    ).toBeDisabled();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Submit the workload
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Check leader spec matches policy defaults
                    const leaderDefaults = policyDefaults.leader;
                    expect(distributedInferenceWorkload.spec.leader.image).toBe(leaderDefaults.image);
                    expect(distributedInferenceWorkload.spec.leader.command).toBe(leaderDefaults.command);
                    expect(distributedInferenceWorkload.spec.leader.args).toBe(leaderDefaults.args);
                    expect(distributedInferenceWorkload.spec.leader.security.runAsUid).toBe(
                        leaderDefaults.security.runAsUid
                    );
                    expect(distributedInferenceWorkload.spec.leader.security.runAsGid).toBe(
                        leaderDefaults.security.runAsGid
                    );
                    expect(distributedInferenceWorkload.spec.leader.security.supplementalGroups).toBe(
                        leaderDefaults.security.supplementalGroups
                    );

                    // Check worker spec matches policy defaults (same as leader)
                    const workerDefaults = policyDefaults.worker;
                    expect(distributedInferenceWorkload.spec.worker.image).toBe(workerDefaults.image);
                    expect(distributedInferenceWorkload.spec.worker.command).toBe(workerDefaults.command);
                    expect(distributedInferenceWorkload.spec.worker.args).toBe(workerDefaults.args);
                    expect(distributedInferenceWorkload.spec.worker.security.runAsUid).toBe(
                        workerDefaults.security.runAsUid
                    );
                    expect(distributedInferenceWorkload.spec.worker.security.runAsGid).toBe(
                        workerDefaults.security.runAsGid
                    );
                    expect(distributedInferenceWorkload.spec.worker.security.supplementalGroups).toBe(
                        workerDefaults.security.supplementalGroups
                    );

                    logger.info(
                        'test distributed inference policy environment rules with similar leader and worker passed!'
                    );
                });

                test('distributed inference environment policy rules - leader and worker different', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('env-rules-policy-diff');

                    const leaderDefaults = {
                        image: 'ubuntu',
                        command: 'leader-command',
                        args: '--leader-arg1',
                        environmentVariables: {
                            instances: [
                                {
                                    name: 'leader-env-var',
                                    value: 'leader-env-value'
                                }
                            ]
                        },
                        security: {
                            uidGidSource: UidGidSource.Custom,
                            runAsUid: 1000,
                            runAsGid: 1000,
                            supplementalGroups: '1000,1001',
                            capabilities: [Capability.NetAdmin]
                        }
                    };

                    const workerDefaults = {
                        image: 'alpine',
                        command: 'worker-command',
                        args: '--worker-arg1',
                        environmentVariables: {
                            instances: [
                                {
                                    name: 'worker-env-var',
                                    value: 'worker-env-value'
                                }
                            ]
                        },
                        security: {
                            uidGidSource: UidGidSource.Custom,
                            runAsUid: 2000,
                            runAsGid: 2000,
                            supplementalGroups: '2000,2001',
                            capabilities: [Capability.SysAdmin]
                        }
                    };

                    // Leader: image and command locked
                    const leaderRules = {
                        image: {
                            canEdit: false
                        },
                        command: {
                            canEdit: false
                        },
                        environmentVariables: {
                            instances: {
                                canAdd: false,
                                locked: ['leader-env-var']
                            }
                        }
                    };

                    // Worker: args and security locked (different from leader)
                    const workerRules = {
                        args: {
                            canEdit: false
                        },
                        environmentVariables: {
                            instances: {
                                canAdd: false,
                                locked: ['worker-env-var']
                            }
                        },
                        security: {
                            runAsUid: {
                                canEdit: false
                            },
                            runAsGid: {
                                canEdit: false
                            }
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: leaderDefaults,
                        worker: workerDefaults
                    };

                    const policyRules = {
                        leader: leaderRules,
                        worker: workerRules,
                        servingPort: {
                            port: {
                                canEdit: false
                            }
                        }
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        policyRules,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();

                    // Validate leader environment fields - image and command disabled, args enabled
                    const hybridFormPage = hybridForm.page;
                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;

                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.hybridImageInput)
                    ).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.cmdInput)).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.argsInput)
                    ).not.toBeDisabled(); // args is editable for leader
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.envVarNameInput)
                    ).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.envVarValue)).toBeDisabled();

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker environment fields - args and security disabled, image/command enabled
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.hybridImageInput)
                    ).not.toBeDisabled(); // image is editable for worker
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.cmdInput)
                    ).not.toBeDisabled(); // command is editable for worker
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.argsInput)).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.envVarNameInput)
                    ).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.envVarValue)).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.uidInput)).toBeDisabled();
                    await expect(hybridFormPage.locator(hybridEnvironmentSection.gidInput)).toBeDisabled();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();
                    // Submit the workload
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Check leader spec - locked fields should match policy defaults
                    expect(distributedInferenceWorkload.spec.leader.image).toBe(leaderDefaults.image);
                    expect(distributedInferenceWorkload.spec.leader.command).toBe(leaderDefaults.command);

                    // Check worker spec - locked fields should match policy defaults
                    expect(distributedInferenceWorkload.spec.worker.args).toBe(workerDefaults.args);
                    expect(distributedInferenceWorkload.spec.worker.security.runAsUid).toBe(
                        workerDefaults.security.runAsUid
                    );
                    expect(distributedInferenceWorkload.spec.worker.security.runAsGid).toBe(
                        workerDefaults.security.runAsGid
                    );

                    logger.info(
                        'test distributed inference policy environment rules with different leader and worker passed!'
                    );
                });
            });

            test.describe('Test compute section', function () {
                test('distributed inference compute policy rules - leader and worker similar', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('compute-rules-policy');

                    const commonComputeDefaults = {
                        gpuDevicesRequest: 1,
                        cpuCoreRequest: 0.5,
                        cpuMemoryRequest: '100M'
                    };

                    const commonComputeRules = {
                        compute: {
                            gpuDevicesRequest: {
                                canEdit: false
                            },
                            cpuCoreRequest: {
                                canEdit: false
                            },
                            cpuMemoryRequest: {
                                canEdit: false
                            }
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: {
                            image: 'ubuntu',
                            compute: commonComputeDefaults
                        },
                        worker: {
                            image: 'ubuntu',
                            compute: commonComputeDefaults
                        }
                    };

                    const policyRules = {
                        leader: commonComputeRules,
                        worker: commonComputeRules,
                        servingPort: {
                            port: {
                                canEdit: false
                            }
                        }
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        policyRules,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Validate leader compute fields are disabled
                    const hybridFormPage = hybridForm.page;
                    const hybridComputeSection = hybridForm.hybridComputeSection;

                    await expect(hybridFormPage.locator(hybridComputeSection.gpuDevicesInput)).toBeDisabled();

                    await hybridForm.toggleFoldingSubSection('.cpu-resources-section', true);
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.cpuCoresRequestInput)
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.cpuMemoryRequestInput)
                    ).toBeDisabled();

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker compute fields are disabled (same as leader)
                    await expect(hybridFormPage.locator(hybridComputeSection.gpuDevicesInput)).toBeDisabled();

                    await hybridForm.toggleFoldingSubSection('.cpu-resources-section', true);
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.cpuCoresRequestInput)
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.cpuMemoryRequestInput)
                    ).toBeDisabled();

                    // Go to workload setup step and submit
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Check leader compute spec - locked fields should match policy defaults
                    expect(distributedInferenceWorkload.spec.leader.compute.gpuDevicesRequest).toBe(
                        commonComputeDefaults.gpuDevicesRequest
                    );
                    expect(distributedInferenceWorkload.spec.leader.compute.cpuCoreRequest).toBe(
                        commonComputeDefaults.cpuCoreRequest
                    );
                    expect(distributedInferenceWorkload.spec.leader.compute.cpuMemoryRequest).toBe(
                        commonComputeDefaults.cpuMemoryRequest
                    );

                    // Check worker compute spec - locked fields should match policy defaults
                    expect(distributedInferenceWorkload.spec.worker.compute.gpuDevicesRequest).toBe(
                        commonComputeDefaults.gpuDevicesRequest
                    );
                    expect(distributedInferenceWorkload.spec.worker.compute.cpuCoreRequest).toBe(
                        commonComputeDefaults.cpuCoreRequest
                    );
                    expect(distributedInferenceWorkload.spec.worker.compute.cpuMemoryRequest).toBe(
                        commonComputeDefaults.cpuMemoryRequest
                    );

                    logger.info(
                        'test distributed inference policy compute rules with similar leader and worker passed!'
                    );
                });

                test('distributed inference compute policy rules - leader and worker different', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('compute-rules-policy-diff');

                    const leaderComputeDefaults = {
                        gpuDevicesRequest: 2,
                        cpuCoreRequest: 1.0,
                        cpuMemoryRequest: '200M'
                    };

                    const workerComputeDefaults = {
                        gpuDevicesRequest: 1,
                        cpuCoreRequest: 0.5,
                        cpuMemoryRequest: '100M'
                    };

                    // Leader: GPU devices locked, CPU cores and memory editable
                    const leaderComputeRules = {
                        compute: {
                            gpuDevicesRequest: {
                                canEdit: false
                            }
                        }
                    };

                    // Worker: CPU cores and memory locked, GPU devices editable
                    const workerComputeRules = {
                        compute: {
                            cpuCoreRequest: {
                                canEdit: false
                            },
                            cpuMemoryRequest: {
                                canEdit: false
                            }
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: {
                            image: 'ubuntu',
                            compute: leaderComputeDefaults
                        },
                        worker: {
                            image: 'ubuntu1',
                            compute: workerComputeDefaults
                        }
                    };

                    const policyRules = {
                        leader: leaderComputeRules,
                        worker: workerComputeRules,
                        servingPort: {
                            port: {
                                canEdit: false
                            }
                        },
                        workers: {
                            canEdit: false
                        }
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        policyRules,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Validate leader compute fields - GPU locked, CPU cores/memory editable
                    const hybridFormPage = hybridForm.page;
                    const hybridComputeSection = hybridForm.hybridComputeSection;

                    await expect(hybridFormPage.locator(hybridComputeSection.gpuDevicesInput)).toBeDisabled();

                    await hybridForm.toggleFoldingSubSection('.cpu-resources-section', true);
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.cpuCoresRequestInput)
                    ).not.toBeDisabled(); // editable for leader
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.cpuMemoryRequestInput)
                    ).not.toBeDisabled(); // editable for leader

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    await expect(hybridFormPage.locator(hybridComputeSection.workersInput)).toBeDisabled(); // editable for worker

                    // Validate worker compute fields - GPU editable, CPU cores/memory locked
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.gpuDevicesInput)
                    ).not.toBeDisabled(); // editable for worker

                    await hybridForm.toggleFoldingSubSection('.cpu-resources-section', true);
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.cpuCoresRequestInput)
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.cpuMemoryRequestInput)
                    ).toBeDisabled();

                    // Go to workload setup step and submit
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Submit the workload
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Check leader compute spec - GPU locked should match policy defaults
                    expect(distributedInferenceWorkload.spec.leader.compute.gpuDevicesRequest).toBe(
                        leaderComputeDefaults.gpuDevicesRequest
                    );

                    // Check worker compute spec - CPU locked should match policy defaults
                    expect(distributedInferenceWorkload.spec.worker.compute.cpuCoreRequest).toBe(
                        workerComputeDefaults.cpuCoreRequest
                    );
                    expect(distributedInferenceWorkload.spec.worker.compute.cpuMemoryRequest).toBe(
                        workerComputeDefaults.cpuMemoryRequest
                    );

                    logger.info(
                        'test distributed inference policy compute rules with different leader and worker passed!'
                    );
                });
            });

            test.describe('Test storage section', function () {
                test('distributed inference storage policy rules - leader and worker similar', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('storage-rules-policy');

                    const commonStorageAttributes = {
                        path: '/home/storage',
                        claimInfo: {
                            storageClass: 'standard',
                            size: '1G',
                            accessModes: {
                                readWriteOnce: true,
                                readOnlyMany: false,
                                readWriteMany: false
                            },
                            volumeMode: PvcVolumeMode.Filesystem
                        }
                    };

                    const commonStorageDefaults = {
                        image: 'ubuntu',
                        storage: {
                            pvc: {
                                attributes: commonStorageAttributes,
                                instances: [
                                    {
                                        claimName: 'common-claim-storage',
                                        ...commonStorageAttributes
                                    }
                                ]
                            }
                        }
                    };

                    const commonStorageRules = {
                        storage: {
                            pvc: {
                                attributes: {
                                    path: {
                                        canEdit: false
                                    },
                                    claimInfo: {
                                        storageClass: {
                                            canEdit: false
                                        },
                                        size: {
                                            canEdit: false
                                        },
                                        accessModes: {
                                            readWriteOnce: {
                                                canEdit: false
                                            },
                                            readOnlyMany: {
                                                canEdit: false
                                            },
                                            readWriteMany: {
                                                canEdit: false
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: commonStorageDefaults,
                        worker: commonStorageDefaults
                    };

                    const policyRules = {
                        leader: commonStorageRules,
                        worker: commonStorageRules
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        policyRules,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Validate leader storage fields are disabled
                    const hybridFormPage = hybridForm.page;
                    const hybridDataSourceSection = hybridForm.hybridDataSourceSection;

                    await expect(hybridFormPage.locator('.storage-class-select')).toHaveAttribute(
                        'aria-disabled'
                    );
                    await expect(hybridFormPage.locator(hybridDataSourceSection.pvc.path)).toBeDisabled();

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker storage fields are disabled (same as leader)
                    await expect(hybridFormPage.locator('.storage-class-select')).toHaveAttribute(
                        'aria-disabled'
                    );
                    await expect(hybridFormPage.locator(hybridDataSourceSection.pvc.path)).toBeDisabled();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Submit the workload
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify leader storage matches policy defaults
                    const leaderPvc = distributedInferenceWorkload.spec.leader.storage.pvc[0];
                    expect(leaderPvc.path).toBe(commonStorageAttributes.path);
                    expect(leaderPvc.claimInfo.storageClass).toBe(
                        commonStorageAttributes.claimInfo.storageClass
                    );
                    expect(leaderPvc.claimInfo.size).toBe(commonStorageAttributes.claimInfo.size);

                    // Verify worker storage matches policy defaults (same as leader)
                    const workerPvc = distributedInferenceWorkload.spec.worker.storage.pvc[0];
                    expect(workerPvc.path).toBe(commonStorageAttributes.path);
                    expect(workerPvc.claimInfo.storageClass).toBe(
                        commonStorageAttributes.claimInfo.storageClass
                    );
                    expect(workerPvc.claimInfo.size).toBe(commonStorageAttributes.claimInfo.size);

                    logger.info(
                        'test distributed inference policy storage rules with similar leader and worker passed!'
                    );
                });

                test('distributed inference storage policy rules - leader and worker different', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('storage-rules-policy-diff');

                    const leaderStorageAttributes = {
                        path: '/home/leader-storage',
                        claimInfo: {
                            storageClass: 'standard',
                            size: '2G',
                            accessModes: {
                                readWriteOnce: true,
                                readOnlyMany: false,
                                readWriteMany: false
                            },
                            volumeMode: PvcVolumeMode.Filesystem
                        }
                    };

                    const workerStorageAttributes = {
                        path: '/home/worker-storage',
                        claimInfo: {
                            storageClass: 'standard',
                            size: '1G',
                            accessModes: {
                                readWriteOnce: true,
                                readOnlyMany: false,
                                readWriteMany: false
                            },
                            volumeMode: PvcVolumeMode.Filesystem
                        }
                    };

                    const leaderStorageDefaults = {
                        image: 'ubuntu',
                        storage: {
                            pvc: {
                                attributes: leaderStorageAttributes,
                                instances: [
                                    {
                                        claimName: 'leader-claim-storage',
                                        ...leaderStorageAttributes
                                    }
                                ]
                            }
                        }
                    };

                    const workerStorageDefaults = {
                        image: 'ubuntu',
                        storage: {
                            pvc: {
                                attributes: workerStorageAttributes,
                                instances: [
                                    {
                                        claimName: 'worker-claim-storage',
                                        ...workerStorageAttributes
                                    }
                                ]
                            }
                        }
                    };

                    // Leader: PVC not locked (can add/edit)
                    const leaderStorageRules = {};

                    // Worker: PVC attributes locked
                    const workerStorageRules = {
                        storage: {
                            pvc: {
                                attributes: {
                                    path: {
                                        canEdit: false
                                    },
                                    claimInfo: {
                                        storageClass: {
                                            canEdit: false
                                        },
                                        size: {
                                            canEdit: false
                                        },
                                        accessModes: {
                                            readWriteOnce: {
                                                canEdit: false
                                            },
                                            readOnlyMany: {
                                                canEdit: false
                                            },
                                            readWriteMany: {
                                                canEdit: false
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: leaderStorageDefaults,
                        worker: workerStorageDefaults
                    };

                    const policyRules = {
                        leader: leaderStorageRules,
                        worker: workerStorageRules
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        policyRules,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();

                    // Validate leader storage fields are enabled (not locked)
                    const hybridFormPage = hybridForm.page;
                    const hybridDataSourceSection = hybridForm.hybridDataSourceSection;

                    await expect(hybridFormPage.locator('.storage-class-select')).not.toHaveAttribute(
                        'aria-disabled'
                    );
                    await expect(hybridFormPage.locator(hybridDataSourceSection.pvc.path)).not.toBeDisabled();

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Validate worker storage fields are disabled (locked)
                    await expect(hybridFormPage.locator('.storage-class-select')).toHaveAttribute(
                        'aria-disabled'
                    );
                    await expect(hybridFormPage.locator(hybridDataSourceSection.pvc.path)).toBeDisabled();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Submit the workload
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify worker storage matches policy defaults (locked)
                    const workerPvc = distributedInferenceWorkload.spec.worker.storage.pvc[0];
                    expect(workerPvc.path).toBe(workerStorageAttributes.path);
                    expect(workerPvc.claimInfo.storageClass).toBe(
                        workerStorageAttributes.claimInfo.storageClass
                    );
                    expect(workerPvc.claimInfo.size).toBe(workerStorageAttributes.claimInfo.size);

                    logger.info(
                        'test distributed inference policy storage rules with different leader and worker passed!'
                    );
                });
            });

            test.describe('Test general section', function () {
                test('distributed inference general policy rules - leader and worker similar', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('general-rules-policy');

                    const commonGeneralDefaults = {
                        image: 'ubuntu',
                        annotations: {
                            instances: [
                                {
                                    name: 'common-annotation',
                                    value: 'common-annotation-value'
                                }
                            ]
                        },
                        labels: {
                            instances: [
                                {
                                    name: 'common-label',
                                    value: 'common-label-value'
                                }
                            ]
                        }
                    };

                    const commonGeneralRules = {
                        annotations: {
                            instances: {
                                canAdd: false,
                                locked: ['common-annotation']
                            }
                        },
                        labels: {
                            instances: {
                                canAdd: false,
                                locked: ['common-label']
                            }
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: commonGeneralDefaults,
                        worker: commonGeneralDefaults
                    };

                    const policyRules = {
                        leader: commonGeneralRules,
                        worker: commonGeneralRules
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        policyRules,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Open general section and validate leader fields are disabled
                    const hybridFormPage = hybridForm.page;
                    const hybridGeneralSection = hybridForm.hybridGeneralSection;
                    await hybridForm.locator(hybridGeneralSection.sectionSelector).click();

                    await expect(
                        hybridFormPage.locator('.annotations-section [aid="itemized-name-input"]')
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.annotations-section [aid="itemized-value-input"]')
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.labels-section [aid="itemized-name-input"]')
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.labels-section [aid="itemized-value-input"]')
                    ).toBeDisabled();

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Open general section and validate worker fields are disabled (same as leader)
                    await hybridForm.locator(hybridGeneralSection.sectionSelector).click();

                    await expect(
                        hybridFormPage.locator('.annotations-section [aid="itemized-name-input"]')
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.annotations-section [aid="itemized-value-input"]')
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.labels-section [aid="itemized-name-input"]')
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.labels-section [aid="itemized-value-input"]')
                    ).toBeDisabled();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Submit the workload
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify leader annotations and labels match policy defaults
                    const leaderDefaults = policyDefaults.leader;
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.leader.annotations,
                            leaderDefaults.annotations.instances[0].name,
                            leaderDefaults.annotations.instances[0].value
                        )
                    ).toBe(true);
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.leader.labels,
                            leaderDefaults.labels.instances[0].name,
                            leaderDefaults.labels.instances[0].value
                        )
                    ).toBe(true);

                    // Verify worker annotations and labels match policy defaults (same as leader)
                    const workerDefaults = policyDefaults.worker;
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.worker.annotations,
                            workerDefaults.annotations.instances[0].name,
                            workerDefaults.annotations.instances[0].value
                        )
                    ).toBe(true);
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.worker.labels,
                            workerDefaults.labels.instances[0].name,
                            workerDefaults.labels.instances[0].value
                        )
                    ).toBe(true);

                    logger.info(
                        'test distributed inference policy general rules with similar leader and worker passed!'
                    );
                });

                test('distributed inference general policy rules - leader and worker different', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('general-rules-policy-diff');

                    const leaderGeneralDefaults = {
                        image: 'ubuntu',
                        annotations: {
                            instances: [
                                {
                                    name: 'leader-annotation',
                                    value: 'leader-annotation-value'
                                }
                            ]
                        },
                        labels: {
                            instances: [
                                {
                                    name: 'leader-label',
                                    value: 'leader-label-value'
                                }
                            ]
                        }
                    };

                    const workerGeneralDefaults = {
                        image: 'ubuntu',
                        annotations: {
                            instances: [
                                {
                                    name: 'worker-annotation',
                                    value: 'worker-annotation-value'
                                }
                            ]
                        },
                        labels: {
                            instances: [
                                {
                                    name: 'worker-label',
                                    value: 'worker-label-value'
                                }
                            ]
                        }
                    };

                    // Leader: annotations locked, labels editable
                    const leaderGeneralRules = {
                        annotations: {
                            instances: {
                                canAdd: false,
                                locked: ['leader-annotation']
                            }
                        }
                    };

                    // Worker: labels locked, annotations editable (different from leader)
                    const workerGeneralRules = {
                        labels: {
                            instances: {
                                canAdd: false,
                                locked: ['worker-label']
                            }
                        }
                    };

                    const policyDefaults = {
                        workers: 1,
                        servingPort: { port: 3131 },
                        leader: leaderGeneralDefaults,
                        worker: workerGeneralDefaults
                    };

                    const policyRules = {
                        leader: leaderGeneralRules,
                        worker: workerGeneralRules
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        policyRules,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Open general section and validate leader fields
                    const hybridFormPage = hybridForm.page;
                    const hybridGeneralSection = hybridForm.hybridGeneralSection;
                    await hybridForm.locator(hybridGeneralSection.sectionSelector).click();

                    // Leader: annotations disabled, labels enabled
                    await expect(
                        hybridFormPage.locator('.annotations-section [aid="itemized-name-input"]')
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.annotations-section [aid="itemized-value-input"]')
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.labels-section [aid="itemized-name-input"]')
                    ).not.toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.labels-section [aid="itemized-value-input"]')
                    ).not.toBeDisabled();

                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Open general section and validate worker fields (different from leader)
                    await hybridForm.locator(hybridGeneralSection.sectionSelector).click();

                    // Worker: labels disabled, annotations enabled
                    await expect(
                        hybridFormPage.locator('.annotations-section [aid="itemized-name-input"]')
                    ).not.toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.annotations-section [aid="itemized-value-input"]')
                    ).not.toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.labels-section [aid="itemized-name-input"]')
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator('.labels-section [aid="itemized-value-input"]')
                    ).toBeDisabled();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Submit the workload
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify leader annotations match policy defaults (locked)
                    const leaderDefaults = policyDefaults.leader;
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.leader.annotations,
                            leaderDefaults.annotations.instances[0].name,
                            leaderDefaults.annotations.instances[0].value
                        )
                    ).toBe(true);

                    // Verify worker labels match policy defaults (locked)
                    const workerDefaults = policyDefaults.worker;
                    expect(
                        findItem(
                            distributedInferenceWorkload.spec.worker.labels,
                            workerDefaults.labels.instances[0].name,
                            workerDefaults.labels.instances[0].value
                        )
                    ).toBe(true);

                    logger.info(
                        'test distributed inference policy general rules with different leader and worker passed!'
                    );
                });
            });
        });

        test.describe('Test workload setup step', function () {
            test.describe('Test inference policy defaults', function () {
                test('distributed inference workload setup policy defaults - all fields', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('workload-setup-policy');

                    const policyDefaults = {
                        workers: 2,
                        replicas: 3,
                        servingPort: { port: 3131 },
                        priorityClass: 'low',
                        startupPolicy: DistributedInferenceStartupPolicy.LeaderReady,
                        leader: {
                            image: 'ubuntu'
                        },
                        worker: {
                            image: 'ubuntu'
                        }
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        {},
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Validate workload setup fields have policy default values
                    const hybridGeneralSection = hybridForm.hybridGeneralSection;

                    // Validate replicas
                    await elementValidationUtils.expectInputValueToEqual(
                        '[aid="replicas-input"]',
                        policyDefaults.replicas.toString()
                    );

                    await inferenceFlatUIUtil.inference.openGeneralSection();

                    // Validate priority
                    await elementValidationUtils.expectElementValueToContain(
                        hybridGeneralSection.prioritySelect,
                        policyDefaults.priorityClass
                    );

                    // Validate startup policy
                    await elementValidationUtils.expectElementValueToContain(
                        hybridGeneralSection.startupPolicy,
                        'Once the leader pod is ready'
                    );

                    // Submit the workload
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify workload setup values
                    expect(distributedInferenceWorkload.spec.priorityClass).toBe(
                        policyDefaults.priorityClass
                    );
                    
                    expect(distributedInferenceWorkload.spec.startupPolicy).toBe(
                        policyDefaults.startupPolicy
                    );
                    expect(distributedInferenceWorkload.spec.replicas).toBe(policyDefaults.replicas);

                    logger.info('test distributed inference workload setup policy defaults passed!');
                });
            });

            test.describe('Test inference policy rules', function () {
                test('distributed inference workload setup policy rules - all fields locked', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('workload-setup-rules-policy');

                    const policyDefaults = {
                        workers: 2,
                        replicas: 3,
                        servingPort: { port: 3131 },
                        priorityClass: 'low',
                        startupPolicy: DistributedInferenceStartupPolicy.LeaderReady,
                        leader: {
                            image: 'ubuntu'
                        },
                        worker: {
                            image: 'ubuntu'
                        }
                    };

                    const policyRules = {
                        priorityClass: { canEdit: false },
                        startupPolicy: { canEdit: false },
                        replicas: { canEdit: false }
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        policyRules,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Validate workload setup fields are disabled
                    const hybridFormPage = hybridForm.page;

                    // Validate replicas is disabled
                    await expect(hybridFormPage.locator('[aid="replicas-input"]')).toBeDisabled();

                    await inferenceFlatUIUtil.inference.openGeneralSection();

                    // Validate priority is disabled (aria-disabled is on the parent label)
                    await expect(hybridFormPage.locator('.priority-select')).toHaveAttribute(
                        'aria-disabled',
                        'true'
                    );

                    // Validate startup policy is disabled (aria-disabled is on the parent label)
                    await expect(hybridFormPage.locator('.startup-policy-select')).toHaveAttribute(
                        'aria-disabled',
                        'true'
                    );

                    // Submit the workload
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify workload setup values match policy defaults (since they were locked)
                    expect(distributedInferenceWorkload.spec.priorityClass).toBe(
                        policyDefaults.priorityClass
                    );
                    expect(distributedInferenceWorkload.spec.startupPolicy).toBe(
                        policyDefaults.startupPolicy
                    );
                    expect(distributedInferenceWorkload.spec.replicas).toBe(policyDefaults.replicas);

                    logger.info('test distributed inference workload setup policy rules passed!');
                });
            });
        });
    });

    supportedFrom(SupportedVersions.V2_24)(function () {
        // this test is similar to the one above, but with the new preemptibility field that supports only in V2_24+
        test.describe('Test workload setup step', function () {
            test.describe('Test inference policy defaults with preemptibility', function () {
                test('distributed inference workload setup policy defaults - all fields including preemptibility', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('workload-setup-policy');

                    const policyDefaults = {
                        workers: 2,
                        replicas: 3,
                        servingPort: { port: 3131 },
                        priorityClass: 'low',
                        preemptibility: 'preemptible',
                        startupPolicy: DistributedInferenceStartupPolicy.LeaderReady,
                        leader: {
                            image: 'ubuntu'
                        },
                        worker: {
                            image: 'ubuntu'
                        }
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        {},
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Validate workload setup fields have policy default values
                    const hybridGeneralSection = hybridForm.hybridGeneralSection;

                    // Validate replicas
                    await elementValidationUtils.expectInputValueToEqual(
                        '[aid="replicas-input"]',
                        policyDefaults.replicas.toString()
                    );

                    await inferenceFlatUIUtil.inference.openGeneralSection();

                    // Validate priority
                    await elementValidationUtils.expectElementValueToContain(
                        hybridGeneralSection.prioritySelect,
                        policyDefaults.priorityClass
                    );

                    // Validate preemptibility
                    await elementValidationUtils.expectElementValueToContain(
                        hybridGeneralSection.preemptibilitySelect,
                        'Preemptible'
                    );

                    // Validate startup policy
                    await elementValidationUtils.expectElementValueToContain(
                        hybridGeneralSection.startupPolicy,
                        'Once the leader pod is ready'
                    );

                    // Submit the workload
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify workload setup values
                    expect(distributedInferenceWorkload.spec.priorityClass).toBe(
                        policyDefaults.priorityClass
                    );
                    expect(distributedInferenceWorkload.spec.preemptibility).toBe(policyDefaults.preemptibility);
                    expect(distributedInferenceWorkload.spec.startupPolicy).toBe(
                        policyDefaults.startupPolicy
                    );
                    expect(distributedInferenceWorkload.spec.replicas).toBe(policyDefaults.replicas);

                    logger.info('test distributed inference workload setup policy defaults passed!');
                });
            });

            test.describe('Test inference policy rules with preemptibility', function () {
                test('distributed inference workload setup policy rules - all fields locked including preemptibility', async function () {
                    const policyName = objectGenerator.generateNameByPrefix('workload-setup-rules-policy');

                    const policyDefaults = {
                        workers: 2,
                        replicas: 3,
                        servingPort: { port: 3131 },
                        priorityClass: 'low',
                        preemptibility: 'preemptible',
                        startupPolicy: DistributedInferenceStartupPolicy.LeaderReady,
                        leader: {
                            image: 'ubuntu'
                        },
                        worker: {
                            image: 'ubuntu'
                        }
                    };

                    const policyRules = {
                        priorityClass: { canEdit: false },
                        preemptibility: { canEdit: false },
                        startupPolicy: { canEdit: false },
                        replicas: { canEdit: false }
                    };

                    await policyUtils.createPolicyV2(
                        policyName,
                        project.id.toString(),
                        PolicyWorkloadType.DistributedInference,
                        policyDefaults,
                        policyRules,
                        ScopeType.Project
                    );

                    // Open distributed inference form
                    const distributedMeta = {
                        ...inferenceMeta,
                        distributed: true
                    };
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(distributedMeta);
                    await inferenceFlatUIUtil.waitForDistributedLeaderFormVisible();
                    
                    // Go to worker step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkerFormVisible();

                    // Go to workload setup step
                    await inferenceFlatUIUtil.inference.clickOnCreateInferenceBtn();
                    await inferenceFlatUIUtil.waitForDistributedWorkloadSetupFormVisible();

                    // Validate workload setup fields are disabled
                    const hybridFormPage = hybridForm.page;

                    // Validate replicas is disabled
                    await expect(hybridFormPage.locator('[aid="replicas-input"]')).toBeDisabled();

                    await inferenceFlatUIUtil.inference.openGeneralSection();

                    // Validate priority is disabled (aria-disabled is on the parent label)
                    await expect(hybridFormPage.locator('.priority-select')).toHaveAttribute(
                        'aria-disabled',
                        'true'
                    );

                    // Validate preemptibility is disabled
                    await expect(hybridFormPage.locator('label.preemptibility-select')).toHaveAttribute(
                        'aria-disabled',
                        'true'
                    );

                    // Validate startup policy is disabled (aria-disabled is on the parent label)
                    await expect(hybridFormPage.locator('.startup-policy-select')).toHaveAttribute(
                        'aria-disabled',
                        'true'
                    );

                    // Submit the workload
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // Verify the created workload
                    distributedInferenceWorkload = await workloadsCrudUtils.getDistributedInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(distributedInferenceWorkload).toBeDefined();
                    logger.debug(
                        `distributedInferenceWorkload: ${JSON.stringify(distributedInferenceWorkload)}`
                    );

                    // Verify workload setup values match policy defaults (since they were locked)
                    expect(distributedInferenceWorkload.spec.priorityClass).toBe(
                        policyDefaults.priorityClass
                    );
                    expect(distributedInferenceWorkload.spec.preemptibility).toBe(
                        policyDefaults.preemptibility
                    );
                    expect(distributedInferenceWorkload.spec.startupPolicy).toBe(
                        policyDefaults.startupPolicy
                    );
                    expect(distributedInferenceWorkload.spec.replicas).toBe(policyDefaults.replicas);

                    logger.info('test distributed inference workload setup policy rules passed!');
                });
            });
        });
    });
});
