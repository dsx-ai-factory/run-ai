/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    credentialUtils,
    dataSourcesUtils,
    flowRootState,
    logger,
    objectGenerator,
    orgUnitUtils,
    runaiHealthChecker,
    tags,
    templateUtils,
    User,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../../modules/helpers';
import { Project } from '../../../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { WorkloadPage } from '../../../../../../../../modules/ui/runaiV2/pages/WorkloadPage';
import { expect } from 'chai';
import { WorkloadType } from '../../../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { test } from '@playwright/test';
import {
    Inference,
    InferenceSpecSpec,
    InferenceTemplate
} from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { InferenceFlatUIUtil } from '../../../../../../../../modules/ui/utils/inference/inferenceFlatUI.util';
import {
    GenericSecret,
    GenericSecretCreationRequest,
    KeyValuePair,
    PVCAsset,
    PVCAssetSpec,
    PVCCreationRequest,
    PvcVolumeMode,
    Scope,
    ServingPortAccessAuthorizationTypeEnum,
    ServingPortProtocol,
    UidGidSource
} from '../../../../../../../../modules/api/apis/assetsApi';
import {
    ScopeType,
    WorkloadType as WorkloadTypeApi
} from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';
import { HuggingFaceElementValidationUtils } from '../../../../../../../../modules/ui/utils/inference/HuggingFaceElementValidationUtils';

test.describe(
    `Inference Hugging Face based on template UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        const suiteName = 'inference-hf-from-template';
        let project: Project;
        let traceName: string;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workload: WorkloadPage;
        let inferenceFlatUIUtil: InferenceFlatUIUtil;
        let inferenceName: string;
        let inferenceWorkload: Inference;
        let template: InferenceTemplate;
        let hfElementValidationUtils: HuggingFaceElementValidationUtils;

        const huggingFaceModelNameWithoutToken = 'microsoft/DialoGPT-small';
        const huggingFaceModelNameWithToken = 'meta-llama/Llama-2-13b-hf';
        const vllmImage = 'vllm/vllm-openai:latest';
        const tgiImage = 'ghcr.io/huggingface/text-generation-inference:latest';
        const hfToken = process.env.HUGGING_FACE_TOKEN;
        const huggingFaceDefaultPort = 8000;
        const huggingFaceDefaultProtocol = ServingPortProtocol.Http;
        const hfHomeDefaultValue = '/tmp/huggingface';

        const currentUser: User = flowRootState.users.system_administrator;

        async function createTemplate(
            metaConfig: {
                workloadType: WorkloadTypeApi;
                scopeType: ScopeType;
                scopeId: string;
                templateName: string;
                description?: string;
            },
            templateSpec: object
        ): Promise<InferenceTemplate> {
            try {
                const response = await templateUtils.createWorkloadTemplateV2(metaConfig, templateSpec);
                return response.body as InferenceTemplate;
            } catch (error) {
                logger.error(
                    `failed to create template ${metaConfig.templateName}. Error: ${JSON.stringify(error)}`
                );
                throw error;
            }
        }

        async function createGenericSecret(
            secretName: string,
            scope: Scope,
            projectId: string | undefined,
            clusterId: string | undefined,
            keyValuePairs: Array<KeyValuePair>
        ): Promise<GenericSecret> {
            const genericSecret: GenericSecretCreationRequest = {
                meta: {
                    name: secretName,
                    description: null,
                    scope: scope,
                    projectId: parseInt(projectId),
                    clusterId: clusterId,
                    departmentId: null
                },
                spec: { existingSecretName: null, keyValuePairs }
            };

            const genericSecretResponse = await credentialUtils.createGenericSecretAsset(genericSecret);
            await credentialUtils.waitForGenericSecretAssetToBeReady(genericSecretResponse);
            return genericSecretResponse;
        }

        async function createPvcAsset(
            name: string,
            projectId: string,
            spec: PVCAssetSpec
        ): Promise<PVCAsset> {
            const pvc: PVCCreationRequest = {
                meta: {
                    name: name,
                    scope: Scope.Project,
                    projectId: parseInt(projectId)
                },
                spec
            };

            const pvcAsset = await dataSourcesUtils.createPvcAsset(pvc);
            logger.info(`created pvc asset ${pvcAsset.meta.name}`);

            await dataSourcesUtils.waitForPvcAssetToBeReady(
                pvcAsset.meta.name,
                pvcAsset.meta.id,
                parseInt(project.id),
                'Inference'
            );
            logger.info(`created pvc asset ${pvcAsset.meta.name} is ready`);
            return pvcAsset;
        }

        async function initBrowser(): Promise<BrowserEngine> {
            // Initialize browser once for all tests with video recording AND tracing support
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workload = new WorkloadPage(browserEngine.currentPage);
            inferenceFlatUIUtil = new InferenceFlatUIUtil(browserEngine.currentPage);
            hfElementValidationUtils = new HuggingFaceElementValidationUtils(browserEngine.currentPage);
            await auth.signIn(currentUser.email, currentUser.password);

            return browserEngine;
        }

        test.beforeAll('before all create inference based on hugging face template UI', async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            browserEngine = new BrowserEngine();

            const projectName = objectGenerator.generateNameByPrefix('project-e2e-for-hf');
            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 1, limit: 1 } }]
            });

            logger.info(`created project ${projectName}`);

            browserEngine = await initBrowser();

            logger.info('Browser initialized and user signed in');
        });

        test.afterAll(
            'after all create inference based on hugging face template UI',
            async ({}, testInfo) => {
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
            }
        );

        test.beforeEach('before each create inference based on hugging face template UI', async function () {
            if (!browserEngine || !browserEngine.browser) {
                browserEngine = await initBrowser();
            }

            inferenceName = objectGenerator.generateNameByPrefix('inference');

            // Generate trace name from test title
            traceName = objectGenerator.generateNameByPrefix(suiteName);

            // Start tracing for this test (context and page stay alive)
            await browserEngine.startTracing(traceName);

            // Navigate to workloads page for each test
            await workload.navigateToWorkloads();
        });

        test.afterEach(
            'after each create inference based on hugging face template UI',
            async ({}, testInfo) => {
                logger.debug(`waiting for workload ${inferenceName} to be deleted`);

                // delete inference
                if (inferenceWorkload) {
                    await workloadsUtilsV2.deleteWorkload(
                        inferenceWorkload.workloadId,
                        WorkloadType.INFERENCE,
                        flowRootState.users.system_administrator.accessToken
                    );
                }
                inferenceWorkload = undefined;

                if (template) {
                    await templateUtils.deleteWorkloadTemplateV2(template.meta.id, WorkloadTypeApi.Inference);
                }
                template = undefined;

                logger.debug(`inference ${inferenceName} deleted!`);

                // Stop and archive tracing for this test (browser, context, and page remain alive)
                await browserEngine.stopAndArchiveTracing(testInfo, traceName);
            }
        );

        supportedFrom(SupportedVersions.V2_23)(function () {
            test('should create new hugging face inference using an existing template with model without token', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('nim-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image: vllmImage,
                    args: `--model ${huggingFaceModelNameWithoutToken}`,
                    environmentVariables: [
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage
                    },
                    servingPort: {
                        protocol: huggingFaceDefaultProtocol,
                        container: huggingFaceDefaultPort
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                // check that the environment variables hide the api key
                await hfElementValidationUtils.expectEnvironmentVariables([
                    {
                        name: 'HF_HOME',
                        value: hfHomeDefaultValue
                    }
                ]);

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                expect(inferenceWorkload.spec.image).to.be.equals(vllmImage);
                expect(inferenceWorkload.spec.args).to.be.equals(
                    `--model ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceWorkload.spec.servingPort.container).to.be.equals(huggingFaceDefaultPort);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equals(huggingFaceDefaultProtocol);
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equals(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equals(1);
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    {
                        name: 'HF_HOME',
                        value: hfHomeDefaultValue,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);

                logger.info('test inference template passed!');
            });

            test('should create new hugging face inference using an existing template with tgi image', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('nim-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image: tgiImage,
                    args: `--model-id ${huggingFaceModelNameWithoutToken}`,
                    environmentVariables: [
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage
                    },
                    servingPort: {
                        protocol: huggingFaceDefaultProtocol,
                        container: huggingFaceDefaultPort
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                // check that the environment variables hide the api key

                await hfElementValidationUtils.expectEnvironmentVariables([
                    {
                        name: 'HF_HOME',
                        value: hfHomeDefaultValue
                    }
                ]);

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                expect(inferenceWorkload.spec.image).to.be.equals(tgiImage);
                expect(inferenceWorkload.spec.args).to.be.equals(
                    `--model-id ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceWorkload.spec.servingPort.container).to.be.equals(huggingFaceDefaultPort);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equals(huggingFaceDefaultProtocol);
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equals(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equals(1);
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    {
                        name: 'HF_HOME',
                        value: hfHomeDefaultValue,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                logger.info('test inference template passed!');
            });

            test('should create new hugging face inference using an existing template with model with token', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('hf-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image: vllmImage,
                    args: `--model ${huggingFaceModelNameWithToken}`,
                    environmentVariables: [
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue
                        },
                        {
                            name: 'HF_TOKEN',
                            value: ''
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage
                    },
                    servingPort: {
                        protocol: huggingFaceDefaultProtocol,
                        container: huggingFaceDefaultPort
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                // check that the environment variables hide the api key
                await hfElementValidationUtils.expectEnvironmentVariables([
                    {
                        name: 'HF_HOME',
                        value: hfHomeDefaultValue
                    },
                    {
                        name: 'HF_TOKEN',
                        value: ''
                    }
                ]);

                await inferenceFlatUIUtil.fillEnvironmentData({
                    runtimeSettings: {
                        environmentVariablesValuesByName: new Map([['HF_TOKEN', hfToken]])
                    }
                });

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                expect(inferenceWorkload.spec.image).to.be.equals(vllmImage);
                expect(inferenceWorkload.spec.args).to.be.equals(`--model ${huggingFaceModelNameWithToken}`);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equals(huggingFaceDefaultPort);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equals(huggingFaceDefaultProtocol);
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equals(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equals(1);
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    {
                        name: 'HF_HOME',
                        value: hfHomeDefaultValue,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                    {
                        name: 'HF_TOKEN',
                        value: undefined,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                logger.info('test inference template passed!');
            });

            test('should create new hugging face inference using an existing template with model with credentials', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('hf-template')
                };

                const secretName = objectGenerator.generateNameByPrefix('generic-secret');
                const genericSecret: GenericSecret = await createGenericSecret(
                    secretName,
                    Scope.Project,
                    project.id,
                    undefined,
                    [{ key: 'HF_TOKEN', value: hfToken }]
                );
                const genericSecretFullName = `genericsecret-${genericSecret.meta.name}`;

                const templateSpec: InferenceSpecSpec = {
                    image: vllmImage,
                    args: `--model ${huggingFaceModelNameWithToken}`,
                    environmentVariables: [
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue
                        },
                        {
                            name: 'HF_TOKEN',
                            secret: {
                                name: genericSecretFullName,
                                key: 'HF_TOKEN'
                            }
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage
                    },
                    servingPort: {
                        protocol: huggingFaceDefaultProtocol,
                        container: huggingFaceDefaultPort
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                expect(inferenceWorkload.spec.image).to.be.equals(vllmImage);
                expect(inferenceWorkload.spec.args).to.be.equals(`--model ${huggingFaceModelNameWithToken}`);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equals(huggingFaceDefaultPort);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equals(huggingFaceDefaultProtocol);
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equals(1);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equals(1);
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    {
                        name: 'HF_HOME',
                        value: hfHomeDefaultValue,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                    {
                        name: 'HF_TOKEN',
                        value: undefined,
                        secret: {
                            name: genericSecretFullName,
                            key: 'HF_TOKEN'
                        },
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                logger.info('test inference template passed!');
            });

            test('should create new hugging face inference using an existing template with serving port values', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('hf-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image: vllmImage,
                    args: `--model ${huggingFaceModelNameWithoutToken}`,
                    environmentVariables: [
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage
                    },
                    servingPort: {
                        protocol: huggingFaceDefaultProtocol,
                        container: huggingFaceDefaultPort,
                        // only the access can be changed
                        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                        authorizedGroups: ['group1', 'group2']
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                // check that the environment variables hide the api key
                await hfElementValidationUtils.expectEnvironmentVariables([
                    {
                        name: 'HF_HOME',
                        value: hfHomeDefaultValue
                    }
                ]);

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                expect(inferenceWorkload.spec.image).to.be.equals(vllmImage);
                expect(inferenceWorkload.spec.args).to.be.equals(
                    `--model ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceWorkload.spec.servingPort.container).to.be.equals(huggingFaceDefaultPort);
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equals(huggingFaceDefaultProtocol);
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.be.equals(
                    templateSpec.servingPort.authorizationType
                );
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.be.deep.equal(
                    templateSpec.servingPort.authorizedGroups
                );
                logger.info('test inference template passed!');
            });

            test('should create new hugging face inference using an existing template with compute values', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('nim-template')
                };

                const autoscaling = {
                    minReplicas: 0,
                    maxReplicas: 5,
                    metric: 'concurrency',
                    metricThreshold: 50,
                    scaleToZeroRetentionSeconds: 300
                };
                const compute = {
                    gpuDevicesRequest: 2,
                    cpuCoreRequest: 1,
                    cpuCoreLimit: 2
                };

                const templateSpec: InferenceSpecSpec = {
                    image: vllmImage,
                    args: `--model ${huggingFaceModelNameWithoutToken}`,
                    environmentVariables: [
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage
                    },
                    servingPort: {
                        protocol: huggingFaceDefaultProtocol,
                        container: huggingFaceDefaultPort
                    },
                    compute,
                    autoscaling
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                // compute
                expect(inferenceWorkload.spec.compute.cpuCoreRequest).to.equal(compute.cpuCoreRequest);
                expect(inferenceWorkload.spec.compute.cpuCoreLimit).to.equal(compute.cpuCoreLimit);
                expect(inferenceWorkload.spec.compute.gpuDevicesRequest).to.equal(compute.gpuDevicesRequest);

                // auto scaling
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.equal(autoscaling.minReplicas);
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.equal(autoscaling.maxReplicas);
                expect(inferenceWorkload.spec.autoscaling.metric).to.equal(autoscaling.metric);
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.equal(
                    autoscaling.metricThreshold
                );
                expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(
                    autoscaling.scaleToZeroRetentionSeconds
                );

                logger.info('test inference template passed!');
            });

            test('should create new hugging face inference using an existing template with pvc storage', async function () {
                const pvcName = objectGenerator.generateNameByPrefix('inf-pvc-storage');
                const pvcSpec = {
                    path: '/home',
                    claimName: objectGenerator.generateNameByPrefix('inf-pvc-storage-claim'),
                    existingPvc: false,
                    readOnly: false,
                    ephemeral: false,
                    claimInfo: {
                        size: '100M',
                        accessModes: {
                            readWriteOnce: true
                        },
                        volumeMode: PvcVolumeMode.Filesystem
                    }
                };

                const pvc = await createPvcAsset(pvcName, project.id, pvcSpec);

                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('hf-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image: vllmImage,
                    args: `--model ${huggingFaceModelNameWithoutToken}`,
                    environmentVariables: [
                        {
                            name: 'HF_HOME',
                            value: pvc.spec.path
                        }
                    ],
                    security: {
                        uidGidSource: UidGidSource.FromTheImage
                    },
                    servingPort: {
                        protocol: huggingFaceDefaultProtocol,
                        container: huggingFaceDefaultPort
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1
                    },
                    storage: {
                        pvc: [
                            {
                                claimName: pvc.spec.claimName,
                                existingPvc: true,
                                path: pvc.spec.path,
                                readOnly: false
                            }
                        ]
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.openAndFillInferenceFirstStep({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                // check that the environment variables hide the api key
                await hfElementValidationUtils.expectEnvironmentVariables([
                    {
                        name: 'HF_HOME',
                        value: pvc.spec.path
                    }
                ]);

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);
                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                expect(inferenceWorkload.spec.image).to.be.equals(vllmImage);
                expect(inferenceWorkload.spec.args).to.be.equals(
                    `--model ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceWorkload.spec.environmentVariables).to.deep.equal([
                    {
                        name: 'HF_HOME',
                        value: pvc.spec.path,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);
                expect(inferenceWorkload.spec.storage.pvc[0]).to.be.exist;
                expect(inferenceWorkload.spec.storage.pvc[0].claimName).to.be.equal(pvc.spec.claimName);
                expect(inferenceWorkload.spec.storage.pvc[0].existingPvc).to.be.true;
                expect(inferenceWorkload.spec.storage.pvc[0].path).to.be.equal(pvc.spec.path);
                expect(inferenceWorkload.spec.storage.pvc[0].readOnly).to.be.false;

                logger.info('test inference template passed!');
            });
        });
    }
);
