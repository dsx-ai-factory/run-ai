/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
    computeResourceUtils,
    dataSourcesUtils,
    environmentUtils,
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
import { PUBLIC_VLLM_IMAGE } from '../../../../../../../../modules/helpers/config/const';
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
    EImagePullPolicyOptions,
    EVolumeAccessMode,
    IAutoScalingOptions,
    IComputeOptions,
    IDataSourceOptions,
    IEnvironmentOptions,
    IInferenceType,
    IPVCOptions,
    IServingPortAccessOptions,
    IServingPortOptions,
    IVolumeOptions
} from '../../../../../../../../modules/helpers/models/inferenceBaseFlatUI.models';
import { WorkloadType } from '../../../../../../../../modules/api/apis/workloadsApi/generated/model/workloadType';
import { expect } from 'chai';
import { InferenceTemplateUIUtil } from '../../../../../../../../modules/ui/utils/inference/inferenceTemplateUI.util';
import { InferenceTemplate } from '../../../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceTemplate';
import {
    InferenceSpecSpec,
    PvcVolumeMode,
    ScopeType,
    ServingPortAccessAuthorizationTypeEnum,
    ServingPortProtocol,
    WorkloadType as WorkloadTypeApi
} from '../../../../../../../../modules/api/apis/workloadsApi/generated/api';
import { Scope } from '../../../../../../../../modules/api/apis/assetsApi/generated/model/scope';
import { PVCAsset } from '../../../../../../../../modules/api/apis/assetsApi/generated/model/pVCAsset';
import { PVCAssetSpec } from '../../../../../../../../modules/api/apis/assetsApi/generated/model/pVCAssetSpec';
import { PVCCreationRequest } from '../../../../../../../../modules/api/apis/assetsApi/generated/model/pVCCreationRequest';

test.describe(
    `Inferences Template Custom UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        let project: Project;
        let videoName: string;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workloadTemplate: WorkloadTemplatePage;
        let inferenceTemplateUIUtil: InferenceTemplateUIUtil;
        let templateName: string;
        let templateMeta: IInferenceTemplateMeta;

        const currentUser: User = flowRootState.users.system_administrator;
        const llmServerImage = 'vllm/vllm-openai:v0.6.4';
        const llmServerCommand = 'bash';
        const llmServerArgs =
            '-c "if [ -n \\"$RUNAI_MODEL_NAME\\" ]; then exec python3 -m vllm.entrypoints.openai.api_server --model \\"$RUNAI_MODEL\\" --served-model-name \\"$RUNAI_MODEL_NAME\\" --host 0.0.0.0 --port 8000; else exec python3 -m vllm.entrypoints.openai.api_server --model \\"$RUNAI_MODEL\\" --host 0.0.0.0 --port 8000; fi"';

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

        test.beforeAll('before all create inference template based on custom UI', async function () {
            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            browserEngine = new BrowserEngine();

            const projectName = objectGenerator.generateNameByPrefix('project-e2e-custom-inf-template');
            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 1, limit: 1 } }]
            });
            logger.info(`created project ${projectName}`);
        });

        test.afterAll('after all create inference template based on custom UI', async function () {
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

        test.beforeEach('before each create inference template based on custom UI', async function () {
            await browserEngine.start();
            auth = new AuthPage(browserEngine.currentPage);
            workloadTemplate = new WorkloadTemplatePage(browserEngine.currentPage);
            inferenceTemplateUIUtil = new InferenceTemplateUIUtil(browserEngine.currentPage);
            await auth.signIn(currentUser.email, currentUser.password);

            templateName = objectGenerator.generateNameByPrefix('inference-template');
            await workloadTemplate.navigateToWorkloadTemplates();

            templateMeta = {
                clusterName: flowRootState.cluster.name,
                inferenceType: IInferenceType.Custom,
                templateName: templateName
            };
        });

        test.afterEach('after each create inference template based on custom UI', async ({}, testInfo) => {
            await browserEngine.stopAndArchive(testInfo, 'inferences-template-' + videoName);

            logger.debug(`waiting for template ${templateName} to be deleted`);

            if (!templateName) {
                logger.warn('template name is not set, skipping cleanup');
                return;
            }

            const inferenceTemplate: InferenceTemplate | undefined = await templateUtils.getTemplateV2ByName(
                WorkloadType.Inference,
                templateName
            );

            if (!inferenceTemplate) {
                logger.warn(`template ${templateName} not found, skipping cleanup`);
                return;
            }
            await templateUtils.deleteWorkloadTemplateV2(inferenceTemplate.meta.id, WorkloadType.Inference);

            logger.debug(`inference template ${templateName} deleted!`);
        });

        supportedFrom(SupportedVersions.V2_23)(function () {
            test('should create new empty inference template with default values only', async function () {
                videoName = 'should-create-new-empty-inference-template-with-default-values-only';

                await inferenceTemplateUIUtil.createCustomTemplate({
                    templateMeta
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceTemplate.spec.servingPort.container).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal('http');
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new inference template with environment values', async function () {
                videoName = 'should-create-new-inference-template-with-environment-values';

                const servingPort: IServingPortOptions = {
                    port: 4000,
                    protocol: 'gRPC'
                };

                const environmentSection: IEnvironmentOptions = {
                    image: PUBLIC_VLLM_IMAGE,
                    imagePullPolicy: EImagePullPolicyOptions.ImageFromRegistry,
                    tool: 'Jupyter',
                    runtimeSettings: {
                        cmd: 'start-notebook.sh',
                        args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                    },
                    servingPort
                };

                await inferenceTemplateUIUtil.createCustomTemplate({
                    templateMeta,
                    environmentSection
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // environment
                expect(inferenceTemplate.spec.image).to.equal(environmentSection.image);
                expect(inferenceTemplate.spec.imagePullPolicy).to.equal('Always');
                expect(inferenceTemplate.spec.args).to.equal(environmentSection.runtimeSettings.args);
                expect(inferenceTemplate.spec.command).to.equal(environmentSection.runtimeSettings.cmd);
                expect(inferenceTemplate.spec.exposedUrls?.length > 0).to.be.true;

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(1);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceTemplate.spec.servingPort.container).to.equal(servingPort.port);
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal(
                    servingPort.protocol.toLowerCase()
                );
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new inference template using an existing environment without filling the environment variables', async function () {
                videoName =
                    'create-new-inference-template-using-an-existing-environment-without-filling-the-environment-variables';

                const llmServerEnvironment = await environmentUtils.get('llm-server');
                expect(llmServerEnvironment.spec.image).to.equal(llmServerImage);

                const environmentSection: IEnvironmentOptions = {
                    assetName: 'llm-server'
                };

                await inferenceTemplateUIUtil.createCustomTemplate({
                    templateMeta,
                    environmentSection
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // environment section
                expect(inferenceTemplate.spec.image).to.equal(llmServerEnvironment.spec.image);
                expect(inferenceTemplate.spec.imagePullPolicy).to.equal('IfNotPresent');
                expect(inferenceTemplate.spec.args).to.equal(llmServerArgs);
                expect(inferenceTemplate.spec.command).to.equal(llmServerCommand);
                expect(inferenceTemplate.spec.environmentVariables.length).to.equal(2);
                expect(inferenceTemplate.spec.environmentVariables[0].name).to.equal('RUNAI_MODEL');
                expect(inferenceTemplate.spec.environmentVariables[0].value).to.be.undefined;
                expect(inferenceTemplate.spec.environmentVariables[1].name).to.equal('RUNAI_MODEL_NAME');
                expect(inferenceTemplate.spec.environmentVariables[1].value).to.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new flat inference from scratch with compute', async function () {
                videoName = 'should-create-new-flat-inference-with-compute';

                const servingPort: IServingPortOptions = {
                    access: {
                        option: IServingPortAccessOptions.Users
                    }
                };

                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 2,
                    maxReplicas: 4,
                    thresholdValue: 50
                };

                const computeSection: IComputeOptions = {
                    gpu: {
                        devicesInput: '3'
                    },
                    autoScaling
                };

                await inferenceTemplateUIUtil.createCustomTemplate({
                    templateMeta,
                    environmentSection: {
                        servingPort
                    },
                    computeSection
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(autoScaling.minReplicas);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(autoScaling.maxReplicas);
                expect(inferenceTemplate.spec.autoscaling.metric).to.equal('throughput');
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.equal(
                    autoScaling.thresholdValue
                );
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                // serving port
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.equal(
                    'authorizedUsersOrGroups'
                );
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.deep.equal([currentUser.email]);
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new inference template with compute with custom auto-scaling period', async function () {
                videoName =
                    'should-create-new-inference-template-with-compute-with-custom-auto-scaling-period';

                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 0,
                    maxReplicas: 4,
                    thresholdValue: 50,
                    autoScalingPeriod: '1m'
                };

                await inferenceTemplateUIUtil.createCustomTemplate({
                    templateMeta,
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

            test('should create new inference template using an existing environment and compute assets', async function ({}, testInfo: TestInfo) {
                videoName = 'create-new-inference-template-using-an-existing-compute-asset';

                const computeResourceName = objectGenerator.generateShortNameByPrefix('cr-two-gpus', 5);
                const computeCreationRequest = {
                    meta: {
                        name: computeResourceName,
                        scope: Scope.Cluster,
                        clusterId: flowRootState.cluster.uuid
                    },
                    spec: {
                        cpuCoreRequest: 0.2,
                        cpuMemoryRequest: '200M',
                        gpuDevicesRequest: 2
                    }
                };

                const computeAsset = await computeResourceUtils.create(computeCreationRequest);
                logger.debug(`computeAsset created: ${JSON.stringify(computeAsset)}`);

                const llmServerEnvironment = await environmentUtils.get('llm-server');
                expect(llmServerEnvironment.spec.image).to.equal(llmServerImage);

                const environmentSection: IEnvironmentOptions = {
                    assetName: 'llm-server',
                    runtimeSettings: {
                        environmentVariablesValues: ['3000', '4000']
                    }
                };

                const autoScaling: IAutoScalingOptions = {
                    minReplicas: 2,
                    maxReplicas: 2
                };

                await inferenceTemplateUIUtil.createCustomTemplate({
                    templateMeta,
                    environmentSection,
                    computeSection: {
                        assetName: computeResourceName,
                        autoScaling
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                // environment section
                expect(inferenceTemplate.spec.image).to.equal(llmServerEnvironment.spec.image);
                expect(inferenceTemplate.spec.imagePullPolicy).to.equal('IfNotPresent');
                expect(inferenceTemplate.spec.args).to.equal(llmServerArgs);
                expect(inferenceTemplate.spec.command).to.equal(llmServerCommand);
                expect(inferenceTemplate.spec.environmentVariables.length).to.equal(2);
                expect(inferenceTemplate.spec.environmentVariables[0].name).to.equal('RUNAI_MODEL');
                expect(inferenceTemplate.spec.environmentVariables[0].value).to.equal('3000');
                expect(inferenceTemplate.spec.environmentVariables[1].name).to.equal('RUNAI_MODEL_NAME');
                expect(inferenceTemplate.spec.environmentVariables[1].value).to.equal('4000');

                // serving port
                expect(inferenceTemplate.spec.servingPort.container).to.equal(8000);
                expect(inferenceTemplate.spec.servingPort.protocol).to.equal('http');
                expect(inferenceTemplate.spec.servingPort.authorizationType).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedUsers).to.be.undefined;
                expect(inferenceTemplate.spec.servingPort.authorizedGroups).to.be.undefined;

                // compute section
                expect(inferenceTemplate.spec.compute.cpuCoreRequest).to.equal(0.2);
                expect(inferenceTemplate.spec.compute.cpuMemoryRequest).to.equal('200M');
                expect(inferenceTemplate.spec.compute.gpuDevicesRequest).to.equal(2);

                // auto scaling
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.equal(autoScaling.minReplicas);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.equal(autoScaling.maxReplicas);
                expect(inferenceTemplate.spec.autoscaling.metric).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.metricThreshold).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);

                // cleanup
                await computeResourceUtils.deleteByName(computeResourceName);
                await computeResourceUtils.waitForComputeResourceToBeDeleted(computeResourceName);
                logger.debug(`computeResource: ${computeResourceName} deleted successfully!`);
                logger.debug('test completed successfully!');
            });

            test('should create new inference template with new nfs from drawer', async function () {
                videoName = 'create-new-inference-template-with-new-nfs';

                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                const dataSourceOptions: IDataSourceOptions = {
                    nfs: dataSourceName
                };

                await inferenceTemplateUIUtil.createCustomTemplate({
                    templateMeta,
                    dataSourceOptions
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);
                expect(inferenceTemplate.spec.storage.nfs).to.not.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new inference template with new pvc from drawer', async function ({}, testInfo: TestInfo) {
                videoName = 'create-new-inference-template-with-new-pvc';

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

                await inferenceTemplateUIUtil.createCustomTemplate({
                    templateMeta,
                    dataSourceOptions
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);
                expect(inferenceTemplate.spec.storage.pvc[0]).to.be.exist;
                expect(inferenceTemplate.spec.storage.pvc[0].path).to.be.equal(pvcOptions.containerPath);
                expect(inferenceTemplate.spec.storage.pvc[0].existingPvc).to.be.true;
                expect(inferenceTemplate.spec.storage.pvc[0].claimName).to.be.contains(pvcOptions.name);
                expect(inferenceTemplate.spec.storage.pvc[0].readOnly).to.be.false;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new inference template with annotations and labels', async function () {
                videoName = 'should-create-new-inference-template-with-annotations-and-labels';

                await inferenceTemplateUIUtil.createCustomTemplate({
                    templateMeta,
                    generalSectionOptions: {
                        annotations: [
                            { name: 'ann1', value: 'ann-val1' },
                            { name: 'ann2', value: 'ann-val2' }
                        ],
                        labels: [{ name: 'lbl1', value: 'lbl-val1' }]
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);
                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);
                expect(inferenceTemplate.spec.annotations).to.deep.equals([
                    { name: 'ann1', value: 'ann-val1', exclude: undefined },
                    { name: 'ann2', value: 'ann-val2', exclude: undefined }
                ]);
                expect(inferenceTemplate.spec.labels).to.deep.equals([
                    { name: 'lbl1', value: 'lbl-val1', exclude: undefined }
                ]);

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });

            test('should create new inference template with new environment asset and compute resource asset and data sources', async function () {
                videoName =
                    'create-new-inference-template-with-new-environment-asset-and-compute-resource-asset-and-data-sources';

                const environmentName = objectGenerator.generateNameByPrefix('custom-environment');
                const computeName = objectGenerator.generateNameByPrefix('custom-compute');
                const servingPort: IServingPortOptions = {
                    protocol: 'HTTP',
                    port: 8000
                };
                const dataSourceName = objectGenerator.generateNameByPrefix('inf-nfs-storage');
                const volumeOptions: IVolumeOptions = {
                    accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                    sizeUnit: 'MB',
                    path: '/home'
                };

                await inferenceTemplateUIUtil.createCustomTemplate({
                    templateMeta,
                    environmentSection: {
                        newEnvironmentResource: {
                            assetName: environmentName,
                            projectName: project.name,
                            image: PUBLIC_VLLM_IMAGE,
                            servingPort: {
                                container: servingPort.port
                            }
                        }
                    },
                    computeSection: {
                        newComputeResource: {
                            assetName: computeName,
                            projectName: project.name,
                            gpu: {
                                devicesInput: 3
                            },
                            isProjectScope: false
                        }
                    },
                    dataSourceOptions: {
                        nfs: dataSourceName,
                        volume: volumeOptions
                    }
                });

                const inferenceTemplate: InferenceTemplate | undefined =
                    await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                expect(inferenceTemplate).to.not.be.undefined;
                logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);

                expect(inferenceTemplate.spec.image).to.be.equal(PUBLIC_VLLM_IMAGE);
                expect(inferenceTemplate.spec.servingPort.container).to.be.equal(servingPort.port);
                expect(inferenceTemplate.spec.servingPort.protocol).to.be.equal(
                    servingPort.protocol.toLowerCase()
                );
                expect(inferenceTemplate.spec.nodePools).to.be.undefined;
                expect(inferenceTemplate.spec.autoscaling.minReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.autoscaling.maxReplicas).to.be.equal(1);
                expect(inferenceTemplate.spec.compute.gpuDevicesRequest).to.be.equal(1);
                expect(inferenceTemplate.spec.compute.gpuPortionRequest).to.be.equal(0.03);
                expect(inferenceTemplate.spec.compute.gpuRequestType).to.be.equal('portion');
                expect(inferenceTemplate.spec.storage.nfs).to.deep.equal([
                    {
                        path: '/test/sample',
                        readOnly: false,
                        server: 'test-server',
                        mountPath: '/home/local/data',
                        name: undefined,
                        exclude: undefined
                    }
                ]);

                const pvc = inferenceTemplate.spec.storage.pvc[0];
                expect(pvc).to.be.exist;
                expect(pvc.path).to.equal('/home');
                expect(pvc.claimInfo.size).to.equal('1M');
                expect(pvc.claimInfo.storageClass).to.be.exist;
                expect(pvc.claimInfo.accessModes.readWriteOnce).to.equal(true);
                expect(pvc.claimInfo.accessModes.readOnlyMany).to.be.undefined;
                expect(pvc.claimInfo.accessModes.readWriteMany).to.be.undefined;

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);

                // cleanup
                const computeId = await computeResourceUtils.getId(computeName);
                await computeResourceUtils.delete(computeId);
                await computeResourceUtils.waitForComputeResourceToBeDeleted(computeName);
            });

            supportedFrom(SupportedVersions.V2_20)(function () {
                test('should create new inference template with volume', async function () {
                    videoName = 'create-new-inference-template-with-new-volume';

                    const volumeOptions: IVolumeOptions = {
                        accessMode: EVolumeAccessMode.ReadWriteByOneNode,
                        sizeUnit: 'MB',
                        path: '/home'
                    };

                    await inferenceTemplateUIUtil.createCustomTemplate({
                        templateMeta,
                        dataSourceOptions: {
                            volume: volumeOptions
                        }
                    });

                    const inferenceTemplate: InferenceTemplate | undefined =
                        await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                    expect(inferenceTemplate).to.not.be.undefined;
                    logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);
                    expect(inferenceTemplate.spec.storage.pvc).to.not.be.undefined;

                    logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
                });
            });

            supportedFrom(SupportedVersions.V2_23)(function () {
                test('should create new inference template with custom request timeout values', async function () {
                    test.setTimeout(300 * 1000); // this test takes longer than the default timeout

                    videoName = 'create-new-inference-template-with-custom-request-timeout-values';

                    await inferenceTemplateUIUtil.createCustomTemplate({
                        templateMeta,
                        generalSectionOptions: {
                            initializationTimeout: {
                                minutes: 5,
                                seconds: 7
                            },
                            requestTimeout: {
                                minutes: 2,
                                seconds: 30
                            }
                        }
                    });

                    const inferenceTemplate: InferenceTemplate | undefined =
                        await templateUtils.getTemplateV2ByName(WorkloadType.Inference, templateName);

                    expect(inferenceTemplate).to.not.be.undefined;
                    logger.debug(`inferenceTemplate: ${JSON.stringify(inferenceTemplate)}`);
                    expect(
                        inferenceTemplate.spec.servingConfiguration?.initializationTimeoutSeconds
                    ).to.equal(307); // 5 minutes * 60 (seconds) + 7 seconds
                    expect(inferenceTemplate.spec.servingConfiguration?.requestTimeoutSeconds).to.equal(150); // 2 minutes * 60 (seconds) + 30 seconds
                    logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
                });
            });

            test('should copy an existing inference template', async function () {
                videoName = 'should-copy-an-existing-inference-template';

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
                const pvc = await createPvcAsset(pvcName, project.id, pvcSpec);
                const pvcStorage = {
                    path: pvc.spec.path,
                    existingPvc: true,
                    claimName: pvc.spec.claimName,
                    readOnly: pvc.spec.readOnly
                };

                const metaConfigOriginal = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName
                };

                const servingPort = {
                    protocol: ServingPortProtocol.Grpc,
                    container: 3131,
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
                    image: 'ubuntu',
                    command: 'start-notebook.sh',
                    args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""',
                    workingDir: '/home',
                    createHomeDir: true,
                    servingPort,
                    compute: {
                        gpuDevicesRequest: 3
                    },
                    autoscaling,
                    storage: {
                        pvc: [pvcStorage]
                    },
                    annotations: [
                        { name: 'ann1', value: 'ann-val1' },
                        { name: 'ann2', value: 'ann-val2' }
                    ],
                    labels: [{ name: 'lbl1', value: 'lbl-val1' }]
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

                // environment section
                expect(inferenceTemplateCopied.spec.image).to.equal('ubuntu');
                expect(inferenceTemplateCopied.spec.command).to.equal('start-notebook.sh');
                expect(inferenceTemplateCopied.spec.args).to.equal(
                    '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""'
                );
                expect(inferenceTemplateCopied.spec.workingDir).to.equal('/home');
                expect(inferenceTemplateCopied.spec.createHomeDir).to.be.true;

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

                // general section
                expect(inferenceTemplateCopied.spec.annotations).to.deep.equals([
                    { name: 'ann1', value: 'ann-val1', exclude: undefined },
                    { name: 'ann2', value: 'ann-val2', exclude: undefined }
                ]);
                expect(inferenceTemplateCopied.spec.labels).to.deep.equals([
                    { name: 'lbl1', value: 'lbl-val1', exclude: undefined }
                ]);
                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);

                await templateUtils.deleteWorkloadTemplateV2(
                    inferenceTemplateCopied.meta.id,
                    WorkloadType.Inference
                );
            });

            test('should edit an existing inference template', async function () {
                videoName = 'should-edit-an-existing-inference-template';

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
                const pvcStorage = {
                    path: pvc.spec.path,
                    existingPvc: true,
                    claimName: pvc.spec.claimName,
                    readOnly: pvc.spec.readOnly
                };

                const metaConfigOriginal = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName
                };

                const templateSpec: InferenceSpecSpec = {
                    image: 'ubuntu',
                    command: 'start-notebook.sh',
                    args: '--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=""',
                    workingDir: '/home',
                    createHomeDir: true,
                    servingPort: {
                        protocol: ServingPortProtocol.Grpc,
                        container: 3131,
                        authorizationType: ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups,
                        authorizedGroups: ['group1', 'group2']
                    },
                    compute: {
                        gpuDevicesRequest: 3
                    },
                    autoscaling: {
                        minReplicas: 0,
                        maxReplicas: 2,
                        metric: 'concurrency',
                        metricThreshold: 98,
                        scaleToZeroRetentionSeconds: 300
                    },
                    storage: {
                        pvc: [pvcStorage]
                    },
                    annotations: [
                        { name: 'ann1', value: 'ann-val1' },
                        { name: 'ann2', value: 'ann-val2' }
                    ],
                    labels: [{ name: 'lbl1', value: 'lbl-val1' }]
                };

                const expectedTemplateSpec: InferenceSpecSpec = {
                    image: 'ubuntu2',
                    command: 'run-script.sh',
                    args: '--arg1 --arg2',
                    workingDir: '/home',
                    createHomeDir: true,
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
                        pvc: [pvcStorage],
                        nfs: [
                            {
                                path: '/test/sample',
                                server: 'test-server',
                                mountPath: '/home/local/data',
                                readOnly: false
                            }
                        ]
                    },
                    annotations: templateSpec.annotations,
                    labels: templateSpec.labels
                };

                function expectTemplate(
                    inferenceTemplate: InferenceTemplate,
                    expectedData: InferenceSpecSpec
                ): void {
                    // environment section
                    expect(inferenceTemplate.spec.image).to.equal(expectedData.image);
                    expect(inferenceTemplate.spec.command).to.equal(expectedData.command);
                    expect(inferenceTemplate.spec.args).to.equal(expectedData.args);
                    expect(inferenceTemplate.spec.workingDir).to.equal(expectedData.workingDir);
                    expect(inferenceTemplate.spec.createHomeDir).to.be.equal(expectedData.createHomeDir);

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

                    // general section
                    expect(inferenceTemplate.spec.annotations.length).to.equal(
                        expectedData.annotations.length
                    );
                    expect(inferenceTemplate.spec.labels.length).to.equal(expectedData.labels.length);
                }

                const originTemplate = await createTemplate(metaConfigOriginal, templateSpec);
                expect(originTemplate).to.not.be.undefined;
                logger.debug(`origin template: ${JSON.stringify(originTemplate)}`);
                logger.debug(`origin template spec: ${JSON.stringify(templateSpec)}`);

                expectTemplate(originTemplate, templateSpec);

                await workloadTemplate.page.reload();
                await workloadTemplate.toggleTemplate(metaConfigOriginal.templateName);
                await workloadTemplate.clickOnEditTemplate();

                const envData: IEnvironmentOptions = {
                    image: expectedTemplateSpec.image,
                    runtimeSettings: {
                        isEditCmd: true,
                        cmd: expectedTemplateSpec.command,
                        args: expectedTemplateSpec.args
                    },
                    servingPort: {
                        protocol: expectedTemplateSpec.servingPort.protocol.toString(),
                        port: expectedTemplateSpec.servingPort.container,
                        access: {
                            option: IServingPortAccessOptions.Public
                        }
                    }
                };

                const computeData: IComputeOptions = {
                    assetName: 'cpu-only',
                    autoScaling: {
                        minReplicas: expectedTemplateSpec.autoscaling.minReplicas,
                        maxReplicas: expectedTemplateSpec.autoscaling.maxReplicas,
                        thresholdValue: expectedTemplateSpec.autoscaling.metricThreshold
                    }
                };

                const dataSourceData: IDataSourceOptions = {
                    nfs: objectGenerator.generateNameByPrefix('nfs-storage')
                };

                await inferenceTemplateUIUtil.fillEnvironmentData(envData);
                await inferenceTemplateUIUtil.fillComputeData(computeData);
                await inferenceTemplateUIUtil.fillDataSourceData(dataSourceData);
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

                // check the nfs data source
                expect(editedTemplate.spec.storage.nfs[0].path).to.equal(
                    expectedTemplateSpec.storage.nfs[0].path
                );
                expect(editedTemplate.spec.storage.nfs[0].server).to.equal(
                    expectedTemplateSpec.storage.nfs[0].server
                );
                expect(editedTemplate.spec.storage.nfs[0].mountPath).to.equal(
                    expectedTemplateSpec.storage.nfs[0].mountPath
                );
                expect(editedTemplate.spec.storage.nfs[0].readOnly).to.equal(
                    expectedTemplateSpec.storage.nfs[0].readOnly
                );

                logger.debug(`inferenceTemplate: ${templateName} created successfully!`);
            });
        });
    }
);
