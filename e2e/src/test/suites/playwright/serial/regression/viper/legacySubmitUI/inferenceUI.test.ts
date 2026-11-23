/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';

import { InferencePage } from '../../../../../../../modules/ui/runaiV2/pages/InferencePage';
import { EnvironmentPage } from '../../../../../../../modules/ui/runaiV2/pages/EnvironmentPage';
import { ComputeResourcePage } from '../../../../../../../modules/ui/runaiV2/pages/ComputeResourcePage';
import { AuthPage } from '../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    apiUtils,
    computeResourceUtils,
    config,
    environmentUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    runaiHealthChecker,
    tags,
    User,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { WorkloadPage } from '../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { Inference, Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/api';
import { expect } from 'chai';

import { WorkloadType } from '../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { test } from '@playwright/test';
import {
    IInferenceType,
    IServingPortAccessOptions
} from '../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';

interface IAutoScalingOptions {
    minReplicas?: number;
    maxReplicas?: number;
    thresholdValue?: number;
    autoScalingPeriod?: string;
}

interface IServingPortOptions {
    protocol?: string;
    port?: number;
    access?: {
        option: IServingPortAccessOptions;
    };
}

interface IAssetBasedOptions {
    inferenceName: string;
    projectName: string;
    environment: {
        name: string;
        servingPort?: IServingPortOptions;
        runtimeSettings?: {
            environmentVariablesValues?: Array<string>;
        };
    };
    computeResourceName: string;
}

test.describe(
    `Inferences UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        let project: Project;
        let videoName: string;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let inference: InferencePage;
        let environment: EnvironmentPage;
        let computeResource: ComputeResourcePage;
        let workload: WorkloadPage;
        const currentUser: User = flowRootState.users.system_administrator;

        async function fillAutoScalingOptions(autoScaling: IAutoScalingOptions): Promise<void> {
            await inference.clickOnAutoScaleToggleBtn();

            if (autoScaling.minReplicas !== undefined) {
                await inference.fillMinReplicasInput(autoScaling.minReplicas);
            }

            if (autoScaling.maxReplicas) {
                await inference.fillMaxReplicasInput(autoScaling.maxReplicas);
            }

            if (autoScaling.thresholdValue) {
                await inference.fillThresholdValueInput(autoScaling.thresholdValue);
            }

            if (autoScaling.autoScalingPeriod && autoScaling.minReplicas !== 0) {
                throw new Error('autoScalingPeriod is not supported for minReplicas > 0');
            }
            if (autoScaling.minReplicas === 0 && autoScaling.autoScalingPeriod) {
                if (autoScaling.autoScalingPeriod === '1m') {
                    await inference.selectAfter1Minute();
                }
            }
        }

        test.beforeAll('creating project and namespace', async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            browserEngine = new BrowserEngine();

            const projectName = objectGenerator.generateNameByPrefix('p-inf');
            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 1, limit: 1 } }]
            });
            await apiUtils.disableHybridSubmission();
            logger.info(`created project ${projectName}`);
        });

        test.afterAll('deleting project and namespace', async function () {
            logger.debug('global after all test');
            if (!(await runaiHealthChecker.isSelfHostedCluster())) {
                // This condition should be removed when fixing problem with saas envs with cluster version.
                return;
            }

            await apiUtils.enableHybridSubmission();

            if (!project) {
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
        });

        test.beforeEach('start browser', async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            inference = new InferencePage(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            await auth.signIn(currentUser.email, currentUser.password);
        });

        test.afterEach('stop browser', async ({}, testInfo) => {
            logger.debug('global after each test');
            await browserEngine.stopAndArchive(testInfo, 'inferences-' + videoName);
        });

        test.describe('Create custom inference based assets', function () {
            let inferenceName: string;
            let environmentName: string;
            let environmentImage: string;
            let computeName: string;
            let newEnvironmentCreated: boolean;
            let newComputeCreated: boolean;
            let inferenceWorkload: Inference;

            test.beforeEach('before each create new inference', async function () {
                environment = new EnvironmentPage(browserEngine.currentPage);
                computeResource = new ComputeResourcePage(browserEngine.currentPage);
                inferenceName = objectGenerator.generateNameByPrefix('inf');
                environmentName = objectGenerator.generateNameByPrefix('environment');
                environmentImage = config.quickStartImage.inferenceHelloWorld;
                computeName = objectGenerator.generateNameByPrefix('compute');
                workload = new WorkloadPage(browserEngine.currentPage);

                await workload.navigateToWorkloads();
            });

            test.afterEach('cleanup after each create new inference', async () => {
                logger.debug('cleanup after each create new inference');

                // delete inference
                if (inferenceWorkload) {
                    logger.debug(`waiting for workload ${inferenceName} to be deleted`);
                    await workloadsUtilsV2.deleteWorkload(
                        inferenceWorkload.workloadId,
                        WorkloadType.INFERENCE,
                        currentUser.accessToken
                    );
                    inferenceWorkload = undefined;
                    logger.debug(`inference ${inferenceName} deleted!`);
                }

                // cleanup assets
                if (newEnvironmentCreated) {
                    await environmentUtils.deleteByName(environmentName);
                }
                if (newComputeCreated) {
                    await computeResourceUtils.deleteByName(computeName);
                }
            });

            async function createInferenceAssetBased(options: IAssetBasedOptions): Promise<void> {
                await workload.openInferenceForm();
                // fill the inference form details
                await inference.fillTheFirstStep(
                    flowRootState.cluster.name,
                    options.projectName,
                    options.inferenceName,
                    IInferenceType.Custom
                );

                // part 2

                // select environment
                newEnvironmentCreated = false;
                await inference.selectEnvironment(options.environment.name);
                if (options.environment.servingPort) {
                    const protocolPortElement = inference.page.locator(inference.servingPortProtocol);
                    expect(await inference.isVisible(protocolPortElement)).to.be.true;
                    const servingPortElement = inference.page.locator(inference.servingPortPort);
                    expect(await inference.isVisible(servingPortElement)).to.be.true;

                    if (options.environment.servingPort.access) {
                        await inference.selectServingPortAccess(
                            options.environment.servingPort.access.option
                        );
                    }

                    if (options.environment.runtimeSettings) {
                        await inference.clickOnEnvSectionMoreSettingsBtn();

                        const environmentVariablesValues =
                            options.environment.runtimeSettings.environmentVariablesValues;
                        if (environmentVariablesValues) {
                            await inference.fillEnvironmentVariablesValues(environmentVariablesValues);
                        }
                    }
                }

                // select compute
                newComputeCreated = false;
                await inference.selectComputeResource(options.computeResourceName);

                // create inference
                await inference.clickOnCreateInferenceBtn();

                await waitUtils.waitForElementToExist(workload, `td:has-text("${options.inferenceName}")`);
                logger.debug(`inference ${options.inferenceName} created successfully!`);
            }

            test('should create new inference from scratch', async function () {
                videoName = 'create-inference-from-scratch';
                await workload.openInferenceForm();

                // fill the inference form details
                await inference.fillTheFirstStep(
                    flowRootState.cluster.name,
                    project.name,
                    inferenceName,
                    IInferenceType.Custom
                );

                // create environment
                await inference.clickOnNewEnvironmentBtn();
                newEnvironmentCreated = true;
                await environment.createNewEnvironment(environmentName, project.name, {
                    image: environmentImage,
                    containerPort: 3000
                });

                // create compute
                newComputeCreated = true;
                await inference.clickOnNewComputeBtn();
                await computeResource.createNewCompute(computeName, project.name);

                // create inference
                await inference.clickOnCreateInferenceBtn();

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    inferenceWorkload.workloadId,
                    Phase.Running,
                    120
                );
                logger.debug(`inference ${inferenceName} created successfully!`);
                logger.debug(`inference ${inferenceName} is running!`);
            });

            test('should create new inference with default auto-scale', async function () {
                videoName = 'create-inference-with-default-auto-scale';
                await workload.openInferenceForm();

                // fill the inference form details
                await inference.fillTheFirstStep(
                    flowRootState.cluster.name,
                    project.name,
                    inferenceName,
                    IInferenceType.Custom
                );

                // create environment
                newEnvironmentCreated = true;
                await inference.clickOnNewEnvironmentBtn();
                await environment.createNewEnvironment(environmentName, project.name, {
                    image: environmentImage,
                    containerPort: 3000
                });

                // create compute
                newComputeCreated = true;
                await inference.clickOnNewComputeBtn();
                await computeResource.createNewCompute(computeName, project.name);

                // auto scale
                await fillAutoScalingOptions({
                    maxReplicas: 4,
                    thresholdValue: 97
                });

                // create inference
                await inference.clickOnCreateInferenceBtn();
                logger.debug('inference with auto scale created successfully');

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    inferenceWorkload.workloadId,
                    Phase.Running,
                    120
                );
                logger.debug(`inference ${inferenceName} is running!`);
            });

            test('should create new inference with auto-scale and custom period', async function () {
                videoName = 'create-inference-with-auto-scale-and-custom-period';
                await workload.openInferenceForm();

                // fill the inference form details
                await inference.fillTheFirstStep(
                    flowRootState.cluster.name,
                    project.name,
                    inferenceName,
                    IInferenceType.Custom
                );

                // create environment
                newEnvironmentCreated = true;
                await inference.clickOnNewEnvironmentBtn();
                await environment.createNewEnvironment(environmentName, project.name, {
                    image: environmentImage,
                    containerPort: 3000
                });

                // create compute
                newComputeCreated = true;
                await inference.clickOnNewComputeBtn();
                await computeResource.createNewCompute(computeName, project.name);

                // auto scale
                await fillAutoScalingOptions({
                    minReplicas: 0,
                    maxReplicas: 1,
                    autoScalingPeriod: '1m'
                });

                // create inference
                await inference.clickOnCreateInferenceBtn();
                logger.debug('inference with auto scale created successfully');

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    inferenceWorkload.workloadId,
                    Phase.Running,
                    120
                );
                logger.debug(`inference ${inferenceName} is running!`);
            });
        });
    }
);
