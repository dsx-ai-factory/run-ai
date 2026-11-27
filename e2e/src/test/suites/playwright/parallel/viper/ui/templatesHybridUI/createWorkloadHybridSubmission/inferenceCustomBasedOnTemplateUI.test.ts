/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { BrowserEngine } from '../../../../../../../../modules/ui/runaiV2/pages/BrowserEngine';
import { AuthPage } from '../../../../../../../../modules/ui/runaiV2/pages/AuthPage';
import {
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
    Capability,
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
    Annotation,
    Label,
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
import { HybridForm } from '../../../../../../../../modules/ui/runaiV2/pages/HybridSubmissionForm';
import { ElementValidationUtils } from '../../../../../../../../modules/ui/utils/ElementValidationUtils';

test.describe(
    `Inferences based on template UI, tags: ${tags.tool.ui} ${tags.resource.inferenceWorkload} ${tags.teams.viper}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        const suiteName = 'inference-custom-from-template';
        let project: Project;
        let traceName: string;

        let browserEngine: BrowserEngine;
        let auth: AuthPage;
        let workload: WorkloadPage;
        let inferenceFlatUIUtil: InferenceFlatUIUtil;
        let hybridForm: HybridForm;
        let inferenceName: string;
        let inferenceWorkload: Inference;
        let template: InferenceTemplate;
        let elementValidationUtils: ElementValidationUtils;

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

        function findItem(
            items: Array<Annotation | Label> | null,
            itemName: string,
            itemValue: string
        ): boolean {
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

        test.beforeAll(
            'before all create inference based on custom inference template UI',
            async function () {
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

                logger.info(`created project ${projectName}`);

                browserEngine = await initBrowser();

                logger.info('Browser initialized and user signed in');
            }
        );

        test.afterAll(
            'after all create inference based on custom inference template UI',
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

        test.beforeEach(
            'before each create inference based on custom inference template UI',
            async function () {
                if (!browserEngine || !browserEngine.browser) {
                    browserEngine = await initBrowser();
                }

                inferenceName = objectGenerator.generateNameByPrefix('inference');

                // Generate trace name from test title
                traceName = objectGenerator.generateNameByPrefix(suiteName);

                // Start tracing for this test (context and page stay alive)
                await browserEngine.startTracing(traceName);

                // Initialize page objects for this test
                elementValidationUtils = new ElementValidationUtils(browserEngine.currentPage);

                // Navigate to workloads page for each test
                await workload.navigateToWorkloads();
            }
        );

        test.afterEach(
            'after each create inference based on custom inference template UI',
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
            test('should create new flat inference using an existing template', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('inf-template')
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
                    annotations: [
                        { name: 'ann1', value: 'ann-val1' },
                        { name: 'ann2', value: 'ann-val2' }
                    ],
                    labels: [{ name: 'lbl1', value: 'lbl-val1' }]
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                expect(inferenceWorkload.spec.image).to.be.equals(templateSpec.image);
                expect(inferenceWorkload.spec.command).to.be.equals(templateSpec.command);
                expect(inferenceWorkload.spec.args).to.be.equals(templateSpec.args);
                expect(inferenceWorkload.spec.workingDir).to.be.equals(templateSpec.workingDir);
                expect(inferenceWorkload.spec.createHomeDir).to.be.equals(templateSpec.createHomeDir);
                expect(inferenceWorkload.spec.servingPort.container).to.be.equals(
                    templateSpec.servingPort.container
                );
                expect(inferenceWorkload.spec.servingPort.protocol).to.be.equals(
                    templateSpec.servingPort.protocol
                );
                expect(inferenceWorkload.spec.servingPort.authorizationType).to.be.equals(
                    templateSpec.servingPort.authorizationType
                );
                expect(inferenceWorkload.spec.servingPort.authorizedGroups).to.deep.equals(
                    templateSpec.servingPort.authorizedGroups
                );
                expect(inferenceWorkload.spec.compute.gpuDevicesRequest).to.be.equals(
                    templateSpec.compute.gpuDevicesRequest
                );
                expect(inferenceWorkload.spec.autoscaling.minReplicas).to.be.equals(
                    templateSpec.autoscaling.minReplicas
                );
                expect(inferenceWorkload.spec.autoscaling.maxReplicas).to.be.equals(
                    templateSpec.autoscaling.maxReplicas
                );
                expect(inferenceWorkload.spec.autoscaling.metric).to.be.equals(
                    templateSpec.autoscaling.metric
                );
                expect(inferenceWorkload.spec.autoscaling.metricThreshold).to.be.equals(
                    templateSpec.autoscaling.metricThreshold
                );
                expect(inferenceWorkload.spec.autoscaling.scaleToZeroRetentionSeconds).to.be.equals(
                    templateSpec.autoscaling.scaleToZeroRetentionSeconds
                );
                expect(findItem(inferenceWorkload.spec.annotations, 'ann1', 'ann-val1')).to.be.true;
                expect(findItem(inferenceWorkload.spec.annotations, 'ann2', 'ann-val2')).to.be.true;
                expect(findItem(inferenceWorkload.spec.labels, 'lbl1', 'lbl-val1')).to.be.true;
                logger.info('test inference template passed!');
            });

            test('should create new flat inference using an existing template with pvc', async function () {
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
                    templateName: objectGenerator.generateNameByPrefix('inf-template')
                };

                const pvcStorage = {
                    path: pvc.spec.path,
                    existingPvc: true,
                    claimName: pvc.spec.claimName,
                    readOnly: pvc.spec.readOnly
                };
                const templateSpec: InferenceSpecSpec = {
                    storage: {
                        pvc: [pvcStorage]
                    }
                };

                template = await createTemplate(metaConfig, templateSpec);
                expect(template).to.not.be.undefined;
                logger.debug(`template: ${JSON.stringify(template)}`);

                await inferenceFlatUIUtil.createFlatWorkload({
                    clusterName: flowRootState.cluster.name,
                    projectName: project.name,
                    workloadName: inferenceName,
                    templateName: metaConfig.templateName,
                    environmentSection: {
                        image: 'ubuntu',
                        servingPort: {
                            port: 4000
                        }
                    }
                });

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                expect(inferenceWorkload.spec.storage.pvc[0]).to.be.exist;
                logger.info('test inference template pvc storage passed!');
            });

            test('should test the inference environment section elements using an existing template', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('inf-template')
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
                    security: {
                        uidGidSource: UidGidSource.Custom,
                        runAsUid: 1000,
                        runAsGid: 1000,
                        supplementalGroups: '1000,1001',
                        capabilities: [Capability.NetBindService]
                    },
                    environmentVariables: [{ name: 'env1', value: 'env-val1' }]
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

                // environment section
                const hybridEnvironmentSection = hybridForm.hybridEnvironmentSection;

                await elementValidationUtils.expectInputValueToEqual(
                    hybridForm.hybridEnvironmentSection.hybridImageInput,
                    templateSpec.image
                );

                // serving port section
                await elementValidationUtils.expectInputValueToEqual(
                    hybridEnvironmentSection.servingPortPort,
                    templateSpec.servingPort.container.toString()
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridEnvironmentSection.servingPortProtocol,
                    templateSpec.servingPort.protocol.toString()
                );

                const authenticatedUsersLabel = 'Groups (2)';
                await elementValidationUtils.expectInputValueToEqual(
                    hybridEnvironmentSection.servingPortAccess.servingPortAccessQuickView,
                    authenticatedUsersLabel
                );

                await hybridForm
                    .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessEditBtn)
                    .click();

                await elementValidationUtils.expectAuthorizedValueToContain(
                    hybridEnvironmentSection.servingPortAccess.servingPortAccessModalOptions,
                    'Specific group(s)'
                );

                await hybridForm
                    .locator(hybridEnvironmentSection.servingPortAccess.servingPortAccessModalSaveBtn)
                    .click();

                // runtime section
                await hybridForm.openRuntimeSection();
                await elementValidationUtils.expectInputValueToEqual(
                    hybridEnvironmentSection.cmdInput,
                    templateSpec.command
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridEnvironmentSection.argsInput,
                    templateSpec.args
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridEnvironmentSection.envVarNameInput,
                    templateSpec.environmentVariables[0].name
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridEnvironmentSection.envVarValue,
                    templateSpec.environmentVariables[0].value
                );

                // security section
                await hybridForm.openSecuritySection();
                await elementValidationUtils.expectInputValueToEqual(
                    hybridEnvironmentSection.uidInput,
                    templateSpec.security.runAsUid.toString()
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridEnvironmentSection.gidInput,
                    templateSpec.security.runAsGid.toString()
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridEnvironmentSection.supplementalGroupsInput,
                    templateSpec.security.supplementalGroups
                );
                await elementValidationUtils.expectElementValueToContain(
                    '.capabilities-select',
                    templateSpec.security.capabilities[0].toString()
                );

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                logger.info('test inference template environment section passed!');
            });

            test('should test the inference compute section elements using an existing template', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('inf-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image: 'ubuntu',
                    servingPort: {
                        container: 3131
                    },
                    compute: {
                        gpuDevicesRequest: 2,
                        cpuCoreRequest: 1,
                        cpuCoreLimit: 2
                    },
                    autoscaling: {
                        minReplicas: 0,
                        maxReplicas: 5,
                        metric: 'concurrency',
                        metricThreshold: 50,
                        scaleToZeroRetentionSeconds: 300
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

                const hybridComputeSection = hybridForm.hybridComputeSection;

                await elementValidationUtils.expectInputValueToEqual(
                    hybridComputeSection.gpuDevicesInput,
                    templateSpec.compute.gpuDevicesRequest.toString()
                );
                const cpuRequestValue = templateSpec.compute.cpuCoreRequest * 1000; // To MB
                await elementValidationUtils.expectInputValueToEqual(
                    hybridComputeSection.cpuCoresRequestInput,
                    cpuRequestValue.toString()
                );
                const cpuLimitValue = templateSpec.compute.cpuCoreLimit * 1000; // To MB
                await elementValidationUtils.expectInputValueToEqual(
                    hybridComputeSection.cpuCoresLimitInput,
                    cpuLimitValue.toString()
                );

                // autoscaling section
                await hybridForm.openAutoscalingSection();
                await elementValidationUtils.expectInputValueToEqual(
                    hybridComputeSection.autoScaling.minReplicasInput,
                    templateSpec.autoscaling.minReplicas.toString()
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridComputeSection.autoScaling.maxReplicasInput,
                    templateSpec.autoscaling.maxReplicas.toString()
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridComputeSection.autoScaling.metricSelect,
                    templateSpec.autoscaling.metric
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridComputeSection.autoScaling.metricThresholdInput,
                    templateSpec.autoscaling.metricThreshold.toString()
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridComputeSection.autoScaling.scaleDownSelect,
                    'After 5 minutes of inactivity'
                );

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                logger.info('test inference template compute section passed!');
            });

            test('should test the inference data source section elements using an existing template', async function () {
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
                    templateName: objectGenerator.generateNameByPrefix('inf-template')
                };

                const pvcStorage = {
                    path: pvc.spec.path,
                    existingPvc: true,
                    claimName: pvc.spec.claimName,
                    readOnly: pvc.spec.readOnly
                };

                const templateSpec: InferenceSpecSpec = {
                    image: 'ubuntu',
                    servingPort: {
                        container: 3131
                    },
                    storage: {
                        pvc: [pvcStorage]
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

                const hybridDataSourceSection = hybridForm.hybridDataSourceSection;

                await hybridForm.locator(hybridDataSourceSection.sectionSelector).click();

                await elementValidationUtils.expectElementValueToContain(
                    '.claim-name-select .selected-value-container',
                    pvc.spec.claimName
                );
                await elementValidationUtils.expectInputValueToEqual(
                    hybridDataSourceSection.containerPath,
                    pvcSpec.path
                );

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                logger.info('test inference template data source section passed!');
            });

            test('should test the inference general section elements using an existing template', async function () {
                const metaConfig = {
                    workloadType: WorkloadTypeApi.Inference,
                    scopeType: ScopeType.Project,
                    scopeId: project.id.toString(),
                    templateName: objectGenerator.generateNameByPrefix('inf-template')
                };

                const templateSpec: InferenceSpecSpec = {
                    image: 'ubuntu',
                    servingPort: {
                        container: 3131
                    },
                    annotations: [
                        {
                            name: 'test-ant-var',
                            value: 'test-ant-value'
                        }
                    ],
                    labels: [
                        {
                            name: 'test-lbl-var',
                            value: 'test-lbl-value'
                        }
                    ]
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

                const hybridGeneralSection = hybridForm.hybridGeneralSection;

                await hybridForm.locator(hybridGeneralSection.sectionSelector).click();

                await elementValidationUtils.expectInputValueToEqual(
                    '.annotations-section [aid="itemized-name-input"]',
                    templateSpec.annotations[0].name
                );
                await elementValidationUtils.expectInputValueToEqual(
                    '.annotations-section [aid="itemized-value-input"]',
                    templateSpec.annotations[0].value
                );
                await elementValidationUtils.expectInputValueToEqual(
                    '.labels-section [aid="itemized-name-input"]',
                    templateSpec.labels[0].name
                );
                await elementValidationUtils.expectInputValueToEqual(
                    '.labels-section [aid="itemized-value-input"]',
                    templateSpec.labels[0].value
                );

                // create inference
                await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                    inferenceName,
                    project.id.toString()
                );
                expect(inferenceWorkload).to.not.be.undefined;
                logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);
                logger.info('test inference template general section passed!');
            });

            supportedFrom(SupportedVersions.V2_23)(function () {
                test('should test the inference request timeout using an existing template', async function () {
                    const metaConfig = {
                        workloadType: WorkloadTypeApi.Inference,
                        scopeType: ScopeType.Project,
                        scopeId: project.id.toString(),
                        templateName: objectGenerator.generateNameByPrefix('inf-template')
                    };

                    const templateSpec: InferenceSpecSpec = {
                        image: 'ubuntu',
                        servingPort: {
                            container: 3131
                        },
                        servingConfiguration: {
                            initializationTimeoutSeconds: 307, // 5 minutes * 60 (seconds) + 7 seconds
                            requestTimeoutSeconds: 150 // 2 minutes * 60 (seconds) + 30 seconds
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

                    const hybridGeneralSection = hybridForm.hybridGeneralSection;

                    await hybridForm.locator(hybridGeneralSection.sectionSelector).click();

                    const servingConfigurationSection =
                        hybridForm.hybridGeneralSection.servingConfigurationSection;
                    await elementValidationUtils.expectInputValueToEqual(
                        servingConfigurationSection.initializationTimeoutSeconds.minutesInput,
                        '5' // 5 minutes
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        servingConfigurationSection.initializationTimeoutSeconds.secondsInput,
                        '7' // 7 seconds
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        servingConfigurationSection.requestTimeoutSeconds.minutesInput,
                        '2' // 2 minutes
                    );
                    await elementValidationUtils.expectInputValueToEqual(
                        servingConfigurationSection.requestTimeoutSeconds.secondsInput,
                        '30' // 30 seconds
                    );

                    // create inference
                    await inferenceFlatUIUtil.clickOnSubmitInferenceBtn(inferenceName);

                    inferenceWorkload = await workloadsCrudUtils.getInferenceByName(
                        inferenceName,
                        project.id.toString()
                    );
                    expect(inferenceWorkload).to.not.be.undefined;
                    logger.debug(`inferenceWorkload: ${JSON.stringify(inferenceWorkload)}`);

                    expect(inferenceWorkload.spec.servingConfiguration).to.be.exist;
                    expect(
                        inferenceWorkload.spec.servingConfiguration.initializationTimeoutSeconds
                    ).to.be.equal(templateSpec.servingConfiguration.initializationTimeoutSeconds);
                    expect(inferenceWorkload.spec.servingConfiguration.requestTimeoutSeconds).to.be.equal(
                        templateSpec.servingConfiguration.requestTimeoutSeconds
                    );

                    logger.info('test inference template request timeout passed!');
                });
            });
        });
    }
);
