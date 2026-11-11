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
    workloadsUtilsV2
} from '../../../../../../../../modules/helpers';
import { Project } from '../../../../../../../../modules/api/apis/orgUnitApi';
import { isKnativeInstalled } from '../../../../../../../../modules/helpers/utils/IsInstalledUtils';
import { test, TestInfo } from '@playwright/test';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { WorkloadTemplatePage } from '../../../../../../../../modules/ui/runaiV2/pages/WorkloadTemplatePage';
import { IInferenceTemplateMeta } from '../../../../../../../../modules/helpers/models/inferenceTemplateUI.models';
import {
    ECredentialsSource,
    EVolumeAccessMode,
    IAutoScalingOptions,
    IComputeOptions,
    IDataSourceOptions,
    IEnvironmentOptions,
    IInferenceType,
    IPVCOptions,
    IServingPortAccessOptions
} from '../../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { WorkloadType } from '../../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadType';
import { expect } from 'chai';
import { InferenceTemplateUIUtil } from '../../../../../../../../modules/ui/utils/inference/inferenceTemplateUI.util';
import { InferenceTemplate } from '../../../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceTemplate';
import {
    EnvironmentVariable,
    GenericSecret,
    GenericSecretCreationRequest,
    KeyValuePair,
    PVCAsset,
    PVCAssetSpec,
    PVCCreationRequest,
    PvcVolumeMode,
    Scope,
    ScopeType,
    ServingPortAccessAuthorizationTypeEnum,
    ServingPortProtocol,
    UidGidSource
} from '../../../../../../../../modules/api/apis/assetsApi';
import { HuggingFaceElementValidationUtils } from '../../../../../../../../modules/ui/utils/inference/HuggingFaceElementValidationUtils';
import { InferenceSpecSpec } from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';

test.describe(
    `Inferences Template HF UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        let project: Project;
        let namespace: string;
        let videoName: string;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloadTemplate: WorkloadTemplatePage;
        let inferenceTemplateUIUtil: InferenceTemplateUIUtil;
        let templateName: string;
        let templateMeta: IInferenceTemplateMeta;

        let huggingFaceElementValidationUtils: HuggingFaceElementValidationUtils;

        const huggingFaceModelNameWithoutToken = 'microsoft/DialoGPT-small';
        const huggingFaceModelNameWithToken = 'meta-llama/Llama-2-13b-hf';
        const vllmImage = 'vllm/vllm-openai:latest';
        const tgiImage = 'ghcr.io/huggingface/text-generation-inference:latest';
        const hfToken = process.env.HUGGING_FACE_TOKEN;
        const huggingFaceDefaultPort = 8000;
        const huggingFaceDefaultProtocol = 'http';
        const hfHomeDefaultValue = '/tmp/huggingface';

        const currentUser: User = flowRootState.users.system_administrator;

        async function createGenericSecret(
            secretName: string,
            scope: Scope,
            projectId: number | undefined,
            clusterId: string | undefined,
            keyValuePairs: Array<KeyValuePair>
        ): Promise<GenericSecret> {
            const genericSecret: GenericSecretCreationRequest = {
                meta: {
                    name: secretName,
                    description: null,
                    scope: scope,
                    projectId: projectId,
                    clusterId: clusterId,
                    departmentId: null
                },
                spec: { existingSecretName: null, keyValuePairs }
            };

            const genericSecretResponse = await credentialUtils.createGenericSecretAsset(genericSecret);
            await credentialUtils.waitForGenericSecretAssetToBeReady(genericSecretResponse);
            return genericSecretResponse;
        }

        async function createPvcAsset(name: string, spec: PVCAssetSpec): Promise<PVCAsset> {
            const pvc: PVCCreationRequest = {
                meta: {
                    name: name,
                    scope: Scope.Project,
                    projectId: parseInt(project.id)
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

        function expectHuggingFaceWorkload(
            inferenceTemplate: InferenceTemplate,
            templateName: string,
            image: string,
            args: string,
            environmentVariables: Array<EnvironmentVariable>,
            checkNfs: boolean = true
        ): void {
            expect(inferenceTemplate).to.not.be.undefined;
            logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

            expect(inferenceTemplate.meta.name).to.equal(templateName);
            expect(inferenceTemplate.spec.image).to.be.equal(image);
            expect(inferenceTemplate.spec.args).to.be.equal(args);
            expect(inferenceTemplate.spec.environmentVariables).to.deep.equal(environmentVariables);
            expect(inferenceTemplate.spec.servingPort.container).to.equal(huggingFaceDefaultPort);
            expect(inferenceTemplate.spec.servingPort.protocol).to.equal(huggingFaceDefaultProtocol);
            expect(inferenceTemplate.spec.servingPort.authorizationType).to.equal('authorizedUsersOrGroups');
            expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);
            expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

            if (checkNfs) {
                expect(inferenceTemplate.spec.storage.nfs).to.not.be.undefined;
                const nfs = inferenceTemplate.spec.storage.nfs[0];
                expect(nfs).to.not.be.undefined;
                expect(nfs.path).to.equal('/test/sample');
                expect(nfs.readOnly).to.be.false;
                expect(nfs.mountPath).to.equal('/home/local/data');
                expect(nfs.server).to.equal('test-server');
            }
        }

        async function createTemplate(
            metaConfig: {
                workloadType: WorkloadType;
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

        test.beforeAll('before all create inference template based on Hugging Face UI', async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            browserEngine = new BrowserEngine();

            const projectName = objectGenerator.generateNameByPrefix('project-e2e-nim-inf-template');
            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 1, limit: 1 } }]
            });
            namespace = project.status.namespace;
            logger.info(`created project ${projectName}`);
        });

        test.afterAll('after all create inference template based on Hugging Face UI', async function () {
            if (!(await runaiHealthChecker.isSelfHostedCluster())) {
                // This condition should be removed when fixing problem with saas envs with cluster version.
                return;
            }

            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
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

        test.beforeEach('before each create inference template based on Hugging Face UI', async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloadTemplate = new WorkloadTemplatePage(browserEngine.currentPage);
            inferenceTemplateUIUtil = new InferenceTemplateUIUtil(browserEngine.currentPage);
            huggingFaceElementValidationUtils = new HuggingFaceElementValidationUtils(
                browserEngine.currentPage
            );
            await auth.signIn(currentUser.email, currentUser.password);

            templateName = objectGenerator.generateNameByPrefix('inference-template');
            await workloadTemplate.navigateToWorkloadTemplates();

            templateMeta = {
                projectName: project.name,
                inferenceType: IInferenceType.VLLM,
                templateName: templateName
            };
        });

        test.afterEach(
            'after each create inference template based on Hugging Face UI',
            async ({}, testInfo) => {
                await browserEngine.stopAndArchive(testInfo, 'inferences-template-' + videoName);

                logger.debug(`waiting for template ${templateName} to be deleted`);

                if (!templateName) {
                    logger.warn('template name is not set, skipping cleanup');
                    return;
                }

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                if (!inferenceTemplate) {
                    logger.warn(`template ${templateName} not found, skipping cleanup`);
                    return;
                }
                await templateUtils.deleteWorkloadTemplateV2(
                    inferenceTemplate.meta.id,
                    WorkloadType.Inference
                );

                logger.debug(`inference template ${templateName} deleted!`);
            }
        );

        supportedFrom(SupportedVersions.V2_23)(function () {
            test('should create new HF template with default values only without requited token', async function () {
                videoName = 'should-create-new-hf-template-with-default-values-only-without-requited-token';

                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithoutToken
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // HF specific data
                expect(inferenceTemplate.spec.image).to.be.equal(vllmImage);
                expect(inferenceTemplate.spec.args).to.be.equal(
                    `--model ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
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

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(huggingFaceDefaultPort);
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal(huggingFaceDefaultProtocol);
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                // other data
                expect(inferenceTemplate.spec.imagePullPolicy).to.be.equal('IfNotPresent');

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new HF template with TGI server type', async function () {
                videoName = 'should-create-new-hf-template-with-tgi-server-type';

                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta: {
                        ...templateMeta,
                        inferenceType: IInferenceType.TGI,
                    },
                    modelSection: {
                        modelName: huggingFaceModelNameWithoutToken,
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // HF specific data
                expect(inferenceTemplate.spec.image).to.be.equal(tgiImage);
                expect(inferenceTemplate.spec.args).to.be.equal(
                    `--model-id ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
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

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(huggingFaceDefaultPort);
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal(huggingFaceDefaultProtocol);
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                // other data
                expect(inferenceTemplate.spec.imagePullPolicy).to.be.equal('IfNotPresent');

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new HF template with default values only when token is requited', async function () {
                videoName = 'should-create-new-hf-template-with-default-values-only-when-token-is-requited';

                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithToken
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // HF specific data
                expect(inferenceTemplate.spec.image).to.be.equal(vllmImage);
                expect(inferenceTemplate.spec.args).to.be.equal(`--model ${huggingFaceModelNameWithToken}`);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
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

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(huggingFaceDefaultPort);
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal(huggingFaceDefaultProtocol);
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                // other data
                expect(inferenceTemplate.spec.imagePullPolicy).to.be.equal('IfNotPresent');

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new HF template with credentials', async function () {
                videoName = 'should-create-new-hf-template-with-credentials';

                const secretName = objectGenerator.generateNameByPrefix('generic-secret');
                const genericSecret: GenericSecret = await createGenericSecret(
                    secretName,
                    Scope.Project,
                    parseInt(project.id),
                    undefined,
                    [{ key: 'HF_TOKEN', value: hfToken }]
                );

                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithToken,
                        credentials: {
                            source: ECredentialsSource.Shared,
                            secretName,
                            sourceDisabled: true
                        }
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // HF specific data
                expect(inferenceTemplate.spec.image).to.be.equal(vllmImage);
                expect(inferenceTemplate.spec.args).to.be.equal(`--model ${huggingFaceModelNameWithToken}`);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
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
                        secret: {
                            name: `genericsecret-${genericSecret.meta.name}`,
                            key: 'HF_TOKEN'
                        },
                        value: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                ]);

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(huggingFaceDefaultPort);
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal(huggingFaceDefaultProtocol);
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                // other data
                expect(inferenceTemplate.spec.imagePullPolicy).to.be.equal('IfNotPresent');

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new HF template with serving port access to specific users', async function () {
                videoName = 'should-create-new-hf-template-with-serving-port-access-to-specific-users';

                const servingPort = {
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };
                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithoutToken,
                        servingPort,
                        shouldOpenServingPortSection: true
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // HF specific data
                expect(inferenceTemplate.spec.image).to.be.equal(vllmImage);
                expect(inferenceTemplate.spec.args).to.be.equal(
                    `--model ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
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

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceTemplate.spec.servingPort.container).to.equal(huggingFaceDefaultPort);
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal(huggingFaceDefaultProtocol);
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.equal(
                    'authorizedUsersOrGroups'
                );
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);

                // other data
                expect(inferenceTemplate.spec.imagePullPolicy).to.be.equal('IfNotPresent');

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create HF template with custom compute resource', async function () {
                test.setTimeout(300 * 1000); // this test takes longer than the default timeout

                videoName = 'create-hf-template-with-custom-compute-resource';

                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithoutToken
                    },
                    computeSection: {
                        gpu: {
                            devicesInput: '3'
                        }
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // HF specific data
                expect(inferenceTemplate.spec.image).to.be.equal(vllmImage);
                expect(inferenceTemplate.spec.args).to.be.equal(
                    `--model ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
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

                // compute section
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.compute.gpuDevicesRequest).to.be.equal(3);
                expect(inferenceTemplate.spec.compute.cpuMemoryRequest).to.be.equal('0M');

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create HF template with compute resource from asset', async function () {
                videoName = 'create-hf-template-with-compute-resource-from-asset';

                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithoutToken
                    },
                    computeSection: {
                        assetName: 'cpu-only'
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // HF specific data
                expect(inferenceTemplate.spec.image).to.be.equal(vllmImage);
                expect(inferenceTemplate.spec.args).to.be.equal(
                    `--model ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
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

                // compute section
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.compute.cpuCoreRequest).to.be.equal(0.1);
                expect(inferenceTemplate.spec.compute.cpuMemoryRequest).to.be.equal('100M');

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create HF template with compute resource from new asset', async function () {
                videoName = 'create-hf-template-with-compute-resource-from-new-asset';

                const computeName = objectGenerator.generateNameByPrefix('custom-compute');
                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithToken,
                        hfToken
                    },
                    computeSection: {
                        newComputeResource: {
                            assetName: computeName,
                            projectName: '',
                            gpu: {
                                devicesInput: 3
                            },
                            isProjectScope: false
                        }
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // HF specific data
                expect(inferenceTemplate.spec.image).to.be.equal(vllmImage);
                expect(inferenceTemplate.spec.args).to.be.equal(`--model ${huggingFaceModelNameWithToken}`);
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
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
                        value: hfToken,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                ]);

                // compute section
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.compute.gpuDevicesRequest).to.be.equal(1);
                expect(inferenceTemplate.spec.compute.gpuPortionRequest).to.be.equal(0.03);
                expect(inferenceTemplate.spec.compute.gpuRequestType).to.be.equal('portion');

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new HF template with custom auto-scaling', async function () {
                videoName = 'should-create-new-hf-template-with-custom-auto-scaling';

                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 0,
                    maxReplicas: 4,
                    thresholdValue: 50,
                    autoScalingPeriod: '1m'
                };

                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithoutToken
                    },
                    computeSection: {
                        gpu: {
                            devicesInput: '3'
                        },
                        autoScaling
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // compute section
                expect(inferenceTemplate.spec.compute.cpuCoreRequest).to.equal(0);
                expect(inferenceTemplate.spec.compute.cpuMemoryRequest).to.equal('0M');
                expect(inferenceTemplate.spec.compute.gpuDevicesRequest).to.equal(3);

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(0);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(autoScaling.maxReplicas);
                expect(inferenceTemplate.spec.autoscaling.metric).to.equal('throughput');
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.equal(
                    autoScaling.thresholdValue
                );
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(60); // 1m

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new HF template with new nfs from drawer', async function () {
                videoName = 'should-create-new-hf-template-with-new-nfs-from-drawer';

                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                const dataSourceOptions: IDataSourceOptions = {
                    nfs: dataSourceName
                };
                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithoutToken
                    },
                    dataSourceOptions
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // HF specific data
                expect(inferenceTemplate.spec.image).to.be.equal(vllmImage);
                expect(inferenceTemplate.spec.args).to.be.equal(
                    `--model ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    {
                        name: 'HF_HOME',
                        value: '/test/sample', // this value is taken from the data storage path
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);

                // storage
                expect(inferenceTemplate.spec.storage.nfs).to.not.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new HF template with pvc storage', async function () {
                videoName = 'create-new-hf-template-with-pvc-storage';

                const dataSourceOptions: IDataSourceOptions = {
                    volume: {
                        accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                        sizeUnit: 'MB',
                        path: '/home'
                    }
                };

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
                const pvc = await createPvcAsset(pvcName, pvcSpec);

                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithoutToken
                    },
                    dataSourceOptions: {
                        assetNames: [pvc.meta.name]
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // HF specific data
                expect(inferenceTemplate.spec.image).to.be.equal(vllmImage);
                expect(inferenceTemplate.spec.args).to.be.equal(
                    `--model ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    {
                        name: 'HF_HOME',
                        value: dataSourceOptions.volume.path, // this value is taken from the data storage path
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);

                expect(inferenceTemplate.spec.storage.pvc).to.deep.equal([
                    {
                        claimInfo: undefined,
                        claimName: pvcSpec.claimName,
                        dataSharing: undefined,
                        ephemeral: false,
                        exclude: undefined,
                        existingPvc: true,
                        name: undefined,
                        path: dataSourceOptions.volume.path,
                        readOnly: false
                    }
                ]);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);

                await dataSourcesUtils.deletePvcAsset(pvc.meta.id);
                await dataSourcesUtils.waitForDataSourceToBeDeleted(pvcName);
            });

            test('should create new HF template with new pvc from drawer', async function () {
                videoName = 'create-new-hf-template-with-new-pvc';

                const pvcOptions: IPVCOptions = {
                    name: objectGenerator.generateNameByPrefix('pvc-test'),
                    storageClass: 'none',
                    accessMode: 'Read-write by one node',
                    claimSize: '1',
                    claimUnit: 'GB',
                    containerPath: '/home',
                    volumeMode: 'Filesystem'
                };
                const dataSourceOptions: IDataSourceOptions = {
                    pvc: pvcOptions
                };

                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithoutToken
                    },
                    dataSourceOptions
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                expect(inferenceTemplate.spec.environmentVariables).to.deep.equal([
                    {
                        name: 'HF_HOME',
                        value: pvcOptions.containerPath,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }
                ]);

                expect(inferenceTemplate.spec.storage.pvc[0]).to.be.exist;
                expect(inferenceTemplate.spec.storage.pvc[0].path).to.be.equal(pvcOptions.containerPath);
                expect(inferenceTemplate.spec.storage.pvc[0].existingPvc).to.be.true;
                expect(inferenceTemplate.spec.storage.pvc[0].claimName).to.be.contains(pvcOptions.name);
                expect(inferenceTemplate.spec.storage.pvc[0].readOnly).to.be.false;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should continue to advanced setup when create hugging face template', async function () {
                videoName = 'continue-to-advanced-setup-when-create-hugging-face-template';

                const servingPort = {
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };

                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithToken,
                        hfToken,
                        servingPort,
                        shouldOpenServingPortSection: true
                    },
                    computeSection: {
                        autoScaling: {
                            minReplicas: 0,
                            maxReplicas: 4,
                            thresholdValue: 97,
                            autoScalingPeriod: '1m'
                        }
                    },
                    dataSourceOptions: {
                        nfs: dataSourceName
                    },
                    advancedSetup: true
                });

                // check that the environment variables are filled correctly
                const environmentVariables = [
                    {
                        name: 'HF_HOME',
                        value: '/test/sample', // this value is taken from the data storage path
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                    {
                        name: 'HF_TOKEN',
                        value: hfToken,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                ];
                await huggingFaceElementValidationUtils.expectEnvironmentVariables(environmentVariables);
                await inferenceTemplateUIUtil.clickOnCreateCustomTemplateBtn(templateName);

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expectHuggingFaceWorkload(
                    inferenceTemplate,
                    templateName,
                    vllmImage,
                    `--model ${huggingFaceModelNameWithToken}`,
                    environmentVariables
                );
                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should continue to advanced setup and create new data source for hugging face template', async function () {
                videoName = 'continue-to-advanced-setup-and-create-new-data-source-for-hugging-face-template';

                const servingPort = {
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };

                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithToken,
                        hfToken,
                        servingPort,
                        shouldOpenServingPortSection: true
                    },
                    computeSection: {
                        autoScaling: {
                            minReplicas: 0,
                            maxReplicas: 4,
                            thresholdValue: 97,
                            autoScalingPeriod: '1m'
                        }
                    },
                    advancedSetup: true
                });

                // we want to make sure that all the data is still there
                const environmentVariables = [
                    {
                        name: 'HF_HOME',
                        value: hfHomeDefaultValue, // default value when no data source is provided
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                    {
                        name: 'HF_TOKEN',
                        value: hfToken,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                ];
                await huggingFaceElementValidationUtils.expectEnvironmentVariables(environmentVariables);

                await inferenceTemplateUIUtil.fillDataSourceData({ nfs: dataSourceName });
                await inferenceTemplateUIUtil.clickOnCreateCustomTemplateBtn(templateName);

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expectHuggingFaceWorkload(
                    inferenceTemplate,
                    templateName,
                    vllmImage,
                    `--model ${huggingFaceModelNameWithToken}`,
                    environmentVariables // here the HF_HOME is still the default value
                );
                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should support modifying advanced setup fields, navigating back to hugging face configuration, and completing submission with the origin data', async function () {
                videoName =
                    'support-modifying-advanced-setup-fields-navigating-back-to-hugging-face-configuration-and-completing-submission-with-the-origin-data';

                const servingPort = {
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };

                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                await inferenceTemplateUIUtil.createHFTemplate({
                    templateMeta,
                    modelSection: {
                        modelName: huggingFaceModelNameWithToken,
                        hfToken,
                        servingPort,
                        shouldOpenServingPortSection: true
                    },
                    computeSection: {
                        autoScaling: {
                            minReplicas: 0,
                            maxReplicas: 4,
                            thresholdValue: 97,
                            autoScalingPeriod: '1m'
                        }
                    },
                    advancedSetup: true
                });

                // check that the environment variables are filled correctly
                const environmentVariables = [
                    {
                        name: 'HF_HOME',
                        value: hfHomeDefaultValue, // default value when no data source is provided
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    },
                    {
                        name: 'HF_TOKEN',
                        value: hfToken,
                        secret: undefined,
                        configMap: undefined,
                        podFieldRef: undefined,
                        userCredential: undefined,
                        description: undefined,
                        exclude: undefined
                    }                    
                ];

                await huggingFaceElementValidationUtils.expectEnvironmentVariables(environmentVariables);

                // change the environment and compute data
                await inferenceTemplateUIUtil.fillEnvironmentData({ assetName: 'gpt2' });
                await inferenceTemplateUIUtil.fillComputeData({ assetName: 'two-gpus' });
                await inferenceTemplateUIUtil.fillDataSourceData({ nfs: dataSourceName });

                // go back to the hugging face page
                await workloadTemplate.locator('[aid="model-section"] [aid="back-btn"]').click();

                // submit the inference
                await inferenceTemplateUIUtil.clickOnCreateCustomTemplateBtn(templateName);

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expectHuggingFaceWorkload(
                    inferenceTemplate,
                    templateName,
                    vllmImage,
                    `--model ${huggingFaceModelNameWithToken}`,
                    [
                        {
                            name: 'HF_HOME',
                            value: hfHomeDefaultValue, // default value when no data source is provided
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        },
                        {
                            name: 'HF_TOKEN',
                            value: hfToken,
                            secret: undefined,
                            configMap: undefined,
                            podFieldRef: undefined,
                            userCredential: undefined,
                            description: undefined,
                            exclude: undefined
                        }
                    ],
                    false
                );
                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should copy an existing inference hf template', async function ({}, testInfo: TestInfo) {
                videoName = 'should-copy-an-existing-inference-hf-template';

                const copiedTemplateName = objectGenerator.generateNameByPrefix('inf-template-copy');
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
                const pvc = await createPvcAsset(pvcName, pvcSpec);
                const pvcStorage = {
                    path: pvc.spec.path,
                    existingPvc: true,
                    claimName: pvc.spec.claimName,
                    readOnly: pvc.spec.readOnly
                };

                const metaConfigOriginal = {
                    workloadType: WorkloadType.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName
                };

                const servingPort = {
                    protocol: huggingFaceDefaultProtocol,
                    container: huggingFaceDefaultPort,
                    authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                    authorizedGroups: ['group1', 'group2']
                };

                const autoscaling = {
                    minReplicas: 0,
                    maxReplicas: 2,
                    metric: 'concurrency',
                    metricThreshold: 98,
                    scaleToZeroRetentionSeconds: 300
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
                        protocol: ServingPortProtocol.Http,
                        container: huggingFaceDefaultPort,
                        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                        authorizedGroups: ['group1', 'group2']
                    },
                    compute: {
                        gpuDevicesRequest: 3
                    },
                    autoscaling,
                    storage: {
                        pvc: [pvcStorage]
                    }
                };

                const originTemplate = await createTemplate(metaConfigOriginal, templateSpec);
                expect(originTemplate).to.not.be.undefined;
                logger.debug(`origin template: ${JSON.stringify(originTemplate)}`);

                await workloadTemplate.page.reload();
                await workloadTemplate.toggleTemplate(metaConfigOriginal.templateName);
                await workloadTemplate.clickOnCopyTemplate();

                await workloadTemplate.fillTemplateName(copiedTemplateName);
                await workloadTemplate.clickOnCreateInferenceTemplateBtn();

                await workloadTemplate.page.waitForTimeout(300);
                const inferenceTemplateCopied: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, copiedTemplateName);

                expect(inferenceTemplateCopied).to.not.be.undefined;
                logger.debug(`copied template: ${JSON.stringify(inferenceTemplateCopied)}`);

                expect(inferenceTemplateCopied.meta.name).to.equal(copiedTemplateName);

                // model section
                expect(inferenceTemplateCopied.spec.image).to.be.equals(vllmImage);
                expect(inferenceTemplateCopied.spec.args).to.be.equals(
                    `--model ${huggingFaceModelNameWithoutToken}`
                );
                expect(inferenceTemplateCopied.spec.environmentVariables).to.deep.equal([
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

                // serving port
                expect(inferenceTemplateCopied.spec.servingPort.container).to.equal(servingPort.container);
                expect(inferenceTemplateCopied.spec.servingPort.protocol).to.equal(servingPort.protocol);
                expect(inferenceTemplateCopied.spec.servingPort.authorizationType).to.equal(
                    servingPort.authorizationType
                );
                expect(inferenceTemplateCopied.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplateCopied.spec.servingPort.authorizedGroups).to.deep.equals(
                    servingPort.authorizedGroups
                );

                // compute section
                expect(inferenceTemplateCopied.spec.compute.gpuDevicesRequest).to.equal(3);

                // auto scaling
                expect(inferenceTemplateCopied.spec.autoscaling.minReplicas).to.equal(
                    autoscaling.minReplicas
                );
                expect(inferenceTemplateCopied.spec.autoscaling.maxReplicas).to.equal(
                    autoscaling.maxReplicas
                );
                expect(inferenceTemplateCopied.spec.autoscaling.metric).to.equal(autoscaling.metric);
                expect(inferenceTemplateCopied.spec.autoscaling.metricThreshold).to.equal(
                    autoscaling.metricThreshold
                );
                expect(inferenceTemplateCopied.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(
                    autoscaling.scaleToZeroRetentionSeconds
                );

                // storage section
                expect(inferenceTemplateCopied.spec.storage.pvc[0]).to.be.exist;
                expect(inferenceTemplateCopied.spec.storage.pvc[0].path).to.equal(pvcStorage.path);
                expect(inferenceTemplateCopied.spec.storage.pvc[0].existingPvc).to.equal(
                    pvcStorage.existingPvc
                );
                expect(inferenceTemplateCopied.spec.storage.pvc[0].claimName).to.equal(pvcStorage.claimName);
                expect(inferenceTemplateCopied.spec.storage.pvc[0].readOnly).to.equal(pvcStorage.readOnly);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);

                await templateUtils.deleteWorkloadTemplateV2(
                    inferenceTemplateCopied.meta.id,
                    WorkloadType.Inference
                );
            });

            test('should edit an existing inference hf template', async function ({}, testInfo: TestInfo) {
                videoName = 'should-edit-an-existing-inference-hf-template';

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
                const pvc = await createPvcAsset(pvcName, pvcSpec);
                const pvcStorage = {
                    path: pvc.spec.path,
                    existingPvc: true,
                    claimName: pvc.spec.claimName,
                    readOnly: pvc.spec.readOnly
                };

                const metaConfigOriginal = {
                    workloadType: WorkloadType.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName
                };

                const servingPort = {
                    protocol: ServingPortProtocol.Http,
                    container: huggingFaceDefaultPort,
                    authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                    authorizedGroups: ['group1', 'group2']
                };

                const autoscaling = {
                    minReplicas: 0,
                    maxReplicas: 2,
                    metric: 'concurrency',
                    metricThreshold: 98,
                    scaleToZeroRetentionSeconds: 300
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
                    servingPort,
                    compute: {
                        gpuDevicesRequest: 3
                    },
                    autoscaling,
                    storage: {
                        pvc: [pvcStorage]
                    }
                };

                const expectedTemplateSpec: InferenceSpecSpec = {
                    image: vllmImage,
                    args: `--model ${huggingFaceModelNameWithoutToken}`,
                    environmentVariables: [
                        {
                            name: 'HF_HOME',
                            value: pvc.spec.path
                        }
                    ],
                    servingPort: {
                        protocol: ServingPortProtocol.Http,
                        container: 8080,
                        authorizationType: undefined, //when it is public, it will be set to undefined
                        authorizedUsers: undefined,
                        authorizedGroups: undefined
                    },
                    compute: {
                        gpuDevicesRequest: 0,
                        cpuCoreRequest: 0.1,
                        cpuMemoryRequest: '100M'
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 2,
                        metric: 'concurrency',
                        metricThreshold: 54
                    },
                    storage: {
                        pvc: [pvcStorage]
                    }
                };

                function expectTemplate(
                    inferenceTemplate: InferenceTemplate,
                    expectedData: InferenceSpecSpec
                ): void {
                    // model section
                    expect(inferenceTemplate.spec.image).to.equal(expectedData.image);
                    expect(inferenceTemplate.spec.args).to.equal(expectedData.args);

                    // serving port
                    expect(inferenceTemplate.spec.servingPort.container).to.equal(
                        expectedData.servingPort.container
                    );
                    expect(inferenceTemplate.spec.servingPort.protocol).to.equal(
                        expectedData.servingPort.protocol
                    );
                    expect(inferenceTemplate.spec.servingPort.authorizationType).to.equal(
                        expectedData.servingPort.authorizationType
                    );
                    expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                    expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.deep.equals(
                        expectedData.servingPort.authorizedGroups
                    );

                    // compute section
                    expect(inferenceTemplate.spec.compute.gpuDevicesRequest).to.equal(
                        expectedData.compute.gpuDevicesRequest
                    );
                    expect(inferenceTemplate.spec.compute.cpuCoreRequest).to.equal(
                        expectedData.compute.cpuCoreRequest
                    );
                    expect(inferenceTemplate.spec.compute.cpuMemoryRequest).to.equal(
                        expectedData.compute.cpuMemoryRequest
                    );

                    // auto scaling
                    expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(
                        expectedData.autoscaling.minReplicas
                    );
                    expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(
                        expectedData.autoscaling.maxReplicas
                    );
                    expect(inferenceTemplate.spec.autoscaling.metric).to.equal(
                        expectedData.autoscaling.metric
                    );
                    expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.equal(
                        expectedData.autoscaling.metricThreshold
                    );
                    expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.equal(
                        expectedData.autoscaling.scaleToZeroRetentionSeconds
                    );

                    // storage section
                    expect(inferenceTemplate.spec.storage.pvc[0]).to.be.exist;
                    expect(inferenceTemplate.spec.storage.pvc[0].path).to.equal(
                        expectedData.storage.pvc[0].path
                    );
                    expect(inferenceTemplate.spec.storage.pvc[0].existingPvc).to.equal(
                        expectedData.storage.pvc[0].existingPvc
                    );
                    expect(inferenceTemplate.spec.storage.pvc[0].claimName).to.equal(
                        expectedData.storage.pvc[0].claimName
                    );
                    expect(inferenceTemplate.spec.storage.pvc[0].readOnly).to.equal(
                        expectedData.storage.pvc[0].readOnly
                    );
                }

                const originTemplate = await createTemplate(metaConfigOriginal, templateSpec);
                expect(originTemplate).to.not.be.undefined;
                logger.debug(`origin template: ${JSON.stringify(originTemplate)}`);
                logger.debug(`origin template spec: ${JSON.stringify(templateSpec)}`);

                expectTemplate(originTemplate, templateSpec);

                await workloadTemplate.page.reload();
                await workloadTemplate.toggleTemplate(metaConfigOriginal.templateName);
                await workloadTemplate.clickOnEditTemplate();

                const computeData: IComputeOptions = {
                    assetName: 'cpu-only',
                    autoScaling: {
                        minReplicas: expectedTemplateSpec.autoscaling.minReplicas,
                        maxReplicas: expectedTemplateSpec.autoscaling.maxReplicas,
                        thresholdValue: expectedTemplateSpec.autoscaling.metricThreshold
                    }
                };

                const envData: IEnvironmentOptions = {
                    image: expectedTemplateSpec.image,
                    servingPort: {
                        protocol: expectedTemplateSpec.servingPort.protocol.toString(),
                        port: expectedTemplateSpec.servingPort.container,
                        access: {
                            option: IServingPortAccessOptions.Public
                        }
                    }
                };
                await inferenceTemplateUIUtil.fillEnvironmentData(envData);
                await inferenceTemplateUIUtil.fillComputeData(computeData);
                await workloadTemplate.clickOnCreateInferenceTemplateBtn();

                await workloadTemplate.page.waitForTimeout(300);
                const editedTemplate: InferenceTemplate | undefined = await templateUtils.getTemplateV2ByName(
                    WorkloadType.Inference,
                    templateName
                );

                expect(editedTemplate).to.not.be.undefined;
                console.log(`edited template: ${JSON.stringify(editedTemplate)}`);
                logger.debug(`edited template: ${JSON.stringify(editedTemplate)}`);
                logger.debug(`edited template spec: ${JSON.stringify(expectedTemplateSpec)}`);

                expectTemplate(editedTemplate, expectedTemplateSpec);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });
        });
    }
);
