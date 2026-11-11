/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    flowRootState,
    logger,
    nodePoolUtils,
    objectGenerator,
    orgUnitUtils,
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
    Inference,
    ServingPortProtocol
} from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { InferenceFlatUIUtil } from '../../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import {
    InferencePolicyChangeRequestV2,
    PvcVolumeMode,
    ScopeType,
    ServingPortAccessAuthorizationTypeEnum
} from '../../../../../../../modules/api/apis/policyApi';
import { HybridForm } from '../../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { IInferenceMeta } from '../../../../../../../modules/helpers/models/inferenceFlatUI.models';
import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { IInferenceType } from '../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { ElementValidationUtils } from '../../../../../../../modules/ui/utils/ElementValidationUtils';
import { backendClusterApi } from '../../../../../../../modules/api';

test.describe(`Inferences NIM Policy  UI, tags: ${tags.tool.ui} ${tags.resource.policy} ${tags.teams.viper}`, function () {
    test.setTimeout(60000);

    const suiteName = 'inference-nim-policy-ui';
    let nodePoolName: string;
    let nodePoolId: number;
    let project: Project;
    let traceName: string;

    let browserEngine: BrowserEngine;
    let auth: AuthPage;
    let workload: WorkloadPage;
    let inferenceFlatUIUtil: InferenceFlatUIUtil;
    let elementValidationUtils: ElementValidationUtils;
    let hybridForm: HybridForm;
    let inferenceMeta: IInferenceMeta;
    let inferenceName: string;
    let inferenceWorkload: Inference;
    const nimModelName: string = 'meta/llama-3.1-405b-instruct';
    const apiKey: string = 'someProfile';

    function createPolicyName(): string {
        const policyName = objectGenerator.generateNameByPrefix('inference-policy');
        logger.debug(`policy name: ${policyName} created for inference policy test`);
        return policyName;
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

    test.beforeAll('before all inference policy test based on NIM UI', async function () {
        if (!(await isKnativeInstalled())) {
            logger.warn('knative is not installed on the cluster, skipping test');
            test.skip();
        }

        browserEngine = new BrowserEngine();

        backendClusterApi.nodePoolsApi.accessToken = flowRootState.users.system_administrator.accessToken;
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

    test.afterAll('after all inference policy test based on NIM UI', async ({}, testInfo) => {
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

        if (!(await isKnativeInstalled())) {
            logger.warn('knative is not installed on the cluster, skipping test');
            return;
        }

        try {
            await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id);
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

    test.beforeEach('before each inference policy test based on NIM UI', async function () {
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

    test.afterEach('after each inference policy test based on NIM UI', async ({}, testInfo) => {
        logger.debug(`waiting for workload ${inferenceName} to be deleted`);

        // cleanup policy
        await workloadUtils.deletePolicy(project.id);

        // delete inference
        if (inferenceWorkload) {
            await workloadsUtilsV2.deleteWorkload(
                inferenceWorkload.workloadId,
                WorkloadType.INFERENCE,
                flowRootState.users.system_administrator.accessToken
            );
        }
        inferenceWorkload = undefined;

        logger.debug(`inference ${inferenceName} deleted!`);

        // Stop and archive tracing for this test (browser, context, and page remain alive)
        await browserEngine.stopAndArchiveTracing(testInfo, traceName);
    });

    supportedFrom(SupportedVersions.V2_20)(function () {
        test.describe('Test NIM inference policy defaults', function () {
            test.describe('Test model section', function () {
                test('nim inference policy serving port defaults', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                servingPort: {
                                    container: 3131,
                                    protocol: ServingPortProtocol.Grpc
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.servingPortPort,
                        policyDefaults.servingPort.container.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.servingPortProtocol,
                        policyDefaults.servingPort.protocol.toString()
                    );

                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    expect(inferenceWorkload.spec.servingPort.container).toBe(
                        policyDefaults.servingPort.container
                    );
                    expect(inferenceWorkload.spec.servingPort.protocol).toBe(
                        policyDefaults.servingPort.protocol
                    );

                    logger.info('test inference serving port policy passed!');
                });

                test('nim inference policy serving port authorization with authenticatedUsers as the authorizationType default', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                servingPort: {
                                    container: 3131,
                                    authorizationType:
                                        ServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers
                                }
                            }
                        }
                    };

                    const authenticatedUsersLabel = 'All authenticated users and service accounts';

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.servingPortPort,
                        policyDefaults.servingPort.container.toString()
                    );

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.servingPortAccess.servingPortAccessQuickView,
                        authenticatedUsersLabel
                    );

                    await hybridForm
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessEditBtn)
                        .click();

                    await elementValidationUtils.expectElementValueToContain(
                        hybridEnvironmentSection.servingPortAccess.servingPortAccessModalOptions,
                        authenticatedUsersLabel
                    );

                    await hybridForm
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessModalSaveBtn)
                        .click();

                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    expect(inferenceWorkload.spec.servingPort.container).toBe(
                        policyDefaults.servingPort.container
                    );
                    expect(inferenceWorkload.spec.servingPort.protocol).toBe(
                        'http' // default value
                    );

                    expect(inferenceWorkload.spec.servingPort.authorizationType).toBe(
                        policyDefaults.servingPort.authorizationType
                    );

                    logger.info('test inference serving port policy passed!');
                });

                test('nim inference policy serving port authorization with authenticatedUsersOrGroups as the authorizationType default and default groups', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                servingPort: {
                                    container: 3131,
                                    authorizationType:
                                        ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                                    authorizedGroups: ['group1', 'group2']
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.servingPortAccess.servingPortAccessQuickView,
                        'Groups (2)'
                    );

                    await hybridForm
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessEditBtn)
                        .click();

                    await elementValidationUtils.expectElementValueToContain(
                        hybridEnvironmentSection.servingPortAccess.servingPortAccessModalOptions,
                        'Specific group(s)'
                    );

                    await hybridForm
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessModalSaveBtn)
                        .click();

                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    expect(inferenceWorkload.spec.servingPort.container).toBe(
                        policyDefaults.servingPort.container
                    );
                    expect(inferenceWorkload.spec.servingPort.protocol).toBe(
                        'http' // default value
                    );

                    expect(inferenceWorkload.spec.servingPort.authorizationType).toBe(
                        policyDefaults.servingPort.authorizationType
                    );

                    expect(inferenceWorkload.spec.servingPort.authorizedGroups).toEqual(
                        policyDefaults.servingPort.authorizedGroups
                    );

                    logger.info('test inference serving port policy passed!');
                });

                test('nim inference policy serving port authorization with authenticatedUsersOrGroups as the authorizationType default and default users', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                servingPort: {
                                    container: 3131,
                                    authorizationType:
                                        ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                                    authorizedUsers: ['user1@mail.com', 'user2@mail.com']
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridEnvironmentSection.servingPortAccess.servingPortAccessQuickView,
                        'Users (2)'
                    );

                    await hybridForm
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessEditBtn)
                        .click();

                    await elementValidationUtils.expectElementValueToContain(
                        hybridEnvironmentSection.servingPortAccess.servingPortAccessModalOptions,
                        'Specific users and service accounts'
                    );

                    await hybridForm
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessModalSaveBtn)
                        .click();

                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    expect(inferenceWorkload.spec.servingPort.container).toBe(
                        policyDefaults.servingPort.container
                    );
                    expect(inferenceWorkload.spec.servingPort.protocol).toBe(
                        'http' // default value
                    );

                    expect(inferenceWorkload.spec.servingPort.authorizationType).toBe(
                        policyDefaults.servingPort.authorizationType
                    );

                    expect(inferenceWorkload.spec.servingPort.authorizedUsers).toEqual(
                        policyDefaults.servingPort.authorizedUsers
                    );

                    logger.info('test inference serving port policy passed!');
                });
            });

            test.describe('Test compute section', function () {
                test('nim non specific NIM inference compute policy defaults', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                // nodeType: 'my-node-type',// TODO: uncomment this when the issue is fixed
                                nodePools: [nodePoolName],
                                compute: {
                                    gpuDevicesRequest: 2,
                                    cpuCoreRequest: 1,
                                    cpuCoreLimit: 2
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridComputeSection = hybridForm.hybridComputeSection;
                    const policyDefaults = policyRequest.policy.defaults;
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.gpuDevicesInput,
                        policyDefaults.compute.gpuDevicesRequest.toString()
                    );
                    const cpuRequestValue = policyRequest.policy.defaults.compute.cpuCoreRequest * 1000; // To MB
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.cpuCoresRequestInput,
                        cpuRequestValue.toString()
                    );
                    const cpuLimitValue = policyRequest.policy.defaults.compute.cpuCoreLimit * 1000; // To MB
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.cpuCoresLimitInput,
                        cpuLimitValue.toString()
                    );
                    // TODO: uncomment this when the issue is fixed
                    // await expectInputValueToEqual(
                    //     hybridFormPage,
                    //     hybridComputeSection.nodeAffinityInput,
                    //     policyDefaults.nodeType
                    // );
                    await elementValidationUtils.expectElementValueToContain(
                        `.node-pools-resource-section [aid="node-pool-${policyRequest.policy.defaults.nodePools[0]}-list-item"]`,
                        policyRequest.policy.defaults.nodePools[0]
                    );

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    // expect(inferenceWorkload.spec.nodeType).toBe(policyDefaults.nodeType); // TODO: uncomment this when the issue is fixed
                    expect(inferenceWorkload.spec.nodePools).toEqual(policyDefaults.nodePools);
                    expect(inferenceWorkload.spec.compute.gpuDevicesRequest).toBe(
                        policyDefaults.compute.gpuDevicesRequest
                    );
                    expect(inferenceWorkload.spec.compute.cpuCoreRequest).toBe(
                        policyDefaults.compute.cpuCoreRequest
                    );
                    expect(inferenceWorkload.spec.compute.cpuCoreLimit).toBe(
                        policyDefaults.compute.cpuCoreLimit
                    );

                    logger.info('test none specific inference policy values passed!');
                });

                test('nim inference autoscaling policy defaults', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                autoscaling: {
                                    minReplicas: 2,
                                    maxReplicas: 5,
                                    metric: 'concurrency',
                                    metricThreshold: 50
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridComputeSection = hybridForm.hybridComputeSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.minReplicasInput,
                        policyDefaults.autoscaling.minReplicas.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.maxReplicasInput,
                        policyDefaults.autoscaling.maxReplicas.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.metricSelect,
                        policyDefaults.autoscaling.metric
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.metricThresholdInput,
                        policyDefaults.autoscaling.metricThreshold.toString()
                    );
                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).toBe(
                        policyDefaults.autoscaling.minReplicas
                    );
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).toBe(
                        policyDefaults.autoscaling.maxReplicas
                    );
                    expect(inferenceWorkload.spec.autoscaling.metric).toBe(policyDefaults.autoscaling.metric);
                    expect(inferenceWorkload.spec.autoscaling.metricThreshold).toBe(
                        policyDefaults.autoscaling.metricThreshold
                    );

                    logger.info('test inference autoscaling policy values passed!');
                });

                test('nim inference autoscaling scale down policy defaults', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                autoscaling: {
                                    minReplicas: 0
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridComputeSection = hybridForm.hybridComputeSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.minReplicasInput,
                        policyDefaults.autoscaling.minReplicas.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.scaleDownSelect,
                        'After 1 minute of inactivity'
                    );
                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    expect(inferenceWorkload.spec.autoscaling.minReplicas).toBe(
                        policyDefaults.autoscaling.minReplicas
                    );
                    expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).toBe(60); // default is 1 minute
                    logger.info(
                        'test inference autoscaling scale down default duration policy values passed!'
                    );
                });

                test('nim inference autoscaling with default scale down retention value policy', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                autoscaling: {
                                    minReplicas: 0,
                                    scaleToZeroRetentionSeconds: 600
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridComputeSection = hybridForm.hybridComputeSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.minReplicasInput,
                        policyDefaults.autoscaling.minReplicas.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.scaleDownSelect,
                        'Custom'
                    );

                    const inMinutes = policyDefaults.autoscaling.scaleToZeroRetentionSeconds / 60;
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.scaleToZeroRetentionSecondsInput,
                        inMinutes.toString()
                    );

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // check the created workload
                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    expect(inferenceWorkload.spec.autoscaling.minReplicas).toBe(
                        policyDefaults.autoscaling.minReplicas
                    );
                    expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).toBe(
                        policyDefaults.autoscaling.scaleToZeroRetentionSeconds
                    );

                    logger.info(
                        'test inference autoscaling scale down custom duration policy values passed!'
                    );
                });

                test('nim inference autoscaling with metric policy when minReplicas and maxReplicas updated', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                autoscaling: {
                                    metric: 'concurrency',
                                    metricThreshold: 50
                                }
                            }
                        }
                    };

                    const maxReplicas = 2;

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridFormPage = hybridForm.page;
                    const hybridComputeSection = hybridForm.hybridComputeSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.minReplicasInput,
                        '1'
                    );

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.maxReplicasInput,
                        '1'
                    );

                    const metricSelectCount = await hybridFormPage
                        .locator(hybridComputeSection.autoScaling.metricSelect)
                        .count();
                    expect(metricSelectCount).toBe(0);

                    const metricThresholdInputCount = await hybridFormPage
                        .locator(hybridComputeSection.autoScaling.metricThresholdInput)
                        .count();
                    expect(metricThresholdInputCount).toBe(0);

                    await inferenceFlatUIUtil.fillAutoScalingOptions({ maxReplicas });

                    // we expected to see the metric threshold input and metric select with the policy default values
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.metricThresholdInput,
                        policyDefaults.autoscaling.metricThreshold.toString()
                    );

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridComputeSection.autoScaling.metricSelect,
                        policyDefaults.autoscaling.metric
                    );

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // check the created workload
                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    expect(inferenceWorkload.spec.autoscaling.minReplicas).toBe(1);

                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).toBe(maxReplicas);

                    expect(inferenceWorkload.spec.autoscaling.metric).toBe(policyDefaults.autoscaling.metric);

                    expect(inferenceWorkload.spec.autoscaling.metricThreshold).toBe(
                        policyDefaults.autoscaling.metricThreshold
                    );
                    logger.info(
                        'test inference autoscaling scale down custom duration policy values passed!'
                    );
                });
            });

            test.describe('Test storage section', function () {
                test('nim inference policy defaults in storage section instance pvc', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                storage: {
                                    pvc: {
                                        instances: [
                                            {
                                                claimName: 'claim-10rkjlqfgw',
                                                path: '/home/log',
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
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );
                    const hybridDataSourceSection = hybridForm.hybridDataSourceSection;
                    const policyDefaults = policyRequest.policy.defaults.storage.pvc.instances[0];

                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.storageClass,
                        policyDefaults.claimInfo.storageClass
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
                        policyDefaults.claimInfo.volumeMode.toString()
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        hybridDataSourceSection.pvc.path,
                        policyDefaults.path
                    );

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // check the created workload
                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    const pvc = inferenceWorkload.spec.storage.pvc[0];
                    expect(pvc.path).toBe(policyDefaults.path);
                    expect(pvc.claimInfo.storageClass).toBe(policyDefaults.claimInfo.storageClass);
                    expect(pvc.claimInfo.size).toBe(policyDefaults.claimInfo.size);
                    expect(pvc.claimInfo.volumeMode).toBe(policyDefaults.claimInfo.volumeMode);
                    expect(pvc.claimInfo.accessModes).toEqual({
                        readWriteOnce: true,
                        readOnlyMany: false,
                        readWriteMany: false
                    });
                    logger.info('test inference policy data source pvc section passed!');
                });
            });
        });

        test.describe('Test NIM inference policy rules', function () {
            test.describe('Test model section', function () {
                test('nim inference policy serving port rules', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                servingPort: {
                                    container: 3131,
                                    protocol: ServingPortProtocol.Grpc
                                }
                            },
                            rules: {
                                servingPort: {
                                    container: {
                                        canEdit: false
                                    },
                                    protocol: {
                                        canEdit: false // can add options
                                    }
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridFormPage = hybridForm.page;
                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });

                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.servingPortProtocol)
                    ).toBeDisabled();

                    await expect(
                        hybridFormPage.locator(hybridEnvironmentSection.servingPortPort)
                    ).toBeDisabled();

                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    expect(inferenceWorkload.spec.servingPort.container).toBe(
                        policyDefaults.servingPort.container
                    );
                    expect(inferenceWorkload.spec.servingPort.protocol).toBe(
                        policyDefaults.servingPort.protocol
                    );

                    expect(inferenceWorkload.spec.environmentVariables.find(env => env.name === 'NIM_SERVER_PORT')?.value).toBe(
                        policyDefaults.servingPort.container.toString()
                    );

                    logger.info('test inference serving port policy passed!');
                });

                test('nim inference policy serving port authorization with when authorizationType can not be edited', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                servingPort: {
                                    container: 3131,
                                    authorizationType:
                                        ServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers
                                }
                            },
                            rules: {
                                servingPort: {
                                    authorizationType: {
                                        canEdit: false
                                    }
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridFormPage = hybridForm.page;
                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });

                    await hybridForm
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessEditBtn)
                        .click();

                    const authorizedOptions = hybridFormPage
                        .locator(
                            hybridForm.hybridEnvironmentSection.servingPortAccess
                                .servingPortAccessModalOptions
                        )
                        .getByRole('radio');
                    const count = await authorizedOptions.count();
                    expect(count).toBe(4);

                    const allOptions = await authorizedOptions.all();
                    for (const op of allOptions) {
                        await expect(op).toHaveAttribute('aria-disabled', 'true');
                    }

                    await hybridFormPage
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessModalSaveBtn)
                        .click();

                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    expect(inferenceWorkload.spec.servingPort.container).toBe(
                        policyDefaults.servingPort.container
                    );
                    expect(inferenceWorkload.spec.servingPort.protocol).toBe(
                        'http' // default value
                    );

                    logger.info('test inference serving port policy passed!');
                });

                test('nim inference policy serving port authorization when authorizedGroups can not be edited', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                servingPort: {
                                    container: 3131,
                                    authorizationType:
                                        ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                                    authorizedGroups: ['group1', 'group2']
                                }
                            },
                            rules: {
                                servingPort: {
                                    authorizedGroups: {
                                        canEdit: false
                                    }
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridFormPage = hybridForm.page;
                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });

                    await hybridForm
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessEditBtn)
                        .click();

                    const addItemBtn = hybridFormPage.locator(
                        hybridEnvironmentSection.servingPortAccess.servingPortAccessModalAddItemBtn
                    );
                    expect(await addItemBtn.getAttribute('disabled')).toBeDefined();

                    const accessInputContainer = hybridFormPage.locator('.access-input-container');
                    expect(await accessInputContainer.count()).toBe(2);
                    for (const row of await accessInputContainer.all()) {
                        const input = row.locator('input');
                        expect(await input.getAttribute('disabled')).toBeDefined();
                        const removeBtn = row.locator(
                            hybridEnvironmentSection.servingPortAccess.servingPortAccessModalRemoveItemBtn
                        );
                        expect(await removeBtn.getAttribute('disabled')).toBeDefined();
                    }

                    await hybridFormPage
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessModalSaveBtn)
                        .click();

                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    expect(inferenceWorkload.spec.servingPort.authorizationType).toBe(
                        policyDefaults.servingPort.authorizationType
                    );

                    expect(inferenceWorkload.spec.servingPort.authorizedGroups).toEqual(
                        policyDefaults.servingPort.authorizedGroups
                    );

                    logger.info('test inference serving port policy passed!');
                });

                test('nim inference policy serving port authorization when authorizedUsers can not be edited', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                servingPort: {
                                    container: 3131,
                                    authorizationType:
                                        ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                                    authorizedUsers: ['user1@mail.com', 'user2@mail.com']
                                }
                            },
                            rules: {
                                servingPort: {
                                    authorizedUsers: {
                                        canEdit: false
                                    }
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridFormPage = hybridForm.page;
                    const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });

                    await hybridForm
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessEditBtn)
                        .click();

                    await hybridForm.expectServingPortAuthorizedUsersChips(
                        policyDefaults.servingPort.authorizedUsers
                    );

                    await hybridFormPage
                        .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessModalSaveBtn)
                        .click();

                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    expect(inferenceWorkload.spec.servingPort.authorizationType).toBe(
                        policyDefaults.servingPort.authorizationType
                    );

                    expect(inferenceWorkload.spec.servingPort.authorizedGroups).toEqual(
                        policyDefaults.servingPort.authorizedGroups
                    );

                    logger.info('test inference serving port policy passed!');
                });
            });

            test.describe('Test compute section', function () {
                test('nim non specific inference compute policy rules', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                // nodeType: 'my-node-type',// TODO: uncomment this when the issue is fixed
                                nodePools: [nodePoolName],
                                compute: {
                                    gpuDevicesRequest: 2,
                                    cpuCoreRequest: 1,
                                    cpuCoreLimit: 2
                                }
                            },
                            rules: {
                                // nodeType: { // TODO: uncomment this when the issue is fixed and add test
                                //     canEdit: false
                                // },
                                nodePools: {
                                    canEdit: false
                                },
                                compute: {
                                    gpuDevicesRequest: {
                                        canEdit: false
                                    },
                                    cpuCoreRequest: {
                                        canEdit: false
                                    },
                                    cpuCoreLimit: {
                                        canEdit: false
                                    }
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridFormPage = hybridForm.page;
                    const hybridComputeSection = hybridForm.hybridComputeSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await expect(hybridFormPage.locator(hybridComputeSection.gpuDevicesInput)).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.cpuCoresRequestInput)
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.cpuCoresLimitInput)
                    ).toBeDisabled();

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    // check the created workload
                    // expect(inferenceWorkload.spec.nodeType).toBe(policyDefaults.nodeType); // TODO: uncomment this when the issue is fixed
                    expect(inferenceWorkload.spec.nodePools).toEqual(policyDefaults.nodePools);
                    expect(inferenceWorkload.spec.compute.gpuDevicesRequest).toBe(
                        policyDefaults.compute.gpuDevicesRequest
                    );
                    expect(inferenceWorkload.spec.compute.cpuCoreRequest).toBe(
                        policyDefaults.compute.cpuCoreRequest
                    );
                    expect(inferenceWorkload.spec.compute.cpuCoreLimit).toBe(
                        policyDefaults.compute.cpuCoreLimit
                    );

                    logger.info('test none specific inference policy values passed!');
                });

                test('nim inference autoscaling with default scale down retention value policy rules', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                autoscaling: {
                                    minReplicas: 0,
                                    maxReplicas: 2,
                                    metric: 'throughput',
                                    metricThreshold: 10,
                                    scaleToZeroRetentionSeconds: 600
                                }
                            },
                            rules: {
                                autoscaling: {
                                    minReplicas: {
                                        canEdit: false
                                    },
                                    maxReplicas: {
                                        canEdit: false
                                    },
                                    scaleToZeroRetentionSeconds: {
                                        canEdit: false
                                    },
                                    metric: {
                                        canEdit: false
                                    },
                                    metricThreshold: {
                                        canEdit: false
                                    }
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );

                    const hybridFormPage = hybridForm.page;
                    const hybridComputeSection = hybridForm.hybridComputeSection;
                    const policyDefaults = policyRequest.policy.defaults;

                    await expect(
                        hybridFormPage.locator(hybridComputeSection.autoScaling.minReplicasInput)
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.autoScaling.maxReplicasInput)
                    ).toBeDisabled();
                    await expect(hybridFormPage.locator('.scale-down-select')).toHaveAttribute(
                        'aria-disabled'
                    ); // the disabled attribute is on the label and not where the aid is located
                    await expect(
                        hybridFormPage.locator(hybridComputeSection.autoScaling.metricThresholdInput)
                    ).toBeDisabled();
                    await expect(
                        hybridFormPage.locator(
                            hybridComputeSection.autoScaling.scaleToZeroRetentionSecondsInput
                        )
                    ).toBeDisabled();

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // check the created workload
                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    expect(inferenceWorkload.spec.autoscaling.minReplicas).toBe(
                        policyDefaults.autoscaling.minReplicas
                    );
                    expect(inferenceWorkload.spec.autoscaling.maxReplicas).toBe(
                        policyDefaults.autoscaling.maxReplicas
                    );
                    expect(inferenceWorkload.spec.autoscaling.metric).toBe(policyDefaults.autoscaling.metric);
                    expect(inferenceWorkload.spec.autoscaling.metricThreshold).toBe(
                        policyDefaults.autoscaling.metricThreshold
                    );
                    expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).toBe(
                        policyDefaults.autoscaling.scaleToZeroRetentionSeconds
                    );

                    logger.info(
                        'test inference autoscaling scale down custom duration policy values passed!'
                    );
                });
            });

            test.describe('Test storage section', function () {

                test('nim inference policy rules in storage section instance pvc', async function () {
                    const policyName = createPolicyName();
                    const policyRequest: InferencePolicyChangeRequestV2 = {
                        meta: {
                            projectId: parseInt(project.id),
                            scope: ScopeType.Project,
                            name: policyName
                        },
                        policy: {
                            defaults: {
                                storage: {
                                    pvc: {
                                        attributes: {
                                            path: '/home/log',
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
                                        },
                                        instances: [
                                            {
                                                claimName: 'claim-10rkjlqfgw',
                                                path: '/home/log',
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
                            },
                            rules: {
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
                                                // volumeMode: { // this field is missing in the policy
                                                //     canEdit: false
                                                // }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    };

                    await workloadUtils.createInferencePolicy(policyRequest);
                    await inferenceFlatUIUtil.openAndFillInferenceFirstStep(
                        inferenceMeta,
                        IInferenceType.Nim
                    );
                    const hybridFormPage = hybridForm.page;
                    const hybridDataSourceSection = hybridForm.hybridDataSourceSection;
                    const policyDefaults = policyRequest.policy.defaults.storage.pvc.instances[0];

                    await expect(hybridFormPage.locator('.storage-class-select')).toHaveAttribute(
                        'aria-disabled'
                    ); // the disabled attribute is on the label and not where the aid is located
                    // TODO: uncomment when the disabled issue will be fixed
                    // await expect(hybridFormPage.locator(hybridDataSourceSection.pvc.accessMode)).toBeDisabled();
                    // await expect(hybridFormPage.locator(hybridDataSourceSection.pvc.claimSize)).toBeDisabled();
                    // await expect(hybridFormPage.locator(hybridDataSourceSection.pvc.unit)).toBeDisabled();
                    // await expect(hybridFormPage.locator('.unit-select')).toHaveAttribute(
                    //     'aria-disabled'
                    // ); // the disabled attribute is on the label and not where the aid is located
                    // await expect(hybridFormPage.locator('.volume-mode-select')).toHaveAttribute(
                    //     'aria-disabled'
                    // ); // the disabled attribute is on the label and not where the aid is located
                    await expect(hybridFormPage.locator(hybridDataSourceSection.pvc.path)).toBeDisabled();

                    await inferenceFlatUIUtil.fillNimModelData({
                        modelName: nimModelName,
                        apiKey
                    });
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    // check the created workload
                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id
                    );
                    expect(inferenceWorkload).toBeDefined();
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    const pvc = inferenceWorkload.spec.storage.pvc[0];
                    expect(pvc.path).toBe(policyDefaults.path);
                    expect(pvc.claimInfo.storageClass).toBe(policyDefaults.claimInfo.storageClass);
                    expect(pvc.claimInfo.size).toBe(policyDefaults.claimInfo.size);
                    expect(pvc.claimInfo.volumeMode).toBe(policyDefaults.claimInfo.volumeMode);
                    expect(pvc.claimInfo.accessModes).toEqual({
                        readWriteOnce: true,
                        readOnlyMany: false,
                        readWriteMany: false
                    });
                    logger.info('test inference policy data source pvc rules section passed!');
                });
            });
        });
    });
});
